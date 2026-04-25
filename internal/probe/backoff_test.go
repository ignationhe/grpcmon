package probe

import (
	"testing"
	"time"
)

func TestDefaultBackoffPolicy(t *testing.T) {
	b := DefaultBackoffPolicy()
	if b.InitialDelay != 100*time.Millisecond {
		t.Errorf("expected InitialDelay 100ms, got %v", b.InitialDelay)
	}
	if b.MaxDelay != 5*time.Second {
		t.Errorf("expected MaxDelay 5s, got %v", b.MaxDelay)
	}
	if b.Multiplier != 2.0 {
		t.Errorf("expected Multiplier 2.0, got %v", b.Multiplier)
	}
}

func TestBackoffPolicy_Delay_Increases(t *testing.T) {
	b := BackoffPolicy{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       0,
	}
	prev := b.Delay(0)
	for i := 1; i <= 5; i++ {
		curr := b.Delay(i)
		if curr <= prev {
			t.Errorf("attempt %d: expected delay %v > %v", i, curr, prev)
		}
		prev = curr
	}
}

func TestBackoffPolicy_Delay_CappedAtMax(t *testing.T) {
	b := BackoffPolicy{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   10.0,
		Jitter:       0,
	}
	for _, attempt := range []int{5, 10, 20} {
		d := b.Delay(attempt)
		if d > b.MaxDelay {
			t.Errorf("attempt %d: delay %v exceeds MaxDelay %v", attempt, d, b.MaxDelay)
		}
	}
}

func TestBackoffPolicy_Delay_NegativeAttempt(t *testing.T) {
	b := DefaultBackoffPolicy()
	d := b.Delay(-1)
	expected := b.Delay(0)
	if d != expected {
		t.Errorf("expected negative attempt to equal attempt 0: got %v want %v", d, expected)
	}
}

func TestBackoffPolicy_Delay_WithJitter(t *testing.T) {
	b := BackoffPolicy{
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.2,
	}
	d := b.Delay(0)
	// With jitter, delay should be >= InitialDelay and <= InitialDelay*(1+Jitter)
	if d < b.InitialDelay {
		t.Errorf("delay %v should be >= InitialDelay %v", d, b.InitialDelay)
	}
	max := time.Duration(float64(b.InitialDelay) * (1 + b.Jitter))
	if d > max {
		t.Errorf("delay %v should be <= %v", d, max)
	}
}
