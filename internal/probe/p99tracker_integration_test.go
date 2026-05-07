package probe_test

import (
	"testing"
	"time"

	"github.com/yourorg/grpcmon/internal/probe"
)

func TestP99Tracker_IntegrationWithWindow(t *testing.T) {
	window := probe.NewWindowAggregator(5 * time.Second)
	p99 := probe.NewP99Tracker(20, 5*time.Second)

	now := time.Now()
	latencies := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
		60 * time.Millisecond,
		70 * time.Millisecond,
		80 * time.Millisecond,
		90 * time.Millisecond,
		100 * time.Millisecond,
	}

	for i, lat := range latencies {
		r := probe.Result{
			Target:  "svc:443",
			Status:  probe.Serving,
			Latency: lat,
			CheckedAt: now.Add(time.Duration(i) * time.Millisecond),
		}
		window.Record(r)
		p99.Record(r)
	}

	stats := window.Stats("svc:443")
	if stats.Count != 10 {
		t.Fatalf("expected 10 window records, got %d", stats.Count)
	}

	p99val, ok := p99.P99("svc:443")
	if !ok {
		t.Fatal("expected P99 to be available")
	}
	if p99val < 90*time.Millisecond {
		t.Errorf("expected P99 >= 90ms, got %v", p99val)
	}
}

func TestP99Tracker_MultiTarget_Independent(t *testing.T) {
	p99 := probe.NewP99Tracker(20, 5*time.Second)
	now := time.Now()

	for i := 0; i < 5; i++ {
		p99.Record(probe.Result{
			Target:    "fast:443",
			Status:    probe.Serving,
			Latency:   time.Duration(i+1) * 5 * time.Millisecond,
			CheckedAt: now.Add(time.Duration(i) * time.Millisecond),
		})
		p99.Record(probe.Result{
			Target:    "slow:443",
			Status:    probe.Serving,
			Latency:   time.Duration(i+1) * 100 * time.Millisecond,
			CheckedAt: now.Add(time.Duration(i) * time.Millisecond),
		})
	}

	fastP99, ok1 := p99.P99("fast:443")
	slowP99, ok2 := p99.P99("slow:443")

	if !ok1 || !ok2 {
		t.Fatal("expected both targets to have P99")
	}
	if fastP99 >= slowP99 {
		t.Errorf("expected fast P99 (%v) < slow P99 (%v)", fastP99, slowP99)
	}
}
