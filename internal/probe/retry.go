package probe

import (
	"context"
	"time"
)

// RetryPolicy defines how failed probes are retried.
type RetryPolicy struct {
	// MaxAttempts is the total number of attempts (including the first).
	MaxAttempts int
	// Delay is the wait time between attempts.
	Delay time.Duration
}

// DefaultRetryPolicy returns a sensible default retry policy.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		Delay:       100 * time.Millisecond,
	}
}

// WithRetry wraps a probe function and retries it according to the policy.
// It returns the first successful Result or the last failed Result.
func WithRetry(ctx context.Context, policy RetryPolicy, fn func(context.Context) Result) Result {
	if policy.MaxAttempts <= 0 {
		policy.MaxAttempts = 1
	}

	var last Result
	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				last.Error = ctx.Err()
				return last
			case <-time.After(policy.Delay):
			}
		}

		last = fn(ctx)
		if last.Error == nil {
			return last
		}
	}
	return last
}
