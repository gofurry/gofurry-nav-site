# Collector v2 本地与测试环境 Smoke Test

本文档用于在后端正式推进 `/api/v2/nav` 前，快速确认 `gofurry-nav-collector` v2 数据面是否完整可用。它只验证旁路数据，不要求改变旧 Redis key、旧日志表、旧前端展示或生产采集频率。

## 前置条件

- 已执行 observation 相关 SQL，存在 `gfn_collector_observation` 表。
- `gfn_collector_domain.site_id` 已完成回填，且有效采集目标满足 `deleted IS NOT TRUE AND site_id > 0`。
- Redis、PostgreSQL 与 GeoIP mmdb 路径可用。
- 测试目标属于自有或明确授权范围；主动 light probe 默认保持关闭，只有需要验证时才临时开启。

## 推荐测试配置

核心 v2 数据面建议开启：

```yaml
collector:
  v2:
    enabled: true
    observation_db: true
    latest_redis: true
    compare_log: false
    edge_hints:
      enabled: true
    protocols:
      ping: true
      http: true
      dns: true
```

低频 light probe 建议按需单项开启：

```yaml
collector:
  v2:
    light_probe:
      rdap:
        enabled: false
      robots:
        enabled: false
      security_txt:
        enabled: false
      page_assets:
        enabled: false
      port_check:
        enabled: false
      waf_canary:
        enabled: false
```

生产默认建议：`rdap`、`robots`、`security_txt`、`page_assets`、`port_check`、`waf_canary` 不需要全部开启；确实需要时保持 7 天或 30 天低频，并确认授权边界。

## 启动验证

在 collector 目录启动：

```powershell
go run .
```

日志中应能看到：

- Ping / HTTP / DNS 模块初始化完成。
- v2 observation 写入失败不会影响旧链路。
- 如果 light probe 未开启，不应出现对应 `light_probe_registered`。
- 如果 run state 开启，应能看到每轮 run start / complete 结构化日志。

## Redis Key 验证

任选一个已知 `site_id` 与 target，例如 `133` / `go-furry.com`。

### latest

```text
GET collector:v2:latest:ping:133:go-furry.com
GET collector:v2:latest:http:133:go-furry.com
GET collector:v2:latest:dns:133:go-furry.com
```

期望：

- JSON 可解析。
- 顶层包含 `site_id`、`target`、`protocol`、`status`、`observed_at`、`payload`、`schema_version`。
- `collector_id`、`job_id` 存在时只作为追踪字段，不影响旧读取。

### summary

```text
GET collector:v2:summary:target:133:go-furry.com
GET collector:v2:summary:site:133
GET collector:v2:summary:site_targets:133
```

期望：

- target summary 包含 `status`、`reason_codes`、`protocols`。
- site summary 包含 `target_count`、`status_counts`、`targets`。
- `edge_provider_hints`、`canonical_target_hint`、`target_relation_hints` 缺失时应被视为正常。

### trend

```text
GET collector:v2:trend:target:133:go-furry.com
```

期望：

- 包含 `windows.24h` 与 `windows.7d`。
- 协议趋势缺失时表示对应窗口内 observation 不足，不应影响 summary。

### change event

```text
GET collector:v2:change:target:133:go-furry.com
```

期望：

- 包含 `events` 数组。
- 没有变化时 `events=[]` 是正常状态。
- 每条事件只包含 old/new 摘要、protocol、category、field 和 observed_at，不包含外部大文本原文。

### run state

```text
GET collector:v2:run:ping:latest
GET collector:v2:run:http:latest
GET collector:v2:run:dns:latest
```

期望：

- 包含 `collector_id`、`job_id`、`protocol`、`status`、`started_at`、`finished_at`、`target_count`、`success_count`、`failure_count`。
- light probe 开启后才应出现对应协议的 run state。

## Observation DB 验证

```sql
SELECT protocol, count(*)
FROM gfn_collector_observation
GROUP BY protocol
ORDER BY protocol;
```

期望：

- `ping`、`http`、`dns` 有数据。
- 低频 light probe 只有开启并跑过后才出现。

```sql
SELECT site_id, target, protocol, status, observed_at
FROM gfn_collector_observation
WHERE site_id = 133
ORDER BY observed_at DESC
LIMIT 20;
```

期望：

- `site_id` 不为 0。
- `target` 与 `gfn_collector_domain` 的有效目标一致。
- `status` 只作为单次 observation 状态，不直接等同站点健康结论。

## 旧链路兼容验证

旧 Redis key 仍应正常：

```text
HGETALL ping:result
GET request:go-furry.com
HGETALL dns:go-furry.com
```

旧日志表仍应写入：

```sql
SELECT count(*) FROM gfn_collector_log_ping;
SELECT count(*) FROM gfn_collector_log_http;
SELECT count(*) FROM gfn_collector_log_dns;
```

## 失败边界验证

- 临时关闭 v2 Redis 或 DB 写入时，旧 Redis key 与旧日志表不应被 v2 旁路失败阻断。
- 删除某个 latest key 后，summary 应能重新生成或在下一轮采集后恢复。
- observation 历史不足时，trend/change key 可以为空或 events 为空，不视为异常。
- light probe 默认关闭时，不应产生额外网络请求。

## 验收结论

满足以下条件即可认为 collector v2 数据面可供后端 `/api/v2/nav` 消费：

- `go test ./...` 与 `go vet ./...` 通过。
- `ping`、`http`、`dns` latest / observation / summary 均可读。
- target trend 与 change event key 可读，缺失或空数组时后端能安全降级。
- 旧 Redis key、旧日志表、旧前端展示不受影响。
- 新增字段遵循兼容新增原则，不要求后端解析未知字段。
