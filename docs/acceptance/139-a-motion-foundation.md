# #139-A — Motion Foundation & Accessibility

日期：2026-10-17。实施前本地 `dev`、`origin/dev` 与只读查询的远端 `dev`
均为 `c9bb0b8affa0f3ad9098b45e419254fd957d791f`，与冻结 V1 文档一致；工作区干净。
在原分支实施，没有 pull/reset/rebase/stash、分支或 Worktree 操作。
持久规范只在 [Motion Contract](../frontend/motion.md)，前端文档导航增加一处链接。

## 所有者审计与 Normal 基线

以下时间来自修改前生产 Nitro 构建的真实组件 Computed Style 与源码，
5 个定向基线 Browser 用例全部通过；受控本地上游不代表生产数据验收。

| Owner | 已核实基线 / 风险 | A 决定 |
| --- | --- | --- |
| `tokens.less` / `mixins.less` | fast 160ms、base 220ms；交互 mixin 已有 Reduced | 保留，无第二套 Token |
| `NavHomePage.vue` | Reveal 原顺序：mount → nextTick → 两帧 → reveal/event → smooth；实测到达内容位置 | 只在滚动调用时读取当前偏好，Reduced 用 instant，保留锁、预热和移动端自动揭示 |
| `NavSpotlightPanels.vue` / `nav.less` | next/prev 520ms；560ms 收口；CSS 停止不能提交目标页，nextTick 后登记 RAF 有清理空窗 | 直接提交最后选中页；取消 Timer/RAF，revision 使迟到的 nextTick 失效；Normal 方向/wrap/时序保留 |
| `GameSearchResult.vue` / `games-search.less` | next/prev 420ms；380+80ms 收口；视觉目标与网络 pending 不可混淆，页码可先于 loading 更新 | Reduced 收敛为单轨；按已接受列表所属页避免旧卡片短暂挂在新页码下；请求未完成保留静态骨架、aria-busy/status；恢复 Normal 时当前 pending 仍静态，下一次分页恢复滑动 |
| `nav.less` | `navHeaderBounce` 1s、`navSearchSpin` 1s 持续；`navSearchFadeIn` 0.5s | 窄媒体查询停用，滚动提示、Spinner、Loading 和建议内容仍可见 |
| `games-search.less` | `search-card-skeleton` 1.35s 持续 | 保留背景/几何的静态骨架 |
| `games.less` | `games-video-loading-spin` 0.82s 持续 | 保留视频等待层和静态 Spinner，媒体解码/失败/卸载逻辑不变 |
| `useGameShowcaseAutoplay` | 已监听系统变化，Reduced 停自播，手动及显示资格独立 | 登记并保留，不迁移成熟控制器 |
| `game-collections.less`、`error.less`、`lottery.less`、`insights.less` | 已有局部 Reduced 规则 | 保留，不变更这些产品的视觉设计 |
| Preferences / `ModeSettingModal` / `BackgroundPreferencesEditor` | 现有局部 Reduced CSS、按偏好即时滚动 | 保留 |
| `FallingLeavesCanvas.vue` | 已监听并释放 Reduced 查询 | 保留原 ambient 及其债务 |
| `GameSidebarSimilar.vue` | 独立 560ms Timer/RAF 分页 owner | B/C 评估候选，本批未迁移或重新设计 |
| Nav 其它揭示/Popover、Site Detail、Insights、Modal | 各自 Transition、显示/焦点与主题 owner；不是本批通用改造对象 | 后续按真实消费者另行审计；保留现状 |

没有发现必须扩大冻结范围才能完成 A 的阻断。现有常驻 `will-change` 等未作
无收益整理。新增 `useReducedMotion` 仅服务 Spotlight/Search 两个真实消费者，
在 mounted 读取系统设置、卸载释放监听；它不持有分页、请求、全局 Store 或插件状态。

## 行为和证据

- Normal：修改前/后均验证 Reveal smooth、Spotlight next/prev/wrap、Search
  next/prev 和结果身份；CSS 时长/曲线、布局、图片、SFW、访问统计没有改动。
- Reduced：初始与同页运行时切换都收敛到唯一正确目标；键盘 Enter 后按钮焦点保留。
  Search 连续选第二/第三页时精确取消旧请求，旧结果释放后仍维持最新 Loading；
  恢复 Normal 不提前显示旧卡片，真实响应到达即显示新页。
- 生命周期：真实 SPA 路由离开中断 Spotlight/Search；媒体切换后无浏览器异常。
  Nuxt Fake Timer 用例不推进时间即验证即时完成，并专门中断 Timer 完成后、
  nextTick 尚未登记 RAF 的窗口，覆盖 reduce 和 unmount。
- 原有失败/空结果/重试、Filter Draft/Apply/Cancel、IME、SSR/Hydration、NSFW、
  视频真实解码与精确取消断言均保留。移动 Home 仍只使用原有严格 Footer 已知债务断言。

## 验证记录

命令在 `apps/cn/nav-web` 执行，日志保留于 Git 忽略的 `.cache/139-a-review/`。

| 检查 | 结果 |
| --- | --- |
| 修改前 `pnpm run build` + 5 个 Normal 基线 Browser 用例 | PASS |
| `pnpm run lint`，最终新增/修改测试与 helper 的定向 `pnpm exec eslint …` | PASS |
| `pnpm run stylelint` | PASS |
| `pnpm run style:policy` | PASS；appearance/arbitrary/important/deep/legacy-dark 均 0，ambient raw 75 不变 |
| `pnpm run typecheck` | PASS |
| 修改后 `pnpm run build` | PASS；保留原有 Tailwind sourcemap / Rollup annotation / Node exports 提示 |
| `pnpm run test:nuxt tests/nuxt/motion-pagination.nuxt.test.ts` | PASS，8/8；包括页码先行和页码/结果同批更新 |
| 下面五个 Browser owner，`--workers=1` | 首轮 FAIL：47/48；唯一失败为新增 Loading 截图误用 ready-only clip helper |
| 修复截图调用后，Header Normal/Reduced 与视频场景定向复测 | PASS，2/2；无产品改动、无放宽断言 |
| 补强 Spotlight 动画中 SPA 离开后的定向复测 | PASS，1/1 |
| 页码先行边界补齐后，重新构建并重跑 Search lifecycle/states 两个 owner | PASS，24/24；typecheck 和受影响文件 lint 再次通过 |
| 最终覆盖结果 | 48 个定向场景全部有通过证据；没有跳过测试或启用 retry |
| 全 Browser / Visual Golden / 全仓库 Go / DB / 云服务 / 生产验收 | NOT RUN，不属于 A 的定向范围 |
| 远端 CI / 维护者人工视觉签核 | NOT RUN（此记录为提交前本地证据），不能由 Fixture 或截图替代 |

Browser 命令：

```sh
pnpm run test:browser tests/browser/regression/nav-revealed-content.spec.ts tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/games-search-lifecycle.spec.ts tests/browser/regression/games-search-states.spec.ts tests/browser/regression/game-detail-interactions.spec.ts --workers=1
pnpm run test:browser tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/game-detail-interactions.spec.ts --workers=1 --grep 'Normal motion baseline|a real video'
pnpm run test:browser tests/browser/regression/nav-revealed-content.spec.ts --workers=1 --grep 'Reduced Reveal'
pnpm run test:browser tests/browser/regression/games-search-lifecycle.spec.ts tests/browser/regression/games-search-states.spec.ts --workers=1
```

截图用 `GOFURRY_MOTION_REVIEW=1` 在相同 Functional owner 中按需导出，无新 runner。
本地五张截图为 `nav-reveal-spotlight-dark.png`、`nav-static-loading.png`、
`search-static-1440-light.png`、`search-static-390-dark.png`、`video-static-loading.png`。
已检查截图中的单页内容、静态等待提示与焦点轮廓；这是 Agent 的静态截图检查，
不是人工动态交互签核或像素级 Normal 等价证明。截图及首轮失败 trace 只保留本地。

A 不授权 B/C、Motion for Vue、API/DB/依赖/CI/主题/CDN/Golden 修改，
不关闭 #139、不创建 PR、不合并 main、不部署。维护者决定后续阶段与视觉签核。
