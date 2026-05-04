# 使用说明

## 环境要求

- Go 1.26
- Node.js 与 npm
- 已安装 pgvector 的 PostgreSQL
- 已安装 `qwen3-embedding:0.6b` 的 Ollama

## 配置

编辑 `config/server.yaml`。最关键的字段如下：

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

默认配置搜索方式与 `gofurry-admin` 一致：先找 `/etc/gofurry-rag/server.yaml`，再找当前工作目录下的 `./config/server.yaml`。也可以通过 `--config` 指定。

环境变量覆盖使用 `APP_` 前缀：

```bash
APP_SERVER_PORT=8081 APP_RAG_TOP_K=8 go run . --config ./config/server.yaml serve
```

## 运行

```bash
go run . --config ./config/server.yaml serve
```

生产控制台地址：

```text
http://127.0.0.1:8080/admin
```

前端开发：

```bash
cd web
npm run dev
```

使用 Vite URL，并让它把 `/api` 代理到 Go 服务。

## CLI

```bash
go run . --config ./config/server.yaml version
go run . --config ./config/server.yaml reset-password --password change-me
go run . --config ./config/server.yaml install
go run . --config ./config/server.yaml uninstall
```

## 文本入库

先登录并保存 HttpOnly Session Cookie：

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

服务会创建 pending 文档，ingest worker 会异步切分并写入 embedding。

## 检索

```bash
curl -X POST http://127.0.0.1:8080/api/v1/chat/query \
  -H "Content-Type: application/json" \
  -d '{"question":"What is GoFurry?","top_k":6}'
```
