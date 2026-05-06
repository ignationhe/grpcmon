package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/yourorg/grpcmon/internal/probe"
)

// HealthScorePanel renders a composite health score for each target.
type HealthScorePanel struct {
	view   *tview.TextView
	scorer *probe.HealthScorer
}

// NewHealthScorePanel creates a panel backed by the given scorer.
func NewHealthScorePanel(scorer *probe.HealthScorer) *HealthScorePanel {
	v := tview.NewTextView()
	v.SetDynamicColors(true)
	v.SetBorder(true)
	v.SetTitle(" Health Scores ")
	return &HealthScorePanel{view: v, scorer: scorer}
}

// Primitive returns the underlying tview primitive.
func (p *HealthScorePanel) Primitive() tview.Primitive { return p.view }

// Refresh recomputes and renders all health scores.
func (p *HealthScorePanel) Refresh() {
	scores := p.scorer.All()
	if len(scores) == 0 {
		p.view.SetText("[gray]no targets[-]")
		return
	}

	var sb strings.Builder
	for _, s := range scores {
		bar := scoreBar(s.Score)
		color := scoreColor(s.Score)
		fmt.Fprintf(&sb, "[white]%-30s[%s]%s [white]%.0f%%[-]\n",
			truncateScore(s.Target, 30),
			color, bar,
			s.Score*100,
		)
	}
	p.view.SetText(sb.String())
}

func scoreBar(score float64) string {
	const width = 10
	filled := int(score * float64(width))
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func scoreColor(score float64) string {
	switch {
	case score >= 0.9:
		return "green"
	case score >= 0.6:
		return "yellow"
	default:
		return "red"
	}
}

func truncateScore(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
