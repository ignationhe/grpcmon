package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/user/grpcmon/internal/probe"
)

// Row represents a single service row in the dashboard.
type Row struct {
	Target  string
	Status  string
	Latency time.Duration
	Updated time.Time
}

// Dashboard holds the state for the terminal UI.
type Dashboard struct {
	rows   map[string]*Row
	width  int
}

// New creates a new Dashboard with the given terminal width.
func New(width int) *Dashboard {
	if width <= 0 {
		width = 80
	}
	return &Dashboard{
		rows:  make(map[string]*Row),
		width: width,
	}
}

// Update applies a probe result to the dashboard state.
func (d *Dashboard) Update(r probe.Result) {
	status := "SERVING"
	if !r.Serving {
		if r.Err != nil {
			status = "ERROR"
		} else {
			status = "NOT_SERVING"
		}
	}
	d.rows[r.Target] = &Row{
		Target:  r.Target,
		Status:  status,
		Latency: r.Latency,
		Updated: r.Timestamp,
	}
}

// Render returns the full dashboard as a string.
func (d *Dashboard) Render() string {
	var sb strings.Builder
	sep := strings.Repeat("-", d.width)
	sb.WriteString(sep + "\n")
	sb.WriteString(fmt.Sprintf("%-40s %-14s %-12s %s\n", "TARGET", "STATUS", "LATENCY", "UPDATED"))
	sb.WriteString(sep + "\n")
	for _, row := range d.rows {
		sb.WriteString(fmt.Sprintf("%-40s %-14s %-12s %s\n",
			row.Target,
			row.Status,
			row.Latency.Round(time.Millisecond),
			row.Updated.Format("15:04:05"),
		))
	}
	sb.WriteString(sep + "\n")
	return sb.String()
}

// RowCount returns the number of tracked services.
func (d *Dashboard) RowCount() int {
	return len(d.rows)
}
