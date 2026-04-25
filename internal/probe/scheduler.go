package probe

import (
	"context"
	"sync"
	"time"

	"github.com/lmittmann/grpcmon/internal/config"
)

// Scheduler drives periodic health-check probes for all configured targets.
// It uses a RateLimiter to prevent bursts and an Aggregator to store results.
type Scheduler struct {
	cfg         *config.Config
	probe       *Probe
	aggregator  *Aggregator
	rateLimiter *RateLimiter
	wg          sync.WaitGroup
	cancel      context.CancelFunc
}

// NewScheduler creates a Scheduler wired to the provided Probe and Aggregator.
func NewScheduler(cfg *config.Config, p *Probe, agg *Aggregator) *Scheduler {
	minDelay := time.Duration(cfg.PollInterval) / 2
	if minDelay < 100*time.Millisecond {
		minDelay = 100 * time.Millisecond
	}
	return &Scheduler{
		cfg:         cfg,
		probe:       p,
		aggregator:  agg,
		rateLimiter: NewRateLimiter(minDelay),
	}
}

// Start launches background goroutines that poll each target at the configured
// interval. It is safe to call Start only once.
func (s *Scheduler) Start(ctx context.Context) {
	ctx, s.cancel = context.WithCancel(ctx)
	for _, t := range s.cfg.Targets {
		target := t
		s.wg.Add(1)
		go s.runTarget(ctx, target)
	}
}

// Stop signals all polling goroutines to exit and waits for them to finish.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

func (s *Scheduler) runTarget(ctx context.Context, target config.Target) {
	defer s.wg.Done()
	ticker := time.NewTicker(time.Duration(s.cfg.PollInterval))
	defer ticker.Stop()

	s.poll(ctx, target)

	for {
		select {
		case <-ticker.C:
			s.poll(ctx, target)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scheduler) poll(ctx context.Context, target config.Target) {
	if err := s.rateLimiter.Wait(ctx, target.Address); err != nil {
		return
	}
	result := s.probe.Check(ctx, target)
	s.aggregator.Record(target.Address, result)
}
