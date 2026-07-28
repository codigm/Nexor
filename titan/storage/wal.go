package storage

import (
	"encoding/binary"
	"os"
	"sync"
)

type WAL struct {
	mu   sync.Mutex
	file *os.File
}

func NewWAL(walFile string) (*WAL, error) {
	f, err := os.OpenFile(walFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	return &WAL{file: f}, nil
}

// LogWrite appends an entry to the Write-Ahead Log and strictly fsyncs it to disk.
func (w *WAL) LogWrite(txnID uint64, pageID uint16, data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Entry Layout:
	// [0-8]: TxnID
	// [8-10]: PageID
	// [10-14]: Length of data
	// [14-...]: Data
	header := make([]byte, 14)
	binary.LittleEndian.PutUint64(header[0:8], txnID)
	binary.LittleEndian.PutUint16(header[8:10], pageID)
	binary.LittleEndian.PutUint32(header[10:14], uint32(len(data)))

	if _, err := w.file.Write(header); err != nil {
		return err
	}
	if _, err := w.file.Write(data); err != nil {
		return err
	}

	// Crucial: strict fsync on every commit for durability.
	// As requested by the user, we accept the throughput penalty for real database correctness.
	return w.file.Sync()
}

func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}
