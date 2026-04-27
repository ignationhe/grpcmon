package probe

import (
	"context"
	"fmt"
	"time"
)

// TimeoutPolicy defines per-target timeout behaviour.
type TimeoutPolicy struct {
	// Default is the fallback timeout when no target-specific value is set.
	Default time.Duration
	// Overrides maps target addresses to custom timeouts.
	Overrides map[string]time.Duration
}

// DefaultTimeoutPolicy returns a TimeoutPolicy with a 5-second default.
func DefaultTimeoutPolicy() *TimeoutPolicy {
	return &TimeoutPolicy{
		Default:   5 * time.Second,
		Overrides: make(map[string]time.Duration),
	}
}

// For returns the effective timeout for the given target address.
func (p *TimeoutPolicy) For(address string) time.Duration {
	if d, ok := p.Overrides[address]; ok && d > 0 {
		return d
	}
	return p.Default
}

// WithTimeout wraps a ProbeFunc, cancelling the context after the timeout
// returned by policy.For(address). It records a timeout error in the result
// when the deadline is exceeded.
func WithTimeout(policy *TimeoutPolicy, address string, fn ProbeFunc) ProbeFunc {
	return func(ctx context.Context) Result {
		timeout := policy.For(address)
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		resultCh := make(chan Result, 1)
		go func() {
			resultCh <- fn(ctx)
		}()

		select {
		case r := <-resultCh:
			return r
		case <-ctx.Done():
			return Result{
				Address: address,
				Err:     fmt.Errorf("probe timed out after %s: %w", timeout, ctx.Err()),
				Latency: timeout,
			}
		}
	}
}
