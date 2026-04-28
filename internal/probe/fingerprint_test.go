package probe

import (
	"errors"
	"testing"
)

func baseResult(target, status string) Result {
	return Result{
		Target:   target,
		Status:   status,
		Metadata: map[string]string{},
	}
}

func TestCompute_Deterministic(t *testing.T) {
	r := baseResult("localhost:50051", "SERVING")
	a := Compute(r)
	b := Compute(r)
	if a != b {
		t.Fatalf("expected deterministic fingerprint, got %q and %q", a, b)
	}
}

func TestCompute_DiffersOnStatus(t *testing.T) {
	r1 := baseResult("localhost:50051", "SERVING")
	r2 := baseResult("localhost:50051", "NOT_SERVING")
	if Compute(r1) == Compute(r2) {
		t.Fatal("expected different fingerprints for different statuses")
	}
}

func TestCompute_DiffersOnError(t *testing.T) {
	r1 := baseResult("localhost:50051", "UNKNOWN")
	r2 := baseResult("localhost:50051", "UNKNOWN")
	r2.Err = errors.New("connection refused")
	if Compute(r1) == Compute(r2) {
		t.Fatal("expected different fingerprints when error differs")
	}
}

func TestCompute_MetadataOrderIndependent(t *testing.T) {
	r1 := Result{Target: "svc", Status: "SERVING", Metadata: map[string]string{"a": "1", "b": "2"}}
	r2 := Result{Target: "svc", Status: "SERVING", Metadata: map[string]string{"b": "2", "a": "1"}}
	if Compute(r1) != Compute(r2) {
		t.Fatal("expected same fingerprint regardless of metadata iteration order")
	}
}

func TestFingerprint_ChangedFirstCall(t *testing.T) {
	f := NewFingerprint()
	r := baseResult("svc", "SERVING")
	if !f.Changed(r) {
		t.Fatal("expected Changed=true on first call")
	}
}

func TestFingerprint_ChangedSameResult(t *testing.T) {
	f := NewFingerprint()
	r := baseResult("svc", "SERVING")
	f.Changed(r) // seed
	if f.Changed(r) {
		t.Fatal("expected Changed=false for identical successive result")
	}
}

func TestFingerprint_ChangedAfterStatusChange(t *testing.T) {
	f := NewFingerprint()
	f.Changed(baseResult("svc", "SERVING"))
	if !f.Changed(baseResult("svc", "NOT_SERVING")) {
		t.Fatal("expected Changed=true after status change")
	}
}

func TestFingerprint_GetAndSet(t *testing.T) {
	f := NewFingerprint()
	f.Set("svc", "abc123")
	if got := f.Get("svc"); got != "abc123" {
		t.Fatalf("expected %q, got %q", "abc123", got)
	}
}

func TestFingerprint_GetMissing(t *testing.T) {
	f := NewFingerprint()
	if got := f.Get("unknown"); got != "" {
		t.Fatalf("expected empty string for missing target, got %q", got)
	}
}

func TestFingerprint_Reset(t *testing.T) {
	f := NewFingerprint()
	r := baseResult("svc", "SERVING")
	f.Changed(r)
	f.Reset()
	if !f.Changed(r) {
		t.Fatal("expected Changed=true after Reset")
	}
}
