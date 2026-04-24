package probe

import (
	"sync"
	"time"
)

// HistoryEntry records a single probe result with its timestamp.
type HistoryEntry struct {
	Timestamp time.Time
	Result    Result
}

// History maintains a rolling window of probe results for a target.
type History struct {
	mu      sync.RWMutex
	entries []HistoryEntry
	maxSize int
}

// NewHistory creates a History that retains up to maxSize entries.
func NewHistory(maxSize int) *History {
	if maxSize <= 0 {
		maxSize = 60
	}
	return &History{
		entries: make([]HistoryEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add appends a new result to the history, evicting the oldest if full.
func (h *History) Add(r Result) {
	h.mu.Lock()
	defer h.mu.Unlock()

	entry := HistoryEntry{Timestamp: time.Now(), Result: r}
	if len(h.entries) >= h.maxSize {
		copy(h.entries, h.entries[1:])
		h.entries[len(h.entries)-1] = entry
	} else {
		h.entries = append(h.entries, entry)
	}
}

// Entries returns a snapshot of all stored entries (oldest first).
func (h *History) Entries() []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	snap := make([]HistoryEntry, len(h.entries))
	copy(snap, h.entries)
	return snap
}

// AvgLatency returns the mean latency across all successful entries.
// Returns 0 if there are no successful entries.
func (h *History) AvgLatency() time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var total time.Duration
	var count int
	for _, e := range h.entries {
		if e.Result.Err == nil {
			total += e.Result.Latency
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / time.Duration(count)
}

// ErrorRate returns the fraction of entries that have errors (0.0–1.0).
func (h *History) ErrorRate() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.entries) == 0 {
		return 0
	}
	var errCount int
	for _, e := range h.entries {
		if e.Result.Err != nil {
			errCount++
		}
	}
	return float64(errCount) / float64(len(h.entries))
}
