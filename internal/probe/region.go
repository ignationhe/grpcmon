package probe

import (
	"sort"
	"sync"
)

// RegionSummary holds aggregated stats for a named region.
type RegionSummary struct {
	Region    string
	Targets   []string
	Healthy   int
	Unhealthy int
	ErrorRate float64 // 0.0–1.0
}

// RegionStore groups probe targets by region tag and computes per-region stats.
type RegionStore struct {
	mu      sync.RWMutex
	regions map[string][]string // region -> target addresses
}

// NewRegionStore creates an empty RegionStore.
func NewRegionStore() *RegionStore {
	return &RegionStore{
		regions: make(map[string][]string),
	}
}

// Assign associates addr with the given region, replacing any prior assignment.
func (r *RegionStore) Assign(addr, region string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove addr from any existing region.
	for reg, addrs := range r.regions {
		for i, a := range addrs {
			if a == addr {
				r.regions[reg] = append(addrs[:i], addrs[i+1:]...)
				break
			}
		}
	}
	r.regions[region] = append(r.regions[region], addr)
}

// Regions returns all known region names in sorted order.
func (r *RegionStore) Regions() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]string, 0, len(r.regions))
	for k := range r.regions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Summarise computes a RegionSummary for the given region using the provided
// Aggregator to determine current health per target.
func (r *RegionStore) Summarise(region string, agg *Aggregator) RegionSummary {
	r.mu.RLock()
	addrs := append([]string(nil), r.regions[region]...)
	r.mu.RUnlock()

	s := RegionSummary{Region: region, Targets: addrs}
	if len(addrs) == 0 {
		return s
	}

	var errCount int
	for _, addr := range addrs {
		res, ok := agg.Latest(addr)
		if !ok || res.Err != nil {
			s.Unhealthy++
			errCount++
		} else {
			s.Healthy++
		}
	}
	s.ErrorRate = float64(errCount) / float64(len(addrs))
	return s
}
