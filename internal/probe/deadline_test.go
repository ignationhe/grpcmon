package probe

import (
	"testing"
	"time"
)

func TestDeadlineTracker_DefaultDeadline(t *testing.T) {
	dt := NewDeadlineTracker(5 * time.Second)
	got := dt.DeadlineFor("svc:443")
	if got != 5*time.Second {
		t.Fatalf("expected 5s default, got %v", got)
	}
}

func TestDeadlineTracker_SetAndGet(t *testing.T) {
	dt := NewDeadlineTracker(5 * time.Second)
	dt.SetDeadline("svc:443", 2*time.Second)
	got := dt.DeadlineFor("svc:443")
	if got != 2*time.Second {
		t.Fatalf("expected 2s, got %v", got)
	}
}

func TestDeadlineTracker_OverrideDeadline(t *testing.T) {
	dt := NewDeadlineTracker(5 * time.Second)
	dt.SetDeadline("svc:443", 2*time.Second)
	dt.SetDeadline("svc:443", 3*time.Second)
	got := dt.DeadlineFor("svc:443")
	if got != 3*time.Second {
		t.Fatalf("expected 3s after override, got %v", got)
	}
}

func TestDeadlineTracker_RecordBreach(t *testing.T) {
	dt := NewDeadlineTracker(5 * time.Second)
	now := time.Now()
	dt.RecordBreach("svc:443", now)
	dt.RecordBreach("svc:443", now.Add(time.Second))

	e, ok := dt.Get("svc:443")
	if !ok {
		t.Fatal("expected entry to exist after breach")
	}
	if e.BreachCount != 2 {
		t.Fatalf("expected 2 breaches, got %d", e.BreachCount)
	}
	if !e.LastBreach.Equal(now.Add(time.Second)) {
		t.Fatalf("unexpected LastBreach: %v", e.LastBreach)
	}
}

func TestDeadlineTracker_GetMissing(t *testing.T) {
	dt := NewDeadlineTracker(5 * time.Second)
	_, ok := dt.Get("missing:443")
	if ok {
		t.Fatal("expected missing entry to return false")
	}
}

func TestDeadlineTracker_RecordBreach_UsesDefaultDeadline(t *testing.T) {
	dt := NewDeadlineTracker(10 * time.Second)
	dt.RecordBreach("new:443", time.Now())
	e, ok := dt.Get("new:443")
	if !ok {
		t.Fatal("expected entry after breach")
	}
	if e.Deadline != 10*time.Second {
		t.Fatalf("expected default deadline 10s, got %v", e.Deadline)
	}
}

func TestDeadlineTracker_All(t *testing.T) {
	dt := NewDeadlineTracker(5 * time.Second)
	dt.SetDeadline("a:443", 1*time.Second)
	dt.SetDeadline("b:443", 2*time.Second)
	all := dt.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all["a:443"].Deadline != 1*time.Second {
		t.Errorf("wrong deadline for a:443")
	}
	if all["b:443"].Deadline != 2*time.Second {
		t.Errorf("wrong deadline for b:443")
	}
}

func TestDeadlineTracker_AllReturnsCopy(t *testing.T) {
	dt := NewDeadlineTracker(5 * time.Second)
	dt.SetDeadline("svc:443", 1*time.Second)
	all := dt.All()
	all["svc:443"] = DeadlineEntry{Deadline: 99 * time.Second}
	got := dt.DeadlineFor("svc:443")
	if got != 1*time.Second {
		t.Fatalf("All() mutation affected internal state")
	}
}
