package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Target represents a single gRPC service endpoint to monitor.
type Target struct {
	Name     string        `yaml:"name"`
	Address  string        `yaml:"address"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
	TLS      bool          `yaml:"tls"`
}

// Config holds the full grpcmon configuration.
type Config struct {
	Targets []Target `yaml:"targets"`
}

// Defaults applied when fields are zero-valued.
const (
	DefaultInterval = 5 * time.Second
	DefaultTimeout  = 2 * time.Second
)

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	cfg.applyDefaults()
	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Targets) == 0 {
		return errors.New("config: no targets defined")
	}
	for i, t := range c.Targets {
		if t.Address == "" {
			return errors.New("config: target at index " + string(rune('0'+i)) + " missing address")
		}
	}
	return nil
}

func (c *Config) applyDefaults() {
	for i := range c.Targets {
		if c.Targets[i].Interval == 0 {
			c.Targets[i].Interval = DefaultInterval
		}
		if c.Targets[i].Timeout == 0 {
			c.Targets[i].Timeout = DefaultTimeout
		}
		if c.Targets[i].Name == "" {
			c.Targets[i].Name = c.Targets[i].Address
		}
	}
}
