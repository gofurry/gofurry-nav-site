# GoFurry Nav Backend Roadmap

## 当前位置

`gofurry-nav-backend` 当前已经把线上导航业务接口收敛到 `/api/v1/nav/...`。这一层作为生产稳定接口继续服务 Nuxt 前端与现有业务，不在 collector v2 数据面打通前做大规模迁移。

下一阶段的重点不是重写前后端接口，而是先配合 `gofurry-nav-collector` 打通 v2 observation 核心能力：collector 旁路写入 observation DB 与 v2 Redis latest，backend 提供默认关闭的只读 v2 观测接口，Nuxt 与 gofurry-admin 暂不切换。

## 路线策略

- 先完成 collector v2 observation 数据面，再考虑站点、分组、搜索、页面素材等业务接口的 v2 化。
- `/api/v1/nav/...` 保持生产兼容，不因为 v2 observation 改造而改变响应结构。
- `/api/v2/nav/...` 优先承载更清晰的资源语义，避免继续扩散 `getXxx`、`page/site/list` 这类历史命名。
- v2 初期只读、默认关闭、可灰度、可回滚，不引入 Prometheus 生态，不做健康评分，不做高频探测。

## 版本计划

### v0.2.0 - Collector Observation 只读接入

**状态：** 规划中  
**范围：** API / 架构 / 稳定性  
**目标：** 在不影响 `/api/v1/nav` 生产展示链路的前提下，提供读取 collector v2 observation 的最小只读接口。

#### 重点

- 只接入 collector v2 observation 核心数据。
- 后端接口默认关闭，仅用于人工验证、灰度观察与后续前端改造准备。
- 不改 Nuxt 当前数据源，不改 gofurry-admin。

#### 计划接口

```text
GET /api/v2/nav/observations/latest?protocol=ping
GET /api/v2/nav/observations/latest?protocol=http
GET /api/v2/nav/observations/latest?protocol=dns
GET /api/v2/nav/sites/:siteId/observations/latest
GET /api/v2/nav/sites/:siteId/observations?protocol=ping&limit=288
GET /api/v2/nav/sites/:siteId/observations?protocol=http&limit=240
GET /api/v2/nav/sites/:siteId/observations?protocol=dns&limit=60
```

#### 响应语义

- `latest` 返回最新观测结果，不做健康评分。
- `observations` 返回历史观测序列，按 `observed_at desc` 排序。
- `protocol` 仅允许 `ping`、`http`、`dns`。
- `siteId` 使用 collector 与主站稳定共享的站点 ID。
- 响应包含统一外壳字段：`site_id`、`protocol`、`status`、`observed_at`、`duration_ms`、`error_code`、`error_message`、`payload`、`schema_version`。
- `payload` 保留协议细节，后端不在这一阶段把协议结果解释成综合状态。

#### 任务

- [ ] 在配置中增加 v2 nav observation API 开关，默认关闭。
- [ ] 新增 observation 只读 DAO，优先读 observation DB；必要时再考虑 v2 Redis latest。
- [ ] 新增 `/api/v2/nav` 路由组，保持 `api := app.Group("/api")`、`v2 := api.Group("/v2")`、`nav := v2.Group("/nav")` 的结构。
- [ ] 新增站点级最新观测接口。
- [ ] 新增站点级历史观测接口。
- [ ] 新增协议级全站最新观测接口。
- [ ] 增加协议、limit、siteId 参数校验。
- [ ] 增加只读查询单元测试或小范围服务测试。
- [ ] 更新 Swagger 或接口文档。

#### 验收标准

- v2 开关关闭时，`/api/v1/nav/...` 行为完全不变。
- v2 开关开启后，可以读取 Ping / HTTP / DNS observation 最新值与历史值。
- v2 接口不写 DB、不写 Redis、不触发 collector 行为。
- 查询失败时返回清晰错误，不影响 v1 生产接口。
- Nuxt 与 gofurry-admin 不需要同步改动即可上线。

---

### v0.3.0 - Nav v2 资源路由规范化

**状态：** 规划中  
**范围：** API / 架构 / 文档  
**目标：** 在 observation 核心链路稳定后，逐步把 nav 业务接口整理为更清晰的 `/api/v2/nav` 资源风格。

#### 重点

- 设计优先，迁移后置。
- 先保留 `/api/v1/nav`，避免一次性牵动 Nuxt 与 gofurry-admin。
- v2 使用资源名与查询参数表达业务，不继续使用历史动作式路径。

#### 建议路由规范

站点与分组：

```text
GET /api/v2/nav/sites
GET /api/v2/nav/sites/:siteId
GET /api/v2/nav/groups
```

搜索建议：

```text
GET /api/v2/nav/search/suggestions?provider=baidu&q=关键词
GET /api/v2/nav/search/suggestions?provider=bing&q=关键词
GET /api/v2/nav/search/suggestions?provider=google&q=关键词
GET /api/v2/nav/search/suggestions?provider=bilibili&q=关键词
```

页面辅助资源：

```text
GET /api/v2/nav/sayings/random
GET /api/v2/nav/backgrounds/random?type=standard
GET /api/v2/nav/backgrounds/random?type=mobile
GET /api/v2/nav/changelog
```

站点观测数据继续沿用 v0.2.0 的 observation 路由：

```text
GET /api/v2/nav/observations/latest
GET /api/v2/nav/sites/:siteId/observations/latest
GET /api/v2/nav/sites/:siteId/observations
```

#### 任务

- [ ] 为 v2 站点列表设计响应结构，明确是否兼容 v1 字段名。
- [ ] 为 v2 站点详情设计响应结构，明确浏览量增加逻辑是否仍由详情接口触发。
- [ ] 为 v2 分组接口设计响应结构，明确 `site_ids` 与站点摘要是否分离。
- [ ] 把搜索建议收敛为统一 provider 参数。
- [ ] 把随机金句、随机背景、更新公告从 `page/header` 历史命名中拆出。
- [ ] 规划 Nuxt 迁移顺序，避免前后端不一致。
- [ ] 规划 gofurry-admin 是否需要消费 v2，避免管理端与前台接口职责混在一起。

#### 验收标准

- v2 路由命名符合资源语义，后续扩展不需要重复 `/nav` 或动作式路径。
- v1 与 v2 可以并存，生产迁移可灰度、可回滚。
- Nuxt 切换 v2 前有明确接口对照表。
- gofurry-admin 的接口边界单独评估，不被 nav 前台 v2 迁移顺手改乱。

---

### v1.0.0-alpha.1 - API 稳定化候选

**状态：** 规划中  
**范围：** API / 测试 / 文档 / 发布  
**目标：** 在 v2 observation 与 nav v2 资源接口稳定后，进入正式稳定版前的 API 冻结阶段。

#### 重点

- 冻结 `/api/v2/nav` 核心接口结构。
- 补齐测试、Swagger、迁移文档。
- 明确 v1 保留期限与迁移策略。

#### 任务

- [ ] 完成 v1 到 v2 的接口对照文档。
- [ ] 补齐 v2 controller/service/dao 的核心测试。
- [ ] 补齐 Nuxt 切换 v2 的验证清单。
- [ ] 明确 `/api/v1/nav` 的维护期限。
- [ ] 更新部署与回滚文档。

#### 验收标准

- v2 接口结构不再频繁变化。
- 前后端迁移路径清晰。
- v1 回滚路径仍然存在。
- 文档足够支撑生产更新。
