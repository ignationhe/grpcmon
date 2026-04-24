package ui

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

const (
	graphWidth   = 20
	graphHeight  = 5
	sparkWidth   = graphWidth
)

// LatencyGraph is a tview TextView that displays a labelled sparkline for a
// single target's recent latency history.
type LatencyGraph struct {
	view      *tview.TextView
	sparkline *Sparkline
	target    string
}

// NewLatencyGraph creates a LatencyGraph for the named target.
func NewLatencyGraph(target string) *LatencyGraph {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle(fmt.Sprintf(" %s ", target))

	return &LatencyGraph{
		view:      tv,
		sparkline: NewSparkline(sparkWidth),
		target:    target,
	}
}

// Update refreshes the sparkline with the provided latency samples.
func (g *LatencyGraph) Update(latencies []time.Duration) {
	values := make([]float64, len(latencies))
	for i, d := range latencies {
		values[i] = float64(d.Milliseconds())
	}

	line := g.sparkline.Render(values)

	var avgMs float64
	if len(latencies) > 0 {
		var sum time.Duration
		for _, d := range latencies {
			sum += d
		}
		avgMs = float64(sum.Milliseconds()) / float64(len(latencies))
	}

	g.view.SetText(fmt.Sprintf("%s\n[yellow]avg: %.1fms[-]", line, avgMs))
}

// View returns the underlying tview primitive.
func (g *LatencyGraph) View() *tview.TextView {
	return g.view
}
