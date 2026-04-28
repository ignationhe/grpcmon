package probe

import (
	"sync"
	"time"
)

// HealthEvent records a single status transition for a target.
type HealthEvent struct {
	Target    string
	Previous  string
	Current   string
	OccurredAt time.Time
}

// HealthLog tracks status transitions across probe results.
type HealthLog struct {
	mu       sync.Mutex
	events   []HealthEvent
	lastSeen map[string]string
	maxSize  int
}

// NewHealthLog creates a HealthLog that retains up to maxSize events.
func NewHealthLog(maxSize int) *HealthLog {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &HealthLog{
		lastSeen: make(map[string]string),
		maxSize:  maxSize,
	}
}

// Record evaluates whether the result represents a status change and, if so,
// appends a HealthEvent to the log.
func (h *HealthLog) Record(r Result) {
	h.mu.Lock()
	defer h.mu.Unlock()

	current := r.Status
	prev, seen := h.lastSeen[r.Target]
	h.lastSeen[r.Target] = current

	if !seen || prev == current {
		return
	}

	event := HealthEvent{
		Target:     r.Target,
		Previous:   prev,
		Current:    current,
		OccurredAt: r.Timestamp,
	}

	h.events = append(h.events, event)
	if len(h.events) > h.maxSize {
		h.events = h.events[len(h.events)-h.maxSize:]
	}
}

// Events returns a copy of all recorded health transition events.
func (h *HealthLog) Events() []HealthEvent {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]HealthEvent, len(h.events))
	copy(out, h.events)
	return out
}

// LastStatus returns the most recently recorded status for a target,
// and whether the target has been seen at all.
func (h *HealthLog) LastStatus(target string) (string, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	s, ok := h.lastSeen[target]
	return s, ok
}
