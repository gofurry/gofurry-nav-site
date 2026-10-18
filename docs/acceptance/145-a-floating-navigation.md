# #145-A — Floating Navigation Shell & Responsive Handoff

2026-10-19；基线 `dev=origin/dev=eeb0ce7`，初始工作区干净。仅实施 A。

## 实现与边界

- NavBar 改为独立品牌岛与图标/操作岛，保留现有 Logo、路由、外链、主题与设置；语言改为目标文字 EN/中，并保留 query/hash。深蓝灰雾面、内端微切角，无整条遮罩；中央留白通过真实 hit test。
- >=768px 固定，768–1023px 紧凑排列仍可直接容纳全部入口；<768px 顶栏正常流动，主入口/GitHub/模式进入可访问菜单。菜单支持 Escape、外部点击、路由/断点关闭，退场立即 inert。
- 手机底栏沿用四项与外观。首页按 Hero 末端进入视窗中段且顶栏离开启动，其他页面按顶栏离开启动；交接后保持到 `scrollY<=24`。同页筛选/query/hash 不重置交接。被动 scroll + RAF 合帧及 ResizeObserver 取代 160ms 轮询，卸载清理全部资源。
- 手机顶栏与底栏互斥可访问；菜单/Preferences 打开期间保持交接状态。Preferences 关闭回实际触发按钮，跨断点关闭时回可见品牌入口；保存逻辑和三页 Editor 未修改。
- 布局/错误布局补桌面 88px 安全区；Site Detail tabs/context 避让；Insights 两级导航链接补局部 scroll margin（实测旧聚焦位置约 24px，会被遮挡）；首页 Reveal 目标仅补 scroll margin，快捷管理弹层调整至浮岛之上。
- Game Detail 侧栏、Insights Matrix 内部 sticky、NavToolDock 与 PageScrollDock 保留原实现。`NavHomePage.vue`、`NavHeader.vue`、Hero 资源逻辑、ModeSettingModal、业务数据、依赖、CI 和 Visual Golden 无改动。

## 实际验证

Windows / Node 24.15.0 / pnpm 12.6.0 / 现有 Playwright Chromium，retries=0、workers=1；Browser 均使用对应源码的生产 Nitro 构建及既有确定性 upstream。

| 状态 | 验证 |
| --- | --- |
| PASS | 修改 Vue/TS 文件定向 `pnpm exec eslint …`；修改 Less/Vue 文件定向 `pnpm exec stylelint …`。 |
| PASS | `pnpm run style:policy`，无新增 debt/豁免；`pnpm run typecheck`；`pnpm run build`；`pnpm run insights:semantics`。 |
| PASS | 最终源码构建上的 49/49 项定向 Browser（2.2 分钟），覆盖七个宽度、Normal/Reduced/实时切换、断点/历史/同页 query、焦点/业务与跨页遮挡。完整命令见下方，日志 `.cache/145-a-final-browser.log`。 |
| PASS | `git diff --check`；保护文件/Golden/lockfile 差异检查。 |
| NOT RUN | 全量 Browser/Visual、数据库/Go 集成、生产操作；本阶段不更新 Golden。 |
| PENDING | 维护者动态视觉签核；本地 Fixture 截图不是生产或人工验收。 |

首轮发现并修复了图标替换造成的 outside-click 误关闭、断点隐藏菜单后的焦点丢失。
额外复现并修复了 Insights 键盘遮挡及同页 query 重置底栏。Quick Modal 的既有中断用例增加入场生命周期收敛断言后再测试 Normal 退场，保留全部时长、命中、焦点和快速重开断言；没有 sleep、retry 或错误豁免。
构建仅保留原有 sourcemap/依赖注解等警告。日志位于根 `.cache/145-a-*.log`，截图在 `.cache/145-a-review/`。

从 `apps/cn/web` 复现最终定向集（PowerShell 可将换行合并）：

```sh
pnpm exec playwright test tests/browser/regression/nav-shell.spec.ts tests/browser/smoke/nav-home-locales.spec.ts tests/browser/regression/preferences-foundation.spec.ts tests/browser/regression/nav-home-header.spec.ts tests/browser/regression/site-detail-shell.spec.ts tests/browser/regression/insights-navigation.spec.ts tests/browser/regression/games-home.spec.ts tests/browser/regression/game-detail-content.spec.ts tests/browser/regression/error-experience.spec.ts tests/browser/regression/game-collections.spec.ts tests/browser/regression/games-search-lifecycle.spec.ts tests/browser/regression/updates.spec.ts tests/browser/regression/lottery.spec.ts tests/browser/regression/page-scroll-dock.spec.ts --workers=1 --grep '145-A|B2 |Desktop Nav Shell|Mobile Nav Shell|Localized Home shell|Preferences interrupts only|Initial Reduced mobile Preferences|Search keeps typography|missing Target evidence|Insights navigation retains|Insights mobile nested|Games Home SSR|populated detail is|real recommended and locale|hydrated keyboard focus safety|real No-JS fallback|Real Index to Detail and locale|Locale navigation cancels old categories|Locale switch preserves article|Lottery localizes all|Lottery language navigation|PageScrollDock tracks'
```

设置 `GOFURRY_145_REVIEW=1` 可同时生成审查截图，不调用 Visual runner。

## 人工检查及后续

1. 对照 `home-desktop-light.png`、`home-desktop-dark.png`、`shell-900.png`：确认切角/雾面、图标辨识、焦点和中央留白；在真实随机/自定义浅深 Hero 上再看对比度。
2. 对照 `home-390.png`、`mobile-body-handoff.png`、`mobile-menu-short.png`：在 390/640/767px 下滚动进入正文、略向上滑、回顶；打开设置、切语言/主题、前进后退、跨 768px 旋转及实时切 Reduced。
3. 在 Games、Insights、Site Detail、静态/错误页检查标题、粘性页签、搜索/管理弹层及链接可用性；1023/1024px 间仅紧凑间距变化。

#145-B 仍负责首页强制 Reveal/吸附、预热/延迟挂载、SSR/SEO；既有移动 Footer hydration debt 保留原窄范围 guard，未增加豁免。
#145-C 负责 Visual/集成终验：`visual/nav-shell.spec.ts` 的八张旧 Golden，以及包含顶栏或受布局安全区影响的 Home/Games/Insights/Site/静态/错误页全页图预计需要人工审查；旧国旗控件断言也需届时迁移。A 未重写这些 Visual spec 或 PNG，不能宣称 Visual PASS。
未创建 PR、合并 main、部署生产或关闭 Issue。
