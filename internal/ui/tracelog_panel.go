package ui

import (
	"fmt"
	"strings"

	"github.com/andrebq/grpcmon/internal/probe"
)

const traceLogMaxRows = 12

// TraceLogPanel renders recent probe trace entries in the terminal dashboard.
type TraceLogPanel struct {
	log    *probe.TraceLog
	target string
	width  int
}

// NewTraceLogPanel creates a panel that shows traces for the given target.
func NewTraceLogPanel(log *probe.TraceLog, target string, width int) *TraceLogPanel {
	if width <= 0 {
		width = 60
	}
	return &TraceLogPanel{log: log, target: target, width: width}
}

// Title returns the panel heading.
func (p *TraceLogPanel) Title() string {
	return fmt.Sprintf("Trace Log — %s", p.target)
}

// Render returns the panel content as a slice of display lines.
func (p *TraceLogPanel) Render() []string {
	entries := p.log.ForTarget(p.target)
	if len(entries) == 0 {
		return []string{"  no trace entries"}
	}
	// Show most recent first, up to traceLogMaxRows.
	start := 0
	if len(entries) > traceLogMaxRows {
		start = len(entries) - traceLogMaxRows
	}
	recent := entries[start:]
	lines := make([]string, 0, len(recent)+1)
	header := fmt.Sprintf("  %-24s %-14s %s", "Time", "Status", "Latency")
	lines = append(lines, header)
	lines = append(lines, "  "+strings.Repeat("-", p.width-4))
	for i := len(recent) - 1; i >= 0; i-- {
		e := recent[i]
		ts := e.Timestamp.Format("15:04:05.000")
		latency := fmt.Sprintf("%7.2fms", float64(e.Duration.Microseconds())/1000.0)
		line := fmt.Sprintf("  %-24s %-14s %s", ts, e.Status, latency)
		lines = append(lines, line)
	}
	return lines
}
