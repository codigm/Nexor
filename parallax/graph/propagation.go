package graph

import (
	"context"
)

// Risk Propagation (PageRank-Inspired)
// Risk decays with each hop: 1.0 at hop 1, 0.5 at hop 2, 0.25 at hop 3. Maximum traversal depth: 3 hops.

type RiskPropagator struct {
	engine GraphEngine
}

func NewRiskPropagator(engine GraphEngine) *RiskPropagator {
	return &RiskPropagator{
		engine: engine,
	}
}

func (rp *RiskPropagator) PropagateRisk(ctx context.Context, userID string) float64 {
	visited := make(map[string]bool)
	// Start hop 1 with 1.0 decay
	return rp.dfsRisk(ctx, userID, visited, 1, 1.0)
}

func (rp *RiskPropagator) dfsRisk(ctx context.Context, nodeID string, visited map[string]bool, hop int, decay float64) float64 {
	if hop > 3 {
		return 0.0 // Beyond hop 3: ignored
	}

	visited[nodeID] = true
	var totalRisk float64 = 0.0

	neighbors, err := rp.engine.GetNeighbors(ctx, nodeID)
	if err != nil {
		return 0.0
	}

	for _, n := range neighbors {
		if !visited[n.ToID] {
			// In a real system, check if neighbor is a fraudster.
			// Simulated base risk using edge weight.
			nodeRisk := n.Weight * decay
			totalRisk += nodeRisk
			
			// Recurse to next hop
			nextDecay := decay * 0.5
			totalRisk += rp.dfsRisk(ctx, n.ToID, visited, hop+1, nextDecay)
		}
	}

	return totalRisk
}
