package ui

import (
	"fmt"
	"sort"
	"time"

	"github.com/rivo/tview"

	"github.com/yourorg/grpcmon/internal/probe"
)

// DeadlinePanel renders per-target deadline configuration and breach counts.
type DeadlinePanel struct {
	view    *tview.Table
	tracker *probe.DeadlineTracker
}

// NewDeadlinePanel creates a DeadlinePanel backed by the given tracker.
func NewDeadlinePanel(tracker *probe.DeadlineTracker) *DeadlinePanel {
	table := tview.NewTable().SetBorders(false)
	table.SetTitle(" Deadlines ").SetBorder(true)
	return &DeadlinePanel{view: table, tracker: tracker}
}

// Primitive returns the underlying tview primitive for layout embedding.
func (p *DeadlinePanel) Primitive() tview.Primitive {
	return p.view
}

// Update refreshes the table with the latest deadline entries.
func (p *DeadlinePanel) Update() {
	p.view.Clear()

	headers := []string{"TARGET", "DEADLINE", "BREACHES", "LAST BREACH"}
	for col, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(headerColor).
			SetSelectable(false).
			SetExpansion(1)
		p.view.SetCell(0, col, cell)
	}

	all := p.tracker.All()
	targets := make([]string, 0, len(all))
	for t := range all {
		targets = append(targets, t)
	}
	sort.Strings(targets)

	for row, target := range targets {
		e := all[target]
		lastBreach := "-"
		if !e.LastBreach.IsZero() {
			lastBreach = e.LastBreach.Format(time.RFC3339)
		}
		breachColor := tcell.ColorWhite
		if e.BreachCount > 0 {
			breachColor = tcell.ColorRed
		}
		fields := []string{
			target,
			e.Deadline.String(),
			fmt.Sprintf("%d", e.BreachCount),
			lastBreach,
		}
		colors := []tcell.Color{
			tcell.ColorWhite,
			tcell.ColorYellow,
			breachColor,
			tcell.ColorGray,
		}
		for col, text := range fields {
			p.view.SetCell(row+1, col,
				tview.NewTableCell(text).
					SetTextColor(colors[col]).
					SetExpansion(1))
		}
	}
}
