package probe

import (
	"sync"
	"time"
)

// CooldownPolicy defines per-target cooldown windows that suppress
// repeated probe attempts after a recent failure.
type CooldownPolicy struct {
	mu       sync.Mutex
	lastFail map[string]time.Time
	window   time.Duration
}

// NewCooldownPolicy creates a CooldownPolicy with the given suppression window.
func NewCooldownPolicy(window time.Duration) *CooldownPolicy {
	if window <= 0 {
		window = 10 * time.Second
	}
	return &CooldownPolicy{
		lastFail: make(map[string]time.Time),
		window:   window,
	}
}

// RecordFailure marks the current time as the last failure for target.
func (c *CooldownPolicy) RecordFailure(target string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastFail[target] = time.Now()
}

// InCooldown reports whether target is still within its suppression window.
func (c *CooldownPolicy) InCooldown(target string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	t, ok := c.lastFail[target]
	if !ok {
		return false
	}
	return time.Since(t) < c.window
}

// Reset clears the cooldown state for target, allowing the next probe
// through immediately regardless of the window.
func (c *CooldownPolicy) Reset(target string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.lastFail, target)
}

// Targets returns all targets that currently have a recorded failure,
// including those whose cooldown has already expired.
func (c *CooldownPolicy) Targets() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, 0, len(c.lastFail))
	for t := range c.lastFail {
		out = append(out, t)
	}
	return out
}
