package probe

import (
	"testing"
	"time"
)

// TestRetryBudget_IntegrationWithScheduler verifies that the RetryBudget
// correctly tracks retries across simulated scheduler ticks.
func TestRetryBudget_IntegrationWithScheduler(t *testing.T) {
	now := time.Now()
	rb := NewRetryBudget(30*time.Second, 3)
	rb.now = func() time.Time { return now }

	target := "svc-integration:443"

	// Simulate three retry attempts within the window
	for i := 0; i < 3; i++ {
		if !rb.Allow(target) {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}

	// Fourth attempt should be blocked
	if rb.Allow(target) {
		t.Fatal("fourth attempt should be blocked")
	}

	if rb.Remaining(target) != 0 {
		t.Fatalf("expected 0 remaining, got %d", rb.Remaining(target))
	}

	// Advance time past the window
	now = now.Add(31 * time.Second)

	if rb.Remaining(target) != 3 {
		t.Fatalf("expected 3 remaining after window expiry, got %d", rb.Remaining(target))
	}

	if !rb.Allow(target) {
		t.Fatal("expected allow after window expiry")
	}
}

// TestRetryBudget_MultiTarget verifies independent budgets per target.
func TestRetryBudget_MultiTarget(t *testing.T) {
	rb := NewRetryBudget(time.Minute, 2)

	targets := []string{"svc-a:443", "svc-b:443", "svc-c:443"}
	for _, tgt := range targets {
		if !rb.Allow(tgt) {
			t.Errorf("first allow for %s should succeed", tgt)
		}
		if !rb.Allow(tgt) {
			t.Errorf("second allow for %s should succeed", tgt)
		}
		if rb.Allow(tgt) {
			t.Errorf("third allow for %s should be blocked", tgt)
		}
	}

	all := rb.Targets()
	if len(all) != len(targets) {
		t.Fatalf("expected %d targets, got %d", len(targets), len(all))
	}
}
