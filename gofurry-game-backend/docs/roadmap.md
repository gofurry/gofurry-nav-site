# GoFurry Game Backend Roadmap

## Current Position

游戏模块已经完成 collector v2、backend v2 read model、admin 采集观测、RAG sync v2、前端主页面 cutover。`gofurry-nav-web` 当前游戏首页、详情页、新闻页、更多新闻页和 sitemap 游戏 URL 来源已经稳定消费 `/api/v2/game/*`，可以认为“动态游戏数据主线”已经切到 v2。

现阶段目标不再是继续维护 v1/v2 双栈，而是逐步把仍在 v1 的接口直接升级为 v2，并移除对应 v1 路由和旧实现。每一步都应该小步完成：先确认调用方，改前端或外部消费者，然后删除旧路由和旧 service/dao 依赖，避免历史包袱继续扩散。

## Stable V2 Mainline

以下接口已经是主线，应继续保留并作为公开合同维护：

- `GET /api/v2/game/list`
- `GET /api/v2/game/info`
- `GET /api/v2/game/news`
- `GET /api/v2/game/news/latest`
- `GET /api/v2/game/panel/main`
- `GET /api/v2/game/sync/list`
- `GET /api/v2/game/sync/info`
- `GET /api/v2/game/sync/news`
- `GET /api/v2/game/sync/creators`
- `GET /api/v2/game/collect/status`
- `GET /api/v2/game/collect/runs`
- `GET /api/v2/game/collect/runs/:run_id`
- `GET /api/v2/game/collect/task-results`
- `GET /api/v2/game/collect/games/:id/status`

这些接口不应该再依赖旧动态表：

- `gfg_game_record`
- `gfg_game_news`
- `gfg_game_player_count`

动态详情、新闻、价格、媒体、在线人数、采集状态都以 collector v2 PostgreSQL 表为事实来源。

## Remaining V1 Surface

本阶段已经移除的 v1 动态接口：

- `GET /api/v1/game/info`
- `GET /api/v1/game/info/list`
- `GET /api/v1/game/info/main`
- `GET /api/v1/game/panel/main`
- `GET /api/v1/game/update/latest`
- `GET /api/v1/game/update/latest/more`
- `GET /api/v1/game/sync/list`
- `GET /api/v1/game/sync/info`
- `GET /api/v1/game/sync/news`
- `GET /api/v1/game/sync/creators`

仍需要升级的 v1 运营和交互接口：

- `GET /api/v1/game/remark`
- `GET /api/v1/game/tag/list`
- `GET /api/v1/game/creator`
- `GET /api/v1/game/recommend/CBF`
- `GET /api/v1/game/recommend/random`
- `POST /api/v1/game/search/simple`
- `POST /api/v1/game/search/page`
- `POST /api/v1/game/review/anonymous`
- `GET /api/v1/game/review/latest`
- `POST /api/v1/game/prize/participation`
- `GET /api/v1/game/prize/participation/activation`
- `GET /api/v1/game/prize/info`

这些接口不应长期以 v1 形式保留。后续每个阶段完成后，直接让调用方切到 v2，并移除对应 v1 路由。

## Roadmap Strategy

优先级按“已稳定、最容易删、最能减少历史包袱”排序：

1. 先删除前端、RAG、admin 已经稳定切走的 v1 动态接口。
2. 再升级搜索和标签，因为它们影响游戏发现体验，也会影响推荐。
3. 然后升级评论和推荐，保证详情页交互可以完全摆脱 v1。
4. 再升级创作者和抽奖，这两块更多是运营能力，风险可控但需要注意邮件、Redis 临时 key 和外链。
5. 最后清理旧包、旧 Redis key、旧 Swagger 文档和旧动态表读取路径。

每个阶段必须满足：

- 不新增长期兼容路由。
- 不把 v2 service 反向依赖 v1 controller。
- 不让公开响应返回 raw snapshot、上游错误体或过大的 HTML。
- 不把 HK 当作 CN 价格 fallback。
- 不让在线人数失败覆盖最近一次成功结果。

## Version Plan

### v2.1.0 - Remove Stable V1 Dynamic Routes

**Status:** Completed
**Scope:** Architecture / Stability / Maintenance
**Goal:** 删除已经稳定切到 v2 的 v1 动态接口，让游戏动态数据只剩一个主线。

#### Focus

- v1 动态路由移除
- 旧动态 service/dao 依赖收缩
- 前端/RAG/admin 调用确认
- 生产回滚边界说明

#### Tasks

- [x] 从 `gameApi` 中移除 `/info`、`/info/list`、`/info/main`、`/panel/main`、`/update/latest`、`/update/latest/more`。
- [x] 从 `gameApi` 中移除 `/sync/list`、`/sync/info`、`/sync/news`、`/sync/creators`。
- [x] 确认 `gofurry-nav-web` 游戏主线调用已经使用 `/api/v2/game/*`，并清理残留的旧动态接口包装函数。
- [x] 确认 `gofurry-rag` `sync_game_base_url` 使用 `/api/v2`。
- [x] 确认 `gofurry-admin` 采集观测通过 admin proxy 调用 `/api/v2/game/collect/*`。
- [x] 删除旧动态 v1 controller 方法，使这些入口不再被 router 或 Swagger 注释引用。
- [x] 更新 roadmap，明确这些 v1 动态路径已经移除。

#### Acceptance Criteria

- 上述 v1 动态路径返回 404 或不再注册。
- `/api/v2/game/*` 主线接口测试通过。
- `gofurry-nav-web` 游戏首页、详情页、新闻页可正常访问。
- RAG 游戏详情、新闻、创作者同步正常。
- admin 游戏采集观测正常。

#### Notes

本阶段不删除 `apps/game` 整个 v1 包，因为评论、标签、创作者等接口仍可能复用其中的站内表模型。只删除已经被 v2 替代的动态读取入口。

旧 service/dao 暂不在本阶段大面积删除，因为部分方法仍被定时任务、缓存刷新或后续迁移阶段间接使用。后续 `v2.6.0` 再做引用归零后的包级清理。

---

### v2.2.0 - Search And Tag V2 Mainline

**Status:** Planned  
**Scope:** User-facing / Architecture / Testing  
**Goal:** 将搜索和标签升级为 v2，并直接替换 v1 搜索/标签路由。

#### Focus

- 简易搜索
- 高级搜索
- 标签列表
- v2 read model 与站内运营表融合

#### Tasks

- [ ] 新增 `POST /api/v2/game/search/simple`，返回 v2 卡片字段：`id`、`appid`、`name`、`summary`、`header_url`、`capsule_url`、`tags`。
- [ ] 新增 `POST /api/v2/game/search/page`，支持分页、关键词、标签、发布时间、更新时间、评分排序、更新时间排序。
- [ ] 新增 `GET /api/v2/game/tags`，替代 `/api/v1/game/tag/list`。
- [ ] 搜索文本优先使用 v2 cleaned 字段：站内名、站内简介、Steam 本地化简介、开发商、发行商、标签。
- [ ] 前端搜索页和侧边栏搜索切到 v2。
- [ ] 移除 `/api/v1/game/search/simple`、`/api/v1/game/search/page`、`/api/v1/game/tag/list`。
- [ ] 补 DAO/service 单元测试，覆盖语言回退、标签组合、分页上限和空结果。

#### Acceptance Criteria

- 搜索页只请求 `/api/v2/game/search/*`。
- 标签筛选结果和当前 v1 行为等价或更准确。
- 搜索响应不再暴露 v1 独有字段名。
- v1 搜索和标签路由已移除。

#### Notes

搜索可以继续基于 PostgreSQL `ILIKE` 起步。全文索引、向量搜索或 RAG 辅助搜索不放进本阶段，避免把替换任务变成新系统建设。

---

### v2.3.0 - Review And Recommendation V2 Mainline

**Status:** Planned  
**Scope:** User-facing / Stability / Security/Safety  
**Goal:** 将详情页评论和推荐切到 v2，彻底移除详情页对 v1 的依赖。

#### Focus

- 评论读取
- 匿名评论提交
- 最新评论
- 详情页推荐
- 随机游戏

#### Tasks

- [ ] 新增 `GET /api/v2/game/reviews`，替代 `/api/v1/game/remark`。
- [ ] 新增 `POST /api/v2/game/reviews/anonymous`，替代 `/api/v1/game/review/anonymous`。
- [ ] 新增 `GET /api/v2/game/reviews/latest`，替代 `/api/v1/game/review/latest`。
- [ ] 新增 `GET /api/v2/game/recommend/similar`，替代 `/api/v1/game/recommend/CBF`。
- [ ] 新增 `GET /api/v2/game/recommend/random`，替代 `/api/v1/game/recommend/random`。
- [ ] 推荐算法读取 v2 结构化字段和站内标签，避免继续依赖旧静态特征拼装。
- [ ] 匿名评论保留现有风控：IP、User-Agent、频率限制、输入长度限制。
- [ ] 前端详情页评论区、推荐栏、随机游戏入口切到 v2。
- [ ] 移除对应 v1 路由。
- [ ] 补测试覆盖评论提交校验、评分统计、推荐空结果、随机游戏只返回可公开游戏。

#### Acceptance Criteria

- 游戏详情页不再请求任何 `/api/v1/game/*`。
- 评论读写和推荐接口均在 `/api/v2/game/*` 下。
- 评论和推荐的响应字段稳定，前端不需要 v1 -> v2 适配层。
- v1 评论和推荐路由已移除。

#### Notes

本阶段不要求重做复杂推荐系统。先保证功能替换和数据来源统一，算法升级可以后置。

---

### v2.4.0 - Creator V2 Mainline

**Status:** Planned  
**Scope:** User-facing / Maintenance / Documentation  
**Goal:** 将创作者列表和 RAG 创作者同步统一到 v2，移除 v1 创作者入口和旧 Redis 聚合缓存依赖。

#### Focus

- 创作者公开列表
- 创作者双语字段
- RAG sync creators 一致性
- 旧缓存 key 清理

#### Tasks

- [ ] 新增 `GET /api/v2/game/creators`，替代 `/api/v1/game/creator`。
- [ ] 保持 `GET /api/v2/game/sync/creators` 作为 RAG 同步入口。
- [ ] 公开创作者列表直接读取 PostgreSQL，不依赖 `game-creator:list` 旧 Redis 聚合。
- [ ] 前端创作者页切到 `/api/v2/game/creators`。
- [ ] 移除 `/api/v1/game/creator`。
- [ ] 清理或停用旧 `game-creator:list` 缓存写入和读取。
- [ ] 补测试覆盖 `zh/en` 名称和简介回退、空链接、空联系方式。

#### Acceptance Criteria

- 创作者公开页和 RAG 创作者同步均使用 v2。
- v1 创作者路由已移除。
- 旧 Redis 创作者聚合 key 不再是功能依赖。

---

### v2.5.0 - Prize V2 Mainline

**Status:** Planned  
**Scope:** User-facing / Security/Safety / Stability  
**Goal:** 将抽奖接口升级到 v2，并清理硬编码旧域名和 v1 激活链接。

#### Focus

- 抽奖信息读取
- 抽奖参与
- 邮件激活
- Redis 临时参与 key
- 安全和滥用控制

#### Tasks

- [ ] 新增 `GET /api/v2/game/prizes`，替代 `/api/v1/game/prize/info`。
- [ ] 新增 `POST /api/v2/game/prizes/participation`，替代 `/api/v1/game/prize/participation`。
- [ ] 新增 `GET /api/v2/game/prizes/participation/activation`，替代 `/api/v1/game/prize/participation/activation`。
- [ ] 将邮件中的激活链接从硬编码 `/api/v1/game/prize/participation/activation` 改为配置化前端 URL 或 v2 URL。
- [ ] 统一 Redis 临时 key 命名，例如 `game:v2:prize:{prize_id}:{key}`。
- [ ] 前端抽奖页和激活页切到 v2。
- [ ] 移除 v1 抽奖路由。
- [ ] 补测试覆盖重复参与、过期活动、错误 key、激活幂等、邮件链接生成。

#### Acceptance Criteria

- 抽奖完整流程只使用 `/api/v2/game/prizes/*`。
- 邮件和前端跳转不再包含 v1 API URL。
- Redis 临时 key 有明确 TTL 和命名空间。
- v1 抽奖路由已移除。

#### Notes

抽奖涉及邮件和真实用户输入，替换时必须先在开发环境完整跑通参与、收信、激活、重复提交和过期场景。

---

### v2.6.0 - Legacy Package And Data Cleanup

**Status:** Planned  
**Scope:** Architecture / Maintenance / Documentation  
**Goal:** 在所有调用方完成 v2 后，删除 v1 历史包袱和旧动态数据依赖。

#### Focus

- v1 package 清理
- 旧 Redis key 清理
- 旧动态表读取移除
- 文档和 Swagger 收敛

#### Tasks

- [ ] 删除不再被引用的 `apps/game/controller` 动态方法。
- [ ] 删除不再被引用的 `apps/game/service` 动态读取逻辑。
- [ ] 删除不再被引用的旧 DAO 方法。
- [ ] 保留或迁移仍有价值的站内表 model，避免误删 `gfg_game`、`gfg_tag`、`gfg_game_comment`、`gfg_game_creator`、`gfg_prize`。
- [ ] 移除旧 Redis key 依赖：`game-info:*`、`game-panel:*`、`game-update:*`、`game-creator:list` 等。
- [ ] 标记旧动态表为归档数据，不再从公开接口读取。
- [ ] 更新 Swagger，删除 v1 game API 文档。
- [ ] 更新部署和运维文档，说明 v2 collector 数据是唯一动态事实来源。

#### Acceptance Criteria

- `routers/url.go` 不再注册 `/api/v1/game/*`。
- `rg "/api/v1/game"` 只在历史说明或迁移文档中出现。
- 公开游戏页面、RAG、admin 均通过 v2 消费数据。
- `go test ./...` 通过。

## Short-Term Direction

下一步优先做 `v2.1.0`：删除已经稳定切走的 v1 动态接口。这个阶段收益最大、风险最低，因为前端主页面、RAG 和 admin 已经验证了 v2 主线。

## Medium-Term Direction

`v2.2.0` 到 `v2.4.0` 逐步把搜索、标签、评论、推荐、创作者迁移到 v2。每次迁移都以“前端切换 + 后端删除 v1 路由”为完成标准，不引入长期双栈。

## Long-Term Direction

`v2.5.0` 和 `v2.6.0` 完成抽奖流程 v2 化和历史包袱清理。最终目标是让 game backend 的公开游戏模块只保留 `/api/v2/game/*`，旧动态表、旧 Redis 聚合 key 和旧 v1 controller 不再参与线上功能。

## Risks

- 搜索和推荐升级时，v2 cleaned text 与旧站内简介可能存在排序差异，需要接受合理差异。
- 抽奖涉及邮件和 Redis 临时 key，不能只做路由重命名，必须验证完整用户流程。
- 删除 v1 路由前，需要确认生产前端、admin、RAG 配置都已经更新。
- 旧 Swagger 或文档如果不清理，会误导后续维护。
- 清理旧动态表读取时不要误删站内运营表，它们仍然是 v2 的一部分。
