package probe

import (
	"testing"
	"time"
)

func TestP99Tracker_NoSamples(t *testing.T) {
	tr := NewP99Tracker(time.Minute)
	if got := tr.Percentile("svc:443", 99); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestP99Tracker_SingleSample(t *testing.T) {
	tr := NewP99Tracker(time.Minute)
	tr.Record("svc:443", 50*time.Millisecond)
	if got := tr.Percentile("svc:443", 99); got != 50*time.Millisecond {
		t.Fatalf("expected 50ms, got %v", got)
	}
}

func TestP99Tracker_Percentiles(t *testing.T) {
	tr := NewP99Tracker(time.Minute)
	for i := 1; i <= 100; i++ {
		tr.Record("svc:443", time.Duration(i)*time.Millisecond)
	}
	p50 := tr.Percentile("svc:443", 50)
	p99 := tr.Percentile("svc:443", 99)
	if p50 < 49*time.Millisecond || p50 > 51*time.Millisecond {
		t.Fatalf("p50 out of range: %v", p50)
	}
	if p99 < 98*time.Millisecond {
		t.Fatalf("p99 too low: %v", p99)
	}
}

func TestP99Tracker_Eviction(t *testing.T) {
	tr := NewP99Tracker(50 * time.Millisecond)
	tr.Record("svc:443", 200*time.Millisecond)
	time.Sleep(60 * time.Millisecond)
	if got := tr.Percentile("svc:443", 99); got != 0 {
		t.Fatalf("expected evicted sample to return 0, got %v", got)
	}
}

func TestP99Tracker_DifferentTargetsIndependent(t *testing.T) {
	tr := NewP99Tracker(time.Minute)
	tr.Record("a:80", 10*time.Millisecond)
	tr.Record("b:80", 100*time.Millisecond)
	if tr.Percentile("a:80", 99) == tr.Percentile("b:80", 99) {
		t.Fatal("different targets should have independent percentiles")
	}
}

func TestP99Tracker_Targets(t *testing.T) {
	tr := NewP99Tracker(time.Minute)
	tr.Record("x:1", time.Millisecond)
	tr.Record("y:2", time.Millisecond)
	targets := tr.Targets()
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
}

func TestP99Tracker_TargetsExcludesEvicted(t *testing.T) {
	tr := NewP99Tracker(30 * time.Millisecond)
	tr.Record("old:1", time.Millisecond)
	time.Sleep(40 * time.Millisecond)
	if len(tr.Targets()) != 0 {
		t.Fatal("evicted target should not appear in Targets()")
	}
}
