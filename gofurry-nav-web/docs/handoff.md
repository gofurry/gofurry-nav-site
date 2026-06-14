# GoFurry Nav Web 样式系统维护 Handoff

## 当前状态

`gofurry-nav-web` 的 `v2.2.x` 样式系统迁移已经完成。后续工作不再从“迁移历史债”出发，而是按维护规则增量演进。

样式主线为：

> Less 负责主题 token、语义组件、复杂视觉；Tailwind 负责布局、间距、尺寸、响应式和简单显示控制。

权威文档：

- `docs/style-system.md`：命名、分层、保留例外和视觉守卫范围。
- `docs/performance/README.md`：性能守卫、视觉守卫、报告产物规则。
- `docs/roadmap.md`：迁移完成态和后续维护规则。

## 开始任何样式工作前

先确认仓库状态：

```powershell
git status --short
```

如果存在未提交改动，先阅读 diff 并确认它们是否属于当前任务；不要把用户或其他任务的改动混进样式维护提交。

## 样式入口

- Tailwind 入口：`app/assets/css/main.css`
- Less 全局入口：`app/assets/styles/index.less`
- 全站 token：`app/assets/styles/tokens.less`
- Less mixin：`app/assets/styles/mixins.less`
- 公共组件：`app/assets/styles/components/`
- 页面样式：`app/assets/styles/pages/`

`index.less` 的导入顺序必须保持为 token、mixin、公共组件、页面样式。页面样式可以消费公共 token 和组件类；公共组件样式不能反向依赖页面类。

## 已完成范围

- 首页导航：`pages/nav.less`
- 游戏首页：`pages/games.less`
- 游戏搜索：`pages/games-search.less`
- 游戏抽奖：`pages/lottery.less`
- 静态页面：`pages/static.less`
- 更新公告：`pages/updates.less`
- 知识库归档：`pages/archive.less`
- 公共组件：button、card、chip、footer、input、modal、nav、pagination、rating、shell

## 维护约束

- 暗色入口统一使用 `html.dark`。
- 不新增 `.games-page--dark`、`.nav-home-page--dark`、`.gf-static-page--dark` 等页面级暗色入口。
- 不新增 `:global(.dark ...)`；确需跨组件暗色选择器时写 `:global(html.dark ...)`，并限制在组件根类边界内。
- 不新增未登记的 `:deep(...)`；确需保留时先写入 `docs/style-system.md` 的保留例外表。
- 模板里不要用复杂 `bg-*`、`text-*`、`border-*`、`hover:*`、`dark:*` 拼完整组件状态；应沉淀为 token 和语义类。
- 页面专属视觉进入 `pages/*.less`，公共复用视觉进入 `components/*.less`。

## 当前保留例外

少量站点详情组件仍保留 `:global(html.dark ...)` 或 `:deep(...)`，原因和边界记录在 `docs/style-system.md`。这些例外禁止扩散；如果未来迁移站点详情页，应单独建立页面 Less 边界并同步收紧守卫。

## 验证建议

常规样式维护完成后运行：

```powershell
cd gofurry-nav-web
npm run typecheck
npm run build
```

涉及页面视觉、主题入口或守卫规则时运行完整回归：

```powershell
npm run preview -- --port 3001
npm run perf:guard -- --base-url http://localhost:3001
npm run visual:guard -- --base-url http://localhost:3001
```

`docs/performance/reports/` 下的截图、manifest、trace 和性能报告是本地临时产物，默认被忽略，不应随提交进入仓库。
