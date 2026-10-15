# #108 Overview Production Consistency Fix

执行日期：2026-10-16（Asia/Shanghai）。基线：本地 `dev` / `93165157549bf0bb3c8c3acc1fadf165f6887c06`；开始时工作区干净。

本批仅修复 Hero 事实关联、两张 Hero 材质及首页冗余眉题。最终视觉审美仍待维护者确认。

## 实现与范围：PASS

- Game Overview 新增 `hero_visual: null | { game_id, name, name_en, visual: { kind: "game_header", asset }, metric_key, as_of }`；原 `featured_visuals` 保持非 null 的 0–3 项数组。
- 主指标沿用 `free → windows → mac → linux` 的首个合法比率（含 0）。内部合同分别为 `free_game_share/1`、`windows_support/1`、`mac_support/1`、`linux_support/1`；日期来自所选指标自身。该日期无效时不换指标补图。
- 复用一次 `ListGameInsightVisualCandidates`，在 `LIMIT 3` 前匹配四列事实身份及 `state='positive'`，正向匹配优先，再按 UTC 日期 + Game ID 的 MD5 和 ID 排序。名称、分类、成人关系、资产身份及有限 Steam URL 语法仍全部有效。
- 服务层复核候选；保留 500ms 子上下文、错误/超时返回 `[]` 与 `null`、原统计/变化失败契约。没有第二次视觉查询、N+1、HTTP 源或 Redis 键。
- 首页校验证明结构、URL、指标键与事实日期；旧后端/无证明/错误证明使用数据 Hero。生态从普通候选中排除合格 Hero ID 后取两项；没有 Hero 时取前两项，图片耗尽时不提升其它作品充当 Hero。
- Hero 图片改为“该指标的观测作品” / “Observed example”；普通缩略图仍为“收录作品”。两张 Hero 使用 `--gf-surface`、外框 0、阴影 none；保留圆角、内部线条和响应式。只删除首页 `GOFURRY / INSIGHTS` 并将 H1 上边距归零。
- 文件范围：Game Insights models/DAO/service、专属查询及生成文件/测试；Overview Hero、Ecosystems 的预览选择、GameVisual 的标签调用、纯函数、类型、两项 locale、index、Header/Hero Less；对应单测、浏览器 fixture/spec、两条旧语义守卫；`docs/public-insights.md` 与本记录。
- `git diff --name-only` 范围检查通过。未修改 Showcase、Collections、Nav/Games Home、其它 Insights、共享图片组件/CDN、Nav Backend、Admin、Collector、迁移、CI、依赖、Tokens、样式债务或 Visual Golden。两次并行 SSR、SEO、链接、Activity/Directory 逻辑不变。

## 验证

| 检查 | 结果与证据 |
| --- | --- |
| `task generate:sqlc` / `task check:sqlc` | **PASS**；根目录生成，工作树契约一致，无手改生成文件。 |
| `go -C apps/cn/game-backend test ./apps/game/v2/service ./apps/game/v2/dao -count=1` | **PASS**；四指标映射、真实零、非法比率/日期、资格和普通候选、UTC、一次查询及失败隔离。 |
| 隔离 PostgreSQL 18 | **PASS**；既有 `TestPostgresReadModelSemantics/insight-featured-visuals`，独立随机测试库，执行迁移、测试后销毁。使用本机临时容器，结束后已停止。覆盖全部旧 URL/分类/成人/归属撤回用例，以及四指标、非正向状态、错键/版本/日期、正向候选位于旧前三名以外和稳定轮换。 |
| EXPLAIN (ANALYZE, BUFFERS) | **PASS**；隔离 fixture 78.769ms，开发只读查询 72.947ms。四个等值谓词与既有事实复合主键匹配；隔离库计划选用 `game_id,fact_date` 索引，开发小表计划选择顺序扫描。没有新增索引，这些耗时不代表生产规模表现。 |
| Nav Web lint / Stylelint / style policy | **PASS**；未扩大任何债务基线。 |
| Nav Web typecheck / insights:semantics | **PASS**。旧“首张随机封面”与“固定 slice(1)”守卫改为专属事实证明/去重入口，未移除原数据语义保护。 |
| 两份定向 unit | **PASS，36 tests**。初次单测暴露纯函数运行时导入不支持 Nuxt `@` 别名，改为相对导入后通过。 |
| Nav Web build | **PASS，执行一次**；保留已有 VueUse 注释/exports deprecation 警告，未改依赖。 |
| 最终完整 `insights-overview.spec.ts --workers=1` | **PASS，142/142，3.4 分钟，一次完整运行，0 retries**。新增 20 项证明/指标用例，保留原 122 项；覆盖中英文、两主题、四视口、两源 SSR/Hydration、无 JS、SEO、日期/百分点/零值、局部失败、安全降级、链接及事件/目录。 |
| Hero 实际计算样式 | **PASS**；有图和数据态均检查两背景相同且等于解析后的 `--gf-surface`、两卡四侧边框均 `0px`、外阴影均 `none`；原几何/数据断言保留。 |
| Fixture 截图矩阵 | **PASS**；16 组合 × 整页/Hero/生态/Activity/Directory = 80 张必需图齐全，另含数据态/失败等共 138 张。 |
| 真实开发页面与图片加载截图 | **NOT RUN**；两份现有 debug 配置的本机 Game/Nav 服务均未监听。没有启动 Collector、调度或生产服务，也没有以 Fixture 代替真实页面验收。 |
| 远端 CI / 生产验证 | **NOT RUN**；本记录仅陈述本机检查，不宣称远端 CI 或生产通过。 |

日志位于忽略的 `.cache/108-consistency/`：`sqlc.log`、`go-integration.log`、`lint.log`、`stylelint.log`、`style-policy.log`、`typecheck.log`、`insights-semantics.log`、`unit.log`、`build.log`、`browser.log`。

## 真实开发数据只读核查：PASS（当前无 Hero 资格）

按已有授权使用本地开发应用账号；先核实 debug/GFG/Tailnet 开发目标，再设置连接默认只读和 `REPEATABLE READ READ ONLY` 事务，并确认 `transaction_read_only=on`。没有写共享库、打印凭据或连接生产。

实际执行本次 DAO/service 后：

| 指标 | 自身 as_of | 比率 | known / eligible | 该日可展示 positive 候选 |
| --- | --- | --- | --- | --- |
| free | 2026-10-12 | null | 0 / 0 | 0 |
| windows | 2026-10-12 | null | 0 / 4 | 0 |
| mac | 2026-10-12 | null | 0 / 4 | 0 |
| linux | 2026-10-12 | null | 0 / 4 | 0 |

四次有界候选审计均返回 3 项普通候选且没有 `hero_match`；由于匹配在 LIMIT 前优先排序，这说明当前完整安全候选池没有对应正向项，而非只抽查前三项。实际 Overview 为 `hero_visual: null`，普通候选 Game IDs 为 `137,139,240`。未知比率未伪装成 0，不存在可宣称“免费正向样本”的真实 Hero 图片。脱敏证据：`real-data.json`、`real-preflight.json`（忽略目录内）。

## 截图与人工复核：PENDING MAINTAINER

全部图片位于 `.cache/108-consistency/fixture/`，为生产构建 + Fixture 数据，不是生产或真实开发美术。

建议先看四张：

- `visual-full-zh-light-1440.png`：桌面整体、眉题消失、两卡材质及既有四区。
- `visual-hero-zh-dark-1440.png`：深色两卡背景、内部分隔、观测作品标签及事实日期。
- `visual-full-zh-light-390.png`：手机完整纵向顺序与普通缩略图。
- `visual-hero-en-dark-390.png`：英文手机折行、百分点说明、日期和链接。

已检查以上画面及自动几何断言；仍请维护者确认留白、主题材质和“观测作品”文案观感。真实服务可用后补验 1440/390px 中文页面；当前开发库应显示数据 Hero，普通预览可保留。

剩余风险：自动分类和 Steam 身份不保证像素 SFW；正向事实只证明其 `as_of` 当日状态，不保证一直免费/支持平台或具有统计代表性。生产数据规模下性能、真实图片可达性与生产截图未验证。分开发布时应先更新严格校验的前端，再更新后端；本次不部署、不合并、不关闭 Issue，不进入其它阶段。
