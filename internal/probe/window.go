package probe

import (
	"sync"
	"time"
)

// WindowStats holds aggregated statistics for a rolling time window.
type WindowStats struct {
	Target    string
	Window    time.Duration
	Count     int
	ErrorRate float64
	AvgLatency time.Duration
	P95Latency time.Duration
}

// WindowAggregator computes rolling-window statistics over probe results.
type WindowAggregator struct {
	mu      sync.Mutex
	window  time.Duration
	buckets map[string][]Result
}

// NewWindowAggregator creates a WindowAggregator with the given rolling window.
func NewWindowAggregator(window time.Duration) *WindowAggregator {
	return &WindowAggregator{
		window:  window,
		buckets: make(map[string][]Result),
	}
}

// Record adds a probe result for the given target.
func (w *WindowAggregator) Record(target string, r Result) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buckets[target] = append(w.buckets[target], r)
	w.evict(target)
}

// evict removes entries older than the window. Must be called with lock held.
func (w *WindowAggregator) evict(target string) {
	cutoff := time.Now().Add(-w.window)
	entries := w.buckets[target]
	idx := 0
	for idx < len(entries) && entries[idx].Timestamp.Before(cutoff) {
		idx++
	}
	w.buckets[target] = entries[idx:]
}

// Stats returns WindowStats for the given target.
func (w *WindowAggregator) Stats(target string) WindowStats {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict(target)
	entries := w.buckets[target]
	stats := WindowStats{Target: target, Window: w.window, Count: len(entries)}
	if len(entries) == 0 {
		return stats
	}
	var errCount int
	var totalLatency time.Duration
	latencies := make([]time.Duration, 0, len(entries))
	for _, e := range entries {
		if e.Err != nil {
			errCount++
		}
		totalLatency += e.Latency
		latencies = append(latencies, e.Latency)
	}
	stats.ErrorRate = float64(errCount) / float64(len(entries))
	stats.AvgLatency = totalLatency / time.Duration(len(entries))
	sortDurations(latencies)
	p95idx := int(float64(len(latencies))*0.95)
	if p95idx >= len(latencies) {
		p95idx = len(latencies) - 1
	}
	stats.P95Latency = latencies[p95idx]
	return stats
}

// Targets returns all tracked target addresses.
func (w *WindowAggregator) Targets() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]string, 0, len(w.buckets))
	for k := range w.buckets {
		out = append(out, k)
	}
	return out
}
