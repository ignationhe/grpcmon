package probe

import (
	"fmt"
	"time"
)

// SLAPolicy defines availability and latency thresholds for a target.
type SLAPolicy struct {
	// MaxErrorRate is the maximum acceptable error rate (0.0–1.0).
	MaxErrorRate float64
	// MaxAvgLatency is the maximum acceptable average latency.
	MaxAvgLatency time.Duration
	// Window is the number of recent history entries to evaluate.
	Window int
}

// SLAViolation describes a single SLA breach.
type SLAViolation struct {
	Target    string
	Kind      string // "error_rate" or "latency"
	Threshold string
	Actual    string
}

// SLAEvaluator checks probe history against configured SLA policies.
type SLAEvaluator struct {
	policies map[string]SLAPolicy
	defaultP SLAPolicy
}

// NewSLAEvaluator creates an evaluator with a default policy applied to all
// targets unless overridden per target address.
func NewSLAEvaluator(defaultPolicy SLAPolicy, overrides map[string]SLAPolicy) *SLAEvaluator {
	if overrides == nil {
		overrides = make(map[string]SLAPolicy)
	}
	return &SLAEvaluator{policies: overrides, defaultP: defaultPolicy}
}

// policyFor returns the effective SLAPolicy for a target address.
func (e *SLAEvaluator) policyFor(target string) SLAPolicy {
	if p, ok := e.policies[target]; ok {
		return p
	}
	return e.defaultP
}

// Evaluate checks the given history for SLA violations and returns any found.
func (e *SLAEvaluator) Evaluate(target string, h *History) []SLAViolation {
	policy := e.policyFor(target)
	window := policy.Window
	if window <= 0 {
		window = 20
	}

	var violations []SLAViolation

	if policy.MaxErrorRate > 0 {
		rate := h.ErrorRate(window)
		if rate > policy.MaxErrorRate {
			violations = append(violations, SLAViolation{
				Target:    target,
				Kind:      "error_rate",
				Threshold: fmt.Sprintf("%.1f%%", policy.MaxErrorRate*100),
				Actual:    fmt.Sprintf("%.1f%%", rate*100),
			})
		}
	}

	if policy.MaxAvgLatency > 0 {
		avg := h.AvgLatency(window)
		if avg > policy.MaxAvgLatency {
			violations = append(violations, SLAViolation{
				Target:    target,
				Kind:      "latency",
				Threshold: policy.MaxAvgLatency.String(),
				Actual:    avg.String(),
			})
		}
	}

	return violations
}
