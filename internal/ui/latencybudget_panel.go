package ui

import (
	"fmt"
	"sort"
	"time"

	"github.com/rivo/tview"
)

// LatencyBudgetSource is satisfied by probe.LatencyBudget.
type LatencyBudgetSource interface {
	All() map[string]time.Duration
	Remaining(target string) time.Duration
	Exceeded(target string) bool
}

// LatencyBudgetPanel renders a tview table showing per-target latency budget
// consumption.
type LatencyBudgetPanel struct {
	table  *tview.Table
	source LatencyBudgetSource
}

// NewLatencyBudgetPanel creates a panel backed by source.
func NewLatencyBudgetPanel(source LatencyBudgetSource) *LatencyBudgetPanel {
	t := tview.NewTable()
	t.SetBorders(false)
	t.SetTitle(" Latency Budget ")
	t.SetBorder(true)
	return &LatencyBudgetPanel{table: t, source: source}
}

// Primitive returns the underlying tview primitive for embedding in a layout.
func (p *LatencyBudgetPanel) Primitive() tview.Primitive { return p.table }

// Update refreshes the table contents from the backing source.
func (p *LatencyBudgetPanel) Update() {
	p.table.Clear()
	headers := []string{"Target", "Spent", "Remaining", "Status"}
	for col, h := range headers {
		p.table.SetCell(0, col, tview.NewTableCell(h).SetTextColor(headerColor))
	}

	all := p.source.All()
	targets := make([]string, 0, len(all))
	for t := range all {
		targets = append(targets, t)
	}
	sort.Strings(targets)

	for row, tgt := range targets {
		spent := all[tgt]
		remaining := p.source.Remaining(tgt)
		exceeded := p.source.Exceeded(tgt)

		status := "OK"
		statusColor := successColor
		if exceeded {
			status = "EXCEEDED"
			statusColor = errorColor
		}

		p.table.SetCell(row+1, 0, tview.NewTableCell(truncateBudgetTarget(tgt, 30)))
		p.table.SetCell(row+1, 1, tview.NewTableCell(fmt.Sprintf("%v", spent.Round(time.Millisecond))))
		p.table.SetCell(row+1, 2, tview.NewTableCell(fmt.Sprintf("%v", remaining.Round(time.Millisecond))))
		p.table.SetCell(row+1, 3, tview.NewTableCell(status).SetTextColor(statusColor))
	}
}

func truncateBudgetTarget(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return "…" + s[len(s)-max+1:]
}
