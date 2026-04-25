package probe

import (
	"testing"
	"time"
)

func makeHistory(results []Result) *History {
	h := NewHistory(len(results))
	for _, r := range results {
		h.Add(r)
	}
	return h
}

func okResult(latency time.Duration) Result {
	return Result{Status: StatusServing, Latency: latency}
}

func errResult() Result {
	return Result{Status: StatusNotServing, Err: errDummy}
}

var errDummy = func() error {
	type e struct{ s string }
	return &e{s: "dummy error"}
}()

func (e *struct{ s string }) Error() string { return e.s }

func TestEvaluate_NoAlert(t *testing.T) {
	h := makeHistory([]Result{okResult(50 * time.Millisecond), okResult(60 * time.Millisecond)})
	p := DefaultAlertPolicy()
	alert := p.Evaluate("svc", h)
	if alert.Level != AlertNone {
		t.Fatalf("expected no alert, got level %d: %s", alert.Level, alert.Message)
	}
}

func TestEvaluate_WarningErrorRate(t *testing.T) {
	results := []Result{okResult(10 * time.Millisecond), okResult(10 * time.Millisecond), errResult(), errResult()}
	h := makeHistory(results)
	p := DefaultAlertPolicy()
	alert := p.Evaluate("svc", h)
	if alert.Level != AlertWarning {
		t.Fatalf("expected warning, got level %d", alert.Level)
	}
}

func TestEvaluate_CriticalErrorRate(t *testing.T) {
	results := []Result{errResult(), errResult(), errResult(), errResult(), errResult(), okResult(10 * time.Millisecond)}
	h := makeHistory(results)
	p := DefaultAlertPolicy()
	alert := p.Evaluate("svc", h)
	if alert.Level != AlertCritical {
		t.Fatalf("expected critical, got level %d", alert.Level)
	}
}

func TestEvaluate_WarningLatency(t *testing.T) {
	h := makeHistory([]Result{okResult(600 * time.Millisecond), okResult(700 * time.Millisecond)})
	p := DefaultAlertPolicy()
	alert := p.Evaluate("svc", h)
	if alert.Level != AlertWarning {
		t.Fatalf("expected warning latency alert, got level %d", alert.Level)
	}
}

func TestEvaluate_CriticalLatency(t *testing.T) {
	h := makeHistory([]Result{okResult(2500 * time.Millisecond), okResult(3000 * time.Millisecond)})
	p := DefaultAlertPolicy()
	alert := p.Evaluate("svc", h)
	if alert.Level != AlertCritical {
		t.Fatalf("expected critical latency alert, got level %d", alert.Level)
	}
}

func TestDefaultAlertPolicy_Values(t *testing.T) {
	p := DefaultAlertPolicy()
	if p.WarningErrorRate >= p.CriticalErrorRate {
		t.Error("warning threshold should be less than critical")
	}
	if p.LatencyWarningMs >= p.LatencyCriticalMs {
		t.Error("latency warning should be less than critical")
	}
}

// TestEvaluate_EmptyHistory verifies that evaluating an empty history
// produces no alert rather than panicking or returning a spurious result.
func TestEvaluate_EmptyHistory(t *testing.T) {
	h := makeHistory([]Result{})
	p := DefaultAlertPolicy()
	alert := p.Evaluate("svc", h)
	if alert.Level != AlertNone {
		t.Fatalf("expected no alert for empty history, got level %d: %s", alert.Level, alert.Message)
	}
}
