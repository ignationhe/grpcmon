package probe

import (
	"sync"
	"time"
)

// LatencyBudget tracks how much of a per-target latency budget has been
// consumed and whether the budget has been exceeded.
type LatencyBudget struct {
	mu      sync.Mutex
	budgets map[string]time.Duration
	spent   map[string]time.Duration
}

// NewLatencyBudget creates a LatencyBudget with the given default budget per
// target. Individual targets may override the default via SetBudget.
func NewLatencyBudget(defaultBudget time.Duration) *LatencyBudget {
	return &LatencyBudget{
		budgets: map[string]time.Duration{"__default__": defaultBudget},
		spent:   make(map[string]time.Duration),
	}
}

// SetBudget overrides the latency budget for a specific target address.
func (lb *LatencyBudget) SetBudget(target string, budget time.Duration) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.budgets[target] = budget
}

// Record adds the observed latency for a target to the spent budget.
func (lb *LatencyBudget) Record(target string, latency time.Duration) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.spent[target] += latency
}

// Exceeded reports whether the total spent latency for target exceeds its
// configured budget.
func (lb *LatencyBudget) Exceeded(target string) bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	budget, ok := lb.budgets[target]
	if !ok {
		budget = lb.budgets["__default__"]
	}
	return lb.spent[target] > budget
}

// Remaining returns how much latency budget is left for a target. A negative
// value means the budget has been exceeded.
func (lb *LatencyBudget) Remaining(target string) time.Duration {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	budget, ok := lb.budgets[target]
	if !ok {
		budget = lb.budgets["__default__"]
	}
	return budget - lb.spent[target]
}

// Reset clears the spent budget for a target, allowing a fresh measurement
// window to begin.
func (lb *LatencyBudget) Reset(target string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	delete(lb.spent, target)
}

// All returns a snapshot of spent durations keyed by target address.
func (lb *LatencyBudget) All() map[string]time.Duration {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	out := make(map[string]time.Duration, len(lb.spent))
	for k, v := range lb.spent {
		out[k] = v
	}
	return out
}
