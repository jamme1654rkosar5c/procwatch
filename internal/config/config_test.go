package config

import (
	"os"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "procwatch-*.yaml")
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
poll_interval: 5s
webhook:
  url: https://hooks.example.com/alert
  timeout_seconds: 10
processes:
  - name: nginx
    match_expr: nginx
    max_cpu_percent: 80
    max_memory_mb: 512
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Webhook.URL != "https://hooks.example.com/alert" {
		t.Errorf("unexpected webhook URL: %s", cfg.Webhook.URL)
	}
	if cfg.PollDuration().Seconds() != 5 {
		t.Errorf("expected 5s poll duration, got %v", cfg.PollDuration())
	}
	if len(cfg.Processes) != 1 || cfg.Processes[0].Name != "nginx" {
		t.Errorf("unexpected processes: %+v", cfg.Processes)
	}
}

func TestLoad_DefaultPollInterval(t *testing.T) {
	path := writeTemp(t, `
webhook:
  url: https://hooks.example.com/alert
processes:
  - name: myapp
    pid_file: /run/myapp.pid
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollDuration().Seconds() != 10 {
		t.Errorf("expected default 10s, got %v", cfg.PollDuration())
	}
}

func TestLoad_MissingWebhookURL(t *testing.T) {
	path := writeTemp(t, `
processes:
  - name: myapp
    pid_file: /run/myapp.pid
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for missing webhook URL")
	}
}

func TestLoad_ProcessMissingLocator(t *testing.T) {
	path := writeTemp(t, `
webhook:
  url: https://hooks.example.com/alert
processes:
  - name: orphan
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error when neither pid_file nor match_expr is set")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	if _, err := Load("/nonexistent/path.yaml"); err == nil {
		t.Fatal("expected error for missing file")
	}
}
