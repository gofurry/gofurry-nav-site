# GoFurry Nav Backend Roadmap

## 当前位置

`gofurry-nav-backend` 当前生产导航接口仍以 `/api/v1/nav/...` 为主。v1 已接入 collector 旧数据面：站点域名列表、Ping latest/history、HTTP latest、DNS latest，并继续服务现有 Nuxt 前端。

collector v2 数据面已经进入可供后端消费的阶段。后端目前已完成 v2 summary 只读接入、collector summary hints 透传，以及站点详情页 v2 只读后端接口：

```text
GET /api/v2/nav/sites/:siteId/detail
GET /api/v2/nav/sites/:siteId/summary
GET /api/v2/nav/sites/:siteId/targets/:target/summary
GET /api/v2/nav/sites/:siteId/targets/:target/latest
GET /api/v2/nav/sites/:siteId/targets/:target/observations
GET /api/v2/nav/sites/:siteId/targets/:target/trend
GET /api/v2/nav/sites/:siteId/targets/:target/changes
GET /api/v2/nav/sites/:siteId/targets/:target/light-probes
```

这些接口当前都受 `nav_v2.summary_enabled` 控制。后端内部已建立 collector v2 read model，可以只读消费 raw observation、target latest、trend、change、run state、light probe，并通过 v2 detail 分接口对外暴露。

已知短板：

- v1 详情接口 `getSiteDetail` 仍带浏览量递增副作用，v2 详情页接口应保持只读。
- `nav_v2.summary_enabled` 当前同时充当 v2 route gate，后续如需更细粒度灰度，可再拆分成独立 detail 开关。

## 路线策略

- `/api/v1/nav/...` 继续保持生产兼容，不因 v2 接入改变响应结构。
- `/api/v2/nav/...` 先补齐 collector v2 read model，再稳定站点详情页聚合接口；前端暂不迁移，等后端 v2 接口稳定后再改造。
- v2 后端只读接入 collector 输出，不触发采集、不重算 collector 健康摘要、不把 light probe 直接纳入上下架或 down 判断。
- 所有 payload 采用兼容读取策略：未知字段忽略，缺失字段给安全默认值；collector 后续只能兼容新增字段。
- 详情页 v2 把读详情和写浏览量拆开，避免继续沿用 v1 的读取副作用。

## Version Plan

### v0.2.x - Summary 接入补齐

**Status:** Completed
**Scope:** API / Documentation / Stability
**Goal:** 在现有 summary 只读接口基础上补齐 collector 已稳定输出但后端丢弃的字段。

#### Focus

- site summary 与 target summary 字段完整性
- summary hints 透传
- collector v1/v2 信息面对照文档

#### Tasks

- [x] 增加 `nav_v2.summary_enabled` 与 `summary_stale_after_seconds` 配置。
- [x] 注册 `/api/v2/nav` summary 路由组。
- [x] 实现 site summary 只读接口。
- [x] 实现 target summary 只读接口。
- [x] 增加 Redis 缺失、过期、JSON 解析失败的 summary 服务测试。
- [x] 在 collector docs 中新增 v1/v2 信息面字段级统计文档。
- [x] 在 backend summary DTO 中补齐 target summary 的 `canonical_target_hint`、`target_relation_hints`、`edge_provider_hints`。
- [x] 在 backend site summary DTO 中补齐顶层 `target_relation_hints` 与 `targets[]` 内的 hints。
- [x] 增加 hints 字段透传测试，确认未知字段不影响旧 summary 响应。

#### Acceptance Criteria

- v2 summary 接口能完整保留 collector summary 稳定字段。
- v2 summary 缺失或过期仍返回成功响应，并清晰标记 `state=missing` 或 `state=stale`。
- `/api/v1/nav/...` 行为完全不变。
- 文档能说明 v1/v2 字段差异和后端当前接入缺口。

---

### v0.3.0 - Collector v2 Read Model 基础

**Status:** Completed
**Scope:** API / Architecture / Testing
**Goal:** 建立后端读取 collector v2 全量信息面的基础层，为站点详情 v2 聚合接口提供稳定数据来源。

#### Focus

- observation DB 查询
- Redis latest/trend/change/light probe 读取
- 参数校验与响应外壳统一

#### Tasks

- [x] 新增 `gfn_collector_observation` 只读 DAO，支持按 `site_id`、`target`、`protocol`、`observed_at` 查询。
- [x] 新增 target latest 读取服务，优先读 `collector:v2:latest:{protocol}:{site_id}:{target}`。
- [x] 新增 raw observation 查询服务，支持协议白名单和 limit 上限。
- [x] 新增 trend 读取服务，读取 `collector:v2:trend:target:{site_id}:{target}`。
- [x] 新增 change 读取服务，读取 `collector:v2:change:target:{site_id}:{target}`。
- [x] 新增 light probe latest 读取服务，覆盖 `rdap`、`robots`、`security_txt`、`page_assets`、`port_check`、`waf_canary`。
- [x] 新增 run state 读取设计或诊断接口规划，读取 `collector:v2:run:{protocol}:latest`。
- [x] 将 latest 与 raw observation 响应统一到 collector envelope：`site_id`、`target`、`protocol`、`status`、`observed_at`、`duration_ms`、`error_code`、`error_message`、`payload`、`schema_version`、`collector_id`、`job_id`。
- [x] 增加 read model 单元测试，覆盖 Redis key 缺失、JSON 解析失败、DB 空结果、非法 protocol、limit 越界。

#### Acceptance Criteria

- 后端可以只读消费 collector v2 的核心 raw/latest/derived/light probe 信息面。
- 查询失败不会影响 v1 接口，也不会触发 collector 行为。
- 所有 v2 read model 接口和服务都有稳定参数校验和可测试错误路径。
- payload 不被后端重新解释成健康结论；健康状态仍以 summary 为准。

---

### v0.4.0 - 站点详情页 v2 后端接口稳定化

**Status:** Completed
**Scope:** API / User-facing / Documentation
**Goal:** 提供稳定的站点详情页 v2 后端接口，让前端后续可以一次性迁移到更完整的 collector v2 数据面。

#### Focus

- 详情页聚合接口
- target 级分接口
- 只读详情语义
- v1/v2 迁移对照

#### Planned Routes

```text
GET /api/v2/nav/sites/:siteId/detail?lang=zh&target={target}
GET /api/v2/nav/sites/:siteId/targets/:target/latest
GET /api/v2/nav/sites/:siteId/targets/:target/observations?protocol={protocol}&limit={limit}
GET /api/v2/nav/sites/:siteId/targets/:target/trend
GET /api/v2/nav/sites/:siteId/targets/:target/changes
GET /api/v2/nav/sites/:siteId/targets/:target/light-probes
```

#### Tasks

- [x] 新增站点详情 v2 API 设计文档，明确响应结构和字段来源。
- [x] 实现 `detail` 只读聚合接口，响应包含 `site`、`targets`、`selected_target`、`site_summary`、`target_summary`、`latest_core`、`derived`、`light_probe_state`、`generated_at`、`schema_version`。
- [x] `detail` 不递增浏览量；浏览量写入另行设计独立接口。
- [x] 实现 target latest 分接口，返回 `ping`、`http`、`dns` 和已启用 light probe 的 latest。
- [x] 实现 target observation 历史分接口，支持 `ping`、`http`、`dns`、`rdap`、`robots`、`security_txt`、`page_assets`、`port_check`、`waf_canary`。
- [x] 实现 target trend 和 changes 分接口，缺失时返回空结构或 `state=missing`。
- [x] 实现 target light-probes 分接口，明确 light probe 不参与健康状态。
- [x] 增加站点不存在、target 不属于站点、summary missing、latest missing、partial failure 的测试。
- [x] 更新 Swagger 或补充 Markdown 接口文档。

#### Acceptance Criteria

- 前端无需调用 v1 Ping/HTTP/DNS 详情接口即可获得详情页所需后端数据。
- v2 detail 是只读接口，不写 DB、不写 Redis、不改变 view count。
- target 选择、summary、latest、trend、change、light probe 都有清晰缺失语义。
- v1 和 v2 可以并存，前端迁移可灰度、可回滚。

---

### v0.4.1 - Nav v2 接口安全与可靠性修复

**Status:** Planned
**Scope:** Security/Safety / Stability / Performance / Testing
**Goal:** 根据 v2 新增代码审计结果，修复详情页 v2 后端接口在归属校验、并发初始化、响应体控制和测试覆盖上的硬化缺口。

#### Focus

- target 归属边界
- v2 singleton / DAO 初始化安全
- raw payload 响应体控制
- detail 聚合性能与 partial failure 语义
- HTTP contract 测试补齐

#### Tasks

- [ ] 修复 P1-001：target 分接口只以 active `gfn_collector_domain` 作为授权来源，summary-only target 仅用于展示补充，不单独授权 raw/latest/history 查询。
- [ ] 为 `targets[]` 增加来源与注册状态字段，例如 `source`、`registered`、`summary_only`，并让 detail 默认 target 优先选择 active DB domain。
- [ ] 修复 P2-001：用 `sync.Once` 或启动期显式初始化替换 v2 detail/readmodel/DAO 的无保护懒加载写入。
- [ ] 增加 `go test -race ./apps/nav/detail/... ./apps/nav/readmodel/...` 验证路径，并补并发初始化测试。
- [ ] 修复 P2-002：为 latest / observations 的 raw `payload` 增加默认 preview 或截断策略，并提供显式 full payload 策略。
- [ ] 为 `observations?limit=500` + 大 payload 场景增加响应体大小和截断测试。
- [ ] 修复 P2-003：优化 detail 聚合读取路径，支持 Redis latest 批量读取或独立分支并发读取。
- [ ] 明确 detail partial failure 语义：summary 核心失败返回错误，trend/change/light probe 等旁路失败返回 `state=missing` 与 reason。
- [ ] 增加 Redis JSON 解析失败结构化日志，记录 `redis_key`、`site_id`、`target`、`protocol` 和错误摘要，但不记录完整 payload。
- [ ] 拆分 `nav_v2.summary_enabled` 粗粒度 route gate，增加 detail/read model 独立灰度开关并保持旧配置兼容。
- [ ] 补齐 detail controller table tests，覆盖 6 个 v2 路由的成功、参数错误、service error 和 URL encoded target。

#### Acceptance Criteria

- 已删除或不在 active domain 表中的 target 不能通过任何 target 分接口读取 raw/latest/history 数据。
- `go test -race ./apps/nav/detail/... ./apps/nav/readmodel/...` 通过。
- 大 payload 不会在默认公开响应中无限制透传，响应会明确标记截断状态。
- detail 聚合接口在旁路数据缺失或失败时保持清晰、可预测的响应语义。
- v2 summary 与 detail 可以独立灰度开启或关闭。
- 新增审计报告中的 P1/P2 项均关闭或有明确剩余风险说明。

#### Notes

- 审计报告：`docs/nav-v2-code-audit.md`。
- 本阶段只修复 v2 后端接口硬化问题，不启动前端迁移。

---

### v1.0.0-alpha.1 - Nav v2 API 冻结候选

**Status:** Planned
**Scope:** API / Testing / Documentation / Release
**Goal:** 在 collector v2 read model 与站点详情 v2 接口稳定后，冻结 `/api/v2/nav` 核心结构，准备前端迁移。

#### Focus

- API freeze
- 前端迁移准备
- 文档与测试补齐
- v1 保留策略

#### Tasks

- [ ] 完成 `/api/v1/nav` 到 `/api/v2/nav` 的接口对照文档。
- [ ] 补齐 v2 controller/service/dao 核心测试。
- [ ] 补齐 Swagger 或 OpenAPI 文档。
- [ ] 明确 v1 生产接口维护期限和废弃策略。
- [ ] 制定 Nuxt 前端迁移验证清单。
- [ ] 制定上线、灰度、回滚步骤。

#### Acceptance Criteria

- v2 接口结构不再频繁变化。
- 前端可以按文档迁移，不需要反向阅读 collector 或 backend 内部代码。
- v1 回滚路径仍存在。
- v2 API 的缺失、过期、解析失败、权限和参数错误都有稳定响应语义。

## Suggested Release Path

- `v0.2.x`：已补齐 summary 字段完整性和文档。
- `v0.3.0`：已建立 collector v2 read model 基础。
- `v0.4.0`：已稳定站点详情页 v2 后端接口。
- `v0.4.1`：修复 v2 新增接口审计发现的安全、并发、性能与测试硬化项。
- `v1.0.0-alpha.1`：冻结 v2 API 候选并准备前端迁移。
- `v1.0.0`：前后端完成 v2 迁移并保留明确 v1 回滚策略后再进入稳定版。
