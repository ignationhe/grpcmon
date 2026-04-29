package probe

import (
	"sync"
	"time"
)

// TraceEntry records a single probe trace event.
type TraceEntry struct {
	Target    string
	Timestamp time.Time
	Duration  time.Duration
	Status    string
	Message   string
}

// TraceLog maintains a bounded, thread-safe log of probe trace entries.
type TraceLog struct {
	mu      sync.RWMutex
	entries []TraceEntry
	maxSize int
}

// NewTraceLog creates a TraceLog that retains at most maxSize entries.
func NewTraceLog(maxSize int) *TraceLog {
	if maxSize <= 0 {
		maxSize = 200
	}
	return &TraceLog{maxSize: maxSize}
}

// Add appends a new trace entry, evicting the oldest if at capacity.
func (t *TraceLog) Add(entry TraceEntry) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.entries) >= t.maxSize {
		t.entries = t.entries[1:]
	}
	t.entries = append(t.entries, entry)
}

// Entries returns a copy of all trace entries in chronological order.
func (t *TraceLog) Entries() []TraceEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]TraceEntry, len(t.entries))
	copy(out, t.entries)
	return out
}

// ForTarget returns trace entries for a specific target.
func (t *TraceLog) ForTarget(target string) []TraceEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var out []TraceEntry
	for _, e := range t.entries {
		if e.Target == target {
			out = append(out, e)
		}
	}
	return out
}

// Clear removes all entries from the log.
func (t *TraceLog) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries = t.entries[:0]
}
