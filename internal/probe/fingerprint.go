package probe

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Fingerprint uniquely identifies a probe result based on its target address
// and status, enabling change detection across successive polls.
type Fingerprint struct {
	mu    sync.RWMutex
	store map[string]string
}

// NewFingerprint creates a new Fingerprint tracker.
func NewFingerprint() *Fingerprint {
	return &Fingerprint{
		store: make(map[string]string),
	}
}

// Compute derives a deterministic hash string from a Result's key fields.
func Compute(r Result) string {
	parts := []string{
		r.Target,
		r.Status,
	}
	if r.Err != nil {
		parts = append(parts, r.Err.Error())
	}
	// Include sorted metadata keys for stability.
	keys := make([]string, 0, len(r.Metadata))
	for k := range r.Metadata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		parts = append(parts, k+"="+r.Metadata[k])
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("%x", sum[:8])
}

// Changed reports whether the fingerprint for the given target has changed
// since the last call to Changed or Set. It updates the stored fingerprint
// when a change is detected.
func (f *Fingerprint) Changed(r Result) bool {
	next := Compute(r)
	f.mu.Lock()
	defer f.mu.Unlock()
	prev, ok := f.store[r.Target]
	if !ok || prev != next {
		f.store[r.Target] = next
		return true
	}
	return false
}

// Set explicitly stores a fingerprint for a target without reporting a change.
func (f *Fingerprint) Set(target, fp string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.store[target] = fp
}

// Get returns the current stored fingerprint for a target, or empty string.
func (f *Fingerprint) Get(target string) string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.store[target]
}

// Reset clears all stored fingerprints.
func (f *Fingerprint) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.store = make(map[string]string)
}
