package probe

import (
	"context"
	"sync"
	"time"
)

// PollConfig describes a polling job for a single target.
type PollConfig struct {
	Target   string
	Service  string
	Interval time.Duration
	Timeout  time.Duration
}

// Poller continuously probes a set of targets and emits Results.
type Poller struct {
	configs []PollConfig
	out     chan Result
	wg      sync.WaitGroup
}

// NewPoller creates a Poller for the provided configurations.
func NewPoller(configs []PollConfig) *Poller {
	return &Poller{
		configs: configs,
		out:     make(chan Result, len(configs)*4),
	}
}

// Results returns the read-only channel of probe results.
func (p *Poller) Results() <-chan Result {
	return p.out
}

// Start launches background goroutines for each target.
// It stops when ctx is cancelled.
func (p *Poller) Start(ctx context.Context) {
	for _, cfg := range p.configs {
		cfg := cfg
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.poll(ctx, cfg)
		}()
	}
}

// Wait blocks until all polling goroutines have exited.
func (p *Poller) Wait() {
	p.wg.Wait()
	close(p.out)
}

func (p *Poller) poll(ctx context.Context, cfg PollConfig) {
	prober, err := New(cfg.Target, cfg.Timeout)
	if err != nil {
		p.out <- Result{Target: cfg.Target, Err: err, Timestamp: time.Now()}
		return
	}
	defer prober.Close()

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			result := prober.Check(ctx, cfg.Service)
			select {
			case p.out <- result:
			default:
				// drop if consumer is slow
			}
		}
	}
}
