package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"grpcmon/internal/probe"
)

// SLAPanel displays current SLA violation status for all monitored targets.
type SLAPanel struct {
	view *tview.TextView
}

// NewSLAPanel creates a new SLA violations panel.
func NewSLAPanel() *SLAPanel {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle(" SLA Violations ")
	return &SLAPanel{view: tv}
}

// Update refreshes the panel with the latest violation data.
func (p *SLAPanel) Update(violations map[string][]probe.SLAViolation) {
	var sb strings.Builder

	if len(violations) == 0 {
		sb.WriteString("[green]✓ All targets within SLA[-]\n")
		p.view.SetText(sb.String())
		return
	}

	for target, vs := range violations {
		if len(vs) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("[yellow]%s[-]\n", target))
		for _, v := range vs {
			icon := "⚠"
			sb.WriteString(fmt.Sprintf("  [red]%s[-] %s: got [red]%s[-] (limit %s)\n",
				icon, v.Kind, v.Actual, v.Threshold))
		}
	}

	p.view.SetText(sb.String())
}

// Primitive returns the underlying tview primitive for layout embedding.
func (p *SLAPanel) Primitive() tview.Primitive {
	return p.view
}
