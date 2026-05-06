package ui

import (
	"strings"
	"testing"
)

type stubRetryBudget struct {
	targets   []string
	remaining map[string]int
}

func (s *stubRetryBudget) Targets() []string { return s.targets }
func (s *stubRetryBudget) Remaining(t string) int {
	return s.remaining[t]
}

func makeRetryBudget(data map[string]int) *stubRetryBudget {
	targets := make([]string, 0, len(data))
	for k := range data {
		targets = append(targets, k)
	}
	return &stubRetryBudget{targets: targets, remaining: data}
}

func TestRetryBudgetPanel_Title(t *testing.T) {
	p := NewRetryBudgetPanel(makeRetryBudget(nil), 5)
	title := p.tv.GetTitle()
	if !strings.Contains(title, "Retry Budget") {
		t.Errorf("expected title to contain 'Retry Budget', got %q", title)
	}
}

func TestRetryBudgetPanel_NoTargets(t *testing.T) {
	p := NewRetryBudgetPanel(makeRetryBudget(nil), 5)
	p.Update()
	text := p.tv.GetText(false)
	if !strings.Contains(text, "no retry activity") {
		t.Errorf("expected 'no retry activity', got %q", text)
	}
}

func TestRetryBudgetPanel_ShowsTarget(t *testing.T) {
	p := NewRetryBudgetPanel(makeRetryBudget(map[string]int{"svc-a:443": 3}), 5)
	p.Update()
	text := p.tv.GetText(false)
	if !strings.Contains(text, "svc-a:443") {
		t.Errorf("expected target name in output, got %q", text)
	}
}

func TestRetryBudgetPanel_ShowsRemainingCount(t *testing.T) {
	p := NewRetryBudgetPanel(makeRetryBudget(map[string]int{"svc-a:443": 2}), 5)
	p.Update()
	text := p.tv.GetText(false)
	if !strings.Contains(text, "2/5") {
		t.Errorf("expected '2/5' in output, got %q", text)
	}
}

func TestRetryBudgetPanel_FullBudgetGreen(t *testing.T) {
	color := retryColor(5, 5)
	if color != "green" {
		t.Errorf("expected green for full budget, got %q", color)
	}
}

func TestRetryBudgetPanel_LowBudgetRed(t *testing.T) {
	color := retryColor(1, 10)
	if color != "red" {
		t.Errorf("expected red for low budget, got %q", color)
	}
}

func TestRetryBudgetPanel_BarWidth(t *testing.T) {
	bar := retryBar(5, 5)
	if len([]rune(bar)) != 10 {
		t.Errorf("expected bar width 10, got %d", len([]rune(bar)))
	}
}

func TestRetryBudgetPanel_TruncateLongTarget(t *testing.T) {
	long := strings.Repeat("x", 40)
	out := truncateRetryTarget(long, 28)
	if len([]rune(out)) > 28 {
		t.Errorf("expected truncation to 28 chars, got %d", len([]rune(out)))
	}
}
