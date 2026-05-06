package probe

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ConcurrencyTracker tracks the number of in-flight probes per target
// and records peak concurrency observed over a sliding window.
type ConcurrencyTracker struct {
	mu       sync.Mutex
	inflight map[string]int
	peak     map[string]int
	window   time.Duration
	reset    map[string]time.Time
}

// NewConcurrencyTracker creates a ConcurrencyTracker with the given peak-reset window.
func NewConcurrencyTracker(window time.Duration) *ConcurrencyTracker {
	return &ConcurrencyTracker{
		inflight: make(map[string]int),
		peak:     make(map[string]int),
		reset:    make(map[string]time.Time),
		window:   window,
	}
}

// Acquire marks one additional in-flight probe for target.
// Returns an error if target is empty.
func (c *ConcurrencyTracker) Acquire(target string) error {
	if target == "" {
		return fmt.Errorf("concurrency: target must not be empty")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.inflight[target]++
	c.maybeResetPeak(target)
	if c.inflight[target] > c.peak[target] {
		c.peak[target] = c.inflight[target]
	}
	return nil
}

// Release decrements the in-flight count for target.
func (c *ConcurrencyTracker) Release(target string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.inflight[target] > 0 {
		c.inflight[target]--
	}
}

// Inflight returns the current number of in-flight probes for target.
func (c *ConcurrencyTracker) Inflight(target string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.inflight[target]
}

// Peak returns the peak concurrency observed within the current window.
func (c *ConcurrencyTracker) Peak(target string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maybeResetPeak(target)
	return c.peak[target]
}

// Targets returns all targets currently being tracked.
func (c *ConcurrencyTracker) Targets() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, 0, len(c.inflight))
	for t := range c.inflight {
		out = append(out, t)
	}
	return out
}

// Track is a convenience helper that acquires, runs fn, then releases.
func (c *ConcurrencyTracker) Track(ctx context.Context, target string, fn func(context.Context) error) error {
	if err := c.Acquire(target); err != nil {
		return err
	}
	defer c.Release(target)
	return fn(ctx)
}

// maybeResetPeak resets peak if the window has elapsed. Caller must hold mu.
func (c *ConcurrencyTracker) maybeResetPeak(target string) {
	if t, ok := c.reset[target]; !ok || time.Since(t) >= c.window {
		c.peak[target] = c.inflight[target]
		c.reset[target] = time.Now()
	}
}
