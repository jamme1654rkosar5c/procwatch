# procwatch

Lightweight process monitor that alerts via webhook when watched processes crash or exceed resource thresholds.

## Installation

```bash
go install github.com/yourusername/procwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/procwatch.git && cd procwatch && go build -o procwatch .
```

## Usage

Create a `config.yaml` file:

```yaml
webhook: "https://hooks.slack.com/services/your/webhook/url"
interval: 10s
processes:
  - name: "nginx"
    max_cpu: 80.0
    max_mem_mb: 512
  - name: "myapp"
    max_cpu: 50.0
    max_mem_mb: 256
```

Then run:

```bash
procwatch --config config.yaml
```

procwatch will poll the specified processes at the given interval and POST a JSON alert to your webhook if a process is not found or exceeds the defined CPU/memory thresholds.

### Alert Payload Example

```json
{
  "process": "nginx",
  "event": "threshold_exceeded",
  "cpu_percent": 91.3,
  "mem_mb": 489,
  "timestamp": "2024-05-01T12:34:56Z"
}
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `config.yaml` | Path to configuration file |
| `--dry-run` | `false` | Log alerts without sending webhook requests |
| `--log-level` | `info` | Logging verbosity (`debug`, `info`, `warn`) |

## License

MIT © 2024 yourusername