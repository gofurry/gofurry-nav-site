<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white" alt="Go Version">&nbsp&nbsp
  <img src="https://img.shields.io/github/license/gofurry/gofurry-nav-site" alt="License">&nbsp&nbsp
  <img src="https://img.shields.io/badge/weekend-project-8B5CF6?style=flat" alt="Weekend Project">&nbsp&nbsp
  <img src="https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange" alt="Made with Love">
</p>

<p align="center">
  ⭐🐺 <a href="https://github.com/gofurry/gofurry-nav-site/README_en.md">English</a> 🐺⭐
</P>

gofurry 是一个面向兽圈文化内容发现与整理的开源多服务仓库，包含公开站点前台、导航与游戏数据接口、采集服务、RAG 服务，以及运营与周边能力模块。

线上主站当前已迁移到 Nuxt 4 前台，重点提升公开页面的 SEO 与首屏可见性；原 Vue 前台保留为迁移参考，不再作为新的生产前台入口。

```text
                  ░██████             ░██████████                                        
                ░██   ░██            ░██                                                
                ░██         ░███████  ░██        ░██    ░██ ░██░████ ░██░████ ░██    ░██ 
                ░██  █████ ░██    ░██ ░█████████ ░██    ░██ ░███     ░███     ░██    ░██ 
                ░██     ██ ░██    ░██ ░██        ░██    ░██ ░██      ░██      ░██    ░██ 
                ░██   ░███ ░██    ░██ ░██        ░██   ░███ ░██      ░██      ░██   ░███ 
                  ░█████░█  ░███████  ░██         ░█████░██ ░██      ░██       ░█████░██ 
                                                                                    ░██ 
                                                                              ░███████
```

## 项目定位

- 公开网站：`https://go-furry.com`
- 面向对象：希望了解、部署、维护或参与 gofurry 的开发者与维护者
- 仓库形态：按服务拆分的 monorepo，而不是单一可执行项目

## 仓库结构

- `gofurry-nav-web`：Nuxt 4 前台，当前生产环境公开站点入口
- `gofurry-nav-frontend-legacy`：旧版 Vue 前台归档代码，仅保留作迁移对照
- `gofurry-nav-backend`：导航站后端 API
- `gofurry-nav-collector`：导航数据采集服务，collector v2 数据面已完成，提供 observation、latest、summary、trend、change event 与低频旁路探测结果
- `gofurry-game-backend`：兽游相关后端 API
- `gofurry-game-collector`：兽游数据采集服务
- `gofurry-rag`：轻量 RAG 服务，包含嵌入式管理控制台、公开问答接口与内容同步
- `gofurry-admin`：运营后台，包含嵌入式前端
- `gofurry-intl`：国际化相关服务与前端实验目录
- `ops`：轻量运维探针与监控中心，包含 `gofurry-ops-agent`、`gofurry-ops-center`、审计模板与路线图
- `sql`：数据库相关脚本与结构文件
- `experimental`：实验性代码，不参与正式发布
- `tools`：辅助脚本与本地工具

## 技术栈

- Go
- Fiber
- PostgreSQL
- Redis
- Nuxt 4 / Vue 3
- Tailwind CSS

## 快速开始

仓库内各服务独立开发、独立运行。下面是当前最常见的几个入口。

前台开发：

```bash
cd gofurry-nav-web
npm install
npm run dev
```

Go 服务开发：

```bash
cd gofurry-nav-backend
go run .
```

RAG 服务开发：

```bash
cd gofurry-rag
go run . --config ./config/server.yaml serve
```

Ops 本地体验：

```bash
cd ops/gofurry-ops-center
go run ./cmd/center check-config --config ./configs/center.example.yaml

cd ../gofurry-ops-agent
go run ./cmd/agent --config ./configs/agent.example.yaml check-config
```

如果你要构建根级打包产物，可使用：

```bat
build.bat all
```

当前支持的目标包括：

- `gofurry-nav-backend`
- `gofurry-nav-collector`
- `gofurry-game-backend`
- `gofurry-game-collector`
- `gofurry-admin`
- `gofurry-rag`
- `gofurry-ops-agent`
- `gofurry-ops-center`

这个脚本会把构建产物输出到根级 `build/` 目录。Nuxt 前台的生产部署则使用 `gofurry-nav-web` 目录内的 Docker 化流程。

## 生产部署

当前仓库内主要有两类部署路径。

Nuxt 前台使用独立 Docker 部署路径，相关说明见：

- [gofurry-nav-web/DEPLOYMENT.md](./gofurry-nav-web/DEPLOYMENT.md)
- [gofurry-nav-web/update.sh](./gofurry-nav-web/update.sh)

当前生产更新的常用方式：

```bash
cd gofurry-nav-web
./update.sh
```

Go 服务与 RAG 则延续各自目录内的二进制 / install 路线，建议分别查看对应子项目文档，尤其是：

- [gofurry-rag/README_zh.md](./gofurry-rag/README_zh.md)
- [gofurry-rag/docs/deployment.md](./gofurry-rag/docs/deployment.md)

Ops Agent / Center 提供轻量 systemd、Docker Compose 与 Nginx 示例，适合用独立 PostgreSQL schema 或独立数据库部署：

- [ops/gofurry-ops-agent/README.md](./ops/gofurry-ops-agent/README.md)
- [ops/gofurry-ops-center/README.md](./ops/gofurry-ops-center/README.md)
- [ops/roadmap.md](./ops/roadmap.md)

## 当前状态

- 主站前台已迁移到 Nuxt 4，并已在生产环境使用
- `gofurry-nav-collector` 已完成 v2 数据面收口，具备进入 `gofurry-nav-backend /api/v2/nav` 后端接口阶段的条件
- `archive` 自由问答页与站内 RAG 能力已经接入生产链路
- `gofurry-rag` 已包含公开问答边界、轻量多轮上下文与站内同步框架
- `ops/gofurry-ops-agent` 与 `ops/gofurry-ops-center` 已形成轻量监控闭环，覆盖节点采集、服务健康、告警状态、Peer 摘要、同步/部署事件和嵌入式控制台
- `ops/code-audit.md` 保持为下一次 Go 代码审计模板，`ops/roadmap.md` 记录 Ops 后续演进方向
- `robots.txt` 与 `sitemap.xml` 由 Nuxt server routes 动态生成
- 旧 Vue 前台已归档为 legacy 参考实现
- 根级 `build.bat` 已覆盖核心 Go 服务、管理后台、RAG 与 Ops Agent / Center 二进制产物

## 贡献说明

欢迎提交 Issue 和 Pull Request。

贡献时建议遵循以下原则：

- 尽量将改动限定在单个服务目录内
- 不提交 `.env`、私钥、数据库凭据或其他敏感配置
- 变更公开行为时，补充必要文档或部署说明
- 尊重现有服务边界，除非确实需要跨服务调整

## 开源与许可

本仓库采用 [MIT License](./LICENSE)。
