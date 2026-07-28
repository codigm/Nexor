package scheduler

import (
	"fmt"
	"time"
)

type Request struct {
	ID        string
	Tokens    []int
	Completed bool
}

type ContinuousBatcher struct {
	ActiveRequests []*Request
	Queue          []*Request
}

func (cb *ContinuousBatcher) Run() {
	// Tight scheduler loop
	for {
		cb.removeCompleted()
		cb.fillFreeSlotsFromQueue()
		
		if len(cb.ActiveRequests) == 0 && len(cb.Queue) == 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		cb.executeOneBatch()
		cb.updateMetrics()
	}
}

func (cb *ContinuousBatcher) removeCompleted() {
	active := make([]*Request, 0)
	for _, req := range cb.ActiveRequests {
		if !req.Completed {
			active = append(active, req)
		}
	}
	cb.ActiveRequests = active
}

func (cb *ContinuousBatcher) fillFreeSlotsFromQueue() {
	maxBatchSize := 32
	for len(cb.ActiveRequests) < maxBatchSize && len(cb.Queue) > 0 {
		req := cb.Queue[0]
		cb.Queue = cb.Queue[1:]
		cb.ActiveRequests = append(cb.ActiveRequests, req)
	}
}

func (cb *ContinuousBatcher) executeOneBatch() {
	fmt.Printf("Executing 1 forward pass for %d requests\n", len(cb.ActiveRequests))
	// Simulating forward pass
	for _, req := range cb.ActiveRequests {
		req.Tokens = append(req.Tokens, 1) // predict 1 token
		if len(req.Tokens) > 10 {          // stop condition
			req.Completed = true
		}
	}
}

func (cb *ContinuousBatcher) updateMetrics() {
	// Log GPU utilization and queue depth
}
