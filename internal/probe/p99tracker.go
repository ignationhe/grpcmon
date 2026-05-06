package probe

import (
	"sort"
	"sync"
	"time"
)

// P99Sample holds a single latency observation with a timestamp.
type P99Sample struct {
	Latency time.Duration
	At      time.Time
}

// P99Tracker maintains a rolling window of latency samples per target
// and computes the p50, p95, and p99 percentiles on demand.
type P99Tracker struct {
	mu      sync.Mutex
	window  time.Duration
	samples map[string][]P99Sample
}

// NewP99Tracker creates a P99Tracker whose samples are evicted after window.
func NewP99Tracker(window time.Duration) *P99Tracker {
	return &P99Tracker{
		window:  window,
		samples: make(map[string][]P99Sample),
	}
}

// Record adds a latency sample for the given target address.
func (t *P99Tracker) Record(addr string, latency time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	t.samples[addr] = append(t.samples[addr], P99Sample{Latency: latency, At: now})
	t.evict(addr, now)
}

// evict removes samples older than the window. Must be called with mu held.
func (t *P99Tracker) evict(addr string, now time.Time) {
	cutoff := now.Add(-t.window)
	ss := t.samples[addr]
	start := 0
	for start < len(ss) && ss[start].At.Before(cutoff) {
		start++
	}
	t.samples[addr] = ss[start:]
}

// Percentile returns the p-th percentile latency (0–100) for addr.
// Returns 0 if no samples are available.
func (t *P99Tracker) Percentile(addr string, p float64) time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.evict(addr, time.Now())
	ss := t.samples[addr]
	if len(ss) == 0 {
		return 0
	}
	vals := make([]time.Duration, len(ss))
	for i, s := range ss {
		vals[i] = s.Latency
	}
	sort.Slice(vals, func(i, j int) bool { return vals[i] < vals[j] })
	idx := int(float64(len(vals)-1) * p / 100.0)
	return vals[idx]
}

// Targets returns all addresses that have at least one live sample.
func (t *P99Tracker) Targets() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	out := make([]string, 0, len(t.samples))
	for addr := range t.samples {
		t.evict(addr, now)
		if len(t.samples[addr]) > 0 {
			out = append(out, addr)
		}
	}
	return out
}
