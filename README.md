# grpcmon

Lightweight terminal dashboard for monitoring live gRPC service health and latency metrics.

---

## Installation

```bash
go install github.com/yourname/grpcmon@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/grpcmon.git && cd grpcmon && go build -o grpcmon .
```

---

## Usage

Point `grpcmon` at one or more gRPC endpoints and watch real-time health and latency stats in your terminal.

```bash
grpcmon --target localhost:50051
```

Monitor multiple services at once:

```bash
grpcmon --target localhost:50051 --target localhost:50052 --interval 2s
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--target` | *(required)* | gRPC endpoint to monitor (repeatable) |
| `--interval` | `1s` | Polling interval |
| `--timeout` | `5s` | Per-request timeout |
| `--tls` | `false` | Enable TLS for connections |

### Dashboard

```
┌─ grpcmon ──────────────────────────────────────┐
│ SERVICE              STATUS   P50    P95   P99  │
│ localhost:50051      ● UP     4ms    12ms  28ms │
│ localhost:50052      ● UP     6ms    18ms  45ms │
│ localhost:50053      ○ DOWN   —      —     —    │
└────────────────────────────────────────────────┘
```

---

## Requirements

- Go 1.21+
- gRPC services with [health checking protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md) enabled

---

## Contributing

Pull requests are welcome. Please open an issue first to discuss any major changes.

---

## License

[MIT](LICENSE)