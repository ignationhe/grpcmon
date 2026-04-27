package probe

import (
	"sync"
	"time"
)

// Deduplicator suppresses repeated alerts for the same target within a
// cooldown window, preventing alert storms from flooding the UI.
type Deduplicator struct {
	mu       sync.Mutex
	cooldown time.Duration
	last     map[string]time.Time
	now      func() time.Time
}

// NewDeduplicator creates a Deduplicator that suppresses repeated alerts
// for the same target within the given cooldown duration.
func NewDeduplicator(cooldown time.Duration) *Deduplicator {
	return &Deduplicator{
		cooldown: cooldown,
		last:     make(map[string]time.Time),
		now:      time.Now,
	}
}

// IsDuplicate returns true if an alert for the given target was already
// seen within the cooldown window. If not a duplicate, it records the
// current time for that target and returns false.
func (d *Deduplicator) IsDuplicate(target string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	if t, ok := d.last[target]; ok && now.Sub(t) < d.cooldown {
		return true
	}
	d.last[target] = now
	return false
}

// Reset clears the recorded time for a specific target, allowing the
// next alert for that target to pass through immediately.
func (d *Deduplicator) Reset(target string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.last, target)
}

// ResetAll clears all recorded targets.
func (d *Deduplicator) ResetAll() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.last = make(map[string]time.Time)
}

// ActiveTargets returns the number of targets currently tracked within
// the cooldown window. Targets whose cooldown has expired are not counted.
func (d *Deduplicator) ActiveTargets() int {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	count := 0
	for _, t := range d.last {
		if now.Sub(t) < d.cooldown {
			count++
		}
	}
	return count
}
