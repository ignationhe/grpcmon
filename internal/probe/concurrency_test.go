package probe

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestConcurrencyTracker_InflightZeroInitially(t *testing.T) {
	ct := NewConcurrencyTracker(time.Minute)
	if got := ct.Inflight("svc:443"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestConcurrencyTracker_AcquireIncrements(t *testing.T) {
	ct := NewConcurrencyTracker(time.Minute)
	_ = ct.Acquire("svc:443")
	_ = ct.Acquire("svc:443")
	if got := ct.Inflight("svc:443"); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestConcurrencyTracker_ReleaseDecrements(t *testing.T) {
	ct := NewConcurrencyTracker(time.Minute)
	_ = ct.Acquire("svc:443")
	_ = ct.Acquire("svc:443")
	ct.Release("svc:443")
	if got := ct.Inflight("svc:443"); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}

func TestConcurrencyTracker_ReleaseDoesNotGoBelowZero(t *testing.T) {
	ct := NewConcurrencyTracker(time.Minute)
	ct.Release("svc:443") // no prior acquire
	if got := ct.Inflight("svc:443"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestConcurrencyTracker_PeakTracked(t *testing.T) {
	ct := NewConcurrencyTracker(time.Minute)
	_ = ct.Acquire("svc:443")
	_ = ct.Acquire("svc:443")
	_ = ct.Acquire("svc:443")
	ct.Release("svc:443")
	ct.Release("svc:443")
	ct.Release("svc:443")
	if got := ct.Peak("svc:443"); got != 3 {
		t.Fatalf("expected peak 3, got %d", got)
	}
}

func TestConcurrencyTracker_AcquireEmptyTargetError(t *testing.T) {
	ct := NewConcurrencyTracker(time.Minute)
	if err := ct.Acquire(""); err == nil {
		t.Fatal("expected error for empty target")
	}
}

func TestConcurrencyTracker_DifferentTargetsIndependent(t *testing.T) {
	ct := NewConcurrencyTracker(time.Minute)
	_ = ct.Acquire("a:443")
	_ = ct.Acquire("a:443")
	_ = ct.Acquire("b:443")
	if got := ct.Inflight("a:443"); got != 2 {
		t.Fatalf("a: expected 2, got %d", got)
	}
	if got := ct.Inflight("b:443"); got != 1 {
		t.Fatalf("b: expected 1, got %d", got)
	}
}

func TestConcurrencyTracker_Track_RunsFn(t *testing.T) {
	ct := NewConcurrencyTracker(time.Minute)
	ran := false
	err := ct.Track(context.Background(), "svc:443", func(_ context.Context) error {
		ran = true
		if got := ct.Inflight("svc:443"); got != 1 {
			t.Errorf("expected inflight 1 during fn, got %d", got)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ran {
		t.Fatal("fn was not called")
	}
	if got := ct.Inflight("svc:443"); got != 0 {
		t.Fatalf("expected inflight 0 after Track, got %d", got)
	}
}

func TestConcurrencyTracker_ConcurrentAcquire(t *testing.T) {
	ct := NewConcurrencyTracker(time.Minute)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = ct.Acquire("svc:443")
		}()
	}
	wg.Wait()
	if got := ct.Inflight("svc:443"); got != 50 {
		t.Fatalf("expected 50, got %d", got)
	}
}
