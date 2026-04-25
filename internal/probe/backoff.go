package probe

import (
	"math"
	"time"
)

// BackoffPolicy defines how retry delays are calculated.
type BackoffPolicy struct {
	// InitialDelay is the delay before the first retry.
	InitialDelay time.Duration
	// MaxDelay caps the exponential growth.
	MaxDelay time.Duration
	// Multiplier is the factor applied on each retry.
	Multiplier float64
	// Jitter adds a random fraction of the delay to avoid thundering herd.
	Jitter float64
}

// DefaultBackoffPolicy returns a sensible exponential backoff configuration.
func DefaultBackoffPolicy() BackoffPolicy {
	return BackoffPolicy{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.1,
	}
}

// Delay returns the backoff duration for the given attempt (0-indexed).
// It applies exponential growth capped at MaxDelay, plus optional jitter.
func (b BackoffPolicy) Delay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	base := float64(b.InitialDelay) * math.Pow(b.Multiplier, float64(attempt))
	if base > float64(b.MaxDelay) {
		base = float64(b.MaxDelay)
	}
	if b.Jitter > 0 {
		// deterministic jitter: use attempt to vary without importing math/rand
		fraction := float64((attempt*2654435761)%100) / 100.0
		base += base * b.Jitter * fraction
	}
	d := time.Duration(base)
	if d > b.MaxDelay {
		d = b.MaxDelay
	}
	return d
}
