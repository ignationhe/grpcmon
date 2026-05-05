package probe

import (
	"testing"
	"time"
)

func TestRegionStore_AssignAndRegions(t *testing.T) {
	rs := NewRegionStore()
	rs.Assign("host-a:443", "us-east")
	rs.Assign("host-b:443", "eu-west")
	rs.Assign("host-c:443", "us-east")

	regs := rs.Regions()
	if len(regs) != 2 {
		t.Fatalf("expected 2 regions, got %d", len(regs))
	}
	if regs[0] != "eu-west" || regs[1] != "us-east" {
		t.Errorf("unexpected order: %v", regs)
	}
}

func TestRegionStore_ReassignMovesTarget(t *testing.T) {
	rs := NewRegionStore()
	rs.Assign("host-a:443", "us-east")
	rs.Assign("host-a:443", "eu-west") // reassign

	regs := rs.Regions()
	for _, r := range regs {
		if r == "us-east" {
			// us-east should be empty now
			s := rs.Summarise("us-east", NewAggregator())
			if len(s.Targets) != 0 {
				t.Errorf("expected us-east to be empty after reassign")
			}
		}
	}
}

func TestRegionStore_Summarise_Empty(t *testing.T) {
	rs := NewRegionStore()
	s := rs.Summarise("nowhere", NewAggregator())
	if s.Healthy != 0 || s.Unhealthy != 0 || s.ErrorRate != 0 {
		t.Errorf("expected zero summary for unknown region")
	}
}

func TestRegionStore_Summarise_AllHealthy(t *testing.T) {
	rs := NewRegionStore()
	agg := NewAggregator()

	addrs := []string{"a:1", "b:1"}
	for _, addr := range addrs {
		rs.Assign(addr, "us-east")
		agg.Record(addr, Result{Target: addr, Status: StatusServing, Latency: 10 * time.Millisecond})
	}

	s := rs.Summarise("us-east", agg)
	if s.Healthy != 2 || s.Unhealthy != 0 {
		t.Errorf("expected 2 healthy, got healthy=%d unhealthy=%d", s.Healthy, s.Unhealthy)
	}
	if s.ErrorRate != 0 {
		t.Errorf("expected 0 error rate, got %f", s.ErrorRate)
	}
}

func TestRegionStore_Summarise_PartialFailure(t *testing.T) {
	rs := NewRegionStore()
	agg := NewAggregator()

	rs.Assign("ok:1", "ap-south")
	rs.Assign("bad:1", "ap-south")

	agg.Record("ok:1", Result{Target: "ok:1", Status: StatusServing, Latency: 5 * time.Millisecond})
	agg.Record("bad:1", Result{Target: "bad:1", Status: StatusNotServing, Err: errNotServing})

	s := rs.Summarise("ap-south", agg)
	if s.Healthy != 1 || s.Unhealthy != 1 {
		t.Errorf("expected 1/1, got healthy=%d unhealthy=%d", s.Healthy, s.Unhealthy)
	}
	if s.ErrorRate != 0.5 {
		t.Errorf("expected 0.5 error rate, got %f", s.ErrorRate)
	}
}
