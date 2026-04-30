package probe

import (
	"testing"
	"time"
)

func TestUptimeTracker_NoRecords(t *testing.T) {
	ut := NewUptimeTracker(time.Minute)
	_, ok := ut.Summary("localhost:50051")
	if ok {
		t.Fatal("expected false for unknown target")
	}
}

func TestUptimeTracker_AllHealthy(t *testing.T) {
	ut := NewUptimeTracker(time.Minute)
	for i := 0; i < 5; i++ {
		ut.Record("svc:50051", Result{Status: StatusServing})
	}
	s, ok := ut.Summary("svc:50051")
	if !ok {
		t.Fatal("expected summary")
	}
	if s.Percent != 100.0 {
		t.Errorf("expected 100%% uptime, got %.2f", s.Percent)
	}
	if s.Total != 5 || s.Healthy != 5 {
		t.Errorf("unexpected counts: total=%d healthy=%d", s.Total, s.Healthy)
	}
}

func TestUptimeTracker_PartialHealthy(t *testing.T) {
	ut := NewUptimeTracker(time.Minute)
	ut.Record("svc:50051", Result{Status: StatusServing})
	ut.Record("svc:50051", Result{Status: StatusServing})
	ut.Record("svc:50051", Result{Status: StatusNotServing})
	ut.Record("svc:50051", Result{Status: StatusNotServing})

	s, ok := ut.Summary("svc:50051")
	if !ok {
		t.Fatal("expected summary")
	}
	if s.Percent != 50.0 {
		t.Errorf("expected 50%% uptime, got %.2f", s.Percent)
	}
}

func TestUptimeTracker_AllUnhealthy(t *testing.T) {
	ut := NewUptimeTracker(time.Minute)
	for i := 0; i < 3; i++ {
		ut.Record("svc:50051", Result{Status: StatusNotServing})
	}
	s, ok := ut.Summary("svc:50051")
	if !ok {
		t.Fatal("expected summary")
	}
	if s.Percent != 0.0 {
		t.Errorf("expected 0%% uptime, got %.2f", s.Percent)
	}
}

func TestUptimeTracker_All(t *testing.T) {
	ut := NewUptimeTracker(time.Minute)
	ut.Record("a:1", Result{Status: StatusServing})
	ut.Record("b:2", Result{Status: StatusNotServing})

	all := ut.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(all))
	}
}

func TestUptimeTracker_Reset(t *testing.T) {
	ut := NewUptimeTracker(time.Minute)
	ut.Record("svc:50051", Result{Status: StatusServing})
	ut.Reset("svc:50051")
	_, ok := ut.Summary("svc:50051")
	if ok {
		t.Fatal("expected false after reset")
	}
}

func TestUptimeTracker_SinceIsSet(t *testing.T) {
	before := time.Now()
	ut := NewUptimeTracker(time.Minute)
	ut.Record("svc:50051", Result{Status: StatusServing})
	after := time.Now()

	s, _ := ut.Summary("svc:50051")
	if s.Since.Before(before) || s.Since.After(after) {
		t.Errorf("Since timestamp out of range: %v", s.Since)
	}
}
