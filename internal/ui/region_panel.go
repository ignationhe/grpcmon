package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/yourorg/grpcmon/internal/probe"
)

// RegionPanel renders a per-region health summary table.
type RegionPanel struct {
	view  *tview.TextView
	store *probe.RegionStore
	agg   *probe.Aggregator
}

// NewRegionPanel creates a RegionPanel backed by the given store and aggregator.
func NewRegionPanel(store *probe.RegionStore, agg *probe.Aggregator) *RegionPanel {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle(" Regions ")
	return &RegionPanel{view: tv, store: store, agg: agg}
}

// Primitive returns the underlying tview primitive.
func (p *RegionPanel) Primitive() tview.Primitive { return p.view }

// Refresh rebuilds the panel content from current state.
func (p *RegionPanel) Refresh() {
	regions := p.store.Regions()
	if len(regions) == 0 {
		p.view.SetText("[gray]no regions assigned[-]")
		return
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "[yellow]%-18s %7s %7s %8s[-]\n",
		"REGION", "HEALTHY", "UNHEALTHY", "ERR RATE")

	for _, reg := range regions {
		s := p.store.Summarise(reg, p.agg)
		color := regionColor(s.ErrorRate)
		fmt.Fprintf(&sb, "[%s]%-18s %7d %9d %7.1f%%[-]\n",
			color, reg, s.Healthy, s.Unhealthy, s.ErrorRate*100)
	}
	p.view.SetText(sb.String())
}

// WorstRegion returns the region name with the highest error rate, or an empty
// string if no regions are available. It can be used to drive alerts or
// auto-scroll focus to the most degraded region.
func (p *RegionPanel) WorstRegion() string {
	regions := p.store.Regions()
	if len(regions) == 0 {
		return ""
	}
	worst := regions[0]
	worstRate := p.store.Summarise(worst, p.agg).ErrorRate
	for _, reg := range regions[1:] {
		if rate := p.store.Summarise(reg, p.agg).ErrorRate; rate > worstRate {
			worst = reg
			worstRate = rate
		}
	}
	return worst
}

// HealthyCount returns the total number of healthy endpoints across all regions.
func (p *RegionPanel) HealthyCount() int {
	total := 0
	for _, reg := range p.store.Regions() {
		total += p.store.Summarise(reg, p.agg).Healthy
	}
	return total
}

// UnhealthyCount returns the total number of unhealthy endpoints across all regions.
func (p *RegionPanel) UnhealthyCount() int {
	total := 0
	for _, reg := range p.store.Regions() {
		total += p.store.Summarise(reg, p.agg).Unhealthy
	}
	return total
}

func regionColor(rate float64) string {
	switch {
	case rate == 0:
		return "green"
	case rate < 0.5:
		return "yellow"
	default:
		return "red"
	}
}
