![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/github/license/GoFurry/gofurry-nav-site)
![Weekend Project](https://img.shields.io/badge/weekend-project-8B5CF6?style=flat)
![Made with Love](https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange)

[English](./README_en.md)

GoFurry 是一个面向兽圈文化内容发现与整理的开源多服务仓库，包含公开站点前台、导航与游戏数据接口、采集服务，以及运营后台。

线上主站当前已迁移到 Nuxt 4 前台，重点提升公开页面的 SEO 与首屏可见性；原 Vue 前台保留为迁移参考，不再作为新的生产前台入口。

```text
  ░██████             ░██████████                                        
 ░██   ░██            ░██                                                
░██         ░███████  ░██        ░██    ░██ ░██░████ ░██░████ ░██    ░██ 
░██  █████ ░██    ░██ ░█████████ ░██    ░██ ░███     ░███     ░██    ░██ 
░██     ██ ░██    ░██ ░██        ░██    ░██ ░██      ░██      ░██    ░██ 
 ░██  ░███ ░██    ░██ ░██        ░██   ░███ ░██      ░██      ░██   ░███ 
  ░█████░█  ░███████  ░██         ░█████░██ ░██      ░██       ░█████░██ 
                                                                     ░██ 
                                                               ░███████
```

## 项目定位

- 公开网站：`https://go-furry.com`
- 面向对象：希望了解、部署、维护或参与 GoFurry 的开发者与维护者
- 仓库形态：按服务拆分的 monorepo，而不是单一可执行项目

## 仓库结构

- `gofurry-nav-web`：Nuxt 4 前台，当前生产环境公开站点入口
- `gofurry-nav-frontend-legacy`：旧版 Vue 前台归档代码，仅保留作迁移对照
- `gofurry-nav-backend`：导航站后端 API
- `gofurry-nav-collector`：导航数据采集服务
- `gofurry-game-backend`：兽游相关后端 API
- `gofurry-game-collector`：兽游数据采集服务
- `gofurry-admin`：运营后台，包含嵌入式前端
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

仓库内各服务独立开发、独立运行。最常见的两类入口如下。

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

如果你要构建旧的根级打包产物，可使用：

```bat
build.bat all
```

这个脚本主要用于当前仓库里的传统构建产物输出到 `build/` 目录。Nuxt 前台的生产部署则使用 `gofurry-nav-web` 目录内的 Docker 化流程。

## 生产部署

Nuxt 前台使用独立 Docker 部署路径，相关说明见：

- [gofurry-nav-web/DEPLOYMENT.md](./gofurry-nav-web/DEPLOYMENT.md)
- [gofurry-nav-web/update.sh](./gofurry-nav-web/update.sh)

当前生产更新的常用方式：

```bash
cd gofurry-nav-web
./update.sh
```

## 当前状态

- 主站前台已迁移到 Nuxt 4，并已在生产环境使用
- `robots.txt` 与 `sitemap.xml` 由 Nuxt server routes 动态生成
- 旧 Vue 前台已归档为 legacy 参考实现
- 基础 GitHub Actions 检查已覆盖核心 Go 服务、管理后台与旧前台路径

## 贡献说明

欢迎提交 Issue 和 Pull Request。

贡献时建议遵循以下原则：

- 尽量将改动限定在单个服务目录内
- 不提交 `.env`、私钥、数据库凭据或其他敏感配置
- 变更公开行为时，补充必要文档或部署说明
- 尊重现有服务边界，除非确实需要跨服务调整

## 开源与许可

本仓库采用 [MIT License](./LICENSE)。
