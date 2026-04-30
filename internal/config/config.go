// Package config loads and validates procwatch configuration from a YAML file.
package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Default values.
const defaultPollInterval = 15 * time.Second

// Process describes a single watched process entry.
type Process struct {
	Name        string  `yaml:"name"`
	PIDFile     string  `yaml:"pid_file"`
	CPULimit    float64 `yaml:"cpu_limit_percent"`  // 0 = disabled
	MemLimitMB  float64 `yaml:"mem_limit_mb"`       // 0 = disabled
}

// Config is the top-level configuration structure.
type Config struct {
	WebhookURL   string        `yaml:"webhook_url"`
	PollInterval time.Duration `yaml:"poll_interval"`
	Processes    []Process     `yaml:"processes"`
}

// Load reads and validates a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.PollInterval <= 0 {
		cfg.PollInterval = defaultPollInterval
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.WebhookURL == "" {
		return errors.New("webhook_url is required")
	}

	for i, p := range cfg.Processes {
		if p.Name == "" && p.PIDFile == "" {
			return fmt.Errorf("process[%d]: must specify name or pid_file", i)
		}
		if p.CPULimit < 0 {
			return fmt.Errorf("process[%d]: cpu_limit_percent must be >= 0", i)
		}
		if p.MemLimitMB < 0 {
			return fmt.Errorf("process[%d]: mem_limit_mb must be >= 0", i)
		}
	}

	return nil
}
