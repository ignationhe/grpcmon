package probe_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/grpcmon/internal/probe"
)

func TestScheduler_PollsTargets(t *testing.T) {
	srv := startHealthServer(t, true)

	p := probe.New(probe.Config{Timeout: time.Second})
	agg := probe.NewAggregator()

	cfg := probe.SchedulerConfig{
		Interval: 50 * time.Millisecond,
		Targets:  []string{srv},
	}

	sched := probe.NewScheduler(cfg, p, agg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sched.Start(ctx)

	// Wait for at least one poll cycle.
	time.Sleep(120 * time.Millisecond)
	sched.Stop()

	result, ok := agg.Latest(srv)
	if !ok {
		t.Fatal("expected a recorded result for target")
	}
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
}

func TestScheduler_StopHaltsPolling(t *testing.T) {
	srv := startHealthServer(t, true)

	p := probe.New(probe.Config{Timeout: time.Second})
	agg := probe.NewAggregator()

	cfg := probe.SchedulerConfig{
		Interval: 30 * time.Millisecond,
		Targets:  []string{srv},
	}

	sched := probe.NewScheduler(cfg, p, agg)
	ctx := context.Background()

	sched.Start(ctx)
	time.Sleep(60 * time.Millisecond)
	sched.Stop()

	// Capture count after stop.
	h := agg.History(srv)
	countAfterStop := len(h.Entries())

	time.Sleep(80 * time.Millisecond)

	countLater := len(agg.History(srv).Entries())
	if countLater != countAfterStop {
		t.Errorf("scheduler continued polling after Stop: got %d entries, want %d",
			countLater, countAfterStop)
	}
}

func TestScheduler_ContextCancellation(t *testing.T) {
	srv := startHealthServer(t, true)

	p := probe.New(probe.Config{Timeout: time.Second})
	agg := probe.NewAggregator()

	cfg := probe.SchedulerConfig{
		Interval: 30 * time.Millisecond,
		Targets:  []string{srv},
	}

	sched := probe.NewScheduler(cfg, p, agg)
	ctx, cancel := context.WithCancel(context.Background())

	sched.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	cancel()

	time.Sleep(80 * time.Millisecond)

	_, ok := agg.Latest(srv)
	if !ok {
		t.Fatal("expected at least one result before cancellation")
	}
}
