# gofurry-rag

![License](https://img.shields.io/github/license/GoFurry/gofurry-rag)
![Go Version](https://img.shields.io/badge/go-1.26.0-00ADD8)

[English](./README.md)

`gofurry-rag` 是面向 GoFurry 内容场景的轻量 RAG 服务。它使用 PostgreSQL + pgvector 存储知识片段，使用本地 Ollama 生成 embedding，并通过腾讯云推理服务生成基于引用的 AI 问答结果。

## 当前能力

- 单二进制 Go 服务，内嵌 Vue 控制台
- PostgreSQL + pgvector 知识库存储
- 本地 Ollama embedding 与并发保护
- 文本文档入库、批量重建、失败重试
- Chunk 预览、编辑、删除与重新向量化
- 检索调试、AI 问答、SSE 流式回答
- 引用详情调试视图，可查看命中文档、chunk、元数据与血缘
- 管理员登录态使用 HttpOnly JWT Cookie

## AI 问答边界

- `POST /api/v1/chat/query` 和 `POST /api/v1/chat/stream` 仍然是公开问答入口
- 默认返回最终答案、`sources` 和 usage 信息
- `include_details=true` 只允许已登录的控制台管理员使用
- 详细引用数据仅用于控制台调试，不建议直接暴露给业务前端
- `GET /api/v1/health` 现已改为管理员接口，会执行深度健康探测

## 快速开始

编辑配置文件：

```bash
go run . --config ./config/server.yaml serve
```

打开控制台：

```text
http://127.0.0.1:9997/admin
```

常用命令：

```bash
go run . --config ./config/server.yaml version
go run . --config ./config/server.yaml reset-password --password change-me
go run . --config ./config/server.yaml install
go run . --config ./config/server.yaml uninstall
```

前端构建与后端验证：

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

## 主要 API

公开接口：

- `POST /api/v1/chat/query`
- `POST /api/v1/chat/stream`
- `GET /livez`
- `GET /readyz`
- `GET /startupz`
- `GET /healthz`

管理员接口：

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
- `DELETE /api/v1/admin/documents/:id`
- `PATCH /api/v1/admin/chunks/:id`
- `DELETE /api/v1/admin/chunks/:id`
- `POST /api/v1/admin/debug/chunk-preview`

## 配置说明

运行时配置来自 `server.yaml`。默认会搜索：

- `/etc/gofurry-rag/server.yaml`
- `./config/server.yaml`

也可以通过 `--config` 指定任意配置文件。

环境变量覆盖使用 `APP_` 前缀，例如：

```bash
APP_SERVER_PORT=8081 APP_RAG_OLLAMA_BASE_URL=http://127.0.0.1:11434 go run . --config ./config/server.yaml serve
```

腾讯云推理建议使用环境变量或本地覆盖文件提供：

```bash
APP_RAG_TENCENT_BASE_URL=https://tokenhub.tencentmaas.com/v1
APP_RAG_TENCENT_MODEL=deepseek-v4-flash
APP_RAG_TENCENT_API_KEY=your-api-key
```

生产模式下会额外校验：

- 不能使用占位的控制台口令
- 不能使用占位的 JWT Secret
- `auth.cookie_secure` 不能为 `false`

## 文档

- [使用说明](./docs/zh/usage.md)
- [Roadmap](./docs/zh/roadmap.md)
- [冒烟测试](./docs/zh/smoke-test.md)
- [部署与回滚](./docs/zh/deployment.md)
- [MVP 设计](./docs/rag-design.md)
- [代码审计](./docs/code-audit.md)

## 许可

沿用父仓库 MIT License。
