package ui

import (
	"strings"
	"testing"
)

func TestSparkline_Empty(t *testing.T) {
	sl := NewSparkline(8)
	out := sl.Render(nil)
	if len([]rune(out)) != 8 {
		t.Fatalf("expected width 8, got %d", len([]rune(out)))
	}
}

func TestSparkline_AllZeros(t *testing.T) {
	sl := NewSparkline(5)
	out := sl.Render([]float64{0, 0, 0, 0, 0})
	// All zeros → all lowest bar character.
	for _, r := range out {
		if r != sparkChars[0] {
			t.Fatalf("expected all lowest bars, got %q", out)
		}
	}
}

func TestSparkline_Ascending(t *testing.T) {
	sl := NewSparkline(4)
	out := sl.Render([]float64{0, 1, 2, 3})
	runes := []rune(out)
	if len(runes) != 4 {
		t.Fatalf("expected 4 chars, got %d", len(runes))
	}
	// Last character should be the tallest bar.
	if runes[3] != sparkChars[len(sparkChars)-1] {
		t.Fatalf("expected tallest bar last, got %q", string(runes[3]))
	}
	// Bars should be non-decreasing.
	for i := 1; i < len(runes); i++ {
		if runes[i] < runes[i-1] {
			t.Fatalf("bars not non-decreasing at index %d", i)
		}
	}
}

func TestSparkline_TruncatesToWidth(t *testing.T) {
	sl := NewSparkline(3)
	out := sl.Render([]float64{1, 2, 3, 4, 5, 6})
	if len([]rune(out)) != 3 {
		t.Fatalf("expected width 3, got %d", len([]rune(out)))
	}
}

func TestSparkline_PadsToWidth(t *testing.T) {
	sl := NewSparkline(6)
	out := sl.Render([]float64{1, 2})
	if len([]rune(out)) != 6 {
		t.Fatalf("expected width 6, got %d", len([]rune(out)))
	}
}

func TestSparkline_ContainsSparkChars(t *testing.T) {
	sl := NewSparkline(5)
	out := sl.Render([]float64{10, 20, 30, 40, 50})
	for _, r := range out {
		if !strings.ContainsRune(string(sparkChars), r) {
			t.Fatalf("unexpected character %q in sparkline", r)
		}
	}
}

func TestNewSparkline_ZeroWidthDefaults(t *testing.T) {
	sl := NewSparkline(0)
	if sl.Width != 10 {
		t.Fatalf("expected default width 10, got %d", sl.Width)
	}
}
