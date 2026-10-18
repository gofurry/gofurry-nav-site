# #145-A3 — Diagonal-Cut Shell Correction

2026-10-19，现有 `dev`；开始时本地/远端均为 `62c885f1b89d1b9399bdc25ad0a083a52eefee2a`，工作区干净。依据维护者 Frozen V1，仅修顶部双岛外壳。**LOCAL ENGINEERING PASS / HUMAN VISUAL PENDING**。

## 实际变更

- 删除 `islandSeam`、窄 SVG 接头和矩形伪元素拼接。两岛各使用完整闭合 SVG 路径，统一裁切整片底色、10px 背景模糊和连续内边线；保留原材质变量、图标与高度。
- 切线均从右上到左下，56px 高度下水平位移约 32px；中间约 82% 为直线，只在上下端点用 4–5px 过渡。外端圆角仍为 18px，内侧仅增加斜口避让留白。验收依据是方向与整体视觉，不是零误差拼合。
- SSR 已输出完整装饰轮廓；局部 ResizeObserver 仅随真实尺寸校准路径，不驱动布局或导航状态，卸载时 disconnect。背景的完整 clip 同时限定命中区域，透明切口穿透；真实控件和焦点环不裁切。
- 未修改 SearchBox、平台网格、导航响应式/滚动状态机、Hero/Reveal、SSR 数据/SEO、Games/Insights、依赖、CI 或 Golden。当前 Design System/Testing 的 A2“12px 完全互补”描述属于被本次冻结决定取代的旧形状依据；本批按授权仅在此说明，不重写 A2 历史记录。

## 验证（工作目录 `apps/cn/web`）

| 命令 / 范围 | 结果 |
| --- | --- |
| `pnpm exec eslint app/components/NavBar.vue tests/browser/regression/nav-shell.spec.ts` | PASS |
| `pnpm exec stylelint app/assets/styles/components/nav.less` | PASS |
| `pnpm run style:policy` | PASS，306 文件；既有 raw 75 未增加，其余 0 |
| `pnpm run typecheck` | PASS |
| `pnpm run build` | PASS；一次最终源码构建，既有依赖 DEP0155/Rollup 注释警告，无构建错误 |
| `pnpm exec playwright test tests/browser/regression/nav-shell.spec.ts --workers=1 --grep-invert '145-A2 inline picker'` | PASS，26/26，38.8s；同一 Nitro 构建，Chromium、单 worker、零 retry；不重复无关 Picker 用例 |
| `git diff --check` | PASS |

新几何断言采样实际完整路径，检查全高斜切方向、中央 80% 直线、同源斜率、真实尺寸、材质裁切、切口/岛间空白穿透、控件命中与焦点环；替换旧 `sameCurve/error=0/12px` 断言，保留其他功能断言和严格网络/错误台账。新增 390/900/1440 深色 Home 焦点检查，不增加 sleep/retry 或测试豁免。

审查截图开关 `GOFURRY_145_A3_REVIEW=1`，忽略目录 `.cache/145-a3-review/`：`shell-{390,900,1440}-light.png` 与 `home-{390,900,1440}-dark-focus.png`。六张均已生成并逐张查看：当前截图呈贯穿全高的直斜切，未见矩形残边、接缝或图标裁切，焦点环可见。日志在根 `.cache/145-a3-{style,typecheck,build,browser}.log`。Fixture 截图仅为本地审查证据，不能替代真实设备或维护者视觉签核。

## 人工待审核 / 未执行

- **HUMAN VISUAL PENDING**：是否真正呈现“一刀从右上到左下切开后拉开”，而非鼓包/尖嘴/角落小裁切；上下柔角是否克制，整片材质、边线是否连续；三个宽度与两类背景下图标、焦点环、透明间隙是否自然。
- NOT RUN：全量 Browser/Visual、远端 Full/Visual、非 Chromium 与真实移动设备、生产访问/部署。没有更新 Golden；本批自动测试结果不构成视觉批准。
- 保留原 fixture 的窄范围手机 Footer hydration 债务守卫，未新增错误忽略规则。不进入 #145-B，不合并 main、不部署、不关闭 Issue。
