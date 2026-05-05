package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/yourorg/grpcmon/internal/probe"
)

const (
	arrowStable    = "→"
	arrowImproving = "↓"
	arrowDegrading = "↑"
)

// TrendlinePanel renders per-target latency trend directions.
type TrendlinePanel struct {
	view     *tview.TextView
	analyzer *probe.TrendAnalyzer
}

// NewTrendlinePanel creates a panel that uses the given TrendAnalyzer.
func NewTrendlinePanel(a *probe.TrendAnalyzer) *TrendlinePanel {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle(" Latency Trends ")
	return &TrendlinePanel{view: tv, analyzer: a}
}

// Primitive returns the underlying tview primitive for layout embedding.
func (p *TrendlinePanel) Primitive() tview.Primitive { return p.view }

// Update refreshes the panel using the provided history map keyed by target.
func (p *TrendlinePanel) Update(histories map[string]*probe.History) {
	if len(histories) == 0 {
		p.view.SetText("[grey]no data[-]")
		return
	}

	var sb strings.Builder
	for target, h := range histories {
		r := p.analyzer.Analyze(target, h)
		arrow, color := directionStyle(r.Direction)
		fmt.Fprintf(&sb, "[%s]%s[-] %-30s slope: %+.1f ms/s\n",
			color, arrow, truncateTrend(target, 30), r.Slope/1e6)
	}
	p.view.SetText(sb.String())
}

func directionStyle(d probe.TrendDirection) (string, string) {
	switch d {
	case probe.TrendDegrading:
		return arrowDegrading, "red"
	case probe.TrendImproving:
		return arrowImproving, "green"
	default:
		return arrowStable, "yellow"
	}
}

func truncateTrend(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
