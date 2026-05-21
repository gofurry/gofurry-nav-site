# Fiber v3 Medium Template

[English](./README.md)

`medium` 是偏均衡的 HTTP 服务模板。它保留了 Redis、WAF、service 安装卸载、embedded UI 和较完整的中间件支持，但去掉了更偏平台化的运行时复杂度。

## 这个版本包含什么

- 默认 SQLite 开箱即用
- 内置完整的 `user` CRUD 示例
- 保留 DB、Redis、logging、WAF、graceful shutdown
- 保留 `kardianos/service` 的 `install` / `uninstall`
- 保留 `pkg/httpkit`、`pkg/abstract`
- 保留较完整的 Fiber 中间件基线：request ID、access log、recover、CORS、timeout、health probes、security headers、compression、ETag、rate limiting
- 按需启用 CSRF、Swagger、WAF、pprof、embedded UI

## 适用场景

- 希望模板更偏真实业务服务
- 不想引入额外的平台运行时负担
- 想要完整中间件和部分增强能力，但业务层仍保持简单

## 快速开始

```bash
go run . serve
```

查看版本：

```bash
go run . version
```

安装或卸载服务：

```bash
go run . install
go run . uninstall
```

## 默认端点

- `GET /healthz`
- `GET /livez`
- `GET /readyz`
- `GET /startupz`
- `GET /api/v1/user/`
- `POST /api/v1/user/`
- `GET /api/v1/user/:id`
- `PUT /api/v1/user/:id`
- `DELETE /api/v1/user/:id`

按需开启：

- `GET /csrf/token`
- `GET /swagger`
- `GET /debug/pprof/...`

## 业务组织方式

- 业务代码位于 `internal/app/<domain>`
- 正常使用 `controller`、`dao`、`service`、`models`
- 路由统一注册在 `internal/transport/http/router/url.go`
- 数据模型注册集中在 `internal/bootstrap/lifecycle.go`

## 配置概览

主配置文件：

```bash
./config/server.yaml
```

重点配置块：

- `server`
- `database`
- `redis`
- `log`
- `middleware`
- `waf`

## 取舍说明

- 比 `heavy` 更轻，但仍然偏工程化
- 比 `light` 保留更多运行时能力和增强中间件
- 适合作为大多数 HTTP 服务的均衡模板

## 使用前检查

- 替换 `go.mod` 模块路径
- 修改 `config/server.yaml` 里的应用标识
- 如果不需要 demo，删除内置 `user` 示例
- 在 `internal/app` 下添加自己的业务域
