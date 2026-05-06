package probe

import "time"

// HealthScore represents a composite score [0.0, 1.0] for a target.
type HealthScore struct {
	Target    string
	Score     float64 // 1.0 = fully healthy, 0.0 = completely unhealthy
	ComputedAt time.Time
}

// HealthScorer computes a composite health score from uptime and error rate.
type HealthScorer struct {
	uptime  *UptimeTracker
	window  *WindowAggregator
	weights scoreWeights
}

type scoreWeights struct {
	UptimeWeight     float64
	ErrorRateWeight  float64
}

// DefaultScoreWeights returns balanced weights that sum to 1.0.
func DefaultScoreWeights() scoreWeights {
	return scoreWeights{
		UptimeWeight:    0.5,
		ErrorRateWeight: 0.5,
	}
}

// NewHealthScorer creates a HealthScorer backed by the given trackers.
func NewHealthScorer(uptime *UptimeTracker, window *WindowAggregator) *HealthScorer {
	return &HealthScorer{
		uptime:  uptime,
		window:  window,
		weights: DefaultScoreWeights(),
	}
}

// Compute returns a HealthScore for the given target.
// Score = (uptimeFraction * w1) + ((1 - errorRate) * w2)
func (h *HealthScorer) Compute(target string) HealthScore {
	uptimeFrac := h.uptime.UptimeFraction(target)

	stats := h.window.Stats(target)
	errorFrac := 0.0
	if stats.Total > 0 {
		errorFrac = stats.ErrorRate
	}

	score := (uptimeFrac * h.weights.UptimeWeight) +
		((1.0 - errorFrac) * h.weights.ErrorRateWeight)

	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return HealthScore{
		Target:     target,
		Score:      score,
		ComputedAt: time.Now(),
	}
}

// All returns HealthScores for every target known to the uptime tracker.
func (h *HealthScorer) All() []HealthScore {
	all := h.uptime.All()
	scores := make([]HealthScore, 0, len(all))
	for _, rec := range all {
		scores = append(scores, h.Compute(rec.Target))
	}
	return scores
}
