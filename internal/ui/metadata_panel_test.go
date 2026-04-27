package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/yourorg/grpcmon/internal/probe"
)

func makeMeta(addr, server, ver string) *probe.ServiceMeta {
	return &probe.ServiceMeta{
		Address:      addr,
		ServerName:   server,
		GRPCVersion:  ver,
		DiscoveredAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
}

func TestMetadataPanel_NoMeta(t *testing.T) {
	p := NewMetadataPanel()
	p.Update(nil)
	// Should not panic and should show placeholder text.
	text := p.view.GetText(false)
	if !strings.Contains(text, "No metadata") {
		t.Errorf("expected placeholder text, got: %q", text)
	}
}

func TestMetadataPanel_ShowsAddress(t *testing.T) {
	p := NewMetadataPanel()
	p.Update([]*probe.ServiceMeta{makeMeta("localhost:50051", "svc", "1.59")})
	text := p.view.GetText(false)
	if !strings.Contains(text, "localhost:50051") {
		t.Errorf("expected address in output, got: %q", text)
	}
}

func TestMetadataPanel_ShowsServerName(t *testing.T) {
	p := NewMetadataPanel()
	p.Update([]*probe.ServiceMeta{makeMeta("localhost:50051", "my-grpc-server", "")})
	text := p.view.GetText(false)
	if !strings.Contains(text, "my-grpc-server") {
		t.Errorf("expected server name in output, got: %q", text)
	}
}

func TestMetadataPanel_ShowsGRPCVersion(t *testing.T) {
	p := NewMetadataPanel()
	p.Update([]*probe.ServiceMeta{makeMeta("localhost:50051", "", "1.59.0")})
	text := p.view.GetText(false)
	if !strings.Contains(text, "1.59.0") {
		t.Errorf("expected grpc version in output, got: %q", text)
	}
}

func TestMetadataPanel_MultipleEntries(t *testing.T) {
	p := NewMetadataPanel()
	metas := []*probe.ServiceMeta{
		makeMeta("host1:50051", "svc-a", "1.58"),
		makeMeta("host2:50051", "svc-b", "1.59"),
	}
	p.Update(metas)
	text := p.view.GetText(false)
	if !strings.Contains(text, "host1:50051") || !strings.Contains(text, "host2:50051") {
		t.Errorf("expected both hosts in output, got: %q", text)
	}
}

func TestMetadataPanel_Title(t *testing.T) {
	p := NewMetadataPanel()
	title := p.view.GetTitle()
	if !strings.Contains(title, "Metadata") {
		t.Errorf("expected 'Metadata' in panel title, got: %q", title)
	}
}

func TestMetadataPanel_PrimitiveNotNil(t *testing.T) {
	p := NewMetadataPanel()
	if p.Primitive() == nil {
		t.Error("expected non-nil primitive")
	}
}
