package probe

import (
	"sync"
	"testing"
	"time"
)

func TestThrottle_AcquireFirstTime(t *testing.T) {
	p := NewThrottlePolicy(2, 100*time.Millisecond)
	if !p.Acquire("host:50051") {
		t.Fatal("expected first acquire to succeed")
	}
	p.Release()
}

func TestThrottle_RejectWithinMinInterval(t *testing.T) {
	p := NewThrottlePolicy(2, 500*time.Millisecond)
	if !p.Acquire("host:50051") {
		t.Fatal("first acquire should succeed")
	}
	p.Release()

	if p.Acquire("host:50051") {
		t.Fatal("second acquire within minInterval should be rejected")
	}
}

func TestThrottle_AllowAfterMinInterval(t *testing.T) {
	p := NewThrottlePolicy(2, 10*time.Millisecond)
	if !p.Acquire("host:50051") {
		t.Fatal("first acquire should succeed")
	}
	p.Release()

	time.Sleep(20 * time.Millisecond)

	if !p.Acquire("host:50051") {
		t.Fatal("acquire after minInterval should succeed")
	}
	p.Release()
}

func TestThrottle_DifferentTargetsIndependent(t *testing.T) {
	p := NewThrottlePolicy(4, 500*time.Millisecond)
	if !p.Acquire("a:50051") {
		t.Fatal("acquire a should succeed")
	}
	p.Release()
	if !p.Acquire("b:50051") {
		t.Fatal("acquire b should succeed independently")
	}
	p.Release()
}

func TestThrottle_MaxInflightBlocks(t *testing.T) {
	p := NewThrottlePolicy(2, 0)

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		if !p.Acquire("host:50051") {
			t.Fatal("should acquire up to maxInflight")
		}
	}

	if p.Inflight() != 2 {
		t.Fatalf("expected 2 inflight, got %d", p.Inflight())
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		p.Release()
	}()

	p.Acquire("other:50051")
	wg.Wait()
	p.Release()
}

func TestThrottle_Reset_AllowsImmediateReacquire(t *testing.T) {
	p := NewThrottlePolicy(2, 500*time.Millisecond)
	if !p.Acquire("host:50051") {
		t.Fatal("first acquire should succeed")
	}
	p.Release()

	p.Reset()

	if !p.Acquire("host:50051") {
		t.Fatal("acquire after reset should succeed")
	}
	p.Release()
}
