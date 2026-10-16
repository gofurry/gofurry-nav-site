# #108-B1 Domain Visual Foundation & Consistency

日期：2026-10-16（Asia/Shanghai）。执行基线为本地 `dev` 的 `528337b39016c4d64bf53bb7dafaf4659122f17f`，开始时工作区干净。按维护者提供的 V1 冻结文档实施；工程检查通过，人工视觉接受仍为 **PENDING MAINTAINER**。

## 修改与边界：PASS

实际修改六个白名单文件：

- `apps/cn/nav-web/app/assets/styles/pages/insights/domain.less`
- `apps/cn/nav-web/app/components/insights/domain/InsightMetricRail.vue`
- `apps/cn/nav-web/app/components/insights/domain/InsightTrendWorkspace.vue`
- `apps/cn/nav-web/app/components/insights/domain/InsightDimensionExplorer.vue`
- `apps/cn/nav-web/tests/browser/regression/insights-domain.spec.ts`
- 本验收记录。

四处收口：

1. Header 保留原组件全部信息，H1 归一为 700，简介 400 并小幅收紧行距/间距；收录数为 26px/600，快照时间为 12px。没有修改 `generated_at`、`as_of`、文案或时间格式。
2. Metric Rail 保留七项 Sites 指标及 Games 商业模式/平台分组，数值字重 600。项间距收敛为 24px，128px 基础宽度结合内容最小宽度保护英文长标签；手机自然露出后续项。恢复原生细滚动条，保留 2px 选中线、原按钮、焦点自动滚入及 URL 切换，不增加手势或翻页控件。
3. 趋势网格由组件 Tailwind 单独负责：Games 在 >1100px 为 9:3，Sites 在 >760px 保留 8:4，其余仍堆叠。删除 Less 中对应的竞争网格声明，保留内部线条和原间距断点。事实区文字行高统一、数值使用 500 字重。
4. 维度选项具有至少 44px 点击高度，选中项为 600；数据名称归一为 600，名称与数字按基线对齐，辅助数据稍增至 12px。选中名称仍使用现有强调色和下划线，没有新卡片材质。

按冻结规范 §2.5，Domain 共用章节标题 650 归一为 600；除此之外没有修饰 Pulse、最近变化或继续观测。Pulse 组件、构图、图像路由与请求完全未改。

三份被修改 Vue 组件的 `<script>` 与基线逐字相同。两个页面入口、Header 组件、composables、ECharts 主/切片组件、API 与 i18n 均未修改。Overview、其它业务页、公共组件、全局 Tokens、样式债务、Backend/数据库、CI、依赖及 Golden 没有改动。`git diff --name-only` 和 `git diff --check` 范围检查通过。

## 验证结果

所有验证命令均在 `apps/cn/nav-web` 执行，实际为 Node 24.15.0 / 项目锁定 pnpm 12.6.0；复用现有依赖，没有安装或升级。

| 检查 | 结果 |
| --- | --- |
| `pnpm run lint` | **PASS**；最后补充测试后另跑该 spec 的定向 ESLint，PASS。 |
| `pnpm run stylelint` | **PASS**。 |
| `pnpm run style:policy` | **PASS**；305 个源文件，既有 5 个 important / 75 个 ambient 原始值债务不变，基线未修改。 |
| `pnpm run typecheck` | **PASS**。 |
| `pnpm run insights:semantics` | **PASS**；指标顺序、维度、公共路径、图表生命周期和数据语义守卫原样保留。 |
| `pnpm run build` | **PASS，执行一次**；已有 VueUse 注释及 package exports deprecation 警告未扩展处理。 |
| `pnpm run test:browser tests/browser/regression/insights-domain.spec.ts --workers=1` | **PASS，47/47，1.1 分钟，一次完整最终状态运行，0 retries**。 |
| 真实开发页面/数据截图 | **NOT RUN**；现有本地 Game/Nav debug 配置存在，但两个服务均未监听。没有启动后端、连接生产或读写数据库。 |
| Go/数据库/跨模块全量浏览器/Visual Golden/远端 CI | **NOT RUN**；本批无对应改动或回归信号，按任务验证预算不执行；本机 PASS 不代表远端 CI 接受。 |

原 25 项浏览器用例全部保留。原 12 项 SSR/焦点矩阵扩展为 Games/Sites × zh/en × Light/Dark × 1440/1024/768/390 共 32 项，再增加两个 Overview/Breakdown 独立失败用例。

旧 Games `>4.5`（10:2）几何断言替换为实际 CSS Grid 有效列宽比例 3 ± 0.15；Sites 为 2 ± 0.15，并核对渲染宽度与网格轨道一致。小屏断言要求侧栏完整落在图表下方，不能仅凭其 top 较大判断堆叠。

新增检查覆盖字体档位、透明无外框/无阴影、手机局部露出下一指标、原生滚动条、末项键盘聚焦后完全可见且焦点轮廓有效、聚焦不切换指标、44px 控件、英文长标签和整页无横溢。快照 `2026-09-09T04:00:00Z` 与指标日期 `2026-09-08` 独立断言；保留比率 0 与未知值、Sites 比率 / Games population、前八条概要和完整十行表格、原生 progress、切片切换/下划线、延迟 all 历史、独立 Panel 失败、主题切换、Sites 三次 / Games 四次初始请求与无额外 Hydration 读取。

## 截图与人工验收

截图和日志全部在 Git 忽略目录 `apps/cn/nav-web/.cache/108-b1-domain-review/`。`before/` 为两个修改前的 Fixture 参考；`fixture/` 为最终生产构建的 **Fixture** 截图，22 张（四张整页、六个代表组合各三张 Rail/Trend/Dimension 局部）。不是生产或已确认的真实开发数据，不更新 Golden。

推荐首先查看：

- `fixture/full-game-zh-light-1440.png`
- `fixture/full-game-zh-light-390.png`
- `fixture/full-site-zh-light-1440.png`
- `fixture/full-site-en-dark-768.png`
- 另一个主题的局部：`fixture/trend-game-en-dark-1440.png`、`fixture/rail-site-zh-dark-390.png`（同组合还有其它局部）。

截图复用本 spec 内轻量可选 `GOFURRY_DOMAIN_REVIEW_DIR` 导出，不新增运行器或截图框架。日志为 `logs/{lint,lint-spec-final,stylelint,style-policy,typecheck,insights-semantics,build,browser}.log`；真实环境脱敏预检为 `logs/real-preflight.json`。

已检查代表性画面，未发现本批新增的文字裁切或页面横溢。请维护者人工确认：Header 与 Overview 的层级衔接、七指标的滚动感知、Games 9:3 事实区宽度、英文/暗色可读性及维度选中反馈。原生滚动条的显示习惯仍受浏览器/系统设置影响，手机后续项露出和键盘滚入提供独立线索。

剩余边界：真实开发长内容与生产观感尚未验证；Fixture 中的游戏图只是布局占位。B2 的 Pulse 图片构图与 Domain 尾部精修、#139 动画均未执行。本批不宣告最终视觉接受，不合并 main、不部署、不关闭 #108；完成后停止，人工接受前不进入 B2。
