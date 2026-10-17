# #108-F2 — Light Card System & Compare / Timeline Polish

日期：2026-10-17。依据维护者提供的 F2 冻结 V1，仅调整游戏/网站 Compare 与 Changes 的表现层。

**工程验收 PASS；最终视觉验收 PENDING MAINTAINER。** 截图全部来自本机生产构建 + 既有浏览器 Fixture，不代表真实开发数据、生产验收或逐资产安全审核。

## 起点与文件范围

- 当前分支始终为 `dev`；开始时工作树干净，无需要保留的未提交修改。
- 实际源码基线：`a30e3a2d0be6c16e9b93a17e637a79c6bc8bf8f3`（F1），当时本地 HEAD 与已记录的 `origin/dev` 相同。没有 pull、reset、rebase、stash、切分支或创建 Worktree。
- 终态为包含本报告的 F2 提交；完整 SHA 见交付消息，也可用 `git log -1 --format=%H -- docs/acceptance/108-f2-light-card-compare-timeline.md` 定位。
- 文件范围 **PASS**：下列 7 个文件均在实施文档第 6 章白名单内。截图、测量及日志均被 Git 忽略。

| 文件（前六项相对 `apps/cn/nav-web/`） | 修改 |
| --- | --- |
| `app/assets/styles/pages/insights/compare.less` | Picker/滚动 region 独立轻表面、控件与行内密度、局部文字对比度、Focus；保留不透明 Sticky 层级。 |
| `app/assets/styles/pages/insights/changes.less` | 筛选器与每条事件轻表面、日期/轴/节点、局部文字对比度与中性 Hover。 |
| `app/components/insights/compare/InsightComparePicker.vue` | 仅根节点增加 `max-w-[960px]`，由 Tailwind 管理结构宽度。 |
| `app/components/insights/InsightsChangeExplorerFeed.vue` | 仅在原有 section/ol/li/link 增加 Grid、间距与定位类；没有新增包装层。 |
| `tests/browser/regression/insights-compare.spec.ts` | 在原用例中增加材质、无嵌套、滚动后 Sticky 坐标/遮挡/层级、Popup 与删除按钮焦点检查。 |
| `tests/browser/regression/insights-changes.spec.ts` | 在原用例中增加单层材质、移动轴/节点几何、整卡焦点及同实体多事件保留检查。 |
| `docs/acceptance/108-f2-light-card-compare-timeline.md` | 本批验收记录。 |

两个 Vue 文件的 `<script>` 与基线逐字一致（归一换行后比较）。必要的模板类是为遵守“Tailwind 管结构、Less 管外观”：若仅在 Less 新增宽度、日期列与事件 Grid，会产生新的结构所有权冲突。没有改变 DOM 层级、链接、事件、computed、watch 或请求。

Overview、Domain、Pulse、Workspace、导航、共享媒体、全局 Tokens、i18n、SEO 源码、Composable、后端、数据库、CI、依赖、样式债务基线和 Visual Golden 均未修改。Picker 的消费者仍仅为两个 Compare 页面，Feed 仍仅供 Changes 使用。

## 轻材质与交互

| 计算样式 | Light | Dark |
| --- | --- | --- |
| 表面 Token 组合 | `color-mix(in srgb, var(--gf-surface) 52%, transparent)` | `color-mix(in srgb, var(--gf-surface-strong) 85%, transparent)` |
| 浏览器背景 | `color(srgb 1 0.980392 0.94902 / 0.375216)` | `color(srgb 0.886275 0.909804 0.941176 / 0.12)` |
| 实际不透明度 | **37.52%** | **12%** |
| 边框 | 1px，现有 `--gf-border` 与透明色各 50%；计算 alpha 约 8.04% | 同一组合，计算 alpha 约 9.02% |
| 圆角 / 外阴影 | 既有 `--gf-radius-lg` = 16px / `none` | 同左 |
| Sticky 表头、事实列例外 | `rgb(242, 227, 208)`，alpha 1 | `rgb(11, 23, 40)`，alpha 1 |

**PASS：无嵌套卡片。** Compare 只有 Picker 与矩阵滚动 region 拥有表面；result、table 与 Chip 背景透明，输入框和 Chip 保留功能性边界。原有 Popup 的不透明底、浮层层级和阴影保持原样，不把浮层变为第三层内容卡。Changes 只有筛选器和每条事件链接拥有表面；日期 section、ol、li 和筛选组透明。

**PASS：Sticky 与交互。** 表头/事实列/左上角继续使用不透明页面背景，z-index 为 3/2/5。滚动区保持内部双向滚动和隔离；不对外层加裁切。搜索弹层在有结果时超出 Picker，命中测试确认未被遮住。原 Combobox 上下键、Enter/Escape、最多四项、移除和缓存交互保留。输入、移除按钮、滚动 region、筛选按钮和整条事件链接有可见的 2px 焦点轮廓；没有新增位移、缩放或动画。

**PASS：移动轴。** 日期在卡片组上方，ol 保留 1px 轴线、16px 内侧间距；桌面日期列 150px、中等宽度 130px，轴内侧间距 28px。节点属于原 li，以事件高度中心定位，实测节点中心横坐标在轴边界 +0.5px，纵向中心误差不超过 1px。

现有全局 `ul, ol { padding: 0 }` 是未分层规则，会压过 Tailwind 的轴间距类。仅在 Changes 专属 ol 使用 `padding: revert-layer`，使原定 Tailwind 间距生效，没有修改全局 reset 或扩大样式债务基线。

辅助文字使用现有 `--gf-text-main` 与 muted/soft 的局部混合，未新增 Token、字号或权重。代表场景按页面背景及祖先表面合成后的对比度：主要事实约 Light 12.83:1 / Dark 11.45:1，辅助文字 6.26:1 / 8.17:1；元数据 Light 5.17–5.41:1 / Dark 6.86–8.81:1。此为选定元素的计算抽检，不等于逐像素爪印背景、所有控件或全站 WCAG 审计。

## Before / After 几何

同一 Fixture、语言、主题与视口，全部 14 组场景没有整页横向溢出；Header 仍为 0px，H1 保持唯一无障碍语义。

| 场景 | Before → After |
| --- | --- |
| 1440px 中文 Light，两实体 Picker 宽度 | 1360 → 960px |
| 同场景 Picker 高度 | 188.59 → 200.59px；单层容器增加必要内边距，并非高度全面缩减。 |
| 390px 英文 Dark，两实体 Picker 高度 | 244.59 → 252.59px；宽度仍为 358px。 |
| 桌面矩阵普通格 padding | 16px 20px → 12px 16px；移动端垂直 14 → 12px。 |
| Changes 1440px 第一条卡片宽 / 高 | 1073.67 / 124.08 → 1157 / 113.08px |
| Changes 768px 第一条卡片宽 / 高 | 538.94 / 151.27 → 542.94 / 140.27px |
| Changes 390px 英文游戏卡片宽 / 高 | 358 / 149.77 → 341 / 134.77px；减少的宽度用于保留轴线。 |
| Changes 筛选器高度，1440px / 390px 英文 | 178 → 178px / 316 → 364px；手机内边距导致类别多一行，44px 操作高度保留。 |

实际双向滚动证据：1024px 四网站 `scrollLeft=180, scrollTop=51`；390px 两网站为 `180/212`，两游戏为 `180/230`。角表头 left/top、事实列 left 均在 region 内边界 1px。1440px 网站矩阵足够宽，水平滚动为 0 是正常布局，不伪称该场景发生水平滚动；另附 1024px 四实体 Sticky 截图证明两方向同时滚动。

## 契约保护

| 项目 | 结果与证据 |
| --- | --- |
| F1 / SSR / SEO | **PASS**：原 SSR 唯一本地化 H1、SR-only、零页头高度与 Compare `noindex, follow` 断言保留；14 组 Before/After 的 H1、title、name/property meta、canonical/alternate 完全一致。 |
| Compare 数据与状态 | **PASS**：原选择顺序、最多四项、地区、URL、缓存、过期响应丢弃、失败保留、insufficient/reverse、0/未知和媒体失败断言全部保留。 |
| Changes 事实与分页 | **PASS**：原相邻日期分组、`occurred_at`/`date` 精度、筛选 URL、跨页追加、cursor 失败重试、过期范围响应丢弃及媒体兜底保持；额外确认前两条同实体 href 相同但仍各自显示。 |
| 请求 / Hydration | **PASS**：沿用原 Fixture 请求与 Hydration 断言，未修改数据所有者、SSR、网络逻辑或 Fixture 响应；无新增 API 请求。 |
| 布局 / 键盘 | **PASS**：zh/en × Light/Dark × 1440/1024/768/390 由既有矩阵覆盖；内部滚动、实际 Sticky 坐标、层级命中、Popup、卡片焦点、移动轴与无页面横溢均通过。 |
| 共享模块隔离 | **PASS**：最终 diff 仅含上述白名单；所有业务脚本不变。没有因为局部样式调整扩大到其它拥有者回归。 |

## 实际检查

全部在开发工作站 `apps/cn/nav-web` 运行，复用现有依赖和 Node 24 / pnpm 12.6.0。

| 命令 / 项目 | 结果 | 日志（相对证据根目录） |
| --- | --- | --- |
| `pnpm run lint` | **PASS**；最终测试断言调整后补充两文件定向 ESLint，PASS。 | `logs/lint.txt`、`logs/lint-final-tests.txt` |
| `pnpm run stylelint` | **PASS**，包含最终辅助文字修正。 | `logs/stylelint.txt` |
| `pnpm run style:policy` | **PASS**；305 源文件，Tailwind 外观/任意外观 0；原 raw 75、important 5 基线未扩大。 | `logs/style-policy.txt` |
| `pnpm run typecheck` | **PASS**。 | `logs/typecheck.txt` |
| `pnpm run insights:semantics` | **PASS**，语义保护文件未修改。 | `logs/insights-semantics.txt` |
| `pnpm run build` | **PASS**；终态构建供最终完整浏览器回归与 After 截图共用。 | `logs/build-final.txt` |
| Compare + Changes 终态完整套件 | **PASS 124/124，2.4 分钟，无重试**；Compare 80、Changes 44。 | `logs/browser-final.txt` |
| Before 键盘基线 | **PASS 2/2**。 | `logs/before-interactions.txt` |
| 最终 Fixture 截图与测量 | **PASS**；14 场景，每阶段 40 张页面视口/局部图，SEO 前后等值与选定文字对比度检查通过。 | `logs/capture-before.txt`、`logs/capture-after.txt`、`measure-before.json`、`measure-after.json` |
| 真实开发数据 / 生产 / 远端 CI | **NOT RUN**。未启动真实数据服务、连接数据库或生产；本地通过不代表远端 CI 已完成。 | 无 |
| 最终产品审美 | **PENDING MAINTAINER**。 | 见下列截图 |

终态完整运行命令：

```sh
pnpm run test:browser tests/browser/regression/insights-compare.spec.ts tests/browser/regression/insights-changes.spec.ts --workers=1
```

诊断过程如实保留：共进行了三次本批生产构建。第一次截图发现未分层 list reset 使时间轴间距丢失，局部修复后第二次构建；随后文字抽检发现旧辅助颜色在轻表面上约 2–3.7:1，局部加强后第三次构建为终态。构建保留既有 sourcemap/exports/PURE 注释警告，无构建错误，未修改依赖。

第一次完整套件为 **123 PASS / 1 FAIL**：新加的 Popup 检查误在已选满四站、匹配结果为空时要求短空态面板超出 Picker。把相同遮挡检查移到已存在的初始非空结果状态，保留所有原业务断言，第二次完整套件 124 PASS。辅助文字修正后又完整运行一次，得到上表最终 **124 PASS**；没有以中间运行或分段相加替代终态回归。早期证据保留于 `logs/browser-initial.txt`、`logs/popup-initial-failure.md`、`logs/browser-layout.txt`；测试总数未减少，没有新增 skip 或 retry。

未运行 Go、数据库、全仓库验证、其它页面浏览器套件或 Visual Golden；本批未发现需要扩展验证范围的跨页回归证据。

## 截图与人工确认

证据根目录：`apps/cn/nav-web/.cache/108-f2-review/`，Git 忽略。`before/` 与 `after/` 同名对照；复用既有生产 Fixture Runner 与响应器，临时导出脚本仅在忽略目录，没有新截图服务、框架或 Golden。

优先查看以下 After 图片，均有 Before 同名图：

1. `after/compare-site-zh-light-1440-ready.png`、`after/compare-game-zh-light-1440-ready.png`：选择区与矩阵的两层级关系。
2. `after/compare-site-zh-light-1024-four-sticky.png`：双向滚动后的不透明 Sticky；另有 `compare-site-zh-light-1440-four-sticky.png` 宽桌面对照。
3. `after/compare-site-en-dark-390-ready.png`、`after/compare-site-en-dark-390-ready-sticky.png`：手机英文选择与滚动矩阵。
4. `after/changes-site-zh-light-1440.png`、`after/changes-site-zh-light-1440-feed.png`：筛选轻面板、左日期轴与逐条卡片。
5. `after/changes-game-en-dark-390.png`、`after/changes-game-en-dark-390-feed.png`：手机细轴、游戏图片兜底和英文时间换行。
6. `after/changes-site-zh-light-768.png`、`after/changes-site-zh-light-1440-empty.png`、`after/changes-site-zh-light-1440-retry.png`：中等屏与异常状态。

已查看代表性实际浏览器画面，未发现本批新增的文字遮挡、Sticky 透叠或页面横溢。仍请维护者确认：暖/暗主题爪印的透出程度；Picker 限宽与矩阵宽度的视觉衔接；英文手机筛选换行后的面板高度；时间轴和长名称的密度。真实数据、真实设备与最终审美均不能由 Fixture 结果替代。

完成本批提交与授权的 `dev` 推送后停止，不创建 PR、不合并 main、不部署生产、不关闭 #108、不进入 #139 或其它阶段。
