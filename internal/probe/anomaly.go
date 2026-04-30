package probe

import (
	"fmt"
	"sync"
	"time"
)

// AnomalyKind describes the type of anomaly detected.
type AnomalyKind string

const (
	AnomalyLatencySpike AnomalyKind = "latency_spike"
	AnomalyErrorBurst  AnomalyKind = "error_burst"
)

// Anomaly represents a detected anomaly for a target.
type Anomaly struct {
	Target    string
	Kind      AnomalyKind
	Message   string
	DetectedAt time.Time
}

// AnomalyDetector detects latency spikes and error bursts relative to a baseline.
type AnomalyDetector struct {
	mu        sync.Mutex
	baseline  *BaselineStore
	latencyMul float64 // multiplier above baseline p50 to trigger spike
	errorThresh float64 // error rate threshold for burst
}

// NewAnomalyDetector creates a detector using the given baseline store.
// latencyMul is how many times the baseline p50 a reading must exceed to be a spike.
// errorThresh is the error rate (0-1) above which an error burst is declared.
func NewAnomalyDetector(baseline *BaselineStore, latencyMul float64, errorThresh float64) *AnomalyDetector {
	return &AnomalyDetector{
		baseline:    baseline,
		latencyMul:  latencyMul,
		errorThresh: errorThresh,
	}
}

// Evaluate checks the history for a target and returns any detected anomalies.
func (d *AnomalyDetector) Evaluate(target string, h *History) []Anomaly {
	d.mu.Lock()
	defer d.mu.Unlock()

	var anomalies []Anomaly
	now := time.Now()

	bl, ok := d.baseline.Get(target)
	if ok && bl.P50 > 0 {
		avg := h.AvgLatency()
		threshold := time.Duration(float64(bl.P50) * d.latencyMul)
		if avg > threshold {
			anomalies = append(anomalies, Anomaly{
				Target:     target,
				Kind:       AnomalyLatencySpike,
				Message:    fmt.Sprintf("avg latency %v exceeds %.1fx baseline p50 %v", avg.Round(time.Millisecond), d.latencyMul, bl.P50.Round(time.Millisecond)),
				DetectedAt: now,
			})
		}
	}

	errRate := h.ErrorRate()
	if errRate > d.errorThresh {
		anomalies = append(anomalies, Anomaly{
			Target:     target,
			Kind:       AnomalyErrorBurst,
			Message:    fmt.Sprintf("error rate %.0f%% exceeds threshold %.0f%%", errRate*100, d.errorThresh*100),
			DetectedAt: now,
		})
	}

	return anomalies
}
