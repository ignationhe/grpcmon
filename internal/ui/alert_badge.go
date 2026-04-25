package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
	"grpcmon/internal/probe"
)

// AlertBadge is a tview component that displays active probe alerts.
type AlertBadge struct {
	TextView *tview.TextView
}

// NewAlertBadge creates and returns a new AlertBadge with default styling.
func NewAlertBadge() *AlertBadge {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle(" ⚠ Alerts ")
	tv.SetTitleAlign(tview.AlignLeft)

	b := &AlertBadge{TextView: tv}
	b.Update(nil)
	return b
}

// Update refreshes the badge content with the provided list of active alerts.
func (b *AlertBadge) Update(alerts []probe.Alert) {
	b.TextView.Clear()

	if len(alerts) == 0 {
		fmt.Fprint(b.TextView, "[green]No active alerts[-]")
		return
	}

	var sb strings.Builder
	for i, a := range alerts {
		if i > 0 {
			sb.WriteString("\n")
		}
		severityColor := severityColor(a.Severity)
		sb.WriteString(fmt.Sprintf(
			"[%s]%-20s[-] %s  [grey]%s[-]",
			severityColor,
			a.Target,
			a.Message,
			a.FiredAt.Format("15:04:05"),
		))
	}
	fmt.Fprint(b.TextView, sb.String())
}

func severityColor(s probe.Severity) string {
	switch s {
	case probe.SeverityCritical:
		return "red"
	case probe.SeverityWarning:
		return "yellow"
	default:
		return "white"
	}
}
