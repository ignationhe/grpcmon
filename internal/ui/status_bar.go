package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"
)

// StatusBar renders a bottom bar showing last refresh time and overall health summary.
type StatusBar struct {
	view *tview.TextView
}

// NewStatusBar creates a new StatusBar widget.
func NewStatusBar() *StatusBar {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(false)
	return &StatusBar{view: tv}
}

// Update refreshes the status bar with the given counts of serving, degraded,
// and total targets, plus the timestamp of the last poll.
func (s *StatusBar) Update(serving, notServing, total int, lastPoll time.Time) {
	var parts []string

	if notServing == 0 && total > 0 {
		parts = append(parts, fmt.Sprintf("[green]● All %d targets healthy[-]", total))
	} else if notServing > 0 {
		parts = append(parts, fmt.Sprintf("[red]● %d/%d targets unhealthy[-]", notServing, total))
	} else {
		parts = append(parts, "[yellow]● No targets[-]")
	}

	if !lastPoll.IsZero() {
		parts = append(parts, fmt.Sprintf("Last poll: [white]%s[-]", lastPoll.Format("15:04:05")))
	}

	s.view.SetText(" " + strings.Join(parts, "  │  "))
}

// View returns the underlying tview primitive.
func (s *StatusBar) View() *tview.TextView {
	return s.view
}
