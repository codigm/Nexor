package saga

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestReserveFunds_And_Compensate(t *testing.T) {
	// Start miniredis for a local mock redis server
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when starting miniredis", err)
	}
	defer mr.Close()

	// Configure redis client pointing to miniredis
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	ctx := context.Background()
	sCtx := SAGAContext{
		TransactionID:   "tx-123",
		SenderAccount:   "ALPH001",
		ReceiverAccount: "BETA001",
		Amount:          "1500.00",
		Currency:        "INR",
	}

	// 1. Reserve Funds
	err = ReserveFunds(ctx, client, sCtx)
	if err != nil {
		t.Fatalf("expected no error on ReserveFunds, got %v", err)
	}

	// Verify key exists in miniredis
	key := fmt.Sprintf("reserve:%s:%s", sCtx.SenderAccount, sCtx.TransactionID)
	val, err := mr.Get(key)
	if err != nil {
		t.Fatalf("expected key to exist in redis, got error %v", err)
	}
	if val != "1500.00" {
		t.Errorf("expected reserved amount '1500.00', got %v", val)
	}

	// Simulate Argus FraudCheck Failure by writing to redis
	err = client.Set(ctx, "fraud:"+sCtx.TransactionID, "BLOCK", 0).Err()
	if err != nil {
		t.Fatalf("failed to write mock fraud score to redis: %v", err)
	}

	err = FraudCheck(ctx, client, sCtx)
	if err != nil {
		// 2. Compensate Reserve Funds upon failure
		err = CompensateReserveFunds(ctx, client, sCtx)
		if err != nil {
			t.Fatalf("expected no error on CompensateReserveFunds, got %v", err)
		}
	} else {
		t.Fatalf("expected FraudCheck to return an error for BLOCK, but got nil")
	}

	// Verify key is deleted successfully
	_, err = mr.Get(key)
	if err == nil {
		t.Errorf("expected key to be deleted, but it still exists")
	}
}
