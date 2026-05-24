# GoFurry Nav Collector Roadmap

## 当前状态

`gofurry-nav-collector` 既定 roadmap 已完成。当前文档不再保留历史完成项，只记录下一阶段需要处理的开放问题。

本轮 roadmap 来自 2026-05-25 代码审计，目标是把已经上线后的稳定性、数据可信度和维护性问题整理为 `v0.5.1` 修复计划。

## 迭代原则

- 不增加默认采集频率、请求次数、DNS 查询目标或并发强度。
- 不改变旧 Redis key、旧日志表、旧后端接口和旧前端展示结构。
- 优先修复会影响数据可信度、错误可见性、运行稳定性的点。
- 对外部站点返回的数据继续按不可信内容处理，所有新增说明都只作为 observation 信号。
- 继续避免不必要复杂度；不引入 Prometheus 生态，不做漏洞扫描，不做高频探测。

## 版本计划

### v0.5.1 - 审计发现修复与稳定性补丁

**状态：** 计划中
**范围：** 稳定性 / 数据可信度 / Redis 错误语义 / 测试
**目标：** 修复 2026-05-25 代码审计发现的问题，让 collector 在现有单实例生产模式下更可靠、更容易排查。

#### Focus

- DNS 递归记录预算和 IPv6 风险信号。
- Redis 写入、lease、Ping 目标刷新错误语义。
- v2 summary 聚合热路径收敛。
- HTTP 与旧 JSON 写入的数据可信度补强。
- 不改变采集强度和旧业务链路。

#### Tasks

- [ ] 为 DNS CNAME / MX / NS 递归 children 加入全局记录预算，避免 `max_dns_records_per_query` 只限制父级 Answer。
- [ ] 为 DNS v2 `private_ip` 风险标记补齐 IPv6 特殊地址识别，例如 `::1`、`fc00::/7`、`fe80::/10`。
- [ ] 优化 site summary 聚合路径，避免每次 observation 写入都用 Redis `SCAN` 全量扫 `collector:v2:summary:target:{site_id}:*`。
- [ ] 将 Redis `SetNX` wrapper 改为返回 `created bool` 和 `GFError`，让 lease 获取失败能区分“其他实例持有”和“Redis 命令失败”。
- [ ] 修复 Ping 目标列表刷新时忽略 `SetNX` 结果的问题；刷新失败必须记录清晰日志并返回旁路错误。
- [ ] HTTP 读取响应体时显式记录是否被 `max_response_bytes` 截断，避免把截断 HTML 当作完整页面语义。
- [ ] 检查 Ping / HTTP / DNS 旧 JSON 写入路径中被忽略的 `sonic.Marshal` 错误，至少记录中文结构化日志。
- [ ] 为 DNS GeoIP / PTR 缓存增加边界或清理策略，避免长期运行中缓存无限增长。
- [ ] 为上述修复补充小范围单元测试，不扩大生产默认行为。

#### Acceptance Criteria

- `gofmt -l .` 为空。
- `go test ./...` 通过。
- `go vet ./...` 通过。
- DNS payload 在异常递归链路下仍受预算保护。
- Redis lease 日志能明确区分锁被占用和 Redis 写入失败。
- Ping 目标刷新失败不再静默吞掉。
- HTTP v2 payload 能标记 body 是否截断，旧 Redis/旧表结构保持兼容。
- 生产未修改配置时，采集频率、并发、旧 Redis key、旧表写入、后端接口和前端展示保持不变。

#### Notes

- 审计报告：`docs/code-audit-20260525.md`。
- 本阶段只做补丁修复，不规划新的协议采集字段。
- v0.5.1 完成后，如没有新的生产问题，roadmap 可以继续保持只记录开放项。
