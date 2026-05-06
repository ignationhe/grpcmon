package ui

import (
	"fmt"
	"sort"

	"github.com/rivo/tview"
)

// QuotaEntry holds display data for a single target's quota state.
type QuotaEntry struct {
	Target    string
	Used      int
	Max       int
	Remaining int
}

// QuotaPanel renders a table showing per-target probe quota consumption.
type QuotaPanel struct {
	view *tview.TextView
}

// NewQuotaPanel creates a new QuotaPanel.
func NewQuotaPanel() *QuotaPanel {
	tv := tview.NewTextView()
	tv.SetBorder(true)
	tv.SetTitle(" Probe Quota ")
	tv.SetDynamicColors(true)
	return &QuotaPanel{view: tv}
}

// Update refreshes the panel with the provided quota entries.
func (p *QuotaPanel) Update(entries []QuotaEntry) {
	if len(entries) == 0 {
		p.view.SetText("[grey]no quota data[-]")
		return
	}

	sorted := make([]QuotaEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Target < sorted[j].Target
	})

	var out string
	out += fmt.Sprintf("[yellow]%-30s %5s %5s %9s[-]\n", "TARGET", "USED", "MAX", "REMAINING")
	for _, e := range sorted {
		color := quotaColor(e.Remaining, e.Max)
		target := e.Target
		if len(target) > 30 {
			target = target[:27] + "..."
		}
		out += fmt.Sprintf("%s%-30s %5d %5d %9d[-]\n",
			color, target, e.Used, e.Max, e.Remaining)
	}
	p.view.SetText(out)
}

// Primitive returns the underlying tview primitive for layout embedding.
func (p *QuotaPanel) Primitive() tview.Primitive {
	return p.view
}

func quotaColor(remaining, max int) string {
	if max <= 0 {
		return "[white]"
	}
	ratio := float64(remaining) / float64(max)
	switch {
	case ratio > 0.5:
		return "[green]"
	case ratio > 0.2:
		return "[yellow]"
	default:
		return "[red]"
	}
}
