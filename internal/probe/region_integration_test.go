package probe_test

import (
	"errors"
	"testing"
	"time"

	"github.com/yourorg/grpcmon/internal/probe"
)

// TestRegionStore_IntegrationWithAggregator verifies that RegionStore and
// Aggregator work together to produce correct region summaries after a
// sequence of probe recordings.
func TestRegionStore_IntegrationWithAggregator(t *testing.T) {
	store := probe.NewRegionStore()
	agg := probe.NewAggregator()

	targets := map[string]string{
		"ny1:443": "us-east",
		"ny2:443": "us-east",
		"ld1:443": "eu-west",
	}
	for addr, region := range targets {
		store.Assign(addr, region)
	}

	// All healthy initially.
	for addr := range targets {
		agg.Record(addr, probe.Result{Target: addr, Status: probe.StatusServing, Latency: 8 * time.Millisecond})
	}

	usEast := store.Summarise("us-east", agg)
	if usEast.Healthy != 2 || usEast.Unhealthy != 0 {
		t.Errorf("us-east: want 2/0, got %d/%d", usEast.Healthy, usEast.Unhealthy)
	}

	euWest := store.Summarise("eu-west", agg)
	if euWest.Healthy != 1 || euWest.Unhealthy != 0 {
		t.Errorf("eu-west: want 1/0, got %d/%d", euWest.Healthy, euWest.Unhealthy)
	}

	// Simulate ny2 going down.
	agg.Record("ny2:443", probe.Result{
		Target: "ny2:443",
		Status: probe.StatusNotServing,
		Err:    errors.New("connection refused"),
	})

	usEast = store.Summarise("us-east", agg)
	if usEast.Healthy != 1 || usEast.Unhealthy != 1 {
		t.Errorf("us-east after failure: want 1/1, got %d/%d", usEast.Healthy, usEast.Unhealthy)
	}
	if usEast.ErrorRate != 0.5 {
		t.Errorf("expected error rate 0.5, got %f", usEast.ErrorRate)
	}

	// eu-west unaffected.
	euWest = store.Summarise("eu-west", agg)
	if euWest.ErrorRate != 0 {
		t.Errorf("eu-west should be unaffected, got error rate %f", euWest.ErrorRate)
	}
}

// TestRegionStore_Regions_SortedOrder ensures region names are always sorted.
func TestRegionStore_Regions_SortedOrder(t *testing.T) {
	store := probe.NewRegionStore()
	store.Assign("z:1", "us-west")
	store.Assign("a:1", "ap-east")
	store.Assign("m:1", "eu-central")

	regs := store.Regions()
	expected := []string{"ap-east", "eu-central", "us-west"}
	for i, r := range regs {
		if r != expected[i] {
			t.Errorf("pos %d: want %q got %q", i, expected[i], r)
		}
	}
}
