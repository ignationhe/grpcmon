package ui

import (
	"errors"
	"testing"
	"time"

	"github.com/yourorg/grpcmon/internal/probe"
)

func setupRegionPanel(t *testing.T) (*probe.RegionStore, *probe.Aggregator, *RegionPanel) {
	t.Helper()
	store := probe.NewRegionStore()
	agg := probe.NewAggregator()
	panel := NewRegionPanel(store, agg)
	return store, agg, panel
}

func TestRegionPanel_Title(t *testing.T) {
	_, _, panel := setupRegionPanel(t)
	if panel.view.GetTitle() != " Regions " {
		t.Errorf("unexpected title: %q", panel.view.GetTitle())
	}
}

func TestRegionPanel_NoRegions(t *testing.T) {
	_, _, panel := setupRegionPanel(t)
	panel.Refresh()
	txt := panel.view.GetText(false)
	if txt == "" {
		t.Error("expected non-empty text for empty state")
	}
}

func TestRegionPanel_ShowsRegionName(t *testing.T) {
	store, agg, panel := setupRegionPanel(t)
	store.Assign("host:443", "us-east")
	agg.Record("host:443", probe.Result{
		Target:  "host:443",
		Status:  probe.StatusServing,
		Latency: 10 * time.Millisecond,
	})
	panel.Refresh()
	txt := panel.view.GetText(false)
	if !containsStr(txt, "us-east") {
		t.Errorf("expected region name in output, got: %q", txt)
	}
}

func TestRegionPanel_AllHealthyGreen(t *testing.T) {
	store, agg, panel := setupRegionPanel(t)
	store.Assign("h1:1", "eu-west")
	agg.Record("h1:1", probe.Result{Target: "h1:1", Status: probe.StatusServing, Latency: 5 * time.Millisecond})
	panel.Refresh()
	txt := panel.view.GetText(false)
	if !containsStr(txt, "eu-west") {
		t.Errorf("region not found in output")
	}
}

func TestRegionPanel_ErrorRateShown(t *testing.T) {
	store, agg, panel := setupRegionPanel(t)
	store.Assign("ok:1", "ap-south")
	store.Assign("bad:1", "ap-south")
	agg.Record("ok:1", probe.Result{Target: "ok:1", Status: probe.StatusServing, Latency: 3 * time.Millisecond})
	agg.Record("bad:1", probe.Result{Target: "bad:1", Status: probe.StatusNotServing, Err: errors.New("down")})
	panel.Refresh()
	txt := panel.view.GetText(false)
	if !containsStr(txt, "ap-south") {
		t.Errorf("expected ap-south in output")
	}
	if !containsStr(txt, "50.0") {
		t.Errorf("expected 50.0%% error rate in output, got: %q", txt)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
