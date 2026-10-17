# #139-C2 — Nav & Insights Selective Refinement

日期：2026-10-18。开始时工作区干净，本地 `dev`、只读远端 `dev` 均为
`1a2081131975657f551a1f4d4eb7d5d6603a02a0`，与冻结基线一致。
沿用 [Motion Contract](../frontend/motion.md)，没有新增规范、Token 或依赖。

## 实际范围

生产修改仅为 `InsightComparePicker.vue` 模板与 `insights/compare.less`：

- Popup 使用 Vue Transition：180ms opacity / -4px 进入，退出同步 `done()`。
  原 `v-if="open"`、listbox/combobox、Loading/结果容器、搜索请求及键盘脚本不变。
- Selected 使用 `TransitionGroup tag="ul"`，保留原类名、li key、ARIA 和删除按钮；
  只有新增项使用已有 Fast 160ms opacity。没有 appear、move 类规则、transform
  或矩阵列动画。SSR 预选直接可见，删除由同步 leave hook 完成真实 DOM 移除。
- Reduced 局部取消过渡并覆盖 enter-from 到可见终态；运行中切换立即到位，
  恢复不重播。没有新增 JS 媒体订阅、Timer、RAF、Observer、Store 或完成状态机。

测试只扩展既有 `insights-compare.spec.ts`，没有修改 fixture、原网络/SSR/URL/键盘
断言、其他生产 owner、依赖、CI 或 Visual Golden。

## Hero 只读审计

- SSR 的初始真实 picture/img 与 displayedId=0 保留到 hydration；Local Blob 在 mounted 后读取，不伪造 SSR Blob。
- 新帧先以 visibility:hidden 留在真实 DOM；Frame 只在 complete/naturalWidth 合格后发 ready，promote 保留同一新元素、移除旧帧并释放旧 Blob。
- 请求代次/disposed 拒绝迟到结果，重复切换丢弃 pending；各帧携带资源快照，主/镜像失败和 768px source 选择仍由真实 renderer 负责。
- nav.less 的 opacity/transform **700ms 是声明**；上述交接没有修改这两个属性，也没有 crossfade Transition。visibility 资格切换与旧节点移除不能称为 700ms 渐变。
- 已阅读 hero-handoff / hero-lifecycle / hero-local-background 既有门禁、SSR 保留与同次绘制字节断言；这些用例不等于 opacity/transform 中间帧测量，C2 没有新增 Computed Style/动画事件采样或重跑 Hero suite。
- 未发现本批可确认且获授权修复的 Hero 缺陷，结论为**保留、不修改**；资源交接、ManagedAsset、Blob、首屏揭示及既有 Footer hydration 债务边界不变。

## 验证

命令工作目录 `apps/cn/nav-web`。基线使用与 C1 源码匹配的现有 Nitro，最终
Browser 使用本次最终源码构建。日志和审查图只保存在忽略目录 `.cache/139-c2-review/`。

| 检查 | 结果 |
| --- | --- |
| 修改前 Popup/Selected Computed Style、Loading 同一容器、即时增删与 SSR/键盘基线 | PASS，4/4；两者均为 0s，SSR 顺序 2,1，Site/Game 原键盘用例通过 |
| 受影响文件 ESLint、Stylelint | PASS |
| `pnpm run style:policy`（一次） | PASS；新增债务 0，ambient raw 75 不变 |
| `pnpm run typecheck` | PASS |
| 最终源码 `pnpm run build`（一次） | PASS；工具链提示保留在 build.log，未修改依赖/构建配置 |
| 定向 Compare Browser | PASS，8/8；3 个新 Motion + 原键盘/竞态 + 2 个 SSR/响应式用例，无失败或重试 |
| Hero 当前运行采样、全量 Browser/Visual、Go/DB、云/生产验收 | NOT RUN，保留只读审计与本批定向范围 |
| 远端 CI / 维护者人工视觉批准 | NOT RUN；本地 Fixture 与截图不能替代 |

```sh
pnpm run test:browser tests/browser/regression/insights-compare.spec.ts -g 'C2 Motion Normal|Compare keyboard picker|Compare SSR selection validation, identity and noindex site zh' --workers=1
pnpm exec eslint app/components/insights/compare/InsightComparePicker.vue tests/browser/regression/insights-compare.spec.ts
pnpm exec stylelint app/assets/styles/pages/insights/compare.less app/components/insights/compare/InsightComparePicker.vue
pnpm run style:policy
pnpm run typecheck
pnpm run build
pnpm run test:browser tests/browser/regression/insights-compare.spec.ts -g 'C2 Motion|Compare keyboard picker|Compare held stale search|Compare SSR selection validation, identity and noindex site zh|Compare hydration and matrix game en 1024 2,1,4,3' --workers=1
```

三个新用例使用真实 DOM 的进入类/transitionrun、同页 emulateMedia 和请求 gate，
没有暂停动画、固定 sleep、retry 或修改 Vue 内部状态。Normal 与 Reduced 的进入态
和删除后真实节点连接状态分别验证，Loading→Ready 保持原节点且只入场一次。

实测进入态：Popup opacity=0 / translateY=-4px / 180ms；新增选项 opacity=0 /
transform=none / 160ms。初始 Reduced 为 opacity=1 / transform=none / 0s；
运行中切换后为最终可见态，恢复时实际动画数为 0，后续操作可正常进入。
旧 Popup/被删 li 的 isConnected=false；入场中删除、重复加回、Escape 和 SPA
About 往返后没有旧节点、错误焦点或重复项。SSR 0/2/4 项、最大四项、1024 矩阵
sticky/顺序及原 stale-search Abort、locale/region/URL/网络断言通过；无 Vue 警告。

日志为 `baseline.log`、`eslint.log`、`stylelint.log`、`policy.log`、`typecheck.log`、
`build.log`、`browser.log`；`normal-entries.json` 保留原生进入类的 Computed Style。
设置 `GOFURRY_C2_REVIEW=1` 导出 `compare-normal-1440-light.png` 与
`compare-reduced-390-dark.png`。Agent 已查看两张图，Popup 未遮住输入、保留原尺寸
与矩阵横向滚动边界，未发现静态截图异常；这些图片仍不等于维护者视觉批准。

## 人工待审与停止边界

维护者仍需确认 1440 Light / 390 Dark 的 Popup 180ms 轻位移是否自然，
以及新增项 160ms 淡入、快速删除和 Reduced 即时反馈是否清楚且无残影。
截图仅用于辅助审查，不代表动态视觉批准或生产数据验收。

不修改 #108 与 A/B/C1 已验收页面；未开 PR、合并 main、部署、关闭 #139 或进入 C3。
