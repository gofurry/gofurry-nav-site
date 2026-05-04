# Usage

## Requirements

- Go 1.26
- Node.js and npm
- PostgreSQL with pgvector
- Ollama with `qwen3-embedding:0.6b`

## Configure

Copy `.env.example` to `.env` and set the PostgreSQL password:

```env
RAG_DB_DSN=postgres://postgres:your_password@192.168.153.121:5432/postgres?sslmode=disable
RAG_OLLAMA_BASE_URL=http://148.70.18.111:43434
RAG_EMBED_MODEL=qwen3-embedding:0.6b
RAG_ADMIN_TOKEN=change-me
```

## Run

```bash
go run ./cmd/server
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

## Add Text

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/documents/text \
  -H "Authorization: Bearer change-me" \
  -H "Content-Type: application/json" \
  -d '{"title":"GoFurry","content":"GoFurry is a content discovery website.","source_type":"manual"}'
```

The service creates a pending document and the ingest worker embeds chunks asynchronously.

## Query

```bash
curl -X POST http://127.0.0.1:8080/api/v1/chat/query \
  -H "Content-Type: application/json" \
  -d '{"question":"What is GoFurry?","top_k":6}'
```
