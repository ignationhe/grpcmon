package probe

import (
	"testing"
	"time"
)

func TestRTTTracker_NoSamples(t *testing.T) {
	tr := NewRTTTracker(10)
	min, max, avg := tr.Stats("host:50051")
	if min != 0 || max != 0 || avg != 0 {
		t.Fatalf("expected zeros for empty tracker, got min=%v max=%v avg=%v", min, max, avg)
	}
}

func TestRTTTracker_SingleSample(t *testing.T) {
	tr := NewRTTTracker(10)
	tr.Record("host:50051", 20*time.Millisecond)
	min, max, avg := tr.Stats("host:50051")
	if min != 20*time.Millisecond {
		t.Errorf("expected min=20ms, got %v", min)
	}
	if max != 20*time.Millisecond {
		t.Errorf("expected max=20ms, got %v", max)
	}
	if avg != 20*time.Millisecond {
		t.Errorf("expected avg=20ms, got %v", avg)
	}
}

func TestRTTTracker_MultiSample_Stats(t *testing.T) {
	tr := NewRTTTracker(10)
	tr.Record("svc", 10*time.Millisecond)
	tr.Record("svc", 30*time.Millisecond)
	tr.Record("svc", 20*time.Millisecond)
	min, max, avg := tr.Stats("svc")
	if min != 10*time.Millisecond {
		t.Errorf("expected min=10ms, got %v", min)
	}
	if max != 30*time.Millisecond {
		t.Errorf("expected max=30ms, got %v", max)
	}
	if avg != 20*time.Millisecond {
		t.Errorf("expected avg=20ms, got %v", avg)
	}
}

func TestRTTTracker_Eviction(t *testing.T) {
	tr := NewRTTTracker(3)
	for i := 0; i < 5; i++ {
		tr.Record("svc", time.Duration(i+1)*time.Millisecond)
	}
	samples := tr.Samples("svc")
	if len(samples) != 3 {
		t.Fatalf("expected 3 samples after eviction, got %d", len(samples))
	}
	// oldest samples should be evicted; last 3 are 3ms, 4ms, 5ms
	if samples[0].RTT != 3*time.Millisecond {
		t.Errorf("expected oldest retained sample=3ms, got %v", samples[0].RTT)
	}
}

func TestRTTTracker_Targets(t *testing.T) {
	tr := NewRTTTracker(10)
	tr.Record("a:1", 1*time.Millisecond)
	tr.Record("b:2", 2*time.Millisecond)
	targets := tr.Targets()
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
}

func TestRTTTracker_DifferentTargetsIndependent(t *testing.T) {
	tr := NewRTTTracker(10)
	tr.Record("a", 5*time.Millisecond)
	tr.Record("b", 50*time.Millisecond)
	_, _, avgA := tr.Stats("a")
	_, _, avgB := tr.Stats("b")
	if avgA == avgB {
		t.Error("expected different averages for different targets")
	}
}

func TestRTTTracker_DefaultWindow(t *testing.T) {
	tr := NewRTTTracker(0)
	if tr.window != 60 {
		t.Errorf("expected default window=60, got %d", tr.window)
	}
}
