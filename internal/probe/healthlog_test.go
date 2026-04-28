package probe

import (
	"testing"
	"time"
)

func makeResult(target, status string) Result {
	return Result{
		Target:    target,
		Status:    status,
		Timestamp: time.Now(),
	}
}

func TestHealthLog_NoTransitionOnFirstSeen(t *testing.T) {
	hl := NewHealthLog(10)
	hl.Record(makeResult("svc", "SERVING"))
	if got := len(hl.Events()); got != 0 {
		t.Fatalf("expected 0 events on first record, got %d", got)
	}
}

func TestHealthLog_RecordsTransition(t *testing.T) {
	hl := NewHealthLog(10)
	hl.Record(makeResult("svc", "SERVING"))
	hl.Record(makeResult("svc", "NOT_SERVING"))

	events := hl.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Previous != "SERVING" || events[0].Current != "NOT_SERVING" {
		t.Errorf("unexpected transition: %+v", events[0])
	}
}

func TestHealthLog_SameStatusNoEvent(t *testing.T) {
	hl := NewHealthLog(10)
	hl.Record(makeResult("svc", "SERVING"))
	hl.Record(makeResult("svc", "SERVING"))
	if got := len(hl.Events()); got != 0 {
		t.Fatalf("expected 0 events for same status, got %d", got)
	}
}

func TestHealthLog_Eviction(t *testing.T) {
	hl := NewHealthLog(3)
	states := []string{"SERVING", "NOT_SERVING", "SERVING", "NOT_SERVING", "SERVING"}
	for _, s := range states {
		hl.Record(makeResult("svc", s))
	}
	if got := len(hl.Events()); got > 3 {
		t.Errorf("expected at most 3 events, got %d", got)
	}
}

func TestHealthLog_MultipleTargetsIndependent(t *testing.T) {
	hl := NewHealthLog(10)
	hl.Record(makeResult("a", "SERVING"))
	hl.Record(makeResult("b", "SERVING"))
	hl.Record(makeResult("a", "NOT_SERVING"))
	hl.Record(makeResult("b", "SERVING"))

	events := hl.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Target != "a" {
		t.Errorf("expected event for target 'a', got %q", events[0].Target)
	}
}

func TestHealthLog_LastStatus(t *testing.T) {
	hl := NewHealthLog(10)
	_, ok := hl.LastStatus("svc")
	if ok {
		t.Fatal("expected no status before any record")
	}
	hl.Record(makeResult("svc", "SERVING"))
	s, ok := hl.LastStatus("svc")
	if !ok || s != "SERVING" {
		t.Errorf("expected SERVING, got %q ok=%v", s, ok)
	}
}
