# 使用说明

## 依赖

- Go 1.26
- Node.js 与 npm
- 已启用 pgvector 的 PostgreSQL
- 已安装 `qwen3-embedding:0.6b` 的 Ollama

## 配置

复制 `.env.example` 为 `.env`，并填写 PostgreSQL 密码：

```env
RAG_DB_DSN=postgres://postgres:your_password@192.168.153.121:5432/postgres?sslmode=disable
RAG_OLLAMA_BASE_URL=http://148.70.18.111:43434
RAG_EMBED_MODEL=qwen3-embedding:0.6b
RAG_ADMIN_TOKEN=change-me
```

## 运行

```bash
go run ./cmd/server
```

生产模式控制台地址：

```text
http://127.0.0.1:8080/admin
```

前端开发模式：

```bash
cd web
npm run dev
```

Vite 会把 `/api` 代理到 Go 服务。

## 文本入库

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/documents/text \
  -H "Authorization: Bearer change-me" \
  -H "Content-Type: application/json" \
  -d '{"title":"GoFurry","content":"GoFurry 是内容发现网站。","source_type":"manual"}'
```

服务会创建 `pending` 文档，后台 worker 异步切分并写入 embedding。

## 检索

```bash
curl -X POST http://127.0.0.1:8080/api/v1/chat/query \
  -H "Content-Type: application/json" \
  -d '{"question":"GoFurry 是什么？","top_k":6}'
```
