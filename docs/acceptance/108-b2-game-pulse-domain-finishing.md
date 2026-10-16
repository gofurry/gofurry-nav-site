# #108-B2 — Game Pulse Composition & Domain Finishing

日期：2026-10-16。范围依据维护者提供的冻结 V1 文档。基线为当前 `dev` 的 `9ae8b27f6ff0137ed8b64e41054075b0ac120ae2`；开始时工作区干净，本地与远端 `dev` 一致。未切换分支、创建 Worktree、pull、reset、stash 或 rebase。

## 原因与测量

**PASS：先测量、后修改。** 修改前使用已有 B1 Production Nitro 构建和现有 Domain Fixture，在四个宽度记录媒体框、img 布局框、自然尺寸、有效比例/max-height/object-fit、Grid 列/行与次项实际间距。

Header Fixture 自然尺寸为 460×215。1440px 时 `domain.less` 的 `max-height:300px` 实际覆盖 `pulse.less` 的 245px；770px 宽的媒体框只能高 300px。共享 `contain` 将图像实际绘制成约 641.86×300px，造成每侧 64.07px 的 CSS 留白。两行 `1fr 1fr`、主项跨行及次项居中让次项间距达到 151.25px。

| 视口 | Before 主图框 | After 主图框 | 单侧 CSS 留白 Before → After | 次项间距 Before → After |
| --- | --- | --- | --- | --- |
| 1440 | 770×300 | 770×359.89 | 64.07 → <0.01px | 151.25 → 24px |
| 1024 | 547.48×255.88 | 547.48×255.88 | 0.02 → 0.02px | 62.09 → 24px |
| 768 | 407.13×190.28 | 407.13×190.28 | <0.01 → <0.01px | 29.30 → 24px |
| 390 | 358×167.31 | 358×167.31 | 0.01 → 0.01px | 20 → 24px |

以上微小余量为像素取整。测量使用 `contain` 的实际绘制尺度与媒体框/图片框双重对比，未仅凭 `object-fit` 判定通过，也不判断真实美术自身是否带内边距。

证据在 Nav Web 忽略目录 `.cache/108-b2-review/measure-before.json` 与 `measure-after.json`；`before/pulse-zh-light-{1440,1024,768,390}.png` 是未修改的基线截图。

## 实施与边界

**PASS：仅七个白名单文件。**

- `apps/cn/nav-web/app/components/insights/domain/InsightsGamePulse.vue`：增加一个次要作品私有包装；Tailwind 管理 7:5 双栏、顶对齐、24px 次项间距及 ≤760px 单列。两个次项在平板和手机均保持横向缩略图行。角色、作品选取和详情链接不变；增加的媒体来源标记只说明本地布局分支，不是安全证明。
- `apps/cn/nav-web/app/assets/styles/pages/insights/pulse.less`：移除 Header 固定高度限制及迁往模板的结构规则；媒体框保持 460:215；直接触及的 650 字重归一为 600。Capsule 按自然比例收缩到内容宽度，主图高度最多 300px、次图最多 92px/72px；不裁切、不拉伸、不使用模糊补图。仅成功选中的 Capsule 应用这组尺寸，失败/缺图继续使用有界的 Header 比例 initials 占位。
- `apps/cn/nav-web/app/assets/styles/pages/insights/domain.less`：仅删除 11 条 Pulse 竞争规则。B1 Header/Rail/Trend/Dimension 以及 Domain Activity/Continue 尾部规则逐字保持；截图未提供需要尾部微调的证据。
- `apps/cn/nav-web/scripts/fixtures/insights-domain.mjs` 与 `tests/browser/fixtures/insights-domain.ts`：只增加默认关闭的角色/媒体/长名称/真实零值场景。另行深比较确认默认 Panel 响应仍等于原 `mockGameHome()`；未修改 Overview Fixture。
- `apps/cn/nav-web/tests/browser/regression/insights-domain.spec.ts`：保留 B1 全部 47 项，增加 28 项 B2 场景与原矩阵内的图片几何、US 价格、本地化链接、焦点及透明表面断言。复用现有可选截图出口，不新增框架或 Golden。
- 本验收报告。

`selectGamePulse()`、`pulseDiscountPrice()`、`entity()` 的 Header/Capsule 选择规则、共享媒体组件与路由均未修改。Header 非空但请求失败只保留 initials；测试明确排除 Capsule 请求，且占位稳定后聚焦各链接不会触发新图片请求。三个角色缺项不补位、不保留固定高度。默认重复候选依旧由原逻辑去重成 91/92/93。

**PASS：隔离。** 无 Overview、Games Home/Showcase/Collections、其它 Insights 页面、全局 Tokens、B1 专属组件、Backend、SQL、数据库、CI、依赖、视觉基线变更；无新增请求、运营配置或动画。继续使用最新成功观测人数，而非实时在线；价格仍为 US/USD，0 与不可用保持不同。

## 验证记录

全部命令在 `apps/cn/nav-web` 执行，Node 24.15.0、项目 pnpm 12.6.0；未安装/升级依赖。

| 检查 | 结果与证据 |
| --- | --- |
| `pnpm run lint` | **PASS**；最后的浏览器断言调整又执行了该 spec 的定向 ESLint，PASS。 |
| `pnpm run stylelint` | **PASS**。 |
| `pnpm run style:policy` | **PASS**；305 个源文件，Tailwind 外观/任意外观均为 0，未扩展债务基线。 |
| `pnpm run typecheck` | **PASS**。 |
| `pnpm run insights:semantics` | **PASS**；导航、价格、地区身份、时间线、维度与 Compare 保护保留。 |
| `pnpm run build` | **PASS**；本批最终产品源码只构建一次。现有依赖的 sourcemap、exports/注释警告保留，未因此调整工具链。 |
| `pnpm run test:browser tests/browser/regression/insights-domain.spec.ts --workers=1` | **PASS：终态完整 75/75，2.0 分钟，无重试**；原 B1 47 项 + B2 28 项。证据 `logs/browser-final.txt`。 |
| 真实开发美术 1440/390px | **NOT RUN**：本地 Game/Nav 配置存在，但只读预检确认 loopback 9998/9999 均未监听；未启动服务、Collector 或连接生产。证据 `logs/real-preflight.json`。 |
| Go/数据库/跨模块全量浏览器/Visual Golden/远端 CI | **NOT RUN**：无本批对应改动或跨模块回归证据；本地通过不代表远端 CI 完成。 |

第一次完整浏览器运行是 **73 PASS / 2 FAIL**。两项失败均为新增的 Header 失败测试误把资源身份契约写成每 URL 仅一次尝试；记录显示请求只涉及三个原 Header URL，SSR/Hydration 期间有重复尝试。改为完整核对允许的 Header URL 集合、禁止任何 Capsule 请求、占位稳定后不再请求；共享请求失败账本及全部既有断言保持。修正后的两项定向复测 **2/2 PASS**，随后完整重跑终态，未将累加通过数当作终态完整运行。

覆盖：Games/Sites × zh/en × Light/Dark × 1440/1024/768/390；Games 初始四次、Sites 三次上游读取及无新增 Hydration 读取；原 URL、SSR、日期、切片、图表、独立失败保护；全部/仅主项/缺主项/缺折扣/缺最新/全缺；460×215 Header、231×326 / 616×353 / 184×69 Capsule、无图及 Header/Capsule 失败；长中英文名称、玩家 0/1,240、峰值 2,800、US 折扣币种、发布日期、无重复 ID 和无整页横溢。

补充调试记录：一次构建前的开发模式几何试探在等待图表时超时，不作验收证据。正式 Before 来自已有生产构建，After 和所有验收截图来自本批生产构建；两者均为 Fixture。

## 截图与人工验收

**PENDING MAINTAINER：工程通过不等于人工审美接受。** 所有截图/日志在 Git 忽略目录 `apps/cn/nav-web/.cache/108-b2-review/`，不上传配置或更新 Golden。`fixture/` 共 74 张生产构建使用本地合成数据的截图，不是真实开发或生产美术。日志为 `logs/{lint,lint-spec-final,stylelint,style-policy,typecheck,insights-semantics,build,browser,browser-fallback-final,browser-final}.txt`。

优先检查这四组整页 + Pulse 局部：

- `fixture/full-game-zh-light-1440.png`、`fixture/pulse-zh-light-1440.png`
- `fixture/full-game-zh-light-390.png`、`fixture/pulse-zh-light-390.png`
- `fixture/full-game-en-dark-1024.png`、`fixture/pulse-en-dark-1024.png`
- `fixture/full-game-zh-light-768.png`、`fixture/pulse-zh-light-768.png`

另有全 16 个 Games 语言/主题/宽度组合的 Pulse 截图，`pulse-capsule-*`、`pulse-fallback-*`、`pulse-roles-*`、`pulse-long-name-*` 关键状态，以及 B1 既有 Sites 代表截图。可选截图导出在焦点断言后恢复静止状态；不把键盘焦点环误作新增卡片外框。

请维护者确认：主封面完整性、7:5 一大两小与紧凑右栏、768px 英文折行、390px 主图高度与信息顺序、Capsule 自然比例和失败占位、明暗主题阅读层级。真实 Steam 图像仍待有可用开发服务时人工查看；本批没有逐资产 SFW 证明，也没有拓展图片资格或改变来源。不得将 Fixture 几何通过当作真实美术/内容安全验收。

完成后只提交并按授权推回当前 `dev`，不创建 PR、合并 main、部署、关闭 #108 或进入 #139。
