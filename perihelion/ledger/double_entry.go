package ledger

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type JournalEntry struct {
	TransactionID string
	DebitAccount  string
	CreditAccount string
	Amount        string // Exact string format to maintain precision
	Currency      string
}

// WriteDoubleEntry performs the strict INSERT-only double-entry bookkeeping with a hash chain
func WriteDoubleEntry(ctx context.Context, db *pgxpool.Pool, entry JournalEntry) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Fetch previous hash for cryptographic chain
	var prevHash string
	err = tx.QueryRow(ctx, "SELECT entry_hash FROM journal_entries ORDER BY created_at DESC LIMIT 1").Scan(&prevHash)
	if err != nil {
		// Genesis block hash if no entries exist
		prevHash = "0000000000000000000000000000000000000000000000000000000000000000"
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	hash := GenerateHash(prevHash, entry.TransactionID, entry.Amount, timestamp)

	_, err = tx.Exec(ctx, `
		INSERT INTO journal_entries 
		(transaction_id, debit_account, credit_account, amount, currency, entry_hash, prev_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		entry.TransactionID, entry.DebitAccount, entry.CreditAccount, entry.Amount, entry.Currency, hash, prevHash)
	
	if err != nil {
		return fmt.Errorf("failed to insert journal entry: %w", err)
	}

	return tx.Commit(ctx)
}
