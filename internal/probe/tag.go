package probe

import "sync"

// TagStore holds arbitrary key-value tags associated with each probe target.
// Tags are user-defined labels (e.g. env=prod, region=us-east) that can be
// attached via config and surfaced in the UI.
type TagStore struct {
	mu   sync.RWMutex
	tags map[string]map[string]string // target address -> tag map
}

// NewTagStore returns an initialised TagStore.
func NewTagStore() *TagStore {
	return &TagStore{
		tags: make(map[string]map[string]string),
	}
}

// Set replaces the full tag map for a target.
func (s *TagStore) Set(target string, tags map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := make(map[string]string, len(tags))
	for k, v := range tags {
		copy[k] = v
	}
	s.tags[target] = copy
}

// Get returns a copy of the tags for a target.
// Returns nil if no tags are registered for the target.
func (s *TagStore) Get(target string) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src, ok := s.tags[target]
	if !ok {
		return nil
	}
	copy := make(map[string]string, len(src))
	for k, v := range src {
		copy[k] = v
	}
	return copy
}

// All returns a snapshot of all target→tags mappings.
func (s *TagStore) All() map[string]map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]map[string]string, len(s.tags))
	for target, src := range s.tags {
		copy := make(map[string]string, len(src))
		for k, v := range src {
			copy[k] = v
		}
		out[target] = copy
	}
	return out
}

// Delete removes all tags for a target.
func (s *TagStore) Delete(target string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tags, target)
}
