# #139-B1 — Core Interaction Motion Refinement

日期：2026-10-17 至 18。实施前本地 `dev`、`origin/dev` 和只读远端查询均为
`57e756f88aef5dd3ba5044e16b328fa39addac3c`，与冻结 V1 文档一致，工作区干净。
沿用 [Motion Contract](../frontend/motion.md)，不新增规范或依赖。

## 实际范围

- SearchBox / `pages/nav.less`：删除旧 500ms Transition、进入类及
  `navSearchFadeIn`；Vue Transition 单独负责 220ms opacity / -6px 入场，
  退出直接完成。Loading → 结果保留同一容器；未改请求、600ms 防抖、
  Abort、IME、Escape、键盘选择、Reveal Lock 或搜索 URL。
- ModeSettingModal / `components/preferences.less`：保留 Teleport + v-if 和
  三页 Editor，遮罩 160ms、面板进入 220ms / 退出 160ms、8px 位移，无缩放。
  关闭立即应用 inert / aria-hidden、禁用指针并移出焦点，业务保存不等待动画。
  局部完成 hook 观察 Vue 施加的 to 类，只等待遮罩/面板的实际 CSS Transition；
  无自建 RAF、逐帧动画或延迟卸载 Timer。Reduced 立即结束等待，取消或
  卸载使旧完成回调失效；恢复 Normal 只影响下一次交互。
- NavBar 仅提供真实焦点返回目标：桌面按钮或不会随移动菜单消失的菜单触发器。
  底部导航通过弹层捕获实际 opener，调用方未改。

通用 Modal、Token、导航菜单动效、Spotlight、Games、其它视觉设计、API/SSR/SEO、
配置保存顺序、依赖、CI 与 Golden 均未修改。没有修改 fixture 的网络预算或错误断言。

## 验证

工作目录 `apps/cn/nav-web`；日志与 3 张独立审查截图在 Git 忽略的
`.cache/139-b1-review/`。Browser 使用真实生产 Nitro 构建和既有隔离上游，
不代表生产验收。中断用例从实际 `transitionrun` 触发 `emulateMedia` 或真实键盘
动作，记录 CSS/DOM/焦点；不暂停动画、不改 Vue 状态或请求，不使用固定 sleep / retries。

| 检查 | 结果 |
| --- | --- |
| 受影响 Vue/Browser 文件 ESLint、Vue/Less 文件 Stylelint | PASS |
| `pnpm run style:policy` | PASS；新增债务 0，原 ambient raw 75 不变 |
| `pnpm run typecheck` | PASS |
| 首次 `pnpm run build` | PASS |
| 首轮 4 个定向 Browser owner，workers=1 | FAIL，27/29；新增两用例暴露中断采样与进入完成时序问题 |
| 搜索中断用例定向复测 | PASS，1/1；取消会干预 CSS 的暂停手段，改为事件触发媒体切换 |
| 修复 Preferences 完成检测后的必要重建 | PASS；因生产代码修复额外构建一次，未复用旧产物；既有工具链提示保留 |
| 最终 Preferences / Nav Shell / Resource Routing 定向复测 | PASS，16/16；修复后 ESLint / Typecheck 通过 |
| 最终覆盖结果 | 29 个场景全部有通过证据；没有跳过或放宽原测试 |
| 全仓库 Go/DB、全部 Browser/Visual、云服务/生产、Golden 更新 | NOT RUN，超出本批范围 |
| 远端 CI、维护者人工视觉签核 | NOT RUN，本地记录不能代替远端或人工验收 |

```sh
pnpm run test:browser tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/nav-shell.spec.ts tests/browser/regression/preferences-foundation.spec.ts tests/browser/regression/resource-routing-preferences.spec.ts --workers=1
pnpm run test:browser tests/browser/regression/nav-home-header.spec.ts -g 'Suggestions own' --workers=1
pnpm run test:browser tests/browser/regression/nav-shell.spec.ts tests/browser/regression/preferences-foundation.spec.ts tests/browser/regression/resource-routing-preferences.spec.ts --workers=1
```

首次进入 hook 的双 RAF 检测可能早于 Vue 自身双 RAF，已改为观察局部 to 类，
保留真实 220ms 过渡和 Reduced 即时完成；未放宽断言。Search 的暂停采样尝试
（含一次定向失败复测）会人为保留动画值，最终测试只被动观测真实浏览器事件。

覆盖搜索单一进入 / 即时退出 / Loading 不重播，首次及同页 Reduced 切换，
Preferences 入场/退场中断、关闭后不可聚焦、快速重开、保存立即落盘、三页键盘及
滚动、真实路由离开、桌面/移动入口与中英文；原有失败、Abort、IME、主题、
响应式和设置语义断言保留。

## 人工视觉待审与保留边界

1. 搜索建议 220ms 的速度与快速重复打开是否自然、是否出现闪烁。
2. Preferences 遮罩/面板是否协调，160ms 退场是否留下可见残影。
3. 1440px Light / 390px Dark 及既有响应式场景是否保持原布局。

既有 Modal 尚无完整通用 focus-trap / Escape 关闭契约；本批仅修复新增退场时间
带来的焦点与交互风险，不宣称完成通用无障碍治理。移动首页已知 Footer hydration
债务仍由原严格断言单独核对，未扩大豁免。截图不替代维护者签核。

审查图：`search-ready.png`、`preferences-desktop-light.png`、
`preferences-mobile-dark.png`。最终日志：`build-final.log`、`browser-final.log`、
`search-recheck-final.log`、`eslint-final.log`、`typecheck-final.log`；均位于上述缓存目录。
