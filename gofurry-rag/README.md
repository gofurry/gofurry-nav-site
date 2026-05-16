# gofurry-rag

![License](https://img.shields.io/github/license/gofurry/gofurry-rag)
![Go Version](https://img.shields.io/badge/go-1.26.0-00ADD8)

[中文说明](./README_zh.md)

`gofurry-rag` is a lightweight RAG service for gofurry content. It stores text knowledge in PostgreSQL with pgvector, creates embeddings through Ollama, and returns relevant source chunks for user questions.

## Features

- Cobra CLI and Viper `server.yaml` configuration, aligned with `gofurry-admin`
- Single Go binary with an embedded Vue + Tailwind dark console
- PostgreSQL + pgvector storage
- Ollama embedding client for `qwen3-embedding:0.6b`
- Async text ingest worker
- HttpOnly JWT Cookie login for admin APIs
- Public retrieval API that returns `sources` first, without LLM answer generation
- Overview console with auto refresh, document/chunk stats, worker status, database status, and Ollama status
- AI Q&A console page for Tencent Cloud inference debugging

## Quick Start

Edit `config/server.yaml`, then start the service:

```bash
go run . --config ./config/server.yaml serve
```

Open the embedded console:

```text
http://127.0.0.1:8080/admin
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
go vet ./...
go build .
go build ./cmd/server
```

## API

- `GET /api/v1/health`
- `GET /api/v1/admin/auth/state`
- `POST /api/v1/admin/auth/login`
- `POST /api/v1/admin/auth/logout`
- `GET /api/v1/admin/auth/me`
- `GET /api/v1/admin/overview`
- `POST /api/v1/admin/documents/text`
- `GET /api/v1/admin/documents`
- `GET /api/v1/admin/documents/:id/chunks`
- `DELETE /api/v1/admin/documents/:id`
- `POST /api/v1/chat/query`
- `POST /api/v1/chat/stream`
- `GET /livez`, `GET /readyz`, `GET /startupz`, `GET /healthz`

Admin routes require logging in through `/api/v1/admin/auth/login`. The service writes a JWT to an HttpOnly cookie, similar to `gofurry-admin`. `POST /api/v1/chat/query` and `POST /api/v1/chat/stream` remain public.

`GET /api/v1/health` includes database and Ollama connection information for the console overview.

## Configuration

Runtime configuration is read from `server.yaml`. By default, the service searches `/etc/gofurry-rag/server.yaml` and `./config/server.yaml`; `--config` can point to any file.

Environment overrides use the `APP_` prefix:

```bash
APP_SERVER_PORT=8081 APP_RAG_OLLAMA_BASE_URL=http://127.0.0.1:11434 go run . --config ./config/server.yaml serve
```

Tencent Cloud inference can be configured the same way without committing secrets:

```bash
APP_RAG_TENCENT_BASE_URL=https://tokenhub.tencentmaas.com/v1 \
APP_RAG_TENCENT_MODEL=deepseek-v4-flash \
APP_RAG_TENCENT_API_KEY=your-secret-key \
go run . --config ./config/server.yaml serve
```

Do not commit real database passwords, console passcodes, or JWT secrets.

## Console Debugging

- The embedded console includes an `AI 问答` page for live RAG debugging.
- The page streams status updates, sources, answer tokens, and final usage from `POST /api/v1/chat/stream`.
- The existing `文档检索` page still exposes the plain retrieval API and source debug details.

## Documentation

- [Usage](./docs/usage.md)
- [Smoke Test](./docs/smoke-test.md)
- [Deployment](./docs/deployment.md)
- [MVP Design](./docs/rag-design.md)

## License

MIT, following the parent gofurry repository.
