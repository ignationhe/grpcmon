package probe

import (
	"context"
	"errors"
	"testing"
	"time"
)

func successFn(_ context.Context) Result {
	return Result{Status: "SERVING", Latency: 5 * time.Millisecond, Timestamp: time.Now()}
}

func failFn(_ context.Context) Result {
	return Result{Status: "NOT_SERVING", Err: errors.New("down"), Timestamp: time.Now()}
}

func TestProbePipeline_SuccessPassesThrough(t *testing.T) {
	pp := NewProbePipeline(DefaultRetryPolicy(), DefaultTimeoutPolicy())
	r := pp.Run(context.Background(), "host:50051", successFn)
	if r.Status != "SERVING" {
		t.Fatalf("expected SERVING, got %s", r.Status)
	}
}

func TestProbePipeline_ThrottleRejects(t *testing.T) {
	th := NewThrottlePolicy(2, 500*time.Millisecond)
	pp := NewProbePipeline(DefaultRetryPolicy(), DefaultTimeoutPolicy(), WithThrottleOption(th))

	pp.Run(context.Background(), "host:50051", successFn)
	r := pp.Run(context.Background(), "host:50051", successFn)
	if r.Status != "THROTTLED" {
		t.Fatalf("expected THROTTLED, got %s", r.Status)
	}
}

func TestProbePipeline_CircuitBreakerOpens(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Second)
	pp := NewProbePipeline(DefaultRetryPolicy(), DefaultTimeoutPolicy(), WithCircuitBreaker(cb))

	pp.Run(context.Background(), "host:50051", failFn)

	r := pp.Run(context.Background(), "host:50051", successFn)
	if r.Status != "OPEN" {
		t.Fatalf("expected OPEN circuit, got %s", r.Status)
	}
}

func TestProbePipeline_CircuitBreakerRecordsSuccess(t *testing.T) {
	cb := NewCircuitBreaker(5, 10*time.Second)
	pp := NewProbePipeline(DefaultRetryPolicy(), DefaultTimeoutPolicy(), WithCircuitBreaker(cb))

	for i := 0; i < 3; i++ {
		r := pp.Run(context.Background(), "host:50051", successFn)
		if r.Status != "SERVING" {
			t.Fatalf("run %d: expected SERVING, got %s", i, r.Status)
		}
	}
}

func TestProbePipeline_ContextCancelledViaRateLimiter(t *testing.T) {
	rl := NewRateLimiter(1 * time.Second)
	pp := NewProbePipeline(DefaultRetryPolicy(), DefaultTimeoutPolicy(), WithRateLimiterOption(rl))

	pp.Run(context.Background(), "host:50051", successFn)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := pp.Run(ctx, "host:50051", successFn)
	if r.Err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
