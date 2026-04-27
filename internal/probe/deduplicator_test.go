package probe

import (
	"testing"
	"time"
)

func TestDeduplicator_FirstAlertPasses(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	if d.IsDuplicate("svc-a") {
		t.Fatal("expected first alert to pass through")
	}
}

func TestDeduplicator_SecondAlertWithinCooldownIsDuplicate(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	d.IsDuplicate("svc-a") // record
	if !d.IsDuplicate("svc-a") {
		t.Fatal("expected second alert within cooldown to be duplicate")
	}
}

func TestDeduplicator_AlertAfterCooldownPasses(t *testing.T) {
	now := time.Now()
	d := NewDeduplicator(30 * time.Second)
	d.now = func() time.Time { return now }
	d.IsDuplicate("svc-a")

	// advance time past cooldown
	d.now = func() time.Time { return now.Add(31 * time.Second) }
	if d.IsDuplicate("svc-a") {
		t.Fatal("expected alert after cooldown to pass through")
	}
}

func TestDeduplicator_DifferentTargetsAreIndependent(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	d.IsDuplicate("svc-a")
	if d.IsDuplicate("svc-b") {
		t.Fatal("expected different target to pass through independently")
	}
}

func TestDeduplicator_Reset_AllowsNextAlertThrough(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	d.IsDuplicate("svc-a")
	d.Reset("svc-a")
	if d.IsDuplicate("svc-a") {
		t.Fatal("expected alert to pass after Reset")
	}
}

func TestDeduplicator_ResetAll_ClearsEverything(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	d.IsDuplicate("svc-a")
	d.IsDuplicate("svc-b")
	d.ResetAll()
	if d.IsDuplicate("svc-a") || d.IsDuplicate("svc-b") {
		t.Fatal("expected all targets to pass after ResetAll")
	}
}

func TestDeduplicator_ZeroCooldown_NeverDuplicates(t *testing.T) {
	d := NewDeduplicator(0)
	d.IsDuplicate("svc-a")
	if d.IsDuplicate("svc-a") {
		t.Fatal("expected zero cooldown to never deduplicate")
	}
}
