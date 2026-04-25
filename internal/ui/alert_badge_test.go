package ui

import (
	"testing"
	"time"

	"github.com/rivo/tview"
	"grpcmon/internal/probe"
)

func makeAlert(target, msg string) probe.Alert {
	return probe.Alert{
		Target:    target,
		Message:   msg,
		FiredAt:   time.Now(),
		Severity:  probe.SeverityWarning,
	}
}

func TestAlertBadge_NoAlerts(t *testing.T) {
	app := tview.NewApplication()
	_ = app
	badge := NewAlertBadge()
	badge.Update(nil)
	text := badge.TextView.GetText(true)
	if text != "No active alerts" {
		t.Errorf("expected 'No active alerts', got %q", text)
	}
}

func TestAlertBadge_SingleAlert(t *testing.T) {
	badge := NewAlertBadge()
	alerts := []probe.Alert{makeAlert("svc-a:50051", "high error rate")}
	badge.Update(alerts)
	text := badge.TextView.GetText(true)
	if text == "" {
		t.Error("expected non-empty alert text")
	}
	if !contains(text, "svc-a:50051") {
		t.Errorf("expected target in alert text, got %q", text)
	}
}

func TestAlertBadge_MultipleAlerts(t *testing.T) {
	badge := NewAlertBadge()
	alerts := []probe.Alert{
		makeAlert("svc-a:50051", "high error rate"),
		makeAlert("svc-b:50052", "latency spike"),
	}
	badge.Update(alerts)
	text := badge.TextView.GetText(true)
	if !contains(text, "svc-a:50051") {
		t.Errorf("expected svc-a in text, got %q", text)
	}
	if !contains(text, "svc-b:50052") {
		t.Errorf("expected svc-b in text, got %q", text)
	}
}

func TestAlertBadge_Title(t *testing.T) {
	badge := NewAlertBadge()
	title := badge.TextView.GetTitle()
	if title == "" {
		t.Error("expected non-empty title for alert badge")
	}
}

func TestAlertBadge_ClearsOnUpdate(t *testing.T) {
	badge := NewAlertBadge()
	badge.Update([]probe.Alert{makeAlert("svc-a:50051", "error")})
	badge.Update(nil)
	text := badge.TextView.GetText(true)
	if text != "No active alerts" {
		t.Errorf("expected cleared alerts, got %q", text)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})())
}
