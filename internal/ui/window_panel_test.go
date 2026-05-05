package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/grpcmon/internal/probe"
)

func addWindowResult(agg *probe.WindowAggregator, target string, latency time.Duration, err error) {
	agg.Record(target, probe.Result{
		Target:    target,
		Timestamp: time.Now(),
		Latency:   latency,
		Err:       err,
	})
}

func TestWindowPanel_Title(t *testing.T) {
	agg := probe.NewWindowAggregator(time.Minute)
	p := NewWindowPanel(agg)
	if !strings.Contains(p.view.GetTitle(), "Window") {
		t.Fatal("expected title to contain 'Window'")
	}
}

func TestWindowPanel_NoData(t *testing.T) {
	agg := probe.NewWindowAggregator(time.Minute)
	p := NewWindowPanel(agg)
	p.Update()
	text := p.view.GetText(false)
	if !strings.Contains(text, "No data") {
		t.Fatalf("expected 'No data', got: %q", text)
	}
}

func TestWindowPanel_ShowsTarget(t *testing.T) {
	agg := probe.NewWindowAggregator(time.Minute)
	addWindowResult(agg, "svc:443", 10*time.Millisecond, nil)
	p := NewWindowPanel(agg)
	p.Update()
	text := p.view.GetText(false)
	if !strings.Contains(text, "svc:443") {
		t.Fatalf("expected target in output, got: %q", text)
	}
}

func TestWindowPanel_ShowsErrorRate(t *testing.T) {
	agg := probe.NewWindowAggregator(time.Minute)
	addWindowResult(agg, "svc:443", 10*time.Millisecond, nil)
	addWindowResult(agg, "svc:443", 10*time.Millisecond, errors.New("fail"))
	p := NewWindowPanel(agg)
	p.Update()
	text := p.view.GetText(false)
	if !strings.Contains(text, "50.0%") {
		t.Fatalf("expected 50.0%% error rate, got: %q", text)
	}
}

func TestWindowPanel_ShowsAvgLatency(t *testing.T) {
	agg := probe.NewWindowAggregator(time.Minute)
	addWindowResult(agg, "svc:443", 20*time.Millisecond, nil)
	addWindowResult(agg, "svc:443", 40*time.Millisecond, nil)
	p := NewWindowPanel(agg)
	p.Update()
	text := p.view.GetText(false)
	if !strings.Contains(text, "30.0ms") {
		t.Fatalf("expected avg latency 30.0ms, got: %q", text)
	}
}

func TestWindowPanel_Primitive(t *testing.T) {
	agg := probe.NewWindowAggregator(time.Minute)
	p := NewWindowPanel(agg)
	if p.Primitive() == nil {
		t.Fatal("expected non-nil primitive")
	}
}
