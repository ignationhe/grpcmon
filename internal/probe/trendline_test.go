package probe

import (
	"testing"
	"time"
)

func buildTrendHistory(latencies []time.Duration) *History {
	h := NewHistory(100)
	base := time.Now().Add(-time.Duration(len(latencies)) * time.Second)
	for i, l := range latencies {
		h.Add(Result{
			Target:  "svc",
			At:      base.Add(time.Duration(i) * time.Second),
			Latency: l,
		})
	}
	return h
}

func TestTrendAnalyzer_StableWhenInsufficientData(t *testing.T) {
	a := NewTrendAnalyzer(5)
	h := buildTrendHistory([]time.Duration{10 * time.Millisecond, 20 * time.Millisecond})
	r := a.Analyze("svc", h)
	if r.Direction != TrendStable {
		t.Fatalf("expected Stable, got %v", r.Direction)
	}
}

func TestTrendAnalyzer_DetectsDegrading(t *testing.T) {
	a := NewTrendAnalyzer(4)
	latencies := []time.Duration{
		5 * time.Millisecond,
		20 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		200 * time.Millisecond,
	}
	h := buildTrendHistory(latencies)
	r := a.Analyze("svc", h)
	if r.Direction != TrendDegrading {
		t.Fatalf("expected Degrading, got %v (slope=%.2f)", r.Direction, r.Slope)
	}
}

func TestTrendAnalyzer_DetectsImproving(t *testing.T) {
	a := NewTrendAnalyzer(4)
	latencies := []time.Duration{
		200 * time.Millisecond,
		100 * time.Millisecond,
		50 * time.Millisecond,
		20 * time.Millisecond,
		5 * time.Millisecond,
	}
	h := buildTrendHistory(latencies)
	r := a.Analyze("svc", h)
	if r.Direction != TrendImproving {
		t.Fatalf("expected Improving, got %v (slope=%.2f)", r.Direction, r.Slope)
	}
}

func TestTrendAnalyzer_StableOnFlatLatency(t *testing.T) {
	a := NewTrendAnalyzer(3)
	latencies := []time.Duration{
		10 * time.Millisecond,
		10 * time.Millisecond,
		10 * time.Millisecond,
		10 * time.Millisecond,
	}
	h := buildTrendHistory(latencies)
	r := a.Analyze("svc", h)
	if r.Direction != TrendStable {
		t.Fatalf("expected Stable, got %v", r.Direction)
	}
}

func TestTrendAnalyzer_IgnoresErrorResults(t *testing.T) {
	a := NewTrendAnalyzer(5)
	h := NewHistory(100)
	base := time.Now().Add(-5 * time.Second)
	for i := 0; i < 5; i++ {
		h.Add(Result{
			Target: "svc",
			At:     base.Add(time.Duration(i) * time.Second),
			Err:    errProbe,
		})
	}
	r := a.Analyze("svc", h)
	if r.Direction != TrendStable {
		t.Fatalf("expected Stable when all results are errors, got %v", r.Direction)
	}
}

func TestNewTrendAnalyzer_MinSamplesFloor(t *testing.T) {
	a := NewTrendAnalyzer(0)
	if a.minSamples < 2 {
		t.Fatalf("expected minSamples >= 2, got %d", a.minSamples)
	}
}
