package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/yourorg/grpcmon/internal/probe"
)

// AnomalyPanel displays recently detected anomalies.
type AnomalyPanel struct {
	view     *tview.TextView
	max      int
	entries  []probe.Anomaly
}

// NewAnomalyPanel creates a panel that shows up to max anomaly entries.
func NewAnomalyPanel(max int) *AnomalyPanel {
	tv := tview.NewTextView()
	tv.SetBorder(true)
	tv.SetTitle(" Anomalies ")
	tv.SetDynamicColors(true)
	tv.SetScrollable(true)
	return &AnomalyPanel{view: tv, max: max}
}

// Update replaces the displayed anomalies with the provided slice.
func (p *AnomalyPanel) Update(anomalies []probe.Anomaly) {
	p.entries = anomalies
	p.render()
}

// Append adds new anomalies, evicting oldest entries beyond max.
func (p *AnomalyPanel) Append(anomalies []probe.Anomaly) {
	p.entries = append(p.entries, anomalies...)
	if len(p.entries) > p.max {
		p.entries = p.entries[len(p.entries)-p.max:]
	}
	p.render()
}

func (p *AnomalyPanel) render() {
	if len(p.entries) == 0 {
		p.view.SetText("[gray]no anomalies detected[-]")
		return
	}
	var sb strings.Builder
	for _, a := range p.entries {
		color := "yellow"
		if a.Kind == probe.AnomalyErrorBurst {
			color = "red"
		}
		ts := a.DetectedAt.Format("15:04:05")
		fmt.Fprintf(&sb, "[%s]%-14s[-] [white]%s[-] %s\n", color, string(a.Kind), a.Target, ts)
		fmt.Fprintf(&sb, "  [gray]%s[-]\n", a.Message)
	}
	p.view.SetText(sb.String())
}

// Primitive returns the underlying tview primitive for layout embedding.
func (p *AnomalyPanel) Primitive() tview.Primitive {
	return p.view
}
