package probe

import (
	"sync"
	"time"
)

// RTTSample holds a single round-trip time measurement for a target.
type RTTSample struct {
	Target    string
	Timestamp time.Time
	RTT       time.Duration
}

// RTTTracker records recent RTT samples per target and provides
// min/max/avg statistics over a sliding window.
type RTTTracker struct {
	mu      sync.Mutex
	window  int
	samples map[string][]RTTSample
}

// NewRTTTracker creates a new RTTTracker that retains at most window
// samples per target.
func NewRTTTracker(window int) *RTTTracker {
	if window <= 0 {
		window = 60
	}
	return &RTTTracker{
		window:  window,
		samples: make(map[string][]RTTSample),
	}
}

// Record adds an RTT sample for the given target.
func (r *RTTTracker) Record(target string, rtt time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := RTTSample{Target: target, Timestamp: time.Now(), RTT: rtt}
	r.samples[target] = append(r.samples[target], s)
	if len(r.samples[target]) > r.window {
		r.samples[target] = r.samples[target][len(r.samples[target])-r.window:]
	}
}

// Stats returns min, max, and average RTT for the given target.
// Returns zeros if no samples exist.
func (r *RTTTracker) Stats(target string) (min, max, avg time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	samples := r.samples[target]
	if len(samples) == 0 {
		return 0, 0, 0
	}
	min = samples[0].RTT
	max = samples[0].RTT
	var total time.Duration
	for _, s := range samples {
		if s.RTT < min {
			min = s.RTT
		}
		if s.RTT > max {
			max = s.RTT
		}
		total += s.RTT
	}
	avg = total / time.Duration(len(samples))
	return min, max, avg
}

// Samples returns a copy of the recorded samples for the given target.
func (r *RTTTracker) Samples(target string) []RTTSample {
	r.mu.Lock()
	defer r.mu.Unlock()
	src := r.samples[target]
	out := make([]RTTSample, len(src))
	copy(out, src)
	return out
}

// Targets returns all targets that have at least one sample.
func (r *RTTTracker) Targets() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, 0, len(r.samples))
	for t := range r.samples {
		out = append(out, t)
	}
	return out
}
