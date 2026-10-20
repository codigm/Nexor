package circuit_breaker

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	ctx := context.Background()
	cb := NewCircuitBreaker(client)
	bankID := "TEST_BANK"

	// 1. Initial state should be CLOSED and allow requests
	state, err := cb.GetState(ctx, bankID)
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, state)

	allowed, err := cb.AllowRequest(ctx, bankID)
	assert.NoError(t, err)
	assert.True(t, allowed)

	// 2. Simulate 3 failures to trigger OPEN state
	cb.RecordFailure(ctx, bankID)
	cb.RecordFailure(ctx, bankID)
	cb.RecordFailure(ctx, bankID)

	state, err = cb.GetState(ctx, bankID)
	assert.NoError(t, err)
	assert.Equal(t, StateOpen, state)

	allowed, err = cb.AllowRequest(ctx, bankID)
	assert.NoError(t, err)
	assert.False(t, allowed) // Should block requests

	// 3. Fast-forward miniredis time by 31 seconds to expire the timeout
	mr.FastForward(31 * time.Second)

	// 4. GetState should now transition to HALF_OPEN
	state, err = cb.GetState(ctx, bankID)
	assert.NoError(t, err)
	assert.Equal(t, StateHalfOpen, state)

	// 5. A single success should close it again
	cb.RecordSuccess(ctx, bankID)
	
	state, err = cb.GetState(ctx, bankID)
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, state)
}
