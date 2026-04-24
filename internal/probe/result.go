package probe

import "time"

// Result holds the outcome of a single health check probe.
type Result struct {
	// Target is the gRPC address that was probed (host:port).
	Target string

	// Serving is true when the remote service reported SERVING status.
	Serving bool

	// Latency is the round-trip time of the health check RPC.
	Latency time.Duration

	// Timestamp records when the probe completed.
	Timestamp time.Time

	// Err is non-nil when the probe itself failed (network error, timeout, etc.).
	Err error
}

// IsHealthy returns true only when the probe succeeded and the service is serving.
func (r Result) IsHealthy() bool {
	return r.Err == nil && r.Serving
}

// String returns a compact human-readable summary of the result.
func (r Result) String() string {
	if r.Err != nil {
		return r.Target + " ERROR: " + r.Err.Error()
	}
	if r.Serving {
		return r.Target + " SERVING (" + r.Latency.Round(time.Millisecond).String() + ")"
	}
	return r.Target + " NOT_SERVING"
}
