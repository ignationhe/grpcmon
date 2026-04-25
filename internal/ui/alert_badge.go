package ui

import (
	"fmt"
	"strings"

	"github.com/user/grpcmon/internal/probe"
)

// AlertBadge renders a compact alert summary line for the dashboard.
type AlertBadge struct {
	alerts []probe.Alert
}

// NewAlertBadge creates an AlertBadge.
func NewAlertBadge() *AlertBadge {
	return &AlertBadge{}
}

// Update replaces the current set of alerts.
func (b *AlertBadge) Update(alerts []probe.Alert) {
	b.alerts = alerts
}

// Render returns a single-line string summarising active alerts.
// Critical alerts are prefixed with [CRIT], warnings with [WARN].
func (b *AlertBadge) Render() string {
	if len(b.alerts) == 0 {
		return "  ✓ All targets healthy"
	}

	var parts []string
	crit, warn := 0, 0
	for _, a := range b.alerts {
		switch a.Level {
		case probe.AlertCritical:
			crit++
			parts = append(parts, fmt.Sprintf("[CRIT] %s: %s", a.Target, a.Message))
		case probe.AlertWarning:
			warn++
			parts = append(parts, fmt.Sprintf("[WARN] %s: %s", a.Target, a.Message))
		}
	}

	header := fmt.Sprintf("  Alerts — %d critical, %d warning", crit, warn)
	return header + "\n  " + strings.Join(parts, "  |  ")
}

// ActiveCount returns the number of non-None alerts.
func (b *AlertBadge) ActiveCount() int {
	return len(b.alerts)
}
