# #145-A1 — Floating Shell & Mobile Search Refinement

2026-10-19；起点 `dev=origin/dev=fc2d422`，工作区干净。基于当前 `apps/cn/web` 实现；本记录补充 A，不改写 A 历史。

## 实际范围

- 两岛全站脱离文档流：桌面 fixed，手机 absolute 随页面滚出。删除两个 layout 的统一顶部留白，Hero/页面画布从顶部开始；仅在 Games、Insights、Site/Groups、Updates、Static/Legal、Error 的实际内容入口调整安全距离。底部安全距离和原 Bottom Tabs 几何阈值、持续接管、回顶、焦点交接不变。
- 内侧材料改成镜像浅切面与圆滑曲线，背景/边缘共用路径；CSS shape 不支持时回退普通圆角。内容、点击和焦点轮廓不裁切；岛内 active/hover 只改变图标颜色，无嵌套实心卡片。
- 手机类别是单行横滑原生按钮；平台网格折叠为当前平台入口。body Teleport 的原生 modal dialog 使用现有全局 Modal 颜色语义、局部 Vue Transition；即时退出、Reduced 初始/运行中收敛。保留八个平台及全部其他类别、关键词与原 URL。
- Sheet 打开取消 debounce/Abort 及旧建议，纳入原 Reveal Lock；选择只切平台；Escape/遮罩/关闭/路由/卸载/旋转释放对话框和原 body overflow。局部 Tab 首尾循环；短视口可滚动选项并保持焦点可见。移动建议列表给平台入口让位并限制在 Hero 内；桌面布局不变。
- Hero/ManagedAsset、BottomTab 业务、Preferences 保存逻辑、SSR/SEO/API、Game/Insights 领域逻辑、依赖、CI、Tokens 与 Golden 均未修改。

- 维护者在验收发现 300px 偏差后追加授权最小 Reveal 修复：首次 Reduced Reveal 接管滚轮时阻止同一事件随后叠加原生 delta；仍尊重 Reveal Lock，Normal 不阻止默认滚动。首次 Reveal 完成即移除 wheel listener，其余预热/挂载/SSR/手势生命周期不变；不启动 #145-B 重构。

## 验证证据

当前结论：本地定向工程验收通过，人工动态视觉待签核；下列失败记录均保留为已解决的迭代历史。

Windows / Node 24.15.0 / pnpm 12.6.0 / 现有 Chromium；retries=0、workers=1。正式 Browser 使用本地生产 Nitro 构建与现有确定性 upstream；不等于生产验收。

| 状态 | 结果 |
| --- | --- |
| PASS | 全部改动 Vue/TS 的定向 ESLint；4 个 Less 与 Error scoped style 的 Stylelint；`pnpm run style:policy`（无新增 debt）、`pnpm run typecheck`。 |
| PASS | `nav-home-header.spec.ts` + `nav-shell.spec.ts` 共 45/45；桌面原搜索/IME/Abort/URL/Quick Access/Preferences、九个宽度、底栏交接、双语、主题、Reduced 与 Sheet 生命周期。日志 `.cache/145-a1-browser-final.log`。 |
| PASS | 短视口修正后的生产构建上，跨页/Sheet 定向集 29 PASS；同次两个 Reveal 旧坐标失败已定位为遗漏 #145-A 的 88px scroll margin，保持 1px 精度修正 Fixture 后单独复核整个 owner（历史及最终补充修复见下）。日志 `.cache/145-a1-browser-cross-page.log`。 |
| FAIL（修复前） | 修正 Fixture 后 Revealed Content **6 PASS / 1 FAIL**：Reduced Reveal 在原被动 wheel / 强制滚动路径出现 300px 终点偏差。Normal、Popover、Core、ToolDock、分页、模式/响应式通过。当时未改冻结产品路径；随后取得最小修复授权。未放宽 1px 断言、未增加自动 retry。日志 `.cache/145-a1-browser-reveal.log`。 |
| PASS | 授权最小修复后的最终生产构建及 **15/15** 定向复核：Reduced/Normal Reveal、原生滚动取消边界、Search Lock、静态 Loading、Sheet、五个手机宽度交接及双语旋转；另行 ESLint、typecheck、style:policy 均通过。日志 `.cache/145-a1-browser-authorized-fix.log`、`145-a1-reveal-typecheck.log`。 |
| PASS | `git diff --check`；保护文件、依赖与 Golden 的差异检查。 |
| NOT RUN | 全量 Browser、Visual/Golden 更新、Go/DB、远端重型工作流、生产操作。 |
| PENDING | 维护者真机与动态外观验收。 |

执行命令（均从 `apps/cn/web`）：`pnpm exec playwright test tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/nav-shell.spec.ts --workers=1`；随后最终生产构建上执行下列定向集，以及 `pnpm exec playwright test tests/browser/regression/nav-revealed-content.spec.ts --workers=1`。

最终定向集的 owner：`nav-home-header`、`nav-shell`、`games-home`、`game-detail-content`、`insights-navigation`、`site-detail-shell`、`error-experience`、`preferences-foundation`、`nav-revealed-content`、`game-collections`、`updates`、`site-groups`、`lottery`、`games-search-contract`（均为 `tests/browser/regression/*.spec.ts`）。实际 grep：

```text
145-A1|Games Home SSR.*light|Games Home mobile|populated detail is rendered|Insights navigation|Insights mobile|145-A Insights|145-A same-page|responsive Target shell (390 light|1440 dark)|sticky tabs leave|hydrated keyboard|No-JS fallback|Preferences interrupts|Initial Reduced mobile Preferences|Reduced Reveal and Spotlight|Normal motion baseline|Index grid.*390px|Timeline geometry.*390px|Release (index|article) SSR and hydration zh 390 light|Site Groups SSR.*light|Lottery Mobile modal|Lottery activation success mobile shell light|English Search ready shell 390 light
```

迭代证据：开发模式 Fixture 首例在客户端挂载等待阶段超时，结束该次测试后改用生产构建；未放宽断言。首轮新增检查发现 Tab 边界进入浏览器工具栏、随后短视口反向 Tab 不滚入可见区，已局部修复；动效采样限定为 Sheet 的实际 transitionrun。保留 `.cache/145-a1-source-browser.log`、`145-a1-browser-initial.log`、`145-a1-short-focus.log`，最终结论以修复后的测试为准。另将 Revealed Content 的 reveal/align 两处几何断言同步到已存在的 scroll-margin-top（仍为 1px 精度，未修改滚动实现或 smooth/instant 断言）。

审查截图仅在忽略目录 `.cache/145-a1-review/`；包括 375/390/430 手机搜索/Sheet、桌面/紧凑双岛、Hero/底栏。未更改任何 Visual Golden。

最终补充验证命令（仅受该最小修复影响的 15 项）：

```sh
pnpm exec playwright test tests/browser/regression/nav-revealed-content.spec.ts tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/nav-shell.spec.ts --workers=1 --grep 'Reduced Reveal and Spotlight|Normal motion baseline|Search keeps typography|Reduced mobile Home|145-A1|Home handoff'
```

## 人工待验

1. 真机 375/390/430：软件键盘、横竖屏、长平台列表、关闭返回焦点、搜索建议与入口无遮挡，滚动位置恢复；Normal/Reduced 实时切换。
2. 768/900/1024/1440 双岛的短切面、颜色反馈及留白；中英/明暗主题。
3. 首页 Hero 顶部、手机底栏持续接管/回顶退出；Games/Insights/Site Detail/静态页首屏重要控件与 sticky 避让。

已解决的阻断：`nav-revealed-content.spec.ts:88` 在 `scene.reveal()` 的终点校验收到误差 **300**（期望 <=1），尚未进入该例后续 Spotlight 断言；可用 `pnpm exec playwright test tests/browser/regression/nav-revealed-content.spec.ts --workers=1 --grep "Reduced Reveal and Spotlight"` 定向复现。`NavHomePage.vue` 与基线无差异，偏差恰等于 Fixture 的 wheel delta；尚未通过基线独立运行证明其历史归属，因此不宣称是已确认的历史失败。维护者已明确授权这个最小补充；首次 Reduced Reveal 取消同一 wheel 默认 delta 后，原测试及新增 Normal/Lock 边界断言均通过，未放宽终点断言。

已知边界：旧 mobile Home Footer hydration 债由原严格 helper 检查，未新增豁免；不支持 CSS shape 的浏览器采用圆角回退。真机键盘/Safari 未在本轮 Chromium Fixture 中代验。#145-B 未启动。
