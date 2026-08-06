package scheduler

import (
	"fmt"
	"sync"
	"time"
)

type Request struct {
	ID        string
	Tokens    []int
	Completed bool
}

type ContinuousBatcher struct {
	mu             sync.Mutex
	ActiveRequests []*Request
	Queue          []*Request
}

func (cb *ContinuousBatcher) Enqueue(req *Request) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Queue = append(cb.Queue, req)
}

func (cb *ContinuousBatcher) QueueLen() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return len(cb.Queue)
}

func (cb *ContinuousBatcher) ActiveLen() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return len(cb.ActiveRequests)
}

func (cb *ContinuousBatcher) Run() {
	// Tight scheduler loop
	for {
		cb.removeCompleted()
		cb.fillFreeSlotsFromQueue()
		
		if cb.ActiveLen() == 0 && cb.QueueLen() == 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		cb.executeOneBatch()
		cb.updateMetrics()
	}
}

func (cb *ContinuousBatcher) removeCompleted() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	active := make([]*Request, 0)
	for _, req := range cb.ActiveRequests {
		if !req.Completed {
			active = append(active, req)
		}
	}
	cb.ActiveRequests = active
}

func (cb *ContinuousBatcher) fillFreeSlotsFromQueue() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	maxBatchSize := 32
	for len(cb.ActiveRequests) < maxBatchSize && len(cb.Queue) > 0 {
		req := cb.Queue[0]
		cb.Queue = cb.Queue[1:]
		cb.ActiveRequests = append(cb.ActiveRequests, req)
	}
}

func (cb *ContinuousBatcher) executeOneBatch() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
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
