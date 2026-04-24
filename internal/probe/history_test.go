package probe

import (
	"errors"
	"testing"
	"time"
)

func TestHistory_AddAndEntries(t *testing.T) {
	h := NewHistory(3)

	h.Add(Result{Latency: 10 * time.Millisecond})
	h.Add(Result{Latency: 20 * time.Millisecond})

	entries := h.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Result.Latency != 10*time.Millisecond {
		t.Errorf("unexpected first entry latency: %v", entries[0].Result.Latency)
	}
}

func TestHistory_Eviction(t *testing.T) {
	h := NewHistory(3)

	for i := 1; i <= 4; i++ {
		h.Add(Result{Latency: time.Duration(i) * time.Millisecond})
	}

	entries := h.Entries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(entries))
	}
	// oldest entry (1ms) should have been evicted; first should now be 2ms
	if entries[0].Result.Latency != 2*time.Millisecond {
		t.Errorf("expected oldest to be 2ms, got %v", entries[0].Result.Latency)
	}
}

func TestHistory_AvgLatency(t *testing.T) {
	h := NewHistory(10)

	h.Add(Result{Latency: 10 * time.Millisecond})
	h.Add(Result{Latency: 30 * time.Millisecond})
	h.Add(Result{Err: errors.New("timeout"), Latency: 0}) // should be excluded

	avg := h.AvgLatency()
	expected := 20 * time.Millisecond
	if avg != expected {
		t.Errorf("expected avg latency %v, got %v", expected, avg)
	}
}

func TestHistory_AvgLatency_Empty(t *testing.T) {
	h := NewHistory(10)
	if h.AvgLatency() != 0 {
		t.Error("expected 0 avg latency for empty history")
	}
}

func TestHistory_ErrorRate(t *testing.T) {
	h := NewHistory(10)

	h.Add(Result{})
	h.Add(Result{})
	h.Add(Result{Err: errors.New("fail")})
	h.Add(Result{Err: errors.New("fail")})

	rate := h.ErrorRate()
	if rate != 0.5 {
		t.Errorf("expected error rate 0.5, got %f", rate)
	}
}

func TestHistory_ErrorRate_Empty(t *testing.T) {
	h := NewHistory(10)
	if h.ErrorRate() != 0 {
		t.Error("expected 0 error rate for empty history")
	}
}

func TestNewHistory_DefaultMaxSize(t *testing.T) {
	h := NewHistory(0)
	if h.maxSize != 60 {
		t.Errorf("expected default maxSize 60, got %d", h.maxSize)
	}
}
