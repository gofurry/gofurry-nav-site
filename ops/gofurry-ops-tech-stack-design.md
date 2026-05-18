# GoFurry 轻量运维系统技术选型与设计方案

> 适用项目：`gofurry-nav-site/ops`  
> 当前目录：
>
> ```text
> ops/
> ├── gofurry-ops-agent
> └── gofurry-ops-center
> ```
>
> 目标：替代原先较重的 `Prometheus + Grafana + 自定义 Prometheus 数据看板` 方案，为 GoFurry 中国站与国际站提供一套更轻、更可控、更适合单人维护的运维观测系统。

---

## 1. 背景与目标

GoFurry 当前有中国站和规划中的国际站：

```text
中国站：go-furry.com
- 腾讯云业务服务器 A：4c8g
- 腾讯云运维服务器 B：2c4g
- Nginx 反代 + 腾讯云 EdgeOne
- 全量业务后端、数据库、采集器、管理后台、运维控制台

国际站：gofurry.com
- Cloudflare + 海外 VPS
- 独立定制化前端
- Go 单体后端
- 精简 PostgreSQL / Redis
- 每日同步中国站公开精简数据
```

原先使用 Prometheus + Grafana + 自定义看板，虽然专业，但对于单人维护的个人站来说偏重：

```text
组件多
资源占用高
配置复杂
告警链路重
看板维护成本高
很多指标看起来专业，但实际决策价值有限
```

新的运维系统目标是：

```text
轻量
低资源占用
低维护成本
跨中国站与国际站
可自定义 GoFurry 业务状态
能快速回答“服务是否正常”
先不做主动告警，重点做观测和状态看板
```

---

## 2. 总体设计原则

### 2.1 不复刻 Prometheus

本系统不是为了重新实现一个通用监控平台，而是服务 GoFurry 自身的轻量运维系统。

它主要回答以下问题：

```text
服务器还活着吗？
CPU / 内存 / 磁盘是否异常？
Docker 容器是否正常？
Nginx / API / Nuxt 是否可访问？
PostgreSQL / Redis 是否正常？
HTTPS 证书是否快过期？
中国站到国际站的数据同步是否成功？
最近一次部署是否正常？
国际站广告位、关键页面是否可访问？
```

### 2.2 Agent 主动上报

采用 Agent 主动 Push 模式，而不是 Center 主动 Pull。

```text
Agent -> Center
```

原因：

```text
服务器不需要额外暴露监控端口
跨区域网络更容易处理
防火墙规则简单
国际站和中国站可以各自自治
Agent 离线时可以本地缓存后补报
```

### 2.3 双 Center，弱一致

由于中国站与国际站跨区域部署，并且不追求监控数据完全一致，因此建议采用双 Center：

```text
中国站 Center：ops-center-cn
国际站 Center：ops-center-intl
```

各自保存本地区详细数据，只互相交换少量摘要状态。

```text
详细指标：本地保存
摘要状态：跨区域同步
告警决策：第一版不做主动通知，仅在看板展示状态
```

---

## 3. 推荐目录结构

当前目录可以继续保持：

```text
ops/
├── gofurry-ops-agent
└── gofurry-ops-center
```

推荐进一步展开为：

```text
ops/
├── gofurry-ops-agent/
│   ├── cmd/
│   │   └── agent/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── collector/
│   │   │   ├── system/
│   │   │   ├── docker/
│   │   │   ├── httpcheck/
│   │   │   ├── postgres/
│   │   │   ├── redis/
│   │   │   └── cert/
│   │   ├── reporter/
│   │   ├── spool/
│   │   ├── identity/
│   │   └── runtime/
│   ├── configs/
│   │   └── agent.example.yaml
│   ├── deploy/
│   │   └── systemd/
│   │       └── gofurry-ops-agent.service
│   ├── go.mod
│   └── README.md
│
└── gofurry-ops-center/
    ├── cmd/
    │   └── center/
    │       └── main.go
    ├── internal/
    │   ├── app/
    │   ├── config/
    │   ├── http/
    │   │   ├── handler/
    │   │   ├── middleware/
    │   │   └── router.go
    │   ├── module/
    │   │   ├── agent/
    │   │   ├── node/
    │   │   ├── metric/
    │   │   ├── servicecheck/
    │   │   ├── alertview/
    │   │   ├── peer/
    │   │   ├── syncstatus/
    │   │   └── auth/
    │   ├── repository/
    │   │   └── db/
    │   │       ├── query.sql
    │   │       └── sqlc/
    │   ├── scheduler/
    │   └── web/
    │       └── dist/
    ├── web/
    │   ├── src/
    │   ├── package.json
    │   └── vite.config.ts
    ├── migrations/
    ├── sqlc.yaml
    ├── configs/
    │   └── center.example.yaml
    ├── deploy/
    │   ├── docker-compose.prod.yml
    │   └── nginx/
    ├── Dockerfile
    ├── go.mod
    └── README.md
```

---

## 4. Agent 技术选型

### 4.1 Agent 定位

`gofurry-ops-agent` 是轻量采集器：

```text
定时采集
主动上报
本地失败缓存
低资源占用
不提供复杂 Web 服务
不依赖数据库
```

### 4.2 Agent 推荐技术栈

| 功能 | 技术选择 | 说明 |
|---|---|---|
| 语言 | Go | 单文件部署、资源占用低、适合服务器探针 |
| CLI | Cobra，可选 | 如果命令复杂再引入；第一版也可直接 flags |
| 配置 | YAML + 环境变量覆盖 | 适合服务器部署和差异化配置 |
| 日志 | `log/slog` | Go 标准库结构化日志 |
| 系统指标 | `gopsutil` | CPU、内存、磁盘、负载、网络 |
| Docker 状态 | Docker Engine SDK for Go | 容器运行状态、重启次数、资源粗略占用 |
| HTTP 检查 | `net/http` | 标准库足够 |
| 上报 | `net/http` | Agent 不需要引入 Web 框架 |
| 本地缓存 | JSONL spool 文件 | 第一版比 SQLite 更简单 |
| 部署 | systemd | Agent 更适合以系统服务运行 |

### 4.3 Agent 不建议使用的技术

```text
不使用 Fiber
不内置复杂 Web UI
不做 Prometheus exporter
不在本地跑数据库
不采集全量进程列表
不做秒级高频采集
```

Agent 的核心是：**稳定、安静、低资源占用**。

### 4.4 Agent 配置示例

```yaml
node:
  id: cn-business-a
  name: CN Business Server A
  region: cn
  role: business

center:
  endpoint: https://ops-cn.go-furry.com/api/v1/agent/ingest
  token: ${GOFURRY_OPS_AGENT_TOKEN}
  timeout: 5s

collect:
  interval: 30s

system:
  enabled: true
  disk_mounts:
    - /
    - /data

docker:
  enabled: true
  containers:
    - nginx
    - postgres
    - redis
    - gofurry-nav-backend

http_checks:
  - name: cn-home
    url: https://go-furry.com/
    timeout: 5s
  - name: cn-api-health
    url: https://go-furry.com/api/health
    timeout: 5s

spool:
  enabled: true
  dir: /var/lib/gofurry-ops-agent/spool
  max_files: 1000
```

---

## 5. Center 技术选型

### 5.1 Center 定位

`gofurry-ops-center` 是轻量中心端：

```text
接收 Agent 上报
保存监控数据
提供看板 API
展示节点和服务状态
展示跨区域摘要
展示同步状态
展示异常状态
第一版不主动发送告警通知
```

### 5.2 Center 推荐技术栈

| 功能 | 技术选择 | 说明 |
|---|---|---|
| 语言 | Go | 与现有后端技术栈一致 |
| Web 框架 | Fiber | 用户熟悉、轻量、开发快 |
| 数据库 | PostgreSQL | 已熟悉，适合状态与历史数据存储 |
| DB Driver | pgx | Go PostgreSQL 生态常用选择 |
| SQL 生成 | sqlc | 写 SQL，生成类型安全 Go 代码 |
| Migration | goose | 简单直接，适合 SQL 迁移 |
| 日志 | `log/slog` | 统一结构化日志 |
| 前端 | Vue 3 + Vite + Tailwind | 内部看板不需要 SSR |
| 前端嵌入 | `go:embed` | 构建后打进 Center，简化部署 |
| 部署 | Docker Compose | Center + PostgreSQL 部署简单 |

### 5.3 为什么 Center 适合 Fiber

Center 需要提供：

```text
Agent 上报 API
Dashboard API
Peer Summary API
管理配置 API
静态前端资源
```

这里用 Fiber 很合适，既和现有 GoFurry 后端技术栈一致，也能保持轻量开发体验。

### 5.4 为什么推荐 sqlc

Center 会有大量固定 SQL：

```text
写入心跳
写入系统采样
写入 HTTP 检查结果
查询节点最新状态
查询最近 24 小时趋势
查询服务当前状态
查询同步状态
查询异常历史
```

这些查询非常适合：

```text
PostgreSQL + pgx + sqlc
```

不建议使用 GORM。这里不是复杂业务系统，ORM 对时序采样、聚合查询、保留策略的帮助有限，反而可能让 SQL 不直观。

---

## 6. 数据库存储设计

### 6.1 第一版不使用时序数据库

第一版不需要引入 Prometheus TSDB、VictoriaMetrics、InfluxDB 等组件。

直接使用 PostgreSQL 即可：

```text
部署简单
查询可控
和 Go/Fiber/sqlc 配合直接
足够支撑个人站监控数据量
```

### 6.2 推荐表结构

```text
nodes
agent_tokens
node_heartbeats
system_samples
docker_container_samples
http_check_results
service_status
alert_states
alert_events
peer_centers
peer_summaries
sync_status
deploy_events
```

说明：

```text
nodes：节点基础信息
agent_tokens：Agent 鉴权信息
node_heartbeats：节点心跳历史
system_samples：CPU / 内存 / 磁盘 / 网络采样
docker_container_samples：容器状态采样
http_check_results：HTTP 检查历史
service_status：当前最新服务状态
alert_states：当前异常状态，不主动发送通知
alert_events：异常变化事件，用于看板查看
peer_centers：对端 Center 配置
peer_summaries：对端 Center 摘要
sync_status：国内到国际同步状态
deploy_events：部署事件记录
```

### 6.3 当前状态和历史数据分离

推荐把“当前状态”和“历史采样”分开。

历史数据：

```text
node_heartbeats
system_samples
http_check_results
docker_container_samples
```

当前状态：

```text
service_status
alert_states
sync_status
peer_summaries
```

这样首页总览不需要扫描大量历史数据。

### 6.4 保留策略

第一版建议：

```text
30 秒 / 60 秒原始数据：保留 7 天
5 分钟聚合数据：第二版再做
1 小时聚合数据：第二版再做
异常事件：长期保留
部署事件：长期保留
同步事件：长期保留
```

第一版可以先用定时清理：

```sql
DELETE FROM system_samples
WHERE created_at < now() - interval '7 days';

DELETE FROM http_check_results
WHERE created_at < now() - interval '7 days';

DELETE FROM docker_container_samples
WHERE created_at < now() - interval '7 days';
```

---

## 7. 看板设计

### 7.1 前端技术

推荐：

```text
Vue 3 + Vite + Tailwind CSS
```

不建议 Nuxt，因为运维看板是内部系统：

```text
不需要 SEO
不需要 SSR
不需要公开内容优化
不需要复杂路由预渲染
```

构建后通过 `go:embed` 嵌入 Center。

### 7.2 页面结构

```text
/overview        总览
/nodes           服务器列表
/nodes/:id       单机详情
/services        服务健康
/sync            数据同步状态
/peers           双 Center 对端摘要
/alerts          异常状态与历史事件
/deployments     部署记录
/settings        节点、Token、Peer 配置
```

### 7.3 总览页重点

总览页不追求 Grafana 式复杂曲线，而是快速回答：

```text
中国站是否正常？
国际站是否正常？
哪台机器异常？
最近一次数据同步是否成功？
当前是否有严重异常？
证书是否快过期？
数据库是否可连接？
关键页面是否可访问？
```

---

## 8. 双 Center 弱一致设计

### 8.1 部署模式

```text
中国运维服务器 B
└── ops-center-cn
    ├── 接收中国业务服务器 A 的 Agent 数据
    ├── 接收中国运维服务器 B 本机 Agent 数据
    ├── 保存中国站详细指标
    └── 拉取国际 Center 摘要

国际站服务器
└── ops-center-intl
    ├── 接收国际站服务器 Agent 数据
    ├── 保存国际站详细指标
    ├── 展示国际站同步状态
    └── 拉取中国 Center 摘要
```

### 8.2 不做全量同步

不建议：

```text
所有 Agent 同时上报两个 Center
两个 Center 全量同步指标
两个 Center 互相复制数据库
两个 Center 都对同一异常做主动告警
```

建议：

```text
Agent 默认只上报本地 Center
Center 之间只交换 summary
详细数据只保存在本地
```

### 8.3 Peer Summary API

建议接口：

```text
GET  /api/v1/peer/summary
POST /api/v1/peer/heartbeat
```

摘要示例：

```json
{
  "region": "cn",
  "center": "ops-center-cn",
  "status": "ok",
  "last_heartbeat_at": "2026-05-18T12:00:00Z",
  "nodes_total": 2,
  "nodes_down": 0,
  "critical_alerts": 0,
  "warning_alerts": 1,
  "last_sync_status": "success",
  "updated_at": "2026-05-18T12:00:00Z"
}
```

### 8.4 跨区不可达的语义

如果国际 Center 访问不到中国 Center，不要直接判断“中国站挂了”。

应该区分：

```text
Service Down：服务本身不可用
Center Unreachable：对端 Center 不可达
Cross-region Check Failed：跨区检查失败
```

跨区网络波动不能等同于服务宕机。

---

## 9. 主动告警策略

### 9.1 第一版不做主动通知

本系统第一版**不实现主动告警通知**，例如：

```text
不发邮件
不发短信
不发 Telegram
不发企业微信
不发 Bark / Server 酱
```

原因：

```text
云厂商已经提供基础资源告警能力
主动通知链路会增加复杂度
单人维护场景下，过多告警容易变成噪音
第一版重点是可视化状态和快速排障
```

### 9.2 仍然保留“异常状态”概念

虽然不主动通知，但 Center 仍然需要计算和展示异常状态。

例如：

```text
节点 3 分钟无心跳
磁盘使用率 > 85%
内存使用率连续 5 分钟 > 90%
HTTP 检查连续 3 次失败
PostgreSQL 不可连接
Redis 不可连接
证书剩余 < 14 天
国际站同步超过 24 小时未成功
```

这些状态只在看板里展示，记录到：

```text
alert_states
alert_events
```

更准确地说，第一版的 `alert` 是“异常状态”和“事件记录”，不是“主动通知”。

### 9.3 后续可扩展主动通知

后续如果需要，可以新增 notifier 模块：

```go
type Notifier interface {
    Send(ctx context.Context, event AlertEvent) error
}
```

可选实现：

```text
Email
Telegram
企业微信
Bark
Server 酱
短信服务
```

但第一版不做。

---

## 10. 安全设计

### 10.1 Agent 到 Center

Agent 上报不建议只靠裸 token，推荐：

```text
HTTPS
Agent Token
HMAC-SHA256 签名
timestamp 防重放
node_id 固定注册
```

请求头示例：

```text
X-GoFurry-Node-ID: cn-business-a
X-GoFurry-Timestamp: 2026-05-18T12:00:00Z
X-GoFurry-Signature: hmac-sha256...
```

### 10.2 Center 到 Center

双 Center 通信使用独立 Peer Token：

```text
只开放 summary API
不开放远程管理权限
不允许国际 Center 管理中国 Center
不允许中国 Center 管理国际 Center
```

### 10.3 Dashboard 访问控制

Center 看板不要裸奔。

第一版可选：

```text
Nginx Basic Auth
Cloudflare Access / Zero Trust
Tailscale
管理员账号 + 密码
单个 ADMIN_TOKEN
```

推荐组合：

```text
Cloudflare Access 或 Nginx Basic Auth
+
Center 内部管理员登录
```

---

## 11. 部署建议

### 11.1 Agent 部署

Agent 推荐用 systemd：

```text
/etc/gofurry-ops-agent/config.yaml
/usr/local/bin/gofurry-ops-agent
/var/lib/gofurry-ops-agent/spool
```

systemd unit 示例：

```ini
[Unit]
Description=GoFurry Ops Agent
After=network-online.target docker.service
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/gofurry-ops-agent --config /etc/gofurry-ops-agent/config.yaml
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
```

### 11.2 Center 部署

Center 推荐 Docker Compose：

```yaml
services:
  ops-center:
    image: gofurry-ops-center:latest
    restart: always
    env_file:
      - ./env/center.env
    depends_on:
      - postgres

  postgres:
    image: postgres:18
    restart: always
    volumes:
      - ./data/postgres:/var/lib/postgresql/data

  nginx:
    image: nginx:alpine
    restart: always
    volumes:
      - ./nginx:/etc/nginx/conf.d
    ports:
      - "80:80"
      - "443:443"
```

---

## 12. MVP 范围

第一版建议只做：

### Agent

```text
心跳
CPU / 内存 / 磁盘 / 网络
Docker 容器状态
HTTP health checks
PostgreSQL ping
Redis ping
本地 JSONL spool
HMAC 上报
```

### Center

```text
Agent ingest API
节点注册 / Token 管理
系统采样入库
HTTP 检查结果入库
服务当前状态计算
异常状态展示，不主动通知
Peer Summary API
Vue 看板
```

### Dashboard

```text
总览页
节点列表
节点详情
服务健康
同步状态
对端 Center 摘要
异常状态页
```

---

## 13. 暂不做的内容

第一版不做：

```text
主动邮件告警
短信告警
Telegram 告警
企业微信告警
复杂告警规则引擎
PromQL 类查询语言
Grafana 式自由拖拽大屏
全量日志采集
全量进程采集
跨 Center 全量指标同步
Kubernetes 部署
高可用集群
```

这些都可以等系统稳定后按需扩展。

---

## 14. 最终技术选型总结

### Agent

```text
Go
log/slog
gopsutil
Docker Engine SDK
net/http
YAML 配置
JSONL spool
systemd 部署
```

Agent 不使用：

```text
Fiber
数据库
Prometheus exporter
内置 Web UI
```

### Center

```text
Go
Fiber
PostgreSQL
pgx
sqlc
goose
log/slog
Vue 3
Vite
Tailwind CSS
go:embed
Docker Compose
```

### 告警

```text
第一版不做主动通知
只做异常状态计算和看板展示
后续预留 notifier 扩展点
```

---

## 15. 结论

GoFurry 轻量运维系统的目标不是成为通用监控平台，而是成为一个**适合单人维护、同时覆盖中国站与国际站的专用运维看板**。

推荐最终形态：

```text
Agent：轻量 Go daemon，主动上报
Center：Fiber + PostgreSQL + Vue 看板
双 Center：本地自治，摘要互通
告警：第一版只展示异常，不主动通知
部署：Agent 用 systemd，Center 用 Docker Compose
```

一句话总结：

> 用 GoFurry 自己真正需要的数据，替代过重的通用监控栈；先做到轻量、可靠、可维护，再考虑扩展主动告警和复杂分析能力。
