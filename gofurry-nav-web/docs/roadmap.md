# GoFurry Nav Web 样式系统统一路线图

## 背景

`gofurry-nav-web` 当前前端样式的主要问题不是 Tailwind 本身，也不是 Vue SFC scoped CSS 本身，而是样式职责边界长期不清晰：

- Tailwind 原子类、组件 scoped CSS、`:global(.dark ...)`、`:global(.games-page--dark ...)`、`:deep(...)` 同时参与同一批元素的视觉控制。
- 明暗色入口存在重复：全局 `.dark` 与页面级 `games-page--dark` 并存。
- 颜色、字重、hover、active、玻璃质感、边框和阴影缺少统一 token。
- games / games-search 这类复杂页面在持续打磨过程中积累了大量局部覆盖，后续维护成本上升。
- 局部样式常出现“已经修改但实际没生效”的情况，主要由 scoped、global、深度选择器和 Tailwind class 优先级交织造成。

本次路线图目标是通过 Less + Tailwind 的职责分层，逐步清理样式历史包袱，提高亮暗色一致性和后续迭代可维护性。

## 技术方向

采用 **Less + Tailwind v4** 的混合方案，但明确职责边界：

- **Tailwind**：只负责布局、间距、尺寸、响应式、简单显示控制。
- **Less**：负责主题 token、语义组件、复杂视觉、hover、active、暗色、玻璃质感和页面级视觉系统。

不要迁移到 UnoCSS；不要全站一次性重写为纯 Less；不要继续扩大 Tailwind 与 scoped CSS 的无边界混用。

## 样式分层规则

### Tailwind 可以继续负责

- `flex` / `grid` / `gap`
- `px` / `py` / `m` / `p`
- `w` / `h` / `max-w` / `min-h`
- `hidden` / `block` / `inline-flex`
- `sm:` / `md:` / `lg:` / `xl:` / `2xl:`
- `sr-only` / `truncate` / `line-clamp`

### Less 应该负责

- 页面背景色、面板色、文字色、弱文字色
- 明暗色主题
- hover / active / disabled / focus 状态
- 卡片、按钮、tag、输入框、分页、模态框、数据面板
- 玻璃质感、边框、阴影、遮罩、动画
- 第三方组件局部重写，例如 date picker、markdown preview

### 禁止继续扩大的模式

- 在模板里继续堆复杂颜色组合：
  - `bg-orange-50/70 dark:bg-slate-800/70 text-orange-900 dark:text-slate-100`
- 同一组件同时依赖 `dark:`、`:global(.dark ...)`、`:global(.games-page--dark ...)` 控制同一视觉状态。
- 为了解决一个局部颜色问题继续追加更高优先级的 global 覆盖。
- 把页面专属视觉塞进公共组件 Less。

## 建议目录结构

```txt
app/assets/css/main.css
app/assets/styles/
  index.less
  tokens.less
  mixins.less
  components/
    button.less
    card.less
    chip.less
    input.less
    modal.less
    pagination.less
    rating.less
  pages/
    games.less
    games-search.less
    nav.less
    archive.less
```

`main.css` 暂时保留 Tailwind v4 入口和 reset。Less 入口通过 `nuxt.config.ts` 的 `css` 数组接入。

## 主题 token 方向

统一使用 CSS variables，让 Less 负责声明和组织，组件只消费变量。

基础 token 示例：

```less
:root {
  --gf-bg-page: #f6ebdc;
  --gf-bg-grid-line: rgba(126, 92, 58, 0.18);
  --gf-surface: rgba(255, 250, 242, 0.72);
  --gf-surface-strong: rgba(255, 250, 242, 0.88);
  --gf-surface-hover: rgba(255, 239, 213, 0.72);
  --gf-border: rgba(126, 92, 58, 0.16);
  --gf-text-main: rgba(28, 25, 23, 0.94);
  --gf-text-muted: rgba(87, 83, 78, 0.72);
  --gf-text-soft: rgba(120, 113, 108, 0.58);
  --gf-accent: rgba(124, 45, 18, 0.86);
}

html.dark {
  --gf-bg-page: #07111f;
  --gf-bg-grid-line: rgba(56, 189, 248, 0.16);
  --gf-surface: rgba(226, 232, 240, 0.08);
  --gf-surface-strong: rgba(226, 232, 240, 0.14);
  --gf-surface-hover: rgba(226, 232, 240, 0.16);
  --gf-border: rgba(226, 232, 240, 0.18);
  --gf-text-main: rgba(248, 250, 252, 0.94);
  --gf-text-muted: rgba(203, 213, 225, 0.76);
  --gf-text-soft: rgba(148, 163, 184, 0.66);
  --gf-accent: rgba(203, 213, 225, 0.82);
}
```

## 语义类方向

公共组件类示例：

```less
.gf-card {}
.gf-card--interactive {}
.gf-panel {}
.gf-button {}
.gf-button--ghost {}
.gf-button--primary {}
.gf-chip {}
.gf-chip--active {}
.gf-input {}
.gf-pagination {}
.gf-rating {}
```

页面专属类示例：

```less
.games-page {}
.games-shell {}
.games-panel {}
.games-game-card {}
.games-stats-card {}
.games-search-page {}
.games-search-card {}
.games-filter-panel {}
```

公共类只表达通用交互和材质；页面类表达当前页面的布局、密度和特殊视觉。

## 版本计划

### v2.2.0-alpha.1 - Less Foundation And Tokens

- [x] 安装 Less 依赖并确认 Nuxt / Vite 能正常编译 Less。
- [x] 新增 `app/assets/styles/index.less`。
- [x] 新增 `tokens.less`，建立亮暗色 CSS variables。
- [x] 新增 `mixins.less`，封装常见 surface、focus ring、hover、motion mixin。
- [x] 在 `nuxt.config.ts` 中接入 Less 入口。
- [x] 保留 Tailwind v4，不改 Tailwind 入口。
- [x] 文档化 Tailwind / Less 职责边界。

验收标准：

- `npm run typecheck` 通过。
- `npm run build` 通过。
- 页面视觉不应发生明显变化。
- 新 token 可在任意组件中被消费。

### v2.2.0-alpha.2 - Games Style Token Migration

- [x] 优先迁移 `/games` 页面。
- [x] 将游戏卡片、信息面板、右侧栏、数据面板、更新情报、分页、按钮、评分迁移到 Less 语义类。
- [x] Tailwind 仅保留布局和响应式。
- [x] 清理 games 页内重复的颜色、字重、hover、暗色 class。
- [x] 统一 games 页亮暗色只走 `html.dark` token。

验收标准：

- `/games` 亮色视觉保持当前稳定效果。
- `/games` 暗色文字、评分、hover、数据面板、右侧栏不再出现局部失效。
- 不继续增加新的 `:global(.games-page--dark ...)` 覆盖。

### v2.2.0-alpha.3 - Games Search Style Token Migration

- [ ] 迁移 `/games/search` 页面。
- [ ] 将搜索卡片、筛选按钮、筛选弹窗、tag、时间选择器、分页、点评按钮迁移到 Less 语义类。
- [ ] 清理 `GameSearchResult.vue` 和 `GameSearchFilter.vue` 中的复杂 scoped CSS。
- [ ] 保留 Tailwind grid 和响应式列数。
- [ ] 统一英文 tag 截断、标题优先级和亮暗色字重。

验收标准：

- 大屏一行 5 个游戏卡片，中屏 4 个，小屏 2 个或 1 个。
- 英文长标题优先省略，tag 不被挤碎。
- 亮暗色卡片字重和文字层级一致。
- 筛选弹窗暗色标题、tag、输入框、时间选择器全部走同一 token。

### v2.2.0-beta.1 - Shared Component Classes

- [ ] 抽取通用 `.gf-button`、`.gf-card`、`.gf-chip`、`.gf-input`、`.gf-pagination`、`.gf-rating`。
- [ ] 迁移 NavBar、Footer、通用 Modal、RatingStar。
- [ ] 收敛导航 active / hover / 语言 / 明暗色按钮样式。
- [ ] 避免公共组件继续依赖页面级暗色类。

验收标准：

- NavBar 中英文路径 active 状态正常。
- Footer 亮暗色一致。
- RatingStar 在亮暗色下评分文字清晰，不再靠局部覆盖修。

### v2.2.0-beta.2 - Dark Mode Entrypoint Cleanup

- [ ] 全站统一暗色入口为 `html.dark`。
- [ ] 逐步移除 `games-page--dark`、`search-results--dark` 等重复暗色状态。
- [ ] 清理可替换的 `dark:` 复杂颜色类。
- [ ] 保留必要的 Tailwind `dark:` 仅用于极轻量布局或显示差异。

验收标准：

- 切换主题后所有页面使用同一暗色 token。
- 不再出现同一组件亮暗色字重明显不一致的问题。
- 新增组件不需要写页面级 dark 覆盖。

### v2.2.0-rc.1 - Visual Regression And Documentation

- [ ] 更新性能与视觉回归脚本覆盖 `/games`、`/games/search`。
- [ ] 增加亮色与暗色截图检查说明。
- [ ] 补充 `docs/style-system.md`，记录样式分层、命名、token、迁移规则。
- [ ] 整理迁移前后重点组件列表。

验收标准：

- `npm run typecheck` 通过。
- `npm run build` 通过。
- `npm run perf:guard` 不因样式迁移出现明显回退。
- 手工检查 `/games`、`/games/search` 亮暗色、英文/中文、桌面/移动端。

### v2.2.0 - Style System Mainline

- [ ] Less + Tailwind 分层成为主线规范。
- [ ] games 与 games-search 完成样式主线迁移。
- [ ] 新增样式必须优先使用 token 和语义类。
- [ ] 保留 Tailwind 作为布局工具，不再作为复杂视觉主线。

## 后续计划

### v2.3.0 - Nav And Static Pages Migration

- [ ] 迁移首页导航模块。
- [ ] 迁移 about / terms / privacy / updates。
- [ ] 清理静态页中重复的 page vars 和局部面板样式。

### v2.4.0 - Archive Style Isolation

- [ ] 单独整理 `/archive` 样式系统。
- [ ] 由于 archive 更像完整应用界面，应保留独立页面层，但复用全站 token。
- [ ] 清理 markdown preview、citation、session sidebar 的深度选择器。

## 维护规则

- 新增视觉样式优先写 Less 语义类。
- 新增复杂颜色必须先进入 token。
- 不允许为了局部修复继续追加高优先级 global 覆盖。
- 每次迁移一个页面或一组组件，不做全站大爆炸重写。
- 每阶段完成后运行 `npm run typecheck` 和必要页面手工检查；构建按阶段或用户要求执行。
