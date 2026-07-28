package explain

import (
	"context"
	"encoding/json"

	"github.com/olympus/argus/graph"
	"github.com/olympus/argus/scoring"
)

type Explanation struct {
	TransactionID string               `json:"transaction_id"`
	RiskScore     int                  `json:"risk_score"`
	Decision      string               `json:"decision"`
	Confidence    float64              `json:"confidence"`
	Factors       []scoring.RuleResult `json:"factors"`
	GraphInsights GraphInsight         `json:"graph_insights"`
}

type GraphInsight struct {
	NetworkSize       int `json:"network_size_3hop"`
	MutualConnections int `json:"mutual_merchant_connections"`
}

type ExplainerEngine struct {
	graphEngine graph.GraphEngine
}

func NewExplainerEngine(ge graph.GraphEngine) *ExplainerEngine {
	return &ExplainerEngine{
		graphEngine: ge,
	}
}

func (ee *ExplainerEngine) GenerateExplanation(ctx context.Context, txn scoring.Transaction, score *scoring.FraudScore) (string, error) {
	
	netSize, _ := ee.graphEngine.GetNetworkSize(ctx, txn.SenderID, 3)
	mutual, _ := ee.graphEngine.GetMutualConnectionsCount(ctx, txn.SenderID, txn.MerchantID)

	exp := Explanation{
		TransactionID: txn.ID,
		RiskScore:     score.TotalScore,
		Decision:      score.Decision,
		Confidence:    score.Confidence,
		Factors:       score.RuleResults,
		GraphInsights: GraphInsight{
			NetworkSize:       netSize,
			MutualConnections: mutual,
		},
	}

	data, err := json.MarshalIndent(exp, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
