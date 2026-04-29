package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/andrebq/grpcmon/internal/probe"
)

func addTrace(log *probe.TraceLog, target, status string, dur time.Duration) {
	log.Add(probe.TraceEntry{
		Target:    target,
		Timestamp: time.Now(),
		Duration:  dur,
		Status:    status,
		Message:   "",
	})
}

func TestTraceLogPanel_Title(t *testing.T) {
	log := probe.NewTraceLog(50)
	p := NewTraceLogPanel(log, "svc:50051", 60)
	if !strings.Contains(p.Title(), "svc:50051") {
		t.Errorf("title missing target: %s", p.Title())
	}
}

func TestTraceLogPanel_NoEntries(t *testing.T) {
	log := probe.NewTraceLog(50)
	p := NewTraceLogPanel(log, "svc:50051", 60)
	lines := p.Render()
	if len(lines) == 0 {
		t.Fatal("expected at least one line for empty state")
	}
	if !strings.Contains(lines[0], "no trace") {
		t.Errorf("expected empty message, got: %s", lines[0])
	}
}

func TestTraceLogPanel_ShowsEntries(t *testing.T) {
	log := probe.NewTraceLog(50)
	addTrace(log, "svc:50051", "SERVING", 12*time.Millisecond)
	addTrace(log, "svc:50051", "NOT_SERVING", 5*time.Millisecond)
	p := NewTraceLogPanel(log, "svc:50051", 80)
	lines := p.Render()
	full := strings.Join(lines, "\n")
	if !strings.Contains(full, "SERVING") {
		t.Errorf("expected SERVING in output:\n%s", full)
	}
	if !strings.Contains(full, "NOT_SERVING") {
		t.Errorf("expected NOT_SERVING in output:\n%s", full)
	}
}

func TestTraceLogPanel_FiltersOtherTargets(t *testing.T) {
	log := probe.NewTraceLog(50)
	addTrace(log, "svc:50051", "SERVING", 10*time.Millisecond)
	addTrace(log, "other:9090", "NOT_SERVING", 10*time.Millisecond)
	p := NewTraceLogPanel(log, "svc:50051", 80)
	lines := p.Render()
	full := strings.Join(lines, "\n")
	if strings.Contains(full, "other:9090") {
		t.Errorf("panel should not show entries for other targets")
	}
}

func TestTraceLogPanel_MaxRows(t *testing.T) {
	log := probe.NewTraceLog(200)
	for i := 0; i < 20; i++ {
		addTrace(log, "svc:50051", "SERVING", time.Millisecond)
	}
	p := NewTraceLogPanel(log, "svc:50051", 80)
	lines := p.Render()
	// header + separator + up to traceLogMaxRows data lines
	if len(lines) > traceLogMaxRows+2 {
		t.Errorf("too many lines rendered: %d", len(lines))
	}
}
