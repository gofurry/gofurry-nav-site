# Smoke Test

1. Build the admin console:

```bash
cd web
npm ci
npm run build
cd ..
```

2. Run backend checks:

```bash
go test ./...
```

3. Configure `.env` with a valid PostgreSQL password and start the service:

```bash
go run ./cmd/server
```

4. Check health:

```bash
curl http://127.0.0.1:8080/api/v1/health
```

Expected result:

```json
{"code":1,"message":"success","data":{"status":"ok"}}
```

5. Create a text document, wait a few seconds, then query:

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/documents/text \
  -H "Authorization: Bearer change-me" \
  -H "Content-Type: application/json" \
  -d '{"title":"Smoke Test","content":"GoFurry stores searchable knowledge chunks.","source_type":"manual"}'

curl -X POST http://127.0.0.1:8080/api/v1/chat/query \
  -H "Content-Type: application/json" \
  -d '{"question":"What does GoFurry store?","top_k":3}'
```

The query should return at least one source after the ingest worker finishes.
