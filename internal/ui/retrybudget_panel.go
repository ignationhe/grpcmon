package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"
)

// RetryBudgetSource is satisfied by probe.RetryBudget.
type RetryBudgetSource interface {
	Remaining(target string) int
	Targets() []string
}

// RetryBudgetPanel displays per-target retry budget remaining.
type RetryBudgetPanel struct {
	tv     *tview.TextView
	source RetryBudgetSource
	max    int
}

// NewRetryBudgetPanel creates a panel backed by the given RetryBudgetSource.
func NewRetryBudgetPanel(source RetryBudgetSource, maxRetries int) *RetryBudgetPanel {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle(" Retry Budget ")
	return &RetryBudgetPanel{tv: tv, source: source, max: maxRetries}
}

// Primitive returns the underlying tview widget.
func (p *RetryBudgetPanel) Primitive() tview.Primitive {
	return p.tv
}

// Update refreshes the panel contents.
func (p *RetryBudgetPanel) Update() {
	targets := p.source.Targets()
	sort.Strings(targets)

	var sb strings.Builder
	if len(targets) == 0 {
		sb.WriteString("[grey]no retry activity[-]\n")
	}

	for _, t := range targets {
		remaining := p.source.Remaining(t)
		color := retryColor(remaining, p.max)
		bar := retryBar(remaining, p.max)
		fmt.Fprintf(&sb, "[white]%-28s[%s]%s[-] %d/%d\n",
			truncateRetryTarget(t, 28), color, bar, remaining, p.max)
	}

	p.tv.SetText(sb.String())
}

func retryColor(remaining, max int) string {
	if max == 0 {
		return "grey"
	}
	ratio := float64(remaining) / float64(max)
	switch {
	case ratio > 0.6:
		return "green"
	case ratio > 0.2:
		return "yellow"
	default:
		return "red"
	}
}

func retryBar(remaining, max int) string {
	const width = 10
	if max == 0 {
		return strings.Repeat("-", width)
	}
	filled := remaining * width / max
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func truncateRetryTarget(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
