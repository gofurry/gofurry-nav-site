# #92-A — HTTP/2 Probe Reliability & Metric V2 验收

日期：2026-10-18。现有本地 `dev`，开始时工作区干净，本地/远端基线均为
`f4bf7feb790b6ec8131be78f1d010281c6dd9cf3`。本记录与实现一起提交；
证据来自本次工作树的本地测试，不代表已部署生产或远端 CI 已通过。

## 范围与合同

- 生产 `performRequest` 的 Transport 启用 `ForceAttemptHTTP2`，删除静态
  `Connection: Keep-Alive`；代理、预算、重定向、响应体限制和独立证书诊断保持原逻辑。
- 成功响应新增 `http2_probe_profile=go_h2_enabled_v1`、`final_scheme`、
  `tls_alpn`，与原 `http_protocol` / `final_url` 一同进入 HTTP Known State。
  ALPN 来自最终 `resp.TLS`，scheme 来自最终 `resp.Request.URL`；空 ALPN 与缺失不同。
- profile 还要求请求完成后 Transport 实际注册 H2 handler：
  `GODEBUG=http2client=0` 的 H1 响应不获可信 profile。无效代理配置的 parse error
  改为不含原 URL 的描述，错误码不变，防止 userinfo 经错误文本泄漏。
- `gfn_project_http2_metric_day_v2` 只接受 HTTP2/2，严格校验 source start /
  UTC 日端点，使用当日 Primary Target 的已 finalized HTTP JSON 和原始观测时间。
  H2+ALPN=h2 为 positive；可信 HTTPS H1 为 negative；可信明文 HTTP 不适用；
  旧/缺失/矛盾证据 unknown，stale/failed/not_probed 保留原语义。
  状态、原因码、Fact 版本 provenance、维度去重和计数守恒均有断言。
- SQLC 分发先匹配 `http2_adoption/2`，再进入已有 version=2 分支。v1、TLS13、
  HSTS/CSP/证书、IPv6/security.txt 的旧 SQL 求值器未变。
- Admin 已按实际 `metric_version` / registry `status` 读取，Overview 只选 active；
  原比率函数零 known 返回 null，不拼接 v1/v2。本批未改 API 或前端。

## 唯一新迁移

`db/nav/migrations/20261018010000_nav_http2_probe_metric_v2.sql`

先退役 `http2_adoption/1`，按维护者补充授权同时退役
`http2_transition/1`，设置 `retired_at`。完整保留两者 Registry、
Checkpoint、Metric/Change 历史与旧求值器；没有 `http2_transition/2`，
其他 Change 检测器保持不变。显式历史 rebuild 可用，active-only backfill 拒绝退休版本。
Down 明确拒绝破坏性回退；既有迁移和 `expected/` 未修改。

唯一 active 的 HTTP2/2 使用独立 checkpoint。起点初值是两个 Fact source start 与
迁移完成 UTC 日期下一日的最大值；任一 Fact 起点未知则保持 NULL。它仅是生产 T
的下限，不是部署完成的证明。新增函数含中文业务注释，Admin DataOps migration
inventory 同步；SQLC 与 `expected-final/gfn.json` 均由既有工具生成，未手写生成结果。

## 实测协议

全部调用生产 `performRequest`，没有另造一个测试专用 H2 Client。

| 本地端点/路径 | 实际 resp.Proto / 最终 TLS ALPN | 结果 |
| --- | --- | --- |
| TLS H2，返回 403、自签证书 | HTTP/2.0 / h2；证书独立验证 false | PASS |
| TLS H1 | HTTP/1.1 / http/1.1 | PASS |
| 明文 HTTP | HTTP/1.1 / 空 | PASS |
| HTTP → HTTPS H2 重定向 | HTTP/2.0 / h2，记录最终 URL | PASS |
| HTTPS H2 → HTTP 重定向 | HTTP/1.1 / 空，最终 scheme=http | PASS |
| HTTP CONNECT → TLS H2 | HTTP/2.0 / h2 | PASS |
| HTTPS CONNECT → TLS H2 | HTTP/2.0 / h2 | PASS |
| TLS 协商失败、CONNECT 407、非法代理配置 | 原失败归因，无可信 profile；敏感 userinfo 不进入错误消息 | PASS |
| GODEBUG 禁用 HTTP2 | H1 响应，无可信 profile | PASS |

## 必要验证

使用 Go 1.26.9；隔离 Docker PostgreSQL **18.6**，仅绑定本机 loopback、
临时数据位于 tmpfs。Go/Goose/SQLC 在工作站运行，无共享库或生产连接。
日志保存在忽略的 `.cache/92-a-review/`，临时 PG 容器验收后移除。

| 验证 / 实际命令（在对应模块目录） | 结果 |
| --- | --- |
| Nav Collector：`go test ./collector/http/service ./collector/facts ./collector/metrics -count=1` | PASS（首轮，无集成配置时集成测试按原契约跳过） |
| 最终 HTTP owner：`go test ./collector/http/service -count=1 -v` | PASS，含上表真实协商和运行时禁用用例 |
| 配置隔离 YAML 后：`go test ./collector/metrics -count=1 -v` | 原有 catalog / P2.3 projector / retired rebuild 通过；初次新增 fixture 失败见下文 |
| 修正后：`go test ./collector/metrics -run '^TestHTTP2MetricV2Integration$' -count=1 -v` | PASS：31 个实体场景 + 3 个拒绝输入子用例；另验重复投影、6 个维度分组、零 known、空集合、原 v2 分发 |
| `go test ./collector/changes -count=1 -v` | PASS，既有 Change projector 与 v1/v2 engine 场景保留 |
| `go test ./collector/metrics -run '^TestP23SiteCapabilityMetricBackfillAndRebuildIntegration$' -count=1 -v` | 最终 PASS：退休拒绝 backfill，显式 rebuild 成功，状态/checkpoint 不变 |
| `go test ./collector/changes -run '^TestP23SiteCapabilityChangeBackfillAndRebuildIntegration$' -count=1 -v` | 最终 PASS：同上，历史 HTTP2 Change 事件仍可重建 |
| `go vet ./collector/http/service ./collector/facts ./collector/metrics ./collector/changes` | PASS；最终 HTTP 改动后另行复核该 owner |
| tools：`go test ./db-baseline -run '^TestPostgresHTTP2MetricV2Upgrade$' -count=1 -v` | PASS：次 UTC 日 / 更晚 Fact 起点 / 起点未知三场景；完整旧行与函数定义不变，其他 registry 不变，拒绝 Down |
| tools：`GOFURRY_UPDATE_SCHEMA_SNAPSHOTS=1 go test ./db-baseline -run '^TestPostgresFreshAndBaselineAdoption$/^gfn$' -count=1 -v` | PASS：隔离 PG18 自动生成唯一 gfn final 快照 |
| 取消更新模式后：`go test ./db-baseline -run '^TestPostgresFreshAndBaselineAdoption$/^gfn$|^TestExpectedFinalReadabilityPolicy$' -count=1 -v` | PASS：Nav fresh、历史升级/adoption、快照比较与可读性；没有跑其他数据库集成 |
| tools：`go tool sqlc generate -f ../sqlc.yaml`、`go tool sqlc vet -f ../sqlc.yaml`、`go run ./check-sqlc` | PASS，生成差异只有 Nav Collector metrics.sql.go |
| Admin：`go test ./internal/app/dataops -run '^TestExpectedRepositoryMigrationVersions$' -count=1` | PASS |
| Admin：`go test ./internal/app/metricadmin -run '^TestRatesUseKnownAndEligibleDenominators$' -count=1` | PASS |
| 格式与范围 / `git diff --check` | PASS |
| 全仓库 Go/DB、前端 Build/Browser/Visual、生产探测/迁移/部署、远端手动重型 CI | NOT RUN，超出本批 |

本地测试修正历史：新增“无 Primary”fixture 初次违反既有
`gfn_site_daily_primary_shape_check`，改为同时置空关联字段后通过；退休 Metric
错误消息断言初次误用 Change 的措辞，随后按 Metric 的既有拒绝信息校正并通过。
未修改产品约束、已有求值器、跳过规则或放宽状态断言。

## 生产待单独授权的操作

1. 确认维护窗口、现有 v1 最后 watermark、两个 Fact 起点和发布计划；暂停旧 Collector。
2. 用生产 migrator 执行新 Goose，再部署匹配的新 Collector；本批均未执行。
3. 从目标环境实际出口/代理验证新版 profile、真实协议/ALPN与失败归因，核对
   v1 Metric/Change retired、v2 唯一 active、旧历史完整。
4. 冻结 T：必须是上述验证完成后的首个完整 UTC 日，且不早于迁移次 UTC 日和
   两个 Fact source start。跨日或验证失败时停下，另行批准前向起点修正；
   不假定今天启用，不伪造过去的 H2 观测。
5. 按 Observation → Facts → Metrics 顺序验收 T，核对 known/eligible 与 null；
   不用旧 v1 0% 代替 v2。HTTP2 Change v2、#92-B/#92-C 和 HTTP/3 不在本批。

结论：**本地开发工程验收 PASS，生产切换 NOT RUN / 待授权**。未创建分支、
Worktree 或 PR；不合并 main、不发版、不关闭 #92。
