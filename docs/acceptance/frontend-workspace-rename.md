# Frontend workspace rename acceptance

日期：2026-10-18。基线：`dev` / `origin/dev` 均为
`dea6dcc119197e97a2efda65d50273528249d9c7`，实施前工作区干净。

## 范围与兼容性

- 798 个跟踪文件从 `apps/cn/nav-web` 移至 `apps/cn/web`，无兼容目录或 symlink。
  原目录只保留本机 ignored 私有文件/缓存，并通过本机 `.git/info/exclude` 继续忽略；未读取或迁移私有配置。
- 迁移文件中仅 `AGENTS.md`、`README.md`、`DEPLOYMENT.md`、`Dockerfile`、
  `docker-compose.prod.yml` 改变内容；其余 blob 与原始文件模式保持一致，
  包括业务源码、测试、Golden、lockfile 和 `update.sh`（`100755`）。
- 修复 Task、CI、doctor、检测/分组脚本、工程政策测试、Admin 平台图标 import 和有效文档路径。
  npm 包名、公共 Task 名、工作流文件/Job/check、Artifact/BuildKit cache 标识均保留。
  历史验收未重写；Changelog 和 alpha.9 release 仅修复指向现行部署说明的链接。
- Compose 固定 `name: nav-web`；服务/容器/镜像、`127.0.0.1:3000:3000` 和原运行时环境保持不变。
  `NODE_BUILD_HEAP_MB=4096` 仅由 Compose build.args 传入 Docker build 阶段；runner 仍从 deps 继承。
- 现有实现与冻结文档无重要路径契约冲突。旧 P0 审计 SHA 对应的目录在 contract 中保留说明；
  其余旧路径仅存在于历史记录、rename 检测 fixture 和维护者配置复制手册。

## 实际验证

工具：Go 1.26.9、Node 24.15.0、pnpm 12.6.0、Task 3.53.1；Docker Desktop Linux Engine 29.8.1 / Compose 5.5.1。
日志保存在 ignored `.cache/frontend-workspace-rename/`。

| 状态 | 命令 / 证据 |
| --- | --- |
| PASS | 根目录 `node --test .github/scripts/detect-changes.test.mjs .github/scripts/nav-web-regression-groups.test.mjs`：17/17。新路径、共享图标 Admin 触发、rename delete/add 事件与原分组完整性均保留。 |
| PASS | 设置 `GOTOOLCHAIN=go1.26.9` 后 `go -C tools test ./check-production-policy -count=1`；所有原守卫及新增部署身份/build-only heap 守卫通过。 |
| PASS | `task check:policy`、`node tools/doctor.mjs`；`task --dry build:nav-web-image`、`task --dry typecheck:nav-web`、`task --dry build:nav-web`。 |
| PASS | 新目录 `pnpm install --frozen-lockfile`、`pnpm run typecheck`、`pnpm run build`。Nitro 产物生成成功。 |
| PASS | Admin React `pnpm run typecheck` 后 `pnpm exec vite build`：按现有 build 脚本的两个步骤执行一次，跨目录图标 import 编译成功。未重复运行等价的 `pnpm run build`。 |
| PASS | 新目录 `docker compose -f docker-compose.prod.yml config -q`；另以公开 `.env.production.example` 渲染 JSON，核对项目 `nav-web`、原服务/镜像/容器/端口、heap build arg 和无 runtime heap 环境变量。 |
| PASS | Git rename/blob/mode 审核、有效旧路径与 Markdown 链接扫描、`git diff --check`；无私有 `.env` / `.env.next`、凭据、生成产物进入提交。 |
| FAIL（环境阻断） | 新目录 `docker build --progress=plain --tag gofurry-nav-web:workspace-rename-check .`：获取 `node:24-bookworm-slim` 的 Docker Hub 匿名 token 时连接超时，未进入 deps/build 阶段。未重试或改镜像来源。日志 `docker-build.log`。 |
| NOT RUN | 镜像构建成功及最终镜像运行时验证；需网络恢复后补验，静态守卫不能替代此项。全量 Browser/Visual、数据库/Go 集成、Goose、生产操作均不属于本次执行。 |

Nuxt 构建保留既有 sourcemap、依赖注解/exports 等警告，Admin 保留 chunk-size 警告；未借本次迁移调整构建策略。
远端 CI 由 push 触发，其结果独立于上述本地通过记录。

## 维护者后续操作

按 [部署说明](../../apps/cn/web/DEPLOYMENT.md#existing-installation-workspace-rename-maintainer-only)
在正式上线时单独备份/复制 `.env`、`.env.next`，确认渲染和实际容器项目名均为 `nav-web`，
检查内存限制，构建成功后再替换并检查 `/healthz`。内存不足时应在其他构建环境生成镜像。
本次未连接生产服务器，未复制配置、启动/替换生产容器或执行 `update.sh`。
