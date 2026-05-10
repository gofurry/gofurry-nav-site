# systemd Deployment and Rollback

This guide covers running `gofurry-rag` as a long-lived Linux service.

## Before You Start

1. Prepare PostgreSQL with pgvector.
2. Prepare Ollama and make sure `rag.ollama_base_url` is reachable.
3. Store production configuration at `/etc/gofurry-rag/server.yaml`.
4. Replace all default credentials before going live.

## Install the Service

The simplest path is the built-in installer:

```bash
go run . --config /etc/gofurry-rag/server.yaml install
systemctl enable --now gofurry-rag
```

If you prefer to manage systemd manually, this unit file is a good starting point:

```ini
[Unit]
Description=gofurry-rag
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/gofurry-rag
ExecStart=/opt/gofurry-rag/gofurry-rag serve --config /etc/gofurry-rag/server.yaml
Restart=always
RestartSec=30
Environment=APP_SERVER_MODE=release

[Install]
WantedBy=multi-user.target
```

## Logs and Health

- `journalctl -u gofurry-rag -f` for structured logs.
- `curl http://127.0.0.1:8080/api/v1/health` for database, Ollama, and worker status.
- `curl http://127.0.0.1:8080/readyz` for readiness.

## Rollback

1. Stop the service:

```bash
sudo systemctl stop gofurry-rag
```

2. Restore the previous binary or configuration.
3. Prefer restoring from backup if you need to undo database changes.
4. Start the service again:

```bash
sudo systemctl start gofurry-rag
```

## Troubleshooting

- If `readyz` fails, check the PostgreSQL connection first.
- If queries time out, review `rag.query_timeout_seconds`, Ollama latency, and worker backlog.
- If the worker stays in `failed`, inspect `worker_recent_error` and `worker_recent_error_at`.
