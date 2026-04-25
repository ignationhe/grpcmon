package probe

import (
	"testing"
	"time"
)

func TestAggregator_RecordAndLatest(t *testing.T) {
	a := NewAggregator(10)

	r := Result{Target: "localhost:50051", Latency: 5 * time.Millisecond, Healthy: true}
	a.Record("localhost:50051", r)

	got, ok := a.Latest("localhost:50051")
	if !ok {
		t.Fatal("expected result to exist")
	}
	if got.Latency != r.Latency {
		t.Errorf("latency: got %v, want %v", got.Latency, r.Latency)
	}
	if !got.Healthy {
		t.Error("expected healthy")
	}
}

func TestAggregator_Latest_Missing(t *testing.T) {
	a := NewAggregator(10)
	_, ok := a.Latest("unknown:9090")
	if ok {
		t.Error("expected no result for unknown target")
	}
}

func TestAggregator_History_Populated(t *testing.T) {
	a := NewAggregator(5)
	target := "svc:443"

	for i := 0; i < 3; i++ {
		a.Record(target, Result{Target: target, Latency: time.Duration(i) * time.Millisecond, Healthy: true})
	}

	h, ok := a.History(target)
	if !ok {
		t.Fatal("expected history to exist")
	}
	if len(h.Entries()) != 3 {
		t.Errorf("entries: got %d, want 3", len(h.Entries()))
	}
}

func TestAggregator_Targets(t *testing.T) {
	a := NewAggregator(10)
	a.Record("a:1", Result{Target: "a:1"})
	a.Record("b:2", Result{Target: "b:2"})

	targets := a.Targets()
	if len(targets) != 2 {
		t.Errorf("targets count: got %d, want 2", len(targets))
	}
}

func TestAggregator_Snapshot(t *testing.T) {
	a := NewAggregator(10)
	a.Record("x:80", Result{Target: "x:80", Healthy: true, Latency: 1 * time.Millisecond})
	a.Record("y:80", Result{Target: "y:80", Healthy: false, Latency: 2 * time.Millisecond})

	snap := a.Snapshot()
	if len(snap) != 2 {
		t.Errorf("snapshot size: got %d, want 2", len(snap))
	}
	if !snap["x:80"].Healthy {
		t.Error("x:80 should be healthy")
	}
	if snap["y:80"].Healthy {
		t.Error("y:80 should not be healthy")
	}
}

func TestAggregator_OverwritesLatest(t *testing.T) {
	a := NewAggregator(10)
	target := "svc:9000"

	a.Record(target, Result{Target: target, Latency: 10 * time.Millisecond, Healthy: false})
	a.Record(target, Result{Target: target, Latency: 20 * time.Millisecond, Healthy: true})

	got, _ := a.Latest(target)
	if got.Latency != 20*time.Millisecond {
		t.Errorf("expected latest latency 20ms, got %v", got.Latency)
	}
	if !got.Healthy {
		t.Error("expected latest to be healthy")
	}
}
