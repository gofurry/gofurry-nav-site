# GoFurry Nav Collector v0.6.5 代码审计报告

## Summary

本次审计面向后端正式进入 `/api/v2/nav` 前的最后一轮 collector 侧收口。审计重点不是继续扩展采集能力，而是确认现有 v2 observation、summary、trend、change event、light probe 能否作为稳定数据源被后端消费。

结论：没有发现 P0 级别问题；但在依赖安全、v2 派生查询成本、低频探测启动行为、仓库配置安全和遗留 HTTP 工具函数上存在需要在进入后端 v2 前处理的风险。

## Scope

- 项目：`gofurry-nav-collector`
- 类型：生产采集 worker / 低频网络观测服务
- 重点代码面：
  - HTTP / TLS / DNS / Ping 采集
  - v2 observation / latest / summary / trend / change event
  - light probe：RDAP、robots.txt、security.txt、page_assets、port_check、waf_canary
  - 配置、Redis、DB、工具函数
- 审计命令：
  - `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`
  - `rg` 静态检查关键网络、DB、配置、锁和启动路径

## Severity Overview

| Severity | Count | Meaning |
|---|---:|---|
| P0 | 0 | Critical |
| P1 | 1 | High |
| P2 | 3 | Medium |
| P3 | 2 | Low |

## Findings

### P0 - Critical

No findings.

### P1 - High

### P1-001: Go toolchain 与 `golang.org/x/net` 存在可达漏洞

- Severity: P1
- Category: Security / Reliability
- Location: `gofurry-nav-collector/go.mod:3`, `gofurry-nav-collector/go.mod:18`, `gofurry-nav-collector/common/util/http.go:153`
- Status: Fixed in v0.6.5
- Confidence: High

#### Problem

`govulncheck` 报告当前代码受 14 个漏洞影响，涉及 Go 标准库 `go1.26` 与 `golang.org/x/net v0.52.0`。其中 `golang.org/x/net/html` 的 HTML 解析漏洞被追踪到 `goquery.NewDocumentFromReader`；标准库 `net/http`、`crypto/tls`、`crypto/x509`、`net/url`、`net` 相关漏洞也存在可达调用路径。

#### Impact

collector 会主动访问和解析第三方站点响应、证书、HTML、DNS 结果。后端 v2 开始正式消费 collector v2 数据后，采集器会成为更关键的数据来源。继续使用有已知漏洞的工具链和依赖，会增加 HTML 解析 DoS、TLS/x509 验证异常、HTTP/2 连接异常、URL 解析边界问题等生产风险。

#### Evidence

`govulncheck` 关键结果：

- `golang.org/x/net@v0.52.0`，HTML 解析相关漏洞固定版本为 `v0.55.0`。
- Go 标准库 `go1.26`，多个 `net/http`、`crypto/tls`、`crypto/x509`、`net/url` 漏洞固定版本为 `go1.26.3` 或更高。
- 示例路径包含 `common/util/http.go:153`、`collector/lightprobe/service/lightProbeService.go:977`、`common/service/redisService.go:32`。

#### Recommendation

在进入后端 v2 前完成依赖安全升级：

- 将 Go toolchain 升级到 `1.26.3+`。
- 将 `golang.org/x/net` 升级到 `v0.55.0+`。
- 运行 `go mod tidy`，确认没有引入不必要依赖漂移。
- 重新运行 `govulncheck`，确保没有 collector 可达漏洞残留；如存在不可避免项，写入接受风险说明。

#### Suggested Change

```bash
go get golang.org/x/net@v0.55.0
go mod tidy
go test ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

同时在 CI 中增加可选 `govulncheck` 步骤，至少在发版前执行。

#### Verification

- `govulncheck` 不再报告 collector 可达漏洞，或所有残留项都有明确接受风险记录。
- `go test ./...`、`go vet ./...` 通过。
- 生产构建使用已升级的 Go toolchain。

### P2 - Medium

### P2-001: Trend / change event 派生查询缺少匹配索引，可能放大 observation 历史表压力

- Severity: P2
- Category: Performance / Reliability
- Location: `gofurry-nav-collector/collector/observation/summary.go:77`, `gofurry-nav-collector/collector/observation/summary.go:85`, `gofurry-nav-collector/collector/observation/dao.go:83`, `gofurry-nav-collector/collector/observation/dao.go:114`
- Status: Fixed in v0.6.5
- Confidence: High

#### Problem

`UpdateSummaryIfEnabled` 会同步触发 trend 与 change event 派生。两类派生都会按 `site_id + target + observed_at` 查询 observation 历史，但当前 observation 建表脚本只包含 `protocol/site_id`、`site_id/protocol`、`protocol/observed_at` 方向的索引，没有覆盖 `site_id + target + protocol + observed_at` 的查询形态。

#### Impact

当前数据量还可控，但后端 v2 正式上线后，summary、trend、change event 更可能被频繁依赖。随着 observation 历史增长，缺少 target 维度索引会增加 DB 扫描成本，严重时拖慢采集 worker 或造成数据库周期性压力。

#### Evidence

- `summary.go:77` 调用 `UpdateTrendIfEnabled`。
- `summary.go:85` 调用 `UpdateChangeEventsIfEnabled`。
- `dao.go:83` 与 `dao.go:114` 查询条件包含 `WHERE site_id = ? AND target = ? AND observed_at >= ?`。
- `sql/20260523_collector_v2_observation.sql` 当前未包含 `(site_id, target, protocol, observed_at DESC, id DESC)` 索引。

#### Recommendation

新增手动 SQL 索引脚本，使用 `CREATE INDEX CONCURRENTLY IF NOT EXISTS`，不要自动迁移生产库：

```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gfn_collector_observation_site_target_protocol_time_id
ON gfn_collector_observation (site_id, target, protocol, observed_at DESC, id DESC);
```

同时给 trend/change 派生增加更明确的查询预算或开关，避免每次 summary 更新都无条件同步做历史查询。

#### Suggested Change

- 新增 `sql/20260525_collector_observation_target_indexes.sql`。
- 为 trend/change 增加配置开关或最小间隔去抖，默认保持当前行为但允许生产按需关闭。
- 增加 DAO 查询构造测试，确保查询字段与索引方向一致。

#### Verification

- 测试库执行 `EXPLAIN ANALYZE`，确认按 `site_id + target + protocol + observed_at` 查询命中索引。
- `go test ./...` 通过。
- 生产执行索引脚本时确认不放在事务块中。

### P2-002: 部分低频探测启用后会在服务启动时立即执行，部署重启可能造成非预期探测峰值

- Severity: P2
- Category: Reliability / Safety
- Location: `gofurry-nav-collector/collector/lightprobe/service/lightProbeService.go:95`, `gofurry-nav-collector/collector/lightprobe/service/lightProbeService.go:105`, `gofurry-nav-collector/collector/lightprobe/service/lightProbeService.go:115`, `gofurry-nav-collector/collector/lightprobe/service/lightProbeService.go:125`, `gofurry-nav-collector/collector/lightprobe/service/lightProbeService.go:135`
- Status: Fixed in v0.6.5
- Confidence: High

#### Problem

RDAP、robots.txt、security.txt、page_assets、port_check 在配置启用后都会启动时立即异步跑一轮。WAF canary 已经有 `run_on_start` 控制，但其他 light probe 没有同等控制。

#### Impact

这些探测虽然低频且默认关闭，但一旦生产配置显式开启，服务部署、重启、回滚都可能触发全量探测。对几百个站点规模来说通常可承受，但这不是“低频计划任务”的直觉行为，容易在生产更新窗口造成突发网络请求。

#### Evidence

`InitLightProbeOnStart` 中存在以下启动即执行路径：

- `go RunRDAP()`
- `go RunRobots()`
- `go RunSecurityTXT()`
- `go RunPageAssets()`
- `go RunPortCheck()`

WAF canary 则已经通过 `cfg.LightProbe.WAFCanary.RunOnStart` 控制。

#### Recommendation

为所有 light probe 增加统一 `run_on_start` 配置，默认 `false`。保留定时调度，避免部署重启天然触发全量探测。

#### Suggested Change

```yaml
collector:
  v2:
    light_probe:
      rdap:
        enabled: false
        run_on_start: false
      robots:
        enabled: false
        run_on_start: false
      security_txt:
        enabled: false
        run_on_start: false
      page_assets:
        enabled: false
        run_on_start: false
      port_check:
        enabled: false
        run_on_start: false
```

#### Verification

- 单元测试覆盖 `run_on_start=false` 时只注册 cron，不立即调用 run。
- 手工启动 collector，确认 Redis run state 不出现对应 light probe 立即运行记录。
- 显式 `run_on_start=true` 时保留当前行为。

### P2-003: 本地 `conf/server.yaml` 包含实配密码和开启状态，需要安全示例配置避免误部署

- Severity: P2
- Category: Security / Maintainability
- Location: `gofurry-nav-collector/conf/server.yaml:14`, `gofurry-nav-collector/conf/server.yaml:21`, `gofurry-nav-collector/conf/server.yaml:43`, `gofurry-nav-collector/conf/server.yaml:70`, `gofurry-nav-collector/conf/server.yaml:77`, `gofurry-nav-collector/conf/server.yaml:102`
- Status: Fixed in v0.6.5
- Confidence: High

#### Problem

当前工作区本地 `conf/server.yaml` 包含测试库 Redis / DB 密码，并且开启了 v2、page_assets、port_check、waf_canary 等探测能力。该文件已被 `.gitignore` 忽略，没有被 Git 跟踪；但仓库缺少可提交的安全示例配置，容易让本地实配被复制成部署默认配置。

#### Impact

风险主要有两个：

- 配置复制或部署时误启用低频主动探测，导致生产行为与预期不一致。
- 如果未来误强制提交本地配置，测试环境凭据可能扩散到仓库历史中。

#### Evidence

`conf/server.yaml` 中存在：

- `db_password: "<测试密码>"`
- `redis_password: "<测试密码>"`
- `collector.v2.enabled: true`
- `page_assets.enabled: true`
- `port_check.enabled: true`
- `waf_canary.enabled: true`
- `waf_canary.canary_path: "/"`

#### Recommendation

保留真实本地配置为 ignored 文件，新增可提交的安全示例配置。示例配置必须使用占位符密码，并默认关闭所有主动 light probe。

#### Suggested Change

- 新增 `conf/server.example.yaml`，所有密码使用占位符。
- 保持 `conf/server.yaml` ignored，不纳入 Git。
- 如果历史凭据有复用风险，轮换测试库密码。

#### Verification

- `git grep` 不再出现真实密码。
- 新人复制 example 后需要显式填写凭据和显式开启主动探测。
- 本地测试文档说明如何创建 local 配置。

### P3 - Low

### P3-001: 遗留 HTTP 工具函数存在不安全默认和无界读取，虽然当前无调用点但容易成为未来踩坑点

- Severity: P3
- Category: Security / Reliability / Maintainability
- Location: `gofurry-nav-collector/common/util/http.go:26`, `gofurry-nav-collector/common/util/http.go:48`, `gofurry-nav-collector/common/util/http.go:153`, `gofurry-nav-collector/common/util/http.go:194`
- Status: Fixed in v0.6.5
- Confidence: Medium

#### Problem

`common/util/http.go` 中的遗留 helper 使用 `InsecureSkipVerify: true`，并在多个路径直接 `io.ReadAll(resp.Body)`。HTML parse helper 直接对响应体创建 goquery document。当前 `rg` 没有发现业务调用点，但这些函数仍在包内保留，并被 `govulncheck` 识别为可达符号路径。

#### Impact

如果后续开发误用这些 helper，可能绕过 TLS 校验、读取超大响应体、解析无界 HTML，造成安全和内存风险。它也会让漏洞扫描结果长期混杂，降低审计信号质量。

#### Evidence

- `http.go:26`：`InsecureSkipVerify: true`
- `http.go:48`、`:69`、`:111`、`:194`：`io.ReadAll(resp.Body)`
- `http.go:153`：`goquery.NewDocumentFromReader(resp.Body)`

#### Recommendation

优先删除无调用点 helper；如果保留，则改成安全版本：

- 默认验证 TLS。
- 必须传入 context。
- 使用 `io.LimitReader`。
- 明确返回代理解析错误，不静默降级。
- 增加单元测试覆盖大响应体限制。

#### Suggested Change

先用 `rg "GetByHttp|PostByHttp"` 确认无调用点，再删除该文件或将函数标记为内部废弃并重写。

#### Verification

- `go test ./...` 通过。
- `govulncheck` 不再因无业务调用 helper 报出误导性路径。

### P3-002: RDAP bootstrap 缓存锁覆盖网络 I/O，后续并发化时可能造成不必要阻塞

- Severity: P3
- Category: Concurrency / Maintainability
- Location: `gofurry-nav-collector/collector/lightprobe/service/lightProbeService.go:1650`
- Status: Fixed in v0.6.5
- Confidence: Medium

#### Problem

`rdapBootstrapServers` 在持有 `rdapBootstrapMu` 的情况下执行 IANA bootstrap HTTP 请求和 JSON 解析。当前 RDAP run 本身是低频串行路径，风险较低；但如果后续对 RDAP 做并发化或多个调用方共享该 helper，锁会覆盖外部网络 I/O，导致其他 goroutine 被同一个慢请求阻塞。

#### Impact

在 IANA bootstrap 网络抖动时，后续 RDAP 查询会被锁串行阻塞。虽然不影响主采集链路，但会降低 light probe 的可预测性。

#### Evidence

`rdapBootstrapServers` 先 `rdapBootstrapMu.Lock()`，随后创建请求并执行 `client.Do(req)`，最后才更新缓存并释放锁。

#### Recommendation

改成 double-check 或 `singleflight` 风格：

- 先短锁检查缓存。
- 缓存过期时释放锁执行网络请求。
- 请求完成后重新加锁写入缓存。
- 或用 `singleflight.Group` 合并并发刷新请求。

#### Suggested Change

不需要引入复杂依赖；当前可以用短锁 + 二次检查完成。

#### Verification

- 单元测试并发调用 `rdapBootstrapServers`，确认只产生一次刷新或不会长时间持锁。
- `go test -race ./collector/lightprobe/service` 通过。

## Recommended Fix Plan

`v0.6.5` 已按以下顺序完成修复：

1. 先升级 Go toolchain 与 `golang.org/x/net`，重新跑 `govulncheck`。
2. 处理仓库配置安全默认，避免后续测试和生产更新误用。
3. 为 observation trend/change 增加 target 维度索引脚本，并加查询预算/开关。
4. 为 light probe 增加统一 `run_on_start`，默认关闭。
5. 删除或加固遗留 HTTP helper。
6. 低优先级处理 RDAP bootstrap 锁范围。

## Verification Suggestions

```bash
cd gofurry-nav-collector
gofmt -l .
go test ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
git diff --check
```

补充建议：

- 对 observation 历史查询执行 `EXPLAIN ANALYZE`。
- 对 light probe 启动行为增加配置测试。
- 对配置文件执行 `git grep`，确认没有真实密码。

## Notes

- 本轮报告不要求立即改动生产配置。
- 本轮没有新增采集内容，也没有建议在 collector 中实现后端 `/api/v2/nav`。
- `port_check`、`waf_canary` 等主动轻探测仍应保持默认关闭、显式授权、低频执行。
