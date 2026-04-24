package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"
	"grpcmon/internal/probe"
)

// TargetPanel renders a single gRPC target's health status and latency.
type TargetPanel struct {
	target string
	textView *tview.TextView
}

// NewTargetPanel creates a new TargetPanel for the given target address.
func NewTargetPanel(target string) *TargetPanel {
	tv := tview.NewTextView()
	tv.SetBorder(true)
	tv.SetTitle(fmt.Sprintf(" %s ", target))
	tv.SetDynamicColors(true)
	return &TargetPanel{
		target:   target,
		textView: tv,
	}
}

// Update refreshes the panel content from the latest probe result and history.
func (p *TargetPanel) Update(result probe.Result, history *probe.History) {
	var sb strings.Builder

	statusColor := "green"
	statusText := "SERVING"
	if result.Err != nil {
		statusColor = "red"
		statusText = fmt.Sprintf("ERROR: %s", result.Err.Error())
	} else if !result.Serving {
		statusColor = "yellow"
		statusText = "NOT SERVING"
	}

	sb.WriteString(fmt.Sprintf("Status : [%s]%s[-]\n", statusColor, statusText))

	if result.Latency > 0 {
		sb.WriteString(fmt.Sprintf("Latency: %s\n", result.Latency.Round(time.Millisecond)))
	} else {
		sb.WriteString("Latency: N/A\n")
	}

	avg := history.AvgLatency()
	if avg > 0 {
		sb.WriteString(fmt.Sprintf("Avg(10) : %s\n", avg.Round(time.Millisecond)))
	}

	errRate := history.ErrorRate()
	sb.WriteString(fmt.Sprintf("Err Rate: %.1f%%\n", errRate*100))

	p.textView.SetText(sb.String())
}

// Primitive returns the underlying tview primitive for layout embedding.
func (p *TargetPanel) Primitive() tview.Primitive {
	return p.textView
}
