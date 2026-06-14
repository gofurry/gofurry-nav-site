# GoFurry Nav Web 性能测量与回归守卫

这套脚本用于样式系统迁移回归。目标是让前端性能与样式迁移有可重复的本地验证方式，防止首页、知识库页、游戏页和游戏搜索页重新误加载重型依赖，也为亮暗色截图检查留下固定入口。

## 使用方式

快速测量可以直接启动本地前端：

```bash
npm run dev
```

然后在另一个终端运行：

```bash
npm run perf:measure -- --base-url http://localhost:3000
```

正式回归守卫建议使用构建后的 preview，避免开发模式 chunk 让 `/archive` 等客户端页面产生不稳定的 JS 请求数：

```bash
npm run build
npm run preview -- --port 3001
npm run perf:guard -- --base-url http://localhost:3001
npm run visual:guard -- --base-url http://localhost:3001
```

脚本会测量以下场景：

- `/` 首页首屏
- `/` 首页内容区 reveal 后
- `/updates`
- `/about`
- `/archive`
- `/games`
- `/games/search`
- `/sites/1`

如果开发环境的站点详情 ID 不是 `1`，可以覆盖：

```bash
$env:PERF_SITE_PATH="/sites/123"
npm run perf:measure -- --base-url http://localhost:3000
```

## 性能守卫

运行，建议指向构建后的 preview 地址：

```bash
npm run perf:guard -- --base-url http://localhost:3001
```

守卫会读取 `docs/performance/budget.json`，检查：

- DOM 节点数
- JS 请求数
- JS 传输体积
- 图片请求数
- Long Task 数量和总耗时
- JS Heap
- 首页首屏是否误加载 `md-editor-v3`、`echarts`、`hls.js`
- `/archive` 空页是否误加载 `md-editor-v3`
- `/games` 首屏是否误加载 `hls.js`
- `/games/search` 是否误加载 `hls.js`

动态详情页依赖本地后端数据，如果失败会进入 warning；核心页面失败才会让守卫退出非 0。

## 视觉截图守卫

运行，建议与 `perf:guard` 使用同一个 preview 地址：

```bash
npm run visual:guard -- --base-url http://localhost:3001
```

脚本会对以下组合截图并生成 manifest：

- `/games` 中文亮色、中文暗色，桌面 `1440x900` 与移动 `390x844`
- `/` 首页导航中文亮色、中文暗色，桌面 `1440x900` 与移动 `390x844`
- `/en` 首页导航英文亮色、英文暗色，桌面 `1440x900` 与移动 `390x844`
- `/games/search` 中文亮色、中文暗色，桌面 `1440x900` 与移动 `390x844`
- `/en/games/search` 英文亮色、英文暗色，桌面 `1440x900` 与移动 `390x844`
- `/about`、`/en/about` 亮色、暗色，桌面 `1440x900` 与移动 `390x844`
- `/updates`、`/en/updates` 亮色、暗色，桌面 `1440x900` 与移动 `390x844`
- `/terms`、`/privacy`、`/en/terms`、`/en/privacy` 亮色、暗色，移动 `390x844`

视觉守卫会做这些硬检查：

- 截图主题必须与 `html.dark` 状态一致。
- 页面关键容器必须存在，例如 `.nav-home-page`、`.nav-header`、`.games-page`、`.games-search-page`、`.search-result-grid`、`.gf-pagination`、`.gf-static-page`、`.updates-summary-shell`。
- 不允许出现旧暗色入口：`games-page--dark`、`search-results--dark`、`is-dark-theme`、`spotlight-panels--dark`。
- 桌面和移动端不能出现明显横向溢出。

本地后端没有游戏数据时，`.game-card` 或 `.search-page-card` 为空只会记录 warning，不会阻断截图产出。

## 手工截图复核

`visual:guard` 负责生成基线截图和硬断言，最终发布前仍建议快速目视：

- `/`、`/en`：首页导航 header、搜索框、内容 reveal、站点卡片和 spotlight 在亮暗色下层级一致。
- `/games`：亮色与暗色背景、信息面板、侧栏搜索、游戏卡片 hover 状态一致。
- `/games/search`：筛选面板、结果卡片、分页和日期选择器在亮暗色下层级清晰。
- `/en/games/search`：英文文案不挤压卡片、按钮、分页和筛选 chip。
- `/about`、`/updates`：静态面板、summary、状态文案和时间线在亮暗色下层级稳定。
- `/terms`、`/privacy`：移动端长文案不造成横向溢出。
- 移动端：无横向滚动，卡片为单列或自然堆叠，固定/浮动按钮不遮挡主要内容。

## 刷新基线

当你确认当前性能表现是可接受版本后，运行：

```bash
npm run perf:baseline -- --base-url http://localhost:3000
```

这会刷新 `docs/performance/baseline.json`。只有在明确接受性能变化时才应该刷新基线。

## Trace

如果任务管理器里仍然看到 GPU 占用异常，可以生成 Chrome trace：

```bash
npm run perf:trace -- --base-url http://localhost:3000 --path /archive --wait 8000
```

首页需要 reveal 后观察时：

```bash
npm run perf:trace -- --base-url http://localhost:3000 --path / --reveal --wait 8000
```

Trace 文件会输出到 `docs/performance/reports/`，可以用 Playwright Trace Viewer 或 Chrome DevTools 辅助分析 paint、composite 和 long task。

## 报告位置

测量报告、截图、manifest 默认输出到：

```text
docs/performance/reports/
```

这个目录下的临时报告默认被 `.gitignore` 忽略，只保留 `.gitkeep`。如果某次报告需要长期归档，可以手动复制到一个明确命名的文档文件中。

## 限制

- Windows 任务管理器的 GPU 百分比不是稳定自动化指标，所以脚本不对 GPU 占用做硬阈值。
- JS 体积优先使用浏览器 PerformanceResourceTiming，缺失时使用响应头或脚本文本大小兜底。
- 本地后端接口、缓存状态、浏览器版本都会影响测量值，预算应保持保守，不要用过窄阈值追求“漂亮数字”。
