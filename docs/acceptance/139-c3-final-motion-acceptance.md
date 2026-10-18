# #139-C3 — Final Motion Acceptance

日期：2026-10-18（Asia/Shanghai）。验收分支为现有本地 `dev`。
代码版本 **S = `b8b3819b6dbf78e19547bc73f8a98736d0f3d02b`**；开始时工作区干净，
本地 HEAD、`origin/dev` 与 `git ls-remote` 只读查询一致，无冻结基线漂移。

本批只新增本记录，不改变产品代码、测试、依赖、CI 或 Golden，不重跑已通过的
本地定向测试。下列完整代码门禁只对应 S；本记录提交以 S 为父提交，
追加后的 HEAD 不同于 S。该提交为 documentation-only，变更仅本文件，
可用 `git diff --name-only b8b3819b6dbf78e19547bc73f8a98736d0f3d02b HEAD` 核对。
文档提交的 CI 选择结果不能代替 S 的完整代码验收。

## 阶段证据索引

下表汇总原记录的本地最终通过证据，包含有说明的必要复测，不表示所有用例
曾在同一次执行中通过。各阶段截图、Fixture 与 Agent 检查均不等于人工或生产验收。

| 阶段 / 原验收记录 | 已实现范围与本地结论 | 人工待审 |
| --- | --- | --- |
| [A](139-a-motion-foundation.md) | Motion Contract；Reveal、Spotlight、Search 最新目标与 Timer/RAF 收敛；持续 Loading 静态降级。静态检查/类型/构建 PASS；Nuxt 8、定向 Browser 48 个场景有通过证据 | 真实等待反馈、即时定位与切页 |
| [B1](139-b1-core-interaction-motion.md) | 搜索建议 220ms 入场；Preferences 局部进入/退出与即时不可交互、焦点返回。静态检查/类型/构建及 29 个定向场景有通过证据 | 入退场节奏、关闭后焦点与点击 |
| [B2](139-b2-navigation-motion.md) | Mobile Menu、Bottom Tabs、Quick Modal Reduced；Site Popover 220ms。仅两份 Less 产品改动；静态检查/类型/构建及 26 个定向场景有通过证据 | 居中、材质、上下方向与快速开关 |
| [C1](139-c1-games-surface-motion.md) | Games Home 卡片/Stats/News Reduced；Similar revision 防止 Timer→nextTick→RAF 过期工作。静态检查/类型/构建 PASS；Nuxt 16/16，Browser 29/29 + Closure 6/6 | wrap、目标位置、四槽与末页占位 |
| [C2](139-c2-nav-insights-motion.md) | Compare Popup 180ms / 新增项 160ms，仅进入、删除即时。静态检查/类型/构建 PASS；基线 4/4、最终定向 8/8；Hero 仅源码/既有测试只读审计 | Popup/新增项节奏、快速删除无残影 |

单一规范仍为 [Motion Contract V1](../frontend/motion.md)。技术选择为
Vue Transition + Less + 局部 `useReducedMotion()`；没有引入 motion-v、GSAP、
全局 Motion Store、第二套 Token 或通用动画框架。复杂手势/共享布局另立议题。

## 当前代码远端门禁

所有代码门禁均须核对 `headSha == S`，时间记录使用 UTC。Fast 工作流名称为
`engineering-foundation`，实际文件为 `.github/workflows/checks.yml`。

| 门禁 | 状态 | 证据 |
| --- | --- | --- |
| Fast / engineering-foundation | PASS | [39057986066](https://github.com/gofurry/gofurry-nav-site/actions/runs/39057986066)，push，2026-10-17 18:25:00–18:27:49Z，headSha=S |
| repository-policy | PASS | [job 117469867993](https://github.com/gofurry/gofurry-nav-site/actions/runs/39057986066/job/117469867993) |
| nav-web-fast | PASS | [job 117469868007](https://github.com/gofurry/gofurry-nav-site/actions/runs/39057986066/job/117469868007)，含静态检查、Unit/Nuxt、类型、Guards、构建与 Smoke |
| nav-web 汇总 | PASS | [job 117470408918](https://github.com/gofurry/gofurry-nav-site/actions/runs/39057986066/job/117470408918) |
| Full | FAIL（门禁未通过；GitHub 原始 conclusion=`cancelled`） | [39059292625](https://github.com/gofurry/gofurry-nav-site/actions/runs/39059292625)，workflow_dispatch / dev，2026-10-17 18:42:41–18:54:36Z，headSha=S |
| nav-web-build | PASS | [job 117473769364](https://github.com/gofurry/gofurry-nav-site/actions/runs/39059292625/job/117473769364)，1m38s；产物 `nav-web-full-output-b8b3819b6dbf78e19547bc73f8a98736d0f3d02b` |
| nav-web-full / games | PASS | [job 117474123080](https://github.com/gofurry/gofurry-nav-site/actions/runs/39059292625/job/117474123080)，217 passed，作业 8m10s |
| nav-web-full / sites-other | PASS | [job 117474123064](https://github.com/gofurry/gofurry-nav-site/actions/runs/39059292625/job/117474123064)，293 passed，作业 6m38s |
| nav-web-full / insights-nav | FAIL（未完成；原始 conclusion=`cancelled`） | [job 117474123041](https://github.com/gofurry/gofurry-nav-site/actions/runs/39059292625/job/117474123041)，10m6s，超过 10m 上限 |
| nav-web-image | PASS | [job 117474122990](https://github.com/gofurry/gofurry-nav-site/actions/runs/39059292625/job/117474122990)，3m4s，无 push/deploy |
| nav-web-full-gate | FAIL | [job 117476264954](https://github.com/gofurry/gofurry-nav-site/actions/runs/39059292625/job/117476264954)，REGRESSION=cancelled，exit 1 |
| Visual | NOT RUN | Full 未通过，按顺序前置条件停止；未触发，故没有本次 run URL/headSha |
| 维护者动态视觉签核 | PENDING | 尚未收到维护者本人通过反馈或关闭授权 |

Fast 中按变更路径跳过的 Go/DB 等无关 job 不记为 PASS。本批未额外执行本地
lint/Unit/Nuxt/Browser/Build、全仓库 Go/DB、真实外部服务或生产验收。

Full 使用一次 Nitro 构建及同 SHA 产物，核对 build、三个 Regression 组、image
和最终 gate。手动 image job 的 `push:false` 只验证构建，不发布/部署。
实际仅触发 Full 一次，未重跑。结束后只读远端查询仍为 S。
Visual 必须后置、同 SHA；本次因阻断未执行，没有更新 Golden。

### 阻断证据与最小复现

GitHub 的 insights-nav 作业注解明确为：`The job has exceeded the maximum
execution time of 10m0s`。作业 18:44:25Z 开始，测试步骤 18:45:23Z 宣告
`Running 506 tests using 1 worker`；18:54:25Z 最后一条完成记录为编号 417，
`insights-workspace.spec.ts:72` 的英文 Prices / US / 390 Dark 场景通过。
18:54:26Z 出现 `context canceled` / `The operation was canceled.`。
日志共有 417 条通过记录，没有已报告的断言失败；剩余 89 项未获得最终结果，
不能把该组当作 PASS，也不能据此断言产品没有回归。

归属为现有 Full `nav-web-full / insights-nav` 作业预算与显式分组容量：
`.github/workflows/nav-web-full.yml` 的 10m 包含环境准备，
`.github/scripts/nav-web-regression-groups.mjs` 分配 18 个 owner。这里只证明本次
运行超过预算，不推断具体慢点或改变测试覆盖。后续最小排查建议是另行授权后
分析该组实测耗时与预算/分组，保留全部 owner、单 worker、零 retry 和原断言；
C3 没有改 workflow、测试、产品或进行优化。

本次触发命令（仓库根目录，执行时 `dev=S`）：

```sh
gh workflow run nav-web-full.yml --repo gofurry/gofurry-nav-site --ref dev
gh run view 39059292625 --repo gofurry/gofurry-nav-site --json headSha,status,conclusion,url,jobs
```

可定位的最小执行单位是该完整 insights-nav 组；复现应保持 S、现有 pinned
Playwright 容器及 production build，从 `apps/cn/nav-web` 使用原 inventory：

```bash
mapfile -t specs < <(node ../../../.github/scripts/nav-web-regression-groups.mjs insights-nav)
pnpm exec playwright test "${specs[@]}" --workers=1
```

需同时保留包含准备时间的 10m job 约束。本批未额外复现或重跑，单独运行最后
一个用例不能复现组总耗时问题。

取消后 failure artifact 步骤没有上传：本 run 只有构建产物与 Docker build
记录，**没有 Regression failure screenshot/trace 链接**，不伪造截图证据。
诊断以以上作业日志/注解链接为准；本地副本位于忽略的
`.cache/139-c3-review/{full.json,insights-nav.log,full-gate.log}`。

## 人工动态视觉清单 — PENDING

使用能核对 S 的本地预览/开发版本，不用未更新的生产站替代。代表环境为
桌面 1440px Light、移动 390px Dark；同页切换 Normal → Reduced → Normal，
其中一次在动画进行时切换，随后再次交互，确认没有旧动作重播。

| 区域 | 需维护者实际操作确认 |
| --- | --- |
| Nav Search / Preferences | 建议 220ms 入场与真实 Loading；设置三页、Save/Cancel、快速关开，关闭后点击/焦点不残留 |
| Nav Mobile / Quick Modal / Popover | 正常开关自然、Reduced 即时；底部导航保持居中；快捷增删/排序/保存；Popover 上下方向、材质及 hover 无闪烁 |
| Nav Reveal / Spotlight | Reduced 即时定位、最新目标正确、自播停止；恢复 Normal 不重播旧动作 |
| Games Home | 分区正反翻页/快速 wrap；Stats/News 到达正确位置，进度与占位不变；Reduced 无旧页残影 |
| Games Detail / Search | Similar 四槽、末页留空、快速重选和中断；Search 网络未就绪仍显示真实 Loading，最终页正确 |
| Insights Compare | Popup 180ms、新增 160ms 是否自然；快速加删、Escape、键盘选择，删除即时且矩阵列无动效/闪动 |

维护者可反馈“通过”或具体页面、操作与问题。CI 或截图通过不能代签；关闭
#139 仍需明确人工视觉通过与关闭授权两项。此清单不要求全浏览器/真机矩阵。

## 已知边界与关闭准备

- Hero 的 700ms CSS 声明不等于已验证 crossfade；C2 未做运行时中间帧测量，
  未修改 Hero、ManagedAsset、Blob 交接或首屏揭示。
- 通用 Modal 完整 focus trap / Escape 是既有债务；B1 的局部焦点保护不能表述为通用修复。
- 移动 Home Footer hydration 债务保留原严格 owner 与限定条件，没有泛化错误豁免。
- B2 对五处 Nav `transition: all` 保留有依据的决策；C3 不批量替换。

当前结论：**BLOCKED（Full insights-nav 超时，最终 Gate 失败；Visual 未运行）**。Issue
[#139](https://github.com/gofurry/gofurry-nav-site/issues/139) 保持 **Open**。
没有发布评论、创建 PR、合并 main 或部署生产。

### Issue 关闭评论草稿（未发布、当前不可用于关闭）

> #139 的 A→C2 已完成 Motion Contract V1、Nav 局部交互动效与 Reduced Motion、
> Games 分页生命周期修复、Insights Compare 轻量进入动效；Hero 保持只读。
> 使用 Vue Transition + Less + 局部 useReducedMotion，未引入新动画库。
> 验收代码为 b8b3819b6dbf78e19547bc73f8a98736d0f3d02b；
> [Fast](https://github.com/gofurry/gofurry-nav-site/actions/runs/39057986066) PASS；
> [Full](https://github.com/gofurry/gofurry-nav-site/actions/runs/39059292625) 未通过：insights-nav
> 超时取消，最终 Gate FAIL；Visual NOT RUN。结论 BLOCKED。
> 维护者人工动态视觉签核及关闭授权 PENDING，因此当前保持 Open，不能关闭。
> 全部门禁与人工授权齐备后可在 dev 关闭，无需等待合并 main 或生产部署。

## C3-R1 修复与复验

日期：2026-10-18（Asia/Shanghai）；下列运行时间使用 UTC。本节为 R1 最新结论，
上文 S 的原始 Full `39059292625` 超时取消、Gate 失败和 BLOCKED 历史原样保留。

### 版本、范围与本地验证

开始时本地 `dev`、远端 `dev` 均为 `d73970d830dd623cbf7d23df5f39fa247d688a5f`，
工作区干净。R = **`ffc59a3a55503f1605330f79c63c38eb5fd843cf`**，已提交并推回 dev。
两次 dispatch 前均确认远端 dev=R；Full 完成后才触发 Visual，二者 `headSha`
均为 R。没有竞争中的同 ref 工作流，没有重试或重复 dispatch。

R 仅修改以下六个文件：

- `.github/scripts/nav-web-regression-groups.mjs`：按冻结顺序将 Changes、Compare、
  Domain、Entity、Navigation 五个 owner 移到新增 `insights-core`，同步 CLI Usage。
- `.github/scripts/nav-web-regression-groups.test.mjs`：保护四组顺序、精确成员及原有分类错误检查。
- `.github/workflows/nav-web-full.yml`：仅矩阵加入 `insights-core`。
- `contracts/nav-web-frontend.md`、`apps/cn/nav-web/AGENTS.md`、`docs/frontend/testing.md`：
  同步现行四组契约与容量重平衡说明；没有改写 #134 历史测量。

Games 14、Sites/Other 15、Insights Core 5、Insights/Nav 13 个 spec，共 47 个，
各归属一次；保留组成员及相对顺序不变。10m job 预算、单 worker、零 retry、
同 SHA 单次构建、pinned 环境和原 aggregate 判断均未改。

本地只执行以下清单验证：守卫 **PASS 8/8**，两个 CLI 清单 **PASS（5 / 13）**。
守卫同时验证全部四组 CLI、遗漏/重复/失效路径/空组/非法组名；未硬编码用例数。

```sh
node --test .github/scripts/nav-web-regression-groups.test.mjs
node .github/scripts/nav-web-regression-groups.mjs insights-core
node .github/scripts/nav-web-regression-groups.mjs insights-nav
```

范围审计和 `git diff --check` PASS。S→R 的 `apps/cn/nav-web/app`、
`apps/cn/nav-web/tests/browser`、Visual workflow 差异为空：产品源码、全部
Browser spec/fixture、Visual Golden 与 S 一致。没有运行本地 Build、Full、Visual、
全量 lint/typecheck 或 Go/DB；没有改变 Browser 测试行为、断言、依赖或超时。

### 同 R 远端结果

Full [39101604530](https://github.com/gofurry/gofurry-nav-site/actions/runs/39101604530)：
**PASS**，headSha=R，workflow_dispatch/dev；2026-10-18 05:44:48–05:54:47Z。
四组使用同次 `nav-web-full-output-ffc59a3a55503f1605330f79c63c38eb5fd843cf` 产物。
下表耗时包含准备/清理；计数取自各组实际 Playwright 日志。

| Job | 结果 / 通过数 | UTC 起止 | 作业耗时 |
| --- | --- | --- | --- |
| [nav-web-build](https://github.com/gofurry/gofurry-nav-site/actions/runs/39101604530/job/117598844736) | PASS | 05:44:52–05:46:44 | 1m52s |
| [games](https://github.com/gofurry/gofurry-nav-site/actions/runs/39101604530/job/117599183886) | PASS / 217 | 05:46:46–05:53:00 | 6m14s |
| [sites-other](https://github.com/gofurry/gofurry-nav-site/actions/runs/39101604530/job/117599183805) | PASS / 293 | 05:46:46–05:52:02 | 5m16s |
| [insights-core](https://github.com/gofurry/gofurry-nav-site/actions/runs/39101604530/job/117599183851) | PASS / 214 | 05:46:46–05:52:19 | 5m33s |
| [insights-nav](https://github.com/gofurry/gofurry-nav-site/actions/runs/39101604530/job/117599183854) | PASS / 292 | 05:46:46–05:54:42 | 7m56s |
| [nav-web-image](https://github.com/gofurry/gofurry-nav-site/actions/runs/39101604530/job/117599183812) | PASS，push:false | 05:46:46–05:48:43 | 1m57s |
| [nav-web-full-gate](https://github.com/gofurry/gofurry-nav-site/actions/runs/39101604530/job/117600512164) | PASS | 05:54:43–05:54:46 | 3s |

全部 1016 项通过，无取消、异常跳过或重试，四组都在原 10m 预算内完成。
这解除原 Full 分组容量阻断；不能代替后置 Visual。

Visual [39102215213](https://github.com/gofurry/gofurry-nav-site/actions/runs/39102215213)：
**FAIL**，headSha=R，workflow_dispatch/dev；2026-10-18 05:55:24–06:02:37Z。
[nav-web-visual job](https://github.com/gofurry/gofurry-nav-site/actions/runs/39102215213/job/117600633566)
05:55:31–06:02:37Z（7m06s）；Build PASS，原 `pnpm run test:visual` 比较
**124 passed / 13 failed**（共 137，测试耗时 5.3m），退出码 1。Golden 更新数 0。

### Visual 阻断证据与待处理项

失败均来自现有 `toHaveScreenshot`，不是本次 Full 超时：

| 现有 Visual owner | 失败范围 | 日志可确认的差异 |
| --- | --- | --- |
| `game-collections.spec.ts` | Light/Dark 的 index-desktop、timeline-desktop、timeline-mobile，共 6 | Index 高 992→944；Timeline 桌面 2809→3146、移动 4011→4671，伴随像素差异 |
| `game-detail.spec.ts` | Detail Insights，Light/Dark × Desktop/Mobile，共 4 | 分别 233、46、218、60 像素差异 |
| `games-home-closure.spec.ts` | Latest Reviews Light/Dark Desktop，共 2 | 831 / 1399 像素差异 |
| `nav-home-header.spec.ts` | Quick Sites validation Desktop，共 1 | 622 像素差异 |

完整 [Visual failure artifact（12127081528）](https://github.com/gofurry/gofurry-nav-site/actions/runs/39102215213/artifacts/12127081528)
名为 `nav-web-visual-failure`，保留 7 天，包含 HTML 报告、expected/actual/diff PNG、
error-context 与 trace。可从 `playwright-visual-report/index.html` 定位全部 13 项；
例如 `visual-test-results/nav-home-header-Nav-Home-Quick-Sites-validation-desktop-visual-chromium/`
下的 `nav-home-quick-sites-modal-desktop-diff.png` 对应最后一项。
日志/元数据副本在忽略的 `.cache/139-c3-r1-review/`；没有把诊断产物写入 Golden。

最小后续排查单位是上述四个现有 Visual owner；应在 R 和相同 pinned 环境下，
先核对 artifact 的 expected/actual/diff 与既有设计验收来源。根因尚未区分为产品
缺陷、已接受设计与基线不一致或环境因素，不能因 Full 通过或像素差异小而豁免。
本批已停止，不重跑、不更新 PNG、不改断言或产品。任何修复/基线处置需另行授权。

### R1 当前结论与文档提交

**BLOCKED — Full PASS / Visual FAIL / HUMAN VISUAL PENDING。**
未达到 `ENGINEERING ACCEPTED`；Issue #139 保持 Open，没有发布关闭评论。
没有创建 PR、切 main、合并或部署。

维护者动态清单沿用上文：1440 Light / 390 Dark，同页 Normal → Reduced → Normal；
Nav 搜索/设置/菜单/快捷弹层的快速开关与焦点，Reveal/Spotlight；Games Home
wrap、Stats/News、Similar 与 Search 的真实 Loading/最终页；Compare 快速加删与
键盘。人工签核仍 PENDING，Visual CI 及 Agent 不代签。新增的四类截图差异需要
维护者另行确定诊断/修复范围，不能直接关闭议题。

本节以独立 documentation-only 提交 D 追加，父提交为 R，D 只改变本验收文件。
**实际 Full/Visual 验证版本为 R，不是 D**；原 Fast 链接仍只对应 S，未将旧 SHA
门禁冒充新版本。可用 `git diff --name-only ffc59a3a55503f1605330f79c63c38eb5fd843cf HEAD`
核对 D 的唯一文件变化；S→R 产品源码不变的证据不等于 Visual 已通过。

## C3-R2 Visual 基线协调与最终验收

日期：2026-10-18（Asia/Shanghai）；远端运行时间使用 UTC。此节追加 R2 真实结果，
保留上文 C3 的 Full 超时和 R1 的 Visual FAIL/BLOCKED 历史，不改写当时结论。

### 授权、版本与修改边界

维护者于 2026-10-18 在本地启动当前前端、审查并明确认可现有外观（“完全没问题，
可以继续”）。这是当前外观的人工认可；**不代表上文每一项 Normal → Reduced →
Normal 动态操作均已逐项签核，也不包含关闭 #139 的授权**。

开始时工作区干净，本地 dev、origin/dev 和 GitHub 只读远端查询均为
`556c1cb8beda76b9b7becd5986881cfea31ab008`。本批 Golden 提交
**T = `449276518267a8acf1652e490180b2ea6d3b7b71`** 已推回 dev；只改变下列
13 张原路径 PNG。136 张现有 PNG 清单不变，其他 123 张字节不变；没有产品、
spec/fixture、容差/mask、工作流、依赖或 lockfile 修改，也没有采样修复。
S 的产品代码及 R1 四组规则保持原样。没有分支/worktree/PR、合并或部署。

### 逐张诊断与更新

先逐张查看 R1 [artifact 12127081528](https://github.com/gofurry/gofurry-nav-site/actions/runs/39102215213/artifacts/12127081528)
的 expected/actual/diff，再对照四个既有 Visual owner 和真实 fixture。在 pinned
容器中独立运行两次原比较：13 张均重现旧 Golden 差异，未出现加载、网络、数据或
就绪断言失败。以下每张均经过诊断后才由受限更新命令重新生成，未把 artifact actual
手工写成 Golden。路径前缀为 `apps/cn/nav-web/tests/browser/visual/__snapshots__/`。

| Owner / PNG | 语义及差异检查 | 两次诊断 / 处理 |
| --- | --- | --- |
| game-collections.spec.ts / game-collections-index-light-desktop.png | 新搜索/筛选与卡片布局；6 项及空集合计数正确 | 两次及 CI actual 字节一致；更新 |
| game-collections.spec.ts / game-collections-index-dark-desktop.png | 同上，Dark 材质正确 | 两次及 CI actual 字节一致；更新 |
| game-collections.spec.ts / game-collection-timeline-light-desktop.png | 新元信息/标签/在线及时间线；11 项顺序、分期和日期正确 | 两次及 CI actual 字节一致；更新 |
| game-collections.spec.ts / game-collection-timeline-dark-desktop.png | 同上，Dark 桌面蛇形布局正确 | 两次及 CI actual 字节一致；更新 |
| game-collections.spec.ts / game-collection-timeline-light-mobile.png | 同数据，移动纵向布局和完整内容正确 | 两次及 CI actual 字节一致；更新 |
| game-collections.spec.ts / game-collection-timeline-dark-mobile.png | 同上，Dark 移动布局正确 | 两次及 CI actual 字节一致；更新 |
| game-detail.spec.ts / game-detail-insights-light-desktop.png | 差异在标题/数值字形边缘；图表曲线、空值间断、零值/免费与轴正确 | 两次及 CI actual 字节一致；更新 |
| game-detail.spec.ts / game-detail-insights-dark-desktop.png | 同上；当前 0 / 峰值 120 / 均值 42.5 不变 | 两次及 CI actual 字节一致；更新 |
| game-detail.spec.ts / game-detail-insights-light-mobile.png | 同数据及曲线，移动布局正确 | 两次及 CI actual 字节一致；更新 |
| game-detail.spec.ts / game-detail-insights-dark-mobile.png | 同上，Dark 图表/文本完成态正确 | 两次及 CI actual 字节一致；更新 |
| games-home-closure.spec.ts / games-home-latest-reviews-light-desktop.png | 正文/元信息字形边缘；3 条评论、图片和 5 分钟/2 小时/3 天正确 | 两次及 CI actual 字节一致；更新 |
| games-home-closure.spec.ts / games-home-latest-reviews-dark-desktop.png | 同上；两次额外差别仅圆角边缘 4 个像素的 RGB 各 1/255 | 原比较器判为一致，详见下文；更新 |
| nav-home-header.spec.ts / nav-home-quick-sites-modal-desktop.png | 说明/提交文字字形边缘；两项入口、名称错误、输入焦点、位置和材质正确 | 两次及 CI actual 字节一致；更新 |

Collections 的六张采用维护者已批准的新版布局。其余七张不是凭差异面积放行：
Detail 固定时钟、30d/CN 数据和缺口/零值已核对；Reviews 的 SSR/浏览器时钟一致；
Quick Sites 已等待 Vue 进入完成、面板 transform 归位，并验证输入焦点和错误文案。
既有 fixture 等待字体、图片解码、有限动画和 RAF；未发现未完成 Canvas 或过渡残影。

Dark Reviews 两次原始位图在 (23,251)、(24,251)、(23,264)、(24,264) 的 RGB
各相差 1、alpha 均 255，属于卡片圆角栅格化边缘；第一轮与 CI actual 字节一致。
使用已安装 Playwright 1.60.0 的原 PNG comparator、原默认参数比较两轮返回 null
（无判定差异），不涉及文字、数据或布局。**没有新增或放宽容差**；后续生成及两轮
完整 spec 比较也通过原有断言。不能表述为全部 13 张均字节完全相同。

### 本地 pinned 证据

使用既有镜像
`mcr.microsoft.com/playwright:v1.60.0-noble@sha256:9bd26ad900bb5e0f4dee75839e957a89ae89c2b7ab1e76050e559790e946b948`，
Linux amd64、Node 24.15.0、pnpm 12.6.0、Playwright 1.60.0 / Chromium，保留
`GOFURRY_VISUAL_ENV=pinned`、单 worker、零 retry。临时容器从冻结 HEAD 的
git archive 构建，frozen install 和一次生产 Build PASS；未占用或重建维护者本地
前端的 .nuxt/.output。任务容器已移除，审查产物保留在忽略的
`.cache/139-c3-r2-review/`（artifact、diag-1/2、stability.json、generated-review.json、
final-1/2 日志与报告）。

在容器 `/work/apps/cn/nav-web` 实际执行，四个 owner 如下；每轮 output/report
写入不同的 /review 目录，不影响比较策略：

```sh
pnpm exec playwright test --config=playwright.visual.config.ts \
  tests/browser/visual/game-collections.spec.ts \
  tests/browser/visual/game-detail.spec.ts \
  tests/browser/visual/games-home-closure.spec.ts \
  tests/browser/visual/nav-home-header.spec.ts --workers=1 \
  -g 'Game Collections|Detail Insights|Latest Reviews|Quick Sites validation'
# 上述原比较独立执行两轮，均重现旧基线 13 FAIL；完成逐项诊断后：
pnpm run test:visual:update -- game-collections.spec.ts
pnpm run test:visual:update --grep 'Detail Insights|Latest Reviews|Quick Sites validation' \
  -- game-detail.spec.ts games-home-closure.spec.ts nav-home-header.spec.ts
# 最终将第一条命令去掉 -g 筛选，连续两次非更新比较四个完整 owner。
```

| 检查 | 实际结果 |
| --- | --- |
| 更新前两次定向诊断 | 各 13 FAIL，均为旧 Golden 差异；作为诊断历史保留 |
| Collections 受限更新 | PASS 6/6，16.7s |
| 其他七张受限更新 | PASS 7/7，20.6s |
| 最终四个 spec 非更新比较 1 | PASS 36/36，1.3m |
| 最终四个 spec 非更新比较 2 | PASS 36/36，1.2m |
| 原路径 PNG 清单 / 变更白名单 / git diff --check | PASS，136 张清单中只变更批准的 13 张 |
| 额外产品审计、全仓库 Go/DB、无关本地测试 | NOT RUN，按范围不执行 |

首次更新命令遗漏 `--` 分隔符，Playwright 把 spec 名解析为可选的 snapshot mode，
在测试启动前报参数错误，未写入 Golden；改用上述文档规定的调用格式后成功，
未改环境守卫或 runner。没有用失败重试、更换断言或扩大更新清单掩盖问题。

### 当前 T 的远端门禁与停止原因

仅 dispatch 一次 Full，触发前及完成后只读确认远端 dev=T。Full
[39107101290](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107101290)
**PASS**，`headSha=T`，2026-10-18 07:17:22–07:27:40Z。四组使用同次
`nav-web-full-output-449276518267a8acf1652e490180b2ea6d3b7b71` 构建产物；
下列通过数来自作业实际日志，耗时包含准备和清理。

| Job | 结果 / 数量 | 作业耗时 |
| --- | --- | --- |
| [nav-web-build](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107101290/job/117615244744) | PASS | 1m56s |
| [games](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107101290/job/117615597599) | PASS / 217 | 6m17s |
| [sites-other](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107101290/job/117615597549) | PASS / 293 | 6m48s |
| [insights-core](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107101290/job/117615597534) | PASS / 214 | 4m50s |
| [insights-nav](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107101290/job/117615597518) | PASS / 292 | 8m10s |
| [nav-web-image](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107101290/job/117615597555) | PASS，push:false | 4m01s |
| [nav-web-full-gate](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107101290/job/117617111363) | PASS | 3s |

总计 **1016/1016 PASS**，四组均在原 10m 预算内，无重试。该结果不覆盖下面
独立的 repository-policy 失败。

T 的 push 自动触发 [engineering-foundation 39107076666](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107076666)，
整体 **FAIL**。其中 [nav-web-fast](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107076666/job/117615203456)
和 [nav-web 汇总](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107076666/job/117615809377)
**PASS**，但 [repository-policy](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107076666/job/117615203458)
**FAIL**（07:17:11–07:17:33Z）：

```text
--- FAIL: TestFrontendFullCIKeepsCompleteGroupsAndExactBuildSource
frontend_ci_test.go:167: Full must run all fixed groups
```

只读定位：`tools/check-production-policy/frontend_ci_test.go:166` 仍要求
`games, sites-other, insights-nav` 三组，而 R1 的 Full 矩阵已为含
`insights-core` 的四组。R1 的 [自动运行 39101568396](https://github.com/gofurry/gofurry-nav-site/actions/runs/39101568396)
也有同一断言失败；这两个源文件在本批均未改变。此为现存跨守卫契约不同步，
不是本次 PNG 更新证明出的产品回归，也不能忽略为已通过。

最小复现入口为在 `tools/` 执行
`go test ./check-production-policy -run '^TestFrontendFullCIKeepsCompleteGroupsAndExactBuildSource$'`。
本批**未执行**本地 Go 测试、未修守卫或工作流；证据直接来自上述 CI 日志
（忽略缓存 `fast.json`、`fast-failed.log`）。失败不涉及浏览器截图，无失败截图
可提供，以上作业链接即最小定位证据。

R2 冻结实施文档第 3 节明确将“某个远端门禁失败”列为停止条件，因此发现自动
门禁失败后只等待已启动的 Full 结束取证，没有 dispatch Visual。T 的远端 Visual
为 **NOT RUN**，没有 T 的 Visual run URL；不得用 R1 的 124/137 或本地 36/36
替代 T 的远端全量验收。没有自动重试、改阈值、跳过失败或擅自扩大到 Go/CI 修复。

实际远端命令及取证入口：

```sh
gh workflow run nav-web-full.yml --repo gofurry/gofurry-nav-site --ref dev
gh run view 39107101290 --repo gofurry/gofurry-nav-site --json headSha,status,conclusion,url,jobs
gh run view 39107076666 --repo gofurry/gofurry-nav-site --log-failed
```

### R2 结论、遗留事项与文档提交

**BLOCKED — Full PASS / engineering-foundation FAIL（Nav Web Fast PASS） /
Visual NOT RUN / HUMAN VISUAL REVIEW ACKNOWLEDGED / ISSUE CLOSE AUTHORIZATION PENDING。**

13 张更新和本地两轮原断言验证已完成；最终远端 Visual 尚待阻断解除后的明确
后续范围。上文 Hero 中间帧、通用 Modal、移动 Footer 等既有边界不因本批消失。
维护者当前外观认可已记录，不代签未明确反馈的逐项动态清单。

本节为独立 documentation-only 提交 D2，父提交为 T，唯一修改是本验收文件的
R2 追加段；**门禁实际验证版本是 T，不是 D2**。D2 的路径选择/跳过结果不能证明
上述 Go 守卫问题已解决，可用 `git diff --name-only 449276518267a8acf1652e490180b2ea6d3b7b71 HEAD`
核对文档提交边界。原 C3/R1 记录保持原文；没有关闭或评论 #139。

### 更新后的关闭评论草稿（未发布；当前仍不可用于关闭）

> #139 的 A–C2 已完成局部 Motion 治理与 Reduced Motion 生命周期修复，继续使用
> Vue Transition + Less + useReducedMotion，没有引入新动画库。维护者于 2026-10-18
> 认可当前本地外观；C3-R2 在 pinned 环境精确协调 13 张 Golden，四个 Visual owner
> 两轮各 36/36 PASS。验收版本为 449276518267a8acf1652e490180b2ea6d3b7b71，
> [Full](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107101290) 1016/1016 PASS。
> [基础门禁](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107076666)
> 因旧三组 Go 守卫与四组矩阵不一致失败，远端 Visual NOT RUN，当前仍 BLOCKED。
> 待明确授权处理阻断并完成后续远端 Visual 后，再取得独立关闭授权；#139 保持 Open。

## C3 最终阻断修复与验收（R2 后续）

日期：2026-10-18（Asia/Shanghai）。本节依据维护者本次明确授权更新当前结论；
上文 C3、R1、R2 的失败和 BLOCKED 均保留为历史，不代表本节完成后的状态。

### 最小修复与本地验证

开始时现有 dev 工作区干净，本地和远端均为
`5dde3dfb9b77532c4f7268072ba6769777d36627`。修复提交
**F = `f6ef4919583e0d97eaa9dce217fe9a8bfd9613b9`** 已推回 dev，仅修改
`tools/check-production-policy/frontend_ci_test.go:166` 的一行预期值：
`games, sites-other, insights-core, insights-nav`。仍用 `slices.Equal` 检查精确
成员及顺序；未删除守卫、改动其余逻辑或降低检查强度。

在 `tools/` 使用 Go 1.26.9（进程级 `GOTOOLCHAIN=go1.26.9`）：

```sh
go test ./check-production-policy -run '^TestFrontendFullCIKeepsCompleteGroupsAndExactBuildSource$' -count=1
go test ./check-production-policy
```

两项分别 **PASS**（0.369s / 3.857s），`git diff --check` PASS。未修改产品、
Visual Golden、测试容差、CI 超时、依赖或其他测试逻辑；没有本地全仓库 Go/DB 检查。
日志与远端元数据留在忽略的 `.cache/139-c3-final-review/`。

### Full 证据复用的精确边界

本次维护者明确授权：相关源码未变时复用已完成 Full，不要求为此 Go 测试预期值
再次运行 Full。因此复用 T=`449276518267a8acf1652e490180b2ea6d3b7b71` 的
[Full 39107101290](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107101290)：
**1016/1016 PASS**（Games 217 / Sites 293 / Core 214 / Nav 292），构建、镜像和 Gate PASS。
**该 run 的实际 headSha 是 T，不是 F**；本次未重新触发 Full。

`git diff --name-status T F` 仅有本验收文档（R2 追加）和上述 Go 测试文件；
`apps/cn/nav-web`、`.github/workflows`、`.github/scripts`、`.node-version`
差异为空。T 与 F 的整个 Nav Web Git tree 均为
`9b64310864ace297f3d65f4d94ef217df11e5f0e`，含产品、Functional/Visual、Golden、
构建配置与依赖。Full 分组、单次构建及执行规则完全相同，符合本次复用条件。

### F 的远端结果

| 门禁 | 真实结果与版本 |
| --- | --- |
| [engineering-foundation 39108634018](https://github.com/gofurry/gofurry-nav-site/actions/runs/39108634018) | PASS，headSha=F，2026-10-18 07:41:50–07:42:31Z |
| [repository-policy job 117619845122](https://github.com/gofurry/gofurry-nav-site/actions/runs/39108634018/job/117619845122) | PASS，07:42:09–07:42:31Z，含 tooling contract tests 和 production policy |
| Nav Web Fast | F 的路径选择为 SKIPPED，不冒充重新执行；T 的 [nav-web-fast](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107076666/job/117615203456) PASS，前端/执行配置相同 |
| [Visual 39108733509](https://github.com/gofurry/gofurry-nav-site/actions/runs/39108733509) | PASS，headSha=F，2026-10-18 07:43:26–07:48:15Z |
| [nav-web-visual job 117620089811](https://github.com/gofurry/gofurry-nav-site/actions/runs/39108733509/job/117620089811) | Build PASS；**137/137 PASS**（日志 137 passed，3.3m），作业耗时 4m46s |

先确认 F 的 repository-policy 成功和远端 dev=F，再仅触发一次
`gh workflow run nav-web-visual.yml --repo gofurry/gofurry-nav-site --ref dev`；
完成后再次确认远端仍为 F。Visual 使用原 pinned Linux / Node 24 / Playwright
环境与非更新型 `pnpm run test:visual`，单 worker、零 retry，137 项全部完成。
没有重试失败用例、改 Golden、加容差或改超时。远端 foundation 按路径跳过的
其他 job 不记为执行通过。

### 最新结论与文档提交

**ENGINEERING ACCEPTED / HUMAN VISUAL REVIEW ACKNOWLEDGED /
ISSUE CLOSE AUTHORIZATION PENDING。**

旧三组断言阻断已解除，远端 Visual 已实际通过；本节取代此前 BLOCKED 作为
**当前结论**。Full 按本次明确授权复用，版本边界如上，不声称 Full 与 Visual
运行在同一 SHA。维护者于 2026-10-18 认可当前本地外观的事实沿用 R2；
未明确反馈的逐项动态操作不由 Agent 代签。既有 Hero、通用 Modal、移动 Footer
边界仍如前文，不属于本次新增工程阻断。

本节单独以 documentation-only 提交追加，父提交为 F，唯一修改为本文件。
后续文档 HEAD 不同于 F；实际 repository-policy/Visual 验证版本为 F，
Full 验证版本为 T。可用 `git diff --name-only f6ef4919583e0d97eaa9dce217fe9a8bfd9613b9 HEAD`
核对文档提交边界。#139 保持 Open；没有 PR、合并 main、生产部署或发布关闭评论。

### 最终关闭评论草稿（未发布；等待明确关闭授权）

> #139 的 A–C2 动效治理与 Reduced Motion 修复完成，继续使用 Vue Transition +
> Less + 局部 useReducedMotion，未引入新动画库。维护者已于 2026-10-18 认可
> 当前本地外观；C3-R2 在 pinned 环境协调 13 张获准 Golden。
> 三组/四组 Go 守卫已以最小修复 f6ef4919583e0d97eaa9dce217fe9a8bfd9613b9 同步；
> [repository-policy](https://github.com/gofurry/gofurry-nav-site/actions/runs/39108634018)
> PASS，[Visual](https://github.com/gofurry/gofurry-nav-site/actions/runs/39108733509)
> 137/137 PASS。经精确源码差异核对，按维护者授权复用 4492765 的
> [Full](https://github.com/gofurry/gofurry-nav-site/actions/runs/39107101290) 1016/1016 PASS。
> 当前工程验收通过；待维护者另行明确授权后关闭，不涉及合并 main 或部署。
