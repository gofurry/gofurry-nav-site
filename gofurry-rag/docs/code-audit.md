# Go 代码审计报告

## 摘要

`gofurry-rag` 现在已经是一个可用的单二进制 RAG 服务，足够支撑内部调试和知识库运维场景。核心链路已经具备：文档入库、切块、Ollama embedding、pgvector 检索、控制台管理、reindex / retry、腾讯云问答生成、引用详情，以及 Ollama 侧的并发保护。

但它当前还不适合直接接到业务前端。主要阻塞点不是功能缺失，而是安全和生产加固还没收口，尤其是仓库中的敏感信息泄露、公开接口暴露过多内部数据、以及调试模式和请求生命周期处理不够严谨。

## 审计范围

- 全仓审计，重点关注以下方面：
- HTTP / API 暴露面与鉴权边界
- 配置与密钥管理
- AI 问答与引用链路
- Ollama 并发控制与 worker 行为
- 面向业务前端接入时的生产可用性

## 严重级别概览

| 严重级别 | 数量 | 含义 |
|---|---:|---|
| P0 | 1 | 致命 |
| P1 | 3 | 高 |
| P2 | 3 | 中 |
| P3 | 1 | 低 |

## 问题明细

### P0 - 致命

#### P0.1 共享配置文件中提交了真实敏感凭据

证据：
- `config/server.yaml:24`
- `config/server.yaml:29`
- `config/server.yaml:30`
- `config/server.yaml:57`

影响说明：
- 当前仓库里直接保存了数据库密码、控制台登录口令、JWT 签名密钥和腾讯云 API key。
- 任何能访问仓库的人都可以伪造管理员身份、签发有效会话、访问数据库，或者消耗外部模型额度。
- 这不是一般性的配置卫生问题，而是需要立即处理的安全事件。

建议：
- 立刻从 `config/server.yaml` 中移除真实密钥，并轮换所有已暴露凭据。
- 仓库里只保留占位值。
- 真实配置改为环境变量或本地未跟踪覆盖文件。
- 后续增加 pre-commit 或 CI 的 secret scan。

### P1 - 高

#### P1.1 公开问答接口可以在无鉴权、无租户隔离的情况下泄露知识库全文

证据：
- `internal/api/server.go:47`
- `internal/api/server.go:48`
- `internal/service/service.go:62`
- `internal/service/chat.go:182`
- `internal/service/chat.go:259`
- `internal/service/chat.go:275`

影响说明：
- `/api/v1/chat/query` 和 `/api/v1/chat/stream` 当前是公开接口。
- 当请求带 `include_details=true` 时，服务会返回命中文档的完整正文、metadata、状态、重试信息和血缘信息。
- 当前没有业务调用方鉴权、没有租户边界、没有知识库范围限制、也没有字段级裁剪。
- 这在调试台里很方便，但如果直接接业务前端，任何能调用 API 的客户端都可能拿到内部知识库原文，而不只是最终答案。

建议：
- 把当前“调试型问答响应”与“业务型问答 API”拆开。
- 业务流量必须先做鉴权，并在检索前加知识库范围或租户范围限制。
- `include_details` 只保留给 admin 接口，或者从公开接口中移除。
- 业务前端只返回裁剪后的引用字段。

#### P1.2 公共健康检查会真实探活 Ollama 和腾讯云，并暴露内部基础设施信息

证据：
- `internal/api/server.go:29`
- `internal/api/server.go:108`
- `internal/service/service.go:206`
- `internal/service/service.go:213`
- `internal/service/service.go:220`
- `internal/service/service.go:251`
- `internal/service/service.go:257`
- `internal/embedder/ollama.go:36`
- `internal/tencentmaas/client.go:133`

影响说明：
- `GET /api/v1/health` 当前是公开的。
- 它会暴露数据库主机/端口、Ollama base URL、腾讯云 base URL / model，以及队列状态。
- 更重要的是，它不是“轻量看状态”，而是真的会发腾讯云 `chat/completions` 请求和 Ollama `/api/tags` 请求。
- 如果这个接口对外可达，别人反复打它就能消耗腾讯云额度、占用本地 Ollama 并发槽位，还能摸清你的内部基础设施。

建议：
- 把 `/api/v1/health` 改成 admin-only，或者给公共版本去掉依赖细节。
- 健康检查拆成两层：
- 一个轻量的 public liveness / readiness
- 一个受保护的深度诊断接口
- 不要在未鉴权的健康接口里调用付费模型服务。

#### P1.3 当前提交配置启用了 debug 模式，并会公开挂载 `pprof`

证据：
- `config/server.yaml:7`
- `internal/transport/http/router/router.go:184`

影响说明：
- 当前提交配置里 `server.mode` 是 `debug`。
- 在 debug 模式下，服务会全局注册 `pprof`。
- 公开 profiling 接口会泄露运行时内部信息，也可能被用于 goroutine / 内存侧的压测和攻击。

建议：
- 把提交配置改为 `prod` 或 `release`。
- `pprof` 只在本地或受保护环境显式开启。
- profiling 应该是运维工具，而不是默认暴露的运行面。

### P2 - 中

#### P2.1 除了流式问答以外，大多数 HTTP handler 仍然丢弃请求取消信号

证据：
- `internal/api/server.go:112`
- `internal/api/server.go:120`
- `internal/api/server.go:135`
- `internal/api/server.go:146`
- `internal/api/server.go:169`
- `internal/api/server.go:184`
- `internal/api/server.go:199`
- `internal/api/server.go:214`
- `internal/api/server.go:228`
- `internal/api/server.go:247`
- `internal/api/server.go:276`
- `internal/api/server.go:291`

影响说明：
- 流式问答这条链路已经改成了请求级 context，但其他大多数 API 还是从 `context.Background()` 起任务。
- 这意味着客户端断开连接或中间件超时后，底层的数据库、embedding、reindex 等逻辑仍然可能继续执行。
- 在你这种本地弱推理环境里，这会直接浪费 Ollama 计算容量，也会让系统过载后恢复更慢。

建议：
- 所有 handler 都改为透传真实请求 context。
- 只从请求 context 派生子 context。
- 服务内仍可保留超时，但要绑定在调用方取消链路上。

#### P2.2 当 `ingest_workers > 1` 时，worker 状态统计会失真

证据：
- `internal/ingest/worker.go:75`
- `internal/ingest/worker.go:91`
- `internal/ingest/worker.go:261`
- `internal/ingest/worker.go:270`
- `internal/ingest/worker.go:280`

影响说明：
- 任何一个 worker 发现当前没有待处理文档时，都会调用 `markIdle()`，而这个方法会直接把 `activeWorkers = 0`。
- 如果未来开启多个 worker，其中一个空闲 worker 可能把另一个正在处理中的 worker 状态覆盖掉。
- 这不会直接造成数据错误，但会让你刚做出来的 worker 观测面板在多 worker 模式下不可信。

建议：
- 按具体 worker / goroutine 跟踪活动状态，而不是全局粗暴清零。
- `markIdle()` 不应该直接清空全局 active 数，除非能确认是对应 worker 自己从 active 变 idle。
- 增加多 worker 状态测试。

#### P2.3 当前默认配置对直接接业务前端仍然过于宽松

证据：
- `config/config.go:470`
- `config/config.go:473`
- `config/server.yaml:33`
- `internal/auth/auth.go:77`

影响说明：
- 如果配置不完整，代码仍然接受默认占位的 console passcode 和 JWT secret。
- 当前提交配置里的 `cookie_secure` 还是 `false`。
- 这对本地开发没问题，但对 HTTPS 下的业务前端来说强度不够。

建议：
- 如果仍然是占位 secret，生产模式下直接启动失败。
- 非本地环境强制要求 `cookie_secure=true`。
- 补一层显式的生产配置校验。

### P3 - 低

#### P3.1 Roadmap 和 README 与当前 AI 问答实现存在滞后

证据：
- `docs/zh/roadmap.md`
- `README_zh.md`

影响说明：
- 项目现在实际上已经有腾讯云回答生成、引用详情和流式问答，但部分文档还把它描述成“只返回 sources 的检索服务”或“尚未接入生成模型”。
- 这会影响后续发布判断、接入判断和自我维护效率。

建议：
- 把当前状态、API 行为和安全边界更新到文档里，确保文档和真实代码一致。

## 就绪性结论

### 功能完整度

如果目标是“内部 RAG 控制台 + 运维调试台”，当前实现已经相当完整：

- 文本 / 手动入库已经具备
- chunk 预览、reindex、retry、chunk 编辑已经具备
- 检索和 AI 回答已经具备
- 引用详情和调试信息已经具备
- Ollama 并发保护已经具备
- health / overview / worker 可观测性已经具备

### 现在能不能接业务前端

暂时不建议直接接。

在接业务前端之前，最低限度需要先解决这 5 件事：

1. 移除并轮换所有已提交的真实密钥。
2. 重构公开问答接口，默认不能把知识库原文和内部 metadata 暴露给业务调用方。
3. 把公共 health 和受保护诊断接口拆开，并移除公共路径上的付费 / 重型探活。
4. 关闭提交配置里的 debug / `pprof` 暴露。
5. 把非流式 API 全面改成请求级 context 传递。

### 解决这些问题之后

当 P0 / P1 问题收掉之后，这个服务就很适合进入第一版业务前端接入阶段，尤其适合单租户或复杂度较低的业务场景。

如果后面要进一步走多租户、细粒度授权、审计合规或更强的回答策略控制，还需要再做一轮专门的安全与产品边界加固。
