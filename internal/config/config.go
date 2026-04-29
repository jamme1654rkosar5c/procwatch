package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ProcessConfig defines a single watched process and its thresholds.
type ProcessConfig struct {
	Name        string  `yaml:"name"`
	PIDFile     string  `yaml:"pid_file,omitempty"`
	MatchExpr   string  `yaml:"match_expr,omitempty"`
	MaxCPU      float64 `yaml:"max_cpu_percent,omitempty"`
	MaxMemoryMB uint64  `yaml:"max_memory_mb,omitempty"`
}

// WebhookConfig holds webhook delivery settings.
type WebhookConfig struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers,omitempty"`
	TimeoutSeconds int        `yaml:"timeout_seconds,omitempty"`
}

// Config is the top-level configuration structure.
type Config struct {
	PollInterval string          `yaml:"poll_interval"`
	Webhook      WebhookConfig   `yaml:"webhook"`
	Processes    []ProcessConfig `yaml:"processes"`

	pollDuration time.Duration
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// PollDuration returns the parsed poll interval duration.
func (c *Config) PollDuration() time.Duration {
	return c.pollDuration
}

func (c *Config) validate() error {
	if c.Webhook.URL == "" {
		return fmt.Errorf("webhook.url is required")
	}
	if len(c.Processes) == 0 {
		return fmt.Errorf("at least one process must be defined")
	}

	interval := c.PollInterval
	if interval == "" {
		interval = "10s"
	}
	d, err := time.ParseDuration(interval)
	if err != nil {
		return fmt.Errorf("poll_interval %q is not a valid duration: %w", interval, err)
	}
	c.pollDuration = d

	for i, p := range c.Processes {
		if p.Name == "" {
			return fmt.Errorf("processes[%d]: name is required", i)
		}
		if p.PIDFile == "" && p.MatchExpr == "" {
			return fmt.Errorf("processes[%d] (%s): pid_file or match_expr is required", i, p.Name)
		}
	}
	return nil
}
