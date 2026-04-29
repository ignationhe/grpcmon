package probe

import (
	"testing"
	"time"
)

func makeTrace(target, status string, dur time.Duration) TraceEntry {
	return TraceEntry{
		Target:    target,
		Timestamp: time.Now(),
		Duration:  dur,
		Status:    status,
		Message:   "test",
	}
}

func TestTraceLog_EmptyInitially(t *testing.T) {
	log := NewTraceLog(10)
	if got := log.Entries(); len(got) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(got))
	}
}

func TestTraceLog_AddAndEntries(t *testing.T) {
	log := NewTraceLog(10)
	log.Add(makeTrace("svc:50051", "SERVING", 10*time.Millisecond))
	log.Add(makeTrace("svc:50052", "NOT_SERVING", 20*time.Millisecond))
	entries := log.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Target != "svc:50051" {
		t.Errorf("unexpected target: %s", entries[0].Target)
	}
}

func TestTraceLog_Eviction(t *testing.T) {
	log := NewTraceLog(3)
	for i := 0; i < 5; i++ {
		log.Add(makeTrace("svc", "SERVING", time.Millisecond))
	}
	if got := len(log.Entries()); got != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", got)
	}
}

func TestTraceLog_ForTarget(t *testing.T) {
	log := NewTraceLog(20)
	log.Add(makeTrace("a:1", "SERVING", time.Millisecond))
	log.Add(makeTrace("b:2", "SERVING", time.Millisecond))
	log.Add(makeTrace("a:1", "NOT_SERVING", time.Millisecond))
	got := log.ForTarget("a:1")
	if len(got) != 2 {
		t.Fatalf("expected 2 entries for a:1, got %d", len(got))
	}
	for _, e := range got {
		if e.Target != "a:1" {
			t.Errorf("unexpected target in filtered result: %s", e.Target)
		}
	}
}

func TestTraceLog_Clear(t *testing.T) {
	log := NewTraceLog(10)
	log.Add(makeTrace("svc", "SERVING", time.Millisecond))
	log.Clear()
	if got := len(log.Entries()); got != 0 {
		t.Fatalf("expected 0 after clear, got %d", got)
	}
}

func TestTraceLog_DefaultMaxSize(t *testing.T) {
	log := NewTraceLog(0)
	for i := 0; i < 250; i++ {
		log.Add(makeTrace("svc", "SERVING", time.Millisecond))
	}
	if got := len(log.Entries()); got != 200 {
		t.Fatalf("expected 200 entries with default max, got %d", got)
	}
}
