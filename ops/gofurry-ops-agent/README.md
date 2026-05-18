# gofurry-ops-agent

`gofurry-ops-agent` is the lightweight host probe for GoFurry Ops. It collects local host and service health data, signs each report with HMAC, and pushes samples to a GoFurry Ops Center.

## Features

- System CPU, memory, load, uptime, disk and network counters
- Docker container status through the local Docker socket
- HTTP checks with status and optional body keyword matching
- Read-only PostgreSQL and Redis health checks
- TLS certificate expiry checks
- HMAC-signed reports to `/api/v1/agent/ingest`
- JSONL spool for short network outages

## Commands

```bash
go run ./cmd/agent --config ./configs/agent.example.yaml check-config
go run ./cmd/agent --config ./configs/agent.example.yaml once
go run ./cmd/agent --config ./configs/agent.example.yaml run
go run ./cmd/agent version
```

## Build

```bash
go test ./...
go build -o bin/gofurry-ops-agent ./cmd/agent
```

## Configuration

Copy `configs/agent.example.yaml` to `/etc/gofurry-ops-agent/config.yaml` and provide secrets through environment variables such as `GOFURRY_OPS_AGENT_TOKEN`.

PostgreSQL and Redis collectors are intentionally read-only. PostgreSQL runs `SELECT 1` and metadata queries. Redis runs `PING` and `INFO`.

## Systemd

```bash
sudo install -m 0755 gofurry-ops-agent /usr/local/bin/gofurry-ops-agent
sudo mkdir -p /etc/gofurry-ops-agent /var/lib/gofurry-ops-agent/spool
sudo cp configs/agent.example.yaml /etc/gofurry-ops-agent/config.yaml
sudo cp deploy/systemd/gofurry-ops-agent.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now gofurry-ops-agent
```

