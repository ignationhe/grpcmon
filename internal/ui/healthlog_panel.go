package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/yourorg/grpcmon/internal/probe"
)

// HealthLogPanel renders recent health status transitions in the dashboard.
type HealthLogPanel struct {
	view *tview.TextView
	maxRows int
}

// NewHealthLogPanel creates a panel that displays up to maxRows transition events.
func NewHealthLogPanel(maxRows int) *HealthLogPanel {
	if maxRows <= 0 {
		maxRows = 10
	}
	tv := tview.NewTextView()
	tv.SetBorder(true)
	tv.SetTitle(" Health Transitions ")
	tv.SetDynamicColors(true)
	tv.SetScrollable(true)
	return &HealthLogPanel{view: tv, maxRows: maxRows}
}

// Update refreshes the panel with the latest events from the HealthLog.
func (p *HealthLogPanel) Update(log *probe.HealthLog) {
	events := log.Events()
	if len(events) > p.maxRows {
		events = events[len(events)-p.maxRows:]
	}

	var sb strings.Builder
	if len(events) == 0 {
		sb.WriteString("[grey]No transitions recorded yet[-]\n")
	} else {
		for i := len(events) - 1; i >= 0; i-- {
			e := events[i]
			ts := e.OccurredAt.Format("15:04:05")
			prevColor := statusColor(e.Previous)
			currColor := statusColor(e.Current)
			line := fmt.Sprintf(
				"[grey]%s[-] [white]%-24s[-] [%s]%s[-] → [%s]%s[-]\n",
				ts, truncateTarget(e.Target, 24),
				prevColor, e.Previous,
				currColor, e.Current,
			)
			sb.WriteString(line)
		}
	}

	p.view.SetText(sb.String())
}

// View returns the underlying tview primitive.
func (p *HealthLogPanel) View() *tview.TextView {
	return p.view
}

func statusColor(status string) string {
	switch status {
	case "SERVING":
		return "green"
	case "NOT_SERVING":
		return "red"
	default:
		return "yellow"
	}
}

func truncateTarget(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
