package settlement

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ProcessRTGS processes real-time gross settlement for amounts >= Rs 10,000.
func ProcessRTGS(ctx context.Context, db *pgxpool.Pool, txID, sender, receiver, amount string) error {
	log.Printf("[RTGS] Processing immediate settlement for %s from %s to %s for %s", txID, sender, receiver, amount)
	
	// Start a transaction for real-time settlement
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Debit Sender
	tag, err := tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE account_number = $2 AND balance >= $1", amount, sender)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient balance or account not found for RTGS")
	}

	// Credit Receiver
	_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE account_number = $2", amount, receiver)
	if err != nil {
		return err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	log.Printf("[RTGS] Successfully settled %s", txID)
	return nil
}
