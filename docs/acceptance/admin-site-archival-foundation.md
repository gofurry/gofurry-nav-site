# Admin Site 归档基础验收

日期：2026-10-18。起始本地/远端 `dev` 均为 `8600c623aa1f0cae5cc203ea720190c0f12669fe`，工作区干净；未切分支、pull、reset、rebase 或 stash。按冻结 V1 文档的实际函数/查询定位，旧行号不作为实现依据。

## 实际范围与契约

- 新迁移 `20261018020000_nav_site_archival_metadata.sql`：两个 nullable text 字段、六种原因代码 CHECK、1000 字符 CHECK、活跃 Site 不得带归档元数据、中文 COMMENT。不回填旧归档原因；Down 明确拒绝丢弃历史。未修改 `20261018010000` 或其他旧迁移。
- `DELETE /api/v1/nav/sites/:id` 必须提交 `deletion_reason_code`，只接受 `unreachable / closed / migrated / duplicate / owner_request / other`；`deletion_reason` 可空，trim 后最多 1000 个 Unicode 字符。缺失、非法、超长、重复归档返回 400，不覆盖首次记录。复用原 Site 锁 → 关闭 Primary/Target 周期 → 软删除 → 刷新当日 Site Daily → Audit → commit 路径；周期原因仍为 `site_deleted`。
- Site 内容、Target 创建/编辑/转移/删除、Primary、分组关系及精选写入拒绝归档 Site；Target 先按 Site ID 顺序加锁，再锁 Domain，避免与归档形成反向锁顺序。采集入队、图标写入复用原 active Site 守卫；content 权限、CSRF 和通用 DELETE helper 未改。
- `site-summaries?status=active|archived` 默认 active，非法值 400；独立 count/list 使用相同过滤，支持名称、ID、简介及历史 Target 域名搜索。active 保持原排序，archived 按 `deleted_at DESC,id DESC`；原主域名来自最后关闭的 Primary 周期。普通 Site GET/List 保持 active 过滤，仅 Admin Workspace 可读取归档记录。
- React 复用 DataTable toolbar、Select、RHF/Zod、现有 Dialog/CSRF transport。状态进入 URL/query key，切换重置页码。归档详情显示原因/时间/说明、原内容/图标、原主域名、历史配置；不挂载编辑、归档、采集、Primary 或图标写控件。旧 NULL 原因显示「未记录（历史归档）」。实际 HistoryPanel 读取 Changes，已有能力为 `changes.read`；冻结文档关于 `audit.read` 的条件说明不改变此消费者。
- 已核对正式 DELETE 消费者为 Site Workspace；两个既有集成测试调用和一个显式 opt-in 的资产验收清理调用同步提供原因。SQLC 正常生成；Admin DataOps 清单同步。Nav Backend/Collector 仅自动生成模型增加字段，无行为修改。

## 验证结果

Go 使用 `GOTOOLCHAIN=go1.26.9`；React 使用 Node 24.15.0 / 项目固定 pnpm 12.6.0。数据库仅为本机 `postgres:18.6` 临时容器、随机 loopback 端口、tmpfs；测试创建独立 GFA/GFN 并清理，无共享/生产数据库访问或真实站点归档。

| 状态 | 实际命令 / 证据 |
|---|---|
| PASS | `tools/`: `go tool sqlc generate -f ../sqlc.yaml`；`go tool sqlc vet -f ../sqlc.yaml`；`go run ./check-sqlc`，生成文件与查询一致 |
| PASS | `tools/`: 设置隔离 `GOFURRY_TEST_POSTGRES_ADMIN_URL`、`GOFURRY_REQUIRE_POSTGRES_MAJOR=18`，以 `GOFURRY_UPDATE_SCHEMA_SNAPSHOTS=1 go test ./db-baseline -run '^TestPostgresFreshAndBaselineAdoption$/^gfn$' -count=1 -v` 生成 expected-final；未启用 baseline 更新 |
| PASS | 取消快照更新后：`go test ./db-baseline -run '^TestPostgresSiteArchivalUpgrade$\|^TestPostgresFreshAndBaselineAdoption$/^gfn$\|^TestExpectedFinalReadabilityPolicy$' -count=1 -v`；fresh、adoption、旧版本升级、真实 schema 对比与中文可读性均通过 |
| PASS | 新迁移定向升级验证：旧 Site 完整内容/时间保留，NULL 不回填，非法 code、active 元数据和超长说明被 DB 拒绝，1000 个 emoji 合法，破坏性 Down 拒绝且版本不变 |
| PASS | `apps/cn/admin/`: 使用仅指向上述临时容器的 `GOFURRY_ADMIN_INTEGRATION_CONFIG`，`go test ./internal/bootstrap -run '^TestAdminSiteArchival$' -count=1 -v` |
| PASS | 真实 Fiber + GFA/GFN：非法 body 零写入；Audit 失败回滚 Site 与周期；重复归档不改首次元数据；Observation、Target/Site Facts、Metric、Change、CollectorDomain 内容保持；活跃周期归零；历史域名/count/分页/原主域名正确；默认 active 与归档只读、权限、入队拒绝、分类/精选保留均通过 |
| PASS | `apps/cn/admin/`: `go test ./internal/app/navadmin/controller -run 'TestNormalizeSiteArchival\|TestValidateSite\|TestSiteGroupMap\|Test.*Asset\|Test.*Icon\|Test.*Curation' -count=1`；`go test ./internal/app/navadmin/controller ./internal/app/dataops -run 'TestValidateSite\|TestExpectedRepositoryMigrationVersions' -count=1` |
| PASS | `apps/cn/admin/`: `go vet ./internal/app/navadmin/... ./internal/db/nav/sqlc ./internal/bootstrap`；Nav Backend、Nav Collector 各自 `go test ./internal/db/nav/sqlc -run '^$'` 编译受影响生成包 |
| PASS | React 最终构建后，Admin `go test ./internal/transport/http/router -run '^TestEmbeddedReactSPAAndAPIBoundary$' -count=1`，内嵌 SPA/静态资源与 API 边界通过 |
| PASS | `tools/`: `go run ./check-production-policy`；`git diff --check` |
| PASS | `apps/cn/admin/react/`: `pnpm exec vitest run src/features/sites/site-pages.test.tsx src/features/sites/group-curation-page.test.tsx src/lib/api.test.ts`，**3 文件 / 12 项通过**；其中 Site 6 项覆盖只读页面、旧 NULL、History 权限、显式原因、错误保留草稿、成功刷新、超长拒绝、取消清空、列表 URL/分页与只读账号 |
| PASS | React `pnpm run typecheck`；`pnpm exec oxlint src/features/sites/site-pages.tsx src/features/sites/site-pages.test.tsx src/lib/types.ts`；`pnpm run build`。保留 2 条既有 `form.watch` React Compiler 提示及 Vite >500KB chunk 提示，无 error，未修改规则 |
| NOT RUN | 维护者真实浏览器人工交互/视觉验收、真实 COS 图标上传、共享/生产 Goose、生产 Collector/部署、全仓库 Go/Game 测试、Nav Web Browser/Full/Visual、远端 CI 结果等待 |

开发中曾出现新增测试夹具遗漏 GFA 账号和 Audit text 字段未显式转 jsonb 的失败；已修正测试夹具/查询，最终上述命令通过。旧直接调用 `deleteSite` 的测试签名也已同步；未降低断言。

## 已知边界与独立上线操作

- GFA Audit 与 GFN 保留既有跨库边界：Audit 失败阻止 GFN commit，但不能保证分布式原子提交。已有普通 GetSite/图标缺失资源错误映射仍为原 500，本批不重构公共错误体系。
- Collector 查询继续排除 deleted Site，已领取且持有目标快照的任务可能完成；归档不等于中断进行中的网络请求。公开 Home/分类/精选/Detail 的既有删除过滤及缓存刷新语义未改，未承诺立即清空所有派生缓存。
- 人工待验：在隔离环境选择原因/取消/确认；active↔archived 与历史域名搜索、翻页刷新；归档旧 NULL 与新原因详情的中文/图标/历史可读性，确认各 Tab 无写入口。自动化夹具通过不替代维护者人工批准。
- 生产需另行批准维护窗口、备份及 Goose 操作，按顺序应用尚未上线的 `20261018010000` 和本次 `20261018020000`，与后续另行批准的迁移协调；Admin Go + 内嵌 React 必须匹配新 schema 后部署。迁移本身不归档任何站点，实际归档仍须操作者逐项明确选择原因。不执行恢复、公开网络遗产陈列馆或 #92-C。
