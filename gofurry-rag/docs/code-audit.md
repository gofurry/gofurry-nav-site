# Go 代码审计报告

## 摘要

`gofurry-rag` 当前已经是一个可用的单二进制 RAG 服务，足够支撑内部调试和知识库运维场景。核心链路已经具备：文档入库、切块、Ollama embedding、pgvector 检索、控制台管理、reindex / retry、腾讯云问答生成、引用详情、SSE 流式问答，以及 Ollama 侧的并发保护。

本报告已按当前代码状态更新。最初审计中提到的部分安全和可用性问题已经收口，包括深度健康检查改为管理员可见、`include_details` 仅允许管理员使用、HTTP handler 传递请求 context、worker 状态统计不再被空闲 worker 粗暴清零。

当前仍不建议直接把公开问答接口接到业务前端。主要剩余工作是业务侧鉴权边界、引用字段裁剪、限流配额策略，以及生产环境配置切换。

## 审计范围

- HTTP / API 暴露面与鉴权边界
- 配置与密钥管理
- AI 问答与引用链路
- Ollama 并发控制与 worker 行为
- 面向业务前端接入时的生产可用性

## 当前状态概览

| 项目 | 状态 | 说明 |
|---|---|---|
| 开发环境临时密钥 | 已知开发态 | `config/server.yaml` 保留开发环境临时值；生产环境必须替换 |
| 公开问答接口 | 待收口 | `chat/query` 和 `chat/stream` 仍公开，默认 sources 仍包含 chunk 内容 |
| `include_details` 管理员保护 | 已处理 | 仅管理员 cookie 可请求详细引用信息 |
| 深度健康检查保护 | 已处理 | `/api/v1/health` 已改为 admin-only |
| 公共探针轻量化 | 已处理 | `/livez`、`/readyz`、`/startupz`、`/healthz` 只做轻量状态返回 |
| debug / pprof 暴露 | 部分处理 | 当前未看到 pprof 注册；提交配置仍是开发用 `debug` |
| 请求 context 传递 | 已处理 | 主要 handler 已透传请求 context |
| 多 worker 状态统计 | 已处理 | `markIdle()` 不再重置全局 active worker 数 |
| 文档同步 | 部分处理 | 中文 README / roadmap 已更新；英文 usage / deployment / smoke-test 仍需补认证说明 |

## 仍需处理的问题

### P1 - 公开问答接口仍缺少业务接入边界

证据：

- `internal/api/server.go:47`
- `internal/api/server.go:48`
- `internal/service/service.go:126`
- `internal/db/repo.go:61`
- `internal/db/repo.go:71`

影响说明：

- `/api/v1/chat/query` 和 `/api/v1/chat/stream` 当前仍是公开接口。
- `include_details=true` 已经只允许管理员使用，这是正确的收口。
- 但默认 `sources` 仍会返回命中 chunk 的 `content`，如果直接接业务前端，调用方仍可能拿到知识库片段原文。
- 当前还没有业务调用方鉴权、租户边界、知识库范围限制和配额策略。

建议：

- 把“控制台调试问答”和“业务前端问答”拆成不同响应边界。
- 业务前端接口默认只返回最终答案、必要的引用标题、URL、rank、score 等裁剪字段。
- 业务流量接入前先确定身份边界、知识域边界、限流和配额策略。
- 如果业务侧确实需要公开调用，建议增加 API key、签名、来源限制或后端代理层。

### P2 - 提交配置仍是开发态

证据：

- `config/server.yaml:7`
- `config/server.yaml:24`
- `config/server.yaml:29`
- `config/server.yaml:30`
- `config/server.yaml:33`
- `config/server.yaml:57`

影响说明：

- `config/server.yaml` 当前保留的是开发环境临时密钥、开发数据库连接和 `debug` 模式。
- 这是当前开发阶段可接受的状态，不再按“已泄露生产密钥”处理。
- 风险边界必须写清楚：这个文件不能直接作为生产配置发布或部署。
- 生产环境需要使用另一套未提交的配置，或通过环境变量覆盖敏感项。

建议：

- 保持当前文件作为开发示例可以接受，但建议在注释或部署文档里明确“仅限开发环境”。
- 生产环境必须替换数据库密码、控制台口令、JWT secret、腾讯云 API key。
- 生产模式应使用 `prod` 或 `release`，并设置 `auth.cookie_secure=true`。
- 上线前执行一次配置检查和 secret scan，确认没有真实生产密钥进入仓库。

### P2 - 英文文档仍需要同步认证边界

证据：

- `README.md`
- `docs/usage.md`
- `docs/deployment.md`
- `docs/smoke-test.md`

影响说明：

- 中文 README 和 roadmap 已经基本说明当前问答、引用和 health 边界。
- 英文 usage / deployment / smoke-test 里仍有直接 `curl /api/v1/health` 的示例，但 `/api/v1/health` 当前已经需要管理员 cookie。
- 这会让新接手的人误以为深度健康检查仍是公开接口，或按旧方式做部署检查。

建议：

- 英文文档补充登录获取 cookie 后再访问 `/api/v1/health` 的示例。
- 把公共探针和深度诊断接口明确拆开描述。
- 在 API 文档里注明 `include_details=true` 仅管理员可用。

## 已处理的问题

### P1 - `include_details` 公开泄露详细引用信息

处理状态：已处理。

当前行为：

- `include_details=false` 时，公开问答接口不会返回完整 citation document 详情。
- `include_details=true` 时，会调用管理员 cookie 校验。
- 未登录请求会返回 unauthorized。

保留风险：

- 公开接口默认 `sources` 仍包含 chunk 内容，所以业务接入前仍需要做字段裁剪。

### P1 - 公共深度健康检查暴露内部基础设施并触发重型探活

处理状态：已处理。

当前行为：

- `/api/v1/health` 已挂载 `requireAdmin`，只允许控制台管理员访问。
- 公共 `/livez`、`/readyz`、`/startupz`、`/healthz` 只返回轻量状态。
- 深度诊断仍会展示数据库、Ollama、Tencent、worker 状态，但已经不在公开路径上。

### P2 - HTTP handler 丢弃请求取消信号

处理状态：已处理。

当前行为：

- 主要 HTTP handler 已使用请求 context 调用 service。
- 流式问答继续使用请求 context，客户端断开时可以向下游传播取消信号。

### P2 - 多 worker 状态统计会被空闲 worker 覆盖

处理状态：已处理。

当前行为：

- `markIdle()` 不再把 `activeWorkers` 直接清零。
- success / failed 路径会按活动 worker 数递减并维护状态。

## 就绪性结论

### 内部调试

适合。

- 文本 / 手动入库已经具备
- chunk 预览、reindex、retry、chunk 编辑已经具备
- 检索和 AI 回答已经具备
- 引用详情和调试信息已经具备
- Ollama 并发保护已经具备
- health / overview / worker 可观测性已经具备

### 业务前端接入

暂时不建议直接接。

最低限度建议先完成：

1. 明确业务问答 API 的鉴权和调用方边界。
2. 对业务响应里的 sources / citations 做字段裁剪。
3. 确认知识库范围、租户范围或业务域范围。
4. 为公开问答路径增加更细的限流、配额和错误降级策略。
5. 用生产环境专用配置替换开发临时密钥，并启用安全 cookie。

完成这些之后，`gofurry-rag` 很适合进入第一版业务前端接入阶段，尤其适合单租户或复杂度较低的业务场景。
