package probe

import (
	"testing"
	"time"
)

func buildEBWindow(target string, total, errors int) *WindowAggregator {
	wa := NewWindowAggregator(30 * time.Second)
	for i := 0; i < total; i++ {
		var r Result
		if i < errors {
			r = Result{Target: target, Err: errSLAResult().Err}
		} else {
			r = Result{Target: target}
		}
		wa.Record(r)
	}
	return wa
}

func TestErrorBudgetTracker_GetMissing(t *testing.T) {
	wa := NewWindowAggregator(30 * time.Second)
	tracker := NewErrorBudgetTracker(wa, 99.9)
	tracker.Evaluate()
	_, ok := tracker.Get("svc:443")
	if ok {
		t.Fatal("expected missing entry")
	}
}

func TestErrorBudgetTracker_NoErrors(t *testing.T) {
	wa := buildEBWindow("svc:443", 10, 0)
	tracker := NewErrorBudgetTracker(wa, 99.9)
	tracker.Evaluate()
	entry, ok := tracker.Get("svc:443")
	if !ok {
		t.Fatal("expected entry")
	}
	if entry.ErrorRate != 0 {
		t.Errorf("expected zero error rate, got %f", entry.ErrorRate)
	}
	if entry.BudgetUsed != 0 {
		t.Errorf("expected zero budget used, got %f", entry.BudgetUsed)
	}
	if entry.Exhausted {
		t.Error("budget should not be exhausted")
	}
}

func TestErrorBudgetTracker_BudgetExhausted(t *testing.T) {
	wa := buildEBWindow("svc:443", 10, 10)
	tracker := NewErrorBudgetTracker(wa, 99.9)
	tracker.Evaluate()
	entry, ok := tracker.Get("svc:443")
	if !ok {
		t.Fatal("expected entry")
	}
	if !entry.Exhausted {
		t.Error("budget should be exhausted")
	}
	if entry.BudgetUsed != 1.0 {
		t.Errorf("expected BudgetUsed=1.0, got %f", entry.BudgetUsed)
	}
}

func TestErrorBudgetTracker_PartialConsumption(t *testing.T) {
	wa := buildEBWindow("svc:443", 1000, 1) // 0.1% errors, SLO allows 0.1%
	tracker := NewErrorBudgetTracker(wa, 99.9)
	tracker.Evaluate()
	entry, ok := tracker.Get("svc:443")
	if !ok {
		t.Fatal("expected entry")
	}
	if entry.Exhausted {
		t.Error("budget should not be exhausted at exactly SLO boundary")
	}
	if entry.BudgetUsed < 0 || entry.BudgetUsed > 1.0 {
		t.Errorf("BudgetUsed out of range: %f", entry.BudgetUsed)
	}
}

func TestErrorBudgetTracker_All(t *testing.T) {
	wa := NewWindowAggregator(30 * time.Second)
	for _, tgt := range []string{"a:1", "b:2", "c:3"} {
		wa.Record(Result{Target: tgt})
	}
	tracker := NewErrorBudgetTracker(wa, 99.0)
	tracker.Evaluate()
	if len(tracker.All()) != 3 {
		t.Errorf("expected 3 entries, got %d", len(tracker.All()))
	}
}

func TestErrorBudgetTracker_InvalidSLODefaults(t *testing.T) {
	wa := NewWindowAggregator(30 * time.Second)
	tracker := NewErrorBudgetTracker(wa, -5)
	if tracker.slo != 99.9 {
		t.Errorf("expected default SLO 99.9, got %f", tracker.slo)
	}
}

func TestErrorBudgetTracker_BurnRateNonZero(t *testing.T) {
	wa := buildEBWindow("svc:443", 100, 10)
	tracker := NewErrorBudgetTracker(wa, 99.0)
	tracker.Evaluate()
	entry, _ := tracker.Get("svc:443")
	if entry.BurnRate <= 0 {
		t.Errorf("expected positive burn rate, got %f", entry.BurnRate)
	}
}
