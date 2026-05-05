package probe

import (
	"sync"
	"time"
)

// DeadlineEntry holds the configured probe deadline and last breach time for a target.
type DeadlineEntry struct {
	Deadline    time.Duration
	LastBreach  time.Time
	BreachCount int
}

// DeadlineTracker tracks per-target probe deadlines and records when probes
// exceed their allowed duration.
type DeadlineTracker struct {
	mu      sync.Mutex
	entries map[string]*DeadlineEntry
	default_ time.Duration
}

// NewDeadlineTracker creates a DeadlineTracker with the given default deadline.
func NewDeadlineTracker(defaultDeadline time.Duration) *DeadlineTracker {
	return &DeadlineTracker{
		entries:  make(map[string]*DeadlineEntry),
		default_: defaultDeadline,
	}
}

// SetDeadline configures a per-target deadline, overriding the default.
func (d *DeadlineTracker) SetDeadline(target string, deadline time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if e, ok := d.entries[target]; ok {
		e.Deadline = deadline
		return
	}
	d.entries[target] = &DeadlineEntry{Deadline: deadline}
}

// DeadlineFor returns the effective deadline for a target.
func (d *DeadlineTracker) DeadlineFor(target string) time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	if e, ok := d.entries[target]; ok {
		return e.Deadline
	}
	return d.default_
}

// RecordBreach marks that a probe for target exceeded its deadline at t.
func (d *DeadlineTracker) RecordBreach(target string, t time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	e, ok := d.entries[target]
	if !ok {
		e = &DeadlineEntry{Deadline: d.default_}
		d.entries[target] = e
	}
	e.LastBreach = t
	e.BreachCount++
}

// Get returns the DeadlineEntry for a target and whether it exists.
func (d *DeadlineTracker) Get(target string) (DeadlineEntry, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	e, ok := d.entries[target]
	if !ok {
		return DeadlineEntry{}, false
	}
	return *e, true
}

// All returns a snapshot of all tracked entries keyed by target.
func (d *DeadlineTracker) All() map[string]DeadlineEntry {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make(map[string]DeadlineEntry, len(d.entries))
	for k, v := range d.entries {
		out[k] = *v
	}
	return out
}
