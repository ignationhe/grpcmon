package ui

import (
	"fmt"
	"sort"
	"strings"
)

// ConcurrencySource is the interface the panel reads from.
type ConcurrencySource interface {
	Targets() []string
	Inflight(target string) int
	Peak(target string) int
}

// ConcurrencyPanel renders a table of per-target in-flight probe counts
// and peak concurrency for the current window.
type ConcurrencyPanel struct {
	src   ConcurrencySource
	width int
}

// NewConcurrencyPanel creates a ConcurrencyPanel backed by src.
func NewConcurrencyPanel(src ConcurrencySource, width int) *ConcurrencyPanel {
	if width <= 0 {
		width = 60
	}
	return &ConcurrencyPanel{src: src, width: width}
}

// Title returns the panel heading.
func (p *ConcurrencyPanel) Title() string {
	return "Concurrency"
}

// Render returns the panel content as a string.
func (p *ConcurrencyPanel) Render() string {
	targets := p.src.Targets()
	if len(targets) == 0 {
		return p.frame("  no targets")
	}
	sort.Strings(targets)

	var sb strings.Builder
	hdr := fmt.Sprintf("  %-30s %8s %8s", "Target", "Inflight", "Peak")
	sb.WriteString(hdr + "\n")
	sb.WriteString("  " + strings.Repeat("-", p.width-4) + "\n")

	for _, t := range targets {
		inflight := p.src.Inflight(t)
		peak := p.src.Peak(t)
		display := t
		if len(display) > 30 {
			display = display[:27] + "..."
		}
		line := fmt.Sprintf("  %-30s %8d %8d", display, inflight, peak)
		sb.WriteString(line + "\n")
	}
	return p.frame(sb.String())
}

func (p *ConcurrencyPanel) frame(body string) string {
	border := strings.Repeat("─", p.width-2)
	return fmt.Sprintf("┌%s┐\n│ %-*s│\n├%s┤\n%s└%s┘",
		border,
		p.width-3, p.Title(),
		border,
		body,
		border,
	)
}
