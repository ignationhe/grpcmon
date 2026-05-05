package probe

import (
	"errors"
	"testing"
	"time"
)

func buildWindowResult(ts time.Time, latency time.Duration, err error) Result {
	return Result{
		Target:    "host:443",
		Timestamp: ts,
		Latency:   latency,
		Err:       err,
	}
}

func TestWindowAggregator_EmptyStats(t *testing.T) {
	w := NewWindowAggregator(time.Minute)
	s := w.Stats("host:443")
	if s.Count != 0 {
		t.Fatalf("expected 0 count, got %d", s.Count)
	}
	if s.ErrorRate != 0 {
		t.Fatalf("expected 0 error rate, got %f", s.ErrorRate)
	}
}

func TestWindowAggregator_CountsResults(t *testing.T) {
	w := NewWindowAggregator(time.Minute)
	now := time.Now()
	w.Record("host:443", buildWindowResult(now, 10*time.Millisecond, nil))
	w.Record("host:443", buildWindowResult(now, 20*time.Millisecond, nil))
	s := w.Stats("host:443")
	if s.Count != 2 {
		t.Fatalf("expected count 2, got %d", s.Count)
	}
}

func TestWindowAggregator_EvictsOldEntries(t *testing.T) {
	w := NewWindowAggregator(50 * time.Millisecond)
	old := time.Now().Add(-100 * time.Millisecond)
	w.Record("host:443", buildWindowResult(old, 10*time.Millisecond, nil))
	w.Record("host:443", buildWindowResult(time.Now(), 20*time.Millisecond, nil))
	s := w.Stats("host:443")
	if s.Count != 1 {
		t.Fatalf("expected 1 after eviction, got %d", s.Count)
	}
}

func TestWindowAggregator_ErrorRate(t *testing.T) {
	w := NewWindowAggregator(time.Minute)
	now := time.Now()
	w.Record("host:443", buildWindowResult(now, 10*time.Millisecond, nil))
	w.Record("host:443", buildWindowResult(now, 10*time.Millisecond, errors.New("fail")))
	s := w.Stats("host:443")
	if s.ErrorRate != 0.5 {
		t.Fatalf("expected error rate 0.5, got %f", s.ErrorRate)
	}
}

func TestWindowAggregator_AvgLatency(t *testing.T) {
	w := NewWindowAggregator(time.Minute)
	now := time.Now()
	w.Record("host:443", buildWindowResult(now, 10*time.Millisecond, nil))
	w.Record("host:443", buildWindowResult(now, 30*time.Millisecond, nil))
	s := w.Stats("host:443")
	if s.AvgLatency != 20*time.Millisecond {
		t.Fatalf("expected avg 20ms, got %v", s.AvgLatency)
	}
}

func TestWindowAggregator_P95Latency(t *testing.T) {
	w := NewWindowAggregator(time.Minute)
	now := time.Now()
	for i := 1; i <= 20; i++ {
		w.Record("host:443", buildWindowResult(now, time.Duration(i)*time.Millisecond, nil))
	}
	s := w.Stats("host:443")
	if s.P95Latency < 18*time.Millisecond {
		t.Fatalf("expected p95 >= 18ms, got %v", s.P95Latency)
	}
}

func TestWindowAggregator_Targets(t *testing.T) {
	w := NewWindowAggregator(time.Minute)
	now := time.Now()
	w.Record("a:443", buildWindowResult(now, 5*time.Millisecond, nil))
	w.Record("b:443", buildWindowResult(now, 5*time.Millisecond, nil))
	targets := w.Targets()
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
}
