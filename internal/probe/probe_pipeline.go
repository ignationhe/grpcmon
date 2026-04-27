package probe

import (
	"context"
	"time"
)

// ProbePipeline composes retry, timeout, circuit-breaker, rate-limiter,
// and throttle policies around a base probe function.
type ProbePipeline struct {
	retry    *RetryPolicy
	timeout  *TimeoutPolicy
	cb       *CircuitBreaker
	rl       *RateLimiter
	throttle *ThrottlePolicy
}

// ProbePipelineOption configures a ProbePipeline.
type ProbePipelineOption func(*ProbePipeline)

// WithCircuitBreaker attaches a CircuitBreaker to the pipeline.
func WithCircuitBreaker(cb *CircuitBreaker) ProbePipelineOption {
	return func(p *ProbePipeline) { p.cb = cb }
}

// WithRateLimiterOption attaches a RateLimiter to the pipeline.
func WithRateLimiterOption(rl *RateLimiter) ProbePipelineOption {
	return func(p *ProbePipeline) { p.rl = rl }
}

// WithThrottleOption attaches a ThrottlePolicy to the pipeline.
func WithThrottleOption(th *ThrottlePolicy) ProbePipelineOption {
	return func(p *ProbePipeline) { p.throttle = th }
}

// NewProbePipeline creates a pipeline with the given retry/timeout policies
// and any additional options.
func NewProbePipeline(retry *RetryPolicy, timeout *TimeoutPolicy, opts ...ProbePipelineOption) *ProbePipeline {
	pp := &ProbePipeline{retry: retry, timeout: timeout}
	for _, o := range opts {
		o(pp)
	}
	return pp
}

// Run executes fn for target through all configured middleware layers.
// Returns the Result produced by fn or a synthetic error Result.
func (pp *ProbePipeline) Run(ctx context.Context, target string, fn func(context.Context) Result) Result {
	if pp.throttle != nil {
		if !pp.throttle.Acquire(target) {
			return Result{Target: target, Status: "THROTTLED", Err: nil, Latency: 0, Timestamp: time.Now()}
		}
		defer pp.throttle.Release()
	}

	if pp.rl != nil {
		if err := pp.rl.Wait(ctx, target); err != nil {
			return Result{Target: target, Status: "NOT_SERVING", Err: err, Latency: 0, Timestamp: time.Now()}
		}
	}

	if pp.cb != nil {
		if !pp.cb.Allow() {
			return Result{Target: target, Status: "OPEN", Err: nil, Latency: 0, Timestamp: time.Now()}
		}
	}

	wrapped := func(c context.Context) Result { return WithTimeout(c, pp.timeout, target, fn) }
	result := WithRetry(ctx, pp.retry, target, wrapped)

	if pp.cb != nil {
		if result.Err != nil {
			pp.cb.RecordFailure()
		} else {
			pp.cb.RecordSuccess()
		}
	}

	return result
}
