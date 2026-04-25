package probe

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter_FirstCallImmediate(t *testing.T) {
	rl := NewRateLimiter(100 * time.Millisecond)
	start := time.Now()
	if err := rl.Wait(context.Background(), "localhost:50051"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 20*time.Millisecond {
		t.Errorf("first call should be immediate, took %v", elapsed)
	}
}

func TestRateLimiter_SecondCallEnforcesDelay(t *testing.T) {
	rl := NewRateLimiter(80 * time.Millisecond)
	addr := "localhost:50051"

	_ = rl.Wait(context.Background(), addr)
	start := time.Now()
	_ = rl.Wait(context.Background(), addr)
	elapsed := time.Since(start)

	if elapsed < 70*time.Millisecond {
		t.Errorf("expected delay >= 70ms, got %v", elapsed)
	}
}

func TestRateLimiter_DifferentAddressesIndependent(t *testing.T) {
	rl := NewRateLimiter(200 * time.Millisecond)
	_ = rl.Wait(context.Background(), "a:1")

	start := time.Now()
	_ = rl.Wait(context.Background(), "b:2")
	if elapsed := time.Since(start); elapsed > 20*time.Millisecond {
		t.Errorf("different addresses should not share rate limit, took %v", elapsed)
	}
}

func TestRateLimiter_ContextCancellation(t *testing.T) {
	rl := NewRateLimiter(500 * time.Millisecond)
	addr := "localhost:50051"
	_ = rl.Wait(context.Background(), addr)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err := rl.Wait(ctx, addr)
	if err == nil {
		t.Fatal("expected error on context cancellation, got nil")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(500 * time.Millisecond)
	addr := "localhost:50051"
	_ = rl.Wait(context.Background(), addr)

	rl.Reset(addr)

	start := time.Now()
	_ = rl.Wait(context.Background(), addr)
	if elapsed := time.Since(start); elapsed > 20*time.Millisecond {
		t.Errorf("after Reset, call should be immediate, took %v", elapsed)
	}
}

func TestRateLimiter_LastProbe(t *testing.T) {
	rl := NewRateLimiter(100 * time.Millisecond)
	addr := "localhost:50051"

	_, ok := rl.LastProbe(addr)
	if ok {
		t.Fatal("expected no last probe before any call")
	}

	before := time.Now()
	_ = rl.Wait(context.Background(), addr)
	after := time.Now()

	ts, ok := rl.LastProbe(addr)
	if !ok {
		t.Fatal("expected last probe to be recorded")
	}
	if ts.Before(before) || ts.After(after) {
		t.Errorf("last probe time %v out of expected range [%v, %v]", ts, before, after)
	}
}
