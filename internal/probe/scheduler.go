package probe

import (
	"context"
	"sync"
	"time"
)

// SchedulerConfig holds configuration for the probe scheduler.
type SchedulerConfig struct {
	Interval time.Duration
	Targets  []string
}

// Scheduler runs probes against a set of targets at a fixed interval,
// feeding results into an Aggregator.
type Scheduler struct {
	cfg        SchedulerConfig
	probe      *Probe
	aggregator *Aggregator
	stopOnce   sync.Once
	stopCh     chan struct{}
}

// NewScheduler creates a Scheduler that will poll cfg.Targets every cfg.Interval.
func NewScheduler(cfg SchedulerConfig, p *Probe, agg *Aggregator) *Scheduler {
	return &Scheduler{
		cfg:        cfg,
		probe:      p,
		aggregator: agg,
		stopCh:     make(chan struct{}),
	}
}

// Start begins polling in the background. It returns immediately.
// Cancel ctx or call Stop to halt the scheduler.
func (s *Scheduler) Start(ctx context.Context) {
	go s.run(ctx)
}

// Stop signals the scheduler to cease polling.
func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
}

func (s *Scheduler) run(ctx context.Context) {
	s.poll(ctx)

	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.poll(ctx)
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scheduler) poll(ctx context.Context) {
	var wg sync.WaitGroup
	for _, target := range s.cfg.Targets {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			result := s.probe.Check(ctx, t)
			s.aggregator.Record(t, result)
		}(target)
	}
	wg.Wait()
}
