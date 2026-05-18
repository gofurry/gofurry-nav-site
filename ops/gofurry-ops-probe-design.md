# GoFurry 轻量运维探针设计方案

> 适用仓库目录：`gofurry-nav-site/ops`  
> 当前规划目录：
>
> ```text
> ops/
> ├── gofurry-ops-agent
> └── gofurry-ops-center
> ```
>
> 目标：替代原有 `Prometheus + Grafana + 自定义 Prometheus 数据看板` 的重型监控体系，改为适合 GoFurry 单人维护、跨中国站与国际站的轻量运维系统。

---

## 1. 背景

GoFurry 当前存在两类站点：

- **中国站 `go-furry.com`**
  - 国内主站
  - 腾讯云业务服务器 A：4c8g
  - 腾讯云运维服务器 B：2c4g
  - Nginx 反代 + 腾讯云 EdgeOne
  - 承担主要业务、管理后台、采集、监控日志、完整数据库等

- **国际站 `gofurry.com`**
  - 国际镜像站
  - Cloudflare 前置
  - 海外服务器
  - 独立定制化前端
  - 单体后端
  - 精简数据库
  - 每日从中国站同步公开精简数据

原有监控体系：

```text
Prometheus + Grafana + 自定义 Prometheus 数据看板
```

优点是专业、企业化、指标能力强；但对 GoFurry 当前的单人维护模式来说存在明显问题：

- 组件偏多
- 资源占用偏高
- 配置和维护成本较高
- 通用指标很多，但实际决策价值有限
- 双站架构下跨区域监控链路复杂

因此，计划重构为：

```text
轻量运维探针 + 自研 Ops Center + 定制化数据看板
```

---

## 2. 设计目标

### 2.1 核心目标

这个系统不是要重新实现 Prometheus，而是要实现一个 **GoFurry 专用的轻量观测系统**。

它主要回答这些问题：

```text
服务器还活着吗？
CPU / 内存 / 磁盘是否异常？
Docker 服务是否正常？
Nginx / API / Nuxt 是否可访问？
PostgreSQL / Redis 是否正常？
HTTPS 证书是否快过期？
国内站到国际站的数据同步是否成功？
最近一次部署是否正常？
中国站和国际站是否互相可见？
```

### 2.2 非目标

第一版不追求：

- 不做通用监控平台
- 不做 Prometheus TSDB 替代品
- 不做 Grafana 式自由拖拽大屏
- 不做强一致多中心同步
- 不做复杂告警编排系统
- 不做 Kubernetes 级别的服务发现
- 不做海量日志检索系统

---

## 3. 总体架构

系统由两个核心部分组成：

```text
gofurry-ops-agent   # 轻量探针，部署在每台服务器
gofurry-ops-center  # 数据接收、存储、看板、告警
```

推荐架构：

```text
中国业务服务器 A
  └── gofurry-ops-agent

中国运维服务器 B
  ├── gofurry-ops-agent
  └── gofurry-ops-center-cn

国际站服务器
  ├── gofurry-ops-agent
  └── gofurry-ops-center-intl
```

双 Center 关系：

```text
ops-center-cn       <---- summary / heartbeat ---->       ops-center-intl
     ↑                                                            ↑
     │                                                            │
中国服务器详细指标                                      国际服务器详细指标
```

---

## 4. 双 Center 设计

### 4.1 为什么允许两边都放 Center

GoFurry 同时有中国站和国际站，跨境链路可能存在不稳定。因此不建议只依赖单一 Center。

推荐模式：

```text
双 Center、弱一致、本地自治、摘要互通
```

含义：

- 中国站有自己的 Center
- 国际站有自己的 Center
- 各自保存本区域详细指标
- 两个 Center 只交换少量摘要状态
- 不追求历史指标完全一致
- 不进行全量指标跨区域复制

### 4.2 中国 Center 职责

部署位置：腾讯云运维服务器 B。

职责：

- 接收中国业务服务器 A 的 Agent 数据
- 接收中国运维服务器 B 本机 Agent 数据
- 监控中国站服务健康
- 展示国内详细指标
- 拉取或接收国际 Center 的摘要状态
- 对中国站本地故障负责告警

### 4.3 国际 Center 职责

部署位置：国际站服务器。

职责：

- 接收国际站服务器本机 Agent 数据
- 监控国际站服务健康
- 展示国际站详细指标
- 记录国内到国际的数据同步状态
- 拉取或接收中国 Center 的摘要状态
- 对国际站本地故障负责告警

### 4.4 Center 之间同步什么

只同步摘要，不同步详细指标。

示例摘要：

```json
{
  "region": "cn",
  "center_id": "ops-center-cn",
  "status": "ok",
  "last_heartbeat_at": "2026-05-18T12:00:00Z",
  "nodes_total": 2,
  "nodes_down": 0,
  "critical_alerts": 0,
  "warning_alerts": 1,
  "last_sync_status": "success"
}
```

### 4.5 告警边界

本地问题由本地 Center 告警：

```text
中国业务服务器 A 磁盘异常 -> ops-center-cn 告警
国际站数据同步失败 -> ops-center-intl 告警
```

跨区域不可达时，只能提示：

```text
Center Unreachable
Cross-region Check Failed
```

不要直接判断：

```text
中国站宕机
国际站宕机
```

因为跨境网络抖动不一定等于对端服务故障。

---

## 5. 仓库目录设计

当前目录：

```text
ops/
├── gofurry-ops-agent
└── gofurry-ops-center
```

建议扩展为：

```text
ops/
├── gofurry-ops-agent/
│   ├── cmd/
│   │   └── agent/
│   │       └── main.go
│   ├── internal/
│   │   ├── collector/
│   │   │   ├── system/
│   │   │   ├── docker/
│   │   │   ├── httpcheck/
│   │   │   ├── postgres/
│   │   │   ├── redis/
│   │   │   └── cert/
│   │   ├── config/
│   │   ├── reporter/
│   │   ├── spool/
│   │   └── security/
│   ├── configs/
│   │   └── agent.example.yaml
│   ├── deploy/
│   │   ├── systemd/
│   │   └── docker/
│   ├── go.mod
│   └── README.md
│
├── gofurry-ops-center/
│   ├── cmd/
│   │   └── center/
│   │       └── main.go
│   ├── internal/
│   │   ├── app/
│   │   ├── config/
│   │   ├── http/
│   │   ├── ingest/
│   │   ├── peer/
│   │   ├── alert/
│   │   ├── repository/
│   │   ├── service/
│   │   ├── model/
│   │   └── security/
│   ├── web/
│   │   ├── src/
│   │   └── package.json
│   ├── migrations/
│   ├── configs/
│   │   └── center.example.yaml
│   ├── deploy/
│   │   ├── docker-compose.yml
│   │   ├── nginx/
│   │   └── systemd/
│   ├── go.mod
│   └── README.md
│
└── docs/
    ├── architecture.md
    ├── api.md
    ├── alerting.md
    ├── deploy.md
    └── security.md
```

也可以保持更简单的两目录结构，后续再逐步补充。

---

## 6. Agent 设计

### 6.1 Agent 定位

Agent 是部署在每台服务器上的轻量探针。

职责：

- 定时采集本机状态
- 执行本机和远程健康检查
- 将数据通过 HTTPS 上报给 Center
- 在网络失败时短暂本地缓存
- 尽量低资源占用

### 6.2 Agent 采集内容

第一版建议采集以下内容。

#### 主机资源

```text
CPU 使用率
内存使用率
磁盘使用率
磁盘 inode 使用率
Load Average
网络收发速率
系统启动时间
Agent 版本
```

#### Docker 状态

```text
容器是否运行
容器重启次数
容器健康状态
容器 CPU / 内存粗略占用
关键容器是否存在
```

#### HTTP 检查

```text
HTTP 状态码
响应时间
响应体关键字
连续失败次数
TLS 是否正常
```

#### PostgreSQL 检查

```text
是否可连接
数据库大小
连接数
简单 SELECT 1 延迟
```

#### Redis 检查

```text
是否可 ping
内存占用
key 数量粗略统计
连接数
```

#### 证书检查

```text
HTTPS 证书剩余天数
证书域名是否匹配
证书是否已过期
```

#### 数据同步状态

由国际站 sync 命令或国际站后端写入 Center：

```text
最近一次同步时间
同步是否成功
同步版本号
同步数据量
checksum 是否一致
失败原因
```

### 6.3 Agent 配置示例

```yaml
node_id: cn-business-a
region: cn
role: business

report:
  endpoint: "https://ops-cn.go-furry.com/api/v1/agent/ingest"
  token: "CHANGE_ME"
  interval: "30s"
  timeout: "5s"
  retry: 3
  spool_dir: "/var/lib/gofurry-ops-agent/spool"

system:
  enabled: true
  disks:
    - "/"
    - "/data"

docker:
  enabled: true
  watch_containers:
    - "nginx"
    - "postgres"
    - "redis"
    - "gofurry-nav-backend"

http_checks:
  - name: "cn-home"
    url: "https://go-furry.com/"
    method: "GET"
    timeout: "5s"
    expect_status: 200

  - name: "cn-api-health"
    url: "https://go-furry.com/api/health"
    method: "GET"
    timeout: "5s"
    expect_status: 200

postgres:
  enabled: true
  dsn: "postgres://user:pass@127.0.0.1:5432/gofurry?sslmode=disable"

redis:
  enabled: true
  addr: "127.0.0.1:6379"
  password: ""
```

---

## 7. Center 设计

### 7.1 Center 定位

Center 是 Ops 系统的中心服务。

职责：

- 接收 Agent 上报
- 存储监控数据
- 计算节点状态
- 生成告警
- 提供 Dashboard API
- 与对端 Center 交换摘要
- 提供轻量 Web 看板

### 7.2 Center 配置示例

```yaml
center_id: ops-center-cn
region: cn

server:
  addr: ":8080"

storage:
  driver: postgres
  dsn: "postgres://ops:pass@127.0.0.1:5432/gofurry_ops?sslmode=disable"

security:
  dashboard_basic_auth:
    username: "admin"
    password: "CHANGE_ME"
  agent_tokens:
    - node_id: "cn-business-a"
      token: "CHANGE_ME"
    - node_id: "cn-ops-b"
      token: "CHANGE_ME"

peer:
  enabled: true
  remote_summary_url: "https://ops-intl.gofurry.com/api/v1/peer/summary"
  token: "CHANGE_ME_PEER_TOKEN"
  interval: "60s"

retention:
  raw_samples_days: 7
  five_minute_samples_days: 30
  hourly_samples_days: 180

alert:
  enabled: true
  cooldown: "30m"
  notify:
    email: false
    telegram: false
```

---

## 8. API 草案

### 8.1 Agent 上报

```http
POST /api/v1/agent/ingest
Authorization: Bearer <agent-token>
Content-Type: application/json
```

请求体示例：

```json
{
  "node_id": "cn-business-a",
  "region": "cn",
  "timestamp": "2026-05-18T12:00:00Z",
  "agent_version": "0.1.0",
  "system": {
    "cpu_usage": 23.5,
    "memory_usage": 61.2,
    "load1": 0.72,
    "uptime_seconds": 123456
  },
  "disks": [
    {
      "mount": "/",
      "usage": 68.4,
      "inode_usage": 12.1
    }
  ],
  "http_checks": [
    {
      "name": "cn-home",
      "status": "ok",
      "status_code": 200,
      "latency_ms": 87
    }
  ],
  "docker": [
    {
      "name": "nginx",
      "running": true,
      "restart_count": 0
    }
  ]
}
```

### 8.2 Center 摘要

```http
GET /api/v1/peer/summary
Authorization: Bearer <peer-token>
```

响应示例：

```json
{
  "region": "intl",
  "center_id": "ops-center-intl",
  "status": "ok",
  "last_heartbeat_at": "2026-05-18T12:00:00Z",
  "nodes_total": 1,
  "nodes_down": 0,
  "critical_alerts": 0,
  "warning_alerts": 0,
  "last_sync_status": "success"
}
```

### 8.3 Dashboard API

```text
GET /api/v1/dashboard/overview
GET /api/v1/nodes
GET /api/v1/nodes/:id
GET /api/v1/services
GET /api/v1/alerts
GET /api/v1/sync-runs
GET /api/v1/peer/status
```

---

## 9. 数据库设计草案

第一版可以使用 PostgreSQL。

核心表：

```sql
CREATE TABLE nodes (
    id              BIGSERIAL PRIMARY KEY,
    node_id          TEXT NOT NULL UNIQUE,
    region           TEXT NOT NULL,
    role             TEXT NOT NULL,
    display_name     TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE node_heartbeats (
    id              BIGSERIAL PRIMARY KEY,
    node_id          TEXT NOT NULL,
    region           TEXT NOT NULL,
    agent_version    TEXT,
    reported_at      TIMESTAMPTZ NOT NULL,
    received_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE system_samples (
    id              BIGSERIAL PRIMARY KEY,
    node_id          TEXT NOT NULL,
    cpu_usage        DOUBLE PRECISION,
    memory_usage     DOUBLE PRECISION,
    load1            DOUBLE PRECISION,
    load5            DOUBLE PRECISION,
    load15           DOUBLE PRECISION,
    uptime_seconds   BIGINT,
    reported_at      TIMESTAMPTZ NOT NULL
);

CREATE TABLE disk_samples (
    id              BIGSERIAL PRIMARY KEY,
    node_id          TEXT NOT NULL,
    mount            TEXT NOT NULL,
    usage            DOUBLE PRECISION,
    inode_usage      DOUBLE PRECISION,
    reported_at      TIMESTAMPTZ NOT NULL
);

CREATE TABLE http_check_results (
    id              BIGSERIAL PRIMARY KEY,
    node_id          TEXT NOT NULL,
    name             TEXT NOT NULL,
    status           TEXT NOT NULL,
    status_code      INTEGER,
    latency_ms       INTEGER,
    error_message    TEXT,
    reported_at      TIMESTAMPTZ NOT NULL
);

CREATE TABLE docker_container_samples (
    id              BIGSERIAL PRIMARY KEY,
    node_id          TEXT NOT NULL,
    name             TEXT NOT NULL,
    running          BOOLEAN NOT NULL,
    restart_count    INTEGER DEFAULT 0,
    health_status    TEXT,
    reported_at      TIMESTAMPTZ NOT NULL
);

CREATE TABLE sync_runs (
    id              BIGSERIAL PRIMARY KEY,
    region           TEXT NOT NULL,
    sync_name        TEXT NOT NULL,
    version          TEXT,
    status           TEXT NOT NULL,
    items_total      INTEGER,
    checksum_ok      BOOLEAN,
    error_message    TEXT,
    started_at       TIMESTAMPTZ,
    finished_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE alerts (
    id              BIGSERIAL PRIMARY KEY,
    region           TEXT NOT NULL,
    node_id          TEXT,
    level            TEXT NOT NULL,
    type             TEXT NOT NULL,
    title            TEXT NOT NULL,
    message          TEXT,
    status           TEXT NOT NULL,
    first_seen_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at      TIMESTAMPTZ
);

CREATE TABLE peer_summaries (
    id              BIGSERIAL PRIMARY KEY,
    peer_region      TEXT NOT NULL,
    peer_center_id   TEXT NOT NULL,
    status           TEXT NOT NULL,
    nodes_total      INTEGER,
    nodes_down       INTEGER,
    critical_alerts  INTEGER,
    warning_alerts   INTEGER,
    last_sync_status TEXT,
    received_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

后续可以增加聚合表，避免长期存储原始样本。

---

## 10. 数据保留策略

建议保留策略：

```text
原始 30s / 60s 采样数据：保留 7 天
5 分钟聚合数据：保留 30 天
1 小时聚合数据：保留 180 天
告警记录：长期保留
部署事件：长期保留
同步记录：长期保留或保留 1 年
```

可以通过每日定时任务清理：

```sql
DELETE FROM system_samples
WHERE reported_at < now() - interval '7 days';

DELETE FROM http_check_results
WHERE reported_at < now() - interval '7 days';
```

---

## 11. 告警设计

### 11.1 第一版告警规则

建议只做少量高价值告警：

```text
节点 3 分钟无心跳
磁盘使用率 > 85%
内存使用率连续 5 分钟 > 90%
关键 HTTP 检查连续 3 次失败
PostgreSQL 不可连接
Redis 不可连接
HTTPS 证书剩余 < 14 天
国际站同步超过 24 小时未成功
对端 Center 超过 5 分钟不可达
```

### 11.2 告警降噪

必须支持：

```text
连续失败 N 次后告警
恢复后发送恢复通知
同一告警冷却 30 分钟
同类告警合并
跨区不可达不等于服务宕机
```

### 11.3 通知渠道

第一版可以先实现一种：

```text
Email
Telegram Bot
企业微信机器人
Server 酱
Bark
```

建议先使用最容易实现的通知方式，不要一开始做全渠道。

---

## 12. Dashboard 设计

自研看板不需要复刻 Grafana，而要做 GoFurry 专用的运维决策面板。

### 12.1 页面结构

```text
/overview       总览
/nodes          节点列表
/nodes/:id      单机详情
/services       服务健康
/sync           国内 -> 国际同步状态
/peer           对端 Center 状态
/alerts         当前告警和历史告警
/deployments    部署事件，第二版
/settings       配置和 token 管理，后续可选
```

### 12.2 总览页内容

总览页只放最关键内容：

```text
中国业务服务器 A：正常 / 异常
中国运维服务器 B：正常 / 异常
国际站服务器：正常 / 异常

go-furry.com 首页：正常 / 异常
go-furry.com API：正常 / 异常
gofurry.com 首页：正常 / 异常
gofurry.com API：正常 / 异常

最近一次国内 -> 国际同步：成功 / 失败
当前严重告警数量
当前警告数量
对端 Center 是否可达
```

### 12.3 不建议第一版做的功能

```text
自由拖拽大屏
复杂图表编辑器
PromQL 类查询语言
多租户权限系统
复杂角色权限系统
无限指标维度
```

---

## 13. 安全设计

### 13.1 Agent 到 Center

最低要求：

```text
HTTPS
Agent Token
node_id 固定注册
时间戳
请求体签名，第二版可做
失败重试
```

建议 Header：

```http
Authorization: Bearer <agent-token>
X-GoFurry-Node-ID: cn-business-a
X-GoFurry-Timestamp: 2026-05-18T12:00:00Z
```

第二版可以加入 HMAC：

```http
X-GoFurry-Signature: sha256=...
```

### 13.2 Center 到 Center

Center 之间只开放摘要 API：

```text
GET /api/v1/peer/summary
POST /api/v1/peer/heartbeat
```

不要开放管理接口。

### 13.3 Dashboard 访问保护

第一版可以使用：

```text
Basic Auth
固定管理员 token
Nginx IP 白名单
Cloudflare Access，国际站可选
Tailscale，后续可选
```

---

## 14. 部署建议

### 14.1 Agent 部署方式

推荐优先使用 systemd。

```text
/etc/gofurry-ops-agent/config.yaml
/usr/local/bin/gofurry-ops-agent
/var/lib/gofurry-ops-agent/spool
/var/log/gofurry-ops-agent/agent.log
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

### 14.2 Center 部署方式

Center 可以使用 Docker Compose。

```yaml
services:
  ops-center:
    image: gofurry-ops-center:latest
    restart: unless-stopped
    ports:
      - "127.0.0.1:8080:8080"
    volumes:
      - ./config.yaml:/etc/gofurry-ops-center/config.yaml:ro
    depends_on:
      - postgres

  postgres:
    image: postgres:18
    restart: unless-stopped
    environment:
      POSTGRES_DB: gofurry_ops
      POSTGRES_USER: ops
      POSTGRES_PASSWORD: change_me
    volumes:
      - postgres-data:/var/lib/postgresql/data

volumes:
  postgres-data:
```

Nginx 反代：

```nginx
server {
    listen 443 ssl http2;
    server_name ops-cn.go-furry.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## 15. MVP 开发计划

### Phase 0：项目骨架

```text
创建 gofurry-ops-agent
创建 gofurry-ops-center
定义配置文件格式
定义 Agent 上报 DTO
定义数据库 migrations
```

### Phase 1：最小可用 Agent

```text
采集 CPU / 内存 / 磁盘
采集心跳
上报 Center
本地失败重试
日志输出
```

### Phase 2：最小可用 Center

```text
接收 Agent 数据
写入 PostgreSQL
节点列表 API
总览 API
简单 Dashboard 页面
```

### Phase 3：服务健康检查

```text
HTTP checks
Docker checks
PostgreSQL checks
Redis checks
```

### Phase 4：告警

```text
节点离线告警
磁盘告警
HTTP 连续失败告警
同步失败告警
告警恢复
告警冷却
```

### Phase 5：双 Center 摘要互通

```text
peer summary API
拉取对端 summary
对端不可达提示
Dashboard 展示对端状态
```

### Phase 6：国际站同步状态接入

```text
国际站 sync 命令写入 sync_runs
Center 展示最近同步状态
同步失败触发告警
```

---

## 16. 第一版验收标准

第一版完成后，应满足：

```text
中国业务服务器 A 能上报状态
中国运维服务器 B 能上报状态
国际站服务器能上报状态
中国 Center 能展示中国站详细状态
国际 Center 能展示国际站详细状态
两个 Center 能看到对端摘要
关键 HTTP 服务异常能触发告警
服务器离线能触发告警
国际站同步失败能触发告警
不再依赖 Prometheus 和 Grafana
```

---

## 17. 最终定位

GoFurry Ops 的定位不是通用监控系统，而是：

> **GoFurry 专用的轻量运维观测系统。**

它应该围绕以下优先级设计：

```text
简单 > 通用
稳定 > 炫酷
低资源 > 大而全
能排障 > 指标数量多
单人可维护 > 企业化复杂度
```

最终目标：

```text
让 GoFurry 的中国站和国际站都能被轻量、清晰、低成本地观察和维护。
```
