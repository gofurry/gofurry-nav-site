# gofurry-rag

![License](https://img.shields.io/github/license/GoFurry/gofurry-rag)
![Go Version](https://img.shields.io/badge/go-1.26.0-00ADD8)

[English](./README.md)

`gofurry-rag` 是面向 GoFurry 内容的轻量 RAG 服务。它使用 PostgreSQL + pgvector 存储知识片段，通过 Ollama 生成 embedding，并在用户提问时返回相关来源片段 `sources`。

## 功能

- Cobra CLI 与 Viper `server.yaml` 配置，启动骨架对齐 `gofurry-admin`
- 单个 Go 二进制，内嵌 Vue + Tailwind 暗色控制台
- PostgreSQL + pgvector 存储
- Ollama embedding client，默认模型 `qwen3-embedding:0.6b`
- 异步文本入库 worker
- 控制台使用唯一口令登录，服务端签发 HttpOnly JWT Cookie
- 文本入库支持手动表单和文件拖拽/批量导入
- 文件导入限制单文件 10 MiB，支持 txt、md、csv、json、yaml、log、html
- 文档管理支持状态、来源类型、分类、语言过滤，分页、删除确认、单文档重新索引
- 支持按范围批量重新索引和失败文档重试
- Chunks 支持查看、编辑、删除；编辑保存时会重新生成 embedding 并写回 pgvector
- 检索接口公开，只返回 `sources`，并提供 rank、score、来源和 chunk 调试字段；支持按 `source_type`、`document_ids`、`category`、`language` 过滤；暂不生成自然语言答案
- 控制台支持 chunk 参数切分预览，不调用 Ollama、不写数据库
- 向量化输入会包含标题和来源上下文，展示和保存的 chunk 内容仍保持原文
- 控制台整体态势自动刷新，展示文档、chunk、失败摘要、队列规模、数据库和 Ollama 状态

## 快速开始

编辑 `config/server.yaml`，然后启动服务：

```bash
go run . --config ./config/server.yaml serve
```

打开内嵌控制台：

```text
http://127.0.0.1:8080/admin
```

常用 CLI：

```bash
go run . --config ./config/server.yaml version
go run . --config ./config/server.yaml reset-password --password change-me
go run . --config ./config/server.yaml install
go run . --config ./config/server.yaml uninstall
```

构建内嵌控制台并验证 Go 代码：

```bash
cd web
npm ci
npm run build
cd ..
go test ./...
go vet ./...
go build .
go build ./cmd/server
```

## API

- `GET /api/v1/health`
- `GET /api/v1/admin/auth/state`
- `POST /api/v1/admin/auth/login`
- `POST /api/v1/admin/auth/logout`
- `GET /api/v1/admin/auth/me`
- `GET /api/v1/admin/overview`
- `POST /api/v1/admin/documents/text`
- `GET /api/v1/admin/documents`
- `POST /api/v1/admin/documents/reindex`
- `POST /api/v1/admin/documents/retry-failed`
- `GET /api/v1/admin/documents/:id/chunks`
- `POST /api/v1/admin/documents/:id/reindex`
- `PATCH /api/v1/admin/chunks/:id`
- `DELETE /api/v1/admin/chunks/:id`
- `POST /api/v1/admin/debug/chunk-preview`
- `DELETE /api/v1/admin/documents/:id`
- `POST /api/v1/chat/query`
- `GET /livez`、`GET /readyz`、`GET /startupz`、`GET /healthz`

管理接口需要先通过 `/api/v1/admin/auth/login` 登录。服务会把 JWT 写入 HttpOnly Cookie，方式与 `gofurry-admin` 类似。`POST /api/v1/chat/query` 保持公开。

## 配置

运行时配置读取 `server.yaml`。默认搜索 `/etc/gofurry-rag/server.yaml` 和 `./config/server.yaml`；也可以通过 `--config` 指定任意文件。

环境变量覆盖使用 `APP_` 前缀：

```bash
APP_SERVER_PORT=8081 APP_RAG_OLLAMA_BASE_URL=http://127.0.0.1:11434 go run . --config ./config/server.yaml serve
```

不要提交真实数据库密码、控制台口令或 JWT Secret。

## 文档

- [使用说明](./docs/zh/usage.md)
- [Roadmap](./docs/zh/roadmap.md)
- [冒烟测试](./docs/zh/smoke-test.md)
- [MVP 设计](./docs/rag-design.md)

## 许可证

沿用父仓库 MIT License。
