# GoFurry Ops Agent / Center 代码审计报告

## Summary

审计日期：2026-05-19

本次审计覆盖 `ops/gofurry-ops-agent` 与 `ops/gofurry-ops-center` 的 Go 代码与关键配置样例，重点看安全边界、认证与签名、采集器、离线 spool、HTTP API、PostgreSQL 入库、监控指标查询、后台循环和可维护性。

整体结论：首版架构清晰，Agent 与 Center 的边界也比较克制；HMAC 上报、时间窗口校验、cookie 登录、SQL 参数化、collector timeout、spool 权限等基础安全面已经具备。当前没有发现 P0 级问题。需要优先处理的是公网或半公网部署下的登录/写入入口防护、告警阈值未生效、监控 raw sample 查询索引不足，以及 Agent/Center 在数据量增长后的边界控制。

已执行验证：

```powershell
cd ops/gofurry-ops-agent; go test ./...; go vet ./...
cd ops/gofurry-ops-center; go test ./...; go vet ./...
```

结果：全部通过。

## Scope

- 项目类型：Agent CLI / 采集 worker；Center API service / 内置控制台。
- 运行上下文：轻量生产运维组件，可能部署在公网反代后，也可能只在内网使用。
- 关键面：HTTP handlers、管理员认证、Agent HMAC 上报、PostgreSQL 写入、后台清理、Peer 拉取、collector 网络调用、spool 文件操作。
- 审计范围：`ops/gofurry-ops-agent` 与 `ops/gofurry-ops-center` 当前 Go 实现；前端仅从 API 暴露数据和风险面角度轻量关注。
- 报告目标：用户指定的 `ops/code-audit.md`。

## Severity Overview

| Severity | Count | Meaning |
|---|---:|---|
| P0 | 0 | Critical |
| P1 | 1 | High |
| P2 | 3 | Medium |
| P3 | 4 | Low |

## Findings

### P0 - Critical

No findings.

### P1 - High

#### P1-001: 管理登录和写入型入口缺少速率限制，且生产密钥强度没有被配置校验兜底

- Severity: P1
- Category: Security / Reliability
- Location: `ops/gofurry-ops-center/internal/httpapi/auth.go:23`, `ops/gofurry-ops-center/internal/httpapi/auth.go:32`, `ops/gofurry-ops-center/internal/httpapi/server.go:55`, `ops/gofurry-ops-center/internal/httpapi/server.go:63`, `ops/gofurry-ops-center/internal/config/config.go:197`
- Status: Open
- Confidence: High

##### Problem

`/api/v1/admin/auth/login` 直接校验 passcode，但没有 IP / 用户维度的失败计数、限速、退避或短期锁定。Agent ingest、sync event、deploy event 也是写入型入口，目前依赖 token/HMAC/session 鉴权，但没有额外的请求速率限制或写入频率保护。

配置校验只检查 `dashboard_passcode`、`session_secret`、`agent_tokens` 非空，没有最小长度、随机性或常见占位值拦截。示例配置明确提醒替换 `change-me-*`，但程序本身不会拒绝这类弱值。

##### Impact

如果 Center 被暴露到公网或弱内网，登录口可被持续撞库；一旦 passcode 或 agent token 使用弱值，攻击者可以拿到控制台访问权或通过合法签名路径向 raw sample 表灌入大量数据，造成告警噪音、存储膨胀和面板失真。

##### Evidence

`authLogin` 在 `auth.go:23` 解析请求，`auth.go:32` 用普通字符串比较 passcode；路由在 `server.go:55` 和 `server.go:63` 挂载时没有 limiter 中间件；配置校验在 `config.go:197` 起仅验证非空。

##### Recommendation

1. 在登录入口加 Fiber limiter，例如按 IP 每分钟 5 次失败，失败后指数退避或锁定 5 分钟。
2. 对 `/api/v1/agent/ingest`、`/api/v1/events/sync`、`/api/v1/events/deploy` 加单独 limiter，Agent 可按 `X-GoFurry-Node-ID` + IP 分桶。
3. 配置校验拒绝 `change-me*`、`test`、`password` 等占位/弱值；建议 `dashboard_passcode >= 12`，`session_secret` 和 token 至少 32 字节随机值。
4. passcode 比较可改成固定时间比较，减少不必要的旁路信息。

##### Suggested Change

在 `httpapi.New` 中为登录和写入型 API 分别挂载 limiter；在 `config.Validate` 中新增 `validateSecret(name, value, minLen)`，集中拒绝短值和占位值。

##### Verification

增加 handler 测试覆盖连续失败登录返回 `429`；增加配置测试覆盖 `change-me-*`、过短 secret、过短 token 被拒绝。

### P2 - Medium

#### P2-001: `http_failure_threshold` 配置目前没有参与 HTTP 告警决策

- Severity: P2
- Category: Correctness / Reliability
- Location: `ops/gofurry-ops-center/internal/service/service.go:55`, `ops/gofurry-ops-center/internal/repository/repository.go:60`, `ops/gofurry-ops-center/internal/service/service.go:361`
- Status: Open
- Confidence: High

##### Problem

`Alert.HTTPFailureThreshold` 被传入 `Store.Ingest`，`Repository.Ingest` 也接收了 `alertThreshold int`，但该参数没有被使用。HTTP 检查只要单次 `status != "ok"`，`evaluatePayloadAlerts` 就立刻创建 critical 告警。

##### Impact

配置里的“连续失败 N 次再告警”语义不会生效。短暂网络抖动、单次 502 或探测超时都会触发 critical 告警，降低面板可信度，也会让后续 notifier 扩展更容易产生误报。

##### Evidence

`service.go:55` 将 `s.cfg.Alert.HTTPFailureThreshold` 传给 store；`repository.go:60` 的 `alertThreshold` 参数未参与任何计算；`service.go:361` 到 `service.go:375` 对 HTTP 失败立即 `UpsertAlert`。

##### Recommendation

把 service status 的 `failure_count` 用于告警判定：HTTP / Postgres / Redis 等服务检查失败时先更新 `service_status.failure_count`，只有 `failure_count >= http_failure_threshold` 时才创建或刷新告警；成功时重置 failure count 并 resolve。

##### Suggested Change

让 repository 返回每个服务检查的最新 failure count，或把 alert evaluation 放入 repository transaction 内，确保 sample、service_status、alert_state 的状态一致。

##### Verification

增加测试：连续失败 1 次不告警，达到阈值才告警；成功样本会清零 failure count 并 resolve 旧告警。

#### P2-002: raw sample 表缺少关键索引，监控面板 30 秒刷新后会放大查询和清理成本

- Severity: P2
- Category: Performance / Reliability
- Location: `ops/gofurry-ops-center/internal/repository/schema.go:55`, `ops/gofurry-ops-center/internal/repository/schema.go:67`, `ops/gofurry-ops-center/internal/repository/schema.go:80`, `ops/gofurry-ops-center/internal/repository/schema.go:93`, `ops/gofurry-ops-center/internal/repository/schema.go:109`, `ops/gofurry-ops-center/internal/repository/metrics.go:274`, `ops/gofurry-ops-center/internal/repository/metrics.go:595`, `ops/gofurry-ops-center/internal/repository/repository.go:443`
- Status: Open
- Confidence: High

##### Problem

`network_samples`、`docker_container_samples`、`http_check_results`、`service_check_results`、`cert_check_results` 没有 `(node_id, reported_at)`、`reported_at` 或 `received_at` 索引。新增的 dashboard metrics API 会按 range 查询趋势和 latest 数据；后台清理每日按 `received_at < $1` 删除旧数据。

##### Impact

当 Agent 数量或采样天数上来后，30 秒轮询的 Overview / Node Metrics 会频繁扫 raw sample 表；清理任务也可能做全表扫描和大量删除，导致 Center 响应变慢，甚至影响 ingest 写入延迟。

##### Evidence

schema 里只有 heartbeats、system、disk 的 node/reported 索引；`metrics.go:274` 查询 HTTP/service latency，`metrics.go:595` 查询网络趋势，`repository.go:443` 起遍历 raw sample 表按 `received_at` 清理。

##### Recommendation

为所有 raw sample 表补齐索引：

```sql
CREATE INDEX IF NOT EXISTS idx_network_samples_node_reported ON network_samples(node_id, reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_network_samples_received ON network_samples(received_at);
CREATE INDEX IF NOT EXISTS idx_http_check_results_node_reported ON http_check_results(node_id, reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_http_check_results_reported ON http_check_results(reported_at);
CREATE INDEX IF NOT EXISTS idx_http_check_results_received ON http_check_results(received_at);
```

同类索引也应补到 docker、service、cert 表。Overview 的 latest 查询可考虑维护轻量 latest 表或物化最新状态，避免每次窗口函数扫历史样本。

##### Suggested Change

短期在内置 schema 中补 `CREATE INDEX IF NOT EXISTS`；中期引入版本化 migration，确保已有部署能自动补索引。

##### Verification

在测试 schema 中灌入多节点、多天样本，使用 `EXPLAIN ANALYZE` 比较 metrics API 的查询计划；确认 dashboard 自动刷新期间 ingest latency 没明显抖动。

#### P2-003: Agent ingest 对 payload 大小、数组数量和字符串长度缺少显式业务边界

- Severity: P2
- Category: Security / Reliability
- Location: `ops/gofurry-ops-center/internal/httpapi/server.go:32`, `ops/gofurry-ops-center/internal/httpapi/server.go:92`, `ops/gofurry-ops-center/internal/repository/repository.go:60`
- Status: Open
- Confidence: Medium

##### Problem

Center 使用 `c.Body()` 获取完整请求体并直接反序列化 Agent payload，之后 repository 会遍历 payload 中的 disks、networks、docker、http、postgres、redis、certs 并逐条写入数据库。当前没有显式设置 Center 业务层 body limit，也没有限制每类数组数量、字段长度或错误消息长度。

##### Impact

正常 Agent 不会产生超大 payload，但如果 agent token 泄露、Agent 配置被误写，或某台机器暴露异常多的网络接口 / 容器 / 检查项，单次 ingest 可能造成大量 DB 写入、事务变长、响应超时和 raw sample 膨胀。

##### Evidence

Fiber 初始化在 `server.go:32` 未设置业务 body 上限；`server.go:92` 读取完整 body；`Repository.Ingest` 从 `repository.go:60` 起逐类循环写入 payload 内容。

##### Recommendation

1. 在 Fiber config 设置明确 `BodyLimit`，例如 1-2 MB，并在 README 中说明。
2. 在 `Service.Ingest` 中做业务校验：每类数组数量上限、`node_id/name/url/error_message` 长度上限、状态枚举白名单。
3. 对异常 payload 返回 400，并记录精简日志，不把长错误原文完整入库。

##### Suggested Change

新增 `validatePayload(payload)`，例如限制 `disks <= 64`、`networks <= 128`、`docker <= 256`、每类 service checks <= 128、字符串字段 <= 512 或 URL <= 2048。

##### Verification

增加 handler/service 测试覆盖超大数组、超长字符串、未知状态值被拒绝；压测单次 ingest 确认失败请求不会打开数据库事务。

### P3 - Low

#### P3-001: Session token 只有过期时间和签名，缺少随机 nonce 与服务端撤销能力

- Severity: P3
- Category: Security / Maintainability
- Location: `ops/gofurry-ops-center/internal/security/hmac.go:32`
- Status: Open
- Confidence: High

##### Problem

`NewSession` 的 payload 只有 expires 秒级时间戳，同一 secret 和同一秒内生成的 session token 完全相同。logout 只清除浏览器 cookie，不会让已泄露 token 在服务端失效。

##### Impact

这不是直接绕过认证的问题，因为攻击者仍需要 `session_secret` 或现有 token。但如果 token 被日志、代理或浏览器环境泄露，它会在 TTL 内持续可用；同秒 token 相同也不利于后续审计和会话级风控。

##### Evidence

`hmac.go:33` 只写入 expires；`hmac.go:35` 到 `hmac.go:38` 对 expires 做 HMAC 后返回。

##### Recommendation

在 session payload 中加入 128-bit 随机 nonce、issued_at、expires；如仍想保持无状态，可以只签名验证；如果要支持 logout 撤销，则增加短 TTL + Redis / DB denylist 或 session store。

##### Verification

测试同一秒连续登录生成不同 token；logout 后旧 token 再访问 `/api/v1/dashboard/overview` 被拒绝。

#### P3-002: Agent Postgres collector 每次采样新建 pgxpool，轻量场景可用但会带来连接抖动

- Severity: P3
- Category: Performance / Reliability
- Location: `ops/gofurry-ops-agent/internal/collector/postgres.go:18`
- Status: Open
- Confidence: Medium

##### Problem

`collectPostgres` 每次采样都会 `pgxpool.New`，完成 ping 和两条元数据查询后 `pool.Close()`。单 Agent、低频采样时影响较小；多 Agent 或更短采样间隔下，会给 PostgreSQL 带来重复建连和认证成本。

##### Impact

在资源较小的数据库上，连接 churn 会让健康检查本身成为噪音，表现为偶发延迟升高或连接数抖动。

##### Evidence

`postgres.go:18` 新建 pool，`postgres.go:24` defer close，采样周期由 runtime 每次触发。

##### Recommendation

如果只做一次性健康检查，优先使用 `pgx.Connect` 单连接；如果希望复用连接，则把 pool 生命周期提升到 Runtime，并设置 `MaxConns=1`、`MinConns=0`、明确 idle timeout。

##### Verification

在本地 PostgreSQL 打开连接日志或查询 `pg_stat_activity`，比较修复前后每 30 秒采样时的连接波动。

#### P3-003: TLS 证书采集的 context 只在连接后检查，拨号和握手不能被父 context 及时取消

- Severity: P3
- Category: Reliability / Correctness
- Location: `ops/gofurry-ops-agent/internal/collector/cert.go:17`, `ops/gofurry-ops-agent/internal/collector/cert.go:20`
- Status: Open
- Confidence: High

##### Problem

`collectCert` 创建了 `checkCtx`，但实际拨号使用 `tls.DialWithDialer`，它没有接收该 context；代码只在连接成功后 `select` 检查一次 `checkCtx.Done()`。

##### Impact

当 Agent 正在退出或上层采集 context 被取消时，证书检查仍可能等到 dialer timeout 或 TLS handshake 结束才返回。当前 timeout 默认 5 秒，影响不大，但会让 shutdown 和 once timeout 不够干净。

##### Evidence

`cert.go:17` 创建 `checkCtx`；`cert.go:20` 调用 `tls.DialWithDialer` 时未传入 context。

##### Recommendation

使用 `tls.Dialer{NetDialer: dialer, Config: tlsCfg}.DialContext(checkCtx, "tcp", addr)`，确保拨号和 TLS 握手都受 context 控制。

##### Verification

增加测试或手工使用不可达地址，取消父 context 后确认 `collectCert` 快速返回 `timeout` / `context canceled`。

#### P3-004: spool replay 使用 Scanner 的 1 MB 单行上限，异常大 payload 会阻塞该文件后续回放

- Severity: P3
- Category: Reliability
- Location: `ops/gofurry-ops-agent/internal/spool/spool.go:71`
- Status: Open
- Confidence: Medium

##### Problem

spool 文件按 JSONL 存储，replay 使用 `bufio.Scanner`，并将单行最大值设置为 1 MB。正常 payload 通常远小于这个值，但如果某次离线上报包含大量采集项并超过上限，Scanner 会返回 `token too long`，该文件不会被删除或隔离，后续每轮 replay 都会继续失败。

##### Impact

离线缓存里一条异常记录可能长期卡住回放流程，让后续有效样本不能按预期补报，或者每轮都记录 replay failed。

##### Evidence

`spool.go:71` 创建 scanner；`spool.go:72` 将 buffer 上限设为 `1024*1024`。

##### Recommendation

将 spool 单条上限与 Center ingest body limit 对齐；对超过上限或无法解析的行做 quarantine 文件移动，例如 `.bad`，并继续处理后续文件。

##### Verification

增加测试：构造超过 1 MB 的 JSONL 行，确认 replay 不会永久阻塞后续正常文件。

## Recommended Fix Plan

1. 优先修 P1：增加 limiter、失败退避、secret/token 强度校验，并补测试。
2. 修 P2-001：让 `http_failure_threshold` 真正参与告警创建，减少误报。
3. 修 P2-002：补 raw sample 查询与清理索引；若已有部署，建议同时引入最小版本化 migration。
4. 修 P2-003：为 ingest 增加 body limit、payload cardinality 和字段长度校验。
5. 逐步处理 P3：session nonce、collector 连接复用/单连接、证书检查 context、spool oversized quarantine。

## Verification Suggestions

建议在修复后执行：

```powershell
cd ops/gofurry-ops-agent
go test ./...
go test -race ./...
go vet ./...

cd ..\gofurry-ops-center
go test ./...
go test -race ./...
go vet ./...
```

建议新增专项验证：

```powershell
# Center handler/security
go test ./internal/httpapi -run "Test.*Login|Test.*Ingest|Test.*Limit" -count=1

# Center alert threshold
go test ./internal/service -run "Test.*HTTP.*Threshold" -count=1

# Center metrics/index smoke with local isolated schema
go test ./internal/repository -run "Test.*Metrics" -count=1

# Agent spool and collectors
go test ./internal/spool ./internal/collector -count=1
```

## Notes

- 本次没有发现 SQL 注入：当前数据库访问使用 pgx 参数化查询；`CleanupRawSamples` 中的表名来自固定白名单切片，不接受外部输入。
- Agent 的 Postgres / Redis collector 默认只做只读健康检查，没有写业务数据。
- Center 的 HMAC 上报校验包含 Bearer token、node id、timestamp、body 签名和时间窗口，基础设计是合理的；当前建议主要是边界硬化和生产抗压。
