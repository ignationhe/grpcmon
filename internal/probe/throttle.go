package probe

import (
	"sync"
	"time"
)

// ThrottlePolicy controls how many probes can be in-flight concurrently
// and enforces a minimum interval between consecutive probes per target.
type ThrottlePolicy struct {
	mu          sync.Mutex
	maxInflight int
	minInterval time.Duration
	lastProbe   map[string]time.Time
	semaphore   chan struct{}
}

// NewThrottlePolicy creates a ThrottlePolicy with the given max concurrent
// probes and minimum interval between probes for the same target.
func NewThrottlePolicy(maxInflight int, minInterval time.Duration) *ThrottlePolicy {
	return &ThrottlePolicy{
		maxInflight: maxInflight,
		minInterval: minInterval,
		lastProbe:   make(map[string]time.Time),
		semaphore:   make(chan struct{}, maxInflight),
	}
}

// Acquire blocks until a probe slot is available, then reserves it.
// Returns false if the target was probed too recently (within minInterval).
func (t *ThrottlePolicy) Acquire(target string) bool {
	t.mu.Lock()
	if last, ok := t.lastProbe[target]; ok {
		if time.Since(last) < t.minInterval {
			t.mu.Unlock()
			return false
		}
	}
	t.lastProbe[target] = time.Now()
	t.mu.Unlock()

	t.semaphore <- struct{}{}
	return true
}

// Release frees a probe slot previously acquired via Acquire.
func (t *ThrottlePolicy) Release() {
	<-t.semaphore
}

// Inflight returns the number of currently active probes.
func (t *ThrottlePolicy) Inflight() int {
	return len(t.semaphore)
}

// Reset clears the last-probe timestamps for all targets.
func (t *ThrottlePolicy) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastProbe = make(map[string]time.Time)
}
