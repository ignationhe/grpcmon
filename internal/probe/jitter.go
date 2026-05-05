package probe

import (
	"math/rand"
	"sync"
	"time"
)

// JitterPolicy adds randomised jitter to poll intervals to prevent
// thundering-herd effects when many targets share the same interval.
type JitterPolicy struct {
	mu      sync.Mutex
	rng     *rand.Rand
	factor  float64 // fraction of interval to jitter, e.g. 0.25 = ±25%
}

// DefaultJitterPolicy returns a JitterPolicy with a 20% jitter factor.
func DefaultJitterPolicy() *JitterPolicy {
	return &JitterPolicy{
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
		factor: 0.20,
	}
}

// NewJitterPolicy returns a JitterPolicy with the given jitter factor.
// factor must be in the range [0, 1]; values outside this range are clamped.
func NewJitterPolicy(factor float64) *JitterPolicy {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	return &JitterPolicy{
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
		factor: factor,
	}
}

// Apply returns d adjusted by a random jitter in the range
// [d*(1-factor), d*(1+factor)], always returning a positive duration.
func (j *JitterPolicy) Apply(d time.Duration) time.Duration {
	if d <= 0 || j.factor == 0 {
		return d
	}
	j.mu.Lock()
	// random value in [-factor, +factor]
	offset := (j.rng.Float64()*2 - 1) * j.factor
	j.mu.Unlock()

	adjusted := time.Duration(float64(d) * (1 + offset))
	if adjusted <= 0 {
		adjusted = 1
	}
	return adjusted
}

// Factor returns the configured jitter factor.
func (j *JitterPolicy) Factor() float64 {
	return j.factor
}
