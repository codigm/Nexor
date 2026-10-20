package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/nexor/perihelion/circuit_breaker"
	"github.com/nexor/perihelion/iso20022"
	"github.com/nexor/perihelion/settlement"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	perihelionTransactionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "perihelion_transactions_total",
			Help: "Total number of transactions processed",
		},
		[]string{"type", "status"},
	)
	tracer = otel.Tracer("perihelion-core")
)

type Orchestrator struct {
	db     *pgxpool.Pool
	rdb    *redis.Client
	reader *kafka.Reader
	cb     *circuit_breaker.CircuitBreaker
}

func NewOrchestrator(db *pgxpool.Pool, rdb *redis.Client, brokers []string) *Orchestrator {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: "perihelion-saga-group",
		Topic:   "payment-initiated",
	})
	cb := circuit_breaker.NewCircuitBreaker(rdb)
	return &Orchestrator{db: db, rdb: rdb, reader: reader, cb: cb}
}

func (o *Orchestrator) Start(ctx context.Context) {
	log.Println("PERIHELION SAGA Orchestrator started. Listening for payments...")

	// Start NEFT batch processor in background
	go settlement.ProcessNEFTBatch(ctx, o.db, o.rdb)

	// Create saga_state table if not exists
	_, err := o.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS saga_state (
			tx_id VARCHAR(255) PRIMARY KEY,
			state VARCHAR(50),
			updated_at TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create saga_state table: %v", err)
	}

	for {
		m, err := o.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		var req struct {
			TransactionId   string  `json:"transactionId"`
			SenderAccount   string  `json:"senderAccount"`
			ReceiverAccount string  `json:"receiverAccount"`
			Amount          float64 `json:"amount"`
			Currency        string  `json:"currency"`
		}

		// Try parsing as ISO 20022 pacs.008 XML first
		var amountStr string
		var rawAmount float64
		pacsDoc, xmlErr := iso20022.ParsePacs008(m.Value)
		if xmlErr == nil && pacsDoc.FIToFICstmrCdtTrf.GrpHdr.MsgId != "" {
			req.TransactionId = pacsDoc.FIToFICstmrCdtTrf.CdtTrfTxInf.PmtId.EndToEndId
			req.SenderAccount = pacsDoc.FIToFICstmrCdtTrf.CdtTrfTxInf.DbtrAcct.Id.Othr.Id
			req.ReceiverAccount = pacsDoc.FIToFICstmrCdtTrf.CdtTrfTxInf.CdtrAcct.Id.Othr.Id
			amountStr = pacsDoc.FIToFICstmrCdtTrf.CdtTrfTxInf.IntrBkSttlmAmt.Value
			req.Currency = pacsDoc.FIToFICstmrCdtTrf.CdtTrfTxInf.IntrBkSttlmAmt.Ccy
			fmt.Sscanf(amountStr, "%f", &rawAmount)
			log.Printf("[ISO_20022] Successfully parsed pacs.008 message: %s", req.TransactionId)
		} else {
			// Fallback to legacy JSON
			if err := json.Unmarshal(m.Value, &req); err != nil {
				log.Printf("Failed to unmarshal payment (XML and JSON): %v", err)
				continue
			}
			rawAmount = req.Amount
			amountStr = fmt.Sprintf("%.4f", req.Amount)
			log.Printf("[LEGACY_JSON] Processed JSON message: %s", req.TransactionId)
		}

		sCtx := SAGAContext{
			TransactionID:   req.TransactionId,
			SenderAccount:   req.SenderAccount,
			ReceiverAccount: req.ReceiverAccount,
			Amount:          amountStr,
			Currency:        req.Currency,
		}

		go o.executeSaga(context.Background(), sCtx, rawAmount)
	}
}

func (o *Orchestrator) updateState(ctx context.Context, txID, state string) {
	_, err := o.db.Exec(ctx, `
		INSERT INTO saga_state (tx_id, state, updated_at) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (tx_id) DO UPDATE SET state = $2, updated_at = $3
	`, txID, state, time.Now())
	if err != nil {
		log.Printf("Failed to persist SAGA state for %s: %v", txID, err)
	}
}

// Extract bank ID from account number
func getBankID(account string) string {
	if len(account) >= 4 {
		return account[:4] // E.g. ALPH001 -> ALPH
	}
	return "UNKNOWN"
}

func (o *Orchestrator) executeSaga(ctx context.Context, sCtx SAGAContext, rawAmount float64) {
	ctx, span := tracer.Start(ctx, "executeSaga")
	span.SetAttributes(attribute.String("transaction.id", sCtx.TransactionID))
	defer span.End()

	o.updateState(ctx, sCtx.TransactionID, "STARTED")
	log.Printf("--- STARTING SAGA FOR %s ---", sCtx.TransactionID)

	senderBank := getBankID(sCtx.SenderAccount)
	receiverBank := getBankID(sCtx.ReceiverAccount)

	for _, bank := range []string{senderBank, receiverBank} {
		allowed, err := o.cb.AllowRequest(ctx, bank)
		if err != nil || !allowed {
			log.Printf("[SAGA] Circuit Breaker OPEN for %s. Queueing payment %s.", bank, sCtx.TransactionID)
			o.updateState(ctx, sCtx.TransactionID, "QUEUED_CIRCUIT_OPEN")
			settlement.QueueNEFT(ctx, o.rdb, sCtx.TransactionID, sCtx.SenderAccount, sCtx.ReceiverAccount, sCtx.Amount)
			return
		}
	}

	o.updateState(ctx, sCtx.TransactionID, "VALIDATING")
	if err := ValidateAccounts(ctx, o.db, sCtx); err != nil {
		o.failSaga(ctx, sCtx, err, rawAmount)
		return
	}

	o.updateState(ctx, sCtx.TransactionID, "RESERVING")
	if err := ReserveFunds(ctx, o.rdb, sCtx); err != nil {
		o.failSaga(ctx, sCtx, err, rawAmount)
		return
	}

	o.updateState(ctx, sCtx.TransactionID, "FRAUD_CHECK")
	if err := FraudCheck(ctx, o.rdb, sCtx); err != nil {
		CompensateReserveFunds(ctx, o.rdb, sCtx)
		o.failSaga(ctx, sCtx, err, rawAmount)
		return
	}

	o.updateState(ctx, sCtx.TransactionID, "SETTLING")
	var settleErr error
	if rawAmount >= 10000 {
		settleErr = settlement.ProcessRTGS(ctx, o.db, sCtx.TransactionID, sCtx.SenderAccount, sCtx.ReceiverAccount, sCtx.Amount)
	} else {
		settleErr = settlement.QueueNEFT(ctx, o.rdb, sCtx.TransactionID, sCtx.SenderAccount, sCtx.ReceiverAccount, sCtx.Amount)
	}

	if settleErr != nil {
		o.cb.RecordFailure(ctx, senderBank)
		CompensateReserveFunds(ctx, o.rdb, sCtx)
		o.failSaga(ctx, sCtx, settleErr, rawAmount)
		return
	}

	o.cb.RecordSuccess(ctx, senderBank)
	o.cb.RecordSuccess(ctx, receiverBank)

	o.updateState(ctx, sCtx.TransactionID, "LEDGER_WRITE")
	if err := WriteLedger(ctx, o.db, sCtx); err != nil {
		o.failSaga(ctx, sCtx, err, rawAmount)
		return
	}

	o.updateState(ctx, sCtx.TransactionID, "COMPLETED")
	txType := "NEFT"
	if rawAmount >= 10000 {
		txType = "RTGS"
	}
	perihelionTransactionsTotal.WithLabelValues(txType, "success").Inc()
	log.Printf("--- SAGA COMPLETED SUCCESSFULLY FOR %s ---", sCtx.TransactionID)
}

func (o *Orchestrator) failSaga(ctx context.Context, sCtx SAGAContext, err error, rawAmount float64) {
	o.updateState(ctx, sCtx.TransactionID, "FAILED")
	txType := "NEFT"
	if rawAmount >= 10000 {
		txType = "RTGS"
	}
	perihelionTransactionsTotal.WithLabelValues(txType, "failure").Inc()
	log.Printf("--- SAGA FAILED FOR %s: %v ---", sCtx.TransactionID, err)
}
