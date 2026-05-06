package ui

import (
	"fmt"
	"sort"
	"time"

	"github.com/rivo/tview"
)

// CheckpointSource is satisfied by probe.Checkpoint.
type CheckpointSource interface {
	LastSeen(target string) (time.Time, bool)
	IsStale(target string, now time.Time) bool
}

// CheckpointPanel renders the last-seen timestamp for each target
// and highlights targets that are currently stale.
type CheckpointPanel struct {
	tv      *tview.TextView
	source  CheckpointSource
	targets []string
}

// NewCheckpointPanel creates a panel backed by the given CheckpointSource.
func NewCheckpointPanel(source CheckpointSource, targets []string) *CheckpointPanel {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle(" Checkpoints ")
	return &CheckpointPanel{tv: tv, source: source, targets: targets}
}

// Update refreshes the panel content based on the current time.
func (p *CheckpointPanel) Update(now time.Time) {
	p.tv.Clear()

	keys := make([]string, len(p.targets))
	copy(keys, p.targets)
	sort.Strings(keys)

	if len(keys) == 0 {
		fmt.Fprint(p.tv, "[gray]no targets[-]")
		return
	}

	for _, tgt := range keys {
		last, ok := p.source.LastSeen(tgt)
		var age string
		var color string
		if !ok {
			age = "never"
			color = "red"
		} else {
			d := now.Sub(last).Truncate(time.Second)
			age = d.String() + " ago"
			if p.source.IsStale(tgt, now) {
				color = "yellow"
			} else {
				color = "green"
			}
		}
		fmt.Fprintf(p.tv, "[%s]%-30s[-] %s\n", color, truncateCP(tgt, 30), age)
	}
}

// Primitive returns the underlying tview widget.
func (p *CheckpointPanel) Primitive() tview.Primitive {
	return p.tv
}

func truncateCP(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
