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
  query_timeout_seconds: 20
  embed_timeout_seconds: 45
  ingest_timeout_seconds: 300
  max_query_question_runes: 4000
  max_query_top_k: 12
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
- 整体态势页会额外展示失败文档数、待处理队列规模和最近失败摘要。
- 文档管理支持手动文本入库，也支持拖拽文件和批量导入文件。
- 文件导入限制单文件最大 10 MiB，允许 `.txt`、`.md`、`.csv`、`.json`、`.yaml`、`.yml`、`.log`、`.html`、`.htm`。
- 文档列表每页 6 条，打开文档 tab 时每 3 秒自动刷新。
- 文档列表支持按 `status`、`source_type`、`category`、`language` 过滤，并可按当前过滤条件批量重新索引或重试失败文档。
- 文档可以重新索引；重新索引会删除旧 chunks，把文档设为 `pending`，由后台 worker 重新切分和向量化。
- Chunks tab 左侧文档列表每页 7 条，可按文档标题或 ID 搜索。
- Chunk 支持查看、编辑和删除；编辑保存时会重新生成 embedding 并写回 pgvector。
- 文档检索页调用公开检索接口，并展示 rank、score、document、chunk、source、URL、token_count 和原始 chunk 内容。
- 文档检索页支持按 `source_type`、`document_ids`、`category`、`language` 收窄检索范围。
- 文档检索页提供切分预览工具，可对已有文档或临时文本预览不同 `chunk_size/chunk_overlap` 的切分结果。

## 导入规范

只有 `content` 是必填字段。`title` 强烈建议填写，方便检索结果和来源展示。

`source_type`、`source_id`、`url` 是可选的来源追踪字段：

- 手动录入：`source_type=manual`，`source_id` 和 `url` 可空。
- 文件导入：`source_type=file`，`source_id` 使用原始文件名，标题使用去掉后缀的文件名。
- 网页内容：推荐 `source_type=website`，`source_id` 使用页面 slug 或路径，`url` 填原始页面地址。
- 导航条目：推荐 `source_type=nav`，`source_id` 使用站点 ID 或 slug。
- 游戏内容：推荐 `source_type=game`，`source_id` 使用游戏 ID 或 slug。

推荐的 metadata 顶层字段：

- `category`：内容分类，例如 `intro`、`faq`、`policy`
- `language`：语言，例如 `zh-CN`、`en-US`
- `tags`：字符串数组，多个标签
- `author`：内容作者或责任人
- `published_at`：发布时间字符串，例如 `2026-05-10`

这些字段对后续批量重建、失败重试、按范围检索和展示引用很有用。

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
  -d '{
    "title":"gofurry",
    "content":"gofurry is a content discovery website.",
    "source_type":"manual",
    "metadata":{
      "category":"intro",
      "language":"en-US",
      "tags":["about","brand"]
    }
  }'
```

服务会创建 `pending` 文档，ingest worker 会异步切分并写入 embedding。向量化时会把标题、来源类型、来源 ID、URL 和 chunk 正文组合成 embedding 输入；数据库中保存和展示的 chunk 内容仍保持原文。

## 重新索引

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/documents/1/reindex \
  -b cookies.txt
```

重新索引会删除该文档旧 chunks，把文档设为 `pending`。在 worker 重新处理完成前，该文档会短暂不可检索。旧数据如果需要使用新的 embedding 输入模板，也通过单文档重新索引迁移。

批量重新索引当前过滤范围：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/documents/reindex \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "scope":"filters",
    "filters":{
      "source_type":["website"],
      "category":["faq"],
      "language":["zh-CN"]
    }
  }'
```

重试失败文档：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/documents/retry-failed \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"scope":"all"}'
```

## 切分预览

切分预览是登录态管理接口，只预览 splitter 结果，不调用 Ollama，不写数据库。

使用已有文档正文：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/debug/chunk-preview \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"document_id":1}'
```

使用临时文本和自定义参数：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/debug/chunk-preview \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"text":"gofurry knowledge text","variants":[{"chunk_size":500,"chunk_overlap":80},{"chunk_size":700,"chunk_overlap":120}]}'
```

响应会返回每组参数的 chunk 数、最短/最长/平均字符数和前 20 个 chunk 预览。

## 检索

```bash
curl -X POST http://127.0.0.1:8080/api/v1/chat/query \
  -H "Content-Type: application/json" \
  -d '{"question":"What is gofurry?","top_k":6}'
```

`POST /api/v1/chat/query` 是公开接口，不需要管理端 Cookie。

返回的 `sources` 会包含调试字段：`source_type`、`source_id`、`chunk_index`、`token_count`。这些字段用于判断命中来自哪个来源、哪个文档和哪个 chunk。

带过滤条件的检索示例：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/chat/query \
  -H "Content-Type: application/json" \
  -d '{
    "question":"What is gofurry?",
    "top_k":6,
    "filters":{
      "source_type":["website"],
      "category":["intro"],
      "language":["en-US"]
    }
  }'
```

## 健康检查

```bash
curl http://127.0.0.1:8080/api/v1/health
```

返回内容包含整体状态、数据库连接信息和 Ollama 模型信息。

## 后续路线图

后续优化方向见 [Roadmap](./roadmap.md)。
