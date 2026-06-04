# GoFurry Nav v1 到 v2 迁移 Roadmap

## 总览

当前站点详情页相关接口已经基本完成 v2 化。剩余的 `/api/v1/nav` 接口主要服务导航首页、搜索建议、随机头图、随机金句、更新公告和 sitemap。后续迁移重点不是继续改站点详情页，而是把导航首页域的 v1 依赖逐步收敛到更稳定的 v2 读接口。

| 阶段 | 状态 | 目标 |
|---|---|---|
| 阶段 0：站点详情 v2 化 | 已完成 | 站点详情页、观测数据、趋势、变更、轻探针、浏览量触达走 v2 |
| 阶段 1：v1 依赖评估与迁移设计 | 已完成 | 明确剩余 v1 接口、前端调用点、性能问题和迁移顺序 |
| 阶段 2：首页 bootstrap v2 | 已完成 | 聚合首页首屏数据，减少 SSR 和浏览器请求数量 |
| 阶段 3：更新公告 v2 | 已完成 | 重构为双语轻量公告表和 v2 时间线接口，移除 CDN markdown 依赖 |
| 阶段 4：搜索建议 v2 | 已完成 | 搜索建议改走 v2，增加 IP 限流、短缓存、请求合并和前端取消机制 |
| 阶段 5：前端请求封装统一 | 未开始 | 收敛 `services/nav.ts` 与旧 axios 封装 |
| 阶段 6：v1 兼容与清理 | 未开始 | sitemap 迁移、幽灵接口清理、保留兼容窗口后下线 v1 |

状态说明：

- 已完成：代码已落地，前端已使用。
- 进行中：评估或设计已开始，但实现还未完全落地。
- 未开始：只完成方向判断，尚未进入实现。

## 阶段 0：站点详情 v2 化

状态：已完成

### 已完成范围

后端 v2 路由定义位置：`gofurry-nav-backend/routers/url_v2.go`

| 路由 | 状态 | 用途 |
|---|---|---|
| `GET /api/v2/nav/sites/:siteId/detail` | 已完成 | 站点详情页聚合数据 |
| `POST /api/v2/nav/sites/:siteId/view` | 已完成 | 站点详情页浏览量触达 |
| `GET /api/v2/nav/sites/:siteId/summary` | 已完成 | 站点健康摘要 |
| `GET /api/v2/nav/sites/:siteId/targets/:target/summary` | 已完成 | 目标健康摘要 |
| `GET /api/v2/nav/sites/:siteId/targets/:target/latest` | 已完成 | 目标最新观测 |
| `GET /api/v2/nav/sites/:siteId/targets/:target/observations` | 已完成 | 目标历史观测 |
| `GET /api/v2/nav/sites/:siteId/targets/:target/trend` | 已完成 | 目标趋势 |
| `GET /api/v2/nav/sites/:siteId/targets/:target/changes` | 已完成 | 目标变更事件 |
| `GET /api/v2/nav/sites/:siteId/targets/:target/light-probes` | 已完成 | 目标轻探针 |

前端已使用位置：

| 前端位置 | 状态 | 说明 |
|---|---|---|
| `app/composables/useSiteDetailPage.ts` | 已完成 | 使用 `/api/v2/nav/sites/:id/detail` |
| `app/components/site/SiteDetailPage.vue` | 已完成 | 使用 `/api/v2/nav/sites/:id/view` |
| `app/components/site/SitePerformance.vue` | 已完成 | 使用 v2 observations |
| `app/composables/useSiteMetadataProbePanel.ts` | 已完成 | 使用 v2 target latest、observations、changes 等接口 |

### 已具备的保护

- `payload_mode=preview` 默认截断 raw payload。
- `full_payload_enabled` 默认关闭。
- 单次响应 payload 有总预算。
- 历史观测列表有 limit 归一化和最大值限制。
- 详情聚合内部并发读取 summary、latest、light probe、trend、changes。

### 后续注意

站点详情 v2 已可作为其他 v2 迁移的风格参考，包括：

- 响应携带 `schema_version`。
- 对公开前端限制 payload 大小。
- 使用明确的 `state` 和 reason 字段表达缺失或不可用状态。
- 对 path 参数进行校验和归属检查。

## 阶段 1：v1 依赖评估与迁移设计

状态：已完成

完成记录：

- 评估并记录剩余 v1 路由、前端调用点、性能风险和迁移顺序。
- 产出本文档作为后续阶段推进依据。

### 剩余 v1 路由

后端 v1 路由定义位置：`gofurry-nav-backend/routers/url.go`

| 路由 | 当前用途 | 迁移建议 | 状态 |
|---|---|---|---|
| `GET /api/v1/nav/page/site/list` | 导航首页站点列表、sitemap | 首页首屏已迁移到 v2 home；sitemap 仍需迁移到 v2 sites index | 部分完成 |
| `GET /api/v1/nav/page/group/list` | 导航首页分组 | 首页首屏已迁移到 v2 home | 已完成 |
| `GET /api/v1/nav/page/ping/list` | 导航首页站点延迟、定时刷新 | 首页首屏和定时刷新已迁移到 v2 home/home-ping | 已完成 |
| `GET /api/v1/nav/page/search/:engine` | 搜索建议代理 | 前端已迁移到 `/api/v2/nav/search/suggestions`，v1 暂保留兼容 | 已迁移 |
| `GET /api/v1/nav/page/header/getSaying` | 首页随机金句 | 首页首屏已合并进 v2 home；客户端 fallback 仍可后续清理 | 部分完成 |
| `GET /api/v1/nav/page/header/image/url` | 首页随机背景图 | 首页首屏已合并进 v2 home；NavHeader fallback 仍可后续清理 | 部分完成 |
| `GET /api/v1/nav/site/changelog` | 更新公告列表 | 已由 `/api/v2/nav/updates` 和 `gfn_nav_update_notice` 替代 | 已迁移 |

### 前端 v1 使用点

主要位置：`gofurry-nav-web`

| 前端位置 | 使用接口 | 问题 | 状态 |
|---|---|---|---|
| `app/components/nav/NavHomePage.vue` | `getGroups`、`getSites`、`getPing`、`getSaying`、`getImageUrl` | 首页首屏并发多个 v1 请求 | 已迁移到 v2 home |
| `app/components/nav/NavContent.vue` | `getGroups`、`getSites`、`getPing` | 语言切换重新加载，且每 60 秒刷新 ping | 已迁移到 v2 home/home-ping |
| `app/components/nav/NavHeader.vue` | `getImageUrl` | 仍通过旧 axios 封装获取背景图 fallback | 待迁移 |
| `app/components/nav/NavTransitionBar.vue` | `getSaying` | 初始数据缺失时客户端补拉 | 待迁移 |
| `app/components/nav/SearchBox.vue` | `/api/v2/nav/search/suggestions` | 搜索建议已走 v2 service，支持更长防抖和旧请求取消 | 已迁移 |
| `app/pages/updates.vue` | `/api/v2/nav/updates` | 更新公告页已重构为 v2 时间线，移除 v1 changelog/content 依赖 | 已迁移 |
| `server/routes/sitemap.xml.ts` | `/api/v1/nav/page/site/list` | sitemap 依赖完整站点列表 | 待迁移 |

### 当前主要问题

| 编号 | 严重度 | 问题 | 建议阶段 |
|---|---|---|---|
| P2-001 | P2 | 首页 v1 接口拆分过细，首屏请求数量偏多 | 阶段 2，已完成 |
| P2-002 | P2 | 搜索建议缺少短缓存、请求合并和旧请求取消 | 阶段 4，已完成 |
| P2-003 | P2 | changelog 只依赖 Redis 缓存，miss 时容易失败 | 阶段 3，已通过重构关闭 |
| P3-001 | P3 | 前端存在 `services/nav.ts` 与旧 axios 两套 nav 请求封装 | 阶段 5 |
| P3-002 | P3 | `/nav/stat/add/count` 前端封装残留，但后端 v1 未注册 | 阶段 6 |

## 阶段 2：首页 bootstrap v2

状态：已完成

完成记录：

- 本地提交：`e824c62 feat(nav): add v2 home bootstrap`
- 后端新增 `GET /api/v2/nav/home` 和 `GET /api/v2/nav/home/ping`。
- 前端 `NavHomePage.vue` 首屏数据改为一次请求 v2 home。
- 前端 `NavContent.vue` 语言切换改走 v2 home，ping 定时刷新改走 v2 home-ping。
- Nuxt 新增 `/api/v2/nav/[...path]` 代理，浏览器端 v2 请求可在本地开发环境正常转发。

### 目标

新增首页启动聚合接口，减少 `/nav` 首屏请求数量，并统一首页数据的缓存状态和降级语义。

建议新增：

```txt
GET /api/v2/nav/home?lang=zh
```

### 建议响应结构

```json
{
  "schema_version": 2,
  "generated_at": "2026-06-03T00:00:00Z",
  "cache_state": {
    "sites": "ready",
    "groups": "ready",
    "ping": "ready",
    "saying": "ready",
    "backgrounds": "ready"
  },
  "sites": [],
  "groups": [],
  "ping": {},
  "saying": null,
  "backgrounds": {
    "desktop": "",
    "mobile": ""
  }
}
```

### 后端任务

- [x] 新增 home controller、service、models。
- [x] 复用现有 `GetSiteList`、`GetGroupList`、`GetPingList`、`GetSayingService`、`GetImageUrl`。
- [x] 为每个子数据块返回 `cache_state`，避免单点失败导致整页失败。
- [x] 保持 site/group 读取 Redis 优先，Redis miss 时回 DB。
- [x] 为首页聚合接口增加单元测试。

### 前端任务

- [x] `NavHomePage.vue` 改为一次请求 v2 home。
- [x] `NavContent.vue` 接收 v2 home 的初始数据。
- [x] 保留 ping 的轻量定时刷新，但改到 v2 接口。
- [x] 移除首屏中重复或 fallback 触发的背景图请求。

### 完成标准

- [x] `/nav` 首屏不再并发请求 v1 site/group/ping/saying/background。
- [x] 首页 SSR 数据加载只依赖 v2 home 聚合接口。
- [x] 首页语言切换仍能正确刷新站点和分组。
- [x] ping 定时刷新仍正常。

验证记录：

- `go test ./...`：通过。
- `go test -race ./apps/nav/...`：通过。
- `npm run typecheck`：通过。
- `npm run build`：通过，仅有既有 Tailwind sourcemap、Nitro external dependency、Node deprecation 类警告。

## 阶段 3：更新公告 v2

状态：已完成

### 目标

更新公告已从旧的 `gfn_log_update + CDN markdown URL + Redis changelog 缓存` 链路重构为新的轻量双语公告模型。前台只读取 v2 时间线接口，Admin 直接维护标题、发布时间和正文，RAG 不再同步更新公告内容。

已新增：

```txt
GET /api/v2/nav/updates?lang=zh
GET /api/v2/nav/updates?lang=en
```

### 后端任务

- [x] 新增 `updates` v2 controller、service 和 model。
- [x] 新增双语轻量公告表 `gfn_nav_update_notice`，字段包含 `title/title_en/body/body_en/published_at/create_time/update_time/deleted`。
- [x] SQL 放入仓库根目录 `sql/20260603_nav_update_notice_bilingual.sql`，按既有 SQL 文件命名方式管理。
- [x] v2 响应包含 `schema_version`、`generated_at`、`state`、`reason_messages` 和 `items`。
- [x] 废弃旧 `gfn_log_update` 读取路径，不再依赖 `site-common:changelog` Redis 缓存。
- [x] 移除 v1 changelog 路由和 Nuxt markdown 内容代理。
- [x] 增加有数据、空数据和 DB 错误场景测试，确认响应不再包含 markdown URL。

### 前端任务

- [x] `updates.vue` 改走 `/api/v2/nav/updates`。
- [x] 更新公告页重构为使用 `GoFurryGridBackground.vue` 的时间线视图。
- [x] 支持按年份折叠，默认展开最新年份，并提供加载更多。
- [x] 正文作为普通文本展示，保留换行，不再解析 markdown。
- [x] 中英文内容跟随站点语言读取。
- [x] 抽出更新页组件和日期工具，降低单文件维护压力。
- [x] 更新页空状态、失败状态、亮色和暗色样式已覆盖。

### Admin 和 RAG 任务

- [x] `gofurry-admin` 新增 `/api/v1/nav/update-notices` 管理接口和 `update-notices` 资源。
- [x] Admin 表单维护中英文标题、中英文正文和发布时间。
- [x] 后端校验标题、正文和发布时间，支持 `datetime-local`、RFC3339、`YYYY-MM-DD HH:mm:ss`。
- [x] `gofurry-rag` 移除 `site_changelog` 同步源、同步卡片、API union 类型和 markdown 拉取逻辑。
- [x] `all` 同步范围收敛为 nav sites、game details、game news、game creators。

### 完成标准

- [x] 更新页首次 SSR 不依赖 v1 changelog。
- [x] 更新页不再请求 `/api/v1/nav/site/changelog` 和 `/api/v1/nav/site/changelog/content`。
- [x] v2 更新公告返回结构具备 schema/version 信息。
- [x] 更新公告内容不再依赖 CDN markdown。
- [x] Admin 可维护新表数据。
- [x] RAG 不再出现或触发 `site_changelog` 同步。

验证记录：

- `gofurry-nav-backend`：`go test ./...` 通过。
- `gofurry-nav-web`：`npm run typecheck`、`npm run build` 通过。
- `gofurry-admin`：`go test ./...`、`cd web && npm run build` 通过。
- `gofurry-rag`：`go test ./...`、`cd web && npm run build` 通过。

## 阶段 4：搜索建议 v2

状态：已完成

### 目标

把搜索建议从实时第三方代理升级为更可控的 v2 接口，减少高频输入对后端和第三方搜索服务的放大效应，并限制单 IP 对本站 suggestions 接口的请求频率，避免本站出口 IP 因接口被盗刷而被搜索引擎封禁。

已新增：

```txt
GET /api/v2/nav/search/suggestions?engine=bing&q=keyword
```

### 后端任务

- [x] 定义 engine 白名单：`baidu`、`bing`、`google`、`bilibili`。
- [x] 复用现有第三方解析逻辑。
- [x] 为 `engine + normalized query` 加 Redis 短 TTL 缓存，当前 TTL 为 90 秒。
- [x] 使用 `singleflight` 合并相同 engine/query 的并发请求。
- [x] 第三方失败时返回空 suggestions 或 error state，不暴露上游细节给前端。
- [x] 保留现有 timeout、body limit、query length limit。
- [x] 增加按可信客户端 IP 的 v2 suggestions 专属限流：30 分钟最多 30 次，超限返回 HTTP 429 和 `Retry-After`。
- [x] 增加缓存命中、查询归一化、非法 engine、v2 envelope 和限流测试。

### 前端任务

- [x] `SearchBox.vue` 改用 `services/nav.ts`。
- [x] 请求使用 `AbortController`，输入变化时取消旧请求。
- [x] 防止旧响应晚于新响应返回后覆盖 suggestions。
- [x] 删除 `utils/api/nav.ts` 中搜索建议相关旧封装。
- [x] 防抖时间从 300ms 增加到 600ms，降低高频输入触发量。

### 完成标准

- [x] 搜索框不再走旧 axios nav 封装。
- [x] 后端同一 engine/query 的并发请求会被合并。
- [x] 第三方搜索服务异常时前端可平稳显示空建议。
- [x] 单 IP 30 分钟最多请求 30 次 v2 suggestions。
- [x] 前端输入变化不会让旧响应覆盖新 suggestions。

验证记录：

- `gofurry-nav-backend`：`go test ./apps/nav/... ./routers/...` 通过。
- `gofurry-nav-web`：`npm run typecheck` 通过。

## 阶段 5：前端请求封装统一

状态：未开始

### 目标

收敛 `gofurry-nav-web` 的 nav 请求入口，减少 baseURL、错误处理、SSR 行为不一致。

### 当前问题

- `app/services/nav.ts` 使用 Nuxt `useApi`。
- `app/utils/api/nav.ts` 使用旧 axios `createRequest`。
- `NavHeader.vue` 仍依赖旧封装。

### 任务

- [ ] 将 `NavHeader.vue` 背景图 fallback 改走 `services/nav.ts`。
- [x] 将 `SearchBox.vue` 搜索建议改走 `services/nav.ts`。
- [ ] 检查 `utils/api/nav.ts` 是否还有真实调用。
- [ ] 删除无调用的旧封装，或明确标记 legacy。
- [ ] 保证客户端和 SSR 的 baseURL 都来自 Nuxt runtime config。

### 完成标准

- nav 业务请求只保留一个主封装。
- `rg "utils/api/nav"` 不再出现新前端业务依赖。
- v1/v2 baseURL 配置路径清晰。

## 阶段 6：v1 兼容与清理

状态：未开始

### 目标

完成 v2 迁移后保留短期兼容，再逐步清理 v1 依赖和历史遗留封装。

### 任务

- [ ] 新增轻量站点索引接口供 sitemap 使用。
- [ ] `server/routes/sitemap.xml.ts` 改用 v2 站点索引接口。
- [ ] 删除或废弃 `/nav/stat/add/count` 前端封装。
- [ ] 如需要全站 PV，重新设计独立统计接口，不复用站点详情浏览量接口。
- [ ] 标记 v1 route deprecated。
- [ ] 为 v1 下线准备兼容期和回滚方案。

### 完成标准

- `gofurry-nav-web` 不再直接依赖 `/api/v1/nav/page/*`。
- sitemap 不再依赖完整首页站点列表。
- v1 下线前有明确兼容窗口和监控指标。

## 推荐接口拆分

如果不希望 `home` 响应过大，可以采用聚合 + 轻量索引组合：

```txt
GET /api/v2/nav/home?lang=zh
GET /api/v2/nav/home/ping
GET /api/v2/nav/sites/index?lang=zh
GET /api/v2/nav/search/suggestions?engine=&q=
GET /api/v2/nav/updates?lang=zh
```

其中 `sites/index` 可服务 sitemap，只返回 `id`、`domains`、`updated_at` 等生成 URL 所需字段。

## 验证建议

后端：

```bash
cd gofurry-nav-backend
go test ./...
go test -race ./apps/nav/...
```

前端：

```bash
cd gofurry-nav-web
npm run typecheck
npm run build
```

性能验证：

- 对比迁移前后首页 SSR 请求数量。
- 对比 `/nav` 首屏 TTFB 和浏览器瀑布图。
- 压测搜索建议接口，观察第三方超时、缓存命中率和后端 goroutine 数量。
- 检查更新公告页是否只请求 `/api/v2/nav/updates`，并确认不再触发 v1 changelog/content。

## 当前下一步建议

阶段 4 已完成。下一步建议进入阶段 5：继续收敛 `services/nav.ts` 与旧 axios 封装，优先处理 `NavHeader.vue` 的背景图 fallback 和剩余 legacy nav 请求。
