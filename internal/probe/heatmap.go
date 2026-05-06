package probe

import (
	"sync"
	"time"
)

// HeatmapBucket represents a single time-bucket of latency samples.
type HeatmapBucket struct {
	Timestamp time.Time
	AvgLatency time.Duration
	ErrorRate  float64
	Samples    int
}

// HeatmapStore accumulates per-minute bucketed latency data per target.
type HeatmapStore struct {
	mu       sync.RWMutex
	buckets  map[string][]HeatmapBucket
	maxBuckets int
	bucketSize time.Duration
}

// NewHeatmapStore creates a HeatmapStore retaining up to maxBuckets time-buckets
// of width bucketSize per target.
func NewHeatmapStore(maxBuckets int, bucketSize time.Duration) *HeatmapStore {
	if maxBuckets <= 0 {
		maxBuckets = 60
	}
	if bucketSize <= 0 {
		bucketSize = time.Minute
	}
	return &HeatmapStore{
		buckets:    make(map[string][]HeatmapBucket),
		maxBuckets: maxBuckets,
		bucketSize: bucketSize,
	}
}

// Record adds a probe result into the appropriate time bucket.
func (h *HeatmapStore) Record(target string, r Result) {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := r.Timestamp.Truncate(h.bucketSize)
	buckets := h.buckets[target]

	if len(buckets) > 0 && buckets[len(buckets)-1].Timestamp.Equal(now) {
		b := &buckets[len(buckets)-1]
		b.Samples++
		if r.Err == nil {
			b.AvgLatency = (b.AvgLatency*time.Duration(b.Samples-1) + r.Latency) / time.Duration(b.Samples)
		} else {
			b.ErrorRate = (b.ErrorRate*float64(b.Samples-1) + 1) / float64(b.Samples)
		}
		h.buckets[target] = buckets
		return
	}

	errRate := 0.0
	if r.Err != nil {
		errRate = 1.0
	}
	buckets = append(buckets, HeatmapBucket{
		Timestamp:  now,
		AvgLatency: r.Latency,
		ErrorRate:  errRate,
		Samples:    1,
	})
	if len(buckets) > h.maxBuckets {
		buckets = buckets[len(buckets)-h.maxBuckets:]
	}
	h.buckets[target] = buckets
}

// Buckets returns a copy of the time-bucketed data for a target.
func (h *HeatmapStore) Buckets(target string) []HeatmapBucket {
	h.mu.RLock()
	defer h.mu.RUnlock()
	src := h.buckets[target]
	out := make([]HeatmapBucket, len(src))
	copy(out, src)
	return out
}

// Targets returns all known targets.
func (h *HeatmapStore) Targets() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	targets := make([]string, 0, len(h.buckets))
	for t := range h.buckets {
		targets = append(targets, t)
	}
	return targets
}
