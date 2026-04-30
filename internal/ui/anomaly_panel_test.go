package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/yourorg/grpcmon/internal/probe"
)

func makeAnomaly(kind probe.AnomalyKind, target, msg string) probe.Anomaly {
	return probe.Anomaly{
		Target:     target,
		Kind:       kind,
		Message:    msg,
		DetectedAt: time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC),
	}
}

func getText(p *AnomalyPanel) string {
	return p.view.GetText(false)
}

func TestAnomalyPanel_Title(t *testing.T) {
	p := NewAnomalyPanel(10)
	if p.view.GetTitle() != " Anomalies " {
		t.Fatalf("unexpected title: %q", p.view.GetTitle())
	}
}

func TestAnomalyPanel_NoAnomalies(t *testing.T) {
	p := NewAnomalyPanel(10)
	p.Update(nil)
	text := getText(p)
	if !strings.Contains(text, "no anomalies") {
		t.Fatalf("expected empty message, got: %q", text)
	}
}

func TestAnomalyPanel_ShowsTarget(t *testing.T) {
	p := NewAnomalyPanel(10)
	p.Update([]probe.Anomaly{
		makeAnomaly(probe.AnomalyLatencySpike, "api.svc:443", "avg 300ms exceeds 2x baseline"),
	})
	text := getText(p)
	if !strings.Contains(text, "api.svc:443") {
		t.Fatalf("expected target in output, got: %q", text)
	}
}

func TestAnomalyPanel_ShowsKind(t *testing.T) {
	p := NewAnomalyPanel(10)
	p.Update([]probe.Anomaly{
		makeAnomaly(probe.AnomalyErrorBurst, "svc", "error rate 60%"),
	})
	text := getText(p)
	if !strings.Contains(text, "error_burst") {
		t.Fatalf("expected kind in output, got: %q", text)
	}
}

func TestAnomalyPanel_Append_Evicts(t *testing.T) {
	p := NewAnomalyPanel(3)
	for i := 0; i < 5; i++ {
		p.Append([]probe.Anomaly{
			makeAnomaly(probe.AnomalyLatencySpike, "svc", "spike"),
		})
	}
	if len(p.entries) > 3 {
		t.Fatalf("expected at most 3 entries, got %d", len(p.entries))
	}
}

func TestAnomalyPanel_Primitive_NotNil(t *testing.T) {
	p := NewAnomalyPanel(10)
	if p.Primitive() == nil {
		t.Fatal("Primitive should not be nil")
	}
}
