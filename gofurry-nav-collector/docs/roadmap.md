# GoFurry Nav Collector 路线图

## 当前状态

`gofurry-nav-collector` 当前已经完成 Phase 0 稳定化，并在生产环境验证通过。已完成内容包括：防重入、局部 worker 状态、网络预算、GeoIP 单例、Redis timeout、保留清理 SQL、zap 中文日志、空 Hash 写入修复、保守日志保留配置。

后续路线图只保留未完成内容。已完成的 Phase 0 细节不再在本文展开，避免路线图变成历史记录。

## 迭代原则

- 不破坏旧接口、旧 Redis key、旧表语义。
- 新能力先旁路，后灰度，再考虑切读。
- 默认低频、低强度，不做漏洞扫描、端口全扫、目录爆破、弱口令尝试或高强度探测。
- 不引入 Prometheus 生态；观测优先依赖中文日志、SQL/Redis 人工查询、可解释的状态数据。
- 外部站点返回的 title、meta、headers、TXT、PTR、证书 subject/issuer 等内容全部按不可信文本处理。
- 推测和事实必须分开：采集结果是 observation，不直接包装成绝对判断。

## 版本计划

### v0.2.0 - v2 Observation 核心数据面

- **状态：** 已完成
- **范围：** 数据模型 / Redis / 灰度安全
- **目标：** 新增 v2 observation 数据面，但不影响现有后端和 Nuxt 前端读取路径。

#### 重点

- 通用 observation 表。
- v2 latest Redis key。
- v1/v2 旁路双写。
- 配置开关与回滚路径。
- `site_id` 显式关联的兼容期基础。

#### 任务

- [x] 新增 `gfn_collector_observation` 表，使用 JSONB 保存协议 payload。
- [x] 为 observation 表增加必要索引，索引脚本必须可手动、可低峰执行、可重复检查。
- [x] 新增 v2 latest Redis key，例如 `collector:v2:latest:{protocol}:{site_id}`。
- [x] 旧表和旧 Redis key 继续写入，v2 只旁路双写。
- [x] 增加 collector v2 写入开关：observation 写入、v2 Redis 写入、按协议启停。
- [x] Ping / HTTP / DNS observation 按 `site_id + protocol` 保留指定条数。
- [x] `gfn_collector_domain` 增加 `site_id` 与 `deleted`，HTTP/DNS 过滤软删除目标。
- [x] Ping 从 `gfn_site.id + gfn_site.domain` 展开带 `site_id` 的内部目标。

#### 验收标准

- 关闭 v2 配置后，v1 行为完全不变。
- 当前 `/api/v1/nav` 展示链路、旧 Redis key、旧表写入不受影响。
- 可以通过 SQL / Redis / 日志人工确认 v2 写入量和写入频率。
- 本轮不要求后端 v2 接口上线，不改变 Nuxt 当前展示。

#### 备注

这一阶段只建立 collector 旁路数据面，不切换前端读取，不引入健康评分新展示。后端只读 v2 API 已记录到 `gofurry-nav-backend/docs/roadmap.md`，不纳入本轮实施。

---

### v0.2.1 - 采集域名与站点关系迁移闭环

- **状态：** 已完成
- **范围：** 数据模型 / Collector / Backend / gofurry-admin / 兼容迁移
- **目标：** 让 `gfn_collector_domain.site_id` 成为采集域名与站点的正式关联，解决 v2 observation 缺少 `site_id` 跳过的问题。

#### 重点

- `gfn_collector_domain` 成为 Ping / HTTP / DNS 的统一采集目标来源。
- backend 导航站点列表的 `domain` 字段改由 `gfn_collector_domain` 聚合生成，wire shape 保持兼容。
- gofurry-admin 的“采集域名”必须绑定站点，删除改为软删除。
- “网站”管理不再编辑历史域名字段，`gfn_site.domain` 已在生产验证稳定后物理移除。
- 提供手动同步脚本，将历史域名数据回填并补齐到 `gfn_collector_domain.site_id`。

#### 任务

- [x] gofurry-admin 采集域名模型、接口、表单增加 `site_id`。
- [x] gofurry-admin 采集域名删除从硬删除改为软删除。
- [x] 新增 `sql/20260523_collector_domain_site_sync.sql`，回填已有采集域名并补齐缺失域名。
- [x] Ping 采集目标从 `gfn_site.domain` 切换到 `gfn_collector_domain`。
- [x] Ping / HTTP / DNS 只采集未软删除、带 `site_id`、且所属站点未软删除的采集域名。
- [x] backend 导航页域名从 `gfn_collector_domain` 聚合，Nuxt 现有读取结构不变。
- [x] gofurry-admin 网站表单移除历史 `domains` 字段，不再写入 `gfn_site.domain`。
- [x] 新增 `sql/20260524_drop_gfn_site_domain.sql`，生产观察稳定后物理移除 `gfn_site.domain`。
- [x] 清理 admin/collector 中对 `gfn_site.domain` 物理列的模型和写入依赖。

#### 验收标准

- 新增或修改采集域名必须绑定站点。
- 软删除采集域名后 collector 不再采集该目标。
- 旧站点展示不因迁移中断，`domain` 返回结构保持兼容。
- v2 observation 不再因 HTTP/DNS 目标缺少 `site_id` 而大量跳过。
- `gfn_site.domain` 物理移除后，collector/backend/admin/Nuxt 均不依赖该列运行。

#### 后续观察

- 观察 admin 新增/修改采集域名是否能被 collector 下一轮采集自动接入。
- 继续通过 `gfn_collector_domain` 维护采集目标；如需回滚，需要先恢复旧列结构和历史备份数据。

---

### v0.3.0 - 协议语义增强

- **状态：** 计划中
- **范围：** Ping / HTTP / TLS / DNS / 安全展示
- **目标：** 在低强度采集前提下，让协议结果更准确、更不容易被误读。

#### 重点

- Ping 只作为辅助信号。
- HTTP 与 TLS 语义拆分。
- DNS 风险标记替代强结论。
- 外部内容限制长度并安全展示。

#### 任务

- [x] Ping v2 payload 增加 `icmp_status`、`loss_rate`、`avg_rtt_ms`、`error_code`、`duration_ms`。
- [ ] Ping 失败只影响健康评分，不单独判定站点 down。
- [ ] 评估 TCP connect fallback，但默认关闭。
- [x] HTTP v2 增加 redirect chain、content type、常见安全 header 是否存在。
- [ ] 评估 HEAD-first 模式，必须放在配置开关后，默认先保持当前 GET 行为。
- [x] HTTP title/meta/header 字符串限制长度，并确保后端/Nuxt 按纯文本展示。
- [x] TLS 拆分 `cert_collected` 与 `cert_verified`。
- [x] TLS 先尝试正常校验，失败后再受控地采集证书详情，并记录 `verify_error`。
- [ ] DNS 将 `hijacked` 语义替换为 `risk_flags`，例如 `private_ip`、`low_ttl`、`resolver_timeout`。
- [ ] DNS TXT、SPF、DMARC、CAA 只保存摘要或限长原文，不保存无限长外部内容。
- [ ] multi-resolver 只作为可选对比能力，默认关闭，不进入 Phase 0 默认路径。

#### 验收标准

- 新字段明确标记为 observation，不包装成最终判断。
- 默认采集频率、并发和探测强度不增加。
- 任一协议增强都可以按配置关闭。
- 前端展示外部字段时不渲染 HTML，不拼接不可信内容。

---

### v0.4.0 - 健康状态聚合与展示旁路

- **状态：** 计划中
- **范围：** Backend / Redis summary / Nuxt 展示 / 兼容性
- **目标：** 基于 v2 observation 生成站点健康摘要，但先旁路展示，不替换旧页面主逻辑。

#### 重点

- 多协议综合健康状态。
- summary Redis key。
- 后端 v2 summary 接口。
- Nuxt 安全展示。

#### 任务

- [ ] 设计健康状态：`healthy`、`warning`、`degraded`、`unknown`、`down`。
- [ ] 定义协议权重：HTTP 高，TLS/DNS 中高，Ping 低。
- [ ] 实现 `collector:v2:summary:{site_id}`，默认旁路写入。
- [ ] 后端增加只读 summary 接口，默认关闭。
- [ ] Nuxt 增加灰度展示入口，不影响当前主展示。
- [ ] 明确规则：HTTP 正常但 Ping 失败不能判 down。
- [ ] 明确规则：DNS 失败但 HTTP latest 仍可用时优先 `degraded` 或 `unknown`。
- [ ] 明确规则：TLS 临期优先 `warning`，已过期但 HTTP 可访问时优先 `degraded`。

#### 验收标准

- 健康状态有可解释规则和示例。
- 旧页面展示路径可随时恢复。
- 单个协议异常不会误伤站点整体状态。
- Nuxt 展示不渲染外部 HTML 字段。

---

### v0.5.0 - 调度治理与可选多实例

- **状态：** 计划中
- **范围：** 调度 / 分布式 / 可靠性
- **目标：** 只有在需要多个 collector 节点时，再引入任务 lease 和 collector 身份，避免重复采集。

#### 重点

- Redis lease。
- collector_id 与 job_id。
- 站点分片。
- 失败重试。

#### 任务

- [ ] 增加 `collector_id` 配置，并写入 v2 observation。
- [ ] 增加 `job_id`，用于追踪单轮采集。
- [ ] 设计 `collector:v2:lease:{protocol}`，必须带 TTL。
- [ ] 获取 lease 失败时跳过本轮，不允许无 TTL 分布式锁。
- [ ] 评估按站点 hash 分片，默认不开启。
- [ ] 设计轻量失败重试策略，避免对失败目标高频重试。
- [ ] 文档化单实例和多实例两种部署方式。

#### 验收标准

- 单实例部署不需要启用 lease。
- 多实例启用 lease 后不会重复采集同一协议任务。
- collector 异常退出后 lease 可依赖 TTL 自动释放。
- 任一分布式能力都可关闭并回退到单实例。

---

### v0.6.0 - 站点治理与采集透明度

- **状态：** 计划中
- **范围：** 产品治理 / 配置 / 文档 / 站点 owner 体验
- **目标：** 让采集行为更透明、更可控，符合朋友站点低侵入健康观测定位。

#### 重点

- 站点采集策略。
- opt-out。
- User-Agent。
- 对外说明。

#### 任务

- [ ] 为站点增加 probe policy：允许 Ping、HTTP、TLS、DNS、采集频率、是否使用代理。
- [ ] 支持站点 opt-out 或单协议关闭。
- [ ] 为 collector HTTP 请求设置明确 User-Agent，包含项目名和联系页。
- [ ] 对外说明采集目的、协议、频率和退出方式。
- [ ] 支持站点 owner 查看基础采集记录。
- [ ] 后端管理侧增加策略修改入口，避免直接改数据库。

#### 验收标准

- 单站点可以降低频率或关闭指定协议。
- 采集行为有公开说明和联系渠道。
- User-Agent 不伪装浏览器，也不隐藏项目身份。
- 站点策略变更不影响其他站点。

---

### v1.0.0-alpha.1 - 稳定版候选

- **状态：** 计划中
- **范围：** 发布 / 文档 / 兼容性 / 回滚
- **目标：** 冻结 v2 payload 和后端接口候选结构，准备进入稳定运行前的反馈期。

#### 重点

- schema version。
- 灰度与回滚文档。
- v1 兼容性测试。
- 运维手册。

#### 任务

- [ ] 文档化 v2 payload schema version 和兼容策略。
- [ ] 文档化生产灰度、切读、回滚流程。
- [ ] 增加 v1 兼容性回归测试。
- [ ] 增加 v2 latest / summary 读取回归测试。
- [ ] 整理生产运维手册：配置、日志、Redis key、SQL 查询、故障排查。
- [ ] 收集生产反馈后再决定是否进入 `v1.0.0`。

#### 验收标准

- v1 行为在 alpha 阶段仍然完整可用。
- 任一 v2 协议或读取路径都可以通过配置关闭。
- schema 和接口有版本说明。
- 进入 `v1.0.0` 前不存在已知阻塞级稳定性问题。

## 短中长期方向

### 短期

- 保持 Phase 0 当前生产配置稳定运行。
- 开始 v2 数据旁路设计，但不切换后端/Nuxt 读取路径。
- 继续观察单目标失败、DNS timeout 和日志保留效果。

### 中期

- 完成 v2 observation、v2 latest、只读 v2 接口。
- 增强协议语义，但全部放在开关后。
- 建立 summary 旁路展示，避免单协议误判站点健康。

### 长期

- 只有出现多节点需求时才做 lease 和分片。
- 完成站点治理、opt-out 和公开采集说明。
- 冻结 v2 schema 后进入 `v1.0.0-alpha.1`。
