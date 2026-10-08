# Game Collection：数据域与公开时间线

Stage A（#140-A）建立人工策展游戏分区及 Game Backend 公开读模型。
当前 #144-A 将成员扩展为人工定义标签规则、固定与排除；公开顺序仍由 Canonical First Available / Release Facts 派生。
Collection 不是 Tag、Showcase、Recommendation；成员关系与 `gfg_game.groups` 社群字段独立，Detail 仅将社群入口数量作为卡片元数据。

## Stage C：Public Discovery 与 Timeline

Nav Web 提供 `/games/collections` 和 `/games/collections/:code`（以及 `/en` 对应路由）。
首页通过一次 `Promise.allSettled` 并行读取 Home、Showcase、Collections Home；两个可选 slice SSR 各有 1 秒预算、retry=0。
Collections Home 永远请求 SFW；SSR 不可用的 Showcase/Collections 在 mounted 后各补读一次（8 秒、retry=0），失败静默隐藏。
SSR 成功（含合法空结果）不补读；合法空 slots 在 Desktop 仍显示“全部分区”。语言变更和卸载取消/丢弃旧恢复，模式变更不请求 Home。
快捷入口仅在 xl 及以上 Sidebar 使用三列，直接继承每日一游的材质和 2.45rem 最小高度；小屏不保留快捷入口 DOM。
后端给出的非空 slots 按 slot 升序压缩展示，最后追加固定的全部分区入口。

Index 视觉以搜索框与高级筛选工具条开场，H1 仅供辅助技术读取，SEO title/description 保留。
卡片复用 Games Home 的透明边框、背景、阴影、hover 和圆角；0/1/2/3 张 preview 原样消费，不复制图片。
输入使用 IME-safe 的 350ms debounce；筛选弹窗仅编辑本地 Draft，Cancel 不提交，Apply 原子提交阶段与排序。
Detail 的单一时间线由后端排序，前端仅稳定分组：已发布、已发布但日期待考、NOW、未来、TBA、未知。
过去和未来各自复用 `serpentineSequence` 几何，900px 以上三列蛇形，以下单列；DOM 顺序始终不变。
日期待考、TBA、未知没有 connector，只有未知作品时不显示 NOW。NOW 取后端 `as_of_date`。
Detail 紧凑 Header 将标题/数量放同组，返回入口靠右，简介下一行。Timeline Card/NOW 使用 Games Home surface 和透明边框。
卡片固定两行标题/摘要空间；日期与最多两个截断标签同排，底部显示带完整可访问名称的评分/在线/社群数，无数据省略，保留 metadata 最小高度。
Connector 为 2px、2.5rem 的低强调动态虚线，left/right/down 跟随 chronology；小屏统一向下。Reduced motion 保留静态虚线，pointer-events:none，位于卡片后方。
显示保留 day/month/quarter/year 与 inferred 精度，不暴露内部 First Available 来源，也不在前端过滤 Adult Tag。

两个分区页面 SSR 一律 SFW。挂载读取本地模式：SFW 零补请求，NSFW 恰好一次刷新。
Detail 保持 `useGameCollectionModeRefresh`。Index 由 `useGameCollectionDiscovery` 单独拥有 committed q/phase/sort、mode、分页和 snapshot；所有搜索/筛选/模式变更共用 generation。
新请求期间保留 ready 卡片，成功后从 page 1 原子替换，失败保留旧内容并局部重试。Load More 携带相同 criteria，按 code 去重，过期响应不追加。模式不写 URL。
Detail 初始/刷新 404 使用真实 Nuxt 404；初始服务失败返回 HTTP 503 与可重试 surface，不伪装空分区。
adult-only 的 SFW Detail 仍为 200，零作品使用中性空态。所有图片复用 SteamAssetImage。

前序 Stage C 收口分别简化了 Home 读取与 Index/Detail 轻量投影，并将缓存 TTL 调整为 1 小时。当时的数据收口将内部缓存 revision 提升到 v2，Public schema_version 保持 1。#144-A 的新增结构与当前 v3 缓存见下文。Functional、固定环境 Visual 比对和维护者人工验收分别记录；
新增六张分区截图与首页四张变更须人工接受后才构成 Visual PASS，代码完成不代表 #140 已关闭。

2026-10-07 Detail Timeline Closure：新增 Detail-only 单次装饰批量读取，3/30 成员均为固定 8 次 SQL，保留原 chronology、SFW 和 Redis 1h 合同。
本地 Game Backend 全量测试/构建、隔离 PG18 集成、Nav Web lint/stylelint/style policy/typecheck/build、Unit 366、Nuxt 75、Collection Browser 32 项通过。
Pinned Linux 的 Detail 材质/几何/方向/reduced-motion 5 项通过；仅四张 Timeline 视觉生成审阅差异，未更新任何 Golden。维护者审计与人工视觉接受仍待完成。

2026-10-07 首页收口：Home 改为单条轻量 SQL；SSR 失败 slice 支持一次 mounted 恢复；快捷入口仅留在 Desktop Sidebar 并继承每日一游材质。
本地 Unit 366、Nuxt 71、Games Home/Showcase Browser 58 项及隔离 PG18 集成通过。桌面两张 Home 基线按本轮授权更新，移动端与 Index/Timeline 基线保持不变；维护者视觉审计仍待进行。

2026-10-06 本地验证：Unit 366、Nuxt 64、focused Browser 76、受影响的既有消费者回归 38 项通过；
lint/stylelint/style:policy/typecheck/build、SEO/Insights guards 与 repository policy 通过。
固定 Playwright image 中六张新分区快照已生成并完成 6/6 复比，仍等待维护者接受。
首页四张旧快照未改，比较出现预期的快捷入口差异；最终 Public 视觉/模式人工验收与 Visual Closure 尚未完成。

## Schema ownership

Goose migration `db/game/migrations/20261006020000_game_collection_foundation.sql`
独占以下 GFG 表；同一迁移提供所有 Table/Column 的中文 COMMENT：

| 表 | 事实与约束 |
| --- | --- |
| `gfg_game_collection` | 内部 identity 主键；唯一 1..64 字符 kebab-case `code`；zh/en 名称与简介；`status`、乐观锁 `version`、发布/归档/维护时间 |
| `gfg_game_collection_item` | `(collection_id, game_id)` 唯一人工成员关系；双向 FK 级联清理关系，反向查询索引 `(game_id, collection_id)` |
| `gfg_game_collection_home_slot` | 1..5 的唯一槽位和唯一 Collection 引用；删除分区时级联清理槽位 |

Item 不含 position、weight、sort_order、note、nsfw。Collection 不含 NSFW、排序或素材字段。
首页第六个“全部分区”入口不入库。本迁移不创建任何 Collection 或 Membership seed。
Down 明确拒绝删除策展数据；恢复需经过验证的备份或重建隔离数据库。

生命周期冻结为 draft → published → draft、draft/published → archived、archived → draft。
`published_at` 是最近一次成功发布时刻，draft 可保留；只有 archived 必须有 `archived_at`。
`version` 从 1 开始。Stage B 才实现版本校验、code 创建后不可变及这些写入转换；
Stage A 没有 Admin mutation、Membership editor、Home curation 或 Audit 功能。

## Hybrid Membership Foundation（#144-A）

`20261008010000_game_collection_hybrid_membership.sql` 新增两张 GFG 表：

| 表 | 事实与约束 |
| --- | --- |
| `gfg_game_collection_tag` | `(collection_id,tag_id)` 复合主键；分区删除级联，Tag 删除 RESTRICT；`(tag_id,collection_id)` 反向索引 |
| `gfg_game_collection_exclusion` | `(collection_id,game_id)` 复合主键；分区/游戏删除级联；`(game_id,collection_id)` 反向索引 |

两表及六个字段在同一迁移中写中文 COMMENT；`created_at` 记录人工配置时刻，不参与 chronology。
原 item 表继续只存人工固定，不回填自动成员，不新增 source/position/nsfw。Down 明确拒绝破坏配置数据。
迁移只在隔离 PostgreSQL 验证并生成 expected-final；历史迁移及 expected 不变。部署需先由维护者应用该 GFG 迁移，再更新 Game Backend/Admin；本阶段不操作共享开发库或生产库。

```text
ActiveTags = 绑定的 Tag 与所属 Category 均未归档的规则
Auto       = 任一 ActiveTags 命中的去重游戏（normal/primary/secondary 均参与，OR）
Manual     = gfg_game_collection_item
Excluded   = gfg_game_collection_exclusion
Effective  = (Auto ∪ Manual) − Excluded
Visible    = Effective 经当前 mode 的 Adult 过滤
```

绑定的是 Tag，不是 Category；类别归档会暂停其下标签规则。归档规则保留，恢复后重新命中。
人工固定不依赖标签继续有效；排除最高优先，可保留当前未命中的游戏。写入禁止人工固定与排除重叠。
未配置规则/排除的旧分区与原人工成员结果一致。规则有效性不影响 Adult 安全判断：归档 adult 标签关系仍具有成人语义。
Public Home/Index/Detail、q/phase/count sort、预览与 Admin 数量/首页资格统一使用 Effective；Public 不公开来源、规则、排除或隐藏成人数。

Collection 配置与派生成员分离：独立 Tag 变更不改变 Collection version/status/published_at/Audit/Home placement。
因此 version 只保护运营配置，不冻结 Effective。Tag Domain 与 Collection Domain 保持各自事务锁，不建立全局锁、定时同步或物化成员副本。

本阶段仅提供 DB 与 Backend；#144-B 的混合成员 Admin Workspace 尚未实施。现有 React 成员编辑仍通过兼容接口只管理 Manual，不显示或写回 Auto。

## Public API

三个 GET 均使用现有 `{code: 1, data: ...}` 成功响应 envelope，与 `/api/v2/game/home` 独立：

| 路径 | data |
| --- | --- |
| `/api/v2/game/collections/home` | 公共 metadata、`slots: [{slot, collection: summary}]`；slot 升序，最多五项 |
| `/api/v2/game/collections` | 公共 metadata、`page`、`page_size`、`total`、`has_more`、`items: summary[]` |
| `/api/v2/game/collections/:code` | 公共 metadata、`collection: info`、`items: timelineItem[]`；完整一维时间线，不分页 |

公共 metadata 为 `schema_version: 1`、UTC `generated_at`、UTC 日历日 `as_of_date`。
`lang=zh|en`，缺失或非法默认 zh；`mode=sfw|nsfw`，缺失或非法默认 sfw。
List 默认 `page=1,page_size=24`，page_size 上限 60；非整数、非正数回安全默认值。
无法安全计算 bigint offset 的 page 回到 1。越界页为 `items=[]` 并保留 total。
List 额外支持 `q`（默认空）、`phase=all|released|upcoming|mixed`、`sort=published_desc|count_desc|count_asc|name_asc|name_desc`。
非法 phase/sort 返回 400。q 不区分大小写、按字面子串匹配 Collection code/双语名称/双语简介，或当前 mode 可见成员的站内双语名称。
SFW 下 adult 成员不参与 keyword、phase 和数量排序，避免通过结果侧信道泄露隐藏成员。
released 指存在 First Available 或 release=available；upcoming 指无 First Available 且 release=upcoming（含 overdue/TBA）；mixed 两者都有，unknown 不计入任一侧。
默认 `published_desc` 按 published_at DESC,id DESC；count 排序按可见数量，再按 published_at DESC,id DESC；name 按当前 lang 实际投影名称，再按 id ASC。
Count/List 使用相同参数化筛选，total/has_more 是筛选后的分页事实。

仅 published 公开。非法 code、draft、archived 和不存在的 Detail 均为 HTTP 404，
同一 `Collection not found` 错误；SQL/连接故障为 503 `Collections unavailable`，不透传内部错误。
空 Index、空 Home 和零可见成员 Detail 都是 200；数组使用 `[]`。
HTTP `Cache-Control: no-store` 保持不变。Redis 与 CDN 是不同层；此合同不推断或修改 CDN/Nginx 策略。

`info` 投影：`code,name,info,visible_game_count,published_at`。
`summary` 在 info 上增加 `preview_games: [{game_id,name,header_url}]`。
`timelineItem`：`game_id,name,summary,header_url,phase,chronology`，game_id 是十进制字符串。
Detail schema v1 additive 字段：`primary_tag/secondary_tag: {code,name}|null`、`rating: {average,count}|null`、`online: {count,collected_at}|null`、`community_count: number`。
标签仅来自 game_tag.role=primary/secondary，按 Game Search 同一双语 fallback，不返回内部 Tag ID 或完整标签列表。
社群 `community_count` 只计算当前游戏 `groups` 中 key/value 经 trim 后均非空的可展示入口；null/非数组/无效元素为 0，不计 resources/links，URL 不公开。Game Detail safeGroups 使用相同 trim 规则。
评分只使用 GoFurry `gfg_game_comment` 的 AVG(score)/COUNT(*)，无评论为 null。在线人数表示最近一次成功观测（latest successful observed player count），不是严格实时；按 collected_at DESC、id DESC 取值，成功的 0 保留对象，只有失败记录则为 null。Tooltip/aria 按 locale 格式化采集时间并显式标注 UTC，保持 SSR/client 一致。
不返回 Collection 内部 ID、status、version、archived_at、原始成员数或隐藏成人数。
分区名称/简介逐字段优先请求语言，空值回退另一语言；游戏文案和图片复用现有 V2 优先级。

## 成人过滤与 Preview

唯一成人事实是 `gfg_game_tag → gfg_tag.code='adult'`。
即使标签已归档，只要关系仍存在就保持成人语义；数字 Tag ID 1014 和名称不参与判断。
sfw 在 Backend 先过滤成人成员，再解析时间线、计数和选 Preview；nsfw 返回完整有效成员。
成人专属分区在 sfw Index/Detail 仍存在，`visible_game_count=0,preview_games/items=[]`。
Home 对当前 mode 下零可见成员的槽位直接省略，也省略非 published 分区；只读跳过不删除 placement，标签恢复后可在缓存重建时重新展示。

Preview 从过滤后的完整有序时间线选取：0 项为 []，1 项取首项，2 项取首尾，
3 项及以上取 `[0, floor(n/2), n-1]`。最多三个不同游戏，不维护 Banner 或手工排序。

## Chronology

Resolver 接收显式 UTC as-of 日期；不解析 release raw text、不在纯函数内读取实时时钟。
First Available 是最高优先级，即使 current release 变回 upcoming 也保持 released。

| 顺序 / phase | 条件 | chronology |
| --- | --- | --- |
| 1 `released` | 存在 First Available；否则 current available 且有可信 window | `source=first_available` 或 `release` |
| 2 `released_unknown` | current available，无可信 window | null |
| 3 `upcoming_overdue` | current upcoming 且 window_end **严格早于** UTC as_of_date | release window |
| 4 `upcoming` | current upcoming，window_end 当天或未来 | release window |
| 5 `upcoming_tba` | current upcoming，无可信 window | null |
| 6 `unknown` | current unknown/missing 且无 First Available | null |

有 window 的 phase 按 `window_start ASC,window_end ASC,game_id ASC`；其余按数值 game_id ASC。
`window_end == as_of_date` 不算 overdue。
chronology 只含 `source,precision,window_start,window_end,inferred`；保留 day/month/quarter/year 精度。
window 是排序区间，不是合成的精确发布日期；没有 synthetic display date。
First Available 的 inferred 原样保留，内部 legacy_manual/steam_backfill/observed_transition 不公开。

## Batch 与 Cache

固定 sqlc 查询负责公开 Collection、count、page、slots 和批量 Effective Membership：按 Collection IDs 限定 `(manual UNION active-tag matches) anti-join exclusions`，不逐游戏/逐标签查询。默认 Browse 仍先分页，只投影当前页成员。
Home 使用单条 sqlc 查询读取 published 槽位、Collection 双语元数据和按 adult code 过滤的可见数量，不加载任何 Game Aggregate 或时间线。
Home schema v1 保留 CollectionSummary 结构，preview_games 固定为空数组；Index/Detail 仍保留完整 preview/timeline。
Index/Detail 使用 `LoadCollectionProjectionGames`：一次批量 Membership，去重 game IDs，随后固定批量读取站内双语文案、localized name/summary、detail name/header、仅 header 类型的 media/assets、First Available 和 Release State。
核心成员与投影合计 6 次 SQL，与成员数量无关；默认 Index（Count/List + 投影）8 次，Home 1 次。
Detail 核心链路审计为 Get + 六次投影，共 7 次；本轮新增过滤后 `LoadCollectionTimelineDecorations` 单次 sqlc batch，总计固定 8 次（零可见成员跳过 decoration）。
装饰仅查询当前 mode 可见 game IDs。Index 不调用装饰读取，保持原查询和 DTO；Detail 只新增主次标签、评论评分、最新成功在线记录与社群数。
不进入完整 Game Aggregate，不读取价格、峰值、配置需求、新闻或完整素材，不内部 HTTP 调用。PG integration 的 3/30 成员验收使用真实 SQL tracing；耗时仅诊断，无 CI 毫秒门槛。
游戏文案继续复用现有 locale fallback；完整 read model 和轻量投影共用 canonical header helper，保持 media/asset 优先级及归档 adult 标签语义。

Redis 是 Origin Read Model acceleration，TTL **1 hour**，不是内容发布时限承诺。
仅默认 Browse（空 q、phase=all、sort=published_desc）、Home、Detail 使用以下命名空间；任何非默认 Discovery query 直接走轻量 DB read，不读写长期结果缓存，避免任意关键词高基数 key。

```text
game:v2:collections:v3:home:{asOfDate}:{lang}:{mode}
game:v2:collections:v3:list:{asOfDate}:{lang}:{mode}:{page}:{pageSize}
game:v2:collections:v3:detail:{asOfDate}:{lang}:{mode}:{code}
```

内部 revision v3 避免读取 v2 的人工-only 旧结果；Public schema_version 仍为 1。旧键自然过期，不扫描、不主动删除。Tag 变化在下一次 Origin cache miss/rebuild 生效，不承诺最终公开时限。
UTC 日期切换立即换 key。有效缓存命中直接返回；miss/error/malformed 回 DB，成功后 best-effort 写缓存。
数据库错误和 404 不缓存。每个 Service 实例用 singleflight 合并同 key 的并发 miss；不同 key 独立。
缓存序列化结果按调用方解码，避免共享可变 slices。单个调用方取消不终止其他等待者，构建有 8 秒总预算。
Redis 读写使用 200ms context 和共享连接池的 read/write timeout clone，不改全局 Redis 配置。
Stage A 不做跨服务精确失效或 stale fallback；公开内容在缓存刷新后更新，不承诺固定的最终可见时限。

## 验证与交付边界

- Service 单测覆盖完整 chronology、UTC 边界、稳定排序、preview、mode/locale、cache TTL/分键/故障及 singleflight。
- Controller 单测覆盖 query normalization、DTO envelope、状态码和错误脱敏。
- `TestPostgresReadModelSemantics/collections` 接入既有 postgres-integration gate，验证 populated read model、实际 HTTP、成人归档语义、固定批次数及 FK/check/unique/cascade。
- `TestPostgresFreshAndBaselineAdoption` 验证 PG18 fresh/adoption/drift/readability；`expected-final/gfg.json` 从实际库生成。
- `task check:db-readability`、`task check:sqlc`、`task check:policy` 保持静态治理。

Stage A 的交付包含 GFG migration 与 Game Backend；Collector 仅新增 generated schema model，不要求单独发布。
Stage A 不包含 Admin 与 Public UX。Stage B 运营合同见下文；Stage C 的 Nav Web 页面、模式与时间线合同及待完成人工验收状态见本文开头。

## Stage B：Admin 运营闭环

Admin 独立拥有 `/api/v1/game/collections`，与 `/api/v1/collection` 的采集控制面无关。
所有 GET 使用现有 `content.read`，所有写操作使用 `content.write`，不新增能力。

| 方法与路径（上述前缀内） | 行为 |
| --- | --- |
| GET / | 按 keyword（code/name/name_en）、status、home_eligible 筛选；page_num 默认 1、page_size 默认 50、上限 200；updated_at DESC、id DESC |
| POST / | 创建 version=1 的草稿；双语名称必填，简介及 0/1 个成员允许暂缺 |
| GET /:id | 完整 Admin DTO：基本内容、状态、版本、生命周期时间、总成员/SFW 数、home_slot |
| PUT /:id | 携带 version 更新双语名称/简介；Code 不可更改 |
| GET /:id/members | 同一只读快照内返回 collection_id、version、完整人工固定成员及 Adult 标记 |
| PUT /:id/members | 携带 version 与完整人工 game_ids；去重、数值排序、存在性及与 Excluded 互斥校验后替换，不写自动派生成员 |
| GET /:id/composition | 同一只读快照返回规则/人工/排除/有效成员、来源、计数、version 与 home_slot |
| PUT /:id/composition | 携带 version 与完整 tag_ids/manual_game_ids/excluded_game_ids 原子替换配置 |
| POST /:id/publish、unpublish、archive、restore | 携带 version 执行显式生命周期转换 |
| GET /home-curation | 固定五位及 canonical placement revision；空位 collection=null |
| PUT /home-curation | 携带 revision，完整提交第 1–5 位各一次；非空分区不可重复 |

静态 `/home-curation` 先于 `/:id` 注册。Code 使用 1–64 位小写 kebab-case；重复返回 409。
名称最多 160、简介最多 500 个 Unicode 字符。未知写入字段被拒绝，不能变相写入 Code、排序或 NSFW。

### 版本、事务与生命周期

所有写操作先取得 GFG transaction advisory lock `gfg.game-collection-domain`，再锁定 Collection 行、比较 version、执行依赖修改。
内容、成员及生命周期成功后 version+1；真正相同的内容/成员集合不改变 version，也不写 Audit。
旧 version 返回 409：`此游戏分区已被其他操作修改，请重新加载后重试。`；客户端不得自动重试 mutation。

- Publish：draft→published；双语名称/简介非空、当次 Effective 至少两个；published_at=now。
- Unpublish：published→draft；保留 published_at，同事务移除首页入口。
- Archive：draft/published→archived；archived_at=now，保留 published_at/成员并移除入口。
- Restore：archived→draft；清空 archived_at，保留 published_at，不恢复发布或首页入口。
- Archived 拒绝内容、成员及 Composition 修改；Published 显式 Composition/Members 修改后 Effective 须至少两个，纯文案编辑仅要求双语完整，不因被动标签变化造成成员不足而阻塞。

Adult 唯一按 Tag code=`adult`，包括已归档 Tag 的现有关系；ID 1014 不具有特殊意义。
成人限定分区可发布，但首页资格必须是 published 且 sfw_member_count>0。
显式 Composition/Members 保存后 Effective SFW=0 时，同一 GFG 事务自动移除 Home Slot；满足总 Effective 至少两个的分区仍可 Published。独立 Tag 变化不删除 Slot、不自动改变生命周期。

### 首页与 Audit

revision 只对第 1–5 位 Collection ID/null 的 canonical placement 计算 SHA-256；名称、version、count 不影响 revision。
完整替换在领域锁下校验旧 revision：新增/移动 Slot 必须符合当前资格，位置未变的暂时失效 Slot 可保留，允许编辑其他位置。原子删除/插入；显式首页编排不增加 Collection.version。
第六个“全部分区”是固定产品入口，只在 Admin 说明，不进入 API 或数据库。

Collection Audit resource 为 `gfg_game_collection`，action 为 create/update/members_update/composition_update/publish/unpublish/archive/restore。
快照仅含有界业务字段，成员保存附带 canonical game ID 集合；自动撤下入口体现在同条记录的 home_slot before→null。
首页显式编排使用 `gfg_game_collection_home_slot` / `home_curation_update`，记录五个 placement。
沿用现有 GFG business transaction + 独立 GFA Audit；审计失败阻止 GFG 提交，但这不是跨库 ACID。

### Composition 写入合同

GET 的 `rule_tags` 含 Tag ID/code/双语名称/active；`manual_members`、`excluded_members` 与 `effective_members` 是完整授权运营数据，Effective 的 source 为 automatic/manual/both。空数组为 `[]`。
`counts.auto_matched` 是排除前 Auto 去重数，manual_pinned/excluded 是配置记录数，effective/sfw_visible 是实际成员数，不能简单相加。响应包含同一版 version 和 home_slot。

PUT 三个数组必须显式提供（允许空，不允许 null/缺失），正整数去重并数值排序，所有 IDs 必须存在，Manual 与 Excluded 交集返回 400。
新绑标签必须 Tag/Category 均有效；之前已绑定但后来失效的规则可原样保留或移除，移除后不能重新绑定失效标签。
沿用领域事务锁→行锁→version 比较→set-based replace→状态校验→version+1→Audit；全相同配置 no-op，无版本或 Audit 写入。
`composition_update` Audit 仅存前后版本、配置 IDs、计数及 home_slot：每组 IDs 最多 256 个 canonical 前缀，超出时 ids_truncated=true，并保存三组完整配置的 SHA-256；不保存海量自动派生成员。不限制或截断请求和 GET 成员数组。

### React Workspace 与刷新

路由：`/game/collections`、`/new`、`/:id`、`/home-curation`（后三者均在相同前缀下）。
列表搜索使用 IME-safe helper 并将 keyword/status/page_num 保存在 URL。Desktop 搜索、固定宽度状态筛选与列菜单处于同一工具行；窄屏可换行。
单页 Workspace 只保留基本内容与收录游戏；成员只做添加/移除，不提供顺序编辑或 Timeline Preview。
成员完整加载并保留本地 draft；已收录游戏搜索按 name/name_en/appid 本地 IME-safe 匹配，每页 20 条，搜索回第 1 页，增删夹到合法页。远程“搜索并添加游戏”独立，跨页编辑最终仍 PUT 完整 canonical game_ids。
Game options 与 Home eligibility 搜索复用 RemoteSelect，不消费 IME 确认键。

内容与成员草稿共享 baseVersion；本页保存成功可推进版本并保留另一份草稿，背景刷新不得覆盖脏草稿或提升其版本。
409 保留草稿并要求显式重新加载。Header 的“重新加载”在干净状态直接取最新 workspace；dirty 时显示“放弃修改并重新加载”，确认后同时丢弃内容/成员草稿并采用新 version，取消则原样保留；成功 reload 清除 error/conflict。草稿通过 useUnsavedChanges 保护导航，生命周期操作在 dirty 时禁用且全部要求确认。
只读用户可查看完整内容；归档后仅 Restore 可写。Audit 入口仅对 audit.read 显示，使用 resource 过滤而不假装支持 target_id。
Home 草稿同样绑定原 revision，背景刷新不会覆盖它。

Stage B 没有新迁移，也不变更 Stage A cache：不 purge Redis、不新增内部失效接口或 Pub/Sub。
所有 Collection 成功 Toast 只显示“已保存”。自动撤下首页以返回的 home_slot 状态为准，不重复解释缓存或刷新时限。Workspace 不设公开刷新说明；首页编排只保留选择资格提示。

验证包含 Controller/route 单测、`TestAdminGameCollectionThreeDatabase` 的隔离 PG18 并发/约束/Audit 集成测试，
以及 React 列表 IME、共享版本、冲突保留、生命周期、只读、五位完整编排和未保存保护测试。
集成测试接入既有 postgres-integration gate，不对共享开发或生产库执行 Goose。

#144-A 追加覆盖 OR/角色/去重、固定与排除、归档规则恢复、SFW 搜索/phase/count 隐私、被动 Home 保留与显式移除、Composition 409/no-op/Audit 回滚和旧 Members 兼容。3/30/100/300 成员的 Detail 与 Index 均维持 8 次 SQL，Home 1 次；诊断执行时间不设 CI 毫秒门槛。
