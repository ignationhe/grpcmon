package probe

import (
	"sync"
	"time"
)

// Checkpoint records the last successful probe time for each target,
// allowing the scheduler to detect stale targets that have not been
// successfully probed within a configurable staleness window.
type Checkpoint struct {
	mu      sync.RWMutex
	last    map[string]time.Time
	stale   time.Duration
}

// NewCheckpoint creates a Checkpoint that considers a target stale
// if it has not been successfully probed within the given window.
func NewCheckpoint(staleAfter time.Duration) *Checkpoint {
	return &Checkpoint{
		last:  make(map[string]time.Time),
		stale: staleAfter,
	}
}

// Record marks the given target as successfully probed at now.
func (c *Checkpoint) Record(target string, now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.last[target] = now
}

// LastSeen returns the time of the last successful probe for target
// and whether any record exists.
func (c *Checkpoint) LastSeen(target string) (time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	t, ok := c.last[target]
	return t, ok
}

// IsStale reports whether target has not been successfully probed
// within the staleness window relative to now.
func (c *Checkpoint) IsStale(target string, now time.Time) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	t, ok := c.last[target]
	if !ok {
		return true
	}
	return now.Sub(t) > c.stale
}

// StaleTargets returns the subset of the provided targets that are
// currently considered stale relative to now.
func (c *Checkpoint) StaleTargets(targets []string, now time.Time) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []string
	for _, tgt := range targets {
		t, ok := c.last[tgt]
		if !ok || now.Sub(t) > c.stale {
			out = append(out, tgt)
		}
	}
	return out
}

// Reset clears the checkpoint record for the given target.
func (c *Checkpoint) Reset(target string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.last, target)
}
