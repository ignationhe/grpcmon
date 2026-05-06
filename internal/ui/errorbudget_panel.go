package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"github.com/yourorg/grpcmon/internal/probe"
)

// ErrorBudgetPanel renders error budget consumption for all targets.
type ErrorBudgetPanel struct {
	view    *tview.TextView
	tracker *probe.ErrorBudgetTracker
}

// NewErrorBudgetPanel creates a panel backed by the given tracker.
func NewErrorBudgetPanel(tracker *probe.ErrorBudgetTracker) *ErrorBudgetPanel {
	v := tview.NewTextView()
	v.SetBorder(true)
	v.SetTitle(" Error Budget ")
	v.SetDynamicColors(true)
	return &ErrorBudgetPanel{view: v, tracker: tracker}
}

// Update re-evaluates and refreshes the panel content.
func (p *ErrorBudgetPanel) Update() {
	p.tracker.Evaluate()

	entries := p.tracker.All()
	if len(entries) == 0 {
		p.view.SetText("[grey]No data[-]")
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Target < entries[j].Target
	})

	var sb strings.Builder
	for _, e := range entries {
		color := budgetColor(e.BudgetUsed)
		pct := e.BudgetUsed * 100
		bar := budgetBar(e.BudgetUsed, 12)
		marker := ""
		if e.Exhausted {
			marker = " [red]EXHAUSTED[-]"
		}
		fmt.Fprintf(&sb, "[white]%-22s[-] [%s]%s %5.1f%%[-] burn=[yellow]%.3f/s[-]%s\n",
			truncateBudgetTarget(e.Target, 22),
			color, bar, pct,
			e.BurnRate,
			marker,
		)
	}
	p.view.SetText(sb.String())
}

// View returns the underlying tview primitive.
func (p *ErrorBudgetPanel) View() *tview.TextView { return p.view }

func budgetColor(used float64) string {
	switch {
	case used >= 1.0:
		return "red"
	case used >= 0.75:
		return "orange"
	case used >= 0.5:
		return "yellow"
	default:
		return "green"
	}
}

func budgetBar(used float64, width int) string {
	filled := int(used * float64(width))
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
