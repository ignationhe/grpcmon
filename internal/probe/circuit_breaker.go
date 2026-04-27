package probe

import (
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // blocking requests
	StateHalfOpen              // testing recovery
)

// CircuitBreaker prevents hammering a failing target by temporarily
// stopping probes after consecutive failures exceed a threshold.
type CircuitBreaker struct {
	mu           sync.Mutex
	state        State
	failures     int
	threshold    int
	cooldown     time.Duration
	openedAt     time.Time
	successes    int
	probeSuccess int // successes required in half-open to close
}

// NewCircuitBreaker creates a CircuitBreaker with the given failure threshold
// and cooldown duration before attempting recovery.
func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold:    threshold,
		cooldown:     cooldown,
		probeSuccess: 1,
	}
}

// Allow reports whether a probe attempt should proceed.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.openedAt) >= cb.cooldown {
			cb.state = StateHalfOpen
			cb.successes = 0
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return false
}

// RecordSuccess records a successful probe result.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.failures = 0
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.probeSuccess {
			cb.state = StateClosed
			cb.failures = 0
		}
	}
}

// RecordFailure records a failed probe result.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.failures++
		if cb.failures >= cb.threshold {
			cb.state = StateOpen
			cb.openedAt = time.Now()
		}
	case StateHalfOpen:
		cb.state = StateOpen
		cb.openedAt = time.Now()
	}
}

// CurrentState returns the current circuit breaker state.
func (cb *CircuitBreaker) CurrentState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
