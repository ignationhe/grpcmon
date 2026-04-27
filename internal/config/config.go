package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Target represents a single gRPC endpoint to monitor.
type Target struct {
	Name    string        `yaml:"name"`
	Address string        `yaml:"address"`
	Timeout time.Duration `yaml:"timeout"`
}

// Config holds the full grpcmon configuration.
type Config struct {
	PollInterval  time.Duration     `yaml:"poll_interval"`
	Targets       []Target          `yaml:"targets"`
	MaxHistory    int               `yaml:"max_history"`
	RetryAttempts int               `yaml:"retry_attempts"`
	Timeouts      map[string]string `yaml:"timeouts"`
}

// Load reads and parses the YAML config file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	if len(cfg.Targets) == 0 {
		return nil, errors.New("config: at least one target is required")
	}

	for i, t := range cfg.Targets {
		if t.Address == "" {
			return nil, fmt.Errorf("config: target[%d] missing address", i)
		}
		if t.Name == "" {
			cfg.Targets[i].Name = t.Address
		}
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 10 * time.Second
	}
	if cfg.MaxHistory <= 0 {
		cfg.MaxHistory = 100
	}
	if cfg.RetryAttempts <= 0 {
		cfg.RetryAttempts = 2
	}
	for i, t := range cfg.Targets {
		if t.Timeout <= 0 {
			cfg.Targets[i].Timeout = 5 * time.Second
		}
	}
}
