package probe

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDefaultTimeoutPolicy(t *testing.T) {
	p := DefaultTimeoutPolicy()
	if p.Default != 5*time.Second {
		t.Fatalf("expected 5s default, got %s", p.Default)
	}
}

func TestTimeoutPolicy_For_UsesDefault(t *testing.T) {
	p := DefaultTimeoutPolicy()
	if got := p.For("localhost:50051"); got != 5*time.Second {
		t.Fatalf("expected 5s, got %s", got)
	}
}

func TestTimeoutPolicy_For_UsesOverride(t *testing.T) {
	p := DefaultTimeoutPolicy()
	p.Overrides["localhost:9090"] = 2 * time.Second
	if got := p.For("localhost:9090"); got != 2*time.Second {
		t.Fatalf("expected 2s override, got %s", got)
	}
}

func TestTimeoutPolicy_For_ZeroOverrideFallsBack(t *testing.T) {
	p := DefaultTimeoutPolicy()
	p.Overrides["localhost:9090"] = 0
	if got := p.For("localhost:9090"); got != 5*time.Second {
		t.Fatalf("expected fallback 5s, got %s", got)
	}
}

func TestWithTimeout_CompletesBeforeDeadline(t *testing.T) {
	policy := DefaultTimeoutPolicy()
	addr := "localhost:50051"

	fn := WithTimeout(policy, addr, func(ctx context.Context) Result {
		return Result{Address: addr, Latency: 1 * time.Millisecond}
	})

	r := fn(context.Background())
	if r.Err != nil {
		t.Fatalf("unexpected error: %v", r.Err)
	}
}

func TestWithTimeout_ExceedsDeadline(t *testing.T) {
	policy := &TimeoutPolicy{
		Default:   20 * time.Millisecond,
		Overrides: make(map[string]time.Duration),
	}
	addr := "localhost:50051"

	fn := WithTimeout(policy, addr, func(ctx context.Context) Result {
		select {
		case <-ctx.Done():
			return Result{Address: addr, Err: ctx.Err()}
		case <-time.After(500 * time.Millisecond):
			return Result{Address: addr}
		}
	})

	r := fn(context.Background())
	if r.Err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(r.Err.Error(), "timed out") {
		t.Fatalf("expected 'timed out' in error, got: %v", r.Err)
	}
	if r.Latency != 20*time.Millisecond {
		t.Fatalf("expected latency=timeout, got %s", r.Latency)
	}
}

func TestWithTimeout_ParentCancellation(t *testing.T) {
	policy := DefaultTimeoutPolicy()
	addr := "localhost:50051"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fn := WithTimeout(policy, addr, func(ctx context.Context) Result {
		<-ctx.Done()
		return Result{Address: addr, Err: ctx.Err()}
	})

	r := fn(ctx)
	if r.Err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
