package probe

import "time"

// TrendDirection indicates whether latency is improving, degrading, or stable.
type TrendDirection int

const (
	TrendStable    TrendDirection = iota
	TrendImproving                // latency decreasing
	TrendDegrading                // latency increasing
)

// TrendSample is a single data point used for trend analysis.
type TrendSample struct {
	At      time.Time
	Latency time.Duration
}

// TrendResult summarises the computed trend for a target.
type TrendResult struct {
	Target    string
	Direction TrendDirection
	Slope     float64 // nanoseconds per second
}

// TrendAnalyzer computes latency trends using simple linear regression
// over recent probe history.
type TrendAnalyzer struct {
	minSamples int
}

// NewTrendAnalyzer returns a TrendAnalyzer that requires at least minSamples
// data points before producing a non-stable result.
func NewTrendAnalyzer(minSamples int) *TrendAnalyzer {
	if minSamples < 2 {
		minSamples = 2
	}
	return &TrendAnalyzer{minSamples: minSamples}
}

// Analyze computes the trend direction for the given target using samples
// drawn from h. Returns TrendStable when insufficient data is available.
func (a *TrendAnalyzer) Analyze(target string, h *History) TrendResult {
	entries := h.Entries()
	var samples []TrendSample
	for _, e := range entries {
		if e.Target == target && e.Err == nil {
			samples = append(samples, TrendSample{At: e.At, Latency: e.Latency})
		}
	}
	if len(samples) < a.minSamples {
		return TrendResult{Target: target, Direction: TrendStable}
	}
	slope := linearRegressionSlope(samples)
	dir := TrendStable
	const threshold = 1e6 // 1 ms/s
	switch {
	case slope > threshold:
		dir = TrendDegrading
	case slope < -threshold:
		dir = TrendImproving
	}
	return TrendResult{Target: target, Direction: dir, Slope: slope}
}

// linearRegressionSlope returns dy/dx where x is elapsed seconds and y is
// latency in nanoseconds.
func linearRegressionSlope(samples []TrendSample) float64 {
	n := float64(len(samples))
	t0 := samples[0].At
	var sumX, sumY, sumXY, sumX2 float64
	for _, s := range samples {
		x := s.At.Sub(t0).Seconds()
		y := float64(s.Latency.Nanoseconds())
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}
