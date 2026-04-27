package probe

import (
	"testing"
	"time"

	"google.golang.org/grpc/metadata"
)

func TestFirstMD_Present(t *testing.T) {
	md := metadata.Pairs("server", "envoy")
	if got := firstMD(md, "server"); got != "envoy" {
		t.Fatalf("expected 'envoy', got %q", got)
	}
}

func TestFirstMD_Missing(t *testing.T) {
	md := metadata.MD{}
	if got := firstMD(md, "server"); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestMetadataCollector_GetMissing(t *testing.T) {
	c := NewMetadataCollector()
	if got := c.Get("localhost:50051"); got != nil {
		t.Fatalf("expected nil for unknown address, got %+v", got)
	}
}

func TestMetadataCollector_StoreAndGet(t *testing.T) {
	c := NewMetadataCollector()
	meta := &ServiceMeta{
		Address:      "localhost:50051",
		ServerName:   "my-service",
		GRPCVersion:  "1.0",
		DiscoveredAt: time.Now(),
	}
	c.mu.Lock()
	c.cache[meta.Address] = meta
	c.mu.Unlock()

	got := c.Get("localhost:50051")
	if got == nil {
		t.Fatal("expected metadata, got nil")
	}
	if got.ServerName != "my-service" {
		t.Fatalf("expected 'my-service', got %q", got.ServerName)
	}
}

func TestMetadataCollector_All(t *testing.T) {
	c := NewMetadataCollector()
	addresses := []string{"host1:50051", "host2:50051", "host3:50051"}
	for _, addr := range addresses {
		c.mu.Lock()
		c.cache[addr] = &ServiceMeta{Address: addr, DiscoveredAt: time.Now()}
		c.mu.Unlock()
	}

	all := c.All()
	if len(all) != len(addresses) {
		t.Fatalf("expected %d entries, got %d", len(addresses), len(all))
	}
}

func TestMetadataCollector_All_Empty(t *testing.T) {
	c := NewMetadataCollector()
	if got := c.All(); len(got) != 0 {
		t.Fatalf("expected empty slice, got %d entries", len(got))
	}
}

func TestMetadataCollector_OverwritesExisting(t *testing.T) {
	c := NewMetadataCollector()
	c.mu.Lock()
	c.cache["host:50051"] = &ServiceMeta{Address: "host:50051", ServerName: "old"}
	c.mu.Unlock()

	c.mu.Lock()
	c.cache["host:50051"] = &ServiceMeta{Address: "host:50051", ServerName: "new"}
	c.mu.Unlock()

	if got := c.Get("host:50051"); got.ServerName != "new" {
		t.Fatalf("expected 'new', got %q", got.ServerName)
	}
}
