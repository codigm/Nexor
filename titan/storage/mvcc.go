package storage

import "sync"

// Multi-Version Concurrency Control (MVCC) Manager

type MVCCManager struct {
	mu           sync.RWMutex
	activeTxns   map[uint64]bool
	nextTxnID    uint64
	minActiveTxn uint64
}

func NewMVCCManager() *MVCCManager {
	return &MVCCManager{
		activeTxns:   make(map[uint64]bool),
		nextTxnID:    1,
		minActiveTxn: 1,
	}
}

func (m *MVCCManager) BeginTransaction() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	txnID := m.nextTxnID
	m.nextTxnID++
	m.activeTxns[txnID] = true
	
	return txnID
}

func (m *MVCCManager) CommitTransaction(txnID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.activeTxns, txnID)
	m.updateMinActive()
}

func (m *MVCCManager) RollbackTransaction(txnID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.activeTxns, txnID)
	m.updateMinActive()
}

func (m *MVCCManager) updateMinActive() {
	min := m.nextTxnID
	for id := range m.activeTxns {
		if id < min {
			min = id
		}
	}
	m.minActiveTxn = min
}

// IsVisible checks if a row is visible to a given transaction
// based on the row's created_txn_id and deleted_txn_id.
func (m *MVCCManager) IsVisible(queryTxnID, createdTxnID, deletedTxnID uint64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Row must be created by a committed transaction (or by us)
	createdCommitted := createdTxnID < queryTxnID && !m.activeTxns[createdTxnID]
	createdByUs := createdTxnID == queryTxnID

	if !createdCommitted && !createdByUs {
		return false
	}

	// Row must not be deleted by a committed transaction (or by us)
	if deletedTxnID != 0 {
		deletedCommitted := deletedTxnID < queryTxnID && !m.activeTxns[deletedTxnID]
		deletedByUs := deletedTxnID == queryTxnID
		if deletedCommitted || deletedByUs {
			return false
		}
	}

	return true
}
