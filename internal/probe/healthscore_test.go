package probe

import (
	"testing"
	"time"
)

func buildScorerHistory(tracker *UptimeTracker, target string, healthy, total int) {
	for i := 0; i < total; i++ {
		serving := i < healthy
		status := StatusServing
		if !serving {
			status = StatusNotServing
		}
		tracker.Record(target, Result{Target: target, Status: status, Latency: 10 * time.Millisecond})
	}
}

func TestHealthScorer_PerfectHealth(t *testing.T) {
	uptime := NewUptimeTracker()
	window := NewWindowAggregator(1*time.Minute, 100)

	buildScorerHistory(uptime, "svc", 10, 10)
	for i := 0; i < 10; i++ {
		window.Record(Result{Target: "svc", Status: StatusServing, Latency: 5 * time.Millisecond})
	}

	scorer := NewHealthScorer(uptime, window)
	got := scorer.Compute("svc")

	if got.Score < 0.99 {
		t.Errorf("expected score ~1.0, got %.4f", got.Score)
	}
	if got.Target != "svc" {
		t.Errorf("unexpected target: %s", got.Target)
	}
}

func TestHealthScorer_AllUnhealthy(t *testing.T) {
	uptime := NewUptimeTracker()
	window := NewWindowAggregator(1*time.Minute, 100)

	buildScorerHistory(uptime, "svc", 0, 10)
	for i := 0; i < 10; i++ {
		window.Record(Result{Target: "svc", Status: StatusNotServing, Err: errSLAResult("down").Err})
	}

	scorer := NewHealthScorer(uptime, window)
	got := scorer.Compute("svc")

	if got.Score > 0.01 {
		t.Errorf("expected score ~0.0, got %.4f", got.Score)
	}
}

func TestHealthScorer_PartialHealth(t *testing.T) {
	uptime := NewUptimeTracker()
	window := NewWindowAggregator(1*time.Minute, 100)

	buildScorerHistory(uptime, "svc", 5, 10) // 50% uptime
	for i := 0; i < 10; i++ {
		status := StatusServing
		if i%2 == 0 {
			status = StatusNotServing
		}
		window.Record(Result{Target: "svc", Status: status})
	}

	scorer := NewHealthScorer(uptime, window)
	got := scorer.Compute("svc")

	if got.Score < 0.4 || got.Score > 0.6 {
		t.Errorf("expected score ~0.5, got %.4f", got.Score)
	}
}

func TestHealthScorer_UnknownTarget(t *testing.T) {
	uptime := NewUptimeTracker()
	window := NewWindowAggregator(1*time.Minute, 100)
	scorer := NewHealthScorer(uptime, window)

	got := scorer.Compute("ghost")
	// No data: uptime fraction = 0, error rate = 0 → score = 0*0.5 + 1.0*0.5 = 0.5
	if got.Score < 0.49 || got.Score > 0.51 {
		t.Errorf("expected score 0.5 for unknown target, got %.4f", got.Score)
	}
}

func TestHealthScorer_All(t *testing.T) {
	uptime := NewUptimeTracker()
	window := NewWindowAggregator(1*time.Minute, 100)

	buildScorerHistory(uptime, "a", 10, 10)
	buildScorerHistory(uptime, "b", 10, 10)

	scorer := NewHealthScorer(uptime, window)
	all := scorer.All()

	if len(all) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(all))
	}
	for _, s := range all {
		if s.ComputedAt.IsZero() {
			t.Errorf("ComputedAt should not be zero for %s", s.Target)
		}
	}
}
