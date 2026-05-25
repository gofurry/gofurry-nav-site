# GoFurry Nav Collector v0.5.8 代码审计报告

## Summary

本次审计聚焦 `gofurry-nav-collector` 当前 v2 observation、summary、light probe、WAF canary、page assets 与 Redis run state。整体没有发现 P0 级直接漏洞；主要问题集中在默认关闭能力开启后的边界保护、长任务可观测性、结果判定可信度和长期 Redis 索引治理。

## Scope

- `collector/lightprobe/service`
- `collector/observation`
- `common/service` Redis 辅助函数
- `roof/env` 配置默认值和协议开关
- `docs/roadmap.md` 与 v2 payload 文档

## Severity Overview

| Severity | Count | Meaning |
|---|---:|---|
| P0 | 0 | Critical |
| P1 | 1 | High |
| P2 | 4 | Medium |
| P3 | 1 | Low |

## Findings

### P0 - Critical

No findings.

### P1 - High

#### P1-001: page_assets 允许同注册域资源但未校验解析后的私网地址

- Severity: P1
- Category: Security
- Location: `collector/lightprobe/service/lightProbeService.go:1100`, `collector/lightprobe/service/lightProbeService.go:1214`
- Status: Open
- Confidence: Medium

##### Problem

`page_assets` 会从 HTTP latest 中读取页面声明的 icon / manifest URL，并允许同注册域或 allowlist host 的资源。当前只校验 URL scheme、host 和注册域，没有在发起请求前解析目标 IP 并拒绝 loopback、private、link-local、unspecified 等地址。

##### Impact

如果某个已授权收录站点的页面声明指向同注册域下解析到内网地址的资源，collector 可能从运行环境向内网地址发起请求。该能力默认关闭且只拉取 1 个 icon / 1 个 manifest，风险受限，但一旦在生产启用，仍属于 SSRF 边界不足。

##### Evidence

`assetURLAllowed` 只根据 host 和注册域判断是否允许，随后 `probeDeclaredAsset` 直接调用 `probeHTTPGetURL`。

##### Recommendation

为 page_assets 增加 `resolvePublicIPsOnly` 检查：请求前解析 host，拒绝私网、环回、链路本地、多播、未指定地址；解析失败则跳过并写 `skipped_reason=asset_ip_not_allowed` 或 `asset_dns_resolve_failed`。

##### Verification

- 单元测试：同注册域但解析到 `127.0.0.1` / `10.0.0.1` 时不发起请求。
- 单元测试：公网 IP 正常放行。

### P2 - Medium

#### P2-001: WAF canary 启用后会在启动时立即跑全量目标

- Severity: P2
- Category: Reliability / Safety
- Location: `collector/lightprobe/service/lightProbeService.go:139`, `collector/lightprobe/service/lightProbeService.go:175`
- Status: Open
- Confidence: High

##### Problem

light probe 当前开启后都会 `go RunXxx()` 立即执行一轮。`waf_canary` 是 30 天低频能力，但一旦启用，服务启动时会马上对所有有效 target 执行 baseline + 11 个 case。

##### Impact

这符合本地验证需求，但生产启用时可能在发布时间点触发一轮集中 canary 请求。虽然目标已授权且请求无害，仍可能造成 WAF 日志突增、运行时间过长或误判为异常流量。

##### Recommendation

为 `waf_canary` 增加 `run_on_start` 配置，默认 `false`；本地测试可显式设为 `true`。其他既有 light probe 行为保持不变。

##### Verification

- 默认启用 `waf_canary.enabled=true` 但 `run_on_start` 未配置时，不在启动时执行。
- 配置 `run_on_start=true` 时启动后执行一轮。

#### P2-002: WAF canary 严格匹配单一预期状态码会制造假阴性

- Severity: P2
- Category: Correctness
- Location: `collector/lightprobe/service/lightProbeService.go:590`, `collector/lightprobe/service/lightProbeService.go:781`
- Status: Open
- Confidence: High

##### Problem

`blocked=true` 认可 `400/403/405/414/415`，但 `matched_expected` 对每个 case 又要求固定状态码。例如路径穿越返回 `400` 会被认为 blocked，但不匹配 `403`。

##### Impact

不同 WAF、Nginx、应用网关对同类拦截会返回不同状态码。过度严格会让“已拦截但状态码不同”的 case 被误读为失败，降低报告可信度。

##### Recommendation

把 WAF case 结果拆成 `blocked`、`status_code_expected`、`matched_expected`：只要被 blocked 即可计入 `expected_blocked_matched_count`；状态码差异单独记录为 `status_code_unexpected_count`。

##### Verification

- 路径穿越返回 `400` 时 `blocked=true` 且 `matched_expected=true`，同时 `status_code_expected=false`。
- 返回 `200` 时仍为 `unexpected_pass`。

#### P2-003: WAF canary 与 light probe 长任务缺少运行中进度状态

- Severity: P2
- Category: Observability / Reliability
- Location: `collector/lightprobe/service/lightProbeService.go:175`, `collector/scheduler/run.go`
- Status: Open
- Confidence: High

##### Problem

run state 在 `Start()` 时写入一次，此时 `target_count=0`。目标加载后只在内存里 `SetTargetCount`，不会再次写 running state；长任务运行时 Redis 里看不到当前 target 数量或已完成计数。

##### Impact

本地测试时会看到 `collector:v2:run:waf_canary:latest` 长时间保持 `running` 且 `target_count=0`，不利于判断任务是否卡住、正在运行还是已经开始处理目标。

##### Recommendation

在加载目标后写一次 running state，并在 light probe executor 每处理 N 个 target 或每隔固定时间刷新一次进度。字段可复用现有 `success_count/failure_count/skipped_count/error_count`。

##### Verification

- 目标加载后 run state 中 `target_count` 立即更新。
- 长任务中途 Redis latest 能看到计数变化。

#### P2-004: WAF canary 缺少 per-run 最大目标数和 per-target case 并发控制

- Severity: P2
- Category: Performance / Safety
- Location: `collector/lightprobe/service/lightProbeService.go:294`, `collector/lightprobe/service/lightProbeService.go:539`
- Status: Open
- Confidence: Medium

##### Problem

`waf_canary` 当前对所有有效目标顺序执行 12 个请求，无法通过配置限制本轮最大目标数，也无法控制 target 级整体执行预算。

##### Impact

当前几百个站点仍可接受，但一旦目标数增长或部分目标响应慢，单轮任务可能持续很久。顺序执行降低瞬时压力，但缺少明确的安全预算，不利于生产控制和本地快速验证。

##### Recommendation

增加 `max_targets_per_run`，默认 `0` 表示不限；增加 `max_cases_per_target` 或固定 case 集合版本号用于测试压缩。保持默认不提高并发，不增加重试。

##### Verification

- 配置 `max_targets_per_run=3` 时只执行前三个目标，并记录 `truncated_target_count`。
- 空值保持现有全量行为。

### P3 - Low

#### P3-001: Site summary target 索引长期只增不减

- Severity: P3
- Category: Maintainability / Reliability
- Location: `collector/observation/summary.go:112`, `collector/observation/summary.go:119`
- Status: Open
- Confidence: Medium

##### Problem

site summary 使用 Redis set 记录 target 索引，但没有在采集域名软删除、site_id 迁移或 target 改名后清理旧 target。`UpdateSiteSummary` 会跳过缺失 summary，但索引本身长期增长。

##### Impact

对当前规模影响较小，但长期会让 summary 更新多读一些无效 key，也会让人工排查时看到历史 target 残留。

##### Recommendation

增加轻量索引清理：当读取 target summary 为空或 site_id 不匹配时，从 `collector:v2:summary:site_targets:{site_id}` 中移除；或新增低频 cleanup 只清理 summary index。

##### Verification

- Redis set 中存在不存在的 target summary key 时，下一次 site summary 更新会删除该 target。
- 不影响有效 target 的 summary 聚合。

## Recommended Fix Plan

将上述问题收敛到 `v0.5.8 - 审计发现修复与轻探测边界加固`：

1. 先修 page_assets 解析后 IP 边界，降低 SSRF 风险。
2. 再修 WAF canary 启动行为、状态码判定和运行进度，提升生产可控性。
3. 最后补 summary index 清理和对应测试。

## Verification Suggestions

- `gofmt -l .`
- `go test ./...`
- `go vet ./...`
- 针对 light probe 追加私网 IP、WAF 状态码、run state 进度、summary index 清理的单元测试。

## Notes

- 本次没有发现需要立即停用生产 collector 的问题。
- WAF canary、port_check、page_assets 都是默认关闭能力；生产启用前仍应明确授权边界和配置预算。
