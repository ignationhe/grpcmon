package probe

import (
	"testing"
	"time"
)

func buildHistory(results []Result) *History {
	h := NewHistory(50)
	for _, r := range results {
		h.Add(r)
	}
	return h
}

func okSLAResult(latency time.Duration) Result {
	return Result{Status: StatusServing, Latency: latency}
}

func errSLAResult() Result {
	return Result{Status: StatusUnknown, Err: errSentinel}
}

var errSentinel = fmt.Errorf("probe error")

func TestSLAEvaluator_NoViolations(t *testing.T) {
	h := buildHistory([]Result{
		okSLAResult(10 * time.Millisecond),
		okSLAResult(12 * time.Millisecond),
		okSLAResult(11 * time.Millisecond),
	})
	e := NewSLAEvaluator(SLAPolicy{
		MaxErrorRate:  0.1,
		MaxAvgLatency: 100 * time.Millisecond,
		Window:        10,
	}, nil)
	v := e.Evaluate("svc:443", h)
	if len(v) != 0 {
		t.Fatalf("expected no violations, got %+v", v)
	}
}

func TestSLAEvaluator_ErrorRateViolation(t *testing.T) {
	h := buildHistory([]Result{
		errSLAResult(), errSLAResult(), errSLAResult(),
		okSLAResult(5 * time.Millisecond),
	})
	e := NewSLAEvaluator(SLAPolicy{MaxErrorRate: 0.1, Window: 10}, nil)
	v := e.Evaluate("svc:443", h)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].Kind != "error_rate" {
		t.Errorf("expected error_rate violation, got %s", v[0].Kind)
	}
}

func TestSLAEvaluator_LatencyViolation(t *testing.T) {
	h := buildHistory([]Result{
		okSLAResult(200 * time.Millisecond),
		okSLAResult(250 * time.Millisecond),
	})
	e := NewSLAEvaluator(SLAPolicy{MaxAvgLatency: 100 * time.Millisecond, Window: 10}, nil)
	v := e.Evaluate("svc:443", h)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].Kind != "latency" {
		t.Errorf("expected latency violation, got %s", v[0].Kind)
	}
}

func TestSLAEvaluator_PerTargetOverride(t *testing.T) {
	h := buildHistory([]Result{
		okSLAResult(200 * time.Millisecond),
		okSLAResult(200 * time.Millisecond),
	})
	overrides := map[string]SLAPolicy{
		"special:443": {MaxAvgLatency: 500 * time.Millisecond, Window: 10},
	}
	e := NewSLAEvaluator(SLAPolicy{MaxAvgLatency: 50 * time.Millisecond, Window: 10}, overrides)
	// special target uses relaxed policy — no violation
	v := e.Evaluate("special:443", h)
	if len(v) != 0 {
		t.Fatalf("expected no violations for overridden target, got %+v", v)
	}
	// default target uses strict policy — violation expected
	v2 := e.Evaluate("other:443", h)
	if len(v2) != 1 {
		t.Fatalf("expected 1 violation for default target, got %d", len(v2))
	}
}

func TestSLAEvaluator_BothViolations(t *testing.T) {
	h := buildHistory([]Result{
		errSLAResult(), errSLAResult(),
		okSLAResult(300 * time.Millisecond),
	})
	e := NewSLAEvaluator(SLAPolicy{
		MaxErrorRate:  0.1,
		MaxAvgLatency: 50 * time.Millisecond,
		Window:        10,
	}, nil)
	v := e.Evaluate("svc:443", h)
	if len(v) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(v))
	}
}
