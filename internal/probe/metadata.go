package probe

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ServiceMeta holds discovered metadata about a gRPC target.
type ServiceMeta struct {
	Address     string
	ServerName  string
	GRPCVersion string
	DiscoveredAt time.Time
}

// MetadataCollector fetches and caches server metadata via gRPC headers.
type MetadataCollector struct {
	mu    sync.RWMutex
	cache map[string]*ServiceMeta
}

// NewMetadataCollector returns a new MetadataCollector.
func NewMetadataCollector() *MetadataCollector {
	return &MetadataCollector{
		cache: make(map[string]*ServiceMeta),
	}
}

// Collect performs a lightweight RPC to capture response headers from the
// target and stores the resulting ServiceMeta in the cache.
func (m *MetadataCollector) Collect(ctx context.Context, address string, conn *grpc.ClientConn) (*ServiceMeta, error) {
	var header metadata.MD
	// Issue a trivial RPC (grpc.EmptyCallOption) just to receive headers.
	_ = conn.Invoke(ctx, "/grpc.health.v1.Health/Check", nil, nil, grpc.Header(&header))

	serverName := firstMD(header, "server")
	grpcVersion := firstMD(header, "grpc-version")

	meta := &ServiceMeta{
		Address:      address,
		ServerName:   serverName,
		GRPCVersion:  grpcVersion,
		DiscoveredAt: time.Now(),
	}

	m.mu.Lock()
	m.cache[address] = meta
	m.mu.Unlock()

	return meta, nil
}

// Get returns cached metadata for the given address, or nil if not present.
func (m *MetadataCollector) Get(address string) *ServiceMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cache[address]
}

// All returns a snapshot of all cached metadata.
func (m *MetadataCollector) All() []*ServiceMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*ServiceMeta, 0, len(m.cache))
	for _, v := range m.cache {
		out = append(out, v)
	}
	return out
}

func firstMD(md metadata.MD, key string) string {
	vals := md.Get(key)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}
