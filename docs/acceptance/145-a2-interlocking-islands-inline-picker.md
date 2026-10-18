# #145-A2 — Interlocking Islands & Inline Platform Picker

2026-10-19；本地 `dev` 起点及核对时远端均为 `0b97b8de68be7a32987221894fb1bf17fcce93cf`，开始时工作区干净。实施依据为维护者提供的 Frozen V1，使用当前 `apps/cn/web` 所有者。状态：**LOCAL ENGINEERING PASS / HUMAN VISUAL PENDING**。

## 实际范围

- `NavBar.vue` / `components/nav.less`：两岛内侧共用唯一的 12px SVG 接口曲线，品牌填左侧、工具岛填其补集；随高度缩放。移除旧独立 `shape()` 切面与普通圆角 fallback，保留外圆角、雾面、原入口、悬浮与点击穿透。装饰 SVG 不裁切真实按钮或焦点环。
- `SearchBox.vue` / `pages/nav.less`：类别组在与输入框相同的 500px 最大宽度内居中，超宽仅轨道横滚。手机平台改为右下原位网格，8 项时第一排第四格为固定主位；其他项与主位交换、选择后保持展开。2/3/5 项按真实数量排列；平台 ID 不依赖翻译，类别/语言切换沿用默认重置，桌面选中状态可正确交接至手机。
- 删除 Bottom Sheet 的 Teleport/dialog、遮罩、滚动锁、焦点陷阱及 CSS。局部进入为 opacity + 4px / 180ms，退出即时，初始及运行中 Reduced 生效；Escape、外部点击、输入焦点、路由/断点与卸载清理均保留。沿用 Search Reveal Lock、请求取消、IME 与 URL 行为。
- 仅同步当前 Design System / Testing 说明；不改历史 A1 记录。未改 `NavHomePage.vue`、Bottom Tabs 状态机、Hero/预热/SSR/SEO、Games/Insights、依赖、CI 或 Golden。

## 实际验证

以下命令工作目录为 `apps/cn/web`，Node 24.15.0 / pnpm 12.6.0；Browser 使用本次最终产品源码的一次 Nuxt/Nitro 生产构建、Chromium、单 worker、零 retry。

| 检查 / 实际命令 | 结果 |
| --- | --- |
| `pnpm exec eslint app/components/NavBar.vue app/components/nav/SearchBox.vue tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/nav-shell.spec.ts tests/nuxt/search-platform-picker.nuxt.test.ts`；新增断言修订后复查对应两个测试文件 | PASS |
| `pnpm exec stylelint app/assets/styles/components/nav.less app/assets/styles/pages/nav.less` | PASS |
| `pnpm run style:policy` | PASS；306 文件，既有 raw-visual-value 75 未增加，其余类别均 0 |
| `pnpm run typecheck` | PASS |
| `pnpm run build` | PASS；既有依赖 DEP0155 / Rollup 注释警告，无构建错误 |
| `pnpm run test:nuxt -- tests/nuxt/search-platform-picker.nuxt.test.ts` | PASS，1/1；真实组件同页语言重置、英文平台/URL 转义、Reveal Lock 与监听解绑 |
| `pnpm exec playwright test tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/nav-shell.spec.ts --workers=1`（截图时 `GOFURRY_145_REVIEW=1`） | 首轮 48 PASS / 1 FAIL，见下方纠正记录 |
| `pnpm exec playwright test tests/browser/regression/nav-home-header.spec.ts --workers=1 --grep '145-A2 inline picker cancels'` | PASS，1/1；49 个独立 Browser 用例最终均 PASS，未重复已通过用例 |
| `git diff --check` | PASS |

首轮唯一失败是本批新增检查误以为手机页初始未 Reveal，断言 ToolDock 为 0；实际已有手机自动 Reveal，原用例明确要求保留。移除此错误前提，锁生命周期由真实 Nuxt 组件验证；原手机自动 Reveal、搜索锁、网络/IME/键盘断言不变。没有为该失败修改 Hero 或放宽已有断言。最初 Typecheck 发现遗漏的 `e621` 稳定 ID，已补齐后通过检查及最终构建。

Browser 覆盖 375/390/430/640/767/768/900/1024/1440 导航，320 中英文溢出与自然 Tab；曲线平移坐标与补集采样、浮层/空隙命中、固定主位及交换、完整类别平台数、URL 编码、输入与请求取消、Normal/初始 Reduced/运行时中断与恢复、快速开关、路由卸载、主题/语言、菜单/Preferences 和手机底栏交接。未增加 sleep、retry、mask、容差或测试豁免。

日志：忽略的根 `.cache/145-a2-{style,typecheck,build,nuxt,browser,browser-final}.log`。代表性截图：`.cache/145-a2-review/picker-390.png`、`picker-640.png`、`shell-1440.png`、`home-desktop-dark.png`；只用于审查。Fixture 的固定 Hero 图片及 API 边界不能视为真实生产验收。

## NOT RUN / 人工待验

- NOT RUN：全量 Browser/Visual、远端 Full/Visual、无关 Go/DB、生产访问与部署；Golden 不更新。
- PENDING：375/390/430 手机真实 Hero 下的雾面和 2×4 轻量感、长英文名截断/完整标签、交换后主位不动；输入与建议、Escape/外部关闭、软件键盘。
- PENDING：640/767 的类别中轴；768/900/1440 首页、Games 或 Insights 的互补切面、焦点环和实际内容避让；真实手机滚动后的底栏接管/回顶退出。
- PENDING：Normal ↔ Reduced 实时切换的动态观感，以及非 Chromium/真实设备人工复核。既有窄范围手机 Footer hydration 债务仍由原 fixture 守卫记录；未新增忽略规则，未在本批修复。

没有已知未解决的本批工程失败；人工外观签核尚未发生。停止于 #145-A2，不进入 #145-B，不合并 main、不部署、不关闭 Issue。
