# GoFurry Nav Web 样式系统规范

## 目标

`gofurry-nav-web` 的样式主线采用 Less + Tailwind v4 分层方案：

- Tailwind 负责布局、间距、尺寸、响应式和简单显示控制。
- Less 负责主题 token、语义组件、复杂视觉、状态、暗色和局部第三方组件重写。

本规范的目标是减少同一视觉状态被 Tailwind、scoped CSS、`:global(...)` 和页面级暗色类同时控制的情况。

## 样式入口

- Tailwind 入口保留在 `app/assets/css/main.css`。
- Less 入口为 `app/assets/styles/index.less`，通过 `nuxt.config.ts` 的 `css` 数组全局加载。
- 基础 token 位于 `app/assets/styles/tokens.less`。
- 常用 Less mixin 位于 `app/assets/styles/mixins.less`。

## 主题规则

统一使用 CSS variables 作为样式 token。组件、页面和后续语义类只消费 token，不直接复制复杂颜色组合。

暗色入口统一为：

```less
html.dark {
  ...
}
```

不要为新样式新增页面级暗色入口，例如 `.games-page--dark`、`.search-results--dark`。迁移旧样式时，应逐步删除这些重复入口。

## Tailwind 使用边界

Tailwind 可以继续用于：

- `flex`、`grid`、`gap`
- `px`、`py`、`m`、`p`
- `w`、`h`、`max-w`、`min-h`
- `hidden`、`block`、`inline-flex`
- `sm:`、`md:`、`lg:`、`xl:`、`2xl:`
- `sr-only`、`truncate`、`line-clamp`

Tailwind 不应继续承载复杂视觉，例如同一元素上同时堆叠亮暗色背景、文字、边框、hover 和 active class。

## Less 使用边界

Less 应优先用于：

- 页面背景、面板、卡片、按钮、chip、输入框、分页和模态框。
- 文字层级、弱文字、hover、active、disabled、focus 状态。
- 玻璃质感、边框、阴影、遮罩、动画。
- 页面专属视觉系统。
- 第三方组件局部重写。

复杂视觉应先提炼 token，再由页面类或语义类消费。

## 命名规则

公共语义类使用 `gf-` 前缀：

```less
.gf-card {}
.gf-button {}
.gf-chip {}
.gf-input {}
.gf-pagination {}
.gf-rating {}
```

页面专属类使用页面前缀：

```less
.games-page {}
.games-game-card {}
.games-search-page {}
.games-search-card {}
.games-filter-panel {}
```

公共类只表达通用材质和交互；页面类表达当前页面密度、布局和特殊视觉。

## 迁移检查项

迁移组件时按以下顺序检查：

1. Tailwind 是否只保留布局、间距、响应式和简单显示控制。
2. 颜色、边框、阴影、hover、active、focus 是否来自 Less token 或语义类。
3. 暗色是否只通过 `html.dark` token 生效。
4. 是否删除了可替换的 `:global(.games-page--dark ...)`、`:deep(...)` 或重复覆盖。
5. 亮暗色下文字层级和字重是否一致。

阶段性验证：

```powershell
npm run typecheck
npm run build
```
