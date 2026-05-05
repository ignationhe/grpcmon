package probe

import (
	"testing"
	"time"
)

func TestCooldownPolicy_NotInCooldownInitially(t *testing.T) {
	cp := NewCooldownPolicy(5 * time.Second)
	if cp.InCooldown("svc:443") {
		t.Fatal("expected no cooldown before any failure recorded")
	}
}

func TestCooldownPolicy_InCooldownAfterFailure(t *testing.T) {
	cp := NewCooldownPolicy(5 * time.Second)
	cp.RecordFailure("svc:443")
	if !cp.InCooldown("svc:443") {
		t.Fatal("expected cooldown to be active immediately after failure")
	}
}

func TestCooldownPolicy_ExpiredWindowNotInCooldown(t *testing.T) {
	cp := NewCooldownPolicy(1 * time.Millisecond)
	cp.RecordFailure("svc:443")
	time.Sleep(5 * time.Millisecond)
	if cp.InCooldown("svc:443") {
		t.Fatal("expected cooldown to have expired")
	}
}

func TestCooldownPolicy_Reset_AllowsImmediateProbe(t *testing.T) {
	cp := NewCooldownPolicy(1 * time.Hour)
	cp.RecordFailure("svc:443")
	cp.Reset("svc:443")
	if cp.InCooldown("svc:443") {
		t.Fatal("expected cooldown cleared after Reset")
	}
}

func TestCooldownPolicy_DifferentTargetsIndependent(t *testing.T) {
	cp := NewCooldownPolicy(5 * time.Second)
	cp.RecordFailure("a:80")
	if cp.InCooldown("b:80") {
		t.Fatal("cooldown for a:80 must not affect b:80")
	}
	if !cp.InCooldown("a:80") {
		t.Fatal("a:80 should still be in cooldown")
	}
}

func TestCooldownPolicy_DefaultWindowApplied(t *testing.T) {
	// Zero window should fall back to default (10s), so cooldown is active.
	cp := NewCooldownPolicy(0)
	cp.RecordFailure("svc:443")
	if !cp.InCooldown("svc:443") {
		t.Fatal("expected default window to keep target in cooldown")
	}
}

func TestCooldownPolicy_Targets_ReturnsRecorded(t *testing.T) {
	cp := NewCooldownPolicy(5 * time.Second)
	cp.RecordFailure("a:80")
	cp.RecordFailure("b:80")
	targets := cp.Targets()
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
}

func TestCooldownPolicy_Reset_RemovesFromTargets(t *testing.T) {
	cp := NewCooldownPolicy(5 * time.Second)
	cp.RecordFailure("a:80")
	cp.Reset("a:80")
	if len(cp.Targets()) != 0 {
		t.Fatal("expected no targets after Reset")
	}
}
