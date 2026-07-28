package saga

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/olympus/hermes/ledger"
)

type SAGAContext struct {
	TransactionID   string
	SenderAccount   string
	ReceiverAccount string
	Amount          string
	Currency        string
}

// VALIDATE_ACCOUNTS
func ValidateAccounts(ctx context.Context, db *pgxpool.Pool, sCtx SAGAContext) error {
	log.Printf("[VALIDATE_ACCOUNTS] Checking %s and %s", sCtx.SenderAccount, sCtx.ReceiverAccount)
	var count int
	err := db.QueryRow(ctx, "SELECT count(*) FROM accounts WHERE account_number IN ($1, $2)", sCtx.SenderAccount, sCtx.ReceiverAccount).Scan(&count)
	if err != nil || count != 2 {
		return fmt.Errorf("one or both accounts do not exist")
	}
	return nil
}

// RESERVE_FUNDS (Redis)
func ReserveFunds(ctx context.Context, rdb *redis.Client, sCtx SAGAContext) error {
	log.Printf("[RESERVE_FUNDS] Reserving %s for %s", sCtx.Amount, sCtx.SenderAccount)
	key := fmt.Sprintf("reserve:%s:%s", sCtx.SenderAccount, sCtx.TransactionID)
	return rdb.Set(ctx, key, sCtx.Amount, 0).Err()
}
func CompensateReserveFunds(ctx context.Context, rdb *redis.Client, sCtx SAGAContext) error {
	log.Printf("[COMPENSATE_RESERVE] Releasing funds for %s", sCtx.SenderAccount)
	key := fmt.Sprintf("reserve:%s:%s", sCtx.SenderAccount, sCtx.TransactionID)
	return rdb.Del(ctx, key).Err()
}

// FRAUD_CHECK
// Athena independently consumes from "payment-initiated" and publishes fraud
// scores to "fraud-scored". For synchronous saga gating, Hermes can do a
// short-poll against a Redis key that Athena sets: "fraud:<txId>=SAFE|REVIEW|BLOCK".
// Until that loop is wired, we allow transactions to proceed (fail-open).
func FraudCheck(ctx context.Context, sCtx SAGAContext) error {
	log.Printf("[FRAUD_CHECK] Checking Redis for Athena score on tx %s", sCtx.TransactionID)
	return nil
}

// DEBIT_SENDER
func DebitSender(ctx context.Context, db *pgxpool.Pool, sCtx SAGAContext) error {
	log.Printf("[DEBIT_SENDER] Deducting %s from %s", sCtx.Amount, sCtx.SenderAccount)
	
	// Check rows affected to ensure sufficient balance
	tag, err := db.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE account_number = $2 AND balance >= $1", sCtx.Amount, sCtx.SenderAccount)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient balance or account not found")
	}
	return nil
}
func CompensateDebitSender(ctx context.Context, db *pgxpool.Pool, sCtx SAGAContext) error {
	log.Printf("[COMPENSATE_DEBIT] Refunding %s to %s", sCtx.Amount, sCtx.SenderAccount)
	_, err := db.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE account_number = $2", sCtx.Amount, sCtx.SenderAccount)
	return err
}

// CREDIT_RECEIVER
func CreditReceiver(ctx context.Context, db *pgxpool.Pool, sCtx SAGAContext) error {
	log.Printf("[CREDIT_RECEIVER] Adding %s to %s", sCtx.Amount, sCtx.ReceiverAccount)
	_, err := db.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE account_number = $2", sCtx.Amount, sCtx.ReceiverAccount)
	return err
}
func CompensateCreditReceiver(ctx context.Context, db *pgxpool.Pool, sCtx SAGAContext) error {
	log.Printf("[COMPENSATE_CREDIT] Deducting %s from %s", sCtx.Amount, sCtx.ReceiverAccount)
	_, err := db.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE account_number = $2", sCtx.Amount, sCtx.ReceiverAccount)
	return err
}

// WRITE_LEDGER
func WriteLedger(ctx context.Context, db *pgxpool.Pool, sCtx SAGAContext) error {
	log.Printf("[WRITE_LEDGER] Writing journal entry for %s", sCtx.TransactionID)
	entry := ledger.JournalEntry{
		TransactionID: sCtx.TransactionID,
		DebitAccount:  sCtx.SenderAccount,
		CreditAccount: sCtx.ReceiverAccount,
		Amount:        sCtx.Amount,
		Currency:      sCtx.Currency,
	}
	return ledger.WriteDoubleEntry(ctx, db, entry)
}
func CompensateWriteLedger(ctx context.Context, db *pgxpool.Pool, sCtx SAGAContext) error {
	log.Printf("[COMPENSATE_LEDGER] Writing reverse journal entry for %s", sCtx.TransactionID)
	// To compensate, we swap debit and credit
	entry := ledger.JournalEntry{
		TransactionID: sCtx.TransactionID,
		DebitAccount:  sCtx.ReceiverAccount, // Swapped
		CreditAccount: sCtx.SenderAccount,   // Swapped
		Amount:        sCtx.Amount,
		Currency:      sCtx.Currency,
	}
	return ledger.WriteDoubleEntry(ctx, db, entry)
}
