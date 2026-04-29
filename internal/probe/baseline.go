package probe

import (
	"sync"
	"time"
)

// BaselineEntry holds the computed baseline latency statistics for a target.
type BaselineEntry struct {
	Target    string
	P50       time.Duration
	P95       time.Duration
	P99       time.Duration
	SampleN   int
	ComputedAt time.Time
}

// BaselineStore computes and stores latency baselines derived from history.
type BaselineStore struct {
	mu        sync.RWMutex
	baselines map[string]BaselineEntry
}

// NewBaselineStore creates a new BaselineStore.
func NewBaselineStore() *BaselineStore {
	return &BaselineStore{
		baselines: make(map[string]BaselineEntry),
	}
}

// Compute derives percentile baselines from the given history for target.
func (b *BaselineStore) Compute(target string, h *History) {
	entries := h.Entries()
	if len(entries) == 0 {
		return
	}

	latencies := make([]time.Duration, 0, len(entries))
	for _, e := range entries {
		if e.Err == nil {
			latencies = append(latencies, e.Latency)
		}
	}
	if len(latencies) == 0 {
		return
	}

	sortDurations(latencies)

	entry := BaselineEntry{
		Target:     target,
		P50:        percentile(latencies, 50),
		P95:        percentile(latencies, 95),
		P99:        percentile(latencies, 99),
		SampleN:    len(latencies),
		ComputedAt: time.Now(),
	}

	b.mu.Lock()
	b.baselines[target] = entry
	b.mu.Unlock()
}

// Get returns the baseline for the given target, and whether it exists.
func (b *BaselineStore) Get(target string) (BaselineEntry, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	e, ok := b.baselines[target]
	return e, ok
}

// All returns a copy of all stored baselines.
func (b *BaselineStore) All() []BaselineEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]BaselineEntry, 0, len(b.baselines))
	for _, e := range b.baselines {
		out = append(out, e)
	}
	return out
}

func sortDurations(d []time.Duration) {
	for i := 1; i < len(d); i++ {
		for j := i; j > 0 && d[j] < d[j-1]; j-- {
			d[j], d[j-1] = d[j-1], d[j]
		}
	}
}

func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := (p * len(sorted)) / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
