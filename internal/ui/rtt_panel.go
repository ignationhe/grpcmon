package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/yourorg/grpcmon/internal/probe"
)

// RTTPanel renders round-trip time statistics (min/max/avg) for all
// monitored targets in a tview TextView.
type RTTPanel struct {
	view    *tview.TextView
	tracker *probe.RTTTracker
}

// NewRTTPanel creates an RTTPanel backed by the given RTTTracker.
func NewRTTPanel(tracker *probe.RTTTracker) *RTTPanel {
	v := tview.NewTextView()
	v.SetDynamicColors(true)
	v.SetBorder(true)
	v.SetTitle(" RTT Statistics ")
	return &RTTPanel{view: v, tracker: tracker}
}

// View returns the underlying tview primitive for layout composition.
func (p *RTTPanel) View() *tview.TextView {
	return p.view
}

// Refresh re-renders the RTT statistics from the current tracker state.
func (p *RTTPanel) Refresh() {
	targets := p.tracker.Targets()
	if len(targets) == 0 {
		p.view.SetText("[grey]No RTT data available[-]")
		return
	}

	var sb strings.Builder
	for _, target := range targets {
		min, max, avg := p.tracker.Stats(target)
		minMs := float64(min.Microseconds()) / 1000.0
		maxMs := float64(max.Microseconds()) / 1000.0
		avgMs := float64(avg.Microseconds()) / 1000.0

		avgColor := rttColor(avgMs)
		fmt.Fprintf(&sb,
			"[white]%-30s[-]  min [cyan]%6.2fms[-]  max [cyan]%6.2fms[-]  avg [%s]%6.2fms[-]\n",
			truncateRTT(target, 30), minMs, maxMs, avgColor, avgMs,
		)
	}
	p.view.SetText(sb.String())
}

// rttColor returns a tview color tag based on the average latency in ms.
func rttColor(avgMs float64) string {
	switch {
	case avgMs < 50:
		return "green"
	case avgMs < 150:
		return "yellow"
	default:
		return "red"
	}
}

func truncateRTT(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
