package ui

import (
	"fmt"
	"strings"

	"github.com/andrebq/grpcmon/internal/probe"
)

// SnapshotPanel renders a summary view of the latest probe snapshot.
type SnapshotPanel struct {
	store  *probe.SnapshotStore
	width  int
	height int
}

// NewSnapshotPanel creates a SnapshotPanel backed by the given store.
func NewSnapshotPanel(store *probe.SnapshotStore, width, height int) *SnapshotPanel {
	return &SnapshotPanel{store: store, width: width, height: height}
}

// Title returns the panel title.
func (p *SnapshotPanel) Title() string {
	return "Snapshot"
}

// Render returns a string representation of the latest snapshot.
func (p *SnapshotPanel) Render() string {
	snap := p.store.Current()
	if snap == nil {
		return centerLine("No snapshot available", p.width)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Captured: %s\n", snap.CapturedAt.Format("15:04:05")))

	changed := p.store.Diff()
	changedSet := make(map[string]bool, len(changed))
	for _, addr := range changed {
		changedSet[addr] = true
	}

	for _, t := range snap.Targets {
		marker := " "
		if changedSet[t.Address] {
			marker = "*"
		}
		line := fmt.Sprintf("%s %-30s %-12s lat=%-8s err=%.1f%%",
			marker,
			truncate(t.Address, 30),
			t.Status,
			t.AvgLatency.Round(1000000),
			t.ErrorRate*100,
		)
		sb.WriteString(line + "\n")
	}
	return sb.String()
}

func centerLine(s string, width int) string {
	if width <= len(s) {
		return s
	}
	pad := (width - len(s)) / 2
	return strings.Repeat(" ", pad) + s
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
