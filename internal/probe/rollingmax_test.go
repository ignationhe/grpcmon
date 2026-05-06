package probe

import (
	"testing"
	"time"
)

func TestRollingMax_NoSamples(t *testing.T) {
	rm := NewRollingMax(time.Minute)
	_, ok := rm.Max("svc")
	if ok {
		t.Fatal("expected false for empty target")
	}
}

func TestRollingMax_SingleSample(t *testing.T) {
	rm := NewRollingMax(time.Minute)
	rm.Record("svc", 42*time.Millisecond)
	max, ok := rm.Max("svc")
	if !ok {
		t.Fatal("expected ok")
	}
	if max != 42*time.Millisecond {
		t.Fatalf("expected 42ms, got %v", max)
	}
}

func TestRollingMax_ReturnsLargest(t *testing.T) {
	rm := NewRollingMax(time.Minute)
	for _, d := range []time.Duration{10, 80, 30, 55} {
		rm.Record("svc", d*time.Millisecond)
	}
	max, ok := rm.Max("svc")
	if !ok {
		t.Fatal("expected ok")
	}
	if max != 80*time.Millisecond {
		t.Fatalf("expected 80ms, got %v", max)
	}
}

func TestRollingMax_Eviction(t *testing.T) {
	rm := NewRollingMax(50 * time.Millisecond)
	// Inject an old entry directly to simulate expiry.
	rm.mu.Lock()
	rm.entries["svc"] = []rollingMaxEntry{
		{latency: 999 * time.Millisecond, at: time.Now().Add(-100 * time.Millisecond)},
	}
	rm.mu.Unlock()

	rm.Record("svc", 5*time.Millisecond)
	max, ok := rm.Max("svc")
	if !ok {
		t.Fatal("expected ok after adding fresh sample")
	}
	if max != 5*time.Millisecond {
		t.Fatalf("expected 5ms after eviction, got %v", max)
	}
}

func TestRollingMax_DifferentTargetsIndependent(t *testing.T) {
	rm := NewRollingMax(time.Minute)
	rm.Record("a", 100*time.Millisecond)
	rm.Record("b", 200*time.Millisecond)

	maxA, _ := rm.Max("a")
	maxB, _ := rm.Max("b")
	if maxA != 100*time.Millisecond {
		t.Fatalf("a: expected 100ms, got %v", maxA)
	}
	if maxB != 200*time.Millisecond {
		t.Fatalf("b: expected 200ms, got %v", maxB)
	}
}

func TestRollingMax_Targets(t *testing.T) {
	rm := NewRollingMax(time.Minute)
	rm.Record("x", time.Millisecond)
	rm.Record("y", time.Millisecond)

	targets := rm.Targets()
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
}

func TestRollingMax_ZeroWindowDefaultsToMinute(t *testing.T) {
	rm := NewRollingMax(0)
	if rm.window != time.Minute {
		t.Fatalf("expected default window of 1m, got %v", rm.window)
	}
}
