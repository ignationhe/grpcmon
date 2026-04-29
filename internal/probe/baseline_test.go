package probe

import (
	"testing"
	"time"
)

func buildBaseline(latencies []time.Duration) *History {
	h := NewHistory(256)
	for _, l := range latencies {
		h.Add(Result{Latency: l, Status: "SERVING"})
	}
	return h
}

func TestBaselineStore_GetMissing(t *testing.T) {
	bs := NewBaselineStore()
	_, ok := bs.Get("missing:443")
	if ok {
		t.Fatal("expected missing baseline to return false")
	}
}

func TestBaselineStore_ComputeAndGet(t *testing.T) {
	h := buildBaseline([]time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
	})

	bs := NewBaselineStore()
	bs.Compute("svc:443", h)

	e, ok := bs.Get("svc:443")
	if !ok {
		t.Fatal("expected baseline to exist after Compute")
	}
	if e.SampleN != 5 {
		t.Errorf("expected SampleN=5, got %d", e.SampleN)
	}
	if e.P50 == 0 {
		t.Error("expected non-zero P50")
	}
	if e.P95 == 0 {
		t.Error("expected non-zero P95")
	}
	if e.P99 == 0 {
		t.Error("expected non-zero P99")
	}
	if e.P50 > e.P95 {
		t.Errorf("P50 (%v) should be <= P95 (%v)", e.P50, e.P95)
	}
}

func TestBaselineStore_SkipsErrorResults(t *testing.T) {
	h := NewHistory(256)
	h.Add(Result{Err: errSentinel, Status: "NOT_SERVING"})
	h.Add(Result{Err: errSentinel, Status: "NOT_SERVING"})

	bs := NewBaselineStore()
	bs.Compute("svc:443", h)

	_, ok := bs.Get("svc:443")
	if ok {
		t.Error("expected no baseline when all results have errors")
	}
}

func TestBaselineStore_All(t *testing.T) {
	bs := NewBaselineStore()
	bs.Compute("a:443", buildBaseline([]time.Duration{5 * time.Millisecond, 10 * time.Millisecond}))
	bs.Compute("b:443", buildBaseline([]time.Duration{15 * time.Millisecond, 20 * time.Millisecond}))

	all := bs.All()
	if len(all) != 2 {
		t.Errorf("expected 2 baselines, got %d", len(all))
	}
}

func TestBaselineStore_ComputeEmptyHistory(t *testing.T) {
	h := NewHistory(256)
	bs := NewBaselineStore()
	bs.Compute("svc:443", h)

	_, ok := bs.Get("svc:443")
	if ok {
		t.Error("expected no baseline for empty history")
	}
}

var errSentinel = &probeError{msg: "connection refused"}

type probeError struct{ msg string }

func (e *probeError) Error() string { return e.msg }
