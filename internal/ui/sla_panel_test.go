package ui

import (
	"strings"
	"testing"

	"grpcmon/internal/probe"
)

func TestSLAPanel_NoViolations(t *testing.T) {
	p := NewSLAPanel()
	p.Update(map[string][]probe.SLAViolation{})
	txt := p.view.GetText(false)
	if !strings.Contains(txt, "All targets within SLA") {
		t.Errorf("expected healthy message, got: %q", txt)
	}
}

func TestSLAPanel_ShowsTarget(t *testing.T) {
	p := NewSLAPanel()
	p.Update(map[string][]probe.SLAViolation{
		"svc:443": {
			{Target: "svc:443", Kind: "error_rate", Threshold: "10.0%", Actual: "75.0%"},
		},
	})
	txt := p.view.GetText(false)
	if !strings.Contains(txt, "svc:443") {
		t.Errorf("expected target name in output, got: %q", txt)
	}
}

func TestSLAPanel_ShowsViolationKind(t *testing.T) {
	p := NewSLAPanel()
	p.Update(map[string][]probe.SLAViolation{
		"svc:443": {
			{Target: "svc:443", Kind: "latency", Threshold: "100ms", Actual: "250ms"},
		},
	})
	txt := p.view.GetText(false)
	if !strings.Contains(txt, "latency") {
		t.Errorf("expected violation kind in output, got: %q", txt)
	}
}

func TestSLAPanel_ShowsThresholdAndActual(t *testing.T) {
	p := NewSLAPanel()
	p.Update(map[string][]probe.SLAViolation{
		"svc:443": {
			{Target: "svc:443", Kind: "latency", Threshold: "100ms", Actual: "350ms"},
		},
	})
	txt := p.view.GetText(false)
	if !strings.Contains(txt, "350ms") {
		t.Errorf("expected actual value in output, got: %q", txt)
	}
	if !strings.Contains(txt, "100ms") {
		t.Errorf("expected threshold in output, got: %q", txt)
	}
}

func TestSLAPanel_Title(t *testing.T) {
	p := NewSLAPanel()
	title := p.view.GetTitle()
	if !strings.Contains(title, "SLA") {
		t.Errorf("expected SLA in panel title, got: %q", title)
	}
}

func TestSLAPanel_MultipleTargets(t *testing.T) {
	p := NewSLAPanel()
	p.Update(map[string][]probe.SLAViolation{
		"alpha:443": {
			{Target: "alpha:443", Kind: "error_rate", Threshold: "5.0%", Actual: "20.0%"},
		},
		"beta:443": {
			{Target: "beta:443", Kind: "latency", Threshold: "50ms", Actual: "200ms"},
		},
	})
	txt := p.view.GetText(false)
	if !strings.Contains(txt, "alpha:443") {
		t.Errorf("expected alpha target, got: %q", txt)
	}
	if !strings.Contains(txt, "beta:443") {
		t.Errorf("expected beta target, got: %q", txt)
	}
}
