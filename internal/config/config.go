package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Target represents a single gRPC service to monitor.
type Target struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

// RetryConfig holds retry behaviour for probes.
type RetryConfig struct {
	MaxAttempts int           `yaml:"max_attempts"`
	Delay       time.Duration `yaml:"delay"`
}

// Config holds the full application configuration.
type Config struct {
	Targets      []Target     `yaml:"targets"`
	PollInterval time.Duration `yaml:"poll_interval"`
	Timeout      time.Duration `yaml:"timeout"`
	HistorySize  int           `yaml:"history_size"`
	Retry        RetryConfig   `yaml:"retry"`
}

// Load reads and parses a YAML config file at path.
// Missing optional fields are filled with defaults.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if len(cfg.Targets) == 0 {
		return nil, errors.New("config: at least one target is required")
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 5 * time.Second
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 2 * time.Second
	}
	if cfg.HistorySize <= 0 {
		cfg.HistorySize = 60
	}
	if cfg.Retry.MaxAttempts <= 0 {
		cfg.Retry.MaxAttempts = 1
	}
	if cfg.Retry.Delay <= 0 {
		cfg.Retry.Delay = 100 * time.Millisecond
	}
	for i, t := range cfg.Targets {
		if t.Name == "" {
			cfg.Targets[i].Name = t.Address
		}
	}
}
