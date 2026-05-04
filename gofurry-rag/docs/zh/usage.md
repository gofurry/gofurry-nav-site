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
  embed_dim: 1024
  chunk_size: 700
  chunk_overlap: 120
  top_k: 6
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

`reset-password` 会更新当前 `server.yaml` 里的 `auth.console_passcode`。

## 控制台

- 登录使用 `auth.console_passcode` 配置的唯一口令。
- 管理接口通过 HttpOnly JWT Cookie 鉴权，不再使用 Admin Token 或 Bearer Header。
- 整体态势页每 5 秒自动刷新，展示文档总量、chunk 总量、状态分布、数据库连接信息和 Ollama 连接信息。
- 文档管理支持手动文本入库，也支持拖拽文件和批量导入文件。
- 文件导入限制单文件最大 10 MiB，允许 `.txt`、`.md`、`.csv`、`.json`、`.yaml`、`.yml`、`.log`、`.html`、`.htm`。
- 文档列表每页 6 条，打开文档 tab 时每 3 秒自动刷新。
- 文档可以重新索引；重新索引会删除旧 chunks，把文档设为 `pending`，由后台 worker 重新切分和向量化。
- Chunks tab 左侧文档列表每页 7 条，可按文档标题或 ID 搜索。
- Chunk 支持查看、编辑和删除；编辑保存时会重新生成 embedding 并写回 pgvector。
- 文档检索页调用公开检索接口，并展示返回的 sources。

## 导入规范

只有 `content` 是必填字段。`title` 强烈建议填写，方便检索结果和来源展示。

`source_type`、`source_id`、`url` 是可选的来源追踪字段：

- 手动录入：`source_type=manual`，`source_id` 和 `url` 可空。
- 文件导入：`source_type=file`，`source_id` 使用原始文件名，标题使用去掉后缀的文件名。
- 网页内容：推荐 `source_type=website`，`source_id` 使用页面 slug 或路径，`url` 填原始页面地址。
- 导航条目：推荐 `source_type=nav`，`source_id` 使用站点 ID 或 slug。
- 游戏内容：推荐 `source_type=game`，`source_id` 使用游戏 ID 或 slug。

这些字段对后续爬虫导入、重新索引、按来源删除、展示引用很有用。

## 文本入库

先登录并保存 HttpOnly Session Cookie：

```bash
curl -c cookies.txt -X POST http://127.0.0.1:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"password":"change-me"}'
```

创建手动文本文档：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/documents/text \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"title":"GoFurry","content":"GoFurry is a content discovery website.","source_type":"manual"}'
```

服务会创建 `pending` 文档，ingest worker 会异步切分并写入 embedding。

## 重新索引

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/documents/1/reindex \
  -b cookies.txt
```

重新索引会删除该文档旧 chunks，把文档设为 `pending`。在 worker 重新处理完成前，该文档会短暂不可检索。

## 检索

```bash
curl -X POST http://127.0.0.1:8080/api/v1/chat/query \
  -H "Content-Type: application/json" \
  -d '{"question":"What is GoFurry?","top_k":6}'
```

`POST /api/v1/chat/query` 是公开接口，不需要管理端 Cookie。

## 健康检查

```bash
curl http://127.0.0.1:8080/api/v1/health
```

返回内容包含整体状态、数据库连接信息和 Ollama 模型信息。

## 后续路线图

后续优化方向见 [Roadmap](./roadmap.md)。
