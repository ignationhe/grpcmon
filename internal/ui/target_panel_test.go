package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"grpcmon/internal/probe"
)

func makeHistory(t *testing.T, results []probe.Result) *probe.History {
	t.Helper()
	h := probe.NewHistory(10)
	for _, r := range results {
		h.Add(r)
	}
	return h
}

func TestTargetPanel_StatusServing(t *testing.T) {
	p := NewTargetPanel("localhost:50051")
	r := probe.Result{Target: "localhost:50051", Serving: true, Latency: 5 * time.Millisecond}
	h := makeHistory(t, []probe.Result{r})
	p.Update(r, h)
	txt := p.textView.GetText(false)
	if !strings.Contains(txt, "SERVING") {
		t.Errorf("expected SERVING in output, got: %s", txt)
	}
}

func TestTargetPanel_StatusNotServing(t *testing.T) {
	p := NewTargetPanel("localhost:50051")
	r := probe.Result{Target: "localhost:50051", Serving: false}
	h := makeHistory(t, []probe.Result{r})
	p.Update(r, h)
	txt := p.textView.GetText(false)
	if !strings.Contains(txt, "NOT SERVING") {
		t.Errorf("expected NOT SERVING in output, got: %s", txt)
	}
}

func TestTargetPanel_StatusError(t *testing.T) {
	p := NewTargetPanel("localhost:50051")
	r := probe.Result{Target: "localhost:50051", Err: errors.New("connection refused")}
	h := makeHistory(t, []probe.Result{r})
	p.Update(r, h)
	txt := p.textView.GetText(false)
	if !strings.Contains(txt, "ERROR") {
		t.Errorf("expected ERROR in output, got: %s", txt)
	}
}

func TestTargetPanel_ShowsLatency(t *testing.T) {
	p := NewTargetPanel("localhost:50051")
	r := probe.Result{Target: "localhost:50051", Serving: true, Latency: 42 * time.Millisecond}
	h := makeHistory(t, []probe.Result{r})
	p.Update(r, h)
	txt := p.textView.GetText(false)
	if !strings.Contains(txt, "42ms") {
		t.Errorf("expected 42ms in output, got: %s", txt)
	}
}

func TestTargetPanel_Title(t *testing.T) {
	target := "myservice:9090"
	p := NewTargetPanel(target)
	if !strings.Contains(p.textView.GetTitle(), target) {
		t.Errorf("expected title to contain %s", target)
	}
}

func TestTargetPanel_PrimitiveNotNil(t *testing.T) {
	p := NewTargetPanel("localhost:50051")
	if p.Primitive() == nil {
		t.Error("expected non-nil primitive")
	}
}
