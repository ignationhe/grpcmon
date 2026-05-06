package probe

import (
	"testing"
	"time"
)

func TestRetryBudget_AllowFirstRetry(t *testing.T) {
	rb := NewRetryBudget(time.Minute, 3)
	if !rb.Allow("svc-a") {
		t.Fatal("expected first retry to be allowed")
	}
}

func TestRetryBudget_AllowUpToMax(t *testing.T) {
	rb := NewRetryBudget(time.Minute, 3)
	for i := 0; i < 3; i++ {
		if !rb.Allow("svc-a") {
			t.Fatalf("expected retry %d to be allowed", i+1)
		}
	}
}

func TestRetryBudget_BlocksWhenExhausted(t *testing.T) {
	rb := NewRetryBudget(time.Minute, 2)
	rb.Allow("svc-a")
	rb.Allow("svc-a")
	if rb.Allow("svc-a") {
		t.Fatal("expected retry to be blocked after exhaustion")
	}
}

func TestRetryBudget_AllowsAfterWindowExpiry(t *testing.T) {
	now := time.Now()
	rb := NewRetryBudget(time.Minute, 2)
	rb.now = func() time.Time { return now }

	rb.Allow("svc-a")
	rb.Allow("svc-a")

	// Advance past the window
	rb.now = func() time.Time { return now.Add(2 * time.Minute) }

	if !rb.Allow("svc-a") {
		t.Fatal("expected retry to be allowed after window expiry")
	}
}

func TestRetryBudget_DifferentTargetsIndependent(t *testing.T) {
	rb := NewRetryBudget(time.Minute, 1)
	rb.Allow("svc-a")
	if !rb.Allow("svc-b") {
		t.Fatal("expected svc-b to be independent of svc-a")
	}
}

func TestRetryBudget_Remaining(t *testing.T) {
	rb := NewRetryBudget(time.Minute, 3)
	if rb.Remaining("svc-a") != 3 {
		t.Fatalf("expected 3 remaining, got %d", rb.Remaining("svc-a"))
	}
	rb.Allow("svc-a")
	if rb.Remaining("svc-a") != 2 {
		t.Fatalf("expected 2 remaining, got %d", rb.Remaining("svc-a"))
	}
}

func TestRetryBudget_Reset_AllowsImmediateRetry(t *testing.T) {
	rb := NewRetryBudget(time.Minute, 1)
	rb.Allow("svc-a")
	rb.Reset("svc-a")
	if !rb.Allow("svc-a") {
		t.Fatal("expected retry to be allowed after reset")
	}
}

func TestRetryBudget_Targets(t *testing.T) {
	rb := NewRetryBudget(time.Minute, 5)
	rb.Allow("svc-a")
	rb.Allow("svc-b")
	targets := rb.Targets()
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
}
