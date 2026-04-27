package probe

import (
	"testing"
	"time"
)

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := NewCircuitBreaker(3, 50*time.Millisecond)
	if !cb.Allow() {
		t.Fatal("expected Allow() == true when closed")
	}
	if cb.CurrentState() != StateClosed {
		t.Fatalf("expected StateClosed, got %v", cb.CurrentState())
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, 50*time.Millisecond)
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.CurrentState() != StateClosed {
		t.Fatal("should still be closed before threshold")
	}
	cb.RecordFailure()
	if cb.CurrentState() != StateOpen {
		t.Fatalf("expected StateOpen after threshold, got %v", cb.CurrentState())
	}
}

func TestCircuitBreaker_BlocksWhenOpen(t *testing.T) {
	cb := NewCircuitBreaker(1, 100*time.Millisecond)
	cb.RecordFailure()
	if cb.Allow() {
		t.Fatal("expected Allow() == false when open")
	}
}

func TestCircuitBreaker_TransitionsToHalfOpenAfterCooldown(t *testing.T) {
	cb := NewCircuitBreaker(1, 30*time.Millisecond)
	cb.RecordFailure()
	time.Sleep(40 * time.Millisecond)
	if !cb.Allow() {
		t.Fatal("expected Allow() == true after cooldown (half-open)")
	}
	if cb.CurrentState() != StateHalfOpen {
		t.Fatalf("expected StateHalfOpen, got %v", cb.CurrentState())
	}
}

func TestCircuitBreaker_ClosesAfterSuccessInHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(1, 20*time.Millisecond)
	cb.RecordFailure()
	time.Sleep(25 * time.Millisecond)
	cb.Allow() // transition to half-open
	cb.RecordSuccess()
	if cb.CurrentState() != StateClosed {
		t.Fatalf("expected StateClosed after success in half-open, got %v", cb.CurrentState())
	}
}

func TestCircuitBreaker_ReopensOnFailureInHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(1, 20*time.Millisecond)
	cb.RecordFailure()
	time.Sleep(25 * time.Millisecond)
	cb.Allow() // transition to half-open
	cb.RecordFailure()
	if cb.CurrentState() != StateOpen {
		t.Fatalf("expected StateOpen after failure in half-open, got %v", cb.CurrentState())
	}
}

func TestCircuitBreaker_SuccessResetFailuresWhenClosed(t *testing.T) {
	cb := NewCircuitBreaker(3, 50*time.Millisecond)
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess() // should reset failure count
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.CurrentState() != StateClosed {
		t.Fatal("expected circuit to remain closed after success reset")
	}
}
