# GoFurry Nav Web 样式系统统一路线图

## 当前决策

样式迁移不再把已完成阶段长期保留在 roadmap 中。已完成的 Less 接入、token、games、games-search、公共组件、暗色入口清理和视觉守卫能力，以 `docs/style-system.md`、`docs/performance/README.md` 和 Git 历史为准。

后续不再把样式债拆到独立的未来 minor 版本作为长期遗留计划；所有剩余样式迁移统一收敛到 `v2.2.x`。目标是在 `v2.2.x` 内完成全站样式主线迁移，避免后续维护继续背负多套暗色入口、深度覆盖和页面级历史包袱。

## 样式主线

- **Tailwind**：只负责布局、间距、尺寸、响应式和简单显示控制。
- **Less**：负责主题 token、语义组件、复杂视觉、hover、active、暗色、玻璃质感、页面级视觉系统和第三方组件局部重写。
- **暗色入口**：统一使用 `html.dark`。
- **公共类**：使用 `gf-` 前缀，例如 `.gf-button`、`.gf-card`、`.gf-input`、`.gf-pagination`。
- **页面类**：使用页面语义前缀，例如 `.games-page`、`.games-search-page`、`.archive-page`、`.updates-page`。

禁止继续扩大以下模式：

- 在模板里堆叠复杂颜色和暗色 class。
- 同一组件同时依赖 `dark:`、`:global(.dark ...)`、页面级 `--dark` class 控制同一视觉状态。
- 为局部修复继续追加更高优先级的 global 覆盖。
- 把页面专属视觉塞进公共组件 Less。

## v2.2.x 迁移计划

### v2.2.1 - Home And Navigation Migration

**Status:** Planned
**Scope:** User-facing / Architecture / Maintainability
**Goal:** 将首页导航相关复杂视觉迁移到 Less 主线，避免首页继续作为 Tailwind、scoped CSS 和 global 覆盖的混合区。

#### Focus

- 首页导航内容区
- 搜索框、分组、站点卡片和 popover
- 导航 header、快捷入口、过渡条
- 中英文 active、hover、focus 和暗色一致性

#### Tasks

- [ ] 迁移首页导航模块中的复杂颜色、边框、阴影、hover 和 active 状态。
- [ ] 为 `NavContent`、`SearchBox`、`NavHeader`、`NavQuickAccess`、`SitePopover`、`GroupPopover` 等模块建立稳定语义类。
- [ ] 将导航相关页面视觉收敛到 `components/nav.less`、`components/shell.less` 或新的 `pages/nav.less`。
- [ ] 清理导航模块中可替换的复杂 `dark:`、`:global(.dark ...)` 和重复 scoped 覆盖。
- [ ] 扩展视觉截图守卫覆盖首页导航亮暗色、中文/英文、桌面/移动端。

#### Acceptance Criteria

- 首页导航、搜索、分组、站点卡片和 popover 在亮暗色下层级一致。
- 中英文路径 active 状态不回退。
- 首页导航模块不再依赖页面级暗色入口。
- `npm run typecheck`、`npm run build`、`npm run visual:guard` 通过。

### v2.2.2 - Static Pages And Updates Finalization

**Status:** Planned
**Scope:** User-facing / Documentation / Maintainability
**Goal:** 完成 about、terms、privacy、updates 的样式收口，让静态内容页和更新页也遵守同一 token 与语义类规则。

#### Focus

- 静态页面卡片、按钮、正文层级
- updates 时间线、摘要条、年份分组
- 页面背景、分隔、空状态和错误状态
- 多语言文案长度与移动端布局

#### Tasks

- [ ] 清理静态页中重复的 page vars 和局部面板样式。
- [ ] 确认 `about`、`terms`、`privacy` 页面只通过 `html.dark` 和页面 token 切换主题。
- [ ] 收敛 `updates` 页面及其子组件的局部颜色、边框、阴影和 hover 状态。
- [ ] 将可复用静态页结构沉淀为 `gf-static-page` 语义规则，页面专属差异留在 `pages/static.less`。
- [ ] 扩展视觉截图守卫覆盖 `/about`、`/updates` 和法律页移动端。

#### Acceptance Criteria

- 静态页和 updates 页面无重复页面级 dark class。
- 亮暗色下正文、按钮、时间线和状态文案对比度稳定。
- 移动端无横向溢出，长英文文案不会挤压按钮或卡片。
- `npm run typecheck`、`npm run build`、`npm run visual:guard` 通过。

### v2.2.3 - Archive Style Isolation

**Status:** Planned
**Scope:** Architecture / User-facing / Maintainability
**Goal:** 单独整理 `/archive` 样式系统，在复用全站 token 的同时保留它作为复杂应用界面的页面隔离层。

#### Focus

- archive 根页面 token
- markdown preview
- citation、session sidebar、message list
- 输入区、空状态、加载态和错误态
- 第三方组件深度选择器边界

#### Tasks

- [ ] 新增或整理 `app/assets/styles/pages/archive.less`。
- [ ] 将 `/archive` 页面核心视觉从长 scoped CSS 中迁移到页面 Less 入口。
- [ ] 清理 markdown preview、citation、session sidebar 中可替换的 `:deep(...)` 和 global 覆盖。
- [ ] 保留 archive 作为独立应用界面的页面层，但复用 `--gf-*` 全站 token。
- [ ] 为 archive 的空态、对话态、引用态和错误态补充视觉截图守卫。

#### Acceptance Criteria

- `/archive` 主题切换只依赖 `html.dark` 和 archive 页面 token。
- markdown、citation、session sidebar 的样式边界清晰，不污染其他页面。
- archive 页面在无数据、本地接口失败和普通对话状态下都可读。
- `npm run typecheck`、`npm run build`、`npm run perf:guard`、`npm run visual:guard` 通过。

### v2.2.4 - Legacy Selector Debt Cleanup

**Status:** Planned
**Scope:** Maintainability / Testing / Architecture
**Goal:** 清理迁移过程中遗留的选择器债务，确保全站样式主线没有隐藏的旧入口和高优先级补丁。

#### Focus

- 旧 dark class
- 可替换的 `:global(...)`
- 可替换的 `:deep(...)`
- 复杂 Tailwind 颜色组合
- 样式入口和命名一致性

#### Tasks

- [ ] 扫描并清理 app 目录中剩余可替换的页面级 `--dark` class。
- [ ] 扫描并清理可迁移到 Less 语义类的复杂 `dark:`、`bg-*`、`text-*`、`border-*` 组合。
- [ ] 审核 `app/assets/styles/index.less` 导入顺序，确保 token、mixin、components、pages 分层稳定。
- [ ] 更新 `docs/style-system.md` 的迁移清单，记录最终保留的例外项和原因。
- [ ] 扩展 `visual:guard` 的旧入口扫描列表，避免已删除模式回流。

#### Acceptance Criteria

- app 目录不再出现已废弃暗色入口。
- 保留的 `:global(...)` 和 `:deep(...)` 都有第三方组件或跨组件边界理由。
- 新增样式可以根据 `docs/style-system.md` 判断应该落在哪个 Less 入口。
- `npm run typecheck`、`npm run build`、`npm run perf:guard`、`npm run visual:guard` 通过。

### v2.2.5 - Style System Migration Complete

**Status:** Planned
**Scope:** Release / Documentation / Testing
**Goal:** 将样式系统迁移作为 `v2.2.x` 完整收口，后续版本只做增量维护，不再保留样式迁移遗留桶。

#### Focus

- 全站迁移验收
- 文档和截图基线
- 性能预算确认
- 发布前清理

#### Tasks

- [ ] 确认 games、games-search、home/nav、static、updates、archive 全部完成 Less + Tailwind 分层迁移。
- [ ] 确认新增样式必须优先使用 token 和语义类。
- [ ] 刷新并归档必要的视觉截图说明，不提交临时截图产物。
- [ ] 确认 `docs/performance/budget.json` 对迁移后的页面保持保守预算。
- [ ] 清理 roadmap 中不再需要的迁移阶段，后续只保留增量维护计划。

#### Acceptance Criteria

- `npm run typecheck` 通过。
- `npm run build` 通过。
- `npm run perf:guard -- --base-url http://localhost:3001` 在构建后的 preview 环境下通过。
- `npm run visual:guard -- --base-url http://localhost:3001` 覆盖核心页面亮暗色、中文/英文、桌面/移动端。
- roadmap 中不再存在独立的未来样式迁移遗留计划。

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

## 维护规则

- 新增视觉样式优先写 Less 语义类。
- 新增复杂颜色必须先进入 token。
- 不允许为了局部修复继续追加高优先级 global 覆盖。
- 每次迁移一个页面或一组组件，完成后清理对应旧样式入口。
- 后续版本不再承接样式迁移历史包袱，只用于新功能、体验增强或非样式架构工作。
