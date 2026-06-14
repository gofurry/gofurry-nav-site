# GoFurry Nav Web 样式系统迁移 Handoff

## 当前任务

准备在新的 session 中启动 `gofurry-nav-web` 的样式规范统一与迁移。

目标不是换掉 Tailwind，而是把项目样式主线调整为：

> Less 负责主题 token、语义组件、复杂视觉；Tailwind 负责布局、间距、响应式和简单工具类。

本文件用于帮助新的 session 快速理解当前问题、决策、约束和下一步。

## 当前仓库状态提示

开始新 session 后第一步请先执行：

```powershell
git status --short
```

当前工作区在上一轮 games / games-search 视觉打磨后存在未提交改动。不要盲目覆盖，先阅读 diff，再继续迁移。

重点可能涉及的文件：

- `gofurry-nav-web/app/components/NavBar.vue`
- `gofurry-nav-web/app/components/Footer.vue`
- `gofurry-nav-web/app/components/common/RatingStar.vue`
- `gofurry-nav-web/app/components/game/main/content/GameInfoGroup.vue`
- `gofurry-nav-web/app/components/game/main/sidebar/GameSidebarActions.vue`
- `gofurry-nav-web/app/components/game/main/sidebar/GameSidebarSearch.vue`
- `gofurry-nav-web/app/components/game/search/GameSearchFilter.vue`
- `gofurry-nav-web/app/components/game/search/GameSearchResult.vue`
- `gofurry-nav-web/app/pages/games/index.vue`
- `gofurry-nav-web/app/pages/games/search.vue`

## 已知上下文

### 当前样式技术栈

- Nuxt 4
- Vue 3
- Tailwind v4，通过 `@tailwindcss/vite` 接入
- 全局 CSS 入口：`app/assets/css/main.css`
- 目前没有 Less / Sass / SCSS / UnoCSS 配置

`main.css` 当前包含：

- `@import "tailwindcss";`
- `@custom-variant dark (&:where(.dark, .dark *));`
- 基础 reset
- scrollbar 样式

### 当前主要问题

- Tailwind 原子类和 scoped CSS 混用边界不清。
- 同一元素的颜色、字重、hover、active 可能同时受 Tailwind、scoped CSS、`:global(...)` 控制。
- 暗色入口重复：`.dark`、`.games-page--dark`、局部 `--dark` class 并存。
- `:global(...)` 和 `:deep(...)` 数量较多，导致覆盖链不透明。
- games / games-search 是当前最明显的样式债集中区。

### 最近修过的问题

- `/games/search` 英文 tag 被长标题挤碎：已改成标题主动省略，tag 保留空间。
- `/games/search` 亮暗色字重差异：筛选 chip 已显式收敛字重。
- 英文路由下 NavBar active 判断：已修复 `/en/games` 同时激活 Navigation 和 Games 的问题。
- NavBar Logo 跳转已改为 `localePath('/')`，避免英文模式点 Logo 回到中文根路径。

这些修复还没有迁移到 Less 体系，新 session 需要保留这些行为。

## 用户偏好与约束

- 用户对视觉一致性很敏感，尤其是亮暗色、hover、字重、卡片质感。
- 不要想当然做大幅视觉重设；迁移初期应保持现有稳定视觉。
- 用户不喜欢频繁跑构建和测试；只有必要时运行。
- 用户希望减少历史包袱，不希望继续维护多套暗色入口和大量局部覆盖。
- 改动应小步进行，每个阶段可审查。

## 推荐第一阶段执行步骤

### 1. 检查当前改动

```powershell
git status --short
git diff -- gofurry-nav-web/app/components/game/search/GameSearchResult.vue
git diff -- gofurry-nav-web/app/components/game/search/GameSearchFilter.vue
git diff -- gofurry-nav-web/app/components/NavBar.vue
```

确认上一轮视觉修复没有被覆盖。

### 2. 接入 Less

建议在 `gofurry-nav-web` 下安装：

```powershell
npm install -D less
```

然后新增：

```txt
app/assets/styles/index.less
app/assets/styles/tokens.less
app/assets/styles/mixins.less
app/assets/styles/components/
app/assets/styles/pages/
```

在 `nuxt.config.ts` 中接入：

```ts
css: [
  '~/assets/css/main.css',
  '~/assets/styles/index.less'
]
```

### 3. 建立 token，但不改变视觉

优先只做变量，不迁移组件：

- `--gf-bg-page`
- `--gf-surface`
- `--gf-surface-strong`
- `--gf-surface-hover`
- `--gf-border`
- `--gf-text-main`
- `--gf-text-muted`
- `--gf-text-soft`
- `--gf-accent`
- `--gf-focus-ring`

暗色只写：

```less
html.dark {
  ...
}
```

不要再新增 `.games-page--dark` 作为主题入口。

### 4. 先迁移 games / games-search

优先顺序：

1. `/games/search` 卡片、tag、筛选弹窗。
2. `/games` 游戏卡片、数据面板、右侧栏。
3. 公共 RatingStar。
4. NavBar / Footer。

每次迁移一个组件，迁移后删除对应的复杂 `:global(.games-page--dark ...)` 覆盖。

## 不建议做的事

- 不要一次性把全站 Tailwind class 全部改成 Less。
- 不要迁移到 UnoCSS。
- 不要新增第三套主题状态 class。
- 不要为了一个暗色问题继续追加更高优先级的 `:global(...)`。
- 不要把 games 页专属样式抽成全局公共组件。

## 验证建议

阶段性验证即可：

```powershell
cd gofurry-nav-web
npm run typecheck
```

构建建议在阶段完成或用户明确要求时运行：

```powershell
npm run build
```

手工重点检查：

- `/games`
- `/games/search`
- `/en/games/search`
- 明暗色切换
- 中文 / 英文切换
- 桌面宽屏、中屏、小屏

## 当前 roadmap

详细阶段计划见：

```txt
gofurry-nav-web/docs/roadmap.md
```

新的 session 应优先从 `v2.2.0-alpha.1 - Less Foundation And Tokens` 开始。
