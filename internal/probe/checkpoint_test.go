package probe

import (
	"testing"
	"time"
)

func TestCheckpoint_IsStale_NoRecord(t *testing.T) {
	cp := NewCheckpoint(30 * time.Second)
	if !cp.IsStale("host:443", time.Now()) {
		t.Fatal("expected stale when no record exists")
	}
}

func TestCheckpoint_IsStale_FreshRecord(t *testing.T) {
	cp := NewCheckpoint(30 * time.Second)
	now := time.Now()
	cp.Record("host:443", now)
	if cp.IsStale("host:443", now.Add(10*time.Second)) {
		t.Fatal("expected fresh within window")
	}
}

func TestCheckpoint_IsStale_ExpiredRecord(t *testing.T) {
	cp := NewCheckpoint(30 * time.Second)
	now := time.Now()
	cp.Record("host:443", now)
	if !cp.IsStale("host:443", now.Add(31*time.Second)) {
		t.Fatal("expected stale after window expires")
	}
}

func TestCheckpoint_LastSeen_Missing(t *testing.T) {
	cp := NewCheckpoint(time.Minute)
	_, ok := cp.LastSeen("missing:443")
	if ok {
		t.Fatal("expected no record for unknown target")
	}
}

func TestCheckpoint_LastSeen_Present(t *testing.T) {
	cp := NewCheckpoint(time.Minute)
	now := time.Now()
	cp.Record("host:443", now)
	got, ok := cp.LastSeen("host:443")
	if !ok {
		t.Fatal("expected record to be present")
	}
	if !got.Equal(now) {
		t.Fatalf("expected %v got %v", now, got)
	}
}

func TestCheckpoint_StaleTargets(t *testing.T) {
	cp := NewCheckpoint(30 * time.Second)
	now := time.Now()
	cp.Record("fresh:443", now)
	cp.Record("stale:443", now.Add(-60*time.Second))

	targets := []string{"fresh:443", "stale:443", "unknown:443"}
	stale := cp.StaleTargets(targets, now)

	if len(stale) != 2 {
		t.Fatalf("expected 2 stale targets, got %d: %v", len(stale), stale)
	}
	for _, s := range stale {
		if s == "fresh:443" {
			t.Fatal("fresh target incorrectly marked stale")
		}
	}
}

func TestCheckpoint_Reset_AllowsNextProbeToBeStale(t *testing.T) {
	cp := NewCheckpoint(30 * time.Second)
	now := time.Now()
	cp.Record("host:443", now)
	cp.Reset("host:443")
	if !cp.IsStale("host:443", now) {
		t.Fatal("expected stale after reset")
	}
}

func TestCheckpoint_DifferentTargetsIndependent(t *testing.T) {
	cp := NewCheckpoint(30 * time.Second)
	now := time.Now()
	cp.Record("a:443", now)
	if !cp.IsStale("b:443", now) {
		t.Fatal("b should be stale independently of a")
	}
	if cp.IsStale("a:443", now) {
		t.Fatal("a should be fresh")
	}
}
