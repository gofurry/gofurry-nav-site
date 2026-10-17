# #139-C1 — Games Surface Motion

日期：2026-10-18。开始时工作区干净，本地 `dev` 与只读远端查询均为
`ea7c6d2b7e8dcaea7e0ba026e013aacb85c87656`，与冻结文档一致。
沿用 [Motion Contract](../frontend/motion.md) 和 A/B 验收成果。

## 实际范围与基线

生产修改只有 `pages/games.less` 和 `GameSidebarSimilar.vue`。Home 的
GameInfoGroup、GameStatsPanels、GameUpdateNews 脚本、数据和结构保持原样。

| Owner | Normal 基线与最终契约 | Reduced |
| --- | --- | --- |
| Home 分区卡片 | opacity 360ms / transform 420ms，next/prev、8 卡、wrap、固定 spacer | 进入即时可读；Vue 暂存的离开页隐藏且不能点击/聚焦，随后由原 Transition 清理 |
| Home Stats | transform 360ms，类型切换、组数、价格/在线人数保持 | 只取消过渡，保留 inline translate3d 与 activeGroup 对应位置 |
| Home News | transform 340ms，索引/进度/限位/测量保持 | 只取消轨道过渡，保留 inline px offset；进度条原有 220ms 保持 |
| Detail Similar | animation 520ms / Timer 560ms、4 条与占位、正反方向保持 | 以最新 currentPage 立即提交唯一页，清理 Timer/RAF 并使尚未登记 RAF 的 nextTick 失效 |

Similar 增加局部 revision 和收敛函数，复用 `useReducedMotion()`。
运行中切 Reduced、重新选页、无效页数和卸载都会取消旧工作；恢复 Normal
不重播，下一次点击仍按原时序执行。页数变化使 source 或目标无效时收敛到
有效页；同长度刷新保留选择；空数据、失败、Loading/Retry 沿用真实分支。
CSS 兜底不是状态完成的依据，没有共享 Store、轮播框架或新依赖。

现有 Nuxt Motion owner 增加 Similar 生命周期断言；Browser 补在 Home 与
Detail Content owner。唯一额外 fixture 修改是 `games-home.ts` 的 News 外观
helper：Reduced 明确断言 none/0s，Normal 仍断言 transform/340ms/原曲线，
避免原 helper 把新无障碍契约误判为失败。网络与错误检查未放宽。

## 定向验证

工作目录 `apps/cn/nav-web`，日志/截图均在 Git 忽略的 `.cache/139-c1-review/`。
修改前使用与基线源码匹配的既有 Nitro 构建；最终 Browser 使用本次最终源码构建。

| 检查 | 结果 |
| --- | --- |
| Normal 基线 Browser（Home + Similar） | 首轮 FAIL 1/2：新增测试误选只有一组的在线人数；改用真实价格组后 Home PASS，两个 owner 均有基线通过证据 |
| 受影响 Vue/TS ESLint、Less/Vue Stylelint | PASS |
| `pnpm run style:policy`（一次） | PASS；新增债务 0，ambient raw 75 不变 |
| `pnpm run typecheck` | PASS |
| Nuxt Motion owner | PASS，最终 16/16；增加页数变更发生在 Timer/nextTick 空窗的验证后必要复测 |
| 最终源码 `pnpm run build`（一次） | PASS；既有 Tailwind sourcemap / Rollup annotation / Node exports 提示保留 |
| Home / Detail Content 两个 Browser owner | PASS，29/29；首次最终构建验证无失败 |
| 原 Home Closure 的 News / 分区布局定向用例 | PASS，6/6；合计 35 个场景，未跑全套 Browser |
| 全仓库 Go/DB、全 Browser/Visual、Golden 更新、云/生产验收 | NOT RUN，超出 C1 范围 |
| 远端 CI、维护者人工视觉签核 | NOT RUN；本地 Fixture 与截图不能替代 |

```sh
pnpm run test:browser tests/browser/regression/games-home.spec.ts tests/browser/regression/game-detail-content.spec.ts --workers=1 -g 'C1 Home Normal|similar games preserve'
pnpm run test:browser tests/browser/regression/games-home.spec.ts --workers=1 -g 'C1 Home Normal'
pnpm exec eslint app/components/game/detail/GameSidebarSimilar.vue tests/nuxt/motion-pagination.nuxt.test.ts tests/browser/regression/games-home.spec.ts tests/browser/regression/game-detail-content.spec.ts tests/browser/fixtures/games-home.ts
pnpm exec stylelint app/assets/styles/pages/games.less app/components/game/detail/GameSidebarSimilar.vue
pnpm run style:policy
pnpm run typecheck
pnpm run test:nuxt tests/nuxt/motion-pagination.nuxt.test.ts
pnpm run build
pnpm run test:browser tests/browser/regression/games-home.spec.ts tests/browser/regression/game-detail-content.spec.ts --workers=1
pnpm run test:browser tests/browser/regression/games-home-closure.spec.ts --workers=1 -g 'News SSR|Group card/rating/spacer'
```

Nuxt 用真实组件、Fake Timers 验证无需推进时间的即时完成、559/560ms 边界、
快速改选、Timer→nextTick→RAF 空窗、已登记 RAF、页数缩减、同长度刷新和卸载。
Browser 从真实 transitionrun / animationstart 触发同页偏好切换；不暂停动画、
不修改内部 Vue 状态，不增加固定 sleep、retry 或错误豁免。

Home 中断的事件内证据为 `connected:true / visibility:hidden / focused:false /
hit:false`，随后单页 DOM、恢复后的反向动画通过。Stats/News 初始 Reduced、
滑动中切换及恢复后的实际矩阵均与 inline 目标一致。Similar 在真实动画开始时
切偏好、切移动 Tab 或点击推荐路由，旧组件工作不会污染重新打开的内容。
原 SSR、网络预算、错误/Retry、Review、新闻外链、Resize 与中英文断言保留。
390/1440 Light/Dark 新用例与既有 390–1440 多断点 8/8/1/wrap 布局用例通过。

最终日志：`nuxt-final.log`、`browser.log`、`browser-closure.log`、`build.log`、
`eslint.log` / `eslint-final.log`、`stylelint.log`、`policy.log`、`typecheck-final.log`。
补充最后一个数据竞态测试后，仅对受影响测试再做 ESLint / Typecheck，均通过。
`GOFURRY_C1_REVIEW=1` 导出的四张图为 `home-normal.png`、
`news-reduced-mobile.png`、`similar-normal-desktop.png`、`similar-reduced-mobile.png`；
已由 Agent 查看单页内容、末页占位、轨道裁切与主题，未发现静态截图异常。
截图、`home-leave-state.json` 与基线日志均只保留上述忽略目录；无 Golden 更新。

## 风险与人工待审

本地工程验证不等于生产数据验收或维护者动态视觉批准。待维护者确认：

1. Home 卡片 Normal 正反方向、快速 wrap 与 Reduced 即时切页是否自然、无残影。
2. Stats/News 切组及偏好切换是否停在正确位置，390px 的新闻限位和进度是否协调。
3. Similar 在桌面/移动端的四槽高度、末页留空与快速改选是否保持原视觉节奏。

Showcase、Search、Collections、Nav、Insights、媒体、API/SSR/SEO、主题、Token、
依赖、CI、Golden 未改。未部署、未创建 PR、未合并 main、未关闭 #139；C2/C3 等待另行安排。
