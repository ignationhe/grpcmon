package probe

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// SnapshotExport is a serialisable representation of an aggregator snapshot.
type SnapshotExport struct {
	Timestamp time.Time        `json:"timestamp"`
	Targets   []TargetExport   `json:"targets"`
}

// TargetExport holds exported metrics for a single target.
type TargetExport struct {
	Address    string        `json:"address"`
	Status     string        `json:"status"`
	Latency    float64       `json:"latency_ms"`
	AvgLatency float64       `json:"avg_latency_ms"`
	ErrorRate  float64       `json:"error_rate"`
	CheckedAt  time.Time     `json:"checked_at"`
}

// Exporter writes aggregator snapshots to an io.Writer in a structured format.
type Exporter struct {
	agg *Aggregator
}

// NewExporter creates an Exporter backed by the given Aggregator.
func NewExporter(agg *Aggregator) *Exporter {
	return &Exporter{agg: agg}
}

// WriteJSON serialises the current aggregator snapshot as JSON to w.
func (e *Exporter) WriteJSON(w io.Writer) error {
	snap := e.agg.Snapshot()
	export := SnapshotExport{
		Timestamp: time.Now().UTC(),
		Targets:   make([]TargetExport, 0, len(snap)),
	}

	for addr, result := range snap {
		h := e.agg.History(addr)
		var avg, errRate float64
		if h != nil {
			avg = float64(h.AvgLatency()) / float64(time.Millisecond)
			errRate = h.ErrorRate()
		}
		export.Targets = append(export.Targets, TargetExport{
			Address:    addr,
			Status:     fmt.Sprintf("%v", result.Status),
			Latency:    float64(result.Latency) / float64(time.Millisecond),
			AvgLatency: avg,
			ErrorRate:  errRate,
			CheckedAt:  result.CheckedAt,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(export)
}
