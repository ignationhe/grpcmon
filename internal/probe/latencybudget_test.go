package probe

import (
	"testing"
	"time"
)

func TestLatencyBudget_NotExceededInitially(t *testing.T) {
	lb := NewLatencyBudget(100 * time.Millisecond)
	if lb.Exceeded("host:443") {
		t.Fatal("expected budget not exceeded initially")
	}
}

func TestLatencyBudget_ExceedsAfterRecord(t *testing.T) {
	lb := NewLatencyBudget(50 * time.Millisecond)
	lb.Record("host:443", 60*time.Millisecond)
	if !lb.Exceeded("host:443") {
		t.Fatal("expected budget exceeded")
	}
}

func TestLatencyBudget_RemainingDecreases(t *testing.T) {
	lb := NewLatencyBudget(100 * time.Millisecond)
	lb.Record("host:443", 30*time.Millisecond)
	if got := lb.Remaining("host:443"); got != 70*time.Millisecond {
		t.Fatalf("expected 70ms remaining, got %v", got)
	}
}

func TestLatencyBudget_SetBudgetOverridesDefault(t *testing.T) {
	lb := NewLatencyBudget(100 * time.Millisecond)
	lb.SetBudget("host:443", 20*time.Millisecond)
	lb.Record("host:443", 25*time.Millisecond)
	if !lb.Exceeded("host:443") {
		t.Fatal("expected custom budget to be exceeded")
	}
}

func TestLatencyBudget_DefaultAppliedToUnknownTarget(t *testing.T) {
	lb := NewLatencyBudget(200 * time.Millisecond)
	lb.Record("other:443", 100*time.Millisecond)
	if lb.Exceeded("other:443") {
		t.Fatal("should not exceed default budget of 200ms with 100ms spent")
	}
}

func TestLatencyBudget_Reset_ClearsSpent(t *testing.T) {
	lb := NewLatencyBudget(50 * time.Millisecond)
	lb.Record("host:443", 60*time.Millisecond)
	lb.Reset("host:443")
	if lb.Exceeded("host:443") {
		t.Fatal("expected budget not exceeded after reset")
	}
}

func TestLatencyBudget_All_ReturnsCopy(t *testing.T) {
	lb := NewLatencyBudget(100 * time.Millisecond)
	lb.Record("a:1", 10*time.Millisecond)
	lb.Record("b:2", 20*time.Millisecond)
	all := lb.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all["a:1"] != 10*time.Millisecond {
		t.Errorf("unexpected value for a:1: %v", all["a:1"])
	}
	// Mutating returned map must not affect internal state.
	all["a:1"] = 999 * time.Second
	if lb.Remaining("a:1") != 90*time.Millisecond {
		t.Error("internal state was mutated via All() return value")
	}
}

func TestLatencyBudget_CumulativeRecord(t *testing.T) {
	lb := NewLatencyBudget(100 * time.Millisecond)
	lb.Record("host:443", 40*time.Millisecond)
	lb.Record("host:443", 40*time.Millisecond)
	if lb.Exceeded("host:443") {
		t.Fatal("80ms should not exceed 100ms budget")
	}
	lb.Record("host:443", 30*time.Millisecond)
	if !lb.Exceeded("host:443") {
		t.Fatal("110ms should exceed 100ms budget")
	}
}
