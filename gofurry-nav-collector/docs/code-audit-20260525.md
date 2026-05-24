# GoFurry Nav Collector 代码审计报告

## Summary

本次审计范围为 `gofurry-nav-collector` 当前代码，重点关注生产运行稳定性、三种协议采集的数据可信度、Redis/DB 旁路写入错误处理、并发和长期运行风险。

结论：没有发现 P0 / P1 级别的立即阻塞问题。当前 collector 的主要风险集中在 P2 / P3：部分预算和错误语义还不够闭合，少数热路径在站点规模继续增长后会变得不够优雅，部分外部观测数据缺少“是否完整”的标记。

## Scope

- Project type: Go 后台采集 worker。
- Runtime context: 单实例生产服务，保留可选多实例 Redis lease。
- Critical surfaces: HTTP / DNS / Ping 网络采集、Redis 写入、DB 写入、v2 observation、summary 聚合、调度运行状态。
- Review scope: `gofurry-nav-collector` 全仓库重点路径。
- Report target: `gofurry-nav-collector/docs/code-audit-20260525.md`。

## Severity Overview

| Severity | Count | Meaning |
|---|---:|---|
| P0 | 0 | Critical |
| P1 | 0 | High |
| P2 | 5 | Medium |
| P3 | 3 | Low |

## Findings

### P0 - Critical

No findings.

### P1 - High

No findings.

### P2 - Medium

### P2-001: DNS 递归 children 未纳入全局记录预算

- Severity: P2
- Category: Reliability / Performance
- Location: `collector/dns/service/dnsService.go:773`, `collector/dns/service/dnsService.go:811`
- Status: Open
- Confidence: High

#### Problem

`max_dns_records_per_query` 当前只在父级 `in.Answer` 循环中检查 `len(results) >= maxRecords`。但 CNAME / MX / NS 递归查询产生的 `children` 会直接 append 到当前记录里，没有纳入同一个全局预算。

#### Impact

遇到异常或刻意构造的 DNS 链路时，单个目标的 v2 payload 和递归处理成本可能超过配置预期。当前 `MaxDepth=2` 已经降低了风险，但预算语义仍不完整。

#### Evidence

父级 Answer 在 `queryDNS` 中按 `maxRecords` 截断；CNAME / MX / NS children 查询结果追加时没有共享剩余预算。

#### Recommendation

为单次域名解析引入递归预算对象，例如 `recordBudget`，父记录和 children 共同消耗预算。children 超出预算时只保留已采到的前 N 条，并在 v2 metadata 中记录 `dns_record_budget_exhausted=true`。

#### Suggested Change

- 将 `queryDNS(domain, qtype, ...)` 内部改为调用带预算参数的私有函数。
- 预算只影响 v2 payload 和旧 DNSRecord children 数量，不增加查询类型和查询次数。
- 增加测试覆盖 CNAME 链和 MX children 超出预算场景。

#### Verification

- 构造递归 children 数量超过 `max_dns_records_per_query` 的测试。
- 运行 `go test ./collector/dns/service`。

### P2-002: DNS IPv6 私有和特殊地址未进入 `private_ip` 风险标记

- Severity: P2
- Category: Correctness / Security
- Location: `collector/dns/service/dnsService.go:1005`
- Status: Open
- Confidence: High

#### Problem

`detectDNSRiskFlags` 目前只检查 IPv4 私网、环回、链路本地和 CGNAT 网段，没有检查 IPv6 的环回、ULA、链路本地等特殊地址。

#### Impact

如果站点 DNS 返回 `::1`、`fc00::/7`、`fe80::/10` 等地址，v2 observation 可能漏掉 `private_ip` 信号，导致安全参考信息不完整。

#### Evidence

`privateRanges` 只包含 IPv4 CIDR：`10.0.0.0/8`、`127.0.0.0/8`、`192.168.0.0/16` 等。

#### Recommendation

补齐 IPv6 地址判断，可优先使用 `net.IP` 的标准方法和显式 CIDR：

- loopback: `ip.IsLoopback()`
- unspecified: `ip.IsUnspecified()`
- link-local: `ip.IsLinkLocalUnicast()`
- ULA: `fc00::/7`

#### Suggested Change

增加 `isPrivateOrSpecialIP(ip net.IP) bool` helper，并用表驱动测试覆盖 IPv4 / IPv6。

#### Verification

- 新增测试：`::1`、`fc00::1`、`fe80::1` 均生成 `private_ip`。
- 运行 `go test ./collector/dns/service`。

### P2-003: v2 site summary 每次更新都扫描 Redis 前缀

- Severity: P2
- Category: Performance / Maintainability
- Location: `collector/observation/summary.go:68`
- Status: Open
- Confidence: High

#### Problem

`UpdateSiteSummary` 每次 target observation 更新后都会调用 `FindByPrefix("collector:v2:summary:target:{site_id}:")`，通过 Redis `SCAN` 找到站点下所有 target summary。

#### Impact

当前几百个站点规模下问题不严重，但这是 observation 写入热路径。随着目标数量、协议数量或写入频率增加，重复 `SCAN` 会造成不必要的 Redis 压力，也会让 summary 写入延迟变得不稳定。

#### Evidence

`UpdateSiteSummary` 先扫描 key，再逐个 `GET` target summary，最后重新聚合 site summary。

#### Recommendation

用更明确的数据结构维护每个站点的 target 索引，例如：

- `collector:v2:summary:site_targets:{site_id}` set 保存 target。
- target summary 写入成功后 `SADD` target。
- site summary 聚合时 `SMEMBERS` 读取目标列表，再按 key 获取 summary。

#### Suggested Change

保留现有 summary key，不影响后端读取；只新增旁路索引 set。缺失索引时可 fallback 到 `SCAN`，便于平滑上线。

#### Verification

- 单元测试覆盖有索引、索引缺失 fallback 两种路径。
- 手工删除索引 set 后确认 summary 仍可恢复。

### P2-004: Redis `SetNX` 无法向调用方区分命令失败和 key 已存在

- Severity: P2
- Category: Reliability / Observability
- Location: `common/service/redisService.go:87`, `collector/scheduler/run.go:189`
- Status: Open
- Confidence: High

#### Problem

Redis wrapper `SetNX` 只返回 bool；Redis 命令错误和 key 已存在都会表现为 `false`。调度 lease 获取失败时会统一记录为 `lease_held_by_other_collector`。

#### Impact

未来启用多实例 lease 时，Redis 短暂错误会被误判为其他实例持有 lease，导致跳过原因不准确，排查困难。单实例默认关闭 lease，生产主路径风险较低。

#### Evidence

`SetNX` 内部记录 Redis 错误后返回 `false`；`AcquireLeaseOrSkip` 收到 `false` 后直接 `r.Skip("lease_held_by_other_collector", 0)`。

#### Recommendation

将 Redis wrapper 调整为返回 `(created bool, err common.GFError)`。调用方根据错误写入 `lease_acquire_failed`，只有无错误且 `created=false` 才记录 `lease_held_by_other_collector`。

#### Suggested Change

- 新增 `SetNXResult` 或调整接口。
- 更新 scheduler memoryStore 测试桩。
- Ping 目标刷新等调用点同步处理错误。

#### Verification

- 单元测试覆盖 Redis 错误、key 已存在、获取成功三种 lease 场景。

### P2-005: Ping 目标列表刷新忽略 `SetNX` 结果

- Severity: P2
- Category: Reliability / Observability
- Location: `collector/ping/service/pingService.go:138`
- Status: Open
- Confidence: High

#### Problem

Ping 每轮从 DB 构建目标列表后先删除旧 Redis key，再调用 `SetNX` 写入新列表，但没有检查 `SetNX` 返回值。

#### Impact

如果 Redis 写入失败，后续读取 `ping:sites` 可能得到空数据或旧数据，表现为 Ping 本轮目标异常，但日志里没有明确说明“刷新目标列表失败”。

#### Evidence

`addAllIpToPing` 检查了 `Del` 错误，但 `cs.SetNX(...)` 返回值被直接忽略。

#### Recommendation

Ping 目标列表刷新应使用明确的覆盖写入语义，例如 `Set` + `SetExpire`，或继续使用 `SetNX` 但必须处理返回值和 Redis 错误。

#### Suggested Change

- 推荐用 `cs.Set` 写目标列表，再 `cs.SetExpire` 或封装 `SetWithTTL`。
- 写入失败记录中文结构化日志，并返回旁路错误。
- 保持旧 key 名 `ping:sites` 不变。

#### Verification

- 单元测试 mock Redis 写入失败，确认错误被返回和记录。

### P3 - Low

### P3-001: HTTP body 是否被限长截断没有进入 payload

- Severity: P3
- Category: Correctness / Observability
- Location: `collector/http/service/httpService.go:1286`
- Status: Open
- Confidence: Medium

#### Problem

HTTP 响应体使用 `io.LimitReader` 限制读取大小，但没有多读 1 字节判断是否被截断。后续 HTML 解析会把截断后的 body 当作完整页面。

#### Impact

大页面的 title/meta/icon/canonical 等信息可能不完整，前端展示或 v2 observation 使用者无法区分“页面本身没有这些字段”和“采集时被截断后没有读到”。

#### Evidence

代码读取 `io.ReadAll(io.LimitReader(resp.Body, maxBytes))`，只记录 `body_read_bytes`。

#### Recommendation

读取 `maxBytes + 1`，如果超过上限则截回 `maxBytes` 并记录：

- `body_truncated=true`
- `body_limit_bytes`
- `body_original_read_bytes` 或 `body_read_bytes_before_truncate`

#### Verification

- 构造超过上限的 response body，确认 v2 payload 标记截断。

### P3-002: 旧 JSON 写入路径仍有少量 `sonic.Marshal` 错误被忽略

- Severity: P3
- Category: Reliability / Maintainability
- Location: `collector/http/service/httpService.go:291`, `collector/ping/service/pingService.go:447`, `collector/dns/service/dnsService.go:380`
- Status: Open
- Confidence: Medium

#### Problem

部分旧 Redis / 旧日志表 JSON 写入路径忽略了 `sonic.Marshal` 返回的错误。

#### Impact

当前模型大多是可序列化结构，实际触发概率较低。但后续字段扩展时如果加入不可序列化类型，旧链路可能写入空 JSON 或错误状态不清晰。

#### Evidence

多处使用 `jsonResult, _ := sonic.Marshal(...)`。

#### Recommendation

统一检查 marshal 错误：旧链路可记录失败状态和中文结构化日志；v2 仍按旁路错误处理，不影响主采集循环继续运行。

#### Verification

- 添加 payload builder 或 marshal helper 测试，覆盖 marshal 失败时的日志/状态处理。

### P3-003: DNS GeoIP / PTR 缓存没有大小或生命周期边界

- Severity: P3
- Category: Performance / Maintainability
- Location: `collector/dns/service/dnsService.go:49`, `collector/dns/service/dnsService.go:982`, `collector/dns/service/dnsService.go:1065`
- Status: Open
- Confidence: Medium

#### Problem

`geoCache` 和 `ptrCache` 使用包级 `sync.Map`，长期运行时不会清理。当前站点数量较少，风险不高，但缓存生命周期没有明确边界。

#### Impact

如果采集目标频繁变化，或外部 DNS 返回大量变化 IP，缓存会缓慢增长，增加长期运行内存占用，也会保留较旧的 PTR/GeoIP 结果。

#### Evidence

`lookupGeoASN` 和 `reversePTR` 命中缓存后直接返回，写入后没有 TTL、最大数量或随 DNS 轮次清理的机制。

#### Recommendation

保持低复杂度即可：优先增加按轮次清理或简单 TTL cache，不必引入大型缓存库。比如每次 DNS run 开始时重建本轮缓存，或缓存项带 `created_at`，超过 24 小时后刷新。

#### Verification

- 单元测试覆盖过期缓存会重新查询。
- 长时间运行观察内存曲线不随历史 IP 无限增长。

## Recommended Fix Plan

建议把所有开放项纳入 `v0.5.1`，按以下顺序处理：

1. 先修 Redis `SetNX` 错误语义和 Ping 目标刷新错误处理，因为这直接影响运行状态可解释性。
2. 再修 DNS 预算和 IPv6 风险标记，因为这影响 observation 数据可信度。
3. 优化 summary 聚合热路径，保留 `SCAN` fallback，降低上线风险。
4. 补充 HTTP truncation 标记、旧 JSON marshal 错误日志、DNS cache 生命周期。
5. 最后补齐测试和手工验证流程。

## Verification Suggestions

建议修复后运行：

```bash
cd gofurry-nav-collector
gofmt -l .
go test ./...
go vet ./...
git diff --check
```

建议增加的重点测试：

- DNS CNAME / MX / NS children 超预算。
- IPv6 私网、环回、链路本地风险标记。
- Redis `SetNX` 成功、key 已存在、命令失败。
- Ping 目标列表刷新失败。
- HTTP body 超过 `max_response_bytes`。
- site summary 有索引和 fallback 两种聚合路径。

## Notes

- 本审计没有建议增加采集频率、请求次数、DNS 查询目标或默认并发。
- 本审计没有建议引入 Prometheus、分布式调度、漏洞扫描或高强度探测。
- 所有建议均可作为补丁版本修复，不需要改变现有 `/api/v1`、旧 Redis key、旧日志表或 Nuxt 展示结构。
