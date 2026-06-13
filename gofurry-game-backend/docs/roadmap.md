# GoFurry Game Backend Roadmap

## Current Position

游戏模块已经完成 collector v2、backend v2 read model、admin 采集观测、RAG sync v2、前端主页面 cutover。`gofurry-nav-web` 当前游戏首页、详情页、新闻页、更多新闻页和 sitemap 游戏 URL 来源已经稳定消费 `/api/v2/game/*`，可以认为“动态游戏数据主线”已经切到 v2。

现阶段目标不再是继续维护 v1/v2 双栈。公开 game API 已经切到 `/api/v2/game/*`，后续重点是删除不再被引用的旧包、旧 Swagger、旧 Redis key 和历史文档说明，避免历史包袱继续扩散。

## Stable V2 Mainline

以下接口已经是主线，应继续保留并作为公开合同维护：

- `GET /api/v2/game/list`
- `GET /api/v2/game/info`
- `GET /api/v2/game/news`
- `GET /api/v2/game/news/latest`
- `GET /api/v2/game/panel/main`
- `GET /api/v2/game/recommend/random`
- `GET /api/v2/game/recommend/similar`
- `GET /api/v2/game/reviews`
- `POST /api/v2/game/reviews/anonymous`
- `GET /api/v2/game/reviews/latest`
- `GET /api/v2/game/creators`
- `GET /api/v2/game/prizes`
- `POST /api/v2/game/prizes/participation`
- `GET /api/v2/game/prizes/participation/activation`
- `POST /api/v2/game/search/simple`
- `POST /api/v2/game/search/page`
- `GET /api/v2/game/tags`
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

已经移除的 v1 动态、搜索和低风险交互接口：

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
- `GET /api/v1/game/creator`
- `GET /api/v1/game/remark`
- `GET /api/v1/game/tag/list`
- `GET /api/v1/game/recommend/random`
- `GET /api/v1/game/recommend/CBF`
- `POST /api/v1/game/search/simple`
- `POST /api/v1/game/search/page`
- `POST /api/v1/game/review/anonymous`
- `GET /api/v1/game/review/latest`
- `POST /api/v1/game/prize/participation`
- `GET /api/v1/game/prize/participation/activation`
- `GET /api/v1/game/prize/info`

当前已无需要长期保留的 v1 game 公开路由。后续重点转向旧包、旧 Redis key、Swagger 和历史文档清理。

## Roadmap Strategy

优先级按“已稳定、最容易删、最能减少历史包袱”排序。前五个阶段已经完成，后续集中到包级清理：

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

**Status:** Completed
**Scope:** User-facing / Architecture / Testing
**Goal:** 将搜索和标签升级为 v2，并直接替换 v1 搜索/标签路由。

#### Focus

- 简易搜索
- 高级搜索
- 标签列表
- v2 read model 与站内运营表融合

#### Tasks

- [x] 新增 `POST /api/v2/game/search/simple`，保持前端兼容字段，同时由 v2 read model 数据源生成。
- [x] 新增 `POST /api/v2/game/search/page`，支持分页、关键词、标签、发布时间、更新时间、评分排序、评论数排序和更新时间排序。
- [x] 新增 `GET /api/v2/game/tags`，替代 `/api/v1/game/tag/list`。
- [x] 搜索文本优先使用 v2 cleaned 字段：站内名、站内简介、Steam 本地化简介、开发商、发行商、标签。
- [x] 前端搜索页和侧边栏搜索切到 v2。
- [x] 移除 `/api/v1/game/search/simple`、`/api/v1/game/search/page`、`/api/v1/game/tag/list`。
- [x] 补 service 单元测试，覆盖语言归一化、分页上限和简单搜索映射。

#### Acceptance Criteria

- 搜索页只请求 `/api/v2/game/search/*`。
- 标签筛选结果和当前 v1 行为等价或更准确。
- 搜索响应当前保留前端兼容字段，底层不再读取旧动态表；后续搜索 UI 改版时再收敛为纯 v2 卡片字段。
- v1 搜索和标签路由已移除。

#### Notes

搜索继续基于 PostgreSQL `ILIKE` 起步。全文索引、向量搜索或 RAG 辅助搜索不放进本阶段，避免把替换任务变成新系统建设。

旧 `apps/search` service/dao 和 `apps/game` tag service/dao 暂不做包级删除，统一留到 `v2.6.0` 引用归零后清理。

---

### v2.3.0 - Review And Random Recommendation V2 Mainline

**Status:** Completed
**Scope:** User-facing / Stability / Security/Safety
**Goal:** 将详情页评论、最新评论和随机游戏切到 v2，先移除低风险 v1 交互入口。

#### Focus

- 评论读取
- 匿名评论提交
- 最新评论
- 随机游戏

#### Tasks

- [x] 新增 `GET /api/v2/game/reviews`，替代 `/api/v1/game/remark`。
- [x] 新增 `POST /api/v2/game/reviews/anonymous`，替代 `/api/v1/game/review/anonymous`。
- [x] 新增 `GET /api/v2/game/reviews/latest`，替代 `/api/v1/game/review/latest`。
- [x] 新增 `GET /api/v2/game/recommend/random`，替代 `/api/v1/game/recommend/random`。
- [x] 匿名评论保留现有风控：IP、User-Agent、频率限制、输入长度限制。
- [x] 前端详情页评论区、最新评论、随机游戏入口切到 v2。
- [x] 移除对应 v1 路由。
- [x] 补 service 测试覆盖评论 IP 脱敏和随机 ID 读取边界。
- [x] `GET /api/v2/game/recommend/similar` 已进入 `v2.3.1` 单独完成。

#### Acceptance Criteria

- 游戏详情页评论读写不再请求 v1。
- 评论读写、最新评论和随机游戏均在 `/api/v2/game/*` 下。
- 评论响应字段保持前端兼容，不需要额外适配层。
- v1 评论、最新评论和随机游戏路由已移除。

#### Notes

本阶段先完成低风险评论与随机推荐。复杂相似推荐已在 `v2.3.1` 单独设计和实现，旧 `/api/v1/game/recommend/CBF` 已在 `v2.3.2` 删除。

---

### v2.3.1 - Similar Recommendation V2 Design

**Status:** Completed
**Scope:** Architecture / Performance / Recommendation Quality
**Goal:** 设计并实现 `GET /api/v2/game/recommend/similar`，让详情页相似推荐从 v1 CBF 切到 v2 hybrid read model。

#### Focus

- v2 相似度特征来源
- 标签、开发商、发行商、价格、在线人数的权重
- PostgreSQL 预计算表与开发环境即时回填
- 冷启动和空结果策略
- 前端推荐栏合同

#### Tasks

- [x] 选型为 v2 hybrid CBF：标签 Weighted Jaccard + 创作者 + 文本 + 平台 + 价格 + 活跃度。
- [x] 新增 `gfg_game_v2_recommendations` 预计算表，记录 `score`、`display_score`、`reason_json`、`algorithm_version`。
- [x] 新增 `GET /api/v2/game/recommend/similar`，只返回已有 v2 details 的游戏。
- [x] API 优先读取预计算结果；单个游戏缺失时即时计算 top 64 并写回，便于开发环境立即验证。
- [x] 响应字段改为纯 v2 合同：`score`、`display_score`、`reasons`、`header_url`、`capsule_url`、`tags`、`price`、`online_count`。
- [x] 前端详情页相似推荐切到 `/api/v2/game/recommend/similar`。
- [x] 在代码中写入中文算法选型说明、权重设计和后续扩展边界。

#### Acceptance Criteria

- 详情页不再请求 `/api/v1/game/recommend/CBF`。
- 推荐接口不依赖旧 `apps/recommend` CBF 运行时计算。
- 新增 SQL migration 可独立执行，不破坏现有 game v2 表。
- 缺少预计算数据时，接口能为单个源游戏计算并回填。
- 推荐结果带可解释理由，方便前端展示和后续人工调权。

#### Notes

当前主线已经切到 v2，不再新增兼容层。后续可以增加 admin 手动重算或 collector 定时重算入口，把即时回填降级为纯兜底。

---

### v2.3.2 - Remove Legacy CBF Route

**Status:** Completed
**Scope:** Maintenance / Cleanup
**Goal:** 当前端详情页相似推荐稳定消费 v2 后，移除 `/api/v1/game/recommend/CBF` 和旧 `apps/recommend` CBF 实现。

#### Tasks

- [x] 线上确认详情页只请求 `/api/v2/game/recommend/similar`。
- [x] 移除 `/api/v1/game/recommend/CBF` 路由。
- [x] 删除不再被引用的 `apps/recommend` CBF controller、service 和 dao。
- [x] 删除 CBF 专用 model：`ContentSimilarities`、`GameRecommendVo`、`GameTemp`。
- [x] 清理旧 Redis key 功能依赖：`recommend:tag-mapping`、`recommend:tag-ids` 不再被代码读取或写入。

#### Acceptance Criteria

- `/api/v1/game/recommend/CBF` 不再注册。
- 后端不再编译旧 CBF controller/service/dao。
- 前端详情页继续使用 `/api/v2/game/recommend/similar`。
- `go test ./...` 通过。

#### Notes

`apps/recommend/models` 暂时保留 `gfg_tag` 与 `gfg_tag_map` 模型，因为旧 `apps/game/dao` 仍引用这些站内标签模型。后续 `v2.6.0` 做包级清理时，再把标签模型迁移到更合适的位置并删除空壳 recommend 包。

---

### v2.4.0 - Creator V2 Mainline

**Status:** Completed
**Scope:** User-facing / Maintenance / Documentation  
**Goal:** 将创作者列表和 RAG 创作者同步统一到 v2，移除 v1 创作者入口和旧 Redis 聚合缓存依赖。

#### Focus

- 创作者公开列表
- 创作者双语字段
- RAG sync creators 一致性
- 旧缓存 key 清理

#### Tasks

- [x] 新增 `GET /api/v2/game/creators`，替代 `/api/v1/game/creator`。
- [x] 保持 `GET /api/v2/game/sync/creators` 作为 RAG 同步入口。
- [x] 公开创作者列表直接读取 PostgreSQL，不依赖 `game-creator:list` 旧 Redis 聚合。
- [x] 前端创作者页切到 `/api/v2/game/creators`。
- [x] 移除 `/api/v1/game/creator`。
- [x] 清理或停用旧 `game-creator:list` 缓存写入和读取。
- [x] 补测试覆盖语言归一化、HTML 清洗、链接和联系方式解析。

#### Acceptance Criteria

- 创作者公开页和 RAG 创作者同步均使用 v2。
- v1 创作者路由已移除。
- 旧 Redis 创作者聚合 key 不再是功能依赖。
- `go test ./...` 和前端 typecheck 通过。

#### Notes

公开创作者列表和 RAG 同步创作者当前复用同一套 v2 PostgreSQL 读取与映射逻辑。旧 `apps/game/dao/gameCreatorDao.go`、v1 creator controller、旧 creator Redis 缓存任务已经删除。

---

### v2.5.0 - Prize V2 Mainline

**Status:** Completed
**Scope:** User-facing / Security/Safety / Stability  
**Goal:** 将抽奖接口升级到 v2，并清理硬编码旧域名和 v1 激活链接。

#### Focus

- 抽奖信息读取
- 抽奖参与
- 邮件激活
- Redis 临时参与 key
- 安全和滥用控制

#### Tasks

- [x] 新增 `GET /api/v2/game/prizes`，替代 `/api/v1/game/prize/info`。
- [x] 新增 `POST /api/v2/game/prizes/participation`，替代 `/api/v1/game/prize/participation`。
- [x] 新增 `GET /api/v2/game/prizes/participation/activation`，替代 `/api/v1/game/prize/participation/activation`。
- [x] 将邮件中的激活链接从硬编码 `/api/v1/game/prize/participation/activation` 改为 v2 URL。
- [x] 统一 Redis 临时 key 命名：`game:v2:prize:email:{prize_id}:{email}`、`game:v2:prize:participation:{prize_id}:{key}`。
- [x] 前端抽奖页和激活提交切到 v2。
- [x] 移除 v1 抽奖路由。
- [x] 补测试覆盖 v2 Redis key 和邮件激活链接生成，避免回退到 v1 URL。

#### Acceptance Criteria

- 抽奖完整流程只使用 `/api/v2/game/prizes/*`。
- 邮件和前端跳转不再包含 v1 API URL。
- Redis 临时 key 有明确 TTL 和命名空间。
- v1 抽奖路由已移除。

#### Notes

抽奖公开入口已经切到 `/api/v2/game/prizes/*`。历史抽奖数据仍使用既有 `gfg_prize`、`gfg_prize_member` 和 `prize:history` 历史展示缓存；本阶段只迁移用户参与临时 key 和公开 API 路由，不强行改写定时开奖历史缓存。

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

下一步优先做 `v2.6.0`：清理旧包、旧 Redis key、旧 Swagger 文档和历史 v1 说明，确保公开游戏模块只剩 v2 主线。

## Medium-Term Direction

`v2.6.0` 重点做引用归零后的包级清理。清理时要保留仍有业务价值的站内运营表模型，避免把 `gfg_game`、`gfg_tag`、`gfg_game_comment`、`gfg_game_creator`、`gfg_prize` 误删。

## Long-Term Direction

`v2.6.0` 完成历史包袱清理后，game backend 的公开游戏模块应只保留 `/api/v2/game/*`，旧动态表、旧 Redis 聚合 key 和旧 v1 controller 不再参与线上功能。

## Risks

- 搜索和推荐升级时，v2 cleaned text 与旧站内简介可能存在排序差异，需要接受合理差异。
- 抽奖涉及邮件和 Redis 临时 key，不能只做路由重命名，必须验证完整用户流程。
- 删除 v1 路由前，需要确认生产前端、admin、RAG 配置都已经更新。
- 旧 Swagger 或文档如果不清理，会误导后续维护。
- 清理旧动态表读取时不要误删站内运营表，它们仍然是 v2 的一部分。
