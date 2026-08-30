# GoFurry Admin

[English](README.md)

GoFurry Admin 是中国站 active 运维后台。前端以 Vue 构建后嵌入 Go 二进制；后端只支持 PostgreSQL：`gfa` 保存后台认证/审计状态，`gfn` 与 `gfg` 使用显式连接池操作 Nav 和 Game 业务数据，Redis 保持现有运行语义。

数据库 schema 只由仓库根目录的 Goose migrations 管理，Admin 启动时不会建表或执行迁移。

Admin 认证为数据库校验的多账号系统，固定使用 `owner`、`developer`、`operator` 三种角色。Cookie JWT 只携带账号与会话身份，每个请求都会重新读取当前启用账号并按编译期 capability 策略授权。详见 [Admin 身份与授权说明](../../../docs/admin-identity.md)。

“采集中心”直接通过现有 `gfg` / `gfn` 连接池管理 durable Schedule / Job / Run / Result / Collector Instance，支持计划启停与 Run Now、Game/Nav 手工采集、队列与历史、取消、受约束重试、审计以及 ECharts outcome/coverage/timing。Admin 不代理调用 Game/Nav Backend；Admin 停机不会停止 Collector 自主调度和 worker。

## 本地开发

需要 Go 1.26.7、Node.js/npm、PostgreSQL 和 Redis。

~~~bash
cd web
npm ci
npm run build
cd ..

cp config/server.example.yaml config/server.yaml
# 修改被 Git 忽略的本地配置后再启动。
go run . serve --config config/server.yaml
~~~

其他命令：

~~~bash
go run . --help
go run . version
go run . reset-password --config config/server.yaml --username owner --password '<新密码>'
~~~

根命令只显示帮助。`serve` 在前台运行，收到 SIGINT/SIGTERM 后会关闭 Fiber、Redis、三个 PostgreSQL 连接池和日志。

## 生产构建与 systemd

根目录 `build.bat admin` 会先构建前端，再构建 Linux 二进制。必须使用最终部署位置的二进制，并从预期工作目录执行安装：

~~~bash
cd /srv/gofurry/gofurry-admin
sudo ./gofurry-admin install --config /etc/gofurry-admin/server.yaml
sudo systemctl cat gofurry-admin
sudo systemctl start gofurry-admin
~~~

安装只写入并 enable `gofurry-admin.service`，不会启动服务。运行用户取 `SUDO_USER`；已有 unit 默认拒绝覆盖，只有显式 `--force` 才会替换。

~~~bash
sudo ./gofurry-admin uninstall
~~~

卸载只删除 systemd 注册，不删除二进制、配置、日志、数据库、Redis 数据或工作目录。完整步骤见[统一 systemd 运维说明](../../../docs/operations/systemd.md)。

## 验证

~~~bash
go vet ./...
go test ./...
go build ./...

cd web
npm ci
npm run build
~~~
