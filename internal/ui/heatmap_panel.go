package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"

	"github.com/yourorg/grpcmon/internal/probe"
)

const (
	heatCold    = "[white]"
	heatWarm    = "[yellow]"
	heatHot     = "[red]"
	heatError   = "[darkred]"
	heatSymbol  = "█"
)

// HeatmapPanel renders a per-target latency heat map as coloured block glyphs.
type HeatmapPanel struct {
	view  *tview.TextView
	store *probe.HeatmapStore
}

// NewHeatmapPanel creates a HeatmapPanel backed by store.
func NewHeatmapPanel(store *probe.HeatmapStore) *HeatmapPanel {
	v := tview.NewTextView()
	v.SetTitle(" Latency Heatmap ")
	v.SetBorder(true)
	v.SetDynamicColors(true)
	return &HeatmapPanel{view: v, store: store}
}

// Primitive returns the underlying tview widget.
func (p *HeatmapPanel) Primitive() tview.Primitive { return p.view }

// Update refreshes the heatmap display for all known targets.
func (p *HeatmapPanel) Update() {
	targets := p.store.Targets()
	if len(targets) == 0 {
		p.view.SetText("[grey]no data[-]")
		return
	}

	var sb strings.Builder
	for _, target := range targets {
		buckets := p.store.Buckets(target)
		sb.WriteString(fmt.Sprintf("[white]%-20s[-] ", truncateHeat(target, 20)))
		for _, b := range buckets {
			sb.WriteString(heatColor(b))
			sb.WriteString(heatSymbol)
		}
		sb.WriteString("[-]\n")
	}
	p.view.SetText(sb.String())
}

func heatColor(b probe.HeatmapBucket) string {
	if b.ErrorRate > 0.5 {
		return heatError
	}
	switch {
	case b.AvgLatency >= 200*time.Millisecond:
		return heatHot
	case b.AvgLatency >= 80*time.Millisecond:
		return heatWarm
	default:
		return heatCold
	}
}

func truncateHeat(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
