package probe

import (
	"testing"
	"time"
)

func buildAnomalyHistory(results []Result) *History {
	h := NewHistory(50)
	for _, r := range results {
		h.Add(r)
	}
	return h
}

func okAnomalyResult(latency time.Duration) Result {
	return Result{Target: "svc", Status: StatusServing, Latency: latency}
}

func errAnomalyResult() Result {
	return Result{Target: "svc", Status: StatusNotServing, Err: errSentinel}
}

var errSentinel = &probeError{"sentinel"}

type probeError struct{ msg string }

func (e *probeError) Error() string { return e.msg }

func makeAnomalyDetector(p50 time.Duration) (*AnomalyDetector, *BaselineStore) {
	bs := NewBaselineStore()
	if p50 > 0 {
		var results []Result
		for i := 0; i < 10; i++ {
			results = append(results, okAnomalyResult(p50))
		}
		bs.Compute("svc", buildAnomalyHistory(results))
	}
	return NewAnomalyDetector(bs, 2.0, 0.5), bs
}

func TestAnomalyDetector_NoAnomaly(t *testing.T) {
	det, _ := makeAnomalyDetector(100 * time.Millisecond)
	h := buildAnomalyHistory([]Result{
		okAnomalyResult(80 * time.Millisecond),
		okAnomalyResult(90 * time.Millisecond),
	})
	anoms := det.Evaluate("svc", h)
	if len(anoms) != 0 {
		t.Fatalf("expected no anomalies, got %d", len(anoms))
	}
}

func TestAnomalyDetector_LatencySpike(t *testing.T) {
	det, _ := makeAnomalyDetector(100 * time.Millisecond)
	h := buildAnomalyHistory([]Result{
		okAnomalyResult(300 * time.Millisecond),
		okAnomalyResult(350 * time.Millisecond),
	})
	anoms := det.Evaluate("svc", h)
	if len(anoms) == 0 {
		t.Fatal("expected latency spike anomaly")
	}
	if anoms[0].Kind != AnomalyLatencySpike {
		t.Fatalf("expected %s, got %s", AnomalyLatencySpike, anoms[0].Kind)
	}
}

func TestAnomalyDetector_ErrorBurst(t *testing.T) {
	det, _ := makeAnomalyDetector(0)
	var results []Result
	for i := 0; i < 6; i++ {
		results = append(results, errAnomalyResult())
	}
	for i := 0; i < 4; i++ {
		results = append(results, okAnomalyResult(10*time.Millisecond))
	}
	h := buildAnomalyHistory(results)
	anoms := det.Evaluate("svc", h)
	if len(anoms) == 0 {
		t.Fatal("expected error burst anomaly")
	}
	if anoms[0].Kind != AnomalyErrorBurst {
		t.Fatalf("expected %s, got %s", AnomalyErrorBurst, anoms[0].Kind)
	}
}

func TestAnomalyDetector_NoBaselineSkipsLatencyCheck(t *testing.T) {
	bs := NewBaselineStore()
	det := NewAnomalyDetector(bs, 2.0, 0.9)
	h := buildAnomalyHistory([]Result{
		okAnomalyResult(999 * time.Millisecond),
	})
	anoms := det.Evaluate("svc", h)
	for _, a := range anoms {
		if a.Kind == AnomalyLatencySpike {
			t.Fatal("should not detect spike without baseline")
		}
	}
}
