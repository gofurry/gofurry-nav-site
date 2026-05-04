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
go vet ./...
go build .
go build ./cmd/server
```

3. Edit `config/server.yaml` with a valid PostgreSQL password and start the service:

```bash
go run . --config ./config/server.yaml serve
```

4. Check health and probes:

```bash
curl http://127.0.0.1:8080/api/v1/health
curl http://127.0.0.1:8080/livez
curl http://127.0.0.1:8080/readyz
curl http://127.0.0.1:8080/startupz
curl http://127.0.0.1:8080/healthz
```

5. Login, create a text document, wait a few seconds, then query:

```bash
curl -c cookies.txt -X POST http://127.0.0.1:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"password":"change-me"}'

curl -X POST http://127.0.0.1:8080/api/v1/admin/documents/text \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"title":"Smoke Test","content":"GoFurry stores searchable knowledge chunks.","source_type":"manual"}'

curl -X POST http://127.0.0.1:8080/api/v1/chat/query \
  -H "Content-Type: application/json" \
  -d '{"question":"What does GoFurry store?","top_k":3}'
```

The query should return at least one source after the ingest worker finishes.
