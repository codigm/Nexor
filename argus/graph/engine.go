package graph

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	_ "github.com/lib/pq"
)

// Edge represents a connection between two nodes in the behavioral graph
type Edge struct {
	ToID      string  `json:"to_id"`
	EdgeType  string  `json:"edge_type"`
	Weight    float64 `json:"weight"`
}

// GraphEngine interface allows for easy mocking in tests
type GraphEngine interface {
	GetNeighbors(ctx context.Context, nodeID string) ([]Edge, error)
	GetMutualConnectionsCount(ctx context.Context, user1, user2 string) (int, error)
	GetNetworkSize(ctx context.Context, nodeID string, maxDepth int) (int, error)
}

// PostgresRedisGraphEngine implements GraphEngine with real backends
type PostgresRedisGraphEngine struct {
	db    *sql.DB
	redis *redis.Client
}

func NewPostgresRedisGraphEngine(db *sql.DB, rdb *redis.Client) *PostgresRedisGraphEngine {
	return &PostgresRedisGraphEngine{
		db:    db,
		redis: rdb,
	}
}

// GetNeighbors gets neighbors, prioritizing Redis cache over Postgres
func (g *PostgresRedisGraphEngine) GetNeighbors(ctx context.Context, nodeID string) ([]Edge, error) {
	cacheKey := fmt.Sprintf("graph:neighbors:%s", nodeID)

	// 1. Check Redis (Hot Subgraph Cache)
	val, err := g.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache hit
		var edges []Edge
		if err := json.Unmarshal([]byte(val), &edges); err == nil {
			return edges, nil
		}
	}

	// 2. Fetch from Postgres (Adjacency List with temporal conditions)
	query := `
		SELECT to_id, edge_type, weight
		FROM graph_edges
		WHERE from_id = $1 
		  AND valid_from <= NOW() 
		  AND valid_to > NOW()
	`
	rows, err := g.db.QueryContext(ctx, query, nodeID)
	if err != nil {
		return nil, fmt.Errorf("postgres query failed: %w", err)
	}
	defer rows.Close()

	var edges []Edge
	for rows.Next() {
		var e Edge
		if err := rows.Scan(&e.ToID, &e.EdgeType, &e.Weight); err != nil {
			continue
		}
		edges = append(edges, e)
	}

	// 3. Update Redis Cache (5-minute TTL)
	if data, err := json.Marshal(edges); err == nil {
		g.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return edges, nil
}

func (g *PostgresRedisGraphEngine) GetMutualConnectionsCount(ctx context.Context, user1, user2 string) (int, error) {
	n1, err := g.GetNeighbors(ctx, user1)
	if err != nil {
		return 0, err
	}
	n2, err := g.GetNeighbors(ctx, user2)
	if err != nil {
		return 0, err
	}

	m2 := make(map[string]bool)
	for _, e := range n2 {
		m2[e.ToID] = true
	}

	count := 0
	for _, e := range n1 {
		if m2[e.ToID] {
			count++
		}
	}
	return count, nil
}

func (g *PostgresRedisGraphEngine) GetNetworkSize(ctx context.Context, nodeID string, maxDepth int) (int, error) {
	visited := make(map[string]bool)
	queue := []string{nodeID}
	visited[nodeID] = true

	for depth := 0; depth < maxDepth; depth++ {
		var nextQueue []string
		for _, id := range queue {
			neighbors, err := g.GetNeighbors(ctx, id)
			if err != nil {
				continue
			}
			for _, n := range neighbors {
				if !visited[n.ToID] {
					visited[n.ToID] = true
					nextQueue = append(nextQueue, n.ToID)
				}
			}
		}
		queue = nextQueue
	}

	return len(visited), nil
}
