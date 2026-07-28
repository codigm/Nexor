package scoring

import (
	"context"
	"time"

	"github.com/olympus/argus/graph"
)

type Transaction struct {
	ID          string
	SenderID    string
	ReceiverID  string
	Amount      float64
	Timestamp   time.Time
	DeviceID    string
	MerchantID  string
	Location    Location
}

type Location struct {
	Lat float64
	Lon float64
}

// History mock for the sake of the rules
type UserHistory interface {
	GetAverageAmount(ctx context.Context, userID string) (float64, error)
	HasSeenDevice(ctx context.Context, userID, deviceID string) (bool, error)
	GetBankTransfersCount(ctx context.Context, userID string, duration time.Duration) (int, error)
	GetUsualLocation(ctx context.Context, userID string) (Location, error)
	HasPaidReceiver(ctx context.Context, senderID, receiverID string) (bool, error)
}

type RuleResult struct {
	Name         string `json:"rule"`
	Contribution int    `json:"contribution"`
}

func EvaluateRules(ctx context.Context, txn Transaction, history UserHistory, ge graph.GraphEngine) []RuleResult {
	var results []RuleResult

	// 1. IsUnusualHour (+20 pts)
	hour := txn.Timestamp.Hour()
	if hour < 6 || hour > 23 { // simple mock for outside normal 2-sigma window
		results = append(results, RuleResult{"UNUSUAL_HOUR", 20})
	}

	// 2. IsAmountAnomaly (+30 pts)
	avgAmt, _ := history.GetAverageAmount(ctx, txn.SenderID)
	if avgAmt > 0 && txn.Amount > 5*avgAmt {
		results = append(results, RuleResult{"AMOUNT_ANOMALY", 30})
	}

	// 3. IsNewDevice (+25 pts)
	seenDevice, _ := history.HasSeenDevice(ctx, txn.SenderID, txn.DeviceID)
	if !seenDevice {
		results = append(results, RuleResult{"NEW_DEVICE", 25})
	}

	// 4. IsMerchantIsolated (+15 pts)
	mutual, _ := ge.GetMutualConnectionsCount(ctx, txn.SenderID, txn.MerchantID)
	if txn.MerchantID != "" && mutual == 0 {
		results = append(results, RuleResult{"MERCHANT_ISOLATION", 15})
	}

	// 5. IsVelocitySpike (+35 pts)
	banks, _ := history.GetBankTransfersCount(ctx, txn.SenderID, 10*time.Minute)
	if banks > 3 {
		results = append(results, RuleResult{"VELOCITY_SPIKE", 35})
	}

	// 6. IsNewLocation (+10 pts)
	usualLoc, _ := history.GetUsualLocation(ctx, txn.SenderID)
	dist := calculateDistance(usualLoc, txn.Location) // > 100km
	if dist > 100.0 {
		results = append(results, RuleResult{"NEW_LOCATION", 10})
	}

	// 7. IsFirstTimeReceiver (+10 pts)
	hasPaid, _ := history.HasPaidReceiver(ctx, txn.SenderID, txn.ReceiverID)
	if !hasPaid {
		results = append(results, RuleResult{"FIRST_TIME_RECEIVER", 10})
	}

	return results
}

func calculateDistance(l1, l2 Location) float64 {
	if l1.Lat == 0 && l1.Lon == 0 {
		return 0 // unknown
	}
	dLat := l1.Lat - l2.Lat
	dLon := l1.Lon - l2.Lon
	distDeg := dLat*dLat + dLon*dLon
	return 111.0 * distDeg 
}
