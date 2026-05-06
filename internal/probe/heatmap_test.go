package probe

import (
	"errors"
	"testing"
	"time"
)

func makeHeatmapResult(ts time.Time, latency time.Duration, err error) Result {
	return Result{
		Target:    "svc:50051",
		Timestamp: ts,
		Latency:   latency,
		Err:       err,
	}
}

func TestHeatmapStore_EmptyInitially(t *testing.T) {
	h := NewHeatmapStore(60, time.Minute)
	if got := h.Buckets("svc:50051"); len(got) != 0 {
		t.Fatalf("expected empty, got %d buckets", len(got))
	}
}

func TestHeatmapStore_RecordSingleBucket(t *testing.T) {
	h := NewHeatmapStore(60, time.Minute)
	ts := time.Now().Truncate(time.Minute)
	h.Record("svc:50051", makeHeatmapResult(ts, 10*time.Millisecond, nil))
	h.Record("svc:50051", makeHeatmapResult(ts.Add(30*time.Second), 20*time.Millisecond, nil))

	buckets := h.Buckets("svc:50051")
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
	if buckets[0].Samples != 2 {
		t.Errorf("expected 2 samples, got %d", buckets[0].Samples)
	}
}

func TestHeatmapStore_MultipleBuckets(t *testing.T) {
	h := NewHeatmapStore(60, time.Minute)
	base := time.Now().Truncate(time.Minute)
	h.Record("svc:50051", makeHeatmapResult(base, 10*time.Millisecond, nil))
	h.Record("svc:50051", makeHeatmapResult(base.Add(time.Minute), 20*time.Millisecond, nil))

	buckets := h.Buckets("svc:50051")
	if len(buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(buckets))
	}
}

func TestHeatmapStore_Eviction(t *testing.T) {
	h := NewHeatmapStore(3, time.Minute)
	base := time.Now().Truncate(time.Minute)
	for i := 0; i < 5; i++ {
		h.Record("svc:50051", makeHeatmapResult(base.Add(time.Duration(i)*time.Minute), 5*time.Millisecond, nil))
	}
	buckets := h.Buckets("svc:50051")
	if len(buckets) != 3 {
		t.Errorf("expected 3 buckets after eviction, got %d", len(buckets))
	}
}

func TestHeatmapStore_ErrorRateTracked(t *testing.T) {
	h := NewHeatmapStore(60, time.Minute)
	ts := time.Now().Truncate(time.Minute)
	h.Record("svc:50051", makeHeatmapResult(ts, 0, errors.New("down")))
	h.Record("svc:50051", makeHeatmapResult(ts.Add(10*time.Second), 0, errors.New("down")))

	buckets := h.Buckets("svc:50051")
	if buckets[0].ErrorRate != 1.0 {
		t.Errorf("expected error rate 1.0, got %f", buckets[0].ErrorRate)
	}
}

func TestHeatmapStore_Targets(t *testing.T) {
	h := NewHeatmapStore(60, time.Minute)
	ts := time.Now()
	h.Record("a:50051", makeHeatmapResult(ts, 5*time.Millisecond, nil))
	h.Record("b:50051", makeHeatmapResult(ts, 5*time.Millisecond, nil))

	targets := h.Targets()
	if len(targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(targets))
	}
}
