# 冒烟测试

1. 构建控制台：

```bash
cd web
npm ci
npm run build
cd ..
```

2. 运行后端检查：

```bash
go test ./...
go vet ./...
go build .
go build ./cmd/server
```

3. 在 `config/server.yaml` 中填入有效 PostgreSQL 密码并启动服务：

```bash
go run . --config ./config/server.yaml serve
```

4. 检查健康接口和探针：

```bash
curl http://127.0.0.1:8080/api/v1/health
curl http://127.0.0.1:8080/livez
curl http://127.0.0.1:8080/readyz
curl http://127.0.0.1:8080/startupz
curl http://127.0.0.1:8080/healthz
```

5. 登录、创建文本文档，等待几秒后检索：

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

ingest worker 完成后，检索结果应该至少包含一个 source。
