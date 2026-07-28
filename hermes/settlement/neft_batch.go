package settlement

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// QueueNEFT queues a transaction for batch settlement (amount < Rs 10,000)
func QueueNEFT(ctx context.Context, rdb *redis.Client, txID, sender, receiver, amount string) error {
	log.Printf("[NEFT] Queueing transaction %s from %s to %s for %s", txID, sender, receiver, amount)
	entry := strings.Join([]string{txID, sender, receiver, amount}, "|")
	return rdb.LPush(ctx, "neft_queue", entry).Err()
}

// ProcessNEFTBatch processes queued NEFT transactions.
func ProcessNEFTBatch(ctx context.Context, db *pgxpool.Pool, rdb *redis.Client) {
	log.Println("[NEFT] Starting NEFT batch processor...")
	for {
		result, err := rdb.BRPop(ctx, 0, "neft_queue").Result()
		if err != nil {
			log.Printf("[NEFT] Error reading from queue: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		entry := result[1]
		parts := strings.Split(entry, "|")
		if len(parts) != 4 {
			log.Printf("[NEFT] Invalid entry format: %s", entry)
			continue
		}

		txID, sender, receiver, amount := parts[0], parts[1], parts[2], parts[3]
		log.Printf("[NEFT] Processing batch entry for %s", txID)
		
		// For MVP, we process them sequentially reusing the RTGS DB logic.
		// A full implementation would calculate net positions before hitting the DB.
		err = ProcessRTGS(ctx, db, txID, sender, receiver, amount)
		if err != nil {
			log.Printf("[NEFT] Failed to process %s: %v", txID, err)
			rdb.LPush(ctx, "neft_dlq", entry)
		} else {
			log.Printf("[NEFT] Successfully batch settled %s", txID)
		}
	}
}
