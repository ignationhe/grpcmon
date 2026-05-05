package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"

	"github.com/yourorg/grpcmon/internal/probe"
)

// WindowPanel renders rolling-window statistics for all tracked targets.
type WindowPanel struct {
	view *tview.TextView
	agg  *probe.WindowAggregator
}

// NewWindowPanel creates a WindowPanel backed by the given WindowAggregator.
func NewWindowPanel(agg *probe.WindowAggregator) *WindowPanel {
	v := tview.NewTextView()
	v.SetDynamicColors(true)
	v.SetBorder(true)
	v.SetTitle(" Rolling Window Stats ")
	return &WindowPanel{view: v, agg: agg}
}

// Update refreshes the panel content from the aggregator.
func (p *WindowPanel) Update() {
	var sb strings.Builder
	targets := p.agg.Targets()
	if len(targets) == 0 {
		fmt.Fprintln(&sb, "[grey]No data yet[-]")
		p.view.SetText(sb.String())
		return
	}
	for _, t := range targets {
		s := p.agg.Stats(t)
		errPct := s.ErrorRate * 100
		errColor := "green"
		if errPct > 10 {
			errColor = "yellow"
		}
		if errPct > 30 {
			errColor = "red"
		}
		fmt.Fprintf(&sb,
			"[white]%-30s[-] cnt=[cyan]%4d[-] err=[%s]%5.1f%%[-] avg=[cyan]%8s[-] p95=[cyan]%8s[-]\n",
			truncateTarget(t, 30),
			s.Count,
			errColor, errPct,
			fmtDur(s.AvgLatency),
			fmtDur(s.P95Latency),
		)
	}
	p.view.SetText(sb.String())
}

// Primitive returns the underlying tview primitive for layout embedding.
func (p *WindowPanel) Primitive() tview.Primitive {
	return p.view
}

func fmtDur(d time.Duration) string {
	if d == 0 {
		return "—"
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%.1fms", float64(d.Milliseconds()))
}
