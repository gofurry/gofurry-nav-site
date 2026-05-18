# GoFurry Ops Center

GoFurry Ops Center is the lightweight receiver and dashboard for `gofurry-ops-agent`. It stores agent samples, service status, alert state, peer summaries, sync runs, and deployment events in Postgres, then serves the embedded `/admin` console from the same Go binary.

## Quick Start

```powershell
cd ops\gofurry-ops-center
Copy-Item configs\center.example.yaml configs\center.local.yaml
go run .\cmd\center check-config --config .\configs\center.local.yaml
go run .\cmd\center serve --config .\configs\center.local.yaml
```

The admin console is available at `http://127.0.0.1:8080/admin`.

## Build

```powershell
cd ops\gofurry-ops-center\web
npm ci
npm run build

cd ..
go test ./...
go vet ./...
go build ./cmd/center
```

The frontend build writes to `internal/web/dist` and is embedded by `internal/web/embed.go`.

## API Surface

- `POST /api/v1/agent/ingest`: agent report ingestion with bearer token, timestamp, node id, and HMAC-SHA256 signature.
- `GET /api/v1/peer/summary`: current center summary for a peer center.
- `POST /api/v1/peer/heartbeat`: peer summary heartbeat.
- `POST /api/v1/events/sync`: sync run event.
- `POST /api/v1/events/deploy`: deploy event.
- `GET /api/v1/dashboard/*`: authenticated dashboard APIs.

## Data Safety

Use a dedicated database or schema for local testing. The center migration only creates `gofurry_ops_*` tables in the connected search path, and agent collectors use read-only checks for Postgres and Redis. Do not place real DSNs, passwords, or tokens in committed config files.

## Deployment Notes

- `configs/center.example.yaml` is a template; copy it to a local config path and replace all `change-me-*` values.
- `deploy/systemd/gofurry-ops-center.service` expects the binary under `/opt/gofurry-ops-center`.
- `deploy/docker-compose.prod.yml` expects a non-committed `deploy/config.yaml`.
- `deploy/nginx/gofurry-ops-center.conf` is a minimal reverse proxy sample.
