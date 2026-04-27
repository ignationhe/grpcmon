package probe

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestExporter_WriteJSON_ContainsTargets(t *testing.T) {
	agg := NewAggregator()
	agg.Record("localhost:50051", Result{
		Address:   "localhost:50051",
		Status:    StatusServing,
		Latency:   25 * time.Millisecond,
		CheckedAt: time.Now(),
	})

	ex := NewExporter(agg)
	var buf bytes.Buffer
	if err := ex.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	var out SnapshotExport
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if len(out.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(out.Targets))
	}
	if out.Targets[0].Address != "localhost:50051" {
		t.Errorf("expected address localhost:50051, got %s", out.Targets[0].Address)
	}
}

func TestExporter_WriteJSON_LatencyInMilliseconds(t *testing.T) {
	agg := NewAggregator()
	agg.Record("svc:9090", Result{
		Address:   "svc:9090",
		Status:    StatusServing,
		Latency:   100 * time.Millisecond,
		CheckedAt: time.Now(),
	})

	ex := NewExporter(agg)
	var buf bytes.Buffer
	_ = ex.WriteJSON(&buf)

	var out SnapshotExport
	_ = json.Unmarshal(buf.Bytes(), &out)

	if out.Targets[0].Latency != 100.0 {
		t.Errorf("expected latency 100ms, got %f", out.Targets[0].Latency)
	}
}

func TestExporter_WriteJSON_TimestampPresent(t *testing.T) {
	agg := NewAggregator()
	ex := NewExporter(agg)
	var buf bytes.Buffer
	_ = ex.WriteJSON(&buf)

	var out SnapshotExport
	_ = json.Unmarshal(buf.Bytes(), &out)

	if out.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp in export")
	}
}

func TestExporter_WriteJSON_EmptyAggregator(t *testing.T) {
	agg := NewAggregator()
	ex := NewExporter(agg)
	var buf bytes.Buffer
	if err := ex.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON on empty aggregator returned error: %v", err)
	}

	var out SnapshotExport
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(out.Targets) != 0 {
		t.Errorf("expected 0 targets, got %d", len(out.Targets))
	}
}
