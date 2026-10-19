# #145-C1.1 — Navigation Clearance & Interaction Polish

2026-10-20。基线 `dev = origin/dev = 8a4dd87f87487ae33cf858f1183ed2189ad27153`，
开工工作区干净。仅在原 dev 工作，未覆盖用户修改。

工程定向验收 **PASS**，当前无未解决阻断；维护者人工外观/拖拽签核 **PENDING**。

## 范围与测量

双岛几何/上四下三、Home SSR/自然滚动、手机平台/底栏保持原实现。
桌面右岛实测底边 116px；layout owner 提供单一 `--gf-nav-content-top: 132px`
（16 + 56 + 44 + 16），只由下列冲突入口选择使用，没有全局 main padding。

| 入口/控件 | 基线 top | 修复 top | 结论 |
| --- | --- | --- | --- |
| 900px Games Search 输入 / 高级筛选 | 104 / 105 | 132 / 133 | 筛选原被截获，修复后可点击 |
| 900px Site Group 返回按钮 | 105.11 | 133.11 | 原被截获，修复后可点击 |
| 900px Site Detail 访问入口 | 120 | 148 | 安全间隙从 4px 增加至 32px |
| 1440px Game Detail Steam / 点评 | 121 | 149 | 安全间隙从 5px 增加至 33px |

390/768/900/1024/1440 的真实 bounds、文字范围和命中检测已覆盖。Games Home、
Insights Sites/Games 的共用导航、Updates、静态 Terms 入口定向审计通过，保持原布局。

Site 主 Tabs 改为普通流，旧 `top=88px` 断言替换为滚动后 `top=-100px`、回顶键盘
切换和精确 query/请求契约；Target/Aside 仅在 xl 保留 sticky，top 使用同一 132px。
路由/历史、Roving Tabindex、横向焦点、Target pending/race 原断言保留。

工具搜索仅修改 `components/nav.less`：局部底边焦点提示，5px WebKit / thin 标准
冷灰滚动条、透明轨道；未改全站滚动条、目录数据或请求行为。两个主题/宽度验证
八条上限、真实溢出、键盘滚动、焦点和 Escape；C1 缓存/外链/设置优先级继续通过。

快捷管理保留现有存储/增删/七项限制：灰色 32px 圆形 PhX，40px 点击区域；只有
Phosphor 把手启动 Pointer capture。原 DOM 单行跟随、邻项 200ms 让位，无原生 Drag Image。
预览不写存储，有效落点仅一次 reorder；取消恢复偏移并释放 RAF/捕获。
Escape、pointercancel、窗口失焦、关闭、路由和卸载有各自清理路径。
键盘上下排序保留同项焦点/状态播报，长列表只在弹窗内滚动以保持把手可见。
Dialog 初始焦点、Tab 环、Esc/遮罩/关闭归还真实 opener、退场 inert 保留。
遮罩/面板仍为 180/220ms，面板改为 4px 淡入退出，Reduced 即时收敛。

## 实际验证

Windows / Node 24.15.0 / pnpm 12.6.0；现有 Nitro 生产构建与确定性 Fixture。
Browser workers=1、retries=0，无固定 sleep、新错误豁免或 Golden 更新。

| 验证 | 结果 |
| --- | --- |
| 受影响 Vue/TS/测试文件 ESLint、Vue/Less Stylelint | PASS |
| `pnpm run style:policy` | PASS，304 源文件，零新增债务 |
| `pnpm run typecheck` | PASS，含最终键盘可见性修复 |
| `pnpm run test:nuxt tests/nuxt/nav-custom-sites.nuxt.test.ts` | PASS，4/4：fresh-mount 顺序/焦点、路由与卸载捕获/RAF 清理、无取消写入、不同行高末端落点与单次写入 |
| `pnpm run build` | PASS，最终 Pointer 边界修复后构建成功；最终 Browser 使用此输出 |
| 布局基线 | 主入口 3 FAIL；其他入口 5 PASS / Game Detail 间距 1 FAIL，均有实际测量记录 |
| 首轮实现 Browser | 32 PASS / 1 FAIL；新增测试误期待默认 Observation query，按既有规范精确修正 |
| 工具搜索外观 Browser | PASS，2/2 |
| 保护回归与 Site 修正复核 | PASS，27/27，覆盖 B SSR/自然滚动、A2、C1、Bottom Tabs、Site selector/history/race |
| 七项键盘可见性补查 | 首次 FAIL，确认仅焦点未保持可见；仅调整弹窗列表 scrollTop，最终复核 PASS |
| 快捷管理最终 Browser | PASS，6/6；上一轮 5 PASS / 1 FAIL 暴露 Normal 末端拖动未让位，改用实际指针意图计算排序、仅限制绘制边界后通过 |
| 全量 Browser/Visual、Golden、远端 Full、Go/DB、云端/生产 | NOT RUN，非本批范围 |
| C1.1 维护者人工视觉/拖拽签核 | PENDING |

实际 Browser 命令（`apps/cn/web`；截图时设置 `GOFURRY_145_C1_1_REVIEW=final`）：

```sh
pnpm run test:browser tests/browser/regression/game-detail-content.spec.ts tests/browser/regression/games-home.spec.ts tests/browser/regression/insights-navigation.spec.ts tests/browser/regression/updates.spec.ts tests/browser/smoke/static-locales.spec.ts --workers=1 --grep '145-C1.1'
pnpm run test:browser tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/site-detail-shell.spec.ts tests/browser/regression/games-search-contract.spec.ts tests/browser/regression/site-groups.spec.ts tests/browser/regression/game-detail-content.spec.ts --workers=1 --grep '145-C1.1|B2 Quick Modal|Quick Access|responsive Target shell|primary tabs|Target selector|target selector|pending Target|primary-tab|keyboard'
pnpm run test:browser tests/browser/regression/nav-revealed-content.spec.ts --workers=1 --grep '145-C1.1'
pnpm run test:browser tests/browser/regression/site-detail-shell.spec.ts tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/nav-revealed-content.spec.ts tests/browser/regression/nav-shell.spec.ts --workers=1 --grep 'missing Target|selector supports|touch selection|145-A2|145-B actual|145-B native|145-B touch|145-C1 continuous|B2 Mobile Menu|Navigation utilities cache|145-C1 utility'
pnpm run test:browser tests/browser/regression/nav-home-header.spec.ts --workers=1 --grep '145-C1.1|B2 Quick Modal|Quick Access'
```

原始失败、日志及 bounds JSON 保留在根 `.cache/145-c1-1-review/`。最终局部修复只复核
受影响快捷管理，不重复已经通过的无变化页面。测试夹具不是生产验收。
新增不同行高 Nuxt 用例最初因 spy 绑定到非实际 Storage 原型而失败；改为监听实际
localStorage 实例后，预览零写入、放下恰好一次写入及取消零写入均得到验证，未放宽断言。
最终日志：`build-pointer-final.log`、`browser-pointer-final.log`、`nuxt-pointer-final.log`。

## 人工待验

代表性截图在 `.cache/145-c1-1-review/final/`：`games-search-900.png`、
`site-group-900.png`、`site-detail-1440.png`、`game-detail-1440.png`、
`search-900-dark.png`、`search-1440-light.png`、`quick-drag-1440.png`。

请确认 900/1440 下入口间距、Site Tabs 自然滚出与 Aside；搜索底边焦点/本机滚动条；
七项拖动上下交换、边缘滚动、取消、键盘连续排序、刷新后顺序以及实时 Reduced。
手机 390 与跨 768 的原平台/底栏保持原样。截图和自动通过不代表人工视觉通过。
没有改动依赖、CI、Golden、后端或历史验收记录；提交推回后停止，不进入 C2。
