package probe

import (
	"sync"
	"time"
)

// UptimeTracker tracks the uptime percentage for each target based on
// the ratio of successful (SERVING) probes to total probes.
type UptimeTracker struct {
	mu      sync.Mutex
	records map[string]*uptimeRecord
}

type uptimeRecord struct {
	total   int
	healthy int
	window  time.Duration
	start   time.Time
}

// UptimeSummary holds uptime statistics for a single target.
type UptimeSummary struct {
	Target  string
	Percent float64
	Total   int
	Healthy int
	Since   time.Time
}

// NewUptimeTracker creates a new UptimeTracker with the given rolling window.
func NewUptimeTracker(window time.Duration) *UptimeTracker {
	return &UptimeTracker{
		records: make(map[string]*uptimeRecord),
		}
}

// Record registers a probe result for the given target.
func (u *UptimeTracker) Record(target string, r Result) {
	u.mu.Lock()
	defer u.mu.Unlock()

	rec, ok := u.records[target]
	if !ok {
		rec = &uptimeRecord{start: time.Now()}
		u.records[target] = rec
	}

	rec.total++
	if r.Status == StatusServing {
		rec.healthy++
	}
}

// Summary returns the uptime summary for the given target.
// Returns zero-value summary and false if the target has no records.
func (u *UptimeTracker) Summary(target string) (UptimeSummary, bool) {
	u.mu.Lock()
	defer u.mu.Unlock()

	rec, ok := u.records[target]
	if !ok || rec.total == 0 {
		return UptimeSummary{}, false
	}

	pct := float64(rec.healthy) / float64(rec.total) * 100.0
	return UptimeSummary{
		Target:  target,
		Percent: pct,
		Total:   rec.total,
		Healthy: rec.healthy,
		Since:   rec.start,
	}, true
}

// All returns summaries for every tracked target.
func (u *UptimeTracker) All() []UptimeSummary {
	u.mu.Lock()
	defer u.mu.Unlock()

	out := make([]UptimeSummary, 0, len(u.records))
	for target, rec := range u.records {
		if rec.total == 0 {
			continue
		}
		pct := float64(rec.healthy) / float64(rec.total) * 100.0
		out = append(out, UptimeSummary{
			Target:  target,
			Percent: pct,
			Total:   rec.total,
			Healthy: rec.healthy,
			Since:   rec.start,
		})
	}
	return out
}

// Reset clears all records for the given target.
func (u *UptimeTracker) Reset(target string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.records, target)
}
