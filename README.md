<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26.7-00ADD8?logo=go&logoColor=white" alt="Go Version">&nbsp&nbsp
  <img src="https://img.shields.io/github/license/gofurry/gofurry-nav-site" alt="License">&nbsp&nbsp
  <img src="https://img.shields.io/badge/weekend-project-8B5CF6?style=flat" alt="Weekend Project">&nbsp&nbsp
  <img src="https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange" alt="Made with Love">
</p>

<p align="center">
  ⭐🐺 <a href="https://github.com/gofurry/gofurry-nav-site/README_en.md">English</a> | 
  <a href="https://go-furry.com">GoFurry 中国站</a> | 
  <a href="https://gofurry.com">GoFurry 国际站</a>
   🐺⭐
</P>

gofurry 是 GoFurry 站点体系的开源多服务仓库，面向兽圈文化内容发现、站点导航、兽游资料整理与用户视角的站点观测。

当前已上线的公开站点是 `https://go-furry.com`，定位为中国站。`https://gofurry.com` 是预留的国际站域名，已经注册，暂未上线。

线上主站当前使用 Nuxt 4 前台，重点提升公开页面的 SEO、首屏可见性、站点观测展示与兽游资料索引体验；原 Vue 前台保留为历史参考，不再作为新的生产前台入口。

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

- 中国站：`https://go-furry.com`，当前生产站点
- 国际站：`https://gofurry.com`，域名已注册，暂未上线
- 面向用户：希望查找兽圈网站、兽游资料、站点可用性与公开观测信息的访问者
- 面向开发者：希望了解、部署、维护或参与 GoFurry 的开发者与维护者
- 仓库形态：按服务拆分的 monorepo，不是单一可执行项目
- 产品边界：不提供用户账号系统，不代理用户访问，不收集用户敏感信息

## 仓库结构

- `apps/cn/nav-web`：Nuxt 4 前台，当前生产公开站点入口
- `apps/cn/nav-backend`：导航站后端 API，提供导航、站点详情、搜索建议、更新公告与监控端点
- `apps/cn/nav-collector`：导航数据采集服务，提供 observation、latest、summary、trend、change event 与低频旁路探测结果
- `apps/cn/game-backend`：兽游相关后端 API，承载游戏列表、详情、搜索、价格与在线人数等数据能力
- `apps/cn/game-collector`：兽游数据采集服务，负责 Steam 游戏资料、价格、在线人数与补充数据采集
- `apps/cn/admin`：运营后台，包含嵌入式前端、标签选项、采集触发、WAF 与管理接口
- `apps/cn/uptime`：独立极简状态服务，使用 Fiber uptime 与本地 Bbolt，不依赖业务 PostgreSQL/Redis
- `apps/intl`：国际站服务和前端占位目录，当前不参与生产构建
- `db`：`gfg`、`gfn`、`gfa` 的 Goose migration source of truth
- `docs`：跨服务架构、开发与部署说明
- `ops`：Nginx、维护页和 WAF 等运维资产
- `legacy`：已下线归档模块，包含旧 Vue 前台、原 RAG 服务、旧 ops agent / center，仅保留历史参考
- `third-party`：随仓库维护的第三方/自研依赖副本，例如 `monitor` 与 `steam-go`
- `experimental`：实验性代码，不参与正式发布
- `tools`：辅助脚本与本地工具

## 技术栈

- Go / Fiber
- PostgreSQL / Redis
- Nuxt 4 / Vue 3
- Tailwind CSS / Less
- ECharts
- Coraza WAF
- GoFurry monitor middleware
- steam-go

## 快速开始

仓库内各服务独立开发、独立运行。下面是当前最常用的入口。

前台开发：

```bash
cd apps/cn/nav-web
npm install
npm run dev
```

Go 服务开发：

```bash
cd apps/cn/nav-backend
cp conf/server.example.yaml conf/server.yaml
# 修改本地配置后：
go run . serve --config conf/server.yaml
```

常见后端入口：

```bash
cd apps/cn/game-backend
go run . serve --config conf/server.yaml

cd ../game-collector
go run . serve --config conf/server.yaml

cd ../nav-collector
go run . serve --config conf/server.yaml

cd ../admin
go run . serve --config config/server.yaml

cd ../uptime
go run . serve --config conf/server.yaml
```

六个 Go 程序的根命令只显示帮助，不会隐式启动；运行服务必须使用显式的 `serve --config`。示例配置只用于复制，真实密码和密钥不要提交。

如果你要构建根级 Go 服务产物，可按目标调用：

```bat
build.bat nav-backend
build.bat nav-collector
build.bat game-backend
build.bat game-collector
build.bat admin
build.bat uptime
```

这个脚本会把构建产物输出到根级 `build/` 目录。Nuxt 前台的生产部署使用 `apps/cn/nav-web` 目录内的独立流程。

## 生产部署

当前仓库内主要有两类生产部署路径。

Nuxt 前台使用独立部署路径，相关说明见：

- [apps/cn/nav-web/DEPLOYMENT.md](./apps/cn/nav-web/DEPLOYMENT.md)
- [apps/cn/nav-web/update.sh](./apps/cn/nav-web/update.sh)
- [跨服务部署说明](./docs/deployment.md)

当前生产更新的常用方式：

```bash
cd apps/cn/nav-web
./update.sh
```

六个 Go 服务使用二进制内置的 Linux/systemd 安装命令。安装必须从最终部署目录执行，并指定真实配置：

```bash
sudo ./gf-nav install --config /etc/gf-nav/server.yaml
# install 只 enable，不会启动；审核 unit 后手动启动：
sudo systemctl start gf-nav
```

完整迁移和卸载顺序见 [systemd 运维说明](./docs/operations/systemd.md)。数据库迁移与二进制部署相互独立，应用启动时不会执行 Goose；本次 V3 前仓库/运行时整理本身没有新增业务 DDL，不需要额外数据库迁移。`legacy/` 下的归档模块不参与默认构建和生产部署。

## 当前状态

- 中国站 `go-furry.com` 已上线并使用 Nuxt 4 前台
- 国际站 `gofurry.com` 已注册域名，暂未上线
- 前台正在进行去卡片化视觉改造，重点覆盖游戏页、游戏搜索页、站点详情页、首页导航面板与个人简历页
- `/steam` 相关页面已调整为“兽游工坊”方向，后续承载兽游工坊与 Steam 相关入口
- `gofurry-nav-backend` 公开主链路使用 `/api/v2/nav`，并接入自研 monitor 中间件展示请求状态
- `gofurry-nav-collector` 已完成 v2 数据面收口，提供 summary、latest、observations、trend、change event 与低频旁路探测结果
- `gofurry-game-backend` 与 `gofurry-game-collector` 继续承担兽游资料、价格、在线人数与采集队列能力
- `gofurry-admin` 已切到 GoFiber 官方 contrib Coraza 中间件，标签选择接口支持全量请求与展示
- `gf-uptime` 已从 Nav Backend 独立出来，使用本地 Bbolt 保存可用性历史；公开状态页默认使用 `https://status.go-furry.com`
- Nav Web 提供 `/healthz`，两个 Collector 可按内网配置启用 `/livez` 与 `/readyz`
- `gofurry-ops-agent` 与 `gofurry-ops-center` 已移入 `legacy/`，当前网站使用自研可观测性中间件路线
- `robots.txt`、`sitemap.xml`、`llms.txt` 与 `/.well-known/security.txt` 已作为公开站点元信息入口提供
- 旧 Vue 前台、原 RAG 服务和旧 Ops 服务均已归档到 `legacy/`

## 贡献说明

欢迎提交 Issue 和 Pull Request。

贡献时建议遵循以下原则：

- 尽量将改动限定在单个服务目录内
- 不提交 `.env`、私钥、数据库凭据或其他敏感配置
- 变更公开行为时，补充必要文档或部署说明
- 尊重现有服务边界，除非确实需要跨服务调整

## 开源与许可

本仓库采用 [MIT License](./LICENSE)。
