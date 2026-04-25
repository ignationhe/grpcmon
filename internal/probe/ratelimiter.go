package probe

import (
	"context"
	"sync"
	"time"
)

// RateLimiter controls how frequently probes may be issued per target.
type RateLimiter struct {
	mu       sync.Mutex
	last     map[string]time.Time
	minDelay time.Duration
}

// NewRateLimiter creates a RateLimiter that enforces a minimum delay between
// successive probes for the same target address.
func NewRateLimiter(minDelay time.Duration) *RateLimiter {
	return &RateLimiter{
		last:     make(map[string]time.Time),
		minDelay: minDelay,
	}
}

// Wait blocks until the rate limit for addr allows the next probe.
// It respects context cancellation and returns ctx.Err() if cancelled.
func (r *RateLimiter) Wait(ctx context.Context, addr string) error {
	r.mu.Lock()
	lastTime, ok := r.last[addr]
	r.mu.Unlock()

	if ok {
		waitUntil := lastTime.Add(r.minDelay)
		delay := time.Until(waitUntil)
		if delay > 0 {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	r.mu.Lock()
	r.last[addr] = time.Now()
	r.mu.Unlock()
	return nil
}

// Reset clears the recorded probe time for addr, allowing an immediate probe.
func (r *RateLimiter) Reset(addr string) {
	r.mu.Lock()
	delete(r.last, addr)
	r.mu.Unlock()
}

// LastProbe returns the time of the most recent probe for addr and whether
// a probe has been recorded at all.
func (r *RateLimiter) LastProbe(addr string) (time.Time, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.last[addr]
	return t, ok
}
