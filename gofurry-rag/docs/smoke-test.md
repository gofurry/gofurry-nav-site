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

3. Edit `config/server.yaml` with a valid PostgreSQL password, console passcode, and JWT secret.

4. Start the service:

```bash
go run . --config ./config/server.yaml serve
```

5. Check health and probes:

```bash
curl http://127.0.0.1:8080/api/v1/health
curl http://127.0.0.1:8080/livez
curl http://127.0.0.1:8080/readyz
curl http://127.0.0.1:8080/startupz
curl http://127.0.0.1:8080/healthz
```

The health response should include `database` and `ollama` objects. If either dependency is unavailable, `status` becomes `degraded`.

6. Confirm admin auth:

```bash
curl -i http://127.0.0.1:8080/api/v1/admin/documents

curl -c cookies.txt -X POST http://127.0.0.1:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"password":"change-me"}'

curl -b cookies.txt http://127.0.0.1:8080/api/v1/admin/overview
```

The first request should return 401. After login, overview should return document and chunk stats.

7. Create a text document, wait a few seconds, then query:

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/documents/text \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"title":"Smoke Test","content":"GoFurry stores searchable knowledge chunks.","source_type":"manual"}'

curl -X POST http://127.0.0.1:8080/api/v1/chat/query \
  -H "Content-Type: application/json" \
  -d '{"question":"What does GoFurry store?","top_k":3}'
```

The query should return at least one source after the ingest worker finishes. The console document list refreshes every 3 seconds, and the overview page refreshes every 5 seconds.
