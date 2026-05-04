# gofurry-rag

![License](https://img.shields.io/github/license/GoFurry/gofurry-rag)
![Go Version](https://img.shields.io/badge/go-1.26.0-00ADD8)

[中文说明](./README_zh.md)

`gofurry-rag` is a lightweight RAG service for GoFurry content. It stores text knowledge in PostgreSQL with pgvector, creates embeddings through Ollama, and returns relevant source chunks for user questions.

## Features

- Single Go binary with an embedded Vue admin console
- PostgreSQL + pgvector storage
- Ollama embedding client for `qwen3-embedding:0.6b`
- Async text ingest worker
- Admin token protection for document APIs and console workflows
- Minimal retrieval API that returns `sources` first, without LLM answer generation

## Quick Start

```bash
cp .env.example .env
go run ./cmd/server
```

Start the admin console in development:

```bash
cd web
npm install
npm run dev
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

Admin routes require logging in through `/api/v1/admin/auth/login`. The service writes a JWT to an HttpOnly cookie, similar to `gofurry-admin`.

## Configuration

See [.env.example](./.env.example). Do not commit real `.env` files or database passwords.

## Documentation

- [Usage](./docs/usage.md)
- [Smoke Test](./docs/smoke-test.md)
- [MVP Design](./docs/rag-design.md)

## License

MIT, following the parent GoFurry repository.
