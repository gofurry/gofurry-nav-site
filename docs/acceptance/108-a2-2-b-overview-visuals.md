# #108-A2.2-B 验证记录

日期：2026-10-10（Asia/Shanghai）。基线：`dev` / `16db2dc`。
范围：Insights 自动图片资格、Nav Overview Logo 投影、Overview 局部构图。
本记录不代表维护者审美验收、逐资产 SFW 审核或生产验收。

## 修改边界

Game 仅改 Insights 查询、生成代码、私有读模型/映射/筛选及专属测试。
Nav 仅改 Insights Overview 的只读投影与专属测试；新增的
`apps/nav/navPage/dao/insights_visuals_integration_test.go` 复用该目录现有隔离库测试生命周期，
没有修改 Nav Page 业务实现或原有集成测试。
Nav Web 仅改 Overview 组件/纯函数/样式、可选类型、局部翻译和测试。

对基线核对：Showcase、Collections、Admin、Collector、Home/Search/Detail、其他 Insights
页面、共享图片组件/路由、全局样式/Tokens、style-debt、Visual Golden、迁移、配置均无修改。
`index.vue`、指标选择算法、`InsightsOverviewActivityItem` 和公共 `publicChanges()` 均未改。

## 数据资格和只读抽检：PASS

- Game：不依赖 `showcase_eligible`；当前有效非 adult 分类是必要条件，任何 adult 关系
  （包括归档标签）均排除。Game/AppID/asset 三方身份、名称、Steam HTTPS URL 有限语法
  在 `LIMIT 3` 前校验；Go 再校验。每日按 UTC 日期与 Game ID 稳定排序，无商业权重。
- Nav：维护者确认真实 GFN 契约为 **`nsfw='0'`**，实施文档的 `'false'` 不适用。
  只从实际返回的事件实体批量读取最多八个 ID，保序去重后最多五项；仅接受归属本站点的
  Managed Asset icon key。空事件不查询，不改原事件 DTO。
- 两种装饰查询均有 500ms 子上下文；失败/超时返回各自 `[]`，原指标/事件错误契约不变。
- 开发库使用已授权的本地应用账号，通过 debug YAML 连接；连接默认只读，并使用只读事务
  和回滚。没有读取/修改 `.env`，没有写共享开发库或连接生产库。
- 抽检 GFG：217 款游戏；58 款 adult，4 款无当前有效分类（也没有标签），155 款通过
  完整资格。无无效 Game/AppID。UTC `2026-10-09` 的三个候选：

| Game ID | AppID | 名称 | Asset ID |
|---|---|---|---|
| 208 | 3363560 | 晶灵小队 / Shard Squad | 765821 |
| 237 | 489830 | The Elder Scrolls V: Skyrim Special Edition | 764008 |
| 144 | 3100580 | 白影酒馆 / Tavern Revisited | 766783 |

三者 `asset.game_id/appid` 匹配、`exists=true`、header 位于受信任 Steam 域名且路径 AppID
匹配。此处没有下载图片、检查像素或承诺当前 CDN 可用性。
实际候选查询 `EXPLAIN (ANALYZE, BUFFERS)`：规划 3.623ms、执行 81.482ms；
扫描 155 个合格游戏的资产。这是当前小规模开发库结果，不外推生产规模。

GFN 分类抽检：`'0'` 共 211 条（未删除 195、其中有 icon 194），`'1'` 共 32 条
（未删除 26、其中有 icon 25）。实际 Overview `recent_changes=[]`，因此 `site_visuals=[]`；
没有捏造近期站点。实际 Overview 只读运行约 180ms。空 ID 批次诊断 EXPLAIN 执行 0.030ms，
不能当作真实非空 Logo 批次的性能结果。非空资格、过滤及超时由隔离库用例验证。

## UI 和语义：PASS

Hero 保持桌面 8:4；有合格首图时游戏内部数据/图片 56:44、证据在底带，手机图片在数据之前。
无合格首图/主指标或图片耗尽时回到 A2.1 纯数据布局。缩略图只用第二/第三个候选，
最多两项且不会补位复用 Hero；失败独立隐藏。Logo 最多五项，失败为中性名称 initials。
Game/Site 实体链接与指标链接独立且本地化，无嵌套链接。图片复用未修改的共享渲染器。
两个并行 SSR Overview 请求、独立失败、Hydration、SEO、零/未知、百分点、样本和事实日期保护保留。
事件仍为中性领域图标，不渲染事件 payload 的图片。

## 检查命令与结果

所有编译/测试在开发机器执行。Go 命令使用 `GOTOOLCHAIN=go1.26.7`；Node 24、pnpm 12.6.0。
隔离集成测试连接本机临时 PostgreSQL 18 容器，自建/销毁测试数据库，不使用共享开发库。

| 检查 | 命令（目录） | 结果 |
|---|---|---|
| 工具链 | `task doctor`（根） | PASS |
| SQLC | `task generate:sqlc`、`task check:sqlc`（根） | PASS；生成代码一致，未手改 |
| 格式/政策 | `task check:fmt`、`task check:policy`（根） | PASS |
| Go 全模块 | `go test ./...`、`go build ./...`、`go vet ./...`（Game Backend、Nav Backend 分别执行） | PASS |
| Game 隔离集成 | `go test ./apps/game/v2/dao -run TestPostgresReadModelSemantics -count=1 -v`（Game Backend） | PASS；包括原 Showcase/Collections 与新增资格用例 |
| Nav 隔离集成 | `go test ./apps/nav/navPage/dao -run TestPostgresNavBackendPersistenceSemantics -count=1 -v`（Nav Backend） | PASS；包括原持久化及新增 Insights Logo 用例 |
| 静态与类型 | `pnpm run lint`、`pnpm run stylelint`、`pnpm run style:policy`、`pnpm run typecheck`（Nav Web） | PASS；style-debt 未扩大 |
| 单测 | `pnpm run test:unit`、`pnpm run test:nuxt`（Nav Web） | PASS；397 + 76 项 |
| 语义/SEO | `pnpm run insights:semantics`、`pnpm run seo:recovery:test`（Nav Web） | PASS |
| 生产构建 | `pnpm run build`（Nav Web） | PASS |
| Overview 浏览器 | `pnpm run test:browser tests/browser/regression/insights-overview.spec.ts --workers=1` | PASS；119 项 |
| 跨模块浏览器 | 下方命令 | PASS；35 个文件、618 项，15.6 分钟 |
| 已接受 Golden 对比 | pinned Linux Visual 工作流 | NOT RUN；本机 Windows 未运行此 Linux 工作流，本轮导出审阅图，不更新 Golden |
| 维护者人工审美验收 | 下方清单 | NOT RUN；等待维护者 |

跨模块命令（Nav Web，使用同一 production build，单 worker；Overview 已单独全跑）：

```powershell
pnpm run test:browser insights games-home game-collections games-search game-detail site-detail site-overview site-observation site-security site-recommendations nav-home nav-revealed nav-shell steam-asset-routing managed-asset-routing --grep-invert 'insights-overview.spec.ts' --workers=1
```

该选择覆盖其他 Insights、Games Home/Showcase、Collections、Nav Home、Games Search、
Site/Game Detail、共享 Steam/Managed 资源路由和相关 smoke 用例。测试及共享 fixture 未修改。
推送后的 CI 必须按实际提交 SHA 单独报告；本文件不预言 CI 结果。

## 截图与维护者检查清单

本地目录：`apps/cn/nav-web/.cache/insights-a2-2-b-review/`。
矩阵为 zh/en × light/dark × 1440/1024/768/390，共 16 组，每组：

- `visual-full-{locale}-{theme}-{width}.png`
- `visual-hero-{locale}-{theme}-{width}.png`
- `visual-ecosystems-{locale}-{theme}-{width}.png`

共 48 张 Playwright 实际截图；图片内容使用稳定 fixture 替身，不作为生产资格证据。
同目录另有 19 张原纯数据/失败降级截图。未写入已接受 Visual Golden。

维护者待核：

- 桌面数据是否仍为主视觉，游戏美术及名称是否保持独立事实表达。
- 1024/768 的纵向节奏、390 图片高度及双生态折行是否符合预期；五个 Logo 在窄屏自然换行。
- 中英文长名称、深浅主题证据文字、键盘焦点及手机底部公共导航的实际观感。
- 无图、单图失败、整域失败时是否接受现有纯数据布局；无多余占位或虚构 Logo。

## 剩余风险与停止边界

分类元数据及 Steam/Managed Asset 身份校验不分析图片像素。标签遗漏、错误分类、第三方同 URL
更换图片及缓存传播均有剩余内容风险，不能宣称逐图审核或绝对 SFW。自动规则不引入人工挑图、
日常审批、审核表或新开关。开发库当前没有近期 Nav 事件，真实非空 Logo 内容审阅没有执行。
本批不实施 A2.3/#139，不合并 main、不部署、不关闭 #108。

## 修改文件清单

~~~text
apps/cn/game-backend/apps/game/v2/dao/insights_visuals_integration_test.go
apps/cn/game-backend/apps/game/v2/dao/insights.go
apps/cn/game-backend/apps/game/v2/models/insights.go
apps/cn/game-backend/apps/game/v2/service/insights_visuals_test.go
apps/cn/game-backend/apps/game/v2/service/insights_visuals.go
apps/cn/game-backend/internal/db/game/queries/insights.sql
apps/cn/game-backend/internal/db/game/sqlc/insights.sql.go
apps/cn/nav-backend/apps/nav/insights/controller/insights_test.go
apps/cn/nav-backend/apps/nav/insights/dao/insights.go
apps/cn/nav-backend/apps/nav/insights/models/insights.go
apps/cn/nav-backend/apps/nav/insights/service/insights_test.go
apps/cn/nav-backend/apps/nav/insights/service/insights.go
apps/cn/nav-backend/apps/nav/insights/service/overview_visuals_test.go
apps/cn/nav-backend/apps/nav/insights/service/overview_visuals.go
apps/cn/nav-backend/apps/nav/navPage/dao/insights_visuals_integration_test.go
apps/cn/nav-backend/internal/db/nav/queries/insights.sql
apps/cn/nav-backend/internal/db/nav/sqlc/insights.sql.go
apps/cn/nav-web/app/assets/styles/pages/insights/overview.less
apps/cn/nav-web/app/components/insights/overview/InsightsOverviewEcosystems.vue
apps/cn/nav-web/app/components/insights/overview/InsightsOverviewGameVisual.vue
apps/cn/nav-web/app/components/insights/overview/InsightsOverviewHero.vue
apps/cn/nav-web/app/components/insights/overview/InsightsOverviewSiteVisual.vue
apps/cn/nav-web/app/components/insights/overview/visuals.ts
apps/cn/nav-web/app/types/insights.ts
apps/cn/nav-web/i18n/locales/en.json
apps/cn/nav-web/i18n/locales/zh.json
apps/cn/nav-web/scripts/insights-semantics.test.mjs
apps/cn/nav-web/tests/browser/fixtures/insights-overview.ts
apps/cn/nav-web/tests/browser/regression/insights-overview.spec.ts
apps/cn/nav-web/tests/unit/insight-overview-visuals.test.ts
docs/acceptance/108-a2-2-b-overview-visuals.md
docs/public-insights.md
~~~
