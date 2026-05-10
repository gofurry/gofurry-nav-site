# Usage

## Requirements

- Go 1.26
- Node.js and npm
- PostgreSQL with pgvector
- Ollama with `qwen3-embedding:0.6b`

## Configure

Edit `config/server.yaml`. The most important fields are:

```yaml
database:
  postgres:
    db_name: "gfr"
    db_username: "postgres"
    db_password: "your_password"
    db_host: "192.168.153.121"
    db_port: "5432"

auth:
  console_passcode: "change-me"
  jwt_secret: "change-this-jwt-secret"

rag:
  ollama_base_url: "http://148.70.18.111:43434"
  embed_model: "qwen3-embedding:0.6b"
  embed_dim: 1024
  chunk_size: 700
  chunk_overlap: 120
  top_k: 6
  query_timeout_seconds: 20
  embed_timeout_seconds: 45
  ingest_timeout_seconds: 300
  max_query_question_runes: 4000
  max_query_top_k: 12
  tencent_base_url: "https://tokenhub.tencentmaas.com/v1"
  tencent_model: "deepseek-v4-flash"
  tencent_api_key: ""
  tencent_timeout_seconds: 60
  tencent_temperature: 0.2
  tencent_top_p: 0.8
  tencent_max_tokens: 1024
  tencent_reasoning_effort: "low"
```

The default config lookup matches `gofurry-admin`: `/etc/gofurry-rag/server.yaml`, then `./config/server.yaml`. You can always pass `--config`.

Environment overrides use the `APP_` prefix:

```bash
APP_SERVER_PORT=8081 APP_RAG_TOP_K=8 go run . --config ./config/server.yaml serve
```

Tencent Cloud inference can be configured through local overrides or environment variables without committing the secret:

```bash
APP_RAG_TENCENT_BASE_URL=https://tokenhub.tencentmaas.com/v1 \
APP_RAG_TENCENT_MODEL=deepseek-v4-flash \
APP_RAG_TENCENT_API_KEY=your-secret-key \
go run . --config ./config/server.yaml serve
```

## Run

```bash
go run . --config ./config/server.yaml serve
```

Open the production console at:

```text
http://127.0.0.1:8080/admin
```

For frontend development:

```bash
cd web
npm run dev
```

Use the Vite URL and let it proxy `/api` to the Go server.

## CLI

```bash
go run . --config ./config/server.yaml version
go run . --config ./config/server.yaml reset-password --password change-me
go run . --config ./config/server.yaml install
go run . --config ./config/server.yaml uninstall
```

`reset-password` updates `auth.console_passcode` in the active `server.yaml`.

## Console

- Login uses the unique passcode from `auth.console_passcode`.
- The overview page refreshes every 5 seconds and shows document totals, chunk totals, status distribution, database connection info, and Ollama connection info.
- The document list refreshes every 3 seconds while the list tab is open, so pending documents move to ready without manual refresh.
- The search page calls the public retrieval API and displays returned sources.
- The new `AI 问答` page streams status updates, sources, and final answers from `POST /api/v1/chat/stream` for real-environment debugging.

## Add Text

Login first and save the HttpOnly session cookie:

```bash
curl -c cookies.txt -X POST http://127.0.0.1:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"password":"change-me"}'
```

Create a manual text document:

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/documents/text \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"title":"GoFurry","content":"GoFurry is a content discovery website.","source_type":"manual"}'
```

Only `content` is required. `title` is strongly recommended for readable sources. `source_type`, `source_id`, and `url` are optional provenance fields:

- `source_type`: where the text came from, such as `manual`, `website`, `nav`, or `game`.
- `source_id`: an external identifier, such as a page slug, article ID, site ID, or game ID.
- `url`: the original page URL, useful for citations and later source review.

These fields are useful for future crawler imports, reindexing, deleting by source, and showing citations. For hand-written text, `source_type: "manual"` is enough.

The service creates a pending document and the ingest worker embeds chunks asynchronously.

## Query

```bash
curl -X POST http://127.0.0.1:8080/api/v1/chat/query \
  -H "Content-Type: application/json" \
  -d '{"question":"What is GoFurry?","top_k":6}'
```

`POST /api/v1/chat/query` is public and does not require the admin session cookie.

Streaming version:

```bash
curl -N -X POST http://127.0.0.1:8080/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"question":"What is GoFurry?","top_k":6}'
```

`POST /api/v1/chat/stream` is public and returns server-sent events that include `status`, `sources`, `delta`, and `done` events.

## Health

```bash
curl http://127.0.0.1:8080/api/v1/health
```

The response includes overall status, database connection information, and Ollama model information.
