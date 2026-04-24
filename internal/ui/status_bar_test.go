package ui

import (
	"strings"
	"testing"
	"time"
)

func TestStatusBar_AllHealthy(t *testing.T) {
	sb := NewStatusBar()
	sb.Update(3, 0, 3, time.Time{})
	txt := sb.View().GetText(false)
	if !strings.Contains(txt, "All 3 targets healthy") {
		t.Errorf("expected healthy message, got: %q", txt)
	}
}

func TestStatusBar_SomeUnhealthy(t *testing.T) {
	sb := NewStatusBar()
	sb.Update(1, 2, 3, time.Time{})
	txt := sb.View().GetText(false)
	if !strings.Contains(txt, "2/3 targets unhealthy") {
		t.Errorf("expected unhealthy message, got: %q", txt)
	}
}

func TestStatusBar_NoTargets(t *testing.T) {
	sb := NewStatusBar()
	sb.Update(0, 0, 0, time.Time{})
	txt := sb.View().GetText(false)
	if !strings.Contains(txt, "No targets") {
		t.Errorf("expected no-targets message, got: %q", txt)
	}
}

func TestStatusBar_ShowsLastPoll(t *testing.T) {
	sb := NewStatusBar()
	pollTime := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	sb.Update(2, 0, 2, pollTime)
	txt := sb.View().GetText(false)
	if !strings.Contains(txt, "14:30:45") {
		t.Errorf("expected poll time 14:30:45 in output, got: %q", txt)
	}
}

func TestStatusBar_ZeroPollTimeHidden(t *testing.T) {
	sb := NewStatusBar()
	sb.Update(1, 0, 1, time.Time{})
	txt := sb.View().GetText(false)
	if strings.Contains(txt, "Last poll") {
		t.Errorf("expected no last poll text for zero time, got: %q", txt)
	}
}

func TestStatusBar_ViewNotNil(t *testing.T) {
	sb := NewStatusBar()
	if sb.View() == nil {
		t.Error("expected non-nil view")
	}
}
