package probe

import (
	"sync"
	"time"
)

// RetryBudget tracks per-target retry allowances within a rolling time window.
// It enforces a maximum number of retries to prevent retry storms.
type RetryBudget struct {
	mu         sync.Mutex
	window     time.Duration
	maxRetries int
	entries    map[string][]time.Time
	now        func() time.Time
}

// NewRetryBudget creates a RetryBudget with the given window and max retries per window.
func NewRetryBudget(window time.Duration, maxRetries int) *RetryBudget {
	return &RetryBudget{
		window:     window,
		maxRetries: maxRetries,
		entries:    make(map[string][]time.Time),
		now:        time.Now,
	}
}

// Allow returns true if a retry is permitted for the given target and records it.
func (rb *RetryBudget) Allow(target string) bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	now := rb.now()
	cutoff := now.Add(-rb.window)

	ts := rb.entries[target]
	filtered := ts[:0]
	for _, t := range ts {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= rb.maxRetries {
		rb.entries[target] = filtered
		return false
	}

	rb.entries[target] = append(filtered, now)
	return true
}

// Remaining returns the number of retries still available for the given target.
func (rb *RetryBudget) Remaining(target string) int {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	now := rb.now()
	cutoff := now.Add(-rb.window)

	count := 0
	for _, t := range rb.entries[target] {
		if t.After(cutoff) {
			count++
		}
	}

	remaining := rb.maxRetries - count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Reset clears the retry history for a target.
func (rb *RetryBudget) Reset(target string) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	delete(rb.entries, target)
}

// Targets returns all targets with recorded retry attempts.
func (rb *RetryBudget) Targets() []string {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	out := make([]string, 0, len(rb.entries))
	for k := range rb.entries {
		out = append(out, k)
	}
	return out
}
