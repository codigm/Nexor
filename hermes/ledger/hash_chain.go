package ledger

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateHash creates a cryptographic hash for a journal entry
func GenerateHash(prevHash, transactionId, amount, timestamp string) string {
	data := fmt.Sprintf("%s%s%s%s", prevHash, transactionId, amount, timestamp)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
