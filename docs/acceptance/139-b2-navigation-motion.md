# #139-B2 — Navigation Consistency & Motion Decision

日期：2026-10-18。开始时本地 `dev`、`origin/dev` 和只读远端查询均为
`5e46469ffa49d11df5fd80bd7506130363adcd75`，工作区干净，与冻结 V1 基线一致。
沿用 [Motion Contract](../frontend/motion.md) 与 [B1 验收](139-b1-core-interaction-motion.md)。

## 实际修改与所有者

生产修改仅在 `components/nav.less` 和 `pages/nav.less`，没有修改组件脚本。

| Owner | Normal | Reduced / 保留边界 |
| --- | --- | --- |
| Mobile Menu | 原 opacity / -8px，200ms | 即时正确可见态；退场 visibility:hidden + pointer-events:none，即刻退出焦点及命中树 |
| Bottom Tabs | 仅 item 背景/边框 420→160ms；浮条显隐仍 220ms | 取消过渡，保留 scoped translate(-50%, …)、缩放终值、安全区、72px 门槛与 640px 断点 |
| Quick Modal | 原遮罩 180ms、面板 transform 220ms / opacity 180ms、8px / .985 缩放不变 | 局部即时显示；关闭退出焦点及命中树，原 Vue Transition 完成真实 DOM 清理 |
| Site Popover | opacity / 有方向的 ±6px，220ms；移除 filter 动画、scale、常驻 will-change | 即时可见/隐藏；保留 left/top 即时定位、静态 backdrop blur(12px)、透明材质、阴影、数据与点击 owner |

运行中切换时，Vue 可能短暂保留原生退场节点；测试同时验证它不能捕获点击或焦点、
真实重开只有一个有效根节点、关闭最终确实移除，未用 opacity 掩盖交互残留。
事件内证据明确记录退场根节点 `connected:true`、`visibility:hidden`、
`focused:false`、`hit:false`，之后真实重开单根、最终关闭零根。
无需给 NavBar/NavHeader 增加监听器、计时器或 B1 完成追踪机制。
测试最终从真实 transitionrun 事件触发同页 emulateMedia，不暂停动画或改 Vue 状态。

Site Popover 上游 `NavSiteGrid` 的 520ms hover 宽限与 320ms 清理寿命保持不变；
本批精修的是其 visible 状态改变后的视觉过渡。Group Popover、A/B1、业务路由、
SSR、主题、链接、数据、设置保存、依赖、CI、Token 和 Visual Golden 均未修改。

## transition: all 与技术决策

只读 Computed Style 已对照 390px Light/Dark、1440px Dark 和真实英文路由：
`.gf-nav` 变化包括背景、底边、阴影及响应式尺寸；`__inner` 有尺寸/padding，
`__center` / `__links` 有尺寸、max-width、语言文本宽度，`__actions` 有尺寸/margin。
这些有限采样不足以证明删去所有布局属性后中间帧完全等价，因此五处 all 均保留，
没有为统一时长改动既有 300/200ms 或 Mobile Panel 的 500ms 颜色过渡。
证据见缓存 `nav-owner-baseline.json` / `nav-owner-final.json`。

Vue Transition + Less 已覆盖 B 阶段的进入/退出、局部方向反馈与实时 Reduced。
不安装 motion-v、GSAP 或新依赖，不做无收益 PoC，不复制通用 Motion 系统。
将来只有确有复杂布局、共享元素或手势收益的独立需求时，才另议技术评估。

## 定向验证

命令工作目录：`apps/cn/nav-web`；所有日志/审查图保留在忽略目录
`.cache/139-b2-review/`。本地受控上游与截图不代表生产或维护者视觉签核。

| 检查 | 结果 |
| --- | --- |
| 复用与 B1 源码匹配的 Nitro 构建，3 个 Normal 基线 | PASS；核实 200、180/220、300、420ms 及旧 filter/will-change |
| 临时 CSS 注入方案探针 | 首轮 FAIL 1/2：遮罩测试点落在既有高层 NavBar；改为真实遮罩区域后相关用例 PASS，未改产品层级；探针已移除 |
| 2 个 Less 的 Stylelint、3 个 Browser TS 的 ESLint | PASS |
| 唯一一次 `pnpm run style:policy` | PASS；新增债务 0，原 ambient raw 75 不变 |
| `pnpm run typecheck` | PASS（本批修改了 Browser TS） |
| 最终源码唯一一次 `pnpm run build` | PASS；仅既有 Tailwind sourcemap / Rollup annotation / Node exports 提示 |
| 三个受影响 Browser owner，workers=1 | 首轮 FAIL，25/26；新增 Popover 断言轮询时已超过原 320ms 节点寿命 |
| 事件内采样并真实 hover 重开后，3 个中断用例定向复测 | PASS，3/3；只修测试观察窗口，没有生产改动、重建或放宽断言 |
| 最终覆盖结果 | 26 个场景均有通过证据；既有请求/键盘/主题/SSR/错误断言保留，补充真实拖拽排序持久化 |
| 全仓库 Go/DB、全部 Browser/Visual、Golden 更新、生产/云验收 | NOT RUN，超出本批范围 |
| Preferences 专属 owner | NOT RUN；未触及其实现或返回焦点，Nav Shell 原有入口回归保留 |
| 远端 CI、维护者视觉签核 | NOT RUN，本记录为提交前本地证据 |

```sh
pnpm exec stylelint app/assets/styles/components/nav.less app/assets/styles/pages/nav.less
pnpm exec eslint tests/browser/regression/nav-shell.spec.ts tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/nav-revealed-content.spec.ts
pnpm run style:policy
pnpm run typecheck
pnpm run build
pnpm run test:browser tests/browser/regression/nav-shell.spec.ts tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/nav-revealed-content.spec.ts --workers=1
pnpm run test:browser tests/browser/regression/nav-shell.spec.ts tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/nav-revealed-content.spec.ts -g 'B2.*(live Reduced|directional motion)' --workers=1
```

日志为 `baseline.log`、`css-probe*.log`、`stylelint.log`、`eslint*.log`、`policy.log`、
`typecheck.log`、`build.log`、`browser.log`、`browser-recheck.log`。
审查图为 `mobile-menu-dark.png`、`bottom-tabs.png`、`quick-modal.png`、
`site-popover-top.png`；退场证据为 `menu-leave-state.json` / `quick-leave-state.json`。

人工待确认：移动菜单 Normal 开合与 Reduced 即时感；Bottom Tabs 高亮速度及水平居中；
Quick Modal 快速关开是否残影；Site Popover 上下方向是否自然、材质是否保持。
B2 仍需维护者人工签核，不据此关闭 #139 或启动 C。
