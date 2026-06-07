# GoFurry Nav Web 性能测量与回归守卫

这套脚本用于 `v2.1.0-rc.1 - Measurement And Regression Guard`。目标是让前端性能优化有可重复的本地验证方式，并防止首页、知识库页和游戏页重新误加载重型依赖。

## 使用方式

先启动本地前端：

```bash
npm run dev
```

然后在另一个终端运行：

```bash
npm run perf:measure -- --base-url http://localhost:3000
```

脚本会测量以下场景：

- `/` 首页首屏
- `/` 首页内容区 reveal 后
- `/updates`
- `/about`
- `/archive`
- `/games`
- `/sites/1`

如果开发环境的站点详情 ID 不是 `1`，可以覆盖：

```bash
$env:PERF_SITE_PATH="/sites/123"
npm run perf:measure -- --base-url http://localhost:3000
```

## 回归守卫

运行：

```bash
npm run perf:guard -- --base-url http://localhost:3000
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

动态详情页依赖本地后端数据，如果失败会进入 warning；核心页面失败才会让守卫退出非 0。

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

测量报告默认输出到：

```text
docs/performance/reports/
```

这个目录下的临时报告默认被 `.gitignore` 忽略，只保留 `.gitkeep`。如果某次报告需要长期归档，可以手动复制到一个明确命名的文档文件中。

## 限制

- Windows 任务管理器的 GPU 百分比不是稳定自动化指标，所以脚本不对 GPU 占用做硬阈值。
- JS 体积优先使用浏览器 PerformanceResourceTiming，缺失时使用响应头或脚本文本大小兜底。
- 本地后端接口、缓存状态、浏览器版本都会影响测量值，预算应保持保守，不要用过窄阈值追求“漂亮数字”。
