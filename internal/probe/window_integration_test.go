package probe_test

import (
	"errors"
	"testing"
	"time"

	"github.com/yourorg/grpcmon/internal/probe"
)

func TestWindowAggregator_IntegrationWithScheduler(t *testing.T) {
	agg := probe.NewWindowAggregator(time.Minute)
	now := time.Now()

	// Simulate a mix of successful and failed probes for two targets.
	for i := 0; i < 5; i++ {
		agg.Record("alpha:443", probe.Result{
			Target:    "alpha:443",
			Timestamp: now,
			Latency:   time.Duration(10+i*5) * time.Millisecond,
		})
	}
	for i := 0; i < 3; i++ {
		agg.Record("beta:443", probe.Result{
			Target:    "beta:443",
			Timestamp: now,
			Latency:   time.Duration(5+i*2) * time.Millisecond,
			Err:       errors.New("unavailable"),
		})
	}
	agg.Record("beta:443", probe.Result{
		Target:    "beta:443",
		Timestamp: now,
		Latency:   8 * time.Millisecond,
	})

	alpha := agg.Stats("alpha:443")
	if alpha.Count != 5 {
		t.Fatalf("alpha: expected 5 results, got %d", alpha.Count)
	}
	if alpha.ErrorRate != 0 {
		t.Fatalf("alpha: expected 0 error rate, got %f", alpha.ErrorRate)
	}
	if alpha.AvgLatency != 20*time.Millisecond {
		t.Fatalf("alpha: expected avg 20ms, got %v", alpha.AvgLatency)
	}

	beta := agg.Stats("beta:443")
	if beta.Count != 4 {
		t.Fatalf("beta: expected 4 results, got %d", beta.Count)
	}
	if beta.ErrorRate != 0.75 {
		t.Fatalf("beta: expected error rate 0.75, got %f", beta.ErrorRate)
	}

	targets := agg.Targets()
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
}

func TestWindowAggregator_AllEvictedReturnsZero(t *testing.T) {
	agg := probe.NewWindowAggregator(1 * time.Millisecond)
	old := time.Now().Add(-10 * time.Millisecond)
	agg.Record("svc:443", probe.Result{
		Target:    "svc:443",
		Timestamp: old,
		Latency:   5 * time.Millisecond,
	})
	time.Sleep(5 * time.Millisecond)
	s := agg.Stats("svc:443")
	if s.Count != 0 {
		t.Fatalf("expected 0 after full eviction, got %d", s.Count)
	}
	if s.AvgLatency != 0 {
		t.Fatalf("expected zero avg latency, got %v", s.AvgLatency)
	}
}
