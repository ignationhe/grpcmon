package probe

import (
	"sync"
	"time"
)

// RollingMax tracks the maximum observed latency for each target within a
// sliding time window.
type RollingMax struct {
	mu     sync.Mutex
	window time.Duration
	entries map[string][]rollingMaxEntry
}

type rollingMaxEntry struct {
	latency time.Duration
	at      time.Time
}

// NewRollingMax creates a RollingMax that retains samples within the given
// sliding window duration.
func NewRollingMax(window time.Duration) *RollingMax {
	if window <= 0 {
		window = time.Minute
	}
	return &RollingMax{
		window:  window,
		entries: make(map[string][]rollingMaxEntry),
	}
}

// Record adds a latency sample for the given target at the current time.
func (r *RollingMax) Record(target string, latency time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	r.evict(target, now)
	r.entries[target] = append(r.entries[target], rollingMaxEntry{latency: latency, at: now})
}

// Max returns the maximum latency observed for the given target within the
// window. Returns 0 and false if no samples exist.
func (r *RollingMax) Max(target string) (time.Duration, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.evict(target, time.Now())
	samples := r.entries[target]
	if len(samples) == 0 {
		return 0, false
	}
	var max time.Duration
	for _, e := range samples {
		if e.latency > max {
			max = e.latency
		}
	}
	return max, true
}

// Targets returns all targets that have at least one sample in the window.
func (r *RollingMax) Targets() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	out := make([]string, 0, len(r.entries))
	for t := range r.entries {
		r.evict(t, now)
		if len(r.entries[t]) > 0 {
			out = append(out, t)
		}
	}
	return out
}

// evict removes entries older than the window. Must be called with r.mu held.
func (r *RollingMax) evict(target string, now time.Time) {
	cutoff := now.Add(-r.window)
	samples := r.entries[target]
	i := 0
	for i < len(samples) && samples[i].at.Before(cutoff) {
		i++
	}
	r.entries[target] = samples[i:]
}
