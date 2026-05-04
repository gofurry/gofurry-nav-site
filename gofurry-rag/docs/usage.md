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
```

The default config lookup matches `gofurry-admin`: `/etc/gofurry-rag/server.yaml`, then `./config/server.yaml`. You can always pass `--config`.

Environment overrides use the `APP_` prefix:

```bash
APP_SERVER_PORT=8081 APP_RAG_TOP_K=8 go run . --config ./config/server.yaml serve
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

## Add Text

Login first and save the HttpOnly session cookie:

```bash
curl -c cookies.txt -X POST http://127.0.0.1:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"password":"change-me"}'
```

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/documents/text \
  -b cookies.txt \
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
