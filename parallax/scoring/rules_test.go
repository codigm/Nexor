package scoring

import (
	"context"
	"testing"
	"time"

	"github.com/nexor/parallax/graph"
	"github.com/stretchr/testify/assert"
)

// MockGraphEngine implements graph.GraphEngine
type MockGraphEngine struct {
	MutualConnections int
}

func (m *MockGraphEngine) GetNeighbors(ctx context.Context, nodeID string) ([]graph.Edge, error) {
	return nil, nil
}

func (m *MockGraphEngine) GetMutualConnectionsCount(ctx context.Context, user1, user2 string) (int, error) {
	return m.MutualConnections, nil
}

func (m *MockGraphEngine) GetNetworkSize(ctx context.Context, nodeID string, maxDepth int) (int, error) {
	return 0, nil
}

// MockUserHistory implements UserHistory
type MockUserHistory struct {
	AvgAmount      float64
	SeenDevice     bool
	BankTransfers  int
	UsualLocation  Location
	HasPaid        bool
}

func (m *MockUserHistory) GetAverageAmount(ctx context.Context, userID string) (float64, error) {
	return m.AvgAmount, nil
}

func (m *MockUserHistory) HasSeenDevice(ctx context.Context, userID, deviceID string) (bool, error) {
	return m.SeenDevice, nil
}

func (m *MockUserHistory) GetBankTransfersCount(ctx context.Context, userID string, duration time.Duration) (int, error) {
	return m.BankTransfers, nil
}

func (m *MockUserHistory) GetUsualLocation(ctx context.Context, userID string) (Location, error) {
	return m.UsualLocation, nil
}

func (m *MockUserHistory) HasPaidReceiver(ctx context.Context, senderID, receiverID string) (bool, error) {
	return m.HasPaid, nil
}

func TestEvaluateRules_AllRulesTrigger(t *testing.T) {
	ctx := context.Background()

	// Setup mocks to trigger all rules
	mockHistory := &MockUserHistory{
		AvgAmount:     100.0, // Txn is 600, so > 5x
		SeenDevice:    false, // Triggers NEW_DEVICE
		BankTransfers: 5,     // > 3 triggers VELOCITY_SPIKE
		UsualLocation: Location{Lat: 10.0, Lon: 10.0},
		HasPaid:       false, // Triggers FIRST_TIME_RECEIVER
	}
	
	mockGraph := &MockGraphEngine{
		MutualConnections: 0, // Triggers MERCHANT_ISOLATION
	}

	// Transaction far away in location and unusual hour
	txn := Transaction{
		ID:         "tx-123",
		SenderID:   "USER_A",
		ReceiverID: "USER_B",
		Amount:     600.0,
		Timestamp:  time.Date(2023, 1, 1, 3, 0, 0, 0, time.UTC), // 3 AM (UNUSUAL_HOUR)
		DeviceID:   "DEV_999",
		MerchantID: "MERCH_1",
		Location:   Location{Lat: 20.0, Lon: 20.0}, // Far from 10,10 (NEW_LOCATION)
	}

	results := EvaluateRules(ctx, txn, mockHistory, mockGraph)

	// All 7 rules should trigger
	assert.Equal(t, 7, len(results))

	expectedRules := map[string]bool{
		"UNUSUAL_HOUR":        true,
		"AMOUNT_ANOMALY":      true,
		"NEW_DEVICE":          true,
		"MERCHANT_ISOLATION":  true,
		"VELOCITY_SPIKE":      true,
		"NEW_LOCATION":        true,
		"FIRST_TIME_RECEIVER": true,
	}

	for _, res := range results {
		_, ok := expectedRules[res.Name]
		assert.True(t, ok, "Unexpected rule triggered: "+res.Name)
	}
}

func TestEvaluateRules_NoRulesTrigger(t *testing.T) {
	ctx := context.Background()

	// Setup mocks so NO rules trigger
	mockHistory := &MockUserHistory{
		AvgAmount:     100.0,
		SeenDevice:    true,
		BankTransfers: 1,
		UsualLocation: Location{Lat: 10.0, Lon: 10.0},
		HasPaid:       true,
	}
	
	mockGraph := &MockGraphEngine{
		MutualConnections: 5, 
	}

	txn := Transaction{
		ID:         "tx-124",
		SenderID:   "USER_A",
		ReceiverID: "USER_B",
		Amount:     150.0, // Not 5x
		Timestamp:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), // Noon (Normal hour)
		DeviceID:   "DEV_123",
		MerchantID: "MERCH_1",
		Location:   Location{Lat: 10.0, Lon: 10.0}, // Same location
	}

	results := EvaluateRules(ctx, txn, mockHistory, mockGraph)

	// 0 rules should trigger
	assert.Equal(t, 0, len(results))
}
