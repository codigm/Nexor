package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/olympus/hermes/saga"
	"github.com/olympus/hermes/tracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("Initializing HERMES core...")

	// Initialize OpenTelemetry Tracing
	jaegerEndpoint := getEnv("JAEGER_ENDPOINT", "localhost:4318")
	tp := tracing.InitTracer("hermes-core", jaegerEndpoint)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Postgres Connection
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		getEnv("DB_USER", "olympus"),
		getEnv("DB_PASSWORD", "olympus_password"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "olympus_db"),
	)
	
	dbPool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Redis Connection
	rdb := redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"),
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}

	// Kafka Connection
	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:29092"), ",")

	// Initialize Orchestrator
	orchestrator := saga.NewOrchestrator(dbPool, rdb, brokers)
	
	// Start Prometheus Metrics Server
	go func() {
		log.Println("Starting Prometheus metrics server on :9091")
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":9091", nil); err != nil {
			log.Fatalf("Metrics server failed: %v", err)
		}
	}()

	// Start blocking
	orchestrator.Start(context.Background())
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
