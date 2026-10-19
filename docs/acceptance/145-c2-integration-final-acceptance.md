# #145-C2 — Integration & Final Acceptance

2026-10-20。起点 `2546d73a1a8ac57ac3d70a4857c34aa3710514c7`，本地/远端 dev 一致且工作区干净。仅在现有 dev 工作。
维护者在本次任务明确签核 A/A1/A2/A3、B、C1、C1.1 的外观与交互；此前记录的人工 PENDING 是当时事实，未改写。

## 既有证据与范围

- [A3](145-a3-diagonal-cut-shell.md)、[B](145-b-hero-scroll-content-readiness.md)、[C1](145-c1-twin-island-utility.md)、[C1.1](145-c1-1-navigation-clearance-interaction-polish.md) 保留原始执行证据。
- C2 不重新设计双岛、二层、Search、平台控件或快捷管理；只修策略、迁移 Visual 契约/审查基线、补当前文档及实际遮挡遗漏。
- 局部产品修正：Collections Index 桌面搜索/高级筛选沿用 C1.1 的 `--gf-nav-content-top`，由 104px 改为 132px。900/1440 的原安全间隙 -12px 已被定向测试证实；手机和业务逻辑不变。
- 集成中复现 A2 的跨断点失焦：桌面 CSS 可先让隐藏控件触发 focusout，再发 matchMedia change，旧回调已无法知道焦点归属。只在现有 picker focusout 分支归还可见类别焦点，不改平台排序、选择、搜索或外观；确定性 Nuxt 反例先 FAIL 后 PASS。
- 未修改依赖、CI、后端、数据库、性能基线或历史验收记录。未部署、运行 update.sh、合并 main 或关闭 #145。

## Repository Policy

[原失败 Run 39291330618](https://github.com/gofurry/gofurry-nav-site/actions/runs/39291330618) 的 Fast/nav-web 成功，Repository Policy 的 Tooling contract tests 失败。
实际日志同时报旧 `deps:nav-web` **和** `build:nav-web`，比冻结文档只列安装任务更完整。Taskfile 的真实任务为 `deps:web` / `build:web`。
策略读取正确名称；保留精确 `pnpm install --frozen-lockfile` 与构建 package cwd，安装也严格核对所属 package cwd。新增旧安装/构建任务、Web 非 frozen、Admin/Web 错误 cwd 反例，保留所有原反例。测试只规范化 CRLF 以使跨行 mutation 可在 Windows/Linux 运行。

在 tools 使用 Go 1.26.9：

```sh
go test ./check-production-policy -run TestTaskfileReleaseContract -count=1
go run ./check-production-policy
```

两项 PASS。初次补充反例的 CRLF 匹配失败已修正；未跳过检查或更改 Taskfile。

## 本地功能与 Visual

钉死 Linux amd64 Playwright v1.60.0 Noble：`sha256:9bd26ad900bb5e0f4dee75839e957a89ae89c2b7ab1e76050e559790e946b948`，Node 24、pnpm 12.6.0、frozen install、`GOFURRY_VISUAL_ENV=pinned`。无重试、容差或 mask 变更。

| 门禁 | 实际结果 |
| --- | --- |
| 定向 ESLint / Stylelint | PASS |
| style:policy / typecheck | PASS |
| Nuxt nav-utilities + nav-custom-sites / search-platform-picker | PASS，6/6 + 2/2（后者包含新增失焦反例） |
| 最终构建 / Functional | PASS；首轮 87 PASS / 2 FAIL，修复后相关 11/11 PASS（不重复无变化用例） |
| Visual 审查后定向更新 / 独立只读 compare | PASS，94/94；首次失败历史见下文 |
| 最终源码 Fast / Repository Policy | NOT RUN，等待推回 |
| 最终源码 Full 四组 / Docker Image / Gate | NOT RUN |
| 最终源码完整 Visual | NOT RUN |
| 生产部署与冒烟 | NOT RUN，维护者执行 |

Visual 迁移：Nav Shell 改检可见 clipped shell 的单层雾面/无框/无高光，并按实际二层底边裁切；手机 Bottom Tabs 通过实际 Hero/content 位置触发，替换旧 73px。
旧 `nav-tool-dock-search-{light,dark}-desktop.png` 一对一迁移为 `nav-right-island-search-{light,dark}-desktop.png`，保留目录、真实结果与焦点，包含完整右岛和二层，排除偶然 hover。Core 的测试定位使用当前安全区，旧 helper 名不是产品 Reveal。Home Overlay 包含完整 center Search；Header / Game Detail Overview 使用完整 viewport，保留此前被旧 header 高度裁掉的真实区域。
Nuxt 的既有 i18n locale-load 控制台警告保留于日志；6 个断言通过，不新增豁免。真实中英文 Browser 的 hydration/网络账本另行严格验证。
首轮完整 compare：41 PASS / 96 FAIL（92 张像素差异 + 4 项过时裁切/滚动定位断言）。迁移后 21 项定向 compare：3 PASS / 18 像素差异，无剩余语义断言失败。
首轮与迁移后的全部 expected/actual/diff 已按 owner 审查；底栏两张、Preferences、Foundation、Review/Lightbox 等原通过基线不刷新。更新白名单为下列 94 张，独立只读 compare 94/94 PASS（2.9m）。哈希核对 136 张中 94 张变更、42 张完全不变，白名单外零变更。4 张生成图与审查 actual 存在边缘采样差异，逐张复核且独立比较通过原阈值；未增加容差或 mask。更新调用按 owner + 精确测试名白名单，经 `test:visual:update` 环境守卫执行；首次 CLI 将 owner 错当可选 mode 而拒绝，未生成图片，改把 `--grep` 放在 owner 前，不修改脚本或绕过守卫。原始日志、expected/actual/diff 保存在忽略目录 `.cache/145-c2-review/`。

Functional 首轮两个失败均保留 trace：语言/history 用例在已打开页面上立即整页 reload，打断 Nuxt 预取；fixture 改为直接从带 query/hash 的地址打开，原 locale/history/scroll/严格资源账本断言全部保留。另一个英文跨断点失焦由上述局部修复解决；中英两条原断言及全部 A2 局部用例复核通过。Collections 手机首轮新增断言曾误用不存在的 CSS 类，改为实际具名 dialog 语义定位，未放宽可见性要求。



实际本地主要命令（除 Go 外在 `apps/cn/web`，Browser 使用最终固定容器生产产物）：

```sh
pnpm run style:policy
pnpm run typecheck
pnpm run test:nuxt tests/nuxt/nav-utilities.nuxt.test.ts tests/nuxt/nav-custom-sites.nuxt.test.ts
pnpm run test:nuxt tests/nuxt/search-platform-picker.nuxt.test.ts
pnpm run build
pnpm run test:browser tests/browser/regression/nav-shell.spec.ts tests/browser/regression/nav-revealed-content.spec.ts tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/game-collections.spec.ts tests/browser/smoke/nav-home-locales.spec.ts tests/browser/regression/hero-lifecycle.spec.ts --workers=1 --grep '145-|Navigation utilities|Localized Home|displayed Hero|B2 Mobile|B2 Quick|Quick Access|Revealed|Core cards|Site Popover|Spotlight'
pnpm run test:browser tests/browser/regression/nav-shell.spec.ts tests/browser/regression/nav-home-header.spec.ts --workers=1 --grep '145-A locale|145-A2'
pnpm run test:visual
# 下面命令只逐个传入白名单 owner 及其精确测试名，非全量更新：
pnpm run test:visual:update --grep '<reviewed test titles>$' <owner>.spec.ts
# 独立只读比较同一白名单，未带 update：
pnpm run test:visual --grep '<all 94 reviewed test titles>$' <15 reviewed owners>
```

构建后因 Collections 真实遮挡、Picker 真实失焦各做一次必要重建；未重复全仓库检查。

## 审查后的 Golden 白名单

只允许以下明确图片变更（同一目录 `tests/browser/visual/__snapshots__/<owner>/`），不全量 update。

| Owner | 数量 | 审查归因 |
| --- | ---: | --- |
| `error-experience.spec.ts` | 4 | 浮层导航入镜，旧 header 占位移除后 viewport/背景构图变化 |
| `footer-shell.spec.ts` | 2 | 页面流变化引起 DOM clip 的 1px 取整差异，Footer 材质/链接不变 |
| `game-collections.spec.ts` | 6 | 浮层入镜、局部入口避让，时间线/数据保持 |
| `game-detail.spec.ts` | 14 | 入口避让、浮层/底栏及视窗 clip/亚像素位置变化，数据/图表不变 |
| `games-home-closure.spec.ts` | 4 | 正常页面流改变局部裁切与亚像素位置，News/Review 数据不变 |
| `games-home.spec.ts` | 4 | 浮层导航与页面入口位置，旧侧 Dock 退役 |
| `games-search.spec.ts` | 11 | 局部避让与背景/clip/亚像素位置，筛选/结果保持 |
| `lottery.spec.ts` | 8 | 浮层入镜、页面流与内容偏移，表单/结果保持 |
| `nav-home-header.spec.ts` | 4 | 已接受的 Hero 全视窗、A2 搜索/平台、C1.1 灰色 PhX 与把手 |
| `nav-revealed-content.spec.ts` | 5 | B 自然流内容位置、C1/C1.1 右岛 Search 替代 Dock 与背景裁切 |
| `nav-shell.spec.ts` | 6 | A–C1 双岛/二层/图标与完整裁切；Bottom Tabs 原 PNG 不变 |
| `site-detail.spec.ts` | 8 | 局部避让、浮层入镜、Tabs 普通流及 Aside 坐标 |
| `site-groups.spec.ts` | 4 | 局部避让与浮层入镜，站点数据保持 |
| `static-pages.spec.ts` | 6 | 浮层入镜与本地内容留白，文案/材质不变 |
| `updates-page.spec.ts` | 8 | 浮层入镜与正常页面流，公告正文保持 |

<details><summary>94 张精确文件名（含 Search 一对一重命名）</summary>

`error-experience.spec.ts`

- `error-404-dark-desktop.png`
- `error-404-dark-mobile.png`
- `error-404-light-desktop.png`
- `error-404-light-mobile.png`

`footer-shell.spec.ts`

- `footer-dark-desktop.png`
- `footer-light-desktop.png`

`game-collections.spec.ts`

- `game-collections-index-dark-desktop.png`
- `game-collection-timeline-dark-desktop.png`
- `game-collection-timeline-dark-mobile.png`
- `game-collections-index-light-desktop.png`
- `game-collection-timeline-light-desktop.png`
- `game-collection-timeline-light-mobile.png`

`game-detail.spec.ts`

- `game-detail-facts-dark-desktop.png`
- `game-detail-facts-light-desktop.png`
- `game-detail-gallery-dark-desktop.png`
- `game-detail-gallery-light-desktop.png`
- `game-detail-insights-dark-desktop.png`
- `game-detail-insights-dark-mobile.png`
- `game-detail-insights-light-desktop.png`
- `game-detail-insights-light-mobile.png`
- `game-detail-news-dark-desktop.png`
- `game-detail-news-light-desktop.png`
- `game-detail-overview-dark-desktop.png`
- `game-detail-overview-dark-mobile.png`
- `game-detail-overview-light-desktop.png`
- `game-detail-overview-light-mobile.png`

`games-home-closure.spec.ts`

- `games-home-latest-reviews-dark-desktop.png`
- `games-home-latest-reviews-light-desktop.png`
- `games-home-news-dark-desktop.png`
- `games-home-news-light-desktop.png`

`games-home.spec.ts`

- `games-home-dark-desktop.png`
- `games-home-dark-mobile.png`
- `games-home-light-desktop.png`
- `games-home-light-mobile.png`

`games-search.spec.ts`

- `games-search-filter-dark-desktop.png`
- `games-search-filter-light-desktop.png`
- `games-search-filter-light-mobile.png`
- `games-search-jump-dark-desktop.png`
- `games-search-jump-light-desktop.png`
- `games-search-results-dark-desktop.png`
- `games-search-results-light-desktop.png`
- `games-home-simple-search-dark-desktop.png`
- `games-home-simple-search-light-desktop.png`
- `games-search-simple-search-dark-desktop.png`
- `games-search-simple-search-light-desktop.png`

`lottery.spec.ts`

- `lottery-activation-fail-dark-mobile.png`
- `lottery-activation-fail-light-mobile.png`
- `lottery-activation-success-dark-desktop.png`
- `lottery-activation-success-light-desktop.png`
- `lottery-page-dark-desktop.png`
- `lottery-page-dark-mobile.png`
- `lottery-page-light-desktop.png`
- `lottery-page-light-mobile.png`

`nav-home-header.spec.ts`

- `nav-home-header-desktop.png`
- `nav-home-quick-sites-modal-desktop.png`
- `nav-home-search-suggestions-desktop.png`
- `nav-home-search-suggestions-mobile.png`

`nav-revealed-content.spec.ts`

- `nav-revealed-core-dark-desktop.png`
- `nav-revealed-core-light-desktop.png`
- `nav-right-island-search-dark-desktop.png`
- `nav-right-island-search-light-desktop.png`
- `nav-spotlight-light-desktop.png`

`nav-shell.spec.ts`

- `nav-shell-mobile-menu-dark.png`
- `nav-shell-mobile-menu-light.png`
- `nav-shell-standard-dark-desktop.png`
- `nav-shell-standard-light-desktop.png`
- `nav-shell-overlay-dark-desktop.png`
- `nav-shell-overlay-light-desktop.png`

`site-detail.spec.ts`

- `site-detail-http-light-desktop.png`
- `site-detail-insights-light-desktop.png`
- `site-detail-overview-light-desktop.png`
- `site-detail-performance-dark-desktop.png`
- `site-detail-performance-light-desktop.png`
- `site-detail-performance-light-mobile.png`
- `site-detail-security-tls-light-desktop.png`
- `site-detail-similar-light-mobile.png`

`site-groups.spec.ts`

- `site-group-dark-desktop.png`
- `site-group-dark-mobile.png`
- `site-group-light-desktop.png`
- `site-group-light-mobile.png`

`static-pages.spec.ts`

- `static-about-dark-desktop.png`
- `static-about-dark-mobile.png`
- `static-about-light-desktop.png`
- `static-about-light-mobile.png`
- `static-terms-dark-mobile.png`
- `static-terms-light-mobile.png`

`updates-page.spec.ts`

- `update-detail-dark-desktop.png`
- `update-detail-dark-mobile.png`
- `update-detail-light-desktop.png`
- `update-detail-light-mobile.png`
- `updates-dark-desktop.png`
- `updates-dark-mobile.png`
- `updates-light-desktop.png`
- `updates-light-mobile.png`

</details>

## 性能与部署交接

沿用 B 的真实测量及限制：SSR HTML 46,383 → 75,810 bytes、正文 SSR 从空变为实际 DOM；首屏存在解析/水合成本，不能宣称整体性能提升。B 的桌面/手机首滚未见长任务；真实网络/低端设备仍需体验确认。本轮不改 SSR、性能预算或基线，不重复无关性能采样。

发布准入目前 **PENDING**，只有最终代码 Fast/Policy、Full 四组与镜像/Gate、完整 Visual 均通过才标 READY。
默认分支是 main。维护者必须确认服务器当前分支确实包含被验收 C2 代码；`git pull --ff-only` 不会把 dev 带入 main。生产若用 main，须先经批准合并并确认该分支 Full；若用 dev，核对接受版本与门禁。
按 [DEPLOYMENT](../../apps/cn/web/DEPLOYMENT.md) 确认工作树、旧镜像/私有配置回退、Compose project `nav-web` 和构建内存，再由维护者从 `apps/cn/web` 执行 `./update.sh`，核验 healthz 及中英文首页、Games/Search、Site/Group 的桌面/手机导航、自然滚动、搜索/弹窗和 hydration。
发生构建失败不强行切换或 prune；Git pull 不是镜像回滚。Issue 保持 Open，生产冒烟后由维护者决定关闭。
