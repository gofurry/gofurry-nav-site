# GoFurry Nav Backend 代码审计报告

## Summary

本次审计范围是 `gofurry-nav-backend` 的公开导航接口、`/api/v2/nav` read model/detail/summary 接口、Redis/DB/config 基础设施、WAF/CORS/限流边界，以及即将接入 Nuxt 前端时会进入用户主路径的代码。

总体结论：未发现 P0 级别问题。审计发现的 P1/P2/P3 项已在 `v0.4.2` 修复：外部搜索建议接口已增加请求预算和失败降级，v2 raw payload 完整输出已改为默认关闭，代理 IP、Redis/DB、v1 Ping 历史脏数据、配置日志、HTTP 工具和 JWT 工具均已做安全收敛。

## Scope

- `routers/router.go`
- `routers/url.go`
- `routers/url_v2.go`
- `middleware/corazaWAF.go`
- `apps/nav/navPage`
- `apps/nav/sitePage`
- `apps/nav/summary`
- `apps/nav/readmodel`
- `apps/nav/detail`
- `common/service/redisService.go`
- `common/util`
- `roof/env`
- `roof/db`

## Severity Overview

| Severity | Count | Meaning |
|---|---:|---|
| P0 | 0 | Critical |
| P1 | 2 | High |
| P2 | 4 | Medium |
| P3 | 3 | Low |

## Findings

### P0 - Critical

No findings.

### P1 - High

#### P1-001: 公开搜索建议接口缺少外部 HTTP 请求预算和响应体上限

- Severity: P1
- Category: Reliability / Performance / Security
- Location: `apps/nav/navPage/service/navPageService.go:116`, `apps/nav/navPage/service/navPageService.go:122`, `apps/nav/navPage/service/navPageService.go:172`, `apps/nav/navPage/service/navPageService.go:178`, `apps/nav/navPage/service/navPageService.go:233`, `apps/nav/navPage/service/navPageService.go:240`, `apps/nav/navPage/service/navPageService.go:246`, `apps/nav/navPage/service/navPageService.go:275`, `apps/nav/navPage/service/navPageService.go:281`
- Status: Fixed in v0.4.2
- Confidence: High

##### Problem

百度、必应、谷歌、B 站搜索建议接口会在公开请求路径中直接访问外部服务。当前实现中 `http.Get` 和手写 `http.Client` 没有统一 `Timeout`，响应体使用 `ReadAll` 读取且没有大小上限，`q` 参数也没有长度限制和统一 `url.QueryEscape`。

##### Impact

外部搜索服务变慢、连接悬挂、返回异常大响应，或用户传入超长 `q` 时，后端 goroutine、连接和内存会被放大占用。当前接口会跟随前端搜索框频繁调用，因此在正式接入前端后容易成为尾延迟和可用性风险点。

##### Evidence

代码中多处直接使用 `http.Get(url)`，随后 `ioutil.ReadAll(resp.Body)`。谷歌搜索建议虽然创建了 `http.Client`，但没有设置 `Timeout`。

##### Recommendation

新增搜索建议专用 HTTP helper：统一 `context`/`Timeout`、最大响应体字节数、最大 `q` 长度、`url.QueryEscape`、外部错误降级为空建议列表。建议默认超时 2 秒到 3 秒，响应体上限 64 KiB。

##### Suggested Change

- 为 `q` 增加 `strings.TrimSpace`、最大长度和空值快速返回。
- 使用 `http.Client{Timeout: ...}` 或 `http.NewRequestWithContext`。
- 用 `io.LimitReader` 代替直接 `ReadAll`。
- 给四个搜索建议接口补 controller/service 单元测试，覆盖超长输入、外部超时、坏 JSON/XML、超大响应。

##### Verification

- 使用 `httptest.Server` 模拟慢响应和超大响应。
- 运行 `go test ./apps/nav/navPage/...`。
- 手工确认外部搜索服务不可用时页面仍可正常展示，只是建议列表为空。

#### P1-002: `payload_mode=full` 允许公开接口绕过 payload 预览限制

- Severity: P1
- Category: Performance / Security
- Location: `apps/nav/detail/service/detailService.go:589`, `apps/nav/detail/service/detailService.go:621`, `apps/nav/readmodel/models/readmodel.go:76`, `apps/nav/readmodel/models/readmodel.go:79`
- Status: Fixed in v0.4.2
- Confidence: High

##### Problem

v2 detail/latest/observations/light-probes 默认会按 `raw_payload_preview_bytes` 截断 payload，但只要请求带 `payload_mode=full`，就会完整返回 collector payload。observation 历史单次最多 500 条，而 collector 单条 v2 payload 已有 512 KiB 级别保护，理论上公开请求仍可触发很大的响应体。

##### Impact

前端迁移后，如果公开页面或恶意请求使用 `payload_mode=full&limit=500`，后端可能从 DB/Redis 读取大量 JSON 并完整序列化返回，造成内存、带宽和延迟压力。这个问题不改变数据权限，但会影响生产可用性。

##### Evidence

`normalizePayloadMode` 接受 `full`；`applyPayloadPolicy` 在 `mode == PayloadModeFull` 时直接返回完整 payload。

##### Recommendation

将公开接口默认只允许 preview。`payload_mode=full` 应通过配置显式开启，或仅允许内网/管理端使用，并增加总响应预算。前端接入站点详情页时不应依赖 full payload。

##### Suggested Change

- 新增配置，例如 `nav_v2.full_payload_enabled: false`。
- `payload_mode=full` 未开启时返回参数非法或自动降级为 preview。
- observations 增加总 payload 预算，例如单响应最多 2 MiB 到 5 MiB。
- 增加测试：未开启 full 时 `payload_mode=full` 不会返回完整 payload。

##### Verification

- 构造 500 条大 payload observation，确认默认响应被截断且总大小受控。
- 确认 Nuxt 详情页只使用 preview/detail fields，不请求 full。

### P2 - Medium

#### P2-001: 代理 IP 信任边界过宽，可能影响限流和浏览量去重

- Severity: P2
- Category: Security / Reliability
- Location: `routers/router.go:49`, `routers/router.go:105`, `common/util/function.go:54`
- Status: Fixed in v0.4.2
- Confidence: Medium

##### Problem

Fiber 配置中启用了 `TrustProxy: true`，限流 key 使用 `c.IP()`，浏览量去重又手动信任 `X-Forwarded-For` 和 `X-Real-IP`。如果部署层没有严格覆盖这些请求头，客户端可以伪造代理 IP 相关 header。

##### Impact

攻击者可能通过伪造 IP header 绕过按 IP 的限流，也可能污染站点浏览量去重。风险取决于 Nginx/CDN 是否清洗或重写了代理 header，因此置信度为 Medium。

##### Evidence

`router.go` 全局 `TrustProxy: true`，limiter 的 `KeyGenerator` 使用 `c.IP()`。`GetClientIP` 会优先返回请求头中的 `X-Forwarded-For` 或 `X-Real-IP`。

##### Recommendation

将可信代理来源配置化，只有来自受信任反代的请求才读取 forwarded header。应用层至少需要提供开关或 trusted proxy CIDR；部署层应明确清洗客户端传入的代理头。

##### Suggested Change

- 新增 `server.trusted_proxy_headers` 或 `server.trusted_proxy_cidrs` 配置。
- `GetClientIP` 仅在远端地址匹配可信代理时读取 `X-Forwarded-For`。
- 限流 key 与浏览量 IP 来源共用同一个 helper。
- 文档中写清 Nginx 需要覆盖 `X-Forwarded-For` 和 `X-Real-IP`。

##### Verification

- 添加单元测试：非可信来源携带伪造 XFF 时仍使用真实远端地址。
- 手工用 curl 带不同 XFF 请求，确认限流和浏览量不会被绕过。

#### P2-002: Redis wrapper 缺少统一命令超时，部分错误会被吞掉

- Severity: P2
- Category: Reliability / Correctness
- Location: `common/service/redisService.go:23`, `common/service/redisService.go:85`, `common/service/redisService.go:156`, `common/service/redisService.go:169`
- Status: Fixed in v0.4.2
- Confidence: High

##### Problem

Redis wrapper 使用包级 `context.Background()`。多数命令没有每次调用的 timeout context。`HDel` 在 Redis 错误时返回 `intVal, nil`，`Incr` 完全忽略错误。

##### Impact

Redis 慢、网络抖动或连接池异常时，HTTP handler 和定时任务可能被拖慢；删除或计数失败也不容易被上层感知。后端 v2 会更依赖 Redis latest/summary/trend/change，这个边界应在前端接入前收紧。

##### Evidence

`var ctx = context.Background()` 被所有命令复用；`HDel` 的 `err != nil` 分支返回 nil 错误；`Incr` 没有返回值。

##### Recommendation

为 Redis wrapper 增加内部 `withRedisTimeout`，默认 2 秒。修复 `HDel` 和 `Incr` 的错误返回，保持调用方兼容时可以先记录日志，再逐步改签名。

##### Suggested Change

- 配置新增或复用 `redis_timeout_seconds`。
- `GetString`、`HGet`、`HGetAll`、`Set`、`SetNX`、`Scan` 等命令都使用带 timeout 的 context。
- `HDel` 错误时返回 `GFError`。
- `Incr` 改为返回 `GFError`，调用方记录失败。

##### Verification

- 使用 mock Redis 或注入 fake client 测试 timeout/error 返回。
- 手工断开 Redis，确认接口快速失败或返回 missing，而不是长时间挂住。

#### P2-003: DB 初始化忽略 `DB()` 错误，DSN 构造对特殊字符不够稳健

- Severity: P2
- Category: Reliability / Maintainability
- Location: `roof/db/db.go:40`, `roof/db/db.go:47`
- Status: Fixed in v0.4.2
- Confidence: Medium

##### Problem

DB 初始化通过 `fmt.Sprintf` 拼接 PostgreSQL DSN，随后忽略 `db.engine.DB()` 返回的错误。当前生产密码可用不代表后续配置都安全，密码中如果出现空格、反斜杠或其他特殊字符，字符串 DSN 可能解析异常。

##### Impact

配置变更后可能出现难定位的启动失败。忽略 `DB()` 错误也可能让后续连接池配置在异常状态下 panic 或产生误导日志。

##### Evidence

`dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", ...)`，`sqlDB, _ := db.engine.DB()`。

##### Recommendation

改用 URL DSN 或对 key/value DSN 做安全转义，并显式处理 `db.engine.DB()` 错误。日志不要输出完整 DSN 或密码。

##### Suggested Change

- 用 `net/url.URL` 构造 `postgres://user:pass@host:port/db?sslmode=disable`。
- `sqlDB, err := db.engine.DB()` 后检查并返回启动错误。
- 增加包含特殊字符密码的 DSN 构造单元测试。

##### Verification

- 使用带空格、`@`、`:`、`\` 的密码运行 DSN 构造测试。
- 确认错误日志不包含密码明文。

#### P2-004: v1 Ping 历史接口对脏数据缺少防护，可能 panic 或产生错误均值

- Severity: P2
- Category: Correctness / Reliability
- Location: `apps/nav/sitePage/service/sitePageService.go:104`, `apps/nav/sitePage/service/sitePageService.go:105`, `apps/nav/sitePage/service/sitePageService.go:119`
- Status: Fixed in v0.4.2
- Confidence: High

##### Problem

v1 Ping 历史接口假设 `Delay` 字段一定以 `ms` 结尾并且长度至少为 2，然后直接切片 `v.Delay[:len(v.Delay)-2]`。同时平均值固定除以 20/60/100，即使实际记录不足，也会输出偏低的平均值。

##### Impact

如果历史表中存在空字符串、异常格式或记录不足，接口可能 panic，或者返回明显不可信的平均延迟和丢包。虽然 recover 中间件会兜底 500，但前端详情页会受到影响。

##### Evidence

`temp.Delay, _ = util.String2Int(v.Delay[:len(v.Delay)-2])` 未检查长度；均值使用固定分母。

##### Recommendation

复用已有 `util.ExtractSuffix2Int` 或新增安全解析 helper，无法解析时跳过或按 0 标记。平均值按实际参与统计数量计算。

##### Suggested Change

- 对 `delay`、`loss` 做安全解析。
- `Twenty/Sixty/Hundred` 分别记录实际统计数量，避免固定分母。
- 增加异常 delay、空 delay、记录不足的单元测试。

##### Verification

- 构造 `delay=""`、`delay="bad"`、`delay="12ms"` 的测试数据。
- 确认接口不 panic，均值符合实际记录。

### P3 - Low

#### P3-001: 配置路径探测日志默认刷屏

- Severity: P3
- Category: Maintainability / Observability
- Location: `roof/env/config.go:201`, `roof/env/config.go:211`, `roof/env/config.go:226`, `roof/env/config.go:232`, `roof/env/config.go:243`
- Status: Fixed in v0.4.2
- Confidence: High

##### Problem

配置加载阶段会直接 `fmt.Println` 输出探测路径和加载路径。命令行、测试、systemd 日志中会出现大量路径追踪。

##### Impact

日志可读性下降，也会暴露部署目录结构。它不是直接安全漏洞，但会干扰生产排障。

##### Recommendation

默认关闭路径追踪，只在 `GF_NAV_BACKEND_CONFIG_TRACE=1` 时输出。找不到配置时保留明确错误。

##### Verification

- 未设置环境变量时运行测试，不出现路径刷屏。
- 设置 `GF_NAV_BACKEND_CONFIG_TRACE=1` 时可看到路径追踪。

#### P3-002: 通用 HTTP 工具保留 `InsecureSkipVerify` 和无上限读取的危险默认

- Severity: P3
- Category: Security / Maintainability
- Location: `common/util/http.go:25`, `common/util/http.go:36`, `common/util/http.go:90`, `common/util/http.go:117`, `common/util/http.go:142`, `common/util/http.go:193`
- Status: Fixed in v0.4.2
- Confidence: Medium

##### Problem

`common/util/http.go` 中多个 helper 使用 `http.Get` 或 `io.ReadAll`，部分 helper 设置 `TLSClientConfig: &tls.Config{InsecureSkipVerify: true}`。当前审计没有发现这些 helper 被业务代码调用，但它们作为公共工具存在危险默认。

##### Impact

后续代码如果复用这些 helper，可能默认跳过证书校验或读取无限响应体，引入安全和稳定性问题。

##### Recommendation

将 helper 改为安全默认：不跳过 TLS 校验、必须传 timeout、响应体上限可配置；确实需要跳过 TLS 时使用显式命名函数并写明用途。

##### Verification

- `rg` 确认新代码不再调用旧危险 helper。
- 为安全 HTTP helper 添加 timeout、body limit、TLS 默认校验测试。

#### P3-003: JWT 工具存在硬编码密钥和空 token panic 风险，未来启用认证前必须处理

- Severity: P3
- Category: Security / Correctness
- Location: `common/constant.go:51`, `common/util/function.go:189`
- Status: Fixed in v0.4.2
- Confidence: Medium

##### Problem

JWT 密钥写在代码常量中，`ParseToken` 在 `jwt.ParseWithClaims` 返回错误且 token 为 nil 的情况下仍访问 `token.Claims`。当前导航公开接口没有使用这套认证工具，因此本轮不升为 P1/P2。

##### Impact

如果后续复用这套工具给后台或管理接口加认证，会存在密钥不可轮换和异常 token 触发 panic 的风险。

##### Recommendation

认证功能正式使用前，将 JWT secret 移入配置或环境变量，并在 `ParseToken` 中先判断 `err` 和 `token == nil`。

##### Verification

- 增加 invalid token 测试，确认不会 panic。
- 确认生产配置中 JWT secret 不进入 Git。

## Recommended Fix Plan

1. `v0.4.2` 已修复 P1：搜索建议外部请求预算、公开 full payload 输出边界。
2. `v0.4.2` 已修复 P2：可信代理 IP、Redis timeout/错误返回、DB 初始化、v1 Ping 脏数据防护。
3. `v0.4.2` 已修复 P3：配置路径追踪开关、HTTP 工具安全默认、JWT 工具防御性修复。
4. `v0.4.2` 额外将 v1 nav/sitePage DAO 从包级 init 改为懒加载，避免纯单元测试必须连接真实数据库。

## Verification Suggestions

```bash
go test ./...
go vet ./...
go test -race ./apps/nav/detail/... ./apps/nav/readmodel/... ./common/service/...
git diff --check
```

建议额外手工验证：

- 外部搜索建议接口在上游超时或不可用时快速返回空列表。
- Nuxt 详情页不使用 `payload_mode=full`。
- 伪造 `X-Forwarded-For` 不会绕过限流或浏览量去重。
- Redis 断连时 v2 summary/detail 接口不会长时间挂起。

## Notes

- 本次审计没有发现 SQL 注入：当前主要查询路径使用 GORM 参数绑定。
- 本次审计没有发现后端主动执行外部命令。
- v2 detail 已保持只读，不递增浏览量；v1 detail 仍保留历史浏览量副作用。
