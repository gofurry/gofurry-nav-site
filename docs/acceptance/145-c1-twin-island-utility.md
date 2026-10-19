# #145-C1 — Twin-Island Refinement & Utility Consolidation

2026-10-19。**定向工程验证 PASS；C1 人工视觉 PENDING。**

维护者已明确签核 #145-B 并允许进入 C1。起始本地 `dev` 与远端均为
`f60c9cc3fe0e6599b624e80a818c5a8b6de83cdd`，工作区干净。仅在现有 dev 工作。

## 实际范围

- 完整 SVG clip 外壳改为相反方向的直线内切面，保留小圆角端点；中间上窄下宽。
  移除 SVG 描边，主/副层共用一层雾面，无新增阴影、渐变或装饰边框；焦点环保留。
- 右岛在 GitHub / 主题 / 设置 / 语言四位下方连续延伸倒梯形；三个 Phosphor
  图标位于相邻上层中心之间，承接本地化抽奖路由、站点目录搜索和原 Issues 外链。
  手机不显示二层。Nav/Game 旧 Dock、其挂载/滚动监听及孤立 Less/token 已退役。
- 新局部 `NavUtilities.vue` 复用目录 API、名称/简介过滤、SFW/NSFW、八项上限、
  域名解析、最近访问和无 view POST 语义；首次打开才加载，按语言缓存。
  关闭/换语言/路由/断点/卸载取消请求并使旧结果失效；Loading 与失败提示区分。
  Tab、Escape、点外部、设置弹层优先级、输入焦点及快速开关有定向覆盖。
- 品牌岛在桌面滚动 >=80px 后隐藏，<=24px 恢复；焦点仍在品牌时延迟隐藏。
  隐藏立即 inert/aria-hidden，160ms opacity/6px 轻移，Reduced 即时。
  手机设置弹层向桌面交还焦点时，NavBar 等待既有 handoff 完成；未改底栏业务。
- 桌面搜索/平台/Quick Access 同步下移 24px；高度 <=520px 使用原位置。
  Hint 在实际内容底部到提示之间不足 24px 时隐藏。手机搜索和 360px Hero 不改。
  工具搜索仅 180ms opacity/4px 入场、即时退场，Reduced 初始/运行中生效。

B 的实际 SSR、Footer、自然滚动以及 A2 平台控件保留。没有产品级 scrollIntoView、
新请求预热、依赖、服务端、CI 或 Golden 修改。同步当前导航文档，并纠正前端 AGENTS
中仍要求旧 Footer mismatch 的过期段落；B 及更早验收历史保留，仅追加 B 的人工签核。

## 验证

环境：Windows、Node 24.15.0、pnpm 12.6.0、仓库 Chromium、现有 Nitro 生产构建
与 Fixture。Browser 始终 `--workers=1`，retries=0，没有固定 sleep 或新增错误豁免。

| 检查 | 实际结果 |
| --- | --- |
| 受影响 Vue/TS/测试文件 ESLint、受影响 Vue/Less Stylelint | PASS |
| `pnpm run style:policy` | PASS，304 文件，零新增样式债务 |
| `pnpm run typecheck` | PASS，最终 NavBar 焦点修复后复核通过 |
| `pnpm run test:nuxt tests/nuxt/nav-utilities.nuxt.test.ts tests/nuxt/search-platform-picker.nuxt.test.ts` | PASS，3/3 |
| `pnpm run build` | PASS；初次浏览器发现焦点问题后仅因源码修复重新构建一次 |
| 首轮定向 Browser | PASS，20/20 |
| 原有保护回归及补充断言 | 初次 64 PASS / 6 FAIL，原因和处理见下 |
| 最终源码对应构建的定向复核 | PASS，9/9；汇总 89 个不同用例的最后结果均 PASS |
| Visual / Golden、全量 Browser、远端 Full、Go/DB、生产/云端 | NOT RUN，不属于 C1 定向范围 |
| 维护者 C1 动态视觉签核 | PENDING，截图不代替人工批准 |

Browser 实际命令（从 `apps/cn/web` 运行；审查截图时设置 `GOFURRY_145_C1_REVIEW=1`）：

```sh
pnpm run test:browser tests/browser/regression/nav-shell.spec.ts tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/nav-revealed-content.spec.ts tests/browser/regression/games-home.spec.ts --workers=1 --grep '145-C1|Navigation utilities cache|floating islands|Games Home SSR'
pnpm run test:browser tests/browser/regression/nav-shell.spec.ts tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/nav-revealed-content.spec.ts tests/browser/regression/games-home.spec.ts tests/browser/smoke/nav-home-locales.spec.ts --workers=1 --grep '145-B|Localized Home|145-A Home|145-A locale|145-A desktop|145-A rotating|145-A short|145-A2|145-A3|Suggestions own|Reduced mobile|Search keeps|Suggestion requests|IME owns|Quick Access|B2 Quick Modal|Provider HTML|Core cards|Real resizing|B2 Mobile Menu|Desktop Nav Shell|Mobile Nav Shell|Games Home mobile|145-C1 utility search'
pnpm run test:browser tests/browser/regression/nav-shell.spec.ts tests/browser/regression/nav-revealed-content.spec.ts --workers=1 --grep '145-C1|145-A rotating|145-A3|145-A2 inline picker'
```

覆盖 375–1440px 壳体/命中与原焦点环、上四下三中心、滚动滞回、品牌聚焦时延迟隐藏、
手机底栏/旋转、路由历史与语言、Normal/Reduced/恢复、搜索 Tab/Escape/外部点击/设置
优先级、目录缓存/外链/最近访问、SSR 数据及单次首页请求、零 hydration 错误、自然
wheel/touch/PageDown/End、两主题以及 Games 原分页/统计/消费者断言。
Nuxt 用真实组件生命周期覆盖 pending close/reopen、过期结果、语言/模式与卸载清理。

初次六项失败未隐藏：

1. 设置打开时 Fixture 尚未提供既有 appearance/patterns 请求：增加仅由真实设置操作
   授权的精确 GET/响应/一次计数，不放宽通用网络断言。
2. 手机设置旋转到桌面，品牌提前 inert 使原焦点交接失败：NavBar 在 inactive handoff
   期间不隐藏，交接变化后 RAF 复核；原 `brand.toBeFocused()` 断言原样通过。
3. 三个旧壳体总数断言：新的 SSR 双断点右壳体总数为 3、实际可见仍为 2，分别精确断言。
4. 英文 A2 旋转后类别焦点断言曾失败一次；最终构建的中英文同用例均通过。
   未修改平台业务或该断言，未单独确认此单次失败根因，保留 trace 供复查。

日志、Browser 汇总和初次失败 trace 在忽略的 `.cache/145-c1-review/`；通过用例未做
无变化的全套重复。初次大范围通过结果与最终局部修复复核分开记录，不宣称一次全绿。

## 截图与人工待审

根目录 `.cache/145-c1-review/`：

- `home-{1440x900,1440x600,1440x480,900x700,390x844}-{light,dark}.png`
- `pattern-home-{390,900,1440}-dark-focus.png`
- `terms-{900,1440}-scrolled.png`、`shared-search-900.png`

请人工确认：双岛外开方向/端点与无框材质，末尾四位下方二层的连续轮廓和错位节奏；
顶部→150px→回顶与键盘焦点；900px 工具搜索展开密度、Hero/快捷入口避让、短高度 Hint，
以及真实随机/自定义背景的可读性。手机 390/640/767 的原平台和底栏仍需体验确认。
Fixture 截图只用于审查，不是生产数据验收或视觉批准。

当前没有复现中的功能阻断；保留上述一次英文焦点失败的调查限制及未执行门禁。
不更新 Golden，不合并、不部署、不关闭 #145。C1 推回后停止，等待维护者签核，未进入 C2。
