package graph

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jEngine struct {
	driver neo4j.DriverWithContext
}

func NewNeo4jEngine(driver neo4j.DriverWithContext) *Neo4jEngine {
	return &Neo4jEngine{
		driver: driver,
	}
}

func (g *Neo4jEngine) GetNeighbors(ctx context.Context, nodeID string) ([]Edge, error) {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		"MATCH (n:User {id: $id})-[r]->(m) RETURN m.id as to_id, type(r) as edge_type, r.weight as weight",
		map[string]any{"id": nodeID})
	if err != nil {
		return nil, err
	}

	var edges []Edge
	for result.Next(ctx) {
		record := result.Record()
		toID, _ := record.Get("to_id")
		edgeType, _ := record.Get("edge_type")
		
		var w float64 = 1.0
		if weight, ok := record.Get("weight"); ok && weight != nil {
			switch v := weight.(type) {
			case float64:
				w = v
			case int64:
				w = float64(v)
			}
		}
		
		edges = append(edges, Edge{
			ToID:     toID.(string),
			EdgeType: edgeType.(string),
			Weight:   w,
		})
	}
	
	if err = result.Err(); err != nil {
		return nil, err
	}

	return edges, nil
}

func (g *Neo4jEngine) GetMutualConnectionsCount(ctx context.Context, user1, user2 string) (int, error) {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		"MATCH (u1:User {id: $user1})-[]-(m)-[]-(u2:User {id: $user2}) RETURN count(DISTINCT m) as count",
		map[string]any{"user1": user1, "user2": user2})
	if err != nil {
		return 0, err
	}

	if result.Next(ctx) {
		record := result.Record()
		count, _ := record.Get("count")
		return int(count.(int64)), nil
	}
	return 0, nil
}

func (g *Neo4jEngine) GetNetworkSize(ctx context.Context, nodeID string, maxDepth int) (int, error) {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := fmt.Sprintf("MATCH (n:User {id: $id})-[*1..%d]-(m) RETURN count(DISTINCT m) as count", maxDepth)
	result, err := session.Run(ctx, query, map[string]any{"id": nodeID})
	if err != nil {
		return 0, err
	}

	if result.Next(ctx) {
		record := result.Record()
		count, _ := record.Get("count")
		return int(count.(int64)), nil
	}
	return 0, nil
}
