# gofurry-rag

![License](https://img.shields.io/github/license/GoFurry/gofurry-rag)
![Go Version](https://img.shields.io/badge/go-1.26.0-00ADD8)

[中文说明](./README_zh.md)

`gofurry-rag` is a lightweight RAG service for GoFurry content. It stores text knowledge in PostgreSQL with pgvector, creates embeddings through Ollama, and returns relevant source chunks for user questions.

## Features

- Cobra CLI and Viper `server.yaml` configuration, aligned with `gofurry-admin`
- Single Go binary with an embedded Vue + Tailwind admin console
- PostgreSQL + pgvector storage
- Ollama embedding client for `qwen3-embedding:0.6b`
- Async text ingest worker
- HttpOnly JWT Cookie login for admin APIs
- Public retrieval API that returns `sources` first, without LLM answer generation

## Quick Start

Edit `config/server.yaml`, then start the service:

```bash
go run . --config ./config/server.yaml serve
```

Common CLI commands:

```bash
go run . --config ./config/server.yaml version
go run . --config ./config/server.yaml reset-password --password change-me
go run . --config ./config/server.yaml install
go run . --config ./config/server.yaml uninstall
```

Build the embedded console and verify Go code:

```bash
cd web
npm ci
npm run build
cd ..
go test ./...
```

## API

- `GET /api/v1/health`
- `GET /api/v1/admin/auth/state`
- `POST /api/v1/admin/auth/login`
- `POST /api/v1/admin/auth/logout`
- `GET /api/v1/admin/overview`
- `POST /api/v1/admin/documents/text`
- `GET /api/v1/admin/documents`
- `GET /api/v1/admin/documents/:id/chunks`
- `DELETE /api/v1/admin/documents/:id`
- `POST /api/v1/chat/query`
- `GET /livez`, `GET /readyz`, `GET /startupz`, `GET /healthz`

Admin routes require logging in through `/api/v1/admin/auth/login`. The service writes a JWT to an HttpOnly cookie, similar to `gofurry-admin`.

## Configuration

Runtime configuration is read from `server.yaml`. By default, the service searches `/etc/gofurry-rag/server.yaml` and `./config/server.yaml`; `--config` can point to any file. Environment overrides use the `APP_` prefix, for example `APP_SERVER_PORT=8081` or `APP_RAG_OLLAMA_BASE_URL=http://127.0.0.1:11434`.

## Documentation

- [Usage](./docs/usage.md)
- [Smoke Test](./docs/smoke-test.md)
- [MVP Design](./docs/rag-design.md)

## License

MIT, following the parent GoFurry repository.
