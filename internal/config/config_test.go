package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/yourorg/grpcmon/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "grpcmon-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	path := writeTemp(t, `
targets:
  - name: payments
    address: localhost:50051
    interval: 10s
    timeout: 3s
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(cfg.Targets))
	}
	if cfg.Targets[0].Name != "payments" {
		t.Errorf("expected name 'payments', got %q", cfg.Targets[0].Name)
	}
	if cfg.Targets[0].Interval != 10*time.Second {
		t.Errorf("expected interval 10s, got %v", cfg.Targets[0].Interval)
	}
}

func TestLoad_Defaults(t *testing.T) {
	path := writeTemp(t, `
targets:
  - address: localhost:9090
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tgt := cfg.Targets[0]
	if tgt.Interval != config.DefaultInterval {
		t.Errorf("expected default interval %v, got %v", config.DefaultInterval, tgt.Interval)
	}
	if tgt.Timeout != config.DefaultTimeout {
		t.Errorf("expected default timeout %v, got %v", config.DefaultTimeout, tgt.Timeout)
	}
	if tgt.Name != "localhost:9090" {
		t.Errorf("expected name to fall back to address, got %q", tgt.Name)
	}
}

func TestLoad_NoTargets(t *testing.T) {
	path := writeTemp(t, `targets: []`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for empty targets, got nil")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
