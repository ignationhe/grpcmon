package probe

import "time"

// AlertLevel indicates the severity of an alert.
type AlertLevel int

const (
	AlertNone    AlertLevel = iota
	AlertWarning            // error rate above warning threshold
	AlertCritical           // error rate above critical threshold or target unreachable
)

// Alert represents a triggered alert for a target.
type Alert struct {
	Target    string
	Level     AlertLevel
	Message   string
	Triggered time.Time
}

// AlertPolicy defines thresholds for alerting.
type AlertPolicy struct {
	WarningErrorRate  float64 // 0.0–1.0
	CriticalErrorRate float64 // 0.0–1.0
	LatencyWarningMs  float64
	LatencyCriticalMs float64
}

// DefaultAlertPolicy returns a sensible default alert policy.
func DefaultAlertPolicy() AlertPolicy {
	return AlertPolicy{
		WarningErrorRate:  0.10,
		CriticalErrorRate: 0.50,
		LatencyWarningMs:  500,
		LatencyCriticalMs: 2000,
	}
}

// Evaluate checks a history snapshot against the policy and returns an Alert.
// Returns an Alert with AlertNone if no threshold is breached.
func (p AlertPolicy) Evaluate(target string, h *History) Alert {
	errRate := h.ErrorRate()
	avgLatency := h.AvgLatency().Seconds() * 1000

	level := AlertNone
	msg := ""

	switch {
	case errRate >= p.CriticalErrorRate:
		level = AlertCritical
		msg = "critical error rate exceeded"
	case avgLatency >= p.LatencyCriticalMs:
		level = AlertCritical
		msg = "critical latency threshold exceeded"
	case errRate >= p.WarningErrorRate:
		level = AlertWarning
		msg = "warning error rate exceeded"
	case avgLatency >= p.LatencyWarningMs:
		level = AlertWarning
		msg = "warning latency threshold exceeded"
	}

	return Alert{
		Target:    target,
		Level:     level,
		Message:   msg,
		Triggered: time.Now(),
	}
}
