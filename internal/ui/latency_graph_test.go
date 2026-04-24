package ui

import (
	"strings"
	"testing"
	"time"
)

func TestLatencyGraph_New(t *testing.T) {
	g := NewLatencyGraph("localhost:50051")
	if g == nil {
		t.Fatal("expected non-nil LatencyGraph")
	}
	if g.View() == nil {
		t.Fatal("expected non-nil tview.TextView")
	}
}

func TestLatencyGraph_Update_Empty(t *testing.T) {
	g := NewLatencyGraph("svc")
	// Should not panic with empty slice.
	g.Update(nil)
	txt := g.view.GetText(false)
	if txt == "" {
		t.Fatal("expected non-empty text after Update with nil")
	}
}

func TestLatencyGraph_Update_ContainsSparkline(t *testing.T) {
	g := NewLatencyGraph("svc")
	latencies := []time.Duration{
		5 * time.Millisecond,
		10 * time.Millisecond,
		15 * time.Millisecond,
		20 * time.Millisecond,
	}
	g.Update(latencies)
	txt := g.view.GetText(false)
	// The text should contain at least one spark character.
	found := false
	for _, r := range sparkChars {
		if strings.ContainsRune(txt, r) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected sparkline chars in output, got: %q", txt)
	}
}

func TestLatencyGraph_Update_ContainsAvg(t *testing.T) {
	g := NewLatencyGraph("svc")
	latencies := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
	}
	g.Update(latencies)
	txt := g.view.GetText(false)
	// avg should be 15.0ms
	if !strings.Contains(txt, "15.0ms") {
		t.Fatalf("expected avg 15.0ms in output, got: %q", txt)
	}
}

func TestLatencyGraph_Title(t *testing.T) {
	target := "my-service:9090"
	g := NewLatencyGraph(target)
	title := g.View().GetTitle()
	if !strings.Contains(title, target) {
		t.Fatalf("expected title to contain %q, got %q", target, title)
	}
}
