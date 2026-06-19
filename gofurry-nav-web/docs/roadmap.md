# GoFurry Nav Web 样式系统路线图

## 当前状态

`v2.2.x` 样式系统迁移已完成。Less + Tailwind 分层、统一 `html.dark` 暗色入口、页面语义类、公共 `gf-*` 组件类、视觉截图守卫和性能守卫都已落地。

后续 roadmap 不再保留“样式迁移遗留桶”。新增页面、组件或视觉调整只按增量维护规则进入对应功能版本，不能再引入多套暗色入口、临时全局覆盖或未登记的深度选择器。

## 样式主线

- **Tailwind**：只负责布局、间距、尺寸、响应式和简单显示控制。
- **Less**：负责主题 token、语义组件、复杂视觉、hover、active、暗色、玻璃质感、页面级视觉系统和第三方组件局部重写。
- **暗色入口**：统一使用 `html.dark`。
- **公共类**：使用 `gf-` 前缀，例如 `.gf-button`、`.gf-card`、`.gf-input`、`.gf-pagination`。
- **页面类**：使用页面语义前缀，例如 `.nav-home-page`、`.games-page`、`.games-search-page`、`.lottery-page`、`.steam-zone-page`、`.updates-page`。

禁止继续扩大以下模式：

- 在模板里堆叠复杂颜色和暗色 class。
- 同一组件同时依赖 `dark:`、`:global(.dark ...)`、页面级 `--dark` class 控制同一视觉状态。
- 为局部修复继续追加更高优先级的 global 覆盖。
- 把页面专属视觉塞进公共组件 Less。

## v2.2.5 - Style System Migration Complete

**Status:** Completed
**Scope:** Release / Documentation / Testing
**Goal:** 将样式系统迁移作为 `v2.2.x` 完整收口，后续版本只做增量维护，不再保留样式迁移遗留桶。

### Completed Checks

- [x] 确认 games、games-search、home/nav、static、updates 全部完成 Less + Tailwind 分层迁移。
- [x] 确认新增样式必须优先使用 token 和语义类，规则沉淀到 `docs/style-system.md`。
- [x] 确认视觉截图和性能报告只作为本地临时产物，`docs/performance/reports/` 不提交截图产物。
- [x] 确认 `docs/performance/budget.json` 对迁移后的页面保持保守预算，并更新到 `v2.2.5`。
- [x] 清理 roadmap 中不再需要的迁移阶段，后续只保留增量维护规则。

### Final Acceptance

- `npm run typecheck` 通过。
- `npm run build` 通过。
- `npm run perf:guard -- --base-url http://localhost:3001` 在构建后的 preview 环境下通过。
- `npm run visual:guard -- --base-url http://localhost:3001` 覆盖核心页面亮暗色、中文/英文、桌面/移动端。
- roadmap 中不再存在独立的未来样式迁移遗留计划。

## 已完成迁移范围

完整迁移清单、命名规则和保留例外以 `docs/style-system.md` 为准。当前完成范围包括：

- 首页导航：`pages/nav.less`
- 游戏首页：`pages/games.less`
- 游戏搜索：`pages/games-search.less`
- 游戏抽奖：`pages/lottery.less`
- 静态页面：`pages/static.less`
- 更新公告：`pages/updates.less`
- 公共组件：`components/button.less`、`card.less`、`chip.less`、`footer.less`、`input.less`、`modal.less`、`nav.less`、`pagination.less`、`rating.less`、`shell.less`

## 后续增量维护规则

- 新增视觉样式优先写 Less 语义类。
- 新增复杂颜色必须先进入 token。
- 新增页面必须有稳定页面根类，并在需要时补充 `visual:guard` 关键容器检查。
- 不允许新增页面级 `--dark` class 作为主题入口。
- 不允许新增未登记的 `:deep(...)`；确需使用时先在 `docs/style-system.md` 的保留例外表中说明边界和原因。
- 不允许新增 `:global(.dark ...)`，暗色入口必须写作 `:global(html.dark ...)` 或全局 Less 中的 `html.dark`。
- `docs/performance/reports/` 只保留 `.gitkeep`，本地截图、manifest、trace 和性能报告不进入提交。

## 验证规则

阶段性验证：

```powershell
npm run typecheck
npm run build
```

完整回归验证：

```powershell
npm run preview -- --port 3001
npm run perf:guard -- --base-url http://localhost:3001
npm run visual:guard -- --base-url http://localhost:3001
```

`perf:guard` 和 `visual:guard` 建议指向构建后的 preview，避免开发模式 chunk 影响性能预算判断。
