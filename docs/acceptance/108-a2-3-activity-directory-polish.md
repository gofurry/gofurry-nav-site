# #108-A2.3 验证记录

日期：2026-10-13（Asia/Shanghai）。实施基线：当前 `dev` / `e7cf0bd`，初始工作区干净。
已完整阅读维护者提供的《Editorial Activity & Topic Directory Polish》冻结 V1 文档。
本记录是本地工程验证，不代表维护者审美验收、真实线上数据验收或远端 CI 通过。

## 修改范围：PASS

仅修改以下五个 Nav Web 文件及本记录：

- `app/components/insights/activity/InsightsOverviewActivity.vue`
- `app/components/insights/overview/InsightsOverviewActivityItem.vue`
- `app/components/insights/overview/InsightsOverviewDirectory.vue`
- `app/assets/styles/pages/insights/overview.less` 的 Activity/Directory 专属规则
- `tests/browser/regression/insights-overview.spec.ts`

Activity 保留 1200px 起的 3:9、窄屏上下排布；收紧标题区间距，H2 为 26–30px，
七日总量为 30px/600。中性领域图标为 20px，容器桌面 36px、手机 30px；事件名称
16px/600，领域和时间 12px/500，说明 13px/400，行内上下留白 20px。
Directory 保留 960px 起的等宽双栏及游戏四项/网站两项；分组标题加入已有 Phosphor
领域图标，条目标题 16px/600、说明 13px/400、上下留白 18px。
两区保留细分隔线和透明默认背景；整行 Hover 仅背景/箭头颜色变化，键盘整行轮廓和
focus-ring 清晰，无新增位移、缩放、动画、图片、依赖或请求。

## 不变量与隔离：PASS

- 首页 `index.vue`、数据工具、排序/合计/UTC 精度、类型、服务、i18n 原文均未修改。
- 保留最多五条事件、实体链接、完整/单源/全失败/空状态、真实 0 与未知值区别。
- 保留游戏四项和网站两项精确本地化目标；无重复 Domain/Changes 入口、嵌套链接或独立箭头按钮。
- 保留两个并行 SSR Overview 源、各源独立失败、既有服务重试数及零 Hydration 补读。
- 保留所有 A2.2-B 图片资格、失败回退、去重、Logo、无 JavaScript SSR 及网络断言。
- 对 Less AST 按选择器比较（仅归一化换行）：所有非 Activity/Directory 声明与基线一致；
  `.overview-heading`、`.overview-text-link`、Hero、双生态和图片规则没有变化。
- Git 文件范围没有触及其他 Insights 页面、共享活动/图片组件、Shell、Nav/Games 业务、
  后端、数据库、全局 Tokens/字体、style-debt 或任何 Visual Golden。

## 本地验证

Node 24.15.0；Nav Web 使用其 `packageManager` 固定的 pnpm 12.6.0；未安装或升级依赖。
以下命令在 `apps/cn/nav-web` 执行：

| 状态 | 命令 / 结果 |
|---|---|
| PASS | `pnpm run lint` |
| PASS | `pnpm run stylelint` |
| PASS | `pnpm run style:policy`；raw 75、important 5，其余债务 0，基线未改 |
| PASS | `pnpm run typecheck` |
| PASS | `pnpm run insights:semantics` |
| PASS | `pnpm exec vitest run --project unit tests/unit/insight-overview-presentation.test.ts`；27 项 |
| PASS | `pnpm run build`；生成当前源码的 production Nitro build |
| PASS | `pnpm run test:browser tests/browser/regression/insights-overview.spec.ts --workers=1` 及下述定向修正复测，累计覆盖全部 122 项 |
| PASS | `pnpm exec eslint tests/browser/regression/insights-overview.spec.ts`；最终测试修改后检查 |
| PASS | `git diff --check`、白名单与非目标样式检查 |
| NOT RUN | 全仓库/Go/数据库/跨模块完整浏览器回归、Visual Golden 更新；没有相应修改或授权需求 |
| NOT RUN | 远端 CI、生产环境验证、维护者最终审美验收；本地结果不替代它们 |

浏览器首轮为 115 PASS / 7 FAIL：本批新增的空态最小高度断言错误地要求 `0px`，
实际 Grid 默认值为合法的 `auto`。未改产品来迁就测试；将该新增断言修正为允许
`auto` 或 `0px`，仍拒绝固定的非零最小高度，并保留事件数量、无空列表、路由、
合计、响应式及错误账本断言。随后执行：

```text
pnpm run test:browser tests/browser/regression/insights-overview.spec.ts --workers=1 --grep "merged ordinary events stay bounded"
```

定向复测 7/7 PASS，未添加重试或固定延迟。其余 115 项对应代码未再改变，不重复执行。
原有测试断言未删除；新增真实 Tab 顺序、整行 Focus/Hover、文字/箭头不重叠、长实体名
完整折行、移动端事件数量及 1440/390 来源失败检查。
Build 有 Tailwind sourcemap 与 Nitro cache-driver external 警告，退出码为 0；
本批未扩展修改构建工具，production runtime 浏览器验证使用此构建完成。

## 临时截图与人工验收

设置 `GOFURRY_OVERVIEW_REVIEW_DIR` 后由原有 Playwright fixture 输出，目录：
`apps/cn/nav-web/.cache/a2-3/review/`（忽略、不提交）。
`zh/en × light/dark × 1440/1024/768/390` 的 16 组均有全页、Activity 和 Directory 截图：

- `visual-full-{locale}-{theme}-{width}.png`
- `activity-{locale}-{theme}-{width}.png`
- `explore-{locale}-{theme}-{width}.png`

另有 `events-{0,1,2,3,4,5,8}-390`、`long-title-390`、`failure-{nav,game,all}-{1440,390}`、
中英文 `partial-empty` 等局部截图；保留原有 Hero/双生态审阅图。
测试日志保存在 `.cache/a2-3/`，包括首轮失败和修正后的定向复测结果。
这些均为 fixture 渲染；游戏美术和网站 Logo 是既有稳定替代素材，不是生产候选证明。

已抽查桌面/手机、中文/英文、Light/Dark、1024/768、部分失败/全失败截图，未见文字
裁切或横向溢出；精确几何、UTC/日精度、六路由和交互由浏览器断言补充。
维护者仍需确认：

- Activity 3:9 的阅读节奏、标题/七日总量是否合适，手机长名称和日期是否舒适。
- Directory 双栏及手机堆叠的留白、说明层级、中性 Hover 与键盘 Focus 是否符合品牌。
- 整页与既有 Hero/双生态衔接、原有 Shell 浮动控件、真实内容下的最终观感。

本批工程检查完成后停止；不进入 A2.4 / #139，不合并 main、不部署、不关闭 #108。
