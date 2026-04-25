package probe

import "sync"

// Aggregator collects Results from multiple targets and provides
// a snapshot of the latest result per target.
type Aggregator struct {
	mu      sync.RWMutex
	latest  map[string]Result
	history map[string]*History
	cap     int
}

// NewAggregator creates an Aggregator that keeps up to historyCap
// results per target in its rolling history.
func NewAggregator(historyCap int) *Aggregator {
	return &Aggregator{
		latest:  make(map[string]Result),
		history: make(map[string]*History),
		cap:     historyCap,
	}
}

// Record stores a Result for the given target.
func (a *Aggregator) Record(target string, r Result) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.latest[target] = r

	h, ok := a.history[target]
	if !ok {
		h = NewHistory(a.cap)
		a.history[target] = h
	}
	h.Add(r)
}

// Latest returns the most recent Result for the given target and
// whether one exists.
func (a *Aggregator) Latest(target string) (Result, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	r, ok := a.latest[target]
	return r, ok
}

// History returns the History for the given target and whether one exists.
func (a *Aggregator) History(target string) (*History, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	h, ok := a.history[target]
	return h, ok
}

// Targets returns the list of known target addresses.
func (a *Aggregator) Targets() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	targets := make([]string, 0, len(a.latest))
	for t := range a.latest {
		targets = append(targets, t)
	}
	return targets
}

// Snapshot returns a copy of all latest results keyed by target.
func (a *Aggregator) Snapshot() map[string]Result {
	a.mu.RLock()
	defer a.mu.RUnlock()

	snap := make(map[string]Result, len(a.latest))
	for k, v := range a.latest {
		snap[k] = v
	}
	return snap
}
