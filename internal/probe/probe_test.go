package probe_test

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/user/grpcmon/internal/probe"
)

func startHealthServer(t *testing.T, serving bool) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	srv := grpc.NewServer()
	hsrv := health.NewServer()
	status := grpc_health_v1.HealthCheckResponse_SERVING
	if !serving {
		status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	}
	hsrv.SetServingStatus("", status)
	grpc_health_v1.RegisterHealthServer(srv, hsrv)

	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	return lis.Addr().String()
}

func TestCheck_Serving(t *testing.T) {
	addr := startHealthServer(t, true)
	p, err := probe.New(addr, 2*time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()

	r := p.Check(context.Background(), "")
	if r.Err != nil {
		t.Fatalf("unexpected error: %v", r.Err)
	}
	if r.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Errorf("expected SERVING, got %v", r.Status)
	}
	if r.LatencyMs < 0 {
		t.Errorf("negative latency: %v", r.LatencyMs)
	}
}

func TestCheck_NotServing(t *testing.T) {
	addr := startHealthServer(t, false)
	p, err := probe.New(addr, 2*time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()

	r := p.Check(context.Background(), "")
	if r.Err != nil {
		t.Fatalf("unexpected error: %v", r.Err)
	}
	if r.Status != grpc_health_v1.HealthCheckResponse_NOT_SERVING {
		t.Errorf("expected NOT_SERVING, got %v", r.Status)
	}
}

func TestCheck_Timeout(t *testing.T) {
	p, err := probe.New("127.0.0.1:1", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()

	r := p.Check(context.Background(), "")
	if r.Err == nil {
		t.Error("expected error for unreachable target")
	}
}
