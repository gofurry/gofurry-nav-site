# GoFurry Game Backend V2 Roadmap

## 当前结论

游戏采集器 v2 已经进入主线，后端下一步应该围绕 collector v2 的 PostgreSQL 结构化数据建立新的读模型和 API 合同。后端 v2 不应该重新理解 Steam raw payload，也不应该继续把 v1 的 `gfg_game_record`、`gfg_game_news`、`gfg_game_player_count` 作为动态游戏数据主来源。

本轮后端迭代目标：

- 使用 collector v2 写入的 PostgreSQL 表作为事实来源。
- Redis 只作为热缓存和高频聚合缓存，未命中时回源 PostgreSQL。
- 保留 `gfg_game`、评论、标签、创作者、抽奖等站内运营表。
- 新增 `/api/v2/game/*`，让 `gofurry-nav-web`、`gofurry-admin`、`gofurry-rag` 有明确迁移目标。
- 在前端、admin、RAG 全部切换完成前，不删除 `/api/v1/game/*`。

## 现状边界

### gofurry-game-backend

当前公开游戏接口仍以 `/api/v1/game/*` 为主，但 `/api/v2/game/list`、`/api/v2/game/info`、`/api/v2/game/news`、`/api/v2/game/news/latest` 已经开放为第一批 public v2 API。v2 已经有独立契约文档 `docs/game-v2-backend-contract.md`，并已完成 read model foundation：新增 v2 table model、PostgreSQL DAO 聚合入口、service 读模型转换、controller 和 router。`panel/main`、`sync/*`、`collect/*` 仍在后续里程碑。

需要继续复用的站内能力：

- `gfg_game`：站内游戏主档案、权重、资源、社群、链接入口。
- `gfg_tag`、`gfg_tag_map`：标签和分类。
- `gfg_game_comment`：评论和评分。
- `gfg_game_creator`：创作者资料。
- `gfg_game_prize`：抽奖活动。
- 推荐、搜索、评论、抽奖等非 Steam 动态采集能力。

### gofurry-nav-web

前端当前仍通过 `NUXT_PUBLIC_GAME_API_BASE` 指向 `/api/v1`，游戏服务调用包括：

- `/game/info/main`
- `/game/panel/main`
- `/game/update/latest`
- `/game/update/latest/more`
- `/game/search/simple`
- `/game/search/page`
- `/game/info`
- `/game/recommend/*`
- `/game/review/*`
- `/game/tag/list`
- `/game/creator`
- `/game/prize/*`

后端 v2 首先要覆盖游戏首页、列表、详情、新闻、聚合面板这些“Steam 动态数据”页面。搜索、推荐、评论、抽奖、创作者可以先继续走 v1 或后续单独迁移，避免一次性改动面过大。

### gofurry-admin

admin 当前直接管理旧游戏运营表，包括游戏主档案、评论、创作者、抽奖、标签、标签映射。它已经连接游戏 PostgreSQL，不需要 MongoDB。

后端 v2 对 admin 的价值主要是新增采集观测和诊断能力：

- 最近采集批次。
- 单游戏采集结果。
- 当前动态详情是否完整。
- 新闻、价格、媒体、在线人数最近更新时间。
- 失败 appid、失败原因、重试建议。

admin 不应该直接暴露 Steam raw snapshot 给普通运营页面；raw snapshot 只适合受保护的调试页或临时排查。

### gofurry-rag

RAG 当前已经按 `/game/sync/list`、`/game/sync/info`、`/game/sync/news`、`/game/sync/creators` 拉取游戏知识。v2 后端要提供兼容 RAG 的稳定同步接口，但内容应变得更干净：

- 使用 cleaned/plain text 作为主要同步内容。
- 不默认返回 raw payload。
- 新闻统一走 Store events。
- 支持 `zh` / `en`。
- 支持分页或 `updated_since`，避免未来全量同步压力过大。

## 总体策略

1. 后端先新增 v2 读模型和 v2 API，不影响现有 v1。
2. 路由方式参考 nav 后端 v2：保留现有 `/api/v1/game/*`，新增独立 `/api/v2/game/*` 分组；建议新增 `routers/url_v2.go` 或同等职责文件注册 `gameV2Api(v2.Group("/game"))`，避免在 v1 router 里混入 v2 逻辑。
3. `gofurry-nav-web` 先切游戏首页、详情、新闻到 v2；搜索、推荐、评论、抽奖按风险后置。
4. `gofurry-admin` 增加 collector v2 观测页，保留原有运营管理。
5. `gofurry-rag` 切到 v2 sync 接口，优先使用 cleaned/plain text。
6. 当前端、admin、RAG 都稳定后，再冻结 v1 动态数据路径，逐步清理旧表读取和旧 Redis key。

## v2.0.0-alpha.1 - Backend V2 Read Model Foundation

状态：已完成。

目标：建立后端读取 collector v2 数据的基础模型，不先追求前端完整切换。

后端改动：

- 新增 v2 table model，覆盖：
  - `gfg_game_v2_details`
  - `gfg_game_v2_localized_details`
  - `gfg_game_v2_prices`
  - `gfg_game_v2_media`
  - `gfg_game_v2_requirements`
  - `gfg_game_v2_news`
  - `gfg_game_v2_player_counts`
  - `gfg_game_v2_collect_runs`
  - `gfg_game_v2_collect_task_results`
- 新增 v2 DAO，按 `game_id` / `appid` 定位游戏，按 `lang` 聚合详情；`region` 只用于价格选择，不进入详情缓存 key。
- 明确语言策略：当前只支持 `zh` / `en`，默认 `zh`，非法语言回退 `zh`。
- 明确价格策略：默认展示 `CN`；当 `CN` 记录缺失，或记录表现为非免费但无币种、无格式化价格时，视为 unavailable，不把 `HK` 作为中国区 fallback。
- 明确在线人数策略：公开接口只读取最近一次 `status='success'` 的结果，失败状态进入后台观测。
- 新增单元测试覆盖 v2 聚合、语言回退、价格 unavailable、在线人数失败不覆盖成功值。

实际完成：

- 新增 `apps/game/v2/models`，覆盖 collector v2 所有主表模型，并定义 `GameV2DetailReadModel`、价格、媒体、新闻、在线人数、站内信息等公开读模型结构。
- 新增 `apps/game/v2/dao`，按 `game_id` / `appid` 聚合 `gfg_game` 站内主档案、v2 详情、本地化详情、价格、媒体、系统需求、新闻、最近成功在线人数和标签。
- 新增 `apps/game/v2/service`，将 DAO 聚合结果转换为稳定读模型，固化 `zh/en` 语言归一、中文回退、默认 `CN` 价格、不使用 `HK` fallback、在线人数缺失为 `unknown`。
- service 层通过接口注入 DAO，避免单元测试依赖数据库配置，也为 alpha.2 的 controller 显式注入 `dao.NewReadModelDAO()` 留出清晰边界。
- 新增单元测试覆盖语言回退、CN 价格 unavailable、不回退 HK、最近成功在线人数和非法语言默认回退。

验收标准：

- 已满足：可以在不依赖旧动态表的情况下聚合一个游戏的 v2 详情。
- 已满足：v2 DAO 不读取 `gfg_game_record`、`gfg_game_news`、`gfg_game_player_count`。
- 已满足：当前 read model foundation 只依赖 PostgreSQL，不依赖 Redis；Redis 缓存读取将在 public API v2 阶段按接口级缓存策略补充。

## v2.0.0-alpha.2 - Public Game API V2

状态：已完成。

目标：提供前端可开始接入的公开 v2 API。

路由方式：

- 保持现有 v1 注册方式不变：`gameApi(v1.Group("/game"))`。
- 新增 v2 注册入口：`gameV2Api(v2.Group("/game"))`，最终路径为 `/api/v2/game/*`。
- v2 controller / service / dao 独立命名或独立包边界，避免 v1 controller 继续膨胀。
- v2 不复用 v1 响应结构硬塞字段；公开响应以 `docs/game-v2-backend-contract.md` 为准。

新增接口：

- `GET /api/v2/game/list`
- `GET /api/v2/game/info`
- `GET /api/v2/game/news`
- `GET /api/v2/game/news/latest`

实际完成：

- 在 router 中新增 `/api/v2/game/*` 分组，保留现有 `/api/v1/game/*` 不变。
- 新增 v2 controller，开放：
  - `GET /api/v2/game/list`
  - `GET /api/v2/game/info`
  - `GET /api/v2/game/news`
  - `GET /api/v2/game/news/latest`
- 新增 v2 list/news DAO 方法，列表从 `gfg_game` + collector v2 表聚合轻量卡片数据；新闻从 `gfg_game_v2_news` 读取 Store events cleaned 内容，并补充站内游戏名和 header。
- 新增 v2 service list/news 方法，统一 `zh/en`、`CN`、分页上限和公开新闻字段。
- 公开接口参数设置保守上限：列表和新闻 `limit` 默认 20，最大 100；非法语言回退 `zh`。
- 本阶段公开 API 先以 PostgreSQL read model 为主，避免引入不稳定缓存行为；Redis read-through 和聚合缓存放到后续面板/高频接口阶段补齐。

响应原则：

- 列表接口返回轻量 `GameV2ListItem`。
- 详情接口返回稳定 `GameV2Detail`。
- 新闻接口返回 `GameV2NewsItem`，包含 `summary`、`plain_text`、必要时包含安全的 `html`。
- raw snapshot、上游错误、payload hash 不进入公开响应。
- 所有列表接口限制 `limit` 上限，防止前端误请求拖垮后端。

缓存策略：

- 与 collector v2 当前实现保持一致，优先读取这些已由 collector 刷新的热缓存：
  - `game:v2:details:{game_id}:{lang}`
  - `game:v2:prices:{game_id}`
  - `game:v2:media:{game_id}`
  - `game:v2:players:{game_id}:current`
- `game:v2:news:{game_id}:{lang}`
- collector 当前不会写入全站 latest news 缓存；`/api/v2/game/news/latest` 应由 backend 从 `gfg_game_v2_news` 查询并写入 backend 自己维护的聚合缓存，例如 `game:v2:news:latest:{lang}`。
- TTL 与 collector 当前实现对齐：详情、价格、媒体、单游戏新闻 7 天；当前在线人数 3 小时；backend 自己维护的 latest news 聚合缓存建议 3 小时或更短。

当前实现说明：

- alpha.2 已开放公开 API，但暂未做 Redis read-through。原因是当前首要目标是验证 v2 响应合同和 PostgreSQL 回源能力，缓存所有权和聚合 key 会在 `v2.0.0-alpha.3 - Frontend Panel Contract` 及后续高频接口里统一处理。

验收标准：

- 已满足：`gofurry-nav-web` 可以用 v2 `list/info/news/news/latest` 接口开始完成游戏详情页和新闻页基础渲染。
- 已满足：新闻正文读取 collector v2 cleaned 字段，前端不需要再承担 BBCode/HTML 清洗。
- 已满足：公开响应不包含 raw snapshot、上游错误体和 collect task 诊断字段；字段结构与 `docs/game-v2-backend-contract.md` 的公开模型保持同向。

## v2.0.0-alpha.3 - Frontend Panel Contract

目标：补齐游戏首页和游戏模块 v2 视觉改造需要的聚合数据。

新增或完善接口：

- `GET /api/v2/game/panel/main`

聚合内容：

- 最新入库或近期更新游戏。
- 热门在线人数排行。
- 免费游戏。
- 最高折扣。
- 低价分组。
- 最新新闻。
- 首页视觉资源：header、capsule、background、screenshots 中适合卡片展示的字段。

前端配合：

- `gofurry-nav-web` 新增 v2 service/types，不复用旧 v1 类型硬塞字段。
- 游戏首页可以先使用 v2 panel，详情页随后切换。
- 环境变量可新增 v2 base，例如 `NUXT_PUBLIC_GAME_V2_API_BASE`，稳定后再替换默认 `NUXT_PUBLIC_GAME_API_BASE`。
- 前端图片 alt 使用游戏名、新闻标题、站点名；纯装饰图保持 `alt=""`。

验收标准：

- 游戏首页不再依赖旧 Redis 聚合 key。
- 首页聚合接口一次请求即可满足首屏主要内容。
- 前端不用为缺失 CN 价格写 HK fallback。

## v2.0.0-alpha.4 - Admin Collection Observability

目标：让运营和排查能看到 collector v2 是否真的稳定，而不是只靠日志。

新增后端接口：

- `GET /api/v2/game/collect/status`
- `GET /api/v2/game/collect/runs`
- `GET /api/v2/game/collect/runs/:run_id`
- `GET /api/v2/game/collect/task-results`
- `GET /api/v2/game/collect/games/:id/status`

接口要求：

- 必须受后台鉴权保护，不作为公开 API。
- 支持按 task、status、appid、时间范围过滤。
- 返回 run 汇总、失败原因、最近成功时间、最近失败时间。
- raw snapshot 不默认返回；如需要调试，应单独做高权限接口并限制响应体大小。

admin 配合：

- `gofurry-admin` 增加“游戏采集状态”资源页。
- 支持查看最近采集批次、失败 appid、失败类型、单游戏动态数据完整度。
- 原有游戏主档案、评论、标签、创作者、抽奖管理继续保留，不和 v2 动态数据混在一个表单里。

验收标准：

- 能从 admin 判断一次全量采集是否完整。
- 能定位某个 appid 是详情失败、新闻失败、价格锁区，还是在线人数失败。
- collect runs 和 task results 的保留策略与 collector 配置一致。
- 当前 collector task result 不记录 Redis 写入失败；admin 第一版以 PostgreSQL run/task 观测为准，Redis key 仅作为热缓存命中和数据新鲜度辅助信号。后续如需要 Redis 写入诊断，再补 collector 观测字段。

## v2.0.0-alpha.5 - RAG Sync V2

目标：让 RAG 使用后端 v2 的 cleaned 内容，减少历史字段和 HTML 噪音。

新增或迁移接口：

- `GET /api/v2/game/sync/list`
- `GET /api/v2/game/sync/info`
- `GET /api/v2/game/sync/news`
- `GET /api/v2/game/sync/creators`

同步内容要求：

- `sync/list` 返回轻量摘要，支持 `lang`、`limit`、`offset` 或 cursor。
- `sync/info` 返回 cleaned/plain text 字段，包含游戏名、简介、详情、标签、开发商、发行商、平台、站内资源。
- `sync/news` 返回 Store events 新闻，正文优先 `plain_text`。
- `sync/creators` 可继续复用站内创作者表，但路径切到 v2。
- 增加 `updated_since` 能力，方便 RAG 做增量同步。

RAG 配合：

- `gofurry-rag` 将 `sync_game_base_url` 切到 `/api/v2`。
- RAG 客户端新增 v2 字段兼容，但保留旧字段解析直到切换完成。
- source id 保持稳定，例如 `game:{id}:{lang}`、`game_news:{event_id}:{lang}`。

验收标准：

- RAG 可以同步游戏详情、游戏新闻、创作者三类内容。
- 同步内容不包含 raw payload。
- 二次同步能通过 checksum 跳过未变化文档。

## v2.0.0-beta.1 - Frontend Cutover

目标：前端游戏模块主要页面切到 v2，验证真实用户体验。

切换范围：

- 游戏首页。
- 游戏详情页。
- 游戏新闻列表。
- 更多新闻页。
- 搜索结果中展示用字段。

暂不强制切换：

- 评论提交。
- 抽奖。
- 复杂推荐。
- 创作者运营页。
- v1 搜索、推荐、标签、评论、抽奖和创作者接口继续服务现有前端，等 games v2 主页面稳定后再评估是否单独迁移。

验证重点：

- SSR 下接口可用且不会因单个游戏异常导致页面 500。
- 图片字段足够支撑 games v2 视觉改造。
- CN 锁区价格显示为 unavailable 或无价格，不展示 HK 替代。
- 新闻和详情正文不会出现 BBCode 残留。
- v1/v2 数据差异有明确解释，不把差异当作线上故障。

验收标准：

- `gofurry-nav-web` 游戏首页和详情页默认请求 v2。
- v1 游戏动态数据路径可以进入冻结状态。
- 前端 v2 类型和后端 v2 响应字段一致。

## v2.0.0-rc.1 - Compatibility Freeze

目标：冻结 v2 API 合同，为 stable 做最后确认。

后端工作：

- 补齐 Swagger 或接口文档。
- 固定响应字段名，避免 stable 前继续变动。
- 补齐 DAO/service/controller 单元测试。
- 加入公开接口 smoke test。
- 明确 v1 动态路径 deprecation note。

跨项目确认：

- `gofurry-nav-web` 已完成主要游戏页面 v2 切换。
- `gofurry-admin` 已能查看 collector v2 采集状态。
- `gofurry-rag` 已切到 v2 sync，并完成至少一次全量同步和一次增量同步。

验收标准：

- collector v2 全量采集后，backend v2、frontend、admin、RAG 都能消费同一套数据。
- v1 动态表不再是公开游戏页面的主要数据源。
- stable 前没有必须修改数据库结构的阻塞项。

## v2.0.0 - Backend V2 Mainline Stable

目标：后端游戏模块 v2 成为主线。

稳定范围：

- `/api/v2/game/list`
- `/api/v2/game/info`
- `/api/v2/game/news`
- `/api/v2/game/news/latest`
- `/api/v2/game/panel/main`
- `/api/v2/game/sync/*`
- `/api/v2/game/collect/*`

v1 处理：

- `/api/v1/game/*` 暂时保留，服务旧入口和非动态功能。
- 冻结旧动态详情、旧新闻、旧在线人数读取路径。
- v1 搜索、推荐、评论、抽奖、标签、创作者等站内运营和交互接口继续保留，不因 v2 stable 立即删除。
- 后续单独计划清理 `gfg_game_record`、`gfg_game_news`、`gfg_game_player_count` 的依赖。

验收标准：

- 新游戏动态内容全部来自 collector v2 数据表。
- 前端主要游戏页面不依赖旧动态表。
- admin 能观测采集状态。
- RAG 能稳定同步 v2 cleaned 内容。

## 风险与约束

- v2 和 v1 并行期间最容易出现字段来源混用，service 层需要明确 `game/v2` 聚合边界。
- CN 价格缺失通常代表锁区或无区域价格，不应该把 HK 当作 fallback。
- 在线人数失败不能覆盖最近一次成功值，也不能简单展示为 0。
- Redis key 必须统一命名，避免后端和 collector 各写各的。
- collector 已写入的 key 必须以 `game:v2:details:{game_id}:{lang}`、`game:v2:prices:{game_id}`、`game:v2:media:{game_id}`、`game:v2:news:{game_id}:{lang}`、`game:v2:players:{game_id}:current` 为准；backend 自己新增的聚合 key 需要明确标注所有权。
- 公开接口不能返回 raw snapshot、上游错误体和过大的 HTML。
- RAG 同步接口要控制分页和响应体大小，否则未来游戏数增长后会拖慢同步。
- admin 采集观测接口必须受保护，避免暴露上游风控、失败模式和内部 appid 排查信息。

## 后续清理方向

v2 stable 后再评估以下清理，不放进第一阶段：

- 删除或归档旧动态表读取逻辑。
- 搜索和推荐从 v1 字段迁移到 v2 cleaned text 与结构化标签。
- 为 player counts 做 90 天趋势图接口。
- 为价格历史做趋势图或低价区比较，但新增地区前要先评估全量采集耗时。
- 为 admin 增加受保护的单游戏重采集触发能力。
