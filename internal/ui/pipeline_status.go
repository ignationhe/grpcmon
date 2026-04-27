package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

// PipelineStatus renders a compact summary of pipeline-level states
// (throttled, circuit-open) alongside normal probe statuses.
type PipelineStatus struct {
	view *tview.TextView
}

// PipelineEntry holds the display state for a single target in the pipeline view.
type PipelineEntry struct {
	Target   string
	Status   string
	Inflight int
}

// NewPipelineStatus creates a new PipelineStatus widget.
func NewPipelineStatus() *PipelineStatus {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle(" Pipeline ")
	return &PipelineStatus{view: tv}
}

// Update refreshes the widget with current pipeline entries.
func (ps *PipelineStatus) Update(entries []PipelineEntry) {
	var sb strings.Builder
	for _, e := range entries {
		color := statusColor(e.Status)
		inflight := ""
		if e.Inflight > 0 {
			inflight = fmt.Sprintf(" [yellow](%d inflight)[-]", e.Inflight)
		}
		fmt.Fprintf(&sb, "[%s]%-6s[-] %s%s\n", color, e.Status, e.Target, inflight)
	}
	ps.view.SetText(sb.String())
}

// Primitive returns the underlying tview primitive for layout embedding.
func (ps *PipelineStatus) Primitive() tview.Primitive {
	return ps.view
}

func statusColor(status string) string {
	switch status {
	case "SERVING":
		return "green"
	case "THROTTLED":
		return "yellow"
	case "OPEN":
		return "orange"
	case "NOT_SERVING":
		return "red"
	default:
		return "white"
	}
}
