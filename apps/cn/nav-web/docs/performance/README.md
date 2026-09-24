# GoFurry Nav Web 性能测量与回归守卫

这套脚本用于样式系统维护回归。目标是让前端性能与样式调整有可重复的本地验证方式，防止首页、游戏页和游戏搜索页重新误加载重型依赖，视觉验证独立使用正式 Playwright Browser / Visual gates。

## 使用方式

快速测量可以直接启动本地前端：

```bash
npm run dev
```

然后在另一个终端运行：

```bash
npm run perf:measure -- --base-url http://localhost:3000
```

正式回归守卫建议使用构建后的 preview，避免开发模式 chunk 让客户端页面产生不稳定的 JS 请求数：

```bash
npm run build
npm run preview -- --port 3001
npm run perf:guard -- --base-url http://localhost:3001
```

脚本会测量以下场景：

- `/` 首页首屏
- `/` 首页内容区 reveal 后
- `/updates`
- `/about`
- `/games`
- `/games/search`
- `/site/1`

如果开发环境的站点详情 ID 不是 `1`，可以覆盖：

```bash
$env:PERF_SITE_PATH="/site/123"
npm run perf:measure -- --base-url http://localhost:3000
```

## 性能守卫

当前预算版本为 `v2.2.5`。预算保持保守，只用于阻断明显回退；接受性能变化时再有意识地刷新 baseline 或调整预算。

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
- 首页首屏是否误加载 `echarts`、`hls.js`
- `/games` 首屏是否误加载 `hls.js`
- `/games/search` 是否误加载 `hls.js`

`md-editor-v3` 已随 #72 删除，当前守卫不再检测它；`baseline.json` 中的同名字段是历史测量记录，保留原样。

动态详情页依赖本地后端数据，如果失败会进入 warning；核心页面失败才会让守卫退出非 0。

## 视觉与 Runtime 验证

旧 visual report runner 已在 #124 P7.2 退役。源码架构检查由 style-policy
负责；页面主题、容器、overflow 和行为由正式 Browser contracts 负责；已接受
像素由 pinned Playwright Visual 比对。不要再调用旧命令或创建平行截图 runner。

参阅[当前测试指南](../../../../../docs/frontend/testing.md)和
[最终验收记录](../../../../../docs/acceptance/issue-124-frontend-engineering-closure.md)。
Visual 不是性能预算；性能 measure/guard/trace/baseline 工具和当前预算保持独立。

## 刷新基线

当你确认当前性能表现是可接受版本后，运行：

```bash
npm run perf:baseline -- --base-url http://localhost:3000
```

这会刷新 `docs/performance/baseline.json`。只有在明确接受性能变化时才应该刷新基线。

## Trace

如果任务管理器里仍然看到 GPU 占用异常，可以生成 Chrome trace：

```bash
npm run perf:trace -- --base-url http://localhost:3000 --path /games --wait 8000
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
