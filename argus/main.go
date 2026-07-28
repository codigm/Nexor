package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"

	"github.com/olympus/argus/explain"
	"github.com/olympus/argus/fingerprint"
	"github.com/olympus/argus/graph"
	"github.com/olympus/argus/scoring"
	"github.com/olympus/titan/query"
	_ "github.com/lib/pq"
)

// MockUserHistory implements scoring.UserHistory for demonstration
type MockUserHistory struct{}

func (m MockUserHistory) GetAverageAmount(ctx context.Context, userID string) (float64, error) {
	return 50.0, nil
}
func (m MockUserHistory) HasSeenDevice(ctx context.Context, userID, deviceID string) (bool, error) {
	return true, nil
}
func (m MockUserHistory) GetBankTransfersCount(ctx context.Context, userID string, duration time.Duration) (int, error) {
	return 1, nil
}
func (m MockUserHistory) GetUsualLocation(ctx context.Context, userID string) (scoring.Location, error) {
	return scoring.Location{Lat: 12.9716, Lon: 77.5946}, nil
}
func (m MockUserHistory) HasPaidReceiver(ctx context.Context, senderID, receiverID string) (bool, error) {
	return true, nil
}

func main() {
	log.Println("===========================================")
	log.Println("   ARGUS — Audit & Fraud Engine v2.0       ")
	log.Println("===========================================")

	// 1. Initialize Postgres
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "olympus")
	dbPass := getEnv("DB_PASSWORD", "olympus_password")
	dbName := getEnv("DB_NAME", "olympus_db")

	connStr := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser + " password=" + dbPass + " dbname=" + dbName + " sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("[ARGUS] Failed to connect to Postgres: %v", err)
	}
	defer db.Close()

	// 2. Initialize Redis
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer rdb.Close()

	// 3. Initialize Engines
	graphEngine := graph.NewPostgresRedisGraphEngine(db, rdb)
	riskPropagator := &graph.RiskPropagator{} // Currently empty struct in graph package

	// Initialize embedded TITAN executor as mock for behavioral fingerprinting
	titanExecutor := query.NewExecutor()
	behavioralEngine := fingerprint.NewBehavioralEngine(titanExecutor)

	userHistory := MockUserHistory{}
	fraudScorer := scoring.NewFraudScorer(graphEngine, riskPropagator, userHistory, behavioralEngine)
	explainerEngine := explain.NewExplainerEngine(graphEngine)

	// 4. Initialize Kafka Consumer
	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:29092")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBrokers},
		GroupID: "argus-audit-group",
		Topic:   "transactions",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer reader.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown handling
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigchan
		log.Println("[ARGUS] Shutting down...")
		cancel()
	}()

	log.Println("[ARGUS] Listening for transactions on Kafka topic 'transactions'...")

	// 5. Processing Loop
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("[ARGUS] Failed to read message: %v\n", err)
			continue
		}

		log.Printf("[ARGUS] Auditing transaction at offset %d: %s\n", m.Offset, string(m.Value))

		// Parse Transaction
		var txn scoring.Transaction
		if err := json.Unmarshal(m.Value, &txn); err != nil {
			log.Printf("[ARGUS] Invalid transaction JSON: %v\n", err)
			continue
		}

		// Score Fraud
		score, err := fraudScorer.ScoreTransaction(ctx, txn)
		if err != nil {
			log.Printf("[ARGUS] Failed to score transaction %s: %v\n", txn.ID, err)
			continue
		}

		// Generate Explanation
		explanation, err := explainerEngine.GenerateExplanation(ctx, txn, score)
		if err != nil {
			log.Printf("[ARGUS] Failed to generate explanation for %s: %v\n", txn.ID, err)
			continue
		}

		log.Printf("[ARGUS] RESULT for %s -> Decision: %s, Score: %d\n", txn.ID, score.Decision, score.TotalScore)
		log.Printf("[ARGUS] EXPLANATION: \n%s\n", explanation)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
