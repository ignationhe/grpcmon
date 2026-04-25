package probe

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestWithRetry_SuccessOnFirstAttempt(t *testing.T) {
	var calls int32
	fn := func(_ context.Context) Result {
		atomic.AddInt32(&calls, 1)
		return Result{Target: "svc", Latency: 5 * time.Millisecond}
	}

	policy := RetryPolicy{MaxAttempts: 3, Delay: 10 * time.Millisecond}
	res := WithRetry(context.Background(), policy, fn)

	if res.Error != nil {
		t.Fatalf("expected no error, got %v", res.Error)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestWithRetry_RetriesOnFailure(t *testing.T) {
	var calls int32
	errFail := errors.New("probe failed")
	fn := func(_ context.Context) Result {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			return Result{Target: "svc", Error: errFail}
		}
		return Result{Target: "svc", Latency: 10 * time.Millisecond}
	}

	policy := RetryPolicy{MaxAttempts: 3, Delay: 5 * time.Millisecond}
	res := WithRetry(context.Background(), policy, fn)

	if res.Error != nil {
		t.Fatalf("expected success on 3rd attempt, got error: %v", res.Error)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_AllAttemptsFail(t *testing.T) {
	errFail := errors.New("always fails")
	var calls int32
	fn := func(_ context.Context) Result {
		atomic.AddInt32(&calls, 1)
		return Result{Target: "svc", Error: errFail}
	}

	policy := RetryPolicy{MaxAttempts: 3, Delay: 5 * time.Millisecond}
	res := WithRetry(context.Background(), policy, fn)

	if res.Error == nil {
		t.Fatal("expected error after all attempts")
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	var calls int32
	fn := func(_ context.Context) Result {
		atomic.AddInt32(&calls, 1)
		return Result{Target: "svc", Error: errors.New("fail")}
	}

	policy := RetryPolicy{MaxAttempts: 5, Delay: 50 * time.Millisecond}
	res := WithRetry(ctx, policy, fn)

	if res.Error == nil {
		t.Fatal("expected error due to cancelled context")
	}
	// Should have run first attempt, then bailed on delay for second.
	if calls > 2 {
		t.Errorf("expected at most 2 calls with cancelled context, got %d", calls)
	}
}

func TestDefaultRetryPolicy(t *testing.T) {
	p := DefaultRetryPolicy()
	if p.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", p.MaxAttempts)
	}
	if p.Delay != 100*time.Millisecond {
		t.Errorf("expected Delay=100ms, got %v", p.Delay)
	}
}
