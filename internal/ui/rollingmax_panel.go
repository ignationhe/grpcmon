package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rivo/tview"
)

// RollingMaxSource is satisfied by probe.RollingMax.
type RollingMaxSource interface {
	Max(target string) (time.Duration, bool)
	Targets() []string
}

// RollingMaxPanel renders per-target peak latency within a sliding window.
type RollingMaxPanel struct {
	view   *tview.TextView
	source RollingMaxSource
}

// NewRollingMaxPanel creates a panel backed by the given RollingMaxSource.
func NewRollingMaxPanel(source RollingMaxSource) *RollingMaxPanel {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle(" Peak Latency (window) ")
	return &RollingMaxPanel{view: tv, source: source}
}

// View returns the underlying tview primitive.
func (p *RollingMaxPanel) View() *tview.TextView { return p.view }

// Refresh redraws the panel with current rolling-max data.
func (p *RollingMaxPanel) Refresh() {
	targets := p.source.Targets()
	if len(targets) == 0 {
		p.view.SetText("[grey]no data[-]")
		return
	}
	sort.Strings(targets)

	var sb strings.Builder
	for _, t := range targets {
		max, ok := p.source.Max(t)
		if !ok {
			continue
		}
		color := peakColor(max)
		short := truncatePeakTarget(t, 28)
		fmt.Fprintf(&sb, "[white]%-28s[%s]%8s[-]\n", short, color, fmtPeak(max))
	}
	p.view.SetText(sb.String())
}

func peakColor(d time.Duration) string {
	switch {
	case d < 100*time.Millisecond:
		return "green"
	case d < 500*time.Millisecond:
		return "yellow"
	default:
		return "red"
	}
}

func fmtPeak(d time.Duration) string {
	if d >= time.Second {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

func truncatePeakTarget(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
