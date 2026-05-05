package probe

import (
	"sync"
	"time"
)

// ReconnectTracker records the number of reconnection attempts and the last
// reconnect time for each target address. It is safe for concurrent use.
type ReconnectTracker struct {
	mu      sync.Mutex
	records map[string]*reconnectRecord
}

type reconnectRecord struct {
	count    int
	lastSeen time.Time
}

// NewReconnectTracker returns an initialised ReconnectTracker.
func NewReconnectTracker() *ReconnectTracker {
	return &ReconnectTracker{
		records: make(map[string]*reconnectRecord),
	}
}

// Record increments the reconnect counter for the given target and updates
// the timestamp to now.
func (r *ReconnectTracker) Record(target string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec, ok := r.records[target]
	if !ok {
		rec = &reconnectRecord{}
		r.records[target] = rec
	}
	rec.count++
	rec.lastSeen = time.Now()
}

// Count returns the total number of reconnects recorded for target.
func (r *ReconnectTracker) Count(target string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	if rec, ok := r.records[target]; ok {
		return rec.count
	}
	return 0
}

// LastSeen returns the time of the most recent reconnect for target and
// whether any reconnect has been recorded.
func (r *ReconnectTracker) LastSeen(target string) (time.Time, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if rec, ok := r.records[target]; ok && rec.count > 0 {
		return rec.lastSeen, true
	}
	return time.Time{}, false
}

// Reset clears all reconnect data for target.
func (r *ReconnectTracker) Reset(target string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.records, target)
}

// Targets returns the list of targets that have at least one recorded
// reconnect.
func (r *ReconnectTracker) Targets() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]string, 0, len(r.records))
	for k := range r.records {
		out = append(out, k)
	}
	return out
}
