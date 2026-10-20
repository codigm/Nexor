package circuit_breaker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type State string

const (
	StateClosed   State = "CLOSED"
	StateOpen     State = "OPEN"
	StateHalfOpen State = "HALF_OPEN"
)

type CircuitBreaker struct {
	rdb *redis.Client
}

func NewCircuitBreaker(rdb *redis.Client) *CircuitBreaker {
	return &CircuitBreaker{rdb: rdb}
}

func (cb *CircuitBreaker) AllowRequest(ctx context.Context, bankID string) (bool, error) {
	state, err := cb.GetState(ctx, bankID)
	if err != nil {
		return false, err
	}

	if state == StateOpen {
		return false, nil
	}
	return true, nil
}

func (cb *CircuitBreaker) GetState(ctx context.Context, bankID string) (State, error) {
	stateKey := fmt.Sprintf("cb:%s:state", bankID)
	
	val, err := cb.rdb.Get(ctx, stateKey).Result()
	if err == redis.Nil {
		return StateClosed, nil
	} else if err != nil {
		return "", err
	}

	state := State(val)
	if state == StateOpen {
		// Check if timeout has expired
		timeoutKey := fmt.Sprintf("cb:%s:timeout", bankID)
		exists, _ := cb.rdb.Exists(ctx, timeoutKey).Result()
		if exists == 0 {
			// Timeout expired, transition to HALF_OPEN
			cb.rdb.Set(ctx, stateKey, string(StateHalfOpen), 0)
			log.Printf("[CB] %s transitioned to HALF_OPEN", bankID)
			return StateHalfOpen, nil
		}
	}

	return state, nil
}

func (cb *CircuitBreaker) RecordSuccess(ctx context.Context, bankID string) error {
	state, _ := cb.GetState(ctx, bankID)
	if state == StateHalfOpen {
		log.Printf("[CB] %s transitioned to CLOSED", bankID)
		cb.rdb.Set(ctx, fmt.Sprintf("cb:%s:state", bankID), string(StateClosed), 0)
	}
	cb.rdb.Del(ctx, fmt.Sprintf("cb:%s:failures", bankID))
	return nil
}

func (cb *CircuitBreaker) RecordFailure(ctx context.Context, bankID string) error {
	state, _ := cb.GetState(ctx, bankID)
	
	if state == StateHalfOpen {
		log.Printf("[CB] %s transitioned to OPEN", bankID)
		cb.transitionToOpen(ctx, bankID)
		return nil
	}

	if state == StateClosed || state == "" {
		failKey := fmt.Sprintf("cb:%s:failures", bankID)
		failures, _ := cb.rdb.Incr(ctx, failKey).Result()
		if failures >= 3 {
			log.Printf("[CB] %s transitioned to OPEN (too many failures)", bankID)
			cb.transitionToOpen(ctx, bankID)
		}
	}
	return nil
}

func (cb *CircuitBreaker) transitionToOpen(ctx context.Context, bankID string) {
	stateKey := fmt.Sprintf("cb:%s:state", bankID)
	timeoutKey := fmt.Sprintf("cb:%s:timeout", bankID)
	
	cb.rdb.Set(ctx, stateKey, string(StateOpen), 0)
	cb.rdb.Set(ctx, timeoutKey, "1", 30*time.Second)
}

// QueuePayment adds a transaction to the bank's recovery queue when the breaker is OPEN.
func (cb *CircuitBreaker) QueuePayment(ctx context.Context, bankID string, payload string) error {
	log.Printf("[CIRCUIT_BREAKER] Queueing payment for bank %s (State: OPEN)", bankID)
	key := fmt.Sprintf("cb:%s:queue", bankID)
	return cb.rdb.LPush(ctx, key, payload).Err()
}

// DrainQueue (Run in background loop) processes queued payments when in HALF_OPEN state.
func (cb *CircuitBreaker) DrainQueue(ctx context.Context, bankID string, processFunc func(payload string) error) {
	key := fmt.Sprintf("cb:%s:queue", bankID)
	for {
		state, _ := cb.GetState(ctx, bankID)
		if state == StateOpen {
			time.Sleep(5 * time.Second)
			continue
		}

		// Pop one transaction
		payload, err := cb.rdb.RPop(ctx, key).Result()
		if err == redis.Nil {
			// Queue empty
			if state == StateHalfOpen {
				log.Printf("[CIRCUIT_BREAKER] Queue empty for %s, transitioning HALF_OPEN -> CLOSED", bankID)
				cb.rdb.Set(ctx, fmt.Sprintf("cb:%s:state", bankID), string(StateClosed), 0)
			}
			time.Sleep(5 * time.Second)
			continue
		} else if err != nil {
			log.Printf("[CIRCUIT_BREAKER] Error popping queue for %s: %v", bankID, err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Process it
		log.Printf("[CIRCUIT_BREAKER] Draining queued payment for %s", bankID)
		err = processFunc(payload)
		if err != nil {
			log.Printf("[CIRCUIT_BREAKER] Failed to process payment for %s, requeueing. Transitioning to OPEN.", bankID)
			cb.rdb.LPush(ctx, key, payload) // Requeue
			cb.transitionToOpen(ctx, bankID)
		} else {
			cb.RecordSuccess(ctx, bankID)
		}

		// Throttle draining in HALF_OPEN
		if state == StateHalfOpen {
			time.Sleep(500 * time.Millisecond)
		}
	}
}
