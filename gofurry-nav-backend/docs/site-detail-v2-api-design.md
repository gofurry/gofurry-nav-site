# 站点详情页后端 v2 接口设计

本文档规划 `gofurry-nav-backend /api/v2/nav` 的站点详情页后端接口。目标是先稳定后端 v2 只读接口，再考虑前端改造。本轮不修改前端。

## 设计目标

- 保留现有 `/api/v1/nav/...` 生产接口，v2 与 v1 并存。
- v2 详情页接口只读，不触发 collector，不写 DB，不写 Redis。
- v2 详情页不沿用 v1 `GET /api/v1/nav/site/getSiteDetail` 的浏览量递增副作用；浏览量写入另行设计独立接口。
- 健康状态以 collector v2 summary 为准，后端不重新评分。
- raw observation 与 latest 保留 collector envelope，payload 中未知字段透传或忽略，不做破坏性转换。
- light probe、trend、change、edge hints 都是旁路解释信息，不参与站点上下架、排序或 down 判断。

## 已存在接口

```text
GET /api/v2/nav/sites/:siteId/summary
GET /api/v2/nav/sites/:siteId/targets/:target/summary
```

### v0.2.x 已补齐字段

后端 summary DTO 已补齐 collector 输出的 summary hints，详情页 v2 后续可以直接读取现有 summary 接口中的这些字段：

| 接口 | 已补字段 |
|---|---|
| site summary | 顶层 `target_relation_hints`。 |
| site summary | `targets[].canonical_target_hint`、`targets[].target_relation_hints`、`targets[].edge_provider_hints`。 |
| target summary | `canonical_target_hint`、`target_relation_hints`、`edge_provider_hints`。 |

## 新增接口

### 站点详情聚合

```text
GET /api/v2/nav/sites/:siteId/detail?lang=zh&target={target}
```

#### 参数

| 参数 | 位置 | 必填 | 说明 |
|---|---|---:|---|
| `siteId` | path | 是 | `gfn_site.id`，必须大于 0。 |
| `lang` | query | 否 | `zh` / `en`，默认 `zh`。 |
| `target` | query | 否 | 指定采集目标；为空时优先使用站点下第一个有效 target。 |

#### 响应字段

```json
{
  "site": {},
  "targets": [],
  "selected_target": "",
  "site_summary": {},
  "target_summary": {},
  "latest_core": {},
  "derived": {},
  "light_probe_state": {},
  "generated_at": "",
  "schema_version": 1
}
```

| 字段 | 类型 | 来源 | 说明 |
|---|---|---|---|
| `site` | object | `gfn_site` | 站点基础信息。 |
| `targets` | object[] | `gfn_collector_domain` + summary target index | 当前站点可采集 target 列表。 |
| `selected_target` | string | query 或默认选择 | 本次聚合使用的 target。 |
| `site_summary` | object | `collector:v2:summary:site:{site_id}` | 站点健康摘要；缺失时返回 `state=missing`。 |
| `target_summary` | object | `collector:v2:summary:target:{site_id}:{target}` | target 健康摘要；缺失时返回 `state=missing`。 |
| `latest_core` | object | target latest Redis | `ping`、`http`、`dns` 最新 observation。 |
| `derived` | object | trend/change Redis | target 趋势和变化事件。 |
| `light_probe_state` | object | light probe latest Redis | 低频旁路探测最新状态。 |
| `generated_at` | string | backend | 后端聚合时间。 |
| `schema_version` | number | backend | v2 detail 响应版本，初始为 `1`。 |

#### `site`

| 字段 | 类型 | 来源 |
|---|---|---|
| `id` | number | `gfn_site.id` |
| `name` | string | `gfn_site.name` / `gfn_site.name_en` |
| `info` | string | `gfn_site.info` / `gfn_site.info_en` |
| `icon` | string or null | `gfn_site.icon` |
| `country` | string or null | `gfn_site.country` |
| `nsfw` | string | `gfn_site.nsfw` |
| `welfare` | string | `gfn_site.welfare` |
| `view_count` | number | `gfn_site.view_count`，只读返回，不递增。 |

#### `targets[]`

| 字段 | 类型 | 来源 |
|---|---|---|
| `target` | string | `TRIM(prefix || name)` |
| `domain_id` | number | `gfn_collector_domain.id` |
| `name` | string | `gfn_collector_domain.name` |
| `prefix` | string or null | `gfn_collector_domain.prefix` |
| `tls` | string | `gfn_collector_domain.tls` |
| `proxy` | string | `gfn_collector_domain.proxy` |
| `summary_state` | string | target summary 读取状态：`ready` / `missing` / `stale` |
| `status` | string | target summary `status`，缺失时为 `unknown` |

### Target latest

```text
GET /api/v2/nav/sites/:siteId/targets/:target/latest
```

#### 响应字段

```json
{
  "site_id": 133,
  "target": "go-furry.com",
  "state": "ready",
  "protocols": {
    "ping": {},
    "http": {},
    "dns": {},
    "rdap": {},
    "robots": {},
    "security_txt": {},
    "page_assets": {},
    "port_check": {},
    "waf_canary": {}
  },
  "generated_at": "",
  "schema_version": 1
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `state` | string | 至少一个协议有 latest 时为 `ready`，全部缺失时为 `missing`。 |
| `protocols` | object | key 为协议名，value 为 collector latest envelope。 |
| `generated_at` | string | 后端响应生成时间。 |
| `schema_version` | number | 后端响应版本。 |

每个协议 value 保留以下 envelope：

| 字段 | 类型 | 说明 |
|---|---|---|
| `site_id` | number | 站点 ID。 |
| `target` | string | 采集目标。 |
| `protocol` | string | 协议名。 |
| `status` | string | `success` / `failure`。 |
| `observed_at` | string | observation 时间。 |
| `duration_ms` | number | 耗时。 |
| `error_code` | string | 可空。 |
| `error_message` | string | 可空。 |
| `payload` | object | 协议原始 payload。 |
| `schema_version` | number | collector schema version。 |
| `collector_id` | string | 可空。 |
| `job_id` | string | 可空。 |

### Target observations

```text
GET /api/v2/nav/sites/:siteId/targets/:target/observations?protocol={protocol}&limit={limit}
```

#### 参数

| 参数 | 位置 | 必填 | 说明 |
|---|---|---:|---|
| `siteId` | path | 是 | 必须大于 0。 |
| `target` | path | 是 | URL path 中需编码。 |
| `protocol` | query | 是 | 允许 `ping`、`http`、`dns`、`rdap`、`robots`、`security_txt`、`page_assets`、`port_check`、`waf_canary`。 |
| `limit` | query | 否 | 默认按协议设置；服务端限制最大值，避免大查询。 |

#### 响应字段

```json
{
  "site_id": 133,
  "target": "go-furry.com",
  "protocol": "http",
  "items": [],
  "limit": 240,
  "generated_at": "",
  "schema_version": 1
}
```

`items[]` 使用与 latest 相同的 collector envelope，并按 `observed_at desc` 排序。

### Target trend

```text
GET /api/v2/nav/sites/:siteId/targets/:target/trend
```

#### 响应字段

| 字段 | 类型 | 来源 |
|---|---|---|
| `state` | string | `ready` / `missing` / `stale`。 |
| `site_id` | number | trend doc。 |
| `target` | string | trend doc。 |
| `windows` | object | `24h` / `7d` 窗口。 |
| `generated_at` | string | collector trend 生成时间。 |
| `schema_version` | number | collector schema version。 |

`windows.{window}.protocols.{protocol}` 保留 collector 字段：`observation_count`、`success_count`、`failure_count`、`success_rate`、`avg_duration_ms`、`p95_duration_ms`、`last_observed_at`、`last_failure_at`、`http`、`ping`、`dns`、`tls`。

### Target changes

```text
GET /api/v2/nav/sites/:siteId/targets/:target/changes
```

#### 响应字段

| 字段 | 类型 | 来源 |
|---|---|---|
| `state` | string | `ready` / `missing`。 |
| `site_id` | number | change doc。 |
| `target` | string | change doc。 |
| `events` | object[] | change events。 |
| `generated_at` | string | collector change 生成时间。 |
| `schema_version` | number | collector schema version。 |

`events[]` 字段：`event_id`、`protocol`、`category`、`field`、`old_value`、`new_value`、`old_observed_at`、`new_observed_at`、`detected_at`。

### Target light probes

```text
GET /api/v2/nav/sites/:siteId/targets/:target/light-probes
```

#### 响应字段

```json
{
  "site_id": 133,
  "target": "go-furry.com",
  "state": "ready",
  "protocols": {
    "rdap": {},
    "robots": {},
    "security_txt": {},
    "page_assets": {},
    "port_check": {},
    "waf_canary": {}
  },
  "generated_at": "",
  "schema_version": 1
}
```

每个协议 value 使用 collector latest envelope。light probe 缺失不影响 `site_summary` 或 `target_summary` 的健康状态。

## 状态与缺失语义

| 场景 | HTTP 语义 | 响应语义 |
|---|---|---|
| site 不存在或已删除 | 业务错误 | 返回错误消息，不伪造空详情。 |
| target 不属于 site | 业务错误 | 返回错误消息，避免跨站 target 读取。 |
| summary key 缺失 | 成功 | `state=missing`、`status=unknown`。 |
| summary 过期 | 成功 | `state=stale`、`status=unknown`。 |
| latest key 缺失 | 成功 | 对应协议缺失；全部缺失时整体 `state=missing`。 |
| trend/change 缺失 | 成功 | `state=missing`，返回空 `windows` 或空 `events`。 |
| raw observation 为空 | 成功 | `items=[]`。 |
| Redis JSON 解析失败 | 业务错误 | 返回明确解析失败错误，并记录日志。 |
| DB 查询失败 | 业务错误 | 返回明确查询失败错误，不影响 v1。 |

## 数据来源

| 模块 | 来源 |
|---|---|
| 站点基础信息 | `gfn_site` |
| target 列表 | `gfn_collector_domain`，可参考 `collector:v2:summary:site_targets:{site_id}` 补齐 summary index |
| site summary | `collector:v2:summary:site:{site_id}` |
| target summary | `collector:v2:summary:target:{site_id}:{target}` |
| core latest | `collector:v2:latest:ping:{site_id}:{target}`、`collector:v2:latest:http:{site_id}:{target}`、`collector:v2:latest:dns:{site_id}:{target}` |
| light probe latest | `collector:v2:latest:{protocol}:{site_id}:{target}` |
| raw history | `gfn_collector_observation` |
| trend | `collector:v2:trend:target:{site_id}:{target}` |
| changes | `collector:v2:change:target:{site_id}:{target}` |

## 实现顺序建议

1. 补齐现有 summary DTO hints 字段，先确保 collector 已稳定输出的信息不再丢失。
2. 新增 observation 只读 DAO 与 Redis v2 read helpers。
3. 实现 target latest / observations / trend / changes / light-probes 分接口。
4. 在分接口稳定后实现 `detail` 聚合接口。
5. 补齐接口测试与文档，再安排前端迁移。

## 测试场景

| 场景 | 期望 |
|---|---|
| summary ready | 返回 collector 原字段，包含 hints。 |
| summary missing | 返回 `state=missing`，状态为 `unknown`。 |
| summary stale | 返回 `state=stale`，状态为 `unknown`。 |
| target latest 部分缺失 | 已存在协议正常返回，缺失协议不阻断整体响应。 |
| raw observation 协议非法 | 返回参数错误。 |
| raw observation limit 越界 | 使用服务端上限或返回参数错误，具体策略在实现时固定并测试。 |
| target 不属于 site | 返回业务错误。 |
| Redis JSON 损坏 | 返回解析错误，日志包含 key。 |
| trend/change 缺失 | 返回 `state=missing` 和空数据结构。 |
| v1 接口回归 | `/api/v1/nav/...` 响应结构不变。 |
