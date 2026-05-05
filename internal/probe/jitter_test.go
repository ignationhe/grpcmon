package probe

import (
	"testing"
	"time"
)

func TestJitterPolicy_DefaultFactor(t *testing.T) {
	j := DefaultJitterPolicy()
	if j.Factor() != 0.20 {
		t.Fatalf("expected factor 0.20, got %v", j.Factor())
	}
}

func TestJitterPolicy_ClampNegativeFactor(t *testing.T) {
	j := NewJitterPolicy(-0.5)
	if j.Factor() != 0 {
		t.Fatalf("expected factor 0, got %v", j.Factor())
	}
}

func TestJitterPolicy_ClampOverOneFactor(t *testing.T) {
	j := NewJitterPolicy(1.5)
	if j.Factor() != 1 {
		t.Fatalf("expected factor 1, got %v", j.Factor())
	}
}

func TestJitterPolicy_ZeroFactor_ReturnsUnchanged(t *testing.T) {
	j := NewJitterPolicy(0)
	d := 500 * time.Millisecond
	if got := j.Apply(d); got != d {
		t.Fatalf("expected %v, got %v", d, got)
	}
}

func TestJitterPolicy_ZeroDuration_ReturnsZero(t *testing.T) {
	j := DefaultJitterPolicy()
	if got := j.Apply(0); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestJitterPolicy_Apply_WithinBounds(t *testing.T) {
	j := NewJitterPolicy(0.25)
	base := 1 * time.Second
	lo := time.Duration(float64(base) * 0.75)
	hi := time.Duration(float64(base) * 1.25)

	for i := 0; i < 200; i++ {
		got := j.Apply(base)
		if got < lo || got > hi {
			t.Fatalf("iteration %d: jittered value %v out of bounds [%v, %v]", i, got, lo, hi)
		}
	}
}

func TestJitterPolicy_Apply_NegativeDuration_ReturnsPositive(t *testing.T) {
	j := NewJitterPolicy(1.0)
	// With factor=1 the offset can be -1, which could produce <=0; ensure clamped.
	for i := 0; i < 50; i++ {
		got := j.Apply(100 * time.Millisecond)
		if got <= 0 {
			t.Fatalf("expected positive duration, got %v", got)
		}
	}
}

func TestJitterPolicy_Apply_Concurrent(t *testing.T) {
	j := DefaultJitterPolicy()
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			for k := 0; k < 50; k++ {
				j.Apply(time.Second)
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}
