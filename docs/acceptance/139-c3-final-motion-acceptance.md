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
