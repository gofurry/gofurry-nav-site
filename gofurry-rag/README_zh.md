# gofurry-rag

![License](https://img.shields.io/github/license/GoFurry/gofurry-rag)
![Go Version](https://img.shields.io/badge/go-1.26.0-00ADD8)

[English](./README.md)

`gofurry-rag` 是面向 GoFurry 内容的轻量 RAG 服务。它使用 PostgreSQL + pgvector 存储知识片段，通过 Ollama 生成 embedding，并在用户提问时返回相关来源片段 `sources`。

## 功能

- 单个 Go 二进制，内嵌 Vue 管理控制台
- PostgreSQL + pgvector 存储
- Ollama embedding client，默认模型 `qwen3-embedding:0.6b`
- 异步文本入库 worker
- 管理 API 使用 Admin Token 鉴权
- 第一版只返回检索来源，不生成自然语言答案

## 快速开始

```bash
cp .env.example .env
go run ./cmd/server
```

开发控制台：

```bash
cd web
npm install
npm run dev
```

构建内嵌控制台并验证 Go 代码：

```bash
cd web
npm ci
npm run build
cd ..
go test ./...
```

## API

- `GET /api/v1/health`
- `POST /api/v1/admin/documents/text`
- `GET /api/v1/admin/documents`
- `GET /api/v1/admin/documents/:id/chunks`
- `DELETE /api/v1/admin/documents/:id`
- `POST /api/v1/chat/query`

管理接口需要：

```http
Authorization: Bearer <RAG_ADMIN_TOKEN>
```

## 配置

配置示例见 [.env.example](./.env.example)。不要提交真实 `.env`、数据库密码或私钥。

## 文档

- [使用说明](./docs/zh/usage.md)
- [冒烟测试](./docs/zh/smoke-test.md)
- [MVP 设计](./docs/rag-design.md)

## 许可证

沿用父仓库 MIT License。
