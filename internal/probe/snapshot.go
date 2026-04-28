package probe

import (
	"sync"
	"time"
)

// Snapshot holds a point-in-time capture of all target metrics.
type Snapshot struct {
	CapturedAt time.Time
	Targets    []TargetSnapshot
}

// TargetSnapshot holds metrics for a single target at snapshot time.
type TargetSnapshot struct {
	Address    string
	Status     string
	Latency    time.Duration
	ErrorRate  float64
	AvgLatency time.Duration
	AlertCount int
}

// SnapshotStore stores and retrieves the most recent snapshot.
type SnapshotStore struct {
	mu       sync.RWMutex
	current  *Snapshot
	previous *Snapshot
}

// NewSnapshotStore creates an empty SnapshotStore.
func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{}
}

// Save stores a new snapshot, promoting the current one to previous.
func (s *SnapshotStore) Save(snap Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.previous = s.current
	copy := snap
	s.current = &copy
}

// Current returns the latest snapshot, or nil if none exists.
func (s *SnapshotStore) Current() *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// Previous returns the snapshot before the latest, or nil if none exists.
func (s *SnapshotStore) Previous() *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.previous
}

// Diff returns target addresses whose status changed between previous and current.
func (s *SnapshotStore) Diff() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.current == nil || s.previous == nil {
		return nil
	}
	prev := make(map[string]string, len(s.previous.Targets))
	for _, t := range s.previous.Targets {
		prev[t.Address] = t.Status
	}
	var changed []string
	for _, t := range s.current.Targets {
		if old, ok := prev[t.Address]; ok && old != t.Status {
			changed = append(changed, t.Address)
		}
	}
	return changed
}
