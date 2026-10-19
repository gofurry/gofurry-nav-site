# #145-B — Hero Scroll & Content Readiness

2026-10-19；现有 `dev`，起点 `1569cad`，开始时工作区干净且与 `origin/dev` 一致。
**工程定向验收 PASS；人工体验/视觉 PENDING。** 未进入 #145-C。

## 实际改动

- Home 直接 SSR 输出 TransitionBar、精选、分组和站点卡片；内层改为 `div`，保留全站唯一 `main`。Footer 从首次 SSR 起位于正常流中。
- 删除 Reveal 强制滚动、wheel/touch/key 拦截、700ms/idle 预热、全局事件及 Search Reveal Lock；中文/英文提示改为“下滑探索更多 / Scroll to explore”。Search 其它业务不改。
- 桌面 ToolDock 沿用 ≥1024px；正文顶部到达视窗中线时启用，越过顶部继续保留，返回 Hero 隐藏。只用 passive scroll + 单 RAF / resize，卸载清理；目录仍在打开搜索时请求。
- 保留单次 Home 数据通道、SFW SSR / mounted 后的本地模式、图片 lazy/async/low 和 content-visibility。天气 iframe 保留原 URL，改用原生 lazy；手机隐藏时不请求，旋转到可见宽度后只请求一次。
- 移除仅供旧 Reveal 的 88px scroll margin（无其它生产锚点/焦点消费者）。测试的显式定位仍自行避让双岛，不伪称为产品滚动。
- `assertHeroHydration` 及消费者不再消费/排除任何 Footer mismatch；保留原始错误、SSR Hero 节点身份和严格网络账本。当前合同/指南加 B 的覆盖说明，历史验收文件保留。

A3 双岛、图标/材质、A2 类别/平台逻辑、MobileBottomTabBar、Hero 资源与偏好、Games/Insights/详情、后端、依赖、CI、Golden 均未修改。

## 定向验证

命令目录均为 `apps/cn/web`；Node **24.15.0**、项目 pnpm **12.6.0**；Chromium **148.0.7778.96**，Windows 本地 Nitro + 原有确定性 upstream。

| 状态 | 验证 |
| --- | --- |
| PASS | 本次 Vue/TS 文件定向 ESLint；测试调整后对五个受影响测试文件补跑 ESLint |
| PASS | `pnpm exec stylelint app/assets/styles/pages/nav.less app/components/nav/NavHomePage.vue app/components/nav/NavTransitionBar.vue app/components/nav/SearchBox.vue app/layouts/default.vue` |
| PASS | `pnpm run style:policy`：305 文件，样式债未增加 |
| PASS | `pnpm run typecheck` |
| PASS | `pnpm run test:nuxt tests/nuxt/search-platform-picker.nuxt.test.ts`：1/1，真实语言变更、平台交换/URL、media 订阅清理 |
| PASS | `pnpm run build`：本批一次构建，所有后续 Browser/性能使用此产物；仅已有 Tailwind sourcemap / Node exports 警告 |
| PASS | 67 个唯一定向 Browser 场景最终通过：首轮 61/67，修正测试采样/预期后定向补测 6/6；零 retry，不重复全套 |
| PASS | `git diff --check`，保护文件/Golden/依赖/CI 差异审计 |
| NOT RUN | 全量 Browser、Visual/Golden、远端 CI、真实生产/低端手机性能、全仓 Go/DB；本批不要求这些门禁 |
| PENDING | 维护者动态视觉及真实设备体验确认，截图不代替批准 |

实际 Browser 命令：

```powershell
$env:GOFURRY_145_B_REVIEW='1'
pnpm exec playwright test tests/browser/regression/nav-revealed-content.spec.ts tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/nav-shell.spec.ts tests/browser/smoke/nav-home-locales.spec.ts tests/browser/regression/hero-lifecycle.spec.ts tests/browser/regression/hero-handoff.spec.ts --workers=1 --grep '145-B|Localized Home|145-A Home|145-A locale|145-A desktop|145-A rotating|145-A2|Suggestions own|Reduced mobile|Search keeps|Suggestion requests|IME owns|Quick Access|B2 Quick Modal|Provider HTML|Core cards|ToolDock caches|Spotlight|Site Popover|displayed Hero|Save commits|Mobile-only'
pnpm exec playwright test tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/nav-revealed-content.spec.ts tests/browser/regression/nav-shell.spec.ts --workers=1 --grep 'actual SSR content.*partial|touch scroll|inline picker has natural|B2 Quick Modal preserves'
```

覆盖：实际 HTML 经 DOMParser 检查（不是 payload 文本）、两种语言/完整/部分/空数据、SFW SSR 后恢复 NSFW；中英文 light/dark × 390/640/767/768/1440 的 Footer/单 main/无溢出/零 hydration 错误；原 SSR Hero 保留；120px 首滚和连续 40px 滚动、PageDown/End、真实 CDP touch 输入、Normal→Reduce→Normal；ToolDock 深处持续、回 Hero 移除面板、路由离开清理、按需目录缓存；Search/IME/Abort、2×4 平台、Quick Sites、Spotlight/Popover、手机接力/旋转及原 history 回归。

首轮失败保留在 `.cache/145-b-review/first-browser/` 与 `browser.log`：

- 两个新 partial 用例误将既有 Spotlight 空分区算作不渲染；改为明确断言 4 个分区、3 个空态，未改产品。
- 合成手势未产生 touch 滚动；改为 CDP touchStart/move/cancel，并验证实际滚动距离，未修改阈值。
- 两个旋转用例的天气账本只看最终宽度；现显式记录旋转到 768px 后的一次原生 lazy 请求，恢复手机后仍保留这条证据。
- Quick Modal 用例有一次首页 upstream 未收到请求、返回 502；未进入 Modal 业务断言。保留原始网络/错误证据，增加仅失败时的 server 日志附件；独立定向验证通过。根因未复现，不能据此声称已修复环境问题；未新增重试或错误豁免。

`RUNTIME-FOOTER-01` 的代码原因是布局按 window 宽度/Reveal 选择 Footer。现在无该分支；上述两语言五宽度的严格检查通过。历史失败记录未改写。

## 真实性能对照与代价

同一 Home Fixture 数据/SVG、900px 高度、无 CPU/网络限速、每宽度各 3 次新浏览器 context。首屏及首次 120px wheel 后各观察 1800ms；该窗口仅用于性能采样，不作为测试 readiness sleep。未运行完整 perf suite、未重置基准或预算。

基线复用原有 Nitro（2026-10-18T19:20:43.628Z 构建），其 B 所涉及 Home/Reveal/SSR 代码与起点一致，但双岛仍为上一轮 32px 斜切参数；当前源为已接受的 18px 参数。它不是精确 `1569cad` 构建，不能拿来证明双岛性能或整体百分比收益；避免了额外旧版构建。最终构建与本批产品源码一致。

| 指标（旧 → 新） | 1440px | 390px |
| --- | --- | --- |
| SSR 未压缩字节 | 46,383 → 75,810 | 46,383 → 75,810 |
| SSR 总元素 / 正文后代元素 / 图片 | 211 / 0 / 3 → 534 / 259 / 37 | 同左 |
| SSR 分组 / 卡片 / 精选 / Footer | 0 / 0 / 0 / 0 → 2 / 6 / 4 / 1 | 同左 |
| 初始化后浏览器元素 / 图片 | 498 / 33 → 547 / 37 | 559 / 39 → 545 / 37 |
| 首屏长任务数量 / 总耗时 | 0 / 0ms → 2 / 116–120ms | 1 / 75–77ms → 2 / 118–128ms |
| 长任务超出 50ms 的累计部分¹ | 0 → 16–20ms | 25–27 → 18–28ms |
| 初始请求 / 首滚后累计请求 | 60 / 91 → 59 / 89 | 72 / 72 → 69 / 69 |
| 首滚长任务 | 0 / 0ms → 0 / 0ms | 0 / 0ms → 0 / 0ms |
| 120px wheel 的最终 scrollY | 812 → 120 | 120 → 120 |
| hydration mismatch | 0 → 0 | 1 → 0 |

¹ 是本地观察窗内 `sum(max(duration-50,0))`，不是 Lighthouse TBT 分数。新版最长单任务 68ms。所有样本仅一次 Home upstream、无初始化 ping/重复 saying/directory，初始图片网络数仍为桌面 2 / 手机 14。SSR 解析/水合有可复现的初始化成本，不能称整体性能提升；本 Fixture 未观察到首次滚动长任务或超过 100ms 的单任务。真实网络/大量站点/低端手机仍需人工体验确认，不通过隐藏正文规避成本。

原始数据/命令脚本：根 `.cache/145-b-review/before.json`、`after.json`；本地采样命令 `node .cache/145-b-review/measure.mjs before` / `after`，脚本位于前端的忽略目录。采样复用已有 Fixture 数据，实际运行生产 Nitro，无替代 Vue/Nuxt 运行时。

## 人工待验 / 结束边界

- 1440：轻滚可停在 Hero/正文交界；ToolDock 中线出现、深处保留、返回 Hero 隐藏；双岛外观未变。
- 390 与 768：正文滚动、首屏初始化体感、平台展开/键盘、底栏接管/反向滚动/回顶、旋转，无横向溢出。
- `/` 与 `/en`：刷新、主题/语言、浏览器前进后退，特别是 Home 历史滚动恢复；图片、工具搜索、Quick Sites、Footer；实时切换 Reduced。

审查截图：根 `.cache/145-b-review/desktop-hero-content-boundary.png`、`mobile-natural-scroll.png`；日志同目录。只作本地 Fixture 审查，不是生产或人工签核。

无产品功能阻断；保留上述初始化成本和未复现的本地 502 风险。仅提交推回当前 `dev`，不合并、不部署、不关闭 #145，不开展 C。
