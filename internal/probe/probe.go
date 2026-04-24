package probe

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Result holds the outcome of a single health probe.
type Result struct {
	Target    string
	Status    grpc_health_v1.HealthCheckResponse_ServingStatus
	LatencyMs float64
	Err       error
	Timestamp time.Time
}

// Prober dials a gRPC endpoint and checks its health.
type Prober struct {
	target  string
	conn    *grpc.ClientConn
	client  grpc_health_v1.HealthClient
	timeout time.Duration
}

// New creates a Prober for the given target address (host:port).
func New(target string, timeout time.Duration) (*Prober, error) {
	conn, err := grpc.NewClient(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &Prober{
		target:  target,
		conn:    conn,
		client:  grpc_health_v1.NewHealthClient(conn),
		timeout: timeout,
	}, nil
}

// Check performs a single health check and returns a Result.
func (p *Prober) Check(ctx context.Context, service string) Result {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	start := time.Now()
	resp, err := p.client.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: service,
	})
	latency := float64(time.Since(start).Microseconds()) / 1000.0

	r := Result{
		Target:    p.target,
		LatencyMs: latency,
		Err:       err,
		Timestamp: time.Now(),
	}
	if err == nil {
		r.Status = resp.GetStatus()
	}
	return r
}

// ConnState returns the current connectivity state of the underlying connection.
func (p *Prober) ConnState() connectivity.State {
	return p.conn.GetState()
}

// Close shuts down the underlying gRPC connection.
func (p *Prober) Close() error {
	return p.conn.Close()
}
