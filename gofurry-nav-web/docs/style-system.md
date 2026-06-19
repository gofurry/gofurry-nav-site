# GoFurry Nav Web 样式系统规范

`gofurry-nav-web` 的样式主线采用 Less + Tailwind v4 分层方案。Tailwind 保留为布局工具，Less 承担主题 token、语义组件、复杂视觉、状态、暗色入口和局部第三方组件重写。

本规范的目标是减少同一视觉状态被 Tailwind、scoped CSS、`:global(...)` 和页面级暗色类同时控制的情况，让后续维护有固定的入口、命名和验收方式。

## 迁移完成状态

`v2.2.x` 样式系统迁移已经完成。新增页面或组件不再进入“后续迁移”队列，而是必须直接遵守本文的 Less + Tailwind 分层规则、`html.dark` 暗色入口和视觉/性能守卫。

## 样式入口

- Tailwind 入口保留在 `app/assets/css/main.css`。
- Less 入口为 `app/assets/styles/index.less`，通过 `nuxt.config.ts` 的 `css` 数组全局加载。
- 基础 token 位于 `app/assets/styles/tokens.less`。
- 常用 Less mixin 位于 `app/assets/styles/mixins.less`。
- 公共组件样式位于 `app/assets/styles/components/`。
- 页面样式位于 `app/assets/styles/pages/`。

`index.less` 的导入顺序必须保持为：token、mixin、公共组件、页面样式。页面样式可以消费公共 token 和组件类，但公共组件样式不能反向依赖页面类。

## 分层职责

Tailwind 可以继续用于：

- `flex`、`grid`、`gap`
- `px`、`py`、`m`、`p`
- `w`、`h`、`max-w`、`min-h`
- `hidden`、`block`、`inline-flex`
- `sm:`、`md:`、`lg:`、`xl:`、`2xl:`
- `sr-only`、`truncate`、`line-clamp`

Less 应优先用于：

- 页面背景、面板、卡片、按钮、chip、输入框、分页和模态框。
- 文字层级、弱文字、hover、active、disabled、focus 状态。
- 玻璃质感、边框、阴影、遮罩、动画。
- 页面专属视觉系统。
- 第三方组件局部重写。

不要在同一元素上继续堆叠复杂的 `bg-*`、`text-*`、`border-*`、`hover:*`、`dark:*` 来表达一个完整组件状态。遇到这种情况，应提炼为 token 和语义类。

## 主题与 Token

统一使用 CSS variables 作为样式 token。组件、页面和后续语义类只消费 token，不直接复制复杂颜色组合。

暗色入口统一为：

```less
html.dark {
  ...
}
```

不要新增页面级暗色入口，例如 `.games-page--dark`、`.search-results--dark`、`.spotlight-panels--dark`、`.is-dark-theme`。迁移旧样式时，应删除这些重复状态，并把差异收敛到 `html.dark` 下的 token 或选择器。

Token 分三层：

- 全站 token：`--gf-*`，定义基础背景、文字、边框、阴影、品牌色和状态色。
- 公共组件 token：`--gf-button-*`、`--gf-card-*`、`--gf-modal-*` 等，只服务对应 `gf-*` 类。
- 页面 token：`--games-*`、`--games-search-*`、`--updates-*`、`--about-*`、`--legal-*` 等，只在对应页面根类下声明。

公共组件 token 可以引用全站 token；页面 token 可以引用全站 token 和少量公共组件 token。公共 `gf-*` 类不能引用页面 token。

## 命名规则

公共语义类使用 `gf-` 前缀：

```less
.gf-app-shell {}
.gf-card {}
.gf-button {}
.gf-chip {}
.gf-input {}
.gf-modal {}
.gf-pagination {}
.gf-rating {}
.gf-nav {}
.gf-footer {}
```

页面专属类使用页面前缀：

```less
.games-page {}
.game-card {}
.games-search-page {}
.search-page-card {}
.game-search-filter-panel {}
.lottery-page {}
.updates-page {}
.steam-zone-page {}
.gf-static-page {}
```

公共类只表达通用材质和交互；页面类表达当前页面密度、布局和特殊视觉。新类优先使用 BEM 风格后缀，例如 `__title`、`__body`、`--active`、`--disabled`。

## 已迁移重点组件

| 范围 | 组件或页面 | 样式入口 | 迁移后约束 |
| --- | --- | --- | --- |
| 游戏首页 | `app/pages/games/index.vue` | `pages/games.less` | 页面背景、卡片、新闻、榜单、侧栏搜索和工具按钮走 `.games-page` token。 |
| 游戏首页卡片 | `GameInfoGroup.vue`、`GameStatsPanels.vue`、`GameSidebarSearch.vue`、`GameToolDock.vue` | `pages/games.less` | Tailwind 保留布局与尺寸，复杂 hover、边框、阴影和文字层级由 Less 管理。 |
| 游戏详情页 | `app/pages/games/[id].vue`、`components/game/detail/*` | `pages/games.less` | 根类为 `.games-page.game-detail-page`，详情 hero、tabs、侧栏、评论、新闻和 gallery 走 `.game-detail-*` 语义类。 |
| 游戏搜索页 | `app/pages/games/search.vue` | `pages/games-search.less` | 根类为 `.games-search-page`，复用 `.games-page` 基础但搜索页 token 独立。 |
| 搜索结果 | `GameSearchResult.vue` | `pages/games-search.less` | 结果卡片使用 `.search-page-card`，数据为空时仍保留网格容器便于截图守卫。 |
| 搜索筛选 | `GameSearchFilter.vue` | `pages/games-search.less` | 筛选面板、chip、日期选择器重写都收敛在搜索页入口。 |
| 游戏抽奖 | `games/prize/index.vue`、`games/prize/activation.vue`、`LotteryJoinModal.vue` | `pages/lottery.less` | 抽奖页、激活页和提交弹窗使用 `.lottery-page`、`.lottery-activation-page`、`.lottery-modal` 页面 token，模板只保留布局和尺寸类。 |
| 分页 | `GamePagination.vue` | `components/pagination.less`、`pages/games-search.less` | 通用分页外观走 `.gf-pagination`，搜索页密度和 active 细节走页面类。 |
| 全站导航 | `NavBar.vue` | `components/nav.less`、`components/shell.less` | active、按钮、移动面板使用 `gf-nav` token，不再依赖页面级 dark class。 |
| 页脚 | `Footer.vue` | `components/footer.less` | 链接、分组标题和 meta 文字由 `gf-footer` token 控制。 |
| 模态与输入 | `ModeSettingModal.vue`、`NsfwConfirmModal.vue` | `components/modal.less`、`components/input.less`、`components/button.less`、`components/chip.less` | 模态框材质、输入框、按钮和 toggle 统一为 `gf-*` 组件类。 |
| 评分 | `RatingStar.vue` | `components/rating.less` | 星级和评分文字走 `.gf-rating`，不再由页面局部覆盖修暗色。 |
| 静态页与更新页 | `about.vue`、`terms.vue`、`privacy.vue`、`updates.vue`、`components/updates/*` | `pages/static.less`、`pages/updates.less` | `.gf-static-page` 承接静态页共享结构，updates summary、状态、年份分组和时间线条目全部由页面 Less token 控制。 |
| 首页导航 | `NavHomePage.vue`、`NavHeader.vue`、`SearchBox.vue`、`NavQuickAccess.vue`、`NavContent.vue`、`SitePopover.vue`、`GroupPopover.vue`、`NavToolDock.vue`、`NavSpotlightPanels.vue` | `pages/nav.less` | 首页根类为 `.nav-home-page`，搜索、快捷入口、内容区、站点卡片、popover、工具栏和 spotlight 复杂视觉统一由页面 token 与 `html.dark` 控制。 |

## 维护检查项

新增或调整组件样式时按以下顺序检查：

1. Tailwind 是否只保留布局、间距、响应式和简单显示控制。
2. 颜色、边框、阴影、hover、active、focus 是否来自 Less token 或语义类。
3. 暗色是否只通过 `html.dark` token 生效。
4. 是否删除了可替换的 `:global(.games-page--dark ...)`、`:global(.search-results--dark ...)`、`:deep(...)` 或重复覆盖。
5. 亮暗色下文字层级和字重是否一致。
6. 页面专属样式是否只落在对应 `pages/*.less`，公共组件样式是否只落在 `components/*.less`。
7. 固定格式 UI，如分页、卡片网格、按钮和筛选 chip，是否有稳定尺寸或响应式约束，避免内容变化造成布局跳动。

## 保留例外

迁移后仍允许少量 `:global(...)` 和 `:deep(...)`，但必须有明确边界：

| 文件 | 保留项 | 原因 | 后续处理 |
| --- | --- | --- | --- |
| `components/common/GoFurryGridBackground.vue` | `:global(html.dark ...)` | 通用背景组件需要跟随全站 `html.dark`，并通过组件根类限制影响范围。 | 保留。 |
| `components/site/*` | `:global(html.dark ...)` | 站点详情观测面板仍处于独立页面迁移之外，本阶段先统一暗色入口，不再保留 `.dark` 简写。 | 后续若迁移站点详情页，再沉淀到页面 Less。 |
| `components/site/SiteDetailPage.vue` | `:deep(...)` | 站点详情页需要包裹子面板生成的 Tailwind 表面，当前只能通过父级边界局部覆盖。 | 保留登记，禁止扩散到其他页面。 |
| `components/site/SitePerformancePanel.vue` | `:deep(...)` | 性能面板包装共享 `SitePerformance` 子组件，局部覆盖用于约束子组件布局和材质。 | 保留登记，禁止扩散到其他页面。 |

`visual:guard` 会在截图前扫描源码：新增废弃暗色 class、`:global(.dark ...)`、未登记 `:deep(...)` 或已迁移游戏详情/抽奖页复杂颜色类都会使守卫失败。

## 视觉回归范围

当前固定检查范围：

- `/`：中文亮色、中文暗色、桌面 `1440x900`、移动 `390x844`。
- `/en`：英文亮色、英文暗色、桌面 `1440x900`、移动 `390x844`。
- `/games`：中文亮色、中文暗色、桌面 `1440x900`、移动 `390x844`。
- `/games/1`：中文亮色、中文暗色、桌面 `1440x900`、移动 `390x844`。
- `/games/prize`：中文亮色、中文暗色、桌面 `1440x900`、移动 `390x844`。
- `/games/prize/activation?status=success`：中文亮色、中文暗色、移动 `390x844`。
- `/games/search`：中文亮色、中文暗色、桌面 `1440x900`、移动 `390x844`。
- `/en/games/search`：英文亮色、英文暗色、桌面 `1440x900`、移动 `390x844`。
- `/about`、`/en/about`：亮色、暗色、桌面 `1440x900`、移动 `390x844`。
- `/updates`、`/en/updates`：亮色、暗色、桌面 `1440x900`、移动 `390x844`。
- `/steam`：兽游专区，亮色、暗色、桌面 `1440x900`、移动 `390x844`。
- `/terms`、`/privacy`、`/en/terms`、`/en/privacy`：亮色、暗色、移动 `390x844`。

自动截图入口：

```powershell
npm run visual:guard -- --base-url http://localhost:3000
```

输出位于 `docs/performance/reports/`，该目录默认被 `.gitignore` 忽略。截图报告重点确认主题入口、关键容器、旧暗色类和横向溢出；数据卡片为空时只记录 warning。

## 阶段性验证

每次完成页面视觉、主题入口或公共组件样式调整后运行：

```powershell
npm run typecheck
npm run build
npm run preview -- --port 3001
npm run perf:guard -- --base-url http://localhost:3001
npm run visual:guard -- --base-url http://localhost:3001
```

`perf:guard` 用来防止样式迁移带来明显性能回退；`visual:guard` 用来保证亮暗色入口、路由和响应式截图仍可复核。
