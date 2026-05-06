package probe

import (
	"testing"
	"time"
)

func TestQuotaPolicy_AllowFirstProbe(t *testing.T) {
	q := NewQuotaPolicy(time.Minute, 3)
	if !q.Allow("svc:443") {
		t.Fatal("expected first probe to be allowed")
	}
}

func TestQuotaPolicy_AllowUpToMax(t *testing.T) {
	q := NewQuotaPolicy(time.Minute, 3)
	for i := 0; i < 3; i++ {
		if !q.Allow("svc:443") {
			t.Fatalf("expected probe %d to be allowed", i+1)
		}
	}
}

func TestQuotaPolicy_BlocksWhenExhausted(t *testing.T) {
	q := NewQuotaPolicy(time.Minute, 2)
	q.Allow("svc:443")
	q.Allow("svc:443")
	if q.Allow("svc:443") {
		t.Fatal("expected probe to be blocked after quota exhausted")
	}
}

func TestQuotaPolicy_AllowsAfterWindowExpiry(t *testing.T) {
	q := NewQuotaPolicy(50*time.Millisecond, 1)
	if !q.Allow("svc:443") {
		t.Fatal("expected first probe allowed")
	}
	if q.Allow("svc:443") {
		t.Fatal("expected second probe blocked")
	}
	time.Sleep(60 * time.Millisecond)
	if !q.Allow("svc:443") {
		t.Fatal("expected probe allowed after window expiry")
	}
}

func TestQuotaPolicy_DifferentTargetsIndependent(t *testing.T) {
	q := NewQuotaPolicy(time.Minute, 1)
	q.Allow("a:443")
	if !q.Allow("b:443") {
		t.Fatal("expected different target to have its own quota")
	}
}

func TestQuotaPolicy_Remaining(t *testing.T) {
	q := NewQuotaPolicy(time.Minute, 3)
	if r := q.Remaining("svc:443"); r != 3 {
		t.Fatalf("expected remaining=3, got %d", r)
	}
	q.Allow("svc:443")
	if r := q.Remaining("svc:443"); r != 2 {
		t.Fatalf("expected remaining=2, got %d", r)
	}
}

func TestQuotaPolicy_Reset_ClearsHistory(t *testing.T) {
	q := NewQuotaPolicy(time.Minute, 1)
	q.Allow("svc:443")
	if q.Allow("svc:443") {
		t.Fatal("expected blocked before reset")
	}
	q.Reset("svc:443")
	if !q.Allow("svc:443") {
		t.Fatal("expected allowed after reset")
	}
}

func TestQuotaPolicy_DefaultsOnZeroArgs(t *testing.T) {
	q := NewQuotaPolicy(0, 0)
	if q.window != time.Minute {
		t.Fatalf("expected default window=1m, got %v", q.window)
	}
	if q.maxProbes != 1 {
		t.Fatalf("expected default maxProbes=1, got %d", q.maxProbes)
	}
}
