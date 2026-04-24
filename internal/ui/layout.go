package ui

import (
	"github.com/rivo/tview"
)

// Layout manages the top-level tview arrangement for grpcmon.
type Layout struct {
	app      *tview.Application
	grid     *tview.Grid
	panels   []*TargetPanel
	statusBar *StatusBar
}

// NewLayout creates a grid layout with target panels and a status bar.
func NewLayout(app *tview.Application, targets []string) *Layout {
	grid := tview.NewGrid().
		SetBorders(false)

	panels := make([]*TargetPanel, len(targets))
	for i, t := range targets {
		panels[i] = NewTargetPanel(t)
	}

	sb := NewStatusBar()

	l := &Layout{
		app:       app,
		grid:      grid,
		panels:    panels,
		statusBar: sb,
	}
	l.build()
	return l
}

// build arranges panels in a two-column grid with the status bar at the bottom.
func (l *Layout) build() {
	n := len(l.panels)
	rows := (n + 1) / 2

	rowSizes := make([]int, rows+1)
	for i := 0; i < rows; i++ {
		rowSizes[i] = 0 // flexible
	}
	rowSizes[rows] = 1 // fixed status bar

	l.grid.SetRows(rowSizes...).SetColumns(0, 0)

	for i, p := range l.panels {
		row := i / 2
		col := i % 2
		l.grid.AddItem(p.Primitive(), row, col, 1, 1, 0, 0, false)
	}

	l.grid.AddItem(l.statusBar.Primitive(), rows, 0, 1, 2, 0, 0, false)
}

// Root returns the root primitive to pass to tview.Application.SetRoot.
func (l *Layout) Root() tview.Primitive {
	return l.grid
}

// Panels returns the target panels for external updates.
func (l *Layout) Panels() []*TargetPanel {
	return l.panels
}

// StatusBar returns the status bar for external updates.
func (l *Layout) StatusBar() *StatusBar {
	return l.statusBar
}
