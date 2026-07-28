package scoring

import (
	"context"
	"math"

	"github.com/olympus/argus/fingerprint"
	"github.com/olympus/argus/graph"
)

// Decision thresholds
const (
	DecisionSafe   = "SAFE"
	DecisionReview = "REVIEW"
	DecisionBlock  = "BLOCK"
)

type FraudScore struct {
	TotalScore       int
	Decision         string
	Confidence       float64
	RuleScore        int
	GraphScore       float64
	AnomalyScore     float64
	RuleResults      []RuleResult
}

type FraudScorer struct {
	graphEngine    graph.GraphEngine
	riskPropagator *graph.RiskPropagator
	userHistory    UserHistory
	fingerprinter  *fingerprint.BehavioralEngine
}

func NewFraudScorer(ge graph.GraphEngine, rp *graph.RiskPropagator, uh UserHistory, fe *fingerprint.BehavioralEngine) *FraudScorer {
	return &FraudScorer{
		graphEngine:    ge,
		riskPropagator: rp,
		userHistory:    uh,
		fingerprinter:  fe,
	}
}

func (fs *FraudScorer) ScoreTransaction(ctx context.Context, txn Transaction) (*FraudScore, error) {
	// 1. Rules (40%)
	ruleResults := EvaluateRules(ctx, txn, fs.userHistory, fs.graphEngine)
	rulePoints := 0
	for _, r := range ruleResults {
		rulePoints += r.Contribution
	}

	// 2. Graph Risk (35%)
	graphRisk := fs.riskPropagator.PropagateRisk(ctx, txn.SenderID)

	// 3. Behavioral Anomaly Score (25%)
	// Fetch actual embedding vector from ATHENA via TITAN integration in fingerprint engine
	anomalyScore, err := fs.fingerprinter.ComputeAnomalyScore(ctx, txn.SenderID, txn)
	if err != nil {
		anomalyScore = 0.0
	}

	// Final Score Formula: (0.4 * rules) + (0.35 * graph risk) + (0.25 * behavioral anomaly)
	// Normalize sub-scores assuming max 100
	normalizedRules := math.Min(100.0, float64(rulePoints))
	normalizedGraph := math.Min(100.0, graphRisk*10) // assuming risk is in 0-10 range originally
	normalizedAnomaly := math.Min(100.0, anomalyScore*100) // anomaly 0.0 - 1.0

	finalScoreFloat := (0.4 * normalizedRules) + (0.35 * normalizedGraph) + (0.25 * normalizedAnomaly)
	finalScore := int(math.Round(finalScoreFloat))

	decision := DecisionSafe
	if finalScore >= 71 {
		decision = DecisionBlock
	} else if finalScore >= 31 {
		decision = DecisionReview
	}

	// Confidence mock
	confidence := 0.85
	if len(ruleResults) > 2 {
		confidence = 0.94
	}

	return &FraudScore{
		TotalScore:   finalScore,
		Decision:     decision,
		Confidence:   confidence,
		RuleScore:    rulePoints,
		GraphScore:   normalizedGraph,
		AnomalyScore: normalizedAnomaly,
		RuleResults:  ruleResults,
	}, nil
}
