package ui

import (
	"testing"
	"time"
)

type stubBudget struct {
	all       map[string]time.Duration
	remaining map[string]time.Duration
	exceeded  map[string]bool
}

func (s *stubBudget) All() map[string]time.Duration { return s.all }
func (s *stubBudget) Remaining(t string) time.Duration { return s.remaining[t] }
func (s *stubBudget) Exceeded(t string) bool          { return s.exceeded[t] }

func makeStubBudget(target string, spent, remaining time.Duration, exceeded bool) *stubBudget {
	return &stubBudget{
		all:       map[string]time.Duration{target: spent},
		remaining: map[string]time.Duration{target: remaining},
		exceeded:  map[string]bool{target: exceeded},
	}
}

func TestLatencyBudgetPanel_Title(t *testing.T) {
	p := NewLatencyBudgetPanel(&stubBudget{
		all:       map[string]time.Duration{},
		remaining: map[string]time.Duration{},
		exceeded:  map[string]bool{},
	})
	if p.table.GetTitle() != " Latency Budget " {
		t.Errorf("unexpected title: %q", p.table.GetTitle())
	}
}

func TestLatencyBudgetPanel_NoData(t *testing.T) {
	p := NewLatencyBudgetPanel(&stubBudget{
		all:       map[string]time.Duration{},
		remaining: map[string]time.Duration{},
		exceeded:  map[string]bool{},
	})
	p.Update()
	// Only header row present.
	if p.table.GetRowCount() != 1 {
		t.Errorf("expected 1 row (header), got %d", p.table.GetRowCount())
	}
}

func TestLatencyBudgetPanel_ShowsTarget(t *testing.T) {
	s := makeStubBudget("svc:443", 30*time.Millisecond, 70*time.Millisecond, false)
	p := NewLatencyBudgetPanel(s)
	p.Update()
	cell := p.table.GetCell(1, 0)
	if cell.Text != "svc:443" {
		t.Errorf("expected target cell to be svc:443, got %q", cell.Text)
	}
}

func TestLatencyBudgetPanel_StatusOK(t *testing.T) {
	s := makeStubBudget("svc:443", 30*time.Millisecond, 70*time.Millisecond, false)
	p := NewLatencyBudgetPanel(s)
	p.Update()
	cell := p.table.GetCell(1, 3)
	if cell.Text != "OK" {
		t.Errorf("expected status OK, got %q", cell.Text)
	}
}

func TestLatencyBudgetPanel_StatusExceeded(t *testing.T) {
	s := makeStubBudget("svc:443", 120*time.Millisecond, -20*time.Millisecond, true)
	p := NewLatencyBudgetPanel(s)
	p.Update()
	cell := p.table.GetCell(1, 3)
	if cell.Text != "EXCEEDED" {
		t.Errorf("expected status EXCEEDED, got %q", cell.Text)
	}
}

func TestLatencyBudgetPanel_TruncatesLongTarget(t *testing.T) {
	long := "very-long-service-name-that-exceeds-the-display-limit.example.com:50051"
	s := makeStubBudget(long, 10*time.Millisecond, 90*time.Millisecond, false)
	p := NewLatencyBudgetPanel(s)
	p.Update()
	cell := p.table.GetCell(1, 0)
	if len(cell.Text) > 31 {
		t.Errorf("expected truncated target, got length %d: %q", len(cell.Text), cell.Text)
	}
}
