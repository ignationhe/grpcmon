package probe

import (
	"sync"
	"time"
)

// QuotaPolicy defines per-target probe budget over a rolling window.
type QuotaPolicy struct {
	mu       sync.Mutex
	window   time.Duration
	maxProbes int
	buckets  map[string][]time.Time
}

// NewQuotaPolicy creates a QuotaPolicy that allows at most maxProbes probes
// per target within the given rolling window duration.
func NewQuotaPolicy(window time.Duration, maxProbes int) *QuotaPolicy {
	if maxProbes <= 0 {
		maxProbes = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &QuotaPolicy{
		window:    window,
		maxProbes: maxProbes,
		buckets:   make(map[string][]time.Time),
	}
}

// Allow returns true and records the probe timestamp if the target is within
// quota, or false if the budget has been exhausted for the current window.
func (q *QuotaPolicy) Allow(target string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-q.window)

	times := q.buckets[target]
	// evict expired timestamps
	valid := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= q.maxProbes {
		q.buckets[target] = valid
		return false
	}

	q.buckets[target] = append(valid, now)
	return true
}

// Remaining returns the number of probes still allowed for target in the
// current window.
func (q *QuotaPolicy) Remaining(target string) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-q.window)

	count := 0
	for _, t := range q.buckets[target] {
		if t.After(cutoff) {
			count++
		}
	}
	rem := q.maxProbes - count
	if rem < 0 {
		return 0
	}
	return rem
}

// Reset clears the probe history for a specific target.
func (q *QuotaPolicy) Reset(target string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.buckets, target)
}
