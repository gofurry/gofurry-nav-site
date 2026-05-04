# 冒烟测试

## 1. 构建控制台

```bash
cd web
npm ci
npm run build
cd ..
```

## 2. 验证后端

```bash
go test ./...
go vet ./...
go build .
go build ./cmd/server
```

## 3. 配置运行环境

在 `config/server.yaml` 中填入有效 PostgreSQL 密码、控制台口令和 JWT Secret。不要提交真实密码。

## 4. 启动服务

```bash
go run . --config ./config/server.yaml serve
```

## 5. 健康检查和探针

```bash
curl http://127.0.0.1:8080/api/v1/health
curl http://127.0.0.1:8080/livez
curl http://127.0.0.1:8080/readyz
curl http://127.0.0.1:8080/startupz
curl http://127.0.0.1:8080/healthz
```

`/api/v1/health` 应包含 `database` 和 `ollama` 对象。如果依赖不可用，`status` 会变成 `degraded`。

## 6. 验证管理端鉴权

```bash
curl -i http://127.0.0.1:8080/api/v1/admin/documents

curl -c cookies.txt -X POST http://127.0.0.1:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"password":"change-me"}'

curl -b cookies.txt http://127.0.0.1:8080/api/v1/admin/overview
```

第一次请求应返回 401。登录后，overview 应返回文档和 chunk 统计。

## 7. 文本入库和状态刷新

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/documents/text \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"title":"Smoke Test","content":"GoFurry stores searchable knowledge chunks.","source_type":"manual"}'

curl -b cookies.txt "http://127.0.0.1:8080/api/v1/admin/documents?page=1&page_size=6"
```

文档初始状态应为 `pending`。等待几秒后再次请求，状态应变为 `ready`，`chunk_count` 应大于 0。

控制台文档列表每 3 秒自动刷新，整体态势页每 5 秒自动刷新。

## 8. 重新索引

假设上一步创建的文档 ID 为 `1`：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/documents/1/reindex \
  -b cookies.txt

curl -b cookies.txt "http://127.0.0.1:8080/api/v1/admin/documents?page=1&page_size=6"
```

重新索引后该文档会回到 `pending`，旧 chunks 会被删除。等待 worker 处理后，文档应重新变为 `ready`。

## 9. Chunk 编辑重向量化

先查看 chunks：

```bash
curl -b cookies.txt "http://127.0.0.1:8080/api/v1/admin/documents/1/chunks?page=1&page_size=20"
```

假设返回的 chunk ID 为 `1`：

```bash
curl -X PATCH http://127.0.0.1:8080/api/v1/admin/chunks/1 \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"content":"GoFurry stores searchable and editable knowledge chunks."}'
```

响应中的 `has_embedding` 应为 `true`，`embedding_dim` 应为配置的 embedding 维度。

## 10. 检索命中

```bash
curl -X POST http://127.0.0.1:8080/api/v1/chat/query \
  -H "Content-Type: application/json" \
  -d '{"question":"What does GoFurry store?","top_k":3}'
```

ingest worker 完成后，检索结果应至少包含一个 source。

## 11. 控制台手动检查

打开：

```text
http://127.0.0.1:8080/admin
```

检查以下流程：

- 登录口令可进入控制台。
- 文本入库可提交。
- 文件拖拽只接受 10 MiB 内的 txt、md、csv、json、yaml、log、html 文件。
- 不支持或过大的文件会显示拒绝原因。
- 文档页可重新索引，且使用非原生确认模态框。
- Chunks 页可查看、编辑、删除，长文本保持换行并可阅读。
- 检索页能展示 sources 和 score。
