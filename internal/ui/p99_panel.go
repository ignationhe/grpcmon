package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/yourorg/grpcmon/internal/probe"
)

// P99Panel renders a table of P50/P95/P99 latency percentiles per target.
type P99Panel struct {
	tracker *probe.P99Tracker
}

// NewP99Panel creates a panel backed by the given tracker.
func NewP99Panel(tracker *probe.P99Tracker) *P99Panel {
	return &P99Panel{tracker: tracker}
}

// Title returns the panel heading.
func (p *P99Panel) Title() string { return "Latency Percentiles" }

// Render returns a formatted string table of percentile data.
func (p *P99Panel) Render() string {
	targets := p.tracker.Targets()
	if len(targets) == 0 {
		return p.Title() + "\n" + "  (no data)\n"
	}
	sort.Strings(targets)

	var sb strings.Builder
	sb.WriteString(p.Title() + "\n")
	sb.WriteString(fmt.Sprintf("  %-28s %10s %10s %10s\n", "Target", "P50", "P95", "P99"))
	sb.WriteString("  " + strings.Repeat("-", 62) + "\n")

	for _, tgt := range targets {
		p50, _ := p.tracker.P50(tgt)
		p95, _ := p.tracker.P95(tgt)
		p99, _ := p.tracker.P99(tgt)

		sb.WriteString(fmt.Sprintf("  %-28s %10s %10s %10s\n",
			truncateP99Target(tgt, 28),
			fmtPercentile(p50),
			fmtPercentile(p95),
			fmtPercentile(p99),
		))
	}
	return sb.String()
}

func fmtPercentile(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000.0)
}

func truncateP99Target(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
