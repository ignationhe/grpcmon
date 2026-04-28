package probe

import (
	"testing"
	"time"
)

func makeSnapshot(targets []TargetSnapshot) Snapshot {
	return Snapshot{
		CapturedAt: time.Now(),
		Targets:    targets,
	}
}

func TestSnapshotStore_InitiallyEmpty(t *testing.T) {
	store := NewSnapshotStore()
	if store.Current() != nil {
		t.Fatal("expected nil current snapshot")
	}
	if store.Previous() != nil {
		t.Fatal("expected nil previous snapshot")
	}
}

func TestSnapshotStore_SaveAndCurrent(t *testing.T) {
	store := NewSnapshotStore()
	snap := makeSnapshot([]TargetSnapshot{{Address: "localhost:50051", Status: "SERVING"}})
	store.Save(snap)
	cur := store.Current()
	if cur == nil {
		t.Fatal("expected non-nil current snapshot")
	}
	if len(cur.Targets) != 1 || cur.Targets[0].Address != "localhost:50051" {
		t.Errorf("unexpected targets: %+v", cur.Targets)
	}
}

func TestSnapshotStore_PromotesToPrevious(t *testing.T) {
	store := NewSnapshotStore()
	first := makeSnapshot([]TargetSnapshot{{Address: "a:1", Status: "SERVING"}})
	second := makeSnapshot([]TargetSnapshot{{Address: "a:1", Status: "NOT_SERVING"}})
	store.Save(first)
	store.Save(second)
	prev := store.Previous()
	if prev == nil {
		t.Fatal("expected non-nil previous snapshot")
	}
	if prev.Targets[0].Status != "SERVING" {
		t.Errorf("expected previous status SERVING, got %s", prev.Targets[0].Status)
	}
}

func TestSnapshotStore_Diff_DetectsChange(t *testing.T) {
	store := NewSnapshotStore()
	store.Save(makeSnapshot([]TargetSnapshot{{Address: "svc:80", Status: "SERVING"}}))
	store.Save(makeSnapshot([]TargetSnapshot{{Address: "svc:80", Status: "NOT_SERVING"}}))
	changed := store.Diff()
	if len(changed) != 1 || changed[0] != "svc:80" {
		t.Errorf("expected [svc:80] in diff, got %v", changed)
	}
}

func TestSnapshotStore_Diff_NoChangeWhenSameStatus(t *testing.T) {
	store := NewSnapshotStore()
	store.Save(makeSnapshot([]TargetSnapshot{{Address: "svc:80", Status: "SERVING"}}))
	store.Save(makeSnapshot([]TargetSnapshot{{Address: "svc:80", Status: "SERVING"}}))
	changed := store.Diff()
	if len(changed) != 0 {
		t.Errorf("expected no diff, got %v", changed)
	}
}

func TestSnapshotStore_Diff_NilWhenOnlyOneSaved(t *testing.T) {
	store := NewSnapshotStore()
	store.Save(makeSnapshot([]TargetSnapshot{{Address: "svc:80", Status: "SERVING"}}))
	changed := store.Diff()
	if changed != nil {
		t.Errorf("expected nil diff with only one snapshot, got %v", changed)
	}
}

func TestSnapshotStore_IsolatesStoredCopy(t *testing.T) {
	store := NewSnapshotStore()
	targets := []TargetSnapshot{{Address: "x:9", Status: "SERVING"}}
	snap := makeSnapshot(targets)
	store.Save(snap)
	targets[0].Status = "NOT_SERVING" // mutate original slice
	if store.Current().Targets[0].Status != "SERVING" {
		t.Error("snapshot should be isolated from external mutation")
	}
}
