package ui_test

import (
	"strings"
	"testing"
	"time"

	"github.com/yourorg/grpcmon/internal/probe"
	"github.com/yourorg/grpcmon/internal/ui"
)

func buildP99Tracker(target string, latencies []time.Duration) *probe.P99Tracker {
	t := probe.NewP99Tracker(50, 30*time.Second)
	now := time.Now()
	for i, l := range latencies {
		t.Record(probe.Result{
			Target:    target,
			Status:    probe.Serving,
			Latency:   l,
			CheckedAt: now.Add(time.Duration(i) * time.Millisecond),
		})
	}
	return t
}

func TestP99Panel_Title(t *testing.T) {
	tracker := probe.NewP99Tracker(50, 30*time.Second)
	p := ui.NewP99Panel(tracker)
	if p.Title() != "Latency Percentiles" {
		t.Errorf("unexpected title: %q", p.Title())
	}
}

func TestP99Panel_NoData(t *testing.T) {
	tracker := probe.NewP99Tracker(50, 30*time.Second)
	p := ui.NewP99Panel(tracker)
	out := p.Render()
	if !strings.Contains(out, "no data") {
		t.Errorf("expected 'no data', got: %q", out)
	}
}

func TestP99Panel_ShowsTarget(t *testing.T) {
	latencies := make([]time.Duration, 20)
	for i := range latencies {
		latencies[i] = time.Duration(i+1) * 10 * time.Millisecond
	}
	tracker := buildP99Tracker("api:443", latencies)
	p := ui.NewP99Panel(tracker)
	out := p.Render()
	if !strings.Contains(out, "api:443") {
		t.Errorf("expected target name in output, got: %q", out)
	}
}

func TestP99Panel_ShowsPercentileColumns(t *testing.T) {
	latencies := make([]time.Duration, 20)
	for i := range latencies {
		latencies[i] = time.Duration(i+1) * 5 * time.Millisecond
	}
	tracker := buildP99Tracker("svc:8080", latencies)
	p := ui.NewP99Panel(tracker)
	out := p.Render()
	for _, col := range []string{"P50", "P95", "P99"} {
		if !strings.Contains(out, col) {
			t.Errorf("expected column %q in output", col)
		}
	}
}

func TestP99Panel_TruncatesLongTarget(t *testing.T) {
	long := strings.Repeat("x", 40) + ":443"
	latencies := []time.Duration{10 * time.Millisecond, 20 * time.Millisecond}
	tracker := buildP99Tracker(long, latencies)
	p := ui.NewP99Panel(tracker)
	out := p.Render()
	if strings.Contains(out, long) {
		t.Errorf("expected long target to be truncated in output")
	}
	if !strings.Contains(out, "…") {
		t.Errorf("expected truncation indicator '…' in output")
	}
}
