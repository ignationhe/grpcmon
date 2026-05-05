package probe

import (
	"testing"
	"time"
)

func TestReconnectTracker_CountZeroInitially(t *testing.T) {
	rt := NewReconnectTracker()
	if got := rt.Count("host:443"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestReconnectTracker_RecordIncrementsCount(t *testing.T) {
	rt := NewReconnectTracker()
	rt.Record("host:443")
	rt.Record("host:443")
	if got := rt.Count("host:443"); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestReconnectTracker_LastSeen_Missing(t *testing.T) {
	rt := NewReconnectTracker()
	_, ok := rt.LastSeen("host:443")
	if ok {
		t.Fatal("expected no last-seen for unknown target")
	}
}

func TestReconnectTracker_LastSeen_Present(t *testing.T) {
	rt := NewReconnectTracker()
	before := time.Now()
	rt.Record("host:443")
	after := time.Now()

	ts, ok := rt.LastSeen("host:443")
	if !ok {
		t.Fatal("expected last-seen to be set")
	}
	if ts.Before(before) || ts.After(after) {
		t.Fatalf("timestamp %v outside expected range [%v, %v]", ts, before, after)
	}
}

func TestReconnectTracker_DifferentTargetsIndependent(t *testing.T) {
	rt := NewReconnectTracker()
	rt.Record("a:443")
	rt.Record("a:443")
	rt.Record("b:443")

	if got := rt.Count("a:443"); got != 2 {
		t.Fatalf("a:443 expected 2, got %d", got)
	}
	if got := rt.Count("b:443"); got != 1 {
		t.Fatalf("b:443 expected 1, got %d", got)
	}
}

func TestReconnectTracker_Reset(t *testing.T) {
	rt := NewReconnectTracker()
	rt.Record("host:443")
	rt.Reset("host:443")

	if got := rt.Count("host:443"); got != 0 {
		t.Fatalf("expected 0 after reset, got %d", got)
	}
	if _, ok := rt.LastSeen("host:443"); ok {
		t.Fatal("expected no last-seen after reset")
	}
}

func TestReconnectTracker_Targets(t *testing.T) {
	rt := NewReconnectTracker()
	rt.Record("a:443")
	rt.Record("b:443")

	targets := rt.Targets()
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
	seen := map[string]bool{}
	for _, tgt := range targets {
		seen[tgt] = true
	}
	if !seen["a:443"] || !seen["b:443"] {
		t.Fatalf("unexpected targets: %v", targets)
	}
}
