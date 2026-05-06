package probe

import (
	"sync"
	"time"
)

// ErrorBudgetEntry holds the computed error budget for a single target.
type ErrorBudgetEntry struct {
	Target     string
	SLOPercent float64 // e.g. 99.9
	ErrorRate  float64 // observed error rate [0,1]
	BudgetUsed float64 // fraction of budget consumed [0,1]
	BurnRate   float64 // errors per second over the window
	Exhausted  bool
	ComputedAt time.Time
}

// ErrorBudgetTracker computes error budget consumption from a WindowAggregator.
type ErrorBudgetTracker struct {
	mu      sync.RWMutex
	window  *WindowAggregator
	slo     float64 // SLO as a percentage, e.g. 99.9
	entries map[string]ErrorBudgetEntry
}

// NewErrorBudgetTracker creates a tracker using the given window aggregator and SLO.
func NewErrorBudgetTracker(window *WindowAggregator, sloPercent float64) *ErrorBudgetTracker {
	if sloPercent <= 0 || sloPercent > 100 {
		sloPercent = 99.9
	}
	return &ErrorBudgetTracker{
		window:  window,
		slo:     sloPercent,
		entries: make(map[string]ErrorBudgetEntry),
	}
}

// Evaluate recomputes the error budget for all known targets.
func (e *ErrorBudgetTracker) Evaluate() {
	allowedErrorRate := 1.0 - (e.slo / 100.0)

	e.mu.Lock()
	defer e.mu.Unlock()

	for _, target := range e.window.Targets() {
		stats := e.window.Stats(target)
		errorRate := stats.ErrorRate

		var budgetUsed float64
		if allowedErrorRate > 0 {
			budgetUsed = errorRate / allowedErrorRate
		} else {
			budgetUsed = 1.0
		}
		if budgetUsed > 1.0 {
			budgetUsed = 1.0
		}

		windowSecs := e.window.WindowDuration().Seconds()
		burnRate := 0.0
		if windowSecs > 0 {
			burnRate = (errorRate * float64(stats.Total)) / windowSecs
		}

		e.entries[target] = ErrorBudgetEntry{
			Target:     target,
			SLOPercent: e.slo,
			ErrorRate:  errorRate,
			BudgetUsed: budgetUsed,
			BurnRate:   burnRate,
			Exhausted:  budgetUsed >= 1.0,
			ComputedAt: time.Now(),
		}
	}
}

// Get returns the latest ErrorBudgetEntry for a target.
func (e *ErrorBudgetTracker) Get(target string) (ErrorBudgetEntry, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	entry, ok := e.entries[target]
	return entry, ok
}

// All returns all computed entries.
func (e *ErrorBudgetTracker) All() []ErrorBudgetEntry {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]ErrorBudgetEntry, 0, len(e.entries))
	for _, v := range e.entries {
		out = append(out, v)
	}
	return out
}
