# GoFurry Nav Web Performance Roadmap

## 当前结论

当前前端性能问题不是单个线上故障，而是多个“常驻型成本”叠加：

- 首页首屏下方内容虽然被隐藏，但已经挂载、水合和加载资源。
- 多个轻内容页使用同一个高成本 fixed grid 背景。
- 知识库页 `/archive` 有全屏 blur + 无限动画，GPU 压力明显。
- 全局滚动按钮在多数页面常驻，并会扫描页面滚动容器。
- 部分重型库同步进入页面，例如 Markdown preview、ECharts、HLS。

优先处理目标：

1. 降低 `/archive` 的持续 GPU 占用。
2. 降低首页初始内存、DOM 和图片资源压力。
3. 降低轻内容页的基础 GPU 成本。
4. 推动重型组件按需加载，避免影响无关页面。

## 诊断摘要

### 首页

首页的主要问题是内容延迟显示，但没有延迟挂载。

`NavHomePage.vue` 当前使用 `v-show="isContentRevealed"` 控制内容区显示。`v-show` 只切换 CSS 显示状态，不会阻止子组件初始化。因此用户刚进入首页时，以下内容已经参与渲染和水合：

- `NavContent`
- `NavSpotlightPanels`
- 全量站点分组和站点卡片
- 大量站点图标
- 站点 hover popover 相关状态
- 首页 ping 定时刷新
- `NavToolDock`
- 内容区 grid 背景

这能解释首页比内容较少页面更容易出现高内存占用。

### 全局和通用背景

`GoFurryGridBackground.vue` 没有持续动画，但默认是 fixed 背景，并使用：

- 多层 `radial-gradient`
- 两层 `repeating-linear-gradient`
- `mask-image`
- 伪元素
- `will-change: opacity`

这类 CSS 在大屏上容易形成常驻合成层。它被用于 `/updates`、`/about`、`/terms`、`/privacy`、站点详情、游戏页等页面，所以即便页面内容很少，也会有固定 GPU 成本。

### 知识库页

`/archive` 是当前最明确的 GPU 热点。

页面包含：

- 大面积 `filter: blur(34px)`
- `archiveGlowDrift` 无限动画
- `archiveMistSweep` 无限动画
- `backdrop-filter`
- `conic-gradient` 滚动进度按钮
- Markdown preview 渲染
- SSE 流式回答期间的频繁状态更新
- 每 8 秒队列状态轮询

这能解释知识库页截图里 GPU 占用明显偏高。这里的核心问题不是内容多少，而是全屏动画背景和 blur 的持续合成压力。

### 全局滚动按钮

`PageScrollDock` 在除 `/archive` 外的页面全局挂载。

它会：

- 使用 `ResizeObserver` 观察 `documentElement` 和 `body`
- 在页面中扫描 `body *` 查找可滚动容器
- 监听 window scroll/resize/load
- 使用 `conic-gradient` 和 `backdrop-filter`

它不是最大问题，但会让所有页面都有额外基础成本，尤其是轻内容页面。

### 重型依赖

当前依赖中存在多个重型库：

- `md-editor-v3`
- `@kangc/v-md-editor`
- `echarts`
- `hls.js`
- `dashjs`
- `html2canvas`
- `dom-to-image-more`

其中 `/archive` 同步引入 `md-editor-v3`，站点详情性能图同步引入 `echarts`，游戏详情图库同步引入 `hls.js`。这些库应该尽量限定在真正需要的交互路径中。

## v2.1.0-alpha.1 - Archive GPU Baseline

目标：优先压低知识库页 `/archive` 的持续 GPU 占用。

状态：已完成。

改进项：

- 移除或默认关闭 `/archive` 的全屏无限动画背景。
- 将 `archiveGlowDrift`、`archiveMistSweep` 改为静态背景或低成本 CSS。
- 去掉大面积 `filter: blur(34px)`，改用静态渐变或低透明度背景。
- 为 `/archive` 增加 motion profile：
  - normal：保留少量过渡。
  - reduced：无背景动画、无 shimmer。
- 对 `answer-placeholder` 的 shimmer 做限制，只在真实等待回答时启用。
- 对 SSE 流式回答做增量缓冲，避免每个 token 都触发回答区重渲染。
- 引用区 Markdown 预览改为展开引用时再挂载，减少长回答页面的初始渲染成本。
- 去掉 `/archive` 常驻滚动圆环的 `backdrop-filter`。

验收标准：

- `/archive` 空会话页面静置时 GPU 占用显著下降。
- 关闭动画后页面视觉仍保持知识库页的深色沉浸感。
- 流式回答时不因每个 token 触发过多 Markdown 重渲染。

## v2.1.0-alpha.2 - Home Initial Render Split

目标：降低首页初始内存、DOM 和图片加载压力。

状态：已完成。

改进项：

- 将首页内容区从 `v-show` 改为真正的延迟挂载。
- 移除 `SearchBox` 下方低价值的羽化模糊伪元素，避免首屏常驻 `filter: blur(...)`。
- 首屏只挂载：
  - `NavHeader`
  - `SearchBox`
  - 必要的首屏背景
  - 必要的首屏导航按钮
- 用户滚动、触摸、键盘触发 reveal 后，再挂载：
  - `NavContent`
  - `NavSpotlightPanels`
  - `NavTransitionBar`
  - 内容区 `GoFurryGridBackground`
- `NavToolDock` 只在内容 reveal 后挂载。
- 首页站点图标增加 `loading="lazy"` 和稳定尺寸。
- `NavContent` 分组块增加 `content-visibility: auto` 和合理的 intrinsic size，降低视窗外布局/绘制成本。
- `NavContent` 与 `NavSpotlightPanels` 的站点图标增加懒加载、异步解码和稳定尺寸。
- `NavSpotlightPanels` 保留原有玻璃质感，仅补充视窗外渲染抑制和图片懒加载。
- 站点列表增加分组级渐进挂载，默认展示前 4 组，滚动接近底部时每批追加 2 组。

验收标准：

- 首页首屏打开时 DOM 节点数明显下降。
- 首页首屏内存低于当前版本。
- reveal 后内容完整，不影响站点浏览体验。
- 移动端仍然自动 reveal 内容。

## v2.1.0-alpha.3 - Lightweight Background System

目标：降低轻内容页和列表页的常驻背景绘制成本。

状态：已完成。

改进项：

- 为 `GoFurryGridBackground` 增加低成本模式：
  - 不使用 fixed。
  - 不使用 mask。
  - 不使用 `will-change`。
  - 使用静态图片或简单 linear-gradient。
- `GoFurryGridBackground profile="light"` 使用组件自身的静态背景层绘制网格，不挂载伪元素 mask。
- 轻内容页默认使用低成本背景：
  - `/updates`
  - `/about`
  - `/terms`
  - `/privacy`
- 复杂页面再按需使用完整 grid 背景：
  - 首页内容区
  - 游戏模块
  - 站点详情页
- 对 `backdrop-filter` 做收敛，避免多个 fixed/floating 元素叠加。

验收标准：

- `/updates` 静置 GPU 占用下降。
- 轻内容页的视觉风格仍与 GoFurry 保持一致。
- 背景组件有明确的性能档位，而不是所有页面统一使用最高成本版本。

## v2.1.0-alpha.4 - Global Widget Cost Control

目标：降低全局小组件对所有页面的基础影响。

状态：已完成。

改进项：

- `PageScrollDock` 不再默认扫描 `body *`。
- 优先只监听 document scroll。
- 需要页面内部滚动容器时，由页面显式声明或传入选择器。
- 仅当页面高度超过阈值后再挂载滚动按钮。
- 去掉滚动按钮里的 `backdrop-filter` 或提供低成本样式。
- 避免在 resize 时频繁触发全量滚动容器解析。
- 滚动进度更新使用 `requestAnimationFrame` 合并，降低滚动中的状态更新频率。

验收标准：

- 内容很少的页面不挂载滚动按钮。
- 长页面滚动进度仍可用。
- 不再因为全局组件让所有页面都有额外 DOM 扫描成本。

## v2.1.0-beta.1 - Heavy Dependency Lazy Loading

目标：重型库只在用户真正进入对应功能时加载。

状态：已完成。

改进项：

- `/archive` 的 `md-editor-v3` 改为异步组件或仅在存在回答/引用时加载。
- 站点详情的 `echarts` 改为按需动态 import。
- 游戏详情图库的 `hls.js` 改为用户点击视频时动态 import。
- `GameTabGallery` 默认优先展示截图，避免进入图库时自动初始化 HLS。
- 清理未使用的重型依赖：
  - `@kangc/v-md-editor`
  - `dashjs`
  - `html2canvas`
  - `dom-to-image-more`
  - `file-saver`
- 检查 chunk 拆分，避免首页或普通页面被重型库污染。

验收标准：

- 首页 chunk 不包含 Markdown、ECharts、HLS。
- `/archive` 空页面不加载 Markdown preview 库。
- 游戏详情进入图片 tab 不自动初始化 HLS。
- 依赖清理不影响已有功能。

## v2.1.0-rc.1 - Measurement And Regression Guard

目标：建立前端性能验证方式，避免优化后回退。

状态：已完成。

验证页面：

- `/`
- `/` reveal 后内容区
- `/updates`
- `/about`
- `/archive`
- `/games`
- `/sites/{id}`

已新增：

- `npm run perf:measure`：生成中文 Markdown 与 JSON 测量报告。
- `npm run perf:guard`：按 `docs/performance/budget.json` 执行回归守卫。
- `npm run perf:trace`：生成 Chrome trace，用于人工分析 paint/composite/long task。
- `npm run perf:baseline`：刷新当前可接受性能基线。
- `docs/performance/budget.json`：保守预算阈值。
- `docs/performance/baseline.json`：当前接受基线。
- `docs/performance/README.md`：中文使用说明。

记录指标：

- JS 请求数与 JS 传输体积。
- 初始 DOM 节点数量。
- 图片请求数与图片传输体积。
- Long Task 数量与总耗时。
- JS Heap。
- CDP Performance metrics。
- 首页首屏是否误加载 `md-editor-v3`、`echarts`、`hls.js`。
- `/archive` 空页是否误加载 `md-editor-v3`。
- `/games` 首屏是否误加载 `hls.js`。

验收标准：

- 本地 3000 端口启动后，性能脚本可完成默认页面测量。
- 后端接口不可用时，动态详情页进入 warning，不阻断核心页面报告。
- 首页首屏不得加载 Markdown、ECharts、HLS。
- `/archive` 空页不得提前加载 Markdown preview。
- `/games` 首屏不得提前加载 HLS。
- GPU 占用不做自动硬阈值；用 trace 和人工观察补充判断。

## 风险点

- 首页延迟挂载可能影响 SSR 内容完整度，需要确认 SEO 需要的内容是否仍可被服务端输出。
- 背景降级可能影响视觉一致性，需要用统一的低成本设计替代，而不是简单删除。
- Markdown preview 异步加载可能影响首条回答出现速度，需要 loading 状态兜底。
- HLS 动态加载需要处理用户首次点击视频时的等待状态。
- `PageScrollDock` 改为页面显式声明后，部分内部滚动页可能需要补配置。

## 推荐执行顺序

1. 先处理 `/archive` 背景动画和 blur。
2. 再处理首页 `v-show` 导致的提前挂载。
3. 然后改造 `GoFurryGridBackground` 的低成本模式。
4. 再收敛 `PageScrollDock`。
5. 最后做重型依赖懒加载和依赖清理。

这个顺序能先解决截图里最明显的 GPU 问题，再逐步处理首页内存和全站基础成本。
