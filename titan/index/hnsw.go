package index

import (
	"math"
	"math/rand"
	"sort"
)

// HNSW (Hierarchical Navigable Small World) Graph implementation

type Node struct {
	ID          int64
	Vector      []float32
	Connections [][]int64 // layer -> list of neighbors
	MaxLayer    int
}

type HNSW struct {
	Nodes        map[int64]*Node
	EntryPoint   int64
	MaxLayer     int
	M            int
	M0           int
	EfConstruction int
	LevelMult    float64
}

func NewHNSW(m, m0, efConstruction int) *HNSW {
	return &HNSW{
		Nodes:        make(map[int64]*Node),
		EntryPoint:   -1,
		MaxLayer:     0,
		M:            m,
		M0:           m0,
		EfConstruction: efConstruction,
		LevelMult:    1.0 / math.Log(float64(m)),
	}
}

// Distance computation using our AVX2 functions (mocked here for pure go fallback if CGo not linked)
func (h *HNSW) distance(vec1, vec2 []float32) float32 {
	// In production, this calls the CGo AVX2 function
	// For simulation, euclidean distance
	var sum float32 = 0
	for i := range vec1 {
		diff := vec1[i] - vec2[i]
		sum += diff * diff
	}
	return float32(math.Sqrt(float64(sum)))
}

func (h *HNSW) Insert(id int64, vector []float32) {
	level := int(-math.Log(rand.Float64()) * h.LevelMult)
	newNode := &Node{
		ID:          id,
		Vector:      vector,
		Connections: make([][]int64, level+1),
		MaxLayer:    level,
	}
	
	h.Nodes[id] = newNode

	if len(h.Nodes) == 1 {
		h.EntryPoint = id
		h.MaxLayer = level
		return
	}

	currNodeID := h.EntryPoint
	currDist := h.distance(vector, h.Nodes[currNodeID].Vector)

	// Phase 1: Traverse upper layers
	for lc := h.MaxLayer; lc > level; lc-- {
		changed := true
		for changed {
			changed = false
			for _, neighborID := range h.Nodes[currNodeID].Connections[lc] {
				d := h.distance(vector, h.Nodes[neighborID].Vector)
				if d < currDist {
					currDist = d
					currNodeID = neighborID
					changed = true
				}
			}
		}
	}

	// Phase 2: Insert at level and below
	for lc := min(level, h.MaxLayer); lc >= 0; lc-- {
		neighbors := h.searchLayer(currNodeID, vector, h.EfConstruction, lc)
		h.Nodes[id].Connections[lc] = h.selectNeighbors(neighbors, h.M)
		
		// Connect back
		for _, neighborID := range h.Nodes[id].Connections[lc] {
			h.Nodes[neighborID].Connections[lc] = append(h.Nodes[neighborID].Connections[lc], id)
		}
	}

	if level > h.MaxLayer {
		h.MaxLayer = level
		h.EntryPoint = id
	}
}

func (h *HNSW) Search(query []float32, k, ef int) []int64 {
	if len(h.Nodes) == 0 {
		return nil
	}

	currNodeID := h.EntryPoint
	currDist := h.distance(query, h.Nodes[currNodeID].Vector)

	for lc := h.MaxLayer; lc > 0; lc-- {
		changed := true
		for changed {
			changed = false
			for _, neighborID := range h.Nodes[currNodeID].Connections[lc] {
				d := h.distance(query, h.Nodes[neighborID].Vector)
				if d < currDist {
					currDist = d
					currNodeID = neighborID
					changed = true
				}
			}
		}
	}

	candidates := h.searchLayer(currNodeID, query, ef, 0)
	
	var res []int64
	for i := 0; i < k && i < len(candidates); i++ {
		res = append(res, candidates[i].id)
	}
	return res
}

type item struct {
	id   int64
	dist float32
}

func (h *HNSW) searchLayer(entryPoint int64, query []float32, ef, lc int) []item {
	visited := make(map[int64]bool)
	visited[entryPoint] = true

	candidates := []item{{entryPoint, h.distance(query, h.Nodes[entryPoint].Vector)}}
	results := []item{{entryPoint, h.distance(query, h.Nodes[entryPoint].Vector)}}

	for len(candidates) > 0 {
		// Pop closest from candidates
		c := candidates[0]
		candidates = candidates[1:]

		farthestResult := results[len(results)-1]
		if c.dist > farthestResult.dist {
			break
		}

		for _, e := range h.Nodes[c.id].Connections[lc] {
			if !visited[e] {
				visited[e] = true
				d := h.distance(query, h.Nodes[e].Vector)
				if d < farthestResult.dist || len(results) < ef {
					candidates = append(candidates, item{e, d})
					results = append(results, item{e, d})
					sort.Slice(candidates, func(i, j int) bool { return candidates[i].dist < candidates[j].dist })
					sort.Slice(results, func(i, j int) bool { return results[i].dist < results[j].dist })
					if len(results) > ef {
						results = results[:ef]
					}
				}
			}
		}
	}
	return results
}

func (h *HNSW) selectNeighbors(candidates []item, m int) []int64 {
	var res []int64
	for i := 0; i < m && i < len(candidates); i++ {
		res = append(res, candidates[i].id)
	}
	return res
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
