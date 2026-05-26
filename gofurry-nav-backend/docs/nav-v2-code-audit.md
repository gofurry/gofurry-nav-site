# Nav v2 后端代码审计报告

## Summary

本次审计范围是 `gofurry-nav-backend` 在 v0.2.x 到 v0.4.0 新增的 `/api/v2/nav` 后端代码，重点覆盖 summary hints、collector v2 read model、站点详情页 detail 聚合接口与 target 分接口。

总体结论：v2 代码已经具备基础只读能力，未发现 P0 级别问题。当前最需要进入 `v0.4.1` 的工作是收紧 target 归属边界、消除首次请求懒加载竞态、控制 raw payload 响应体大小，并补充接口级硬化测试。

## Scope

- `apps/nav/summary`
- `apps/nav/readmodel`
- `apps/nav/detail`
- `routers/url_v2.go`
- `routers/router.go` 中 v2 route gate
- `docs/roadmap.md` 与 `docs/site-detail-v2-api-design.md` 的实现状态一致性

## Severity Overview

| Severity | Count | Meaning |
|---|---:|---|
| P0 | 0 | Critical |
| P1 | 1 | High |
| P2 | 3 | Medium |
| P3 | 3 | Low |

## Findings

### P0 - Critical

No findings.

### P1 - High

#### P1-001: target 归属校验信任 Redis summary index，可能绕过 DB 软删除边界

- Severity: P1
- Category: Security / Correctness
- Location: `apps/nav/detail/service/detailService.go:219`, `apps/nav/detail/service/detailService.go:241`, `apps/nav/detail/service/detailService.go:313`
- Status: Open
- Confidence: Medium

##### Problem

`ensureSiteTarget` 先检查 `gfn_collector_domain`，随后在 Redis site summary 的 `targets[]` 中找到 target 也会放行。`mergeTargets` 也会把 site summary 中存在、但 DB domain 表中不存在的 target 合并进详情页 target 列表。

##### Impact

如果 Redis summary index 残留了已经删除、迁移或错误归属的 target，公开 target 分接口可能继续暴露该 target 的 latest、observations、trend、changes 或 light probe 数据。这个风险取决于 Redis summary 的可信度和清理及时性，因此置信度为 Medium，但它影响的是公开接口的归属边界，优先级应高于普通清理项。

##### Evidence

`ensureSiteTarget` 在 DB domain 未命中后读取 site summary，并在 `siteSummary.Targets` 命中时直接返回成功。`mergeTargets` 也把 summary-only target 当作可选 target 合并到响应。

##### Recommendation

把 active `gfn_collector_domain` 作为 target 分接口的授权来源。site summary 中的 target 可以作为详情页展示补充，但应标记 `source=summary_index` 或 `registered=false`，不能单独授权 raw/latest/history 分接口。

##### Suggested Change

- 新增 target inventory DTO 字段：`source`、`registered`、`summary_only`。
- `ensureSiteTarget` 只允许 active DB domain 通过。
- detail 聚合默认 target 优先选择 active DB domain；没有 active target 时返回 missing 空结构。
- 删除或停用 domain 时同步清理 collector summary target index。

##### Verification

- 增加测试：DB domain 不存在但 site summary targets 命中时，`latest/observations/trend/changes/light-probes` 返回 `target 不属于当前 site`。
- 增加测试：detail 响应可以展示 summary-only target，但不会默认选中为 raw data target。

### P2 - Medium

#### P2-001: v2 懒加载单例存在首次并发请求数据竞态

- Severity: P2
- Category: Concurrency / Reliability
- Location: `apps/nav/detail/controller/detail.go:23`, `apps/nav/detail/controller/detail.go:117`, `apps/nav/detail/service/detailService.go:43`, `apps/nav/detail/service/detailService.go:45`, `apps/nav/readmodel/service/readModelService.go:25`, `apps/nav/readmodel/service/readModelService.go:27`, `apps/nav/detail/dao/detailDao.go:13`, `apps/nav/detail/dao/detailDao.go:17`, `apps/nav/readmodel/dao/observationDao.go:11`, `apps/nav/readmodel/dao/observationDao.go:15`
- Status: Open
- Confidence: High

##### Problem

新增 v2 模块使用包级变量做懒加载：`detailSvc`、`detailSingleton`、`readModelSingleton`、`newDetailDao`、`newObservationDao`。这些变量在首次请求时会被读写，但没有 `sync.Once`、锁或启动期初始化保护。

##### Impact

生产服务启动后，如果多个请求同时打到 v2 接口，Go race detector 会认为这些读写存在数据竞态。实际影响可能是重复初始化 DAO/service，或者在极端情况下读到部分初始化状态，增加首批请求的不稳定性。

##### Evidence

`currentDetailService` 在 `detailSvc == nil` 时直接赋值。`GetDetailService` 与 `GetReadModelService` 在字段为 nil 时直接写入 singleton 字段。DAO getter 也在 `Gm == nil` 时直接调用 `Init()`。

##### Recommendation

使用 `sync.Once` 或启动期显式注入完成初始化。测试注入可以保留 package-level hook，但生产路径应避免请求期无保护写全局变量。

##### Suggested Change

- 为 controller/service/DAO singleton 分别增加 `sync.Once`。
- controller 测试使用显式 setter，如 `setDetailReaderForTest`，并在测试后恢复。
- 增加 `go test -race ./apps/nav/detail/... ./apps/nav/readmodel/...` 作为 v0.4.1 验证项。

##### Verification

- 新增并发测试：100 个 goroutine 同时调用 `currentDetailService()` 和 service getter。
- 运行 `go test -race ./apps/nav/detail/... ./apps/nav/readmodel/...`。

#### P2-002: raw payload 直接对外透传，响应体大小缺少上限策略

- Severity: P2
- Category: Performance / Security
- Location: `apps/nav/readmodel/models/readmodel.go:25`, `apps/nav/readmodel/models/readmodel.go:56`, `apps/nav/readmodel/service/readModelService.go:122`, `apps/nav/readmodel/service/readModelService.go:128`, `apps/nav/readmodel/service/readModelService.go:134`, `apps/nav/readmodel/service/readModelService.go:297`
- Status: Open
- Confidence: Medium

##### Problem

latest 与 observations 会把 collector `payload` 以 `json.RawMessage` 原样返回。observations 虽然限制最多 500 条，但没有限制单条 payload 字节数，也没有字段选择或 preview 模式。

##### Impact

如果 collector payload 变大，公开 API 可能返回很大的响应体，带来内存、带宽和延迟压力。对于导航站详情页，前端通常并不需要完整 raw payload，尤其是 `observations?limit=500` 场景。

##### Evidence

`GfnCollectorObservation.Payload` 是完整 JSONB 字符串，`CollectorEnvelope.Payload` 是 `json.RawMessage`，`ListObservations` 会把查询结果逐条转换后完整返回。

##### Recommendation

在公开 HTTP 层增加 payload 输出策略，默认返回安全 preview 或核心字段，完整 payload 使用显式参数或内部诊断接口暴露。

##### Suggested Change

- 增加 `payload_mode=summary|full`，默认 `summary`。
- 增加单条 payload 最大字节数配置，例如 `nav_v2.max_payload_bytes`。
- 对被截断 payload 增加 `payload_truncated=true` 和 `payload_bytes`。
- 为 `limit=500` + 大 payload 添加响应体大小测试。

##### Verification

- 构造 1MB payload observation，确认默认响应被截断或降级。
- 确认 `payload_mode=full` 受配置开关或内部权限控制。

#### P2-003: detail 聚合路径串行读取多个 Redis key，高并发下容易放大尾延迟

- Severity: P2
- Category: Performance / Reliability
- Location: `apps/nav/detail/service/detailService.go:70`, `apps/nav/detail/service/detailService.go:104`, `apps/nav/readmodel/service/readModelService.go:72`
- Status: Open
- Confidence: Medium

##### Problem

`GetSiteDetail` 会依次读取 site summary、target summary、core latest、light probe latest、trend 和 changes。read model 的 latest 读取又按协议串行读 Redis key。

##### Impact

单次 detail 请求会产生多次 Redis round-trip。低流量下问题不明显，但详情页成为前端主路径后，高并发会放大 Redis 压力和接口尾延迟。

##### Evidence

`GetSiteDetail` 顺序调用多个 read service；`GetTargetLatest` 按协议循环读取 Redis key。

##### Recommendation

v0.4.1 先做小步优化：在 detail 聚合内为独立数据面并发读取，或者为 Redis latest 使用 pipeline/mget 读取。并增加超时与部分失败策略，避免某一路派生信息拖垮整个 detail。

##### Suggested Change

- read model 增加批量 latest 读取接口。
- detail 聚合并发读取 `target_summary`、`latest_core`、`light_probe`、`trend`、`changes`。
- 明确 partial failure：核心 summary 失败返回错误，旁路 trend/change/light probe 失败可返回 `state=missing` 并附 reason。

##### Verification

- 增加 fake Redis 延迟测试，确认 detail 聚合耗时接近最长分支而不是所有分支累加。
- 增加 partial failure 测试，确认旁路失败不会阻断核心详情。

### P3 - Low

#### P3-001: Redis JSON 解析失败缺少 key/site/target/protocol 结构化日志

- Severity: P3
- Category: Reliability / Maintainability
- Location: `apps/nav/readmodel/service/readModelService.go:87`, `apps/nav/readmodel/service/readModelService.go:153`, `apps/nav/readmodel/service/readModelService.go:182`, `apps/nav/summary/service/summaryService.go:53`, `apps/nav/summary/service/summaryService.go:84`
- Status: Open
- Confidence: High

##### Problem

summary/readmodel 在 JSON 解析失败时返回业务错误，但没有记录 Redis key、site_id、target、protocol 等定位信息。

##### Impact

线上出现 collector 写入异常、schema 漂移或 Redis 脏数据时，后端日志难以直接定位坏 key，需要人工复现或扫 Redis。

##### Evidence

相关代码直接返回 `站点健康摘要解析失败`、`目标健康摘要解析失败`、`target latest 解析失败` 等错误。

##### Recommendation

解析失败时增加结构化日志，至少包含 `redis_key`、`site_id`、`target`、`protocol`、`schema_version` 和原始错误。

##### Suggested Change

在 decode 分支调用项目现有 `common/log`，避免把完整 payload 打进日志，只记录 key 与错误摘要。

##### Verification

- 增加 bad JSON 测试时注入 logger 或检查日志 hook。
- 手工构造坏 Redis key，确认日志可定位 key。

#### P3-002: v2 detail 与 summary 共用 `nav_v2.summary_enabled`，灰度开关粒度偏粗

- Severity: P3
- Category: Configuration / Maintainability
- Location: `routers/router.go:75`, `routers/url_v2.go:9`, `docs/roadmap.md:25`
- Status: Open
- Confidence: High

##### Problem

当前 `/api/v2/nav` 所有接口都受 `nav_v2.summary_enabled` 控制。这个开关名称已经无法准确表达 detail、read model 分接口和 summary 的启停边界。

##### Impact

后续灰度或回滚时，只想关闭 detail 分接口会连带关闭 summary；只想保留 summary 也无法单独控制 detail。

##### Evidence

`registerRoutes` 只判断 `NavV2.SummaryEnabled` 后注册整个 `navV2Api`。

##### Recommendation

增加独立配置，例如 `nav_v2.enabled`、`nav_v2.summary_enabled`、`nav_v2.detail_enabled`、`nav_v2.read_model_enabled`。保持向后兼容：旧 `summary_enabled=true` 可继续开启 summary。

##### Verification

- 增加路由注册测试，覆盖 summary/detail 分别开启和关闭。
- 更新 `conf/server.yaml` 示例与文档。

#### P3-003: controller 级接口测试覆盖不足

- Severity: P3
- Category: Testing / Maintainability
- Location: `apps/nav/detail/controller/detail_test.go:17`
- Status: Open
- Confidence: High

##### Problem

当前 controller 测试只覆盖非法 siteId、非法 limit 和 latest 成功。detail、observations、trend、changes、light-probes 的成功/错误响应没有完整 HTTP 层测试。

##### Impact

后续修改路由、参数名或响应包装时，service 测试可能仍然通过，但公开 HTTP contract 发生回归。

##### Evidence

`detail/controller` 目前只有 3 个测试函数，而 v0.4.0 新增了 6 个公开路由。

##### Recommendation

补齐 controller table tests，覆盖每个路由的成功响应、参数错误、service error 透传和 URL encoded target。

##### Verification

- 新增 `TestDetailRoutes` table test。
- 运行 `go test ./apps/nav/detail/controller -count=1`。

## Recommended Fix Plan

1. `v0.4.1` 第一优先级：收紧 target 归属校验，区分 DB registered target 与 summary-only target。
2. `v0.4.1` 第二优先级：用 `sync.Once` 或启动期初始化消除新增 v2 singleton/DAO 懒加载竞态。
3. `v0.4.1` 第三优先级：为 raw payload 增加默认 preview/截断策略，并补大 payload 测试。
4. `v0.4.1` 第四优先级：优化 detail 聚合读取路径，至少完成 latest 批量读取或独立分支并发读取。
5. `v0.4.1` 收尾：补齐日志、开关粒度和 controller HTTP contract 测试。

## Verification Suggestions

```bash
go test ./apps/nav/detail/...
go test ./apps/nav/readmodel/...
go test ./apps/nav/summary/...
go test ./...
go test -race ./apps/nav/detail/... ./apps/nav/readmodel/...
git diff --check
```

## Notes

- 本次审计没有发现 SQL 注入：新增 DAO 查询都使用 GORM 参数绑定。
- 本次审计没有发现 v2 detail 会递增 view count；新增 detail 路径没有调用 v1 `touchSiteViewCount`。
- 本次审计没有发现文件读写、外部命令执行或 SSRF 型网络访问。
