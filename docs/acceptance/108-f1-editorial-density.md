# #108-F1 — Editorial Density & Redundant Copy Cleanup

日期：2026-10-17。按维护者提供的冻结 V1 执行，仅清理 Insights 可见冗余页头和局部间距。

基线：`dev` / `bb707f6c411397981efe94a2b680ae25ef364e52`。开始时工作区干净；只读核对远端 `dev` 仍为相同提交。未新建分支、Worktree，未 pull、reset、stash、rebase 或覆盖用户工作。

| 验收层级 | 结论 |
| --- | --- |
| ENGINEERING | **PASS**；静态检查、一次构建、六个拥有者各自的终态完整回归及白名单检查通过。 |
| FIXTURE VISUAL | **PASS**；指定截图和几何/可访问性检查通过，仅代表合成数据工程验证。 |
| REAL DATA / PRODUCTION | **NOT RUN**。 |
| MAINTAINER VISUAL | **PENDING MAINTAINER**；本报告不代签最终审美验收。 |

## 分页面契约

| 页面 | 状态 | 清除与保留 |
| --- | --- | --- |
| Overview，zh/en | **PASS** | 移除可见口号及介绍；唯一 SSR H1 使用既有 `insights.overview.title`，通过 `sr-only` 保留无障碍语义。保留 `data-overview-header`，容器高度 0，无空页头外边距；导航后直接 Hero。Hero/美术/双生态/动态/目录及两源并行 SSR 未改。 |
| Sites / Games Domain，zh/en | **PASS** | H1 隐藏，删除泛化介绍；计数、单位、未知 `—`、快照标签及 `<time datetime>` 保留，改为紧凑事实条。计数 26px/600、快照 12px 不变；不把生成时间当指标事实日。B1 Rail/Trend/Dimension 与 B2 Pulse 完整保留。 |
| 证书，zh/en | **PASS** | 共享 Header 显式使用 `visually-hidden`，无重复领域眉题、可见大标题或介绍。直接进入证书验证与到期概况；verified/failed/unknown、风险列表、口径与事实日期、完整说明折叠区保留。 |
| 网站 / 游戏 Compare，zh/en | **PASS** | 隐藏通用 H1、去掉眉题/套话；收紧 picker 顶部、结果区和快照到矩阵的间距。搜索标签、已选数量、实体身份、地区、结果和不同事实时间保留；矩阵/选择器内部 DOM、材质、滚动、固定表头和业务结构均未改。 |
| Changes，zh/en | **PASS** | 隐藏通用 H1，删除眉题/介绍；缩短导航至筛选、筛选至正文的空白。领域/范围/分类标签、keyset/cursor、加载、事件布局及时间精度不变。 |
| 玩家 / 价格 / 语言，zh/en | **PASS** | Header 默认态只输出原有本地化可见 H1，删除重复领域眉题和泛化介绍；标题文本及 i18n 资源不改。排行/价格/语言正文结构、控件、指标、数据说明及章节间距保留。 |

Workspace Header 只增加一个可选 Boolean `visuallyHidden`，不做路由判断或多模式系统。普通态保留可见 H1 和紧凑下边距，隐藏态 wrapper `m-0`；两个模式都只输出一个真实 H1。

## 实际首屏密度

使用现有 Production Nitro Fixture：修改前复用 B2 构建，修改后使用本批唯一一次生产构建。下表单位为 CSS px，表示**导航容器底部到首个业务模块顶部**，不是产品固定像素合同；Domain 包含必须保留的事实条高度。

| 页面 | zh Light 1440 Before → After | en Dark 390 Before → After |
| --- | --- | --- |
| Overview → Hero | 164.38 → 40 | 189.59 → 24 |
| Sites → Metric Rail | 262.69 → 104 | 280.78 → 115.19 |
| Games → Pulse | 262.69 → 104 | 280.78 → 115.19 |
| 证书 → 证书状态模块 | 195.39 → 24 | 189.59 → 16 |
| 网站 Compare → picker | 195.39 → 24 | 189.59 → 16 |
| 游戏 Compare → picker | 195.39 → 24 | 189.59 → 16 |
| Changes → filters | 195.39 → 24 | 189.59 → 16 |
| 玩家 → metric selector | 195.39 → 114.39 | 189.59 → 77.59 |
| 价格 → region selector | 195.39 → 114.39 | 189.59 → 77.59 |
| 语言 → summary | 195.39 → 114.39 | 189.59 → 77.59 |

Overview、证书、两种 Compare、Changes 的通用 Header 均实测为 **0px**。Domain 事实条 Header 从 186.69/220.78px 降到 56/79.19px；移动端仍允许计数与快照换行。普通 Workspace Header 从 123.39/133.59px 降到 50.39/33.59px，只留下真实标题。

所有 20 个前后截图场景均无全页横向溢出。全局导航/section gap Token、公共背景、页面容器与导航组件保持原样；局部结构和外间距由 Tailwind 管理，Less 仅清理废弃页头规则及收紧事实条内边距。

## H1、SEO 与运行保护

- **PASS：语义 H1。** 浏览器按 SSR `<h1>` 数量和精确译文检查唯一性，再通过无障碍 heading role 查询验证客户端标题。隐藏态检查 `sr-only`、实际 `clip-path: inset(50%)` 及零 Header 高度；普通态检查真实可见、Header 高度由 H1 本身组成。没有 `display:none`、`v-if=false` 或删除语义标题。
- **PASS：SEO 不变。** 所有被改页面的 `<script>`（归一换行后）与基线逐字一致，包含 fetch、async-data、URL、SEO、robots 代码。额外对 10 页 × 两语言的 20 组渲染结果逐项比较 `<title>`、全部 name/property meta、canonical/alternate link，前后完全一致。元信息证据在 `measure-before.json` / `measure-after.json`，结论在 `logs/capture-after.txt`；未改 `seo.ts` 或翻译资源。
- **PASS：旧外观断言替换。** 仅替换原来要求无属性可见 H1、旧标题字号/引言或点击隐藏标题的断言。Domain 截图复位改为点击可见收录数。保留所有导航、SEO、SSR、请求数、日期、数据/安全降级、选中态、矩阵、时间线、错误和交互保护；没有 skip 或删除整项测试。
- **PASS：运行回归。** 六个拥有者各自的终态完整回归全部通过；原有 Overview 两源、Sites 三次、Games 四次初始读取，以及 Compare noindex、实体顺序、地区切换、真实 0/未知、变化分页、独立错误等由六套原有业务断言继续覆盖。

## 检查记录

全部在开发工作站 `apps/cn/nav-web` 执行，复用既有依赖、Node 24 / 项目 pnpm 12.6.0。没有 Go、数据库、全仓库验证、依赖升级、Golden 更新或生产操作。

| 检查 | 结果 |
| --- | --- |
| `pnpm run lint` | **PASS**；最终测试文件变更另行定向 ESLint，PASS。 |
| `pnpm run stylelint` | **PASS**。 |
| `pnpm run style:policy` | **PASS**；305 源文件，Tailwind 外观与任意外观计数 0，基线未变。 |
| `pnpm run typecheck` | **PASS**。 |
| `pnpm run insights:semantics` | **PASS**；原业务/时间/价格/导航保护保留，仅更新 F1 页头契约。 |
| `pnpm run seo:recovery:test` | **PASS**。 |
| `pnpm run build` | **PASS**；本批一次生产构建，后续截图及六个套件复用同一份。依赖的 sourcemap/exports 警告仍存在，无构建错误。 |
| 六个 Insights 浏览器拥有者 | **PASS：六套各自终态完整通过，无重试**。联合运行先为 437 PASS / 2 FAIL；修正两条旧外观断言后，Overview 完整复测 142/142 PASS（3.2 分钟）。不是单次联合 439 全绿；详细记录如下。 |
| 真实开发数据 / 生产观感 | **NOT RUN**；本批证据只来自本地 Fixture，未连接真实数据或生产。 |
| 远端 CI | **NOT RUN**；本地结果不能代表远端完成。 |
| 维护者最终审美 | **PENDING MAINTAINER**，由维护者查看截图后确认。 |

完整定向命令：

```sh
pnpm run test:browser tests/browser/regression/insights-overview.spec.ts tests/browser/regression/insights-domain.spec.ts tests/browser/regression/insights-workspace.spec.ts tests/browser/regression/insights-compare.spec.ts tests/browser/regression/insights-changes.spec.ts tests/browser/regression/insights-navigation.spec.ts --workers=1
```

| 拥有者 | 最终完整套件结果 | 日志 |
| --- | --- | --- |
| Overview | **PASS 142/142**（原联合运行 140 PASS / 2 FAIL） | `logs/overview-final.txt` |
| Domain | **PASS 75/75** | `logs/browser-final.txt` |
| Workspace | **PASS 96/96** | 同上 |
| Compare | **PASS 80/80** | 同上 |
| Changes | **PASS 44/44** | 同上 |
| Navigation | **PASS 2/2** | 同上 |

诊断记录：首次启动六套测试时，新 F1 断言错误地假定 `sr-only` 使用旧 CSS `clip`；实际已安装 Tailwind 使用 `clip-path: inset(50%)`。确认编译 CSS 和实际零高度后中止该轮，修正新断言并完整重跑；产品源码和构建不变。初轮日志保留在 `logs/browser.txt`。

随后联合完整运行 439 项，得到 **437 PASS / 2 FAIL（10.4 分钟）**。失败来自原有 Overview Dark ready shell 对整个 Header 的 `toBeVisible()`，正是 F1 要替换的旧外观要求。两项更新为 SSR 唯一本地化 H1、无障碍 heading role、SR-only 隐藏规则、零高度 Header 和真实 Hero 可见，保留原数据/请求/暗色检查。失败证据在 `logs/legacy-header-{1440,390}.md`；之后只完整重跑受影响的 Overview，另外五套终态代码和断言未变，不重复运行。没有跳过测试，也不把分段相加冒充单次联合全绿。

矩阵覆盖 zh/en × Light/Dark × 1440/1024/768/390：Overview、Domain 复用原矩阵；Workspace、Changes 扩展原矩阵；Compare 补上 768px 和双语言暗色布局。六套以外未增加回归范围。

## 截图与证据

根目录（Git 忽略）：`apps/cn/nav-web/.cache/108-f1-review/`。

- `before/`：20 张修改前首屏。
- `after/`：20 张修改后首屏，与 Before 同页、同数据、同语言/主题/尺寸。
- `measure-before.json` / `measure-after.json`：导航、Header、H1、业务模块位置和渲染 SEO 元信息。
- `owners/overview/` / `owners/domain/`：复用已有可选 Playwright 截图机制的完整页面及局部补充。
- `logs/`：静态、构建、浏览器、截图检查日志与白名单清单。临时截图脚本复用现有 fixture responder/app，无新的测试框架或服务。

优先人工检查以下六张独立 **Fixture** 截图，均有 `before/` 下同名对照：

1. `after/overview-zh-light-1440.png`
2. `after/sites-zh-light-1440.png`
3. `after/sites-certificates-zh-light-1440.png`
4. `after/sites-compare-zh-light-1440.png`
5. `after/changes-zh-light-1440.png`
6. `after/games-compare-en-dark-390.png`

已人工式查看这些浏览器画面，页头无空占位，操作与事实内容靠前，未发现本批新增裁切或全页横溢；不代替维护者审美验收。请确认局部间距、Domain 事实条、专业页保留标题的层级，以及英文手机的标签/日期折行。移动端对比矩阵继续保持原有内部横向滚动。

## 文件范围与停止边界

产品文件 15 个，均在白名单：

- 页级模板 8 个：`app/pages/insights/index.vue`、`changes.vue`、`games/{players,prices,languages,compare}.vue`、`sites/{certificates,compare}.vue`。
- Header 2 个：`app/components/insights/domain/InsightsDomainHeader.vue`、`app/components/insights/workspace/InsightsWorkspaceHeader.vue`。
- 样式 5 个：`app/assets/styles/pages/insights/{overview,domain,workspace,compare,changes}.less`，仅 F1 页头/间距规则。

验证文件 6 个：`scripts/insights-semantics.test.mjs` 和 `tests/browser/regression/insights-{overview,domain,workspace,compare,changes}.spec.ts`。Navigation 套件原样运行。加本报告共 **22 个文件**。

**PASS：无越界。** Domain 页面本身、Hero 与其它 Overview 组件、Pulse、Rail、趋势/维度、Compare Picker/Matrix、事件 Feed、媒体/CDN、导航/背景、全局 Tokens、翻译、SEO 工具、数据 owner、Backend/数据库、CI、依赖与 Golden 均未修改。未加入 F2 卡片、矩阵或时间线重构，未进入 #139。

完成后仅提交并按授权推回当前 `dev`，不创建 PR、不合并 main、不部署、不关闭 #108；人工验收和下一批授权由维护者决定。
