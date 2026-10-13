# #108-A2.4 — Overview 综合验收记录

日期：2026-10-13（Asia/Shanghai）；证据核对截止约 10:07 UTC。
分支：当前本地 `dev`。受测产品提交：`4f2e7a9f9d154e3aeff1fd226698b22e256e326f`。
任务开始时工作区干净，远端 `dev` 与本地产品提交相同。本次唯一拟提交变更是本报告。
未切换/创建分支，未 pull、merge、rebase、reset、stash 或清理已有文件。

已阅读维护者提供的 A2.4 中文文档、根及 Nav Web `AGENTS.md`、前端契约、测试指南、
Public Insights 契约、A2.2-B/A2.3 验收记录及当前 Overview 组件、数据投影和测试。
本批是只读审计，没有修改产品源码、测试、依赖、配置、样式债务或 Visual Golden。

**工程检查通过 ≠ 维护者最终视觉验收通过。** 本报告不替其它 Insights 页面签收。

## G1～G4 判定

| Gate | 状态 | 本次证据与限制 |
|---|---|---|
| G1 整页视觉 | **PENDING MAINTAINER** | Fixture 16 组合自动检查和截图完整性 PASS；已审视四张代表整页及手机/平板局部。Header、Hero、双生态、Activity、Directory 顺序和主次保留；最终密度、留白、主题与真实美术观感仍待维护者确认。 |
| G2 响应式与交互 | **PENDING MAINTAINER** | 完整浏览器套件 PASS，含四宽度、两语言/主题、长文折行、Tab/Focus/Hover、图片比例、局部失败及精确链接；390px 局部截图存在既有 Shell 浮动导航与文字相交的 P2 待审议项，见下文。不能将自动断言通过等同于所有叠层视觉均获认可。 |
| G3 真实数据与降级 | **NOT RUN** | 本机 Game/Nav 后端没有监听各自现有 debug 配置的 HTTP 端口，真实页面数据链路未就绪。真实候选数量、当前图片/Logo 加载、当前空数据、真实 SSR 请求账本及真实截图均未验收。Fixture 降级 PASS 单独记录，不替代本 Gate。 |
| G4 工程契约与隔离 | **PASS** | 当前产品代码重新构建一次、Overview 完整套件单次 122/122、语义及 31 项单测通过；同产品 SHA 的远端 Fast/工程 CI 完成且成功；本次产品与测试零修改。 |

工程与 Fixture 证据已可供维护者审阅；G3 缺口和 P2 观察项如实保留。
**Final Visual ACCEPTED：PENDING MAINTAINER。** 没有代替维护者确认、关闭 #108 或签收生产。

## 本次实际执行：PASS

工作目录均为 `apps/cn/nav-web`；Node 24.15.0、项目固定 pnpm 12.6.0、已有依赖。
没有安装/升级依赖，没有执行全仓库验证、Go/数据库集成或全套跨模块浏览器测试。

| 检查 | 实际命令 | 本次结果 |
|---|---|---|
| 当前源码构建 | `pnpm run build` | **PASS**，仅一次；同一 production Nitro 构建供下列浏览器测试使用 |
| 最终完整 Overview | `pnpm run test:browser tests/browser/regression/insights-overview.spec.ts --workers=1` | **PASS，122 passed，0 failed，0 skipped，约 3.0 分钟；单次完整执行，零重试、无定向补测** |
| 语义 | `pnpm run insights:semantics` | **PASS** |
| 已有纯函数测试 | `pnpm exec vitest run --project unit tests/unit/insight-overview-presentation.test.ts tests/unit/insight-overview-visuals.test.ts` | **PASS，2 个文件、31 项** |
| 截图完整性 | 只读校验文件、PNG 签名和宽高 | **PASS，16 组合 × 5 种视图 = 80 张必需图；含补充状态共 138 张** |
| 范围核对 | `git diff --name-only`、`git status --short`、提交前 `git diff --check` | **PASS，仅本报告；构建前后产品与测试内容无变化** |

浏览器结果是本次对 `4f2e7a9` 最终代码的完整运行，**不是 A2.3 的 115+7 拼接结果**。
`GOFURRY_OVERVIEW_REVIEW_DIR` 仅为本次测试进程设置，结束时移除；未写配置。
日志位于忽略的 `.cache/a2-4/build.log`、`browser.log`、`semantics.log`、`unit.log`。
构建仍有 Nuxt/Tailwind sourcemap 和 Nitro cache-driver external 警告，退出码 0；
已用该产物跑完真实 Nitro 的 fixture 浏览器套件，没有为消除警告改工具链。

## 运行契约只读核对：PASS

- `index.vue` 使用 `Promise.allSettled()` 并行读取 Nav/Game Overview，分别处理失败；
  成功时各一次 SSR，请求源只有两个，无 Game Home、每实体读取或 Hydration 补读。
  保留既有 503 重试行为；失败测试不把物理重试次数误称为始终两个 HTTP 请求。
- Hero 桌面 8:4；有图游戏内部 56:44，460:215 图片与底部证据带；无图/无可靠指标/
  图片耗尽回到数据版。艺术作品和统计入口独立，游戏 ID 与 AppID 不互换。
- 双生态及 Directory 在 960px 起两等宽栏；Activity 在 1200px 起 3:9，窄屏堆叠。
  游戏预览只取候选第二/第三项，首项保留给 Hero；不补位复用。Logo 最多五项、48px，
  失败使用名称 initials；不是用其它品牌或事件 payload 伪造图像。
- Game SQL/映射仍验证当前有效非 adult 分类、任何 adult 关系排除、Game/AppID/asset
  归属和有限 Steam HTTPS URL；每日 UTC 日期与 ID 确定性排序、最多三项。
  Nav 仍从实际近期事件对应 ID 投影明确 `nsfw='0'` 且未删除的合法本站点 icon，
  无事件不补造 Logo。A2.3 未修改这些查询、映射、视觉组件或过滤函数。
- 零值/未知值区分、Hero 指标优先级、各生态排除自己的 Hero 指标、30 日百分点、
  合法样本/覆盖率和每项 `as_of` 保留；不借 `generated_at` 伪造统一事实时间。
- Activity 按真实事件时间排序、最多五条；只有双源完整且计数合法才显示七日合计。
  精确时间为 UTC，日精度仅日期；原空态、部分失败、全失败和 Changes 链接保留。
  活动始终使用中性图标和文字，没有从事件资源取游戏美术。
- Directory 保留四个 Game 路径（players/prices/languages/compare）、两个 Site 路径
  （certificates/compare），中英文精确链接、无嵌套链接、整行键盘焦点均被测试覆盖。
- 保留无 JavaScript SSR、SEO/H1、两语言 Hydration、图片安全/失效、长标题、真实 0、
  不可靠数字和西部时区用例；没有删除、修改或放宽任何测试。

## 远端 CI

核对对象明确为受测产品 SHA `4f2e7a9`，不是本报告提交后尚未运行的检查。

| 对象 | 状态 | 证据 |
|---|---|---|
| engineering-foundation，当前产品 SHA | **PASS / completed** | [Run 38370723890](https://github.com/gofurry/gofurry-nav-site/actions/runs/38370723890)，2026-10-13 08:43:09 UTC 完成 |
| nav-web-fast 与汇总 nav-web | **PASS / completed** | 同上；Fast 包括 lint、Stylelint、style policy、Unit/Nuxt、typecheck、语义、SEO、build、Browser Smoke |
| 其余工程 Job | **PASS / completed** | 同上：detect-changes、repository-policy、foundation、postgres-integration、active-vulnerability、六个 production-go；共 13 个 Job 均 success |
| 当前产品 SHA 的远端 Full/Visual | **NOT RUN** | 未找到该 SHA 的 Full/Visual 执行，不触发新工作流；Fast 不替代 Full/Visual |
| 最近历史 nav-web-full | **FAIL，历史结果** | [Run 38335032930](https://github.com/gofurry/gofurry-nav-site/actions/runs/38335032930)，启动于当前产品提交之前；不能作为当前代码通过或本次 Overview 失败证据 |

只读查询结果保存在 `.cache/a2-4/ci.json`。本次未在本机重复执行 CI 中的 Go/数据库检查。

## Fixture 截图

**仅本机文件，不是已上传/可远程共享的附件。** 本机目录：
`E:/Git/gofurry/gofurry-nav-site/apps/cn/nav-web/.cache/a2-4/fixture/`。
该目录被 Git 忽略；游戏图是 `GAME · FIXTURE` 替身，站点图标也来自隔离 fixture。
它们不证明真实美术的主题、SFW 像素或当前 CDN 可用性。

| 宽度 | zh Light | zh Dark | en Light | en Dark |
|---|---|---|---|---|
| 1440 | PASS | PASS | PASS | PASS |
| 1024 | PASS | PASS | PASS | PASS |
| 768 | PASS | PASS | PASS | PASS |
| 390 | PASS | PASS | PASS | PASS |

这里的 PASS 指对应自动布局/运行断言通过及五种截图存在，不代表审美终验。
每组有 `visual-full-*`、`visual-hero-*`、`visual-ecosystems-*`、`activity-*`、`explore-*`；
宽高/文件大小清单为 `.cache/a2-4/fixture-manifest.json`。另有纯数据、长标题、0～8
事件、部分空源、全失败图，共 138 PNG；没有更新任何 Visual Golden。

四张优先交给维护者的 **Fixture 整页图**：

1. `visual-full-zh-light-1440.png`（1440×3232）：桌面整体层级及暖色表面。
2. `visual-full-zh-dark-1440.png`（1440×3232）：Dark 的数据/证据/分隔线层次。
3. `visual-full-zh-light-390.png`（390×5330）：手机纵向节奏、Logo 换行及完整收尾。
4. `visual-full-en-light-1024.png`（1024×3709）：英文中间断点，Hero/Activity 堆叠而双生态/Directory 双栏。

已查看上述四张原始整页及 `visual-hero-zh-light-390.png`、`explore-en-dark-390.png`、
`activity-en-light-768.png` 局部；其它组合的几何/换行/交互由本次完整套件检查。

## 真实开发环境：NOT RUN

只读预检确认本机 Game/Nav 的现有配置文件存在、模式为 debug，但配置的 HTTP 监听
均不存在。存在配置不等于数据链路已就绪；不自动新建/改写配置或启动未核验的业务进程。
脱敏预检记录：`.cache/a2-4/real-preflight.log`，不包含地址、密码或连接串。

- 当前真实 `featured_visuals`、`site_visuals`、近期事件数量：**未知 / NOT RUN**，不是 0。
- 真实图片和 Logo 成功/失败、真实数字/日期、真实两次 SSR 及四张真实整页图：**NOT RUN**。
- 当日及未来 UTC 日期候选主题相关性抽样：**NOT RUN**；未连接开发或生产数据库。
- 没有创建 `.cache/a2-4/real/` 截图，也没有把 Fixture 图复制或更名为真实图。
- 最短补验方式：维护者按既有开发配置提供已运行且确认只读的本机 Game/Nav 服务，
  确认 Nav Web 指向这两个本机源后，再只读访问 `/insights` 与 `/en/insights`，
  采集上述四组合到独立 `.cache/a2-4/real/` 并核对 SSR/资源加载；不要发送密钥。

## 问题等级、风险与等待授权事项

**P0/P1：本次已执行的 Fixture 与静态核对未发现。** 真实数据未运行，不能据此证明
当日/未来候选全部安全或主题相关。历史 A2.2-B 样本与通过记录不充当今日实测。

**P2-01：390px 局部截图中，既有移动浮动导航覆盖部分阅读内容，待维护者审议。**

- 原始证据：`visual-hero-zh-light-390.png` 中覆盖网站样本/日期附近；
  `explore-en-dark-390.png` 中覆盖最后一个网站专题的部分标题/说明。
- 复现：本次既有 visual composition 的 390px 场景，区域截图滚动到 Hero/Directory
  时可见该叠层；无需修改数据或测试。全页图仍包含完整文本，现有 Tab 用例通过，
  尚未证实永久遮挡、焦点陷阱或链接不可达，因此不提升为 P1。
- 归属依据：未修改的 `app/components/common/MobileBottomTabBar.vue` 使用固定底部
  定位；外观归 `app/assets/styles/components/nav.less`。这是共享 Shell，不是本批
  Overview 文字溢出，不能借 A2.4 擅自修理。
- 最小建议：先确认维护者是否接受这种已有浮动阅读取舍；若不接受，另行授权仅处理
  移动导航的停靠/内容避让，不改变 Overview 数据或整页设计，也不引入 #139 动画。
  候选修复白名单为上述两个 Shell 文件及已有
  `tests/browser/regression/nav-shell.spec.ts`、`insights-navigation.spec.ts`；
  最小验证为 390px 两语言/主题的滚动与焦点遮挡、Shell/Insights navigation 定向回归
  及本 Overview 回归。具体方案需维护者授权，本批没有修改这些文件或新增测试。

其它剩余风险：自动分类、成人标签排除和 URL/资产身份校验并非逐图 SFW 证明，也不
保证 Furry 主题代表性；同 URL 替换像素、错误分类、每日轮换及 CDN 缓存仍在既有边界内。
未引入日常人工审核、运营开关、调度或新系统，也未为了截图补造游戏/事件/Logo。

维护者待办：审阅四张代表图的整页节奏/主题/密度，裁定 P2-01，待真实服务就绪后完成
G3，再明确是否接受最终视觉。未获该确认前，不标记最终视觉通过。

## 提交与停止边界

只提交本报告到当前 `dev` 并按本次授权推回。截图、日志及私密本地配置不提交。
所有产品/测试/配置/Golden 均与受测产品提交一致；不涉及其它已完成模块。
本次停止，不进入 #139 或其它阶段，不创建 PR、不合并 main、不部署生产、不关闭 #108。
