# #108-F3 — Insights Final Visual & Frontend Engineering Closure

日期：2026-10-17。基线为当前本地 `dev` 的 `2cc40b6047462762cf4cc3446a8054e117278044`，与维护者文档一致。开始时工作树干净；没有 pull、reset、rebase、stash、切分支或 Worktree 操作。

本记录区分本机 Fixture 工程验证、真实数据/生产、远端 CI 和维护者视觉认可。终态提交为本报告所在的 F3 提交，SHA 见交付消息，或用 `git log -1 --format=%H -- docs/acceptance/108-f3-frontend-engineering-closure.md` 定位。

**本地工程 PASS；生产/远端 CI NOT RUN；最终视觉 PENDING MAINTAINER。** 没有发现需要越过本批所有权边界才能处理的阻塞，不把工程通过等同于 #108 已获最终关闭许可。

## F3-A：两个末尾说明入口

**PASS：局部视觉与原生交互。** 网站说明区仅由 `.insights-domain-page.insights-site-domain .insights-data-info` 收紧；证书说明区仅由 `.insights-workspace-page[data-certificate-intelligence]` 收紧。没有通用 details/summary 覆写，没有修改 Footer。Games Domain 继续保留原 36px 间距和 1px 说明区顶线；其它 Workspace 的章节间距与顶线保留。

生产构建 Fixture 实测，zh/en × Light/Dark × 1440/390 全部一致：

| 位置 | Before | After |
| --- | --- | --- |
| Sites「继续观测」之后的说明间距 | 36px | 12px |
| Sites summary 高度 | 58px | 42px，垂直 padding 12px |
| 证书风险列表之后的说明间距（1440 / 390） | 72 / 48px | 16 / 16px |
| 证书 summary 高度 | 64.69px | 48.69px，垂直 padding 12px |
| 两处说明区顶部边线 | 1px | 0px，上方业务列表的末行边线保留 |

16 个页面场景通过键盘 Enter 展开/关闭，说明原文与 Before 完全相同，关闭后 summary 高度恢复，无整页横溢，Footer 存在。Domain 展开内容保留事实日期、`known / eligible`、可用日期和 caveat；三列在手机变为一列。现有 SSR/空数据与错误用例继续覆盖对应事实来源，未修改文案、业务计算或日期语义。

F3-A 初期使用已有 source Fixture 尝试预检，遇到 Hydration 等待超时，没有修改运行时来绕过；随后在原生产 Fixture 中注入本机 Less 编译结果，仅作局部 CSS 预检。正式 After 来自本批唯一一次终态生产构建，未注入 CSS；预检不计入生产构建通过证据。

## F3-B：消费者与职责审计

审计同时检查实际 Vue 模板/脚本、显式 import 链、Nuxt 自动生成的组件注册（包含 Lazy 名称）、动态 class/component 表达式及浏览器渲染。没有仅凭文件位置或一次无命中的文本搜索删除共享代码。

| 规则 / 消费者 | 处理与证据 |
| --- | --- |
| `.insights-ranges`：Domain `InsightMetricTrend`；Game Detail `GameTabInsights` 与 `GameInsightsTimeline` | 保留公共类与行为契约，拆开视觉作用域。Domain 由 `domain.less` 拥有扁平按钮；原分段材质限定到 `:where(.game-insights-tab)`，不增加祖先 specificity。状态与按钮归同一 owner，不再用 `!important` 互相覆盖。 |
| `.entity-insights-eyebrow`、旧 heading h3/p | **删除死规则。** 冻结文档把眉题列为 Game Detail 消费者，但当前 GameDetailMain → GameTabInsights 的标题区只有 h2 + 返回链接；没有眉题、说明 p 或 heading h3。模板/动态类与 Nuxt 注册审计一致，8 个真实 Fixture 页面均为 0 个眉题节点。不存在可保留的当前眉题计算间距。 |
| `.insights-section` 卡片、`.insights-change-list` / `.insights-change*` 旧列表 | **删除死规则。** 当前 Domain/Overview/Changes 分别使用自己的 section/活动行；没有旧卡片或旧列表实例。名称相似的 `.insights-section__heading` 仍被两个趋势组件使用，未误删组件或其语义。 |
| `.insights-section__heading`、趋势/切片 chart shell、DataInfo dl | 结构归回三个原 Vue 模板的 Tailwind：原断点、间距和 chart 高度保持。`domain.less` 保留外观；删除 `insights.less` 中已被覆盖的旧 padding、尺寸、标题和 card fallback，避免两处拥有同一结构。三个脚本逐字不变。 |
| `InsightsDataInfo` | 仅 Sites/Games Domain 消费；有效表现规则集中到 `domain.less`，删除已经被 Domain 扁平规则覆盖的旧材质。F3-A 间距特例仅限 Sites；保留 Games 原入口、内容和布局。 |
| Game Detail 实际仍消费的标题、概要、玩家/价格、状态、历史图、时间线、范围/地区和 retry | 全部保留在原 `insights.less`；不搬文件、不改业务脚本、生命周期、请求或媒体。只对本批授权的旧字重作角色归一。 |
| 通用 chart/loading/empty、动态 timeline connector、F2 列表 `revert-layer`、其它独占 Less 结构 | 保留。分别存在趋势/加载状态、实体图表/时间线或全局 list reset 的真实消费者；不把独占的历史结构都当成竞争，更不为“清零”重写整页。 |

**计算样式兼容证据：** 8 组 Game Detail（zh/en × Light/Dark × 1440/390）的选中、普通、Hover 范围按钮颜色、背景、padding、字重、圆角，Before/After JSON 完全一致。Domain 8 组同样逐字段相等：静止透明、选中下划线与 accent 文字；Hover 使用现有 accent-soft 底色。这个 Hover 底色来自原真实层叠，未因“扁平”的文档描述而丢失。两类控制均验证键盘焦点；Game 缓存/过期响应/图表主题与时间线排序继续由原业务用例保护。

### 债务与字体

| 项目 | Before | After |
| --- | --- | --- |
| `insights.less` important | 3 | 0：两条范围状态声明改为明确 owner；一条无消费者眉题声明删除。 |
| `insights/domain.less` important | 2 | 0：Domain 不再接收 Game 分段材质，无需强制重置。 |
| 总 important | 5 | 0，未转移到其它文件。 |
| Ambient raw | 75 | 75，不属于本批。 |
| Tailwind appearance / arbitrary、deep、legacy-dark | 0 | 0。 |
| 临时豁免范围内非标准字重声明 | 23 | 0，包含删除的不可达旧规则；另删除死眉题的小标签 800。 |

字重按语义选择而非四舍五入：Game 实体 h2 和 Workspace 可见 H1、证书主要结果为 700；section/card 标题、返回按钮、状态/折扣与普通指标为 600；价格地区、日期和时间线类别为 500。原标准字重保持，字号、line-height、tracking、字体家族和全局 Token 不变。Overview/Pulse 样式文件本身未改；Insights 子导航和 Domain 活动文字只退出旧 650，没有重设导航或媒体基础设施。

预算/豁免退出：**PASS**。先运行已有 Policy 得到唯一两项 stale（3→0、2→0），没有新增债务；终态 465 项页面回归全部通过后，才执行 `pnpm run style:policy:update` 并移除 Stylelint #108 允许列表。更新器仅移除两项旧 important 预算，ambient raw 75 和其它预算不变；没有手工编辑计数。撤销豁免后完整 Stylelint 与 Policy/Policy tests 再次通过。

## 验证

全部在开发工作站 `apps/cn/nav-web`，复用 Node 24 / pnpm 12.6.0 和既有依赖、生产 Fixture、Playwright owner。没有新工具/框架、Go/数据库、生产服务或 Golden 更新。

| 检查 | 状态 / 证据（相对 `.cache/108-f3-review/`） |
| --- | --- |
| `pnpm run lint` | **PASS**，`logs/lint.txt`；最后仅 Stylelint 配置变更，定向 ESLint 也 PASS，`logs/lint-config-final.txt`。 |
| `pnpm run stylelint` | **PASS**；退出豁免后的完整严格检查，`logs/stylelint-final.txt`。 |
| `pnpm run style:policy` / `style:policy:test` | **PASS**，终态 75/75 Policy tests；`logs/style-policy-final.txt`、`logs/style-policy-tests-final.txt`。初查的 74 PASS / 1 stale-budget FAIL 如实保留，随后按流程下调预算，未改测试规避。 |
| `pnpm run typecheck` | **PASS**，`logs/typecheck.txt`。 |
| `pnpm run insights:semantics` | **PASS**，原语义保护未修改，`logs/insights-semantics.txt`。 |
| `pnpm run seo:recovery:test` | **PASS**，路由身份、权威状态、sitemap 与 noindex 合约保持，`logs/seo-recovery.txt`。 |
| `pnpm run build` | **PASS**，本批一次正式构建，`logs/build-final.txt`；保留依赖的 sourcemap/exports/PURE 警告，无构建错误。 |
| Game Detail 修改前局部基线 | **PASS 8/8**，`logs/game-before.txt`。 |
| 终态 8 个 Browser owner | **PASS 465/465，9.8 分钟，无重试**，一次完整运行，`logs/browser-final.txt`。 |
| F3-A 生产 Fixture Before/After | **PASS**，16 场景 × 展开/关闭，`measure-before.json` / `measure-after.json`、`logs/capture-after.txt`。 |
| 真实开发数据 / 真实生产截图 | **NOT RUN**，没有把 Fixture 图片当成真实服务验证。 |
| 远端 CI | **NOT RUN**，本地结果不代表 CI 已完成。 |
| 维护者最终审美 | **PENDING MAINTAINER**。 |

终态浏览器命令：

```sh
pnpm run test:browser tests/browser/regression/insights-domain.spec.ts tests/browser/regression/insights-workspace.spec.ts tests/browser/regression/insights-overview.spec.ts tests/browser/regression/insights-compare.spec.ts tests/browser/regression/insights-changes.spec.ts tests/browser/regression/insights-navigation.spec.ts tests/browser/regression/insights-entity.spec.ts tests/browser/regression/game-detail-insights.spec.ts --workers=1
```

Domain/Workspace 在原用例中补充末尾说明、范围状态与展开布局断言；Game Detail 在原 spec 增加 8 个字体/层叠主题尺寸用例与可选截图输出。没有删除/跳过/弱化原有 SSR、SEO、请求预算、URL、日期、0/未知、分页、缓存、失败、媒体安全和图片回退检查，没有增加重试。

| 拥有者 | 完整终态结果 |
| --- | --- |
| Domain | **PASS 75/75** |
| Workspace | **PASS 96/96** |
| Overview | **PASS 142/142** |
| Compare | **PASS 80/80** |
| Changes | **PASS 44/44** |
| Navigation | **PASS 2/2** |
| Entity bridge | **PASS 10/10** |
| Game Detail Insights | **PASS 16/16**（原 8 项 + 本批 8 项） |

运行后只更新预算、Stylelint 配置与文档，没有再改产品或浏览器测试源码；未重复生产构建或把多轮局部结果相加冒充完整通过。构建后的自动检查通过不代表所有真实设备/字体渲染均获人工认可。

## 文件范围与截图

**PASS：共 19 个文件。** 产品范围：`app/assets/styles/pages/insights.less`；`insights/{domain,workspace,changes,compare,navigation,media}.less`；`InsightsDataInfo.vue`、`domain/InsightMetricTrend.vue`、`domain/InsightSliceTrend.vue`。测试范围：原 `insights-domain.spec.ts`、`insights-workspace.spec.ts`、`game-detail-insights.spec.ts`。治理范围：`stylelint.config.mjs`、`frontend-style-debt.json`、Nav Web `AGENTS.md`、`contracts/nav-web-frontend.md`、`docs/frontend/design-system.md` 及本报告；AGENTS 仅同步豁免/剩余预算的当前状态，没有重写历史验收记录。

禁止范围均未修改：Overview、Pulse 文件；Game Home、Showcase、Collections、Nav Home；共享图片组件/CDN；Backend/DB/API；全局 Tokens、Footer、依赖/锁文件、CI 与 Visual Golden。没有修改任何产品 `<script>`。

Git 忽略证据目录：`apps/cn/nav-web/.cache/108-f3-review/`。

- `before/` / `after/`：F3-A 各 32 张末尾截图 + Game Detail 各 8 张页面截图与 8 份样式 JSON。
- `measure-before.json` / `measure-after.json`：间距、border、字体、展开内容、范围状态、无横溢及 Footer。
- `weight-audit.json`、`debt-before.json`、`logs/consumers.txt`：源码与消费者审计。
- `f3a-preview/` 是早期 CSS 注入预检，`after/` 才是最终正式构建证据。

优先人工核对下列同名 Before/After：

1. `after/domain-zh-light-1440-closed.png`：上方链接末线与说明入口的衔接。
2. `after/domain-en-dark-390-open.png`：手机展开三项事实与文案换行。
3. `after/workspace-zh-light-1440-closed.png`：证书风险末行与说明入口。
4. `after/workspace-en-dark-390-open.png`：英文手机证书依据与 Footer。
5. `after/game-detail-zh-light-1440.png`、`after/game-detail-en-dark-390.png`：Game Detail 字体角色归一后的视觉兼容。

F3-A 的最终产品观感和 Game Detail 字重仍待维护者审美确认；Fixture 工程通过不能关闭这个人工门槛。本批不关闭 #108，不创建 PR、不合并 main、不部署、不进入 #139 或其它阶段。
