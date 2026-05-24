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

### v0.3.0 - v2 Observation 协议语义增强

- **状态：** 已完成
- **范围：** Ping / HTTP / TLS / DNS / 安全展示
- **目标：** 在不改变采集强度和旧展示链路的前提下，让 v2 observation 的协议 payload 更准确、更不容易被误读。

#### 重点

- Ping 只作为辅助信号。
- HTTP 与 TLS 语义拆分。
- DNS 风险标记替代强结论。
- 外部内容限制长度并安全展示。

#### 任务

- [x] Ping v2 payload 增加 `icmp_status`、`loss_rate`、`avg_rtt_ms`、`error_code`、`duration_ms`。
- [x] HTTP v2 增加 redirect chain、content type、常见安全 header 是否存在。
- [x] HTTP title/meta/header 字符串限制长度，并确保后端/Nuxt 按纯文本展示。
- [x] TLS 拆分 `cert_collected` 与 `cert_verified`。
- [x] TLS 先尝试正常校验，失败后再受控地采集证书详情，并记录 `verify_error`。
- [x] DNS 将 `hijacked` 语义替换为 `risk_flags`，例如 `private_ip`、`low_ttl`、`nxdomain_with_answer`、`ptr_empty`。
- [x] DNS TXT、SPF、DMARC、CAA 只保存摘要或限长原文，不保存无限长外部内容。

#### 验收标准

- 新字段明确标记为 observation，不包装成最终判断。
- 默认采集频率、并发和探测强度不增加。
- 任一协议增强都可以按配置关闭。
- 前端展示外部字段时不渲染 HTML，不拼接不可信内容。

#### 后续归位

- Ping 是否影响站点整体 `down` 状态归入 v0.4.0 健康状态聚合，不在 v0.3.0 单协议 payload 阶段下结论。
- TCP connect fallback、HTTP HEAD-first、DNS multi-resolver 都会改变探测行为，归入后续可选协议能力，默认不启用。

---

### v0.3.1 - Collector 可靠性审计修复

- **状态：** 已完成
- **范围：** DNS / Ping / Redis / DB / 配置 / 日志
- **目标：** 修复 2026-05-24 代码审计中发现的可靠性和数据可信度问题，不增加采集强度，不改变旧接口、旧 Redis key、旧表语义。

#### 重点

- 先修正影响 observation 数据可信度的问题。
- 让旧表、Redis、DB 初始化错误更容易被发现。
- 收紧配置语义，减少“配置写了但不生效”的维护陷阱。

#### 任务

- [x] 修复 DNS GeoIP reader 参数顺序，确保 Country / City / ASN reader 不被交换。
- [x] 为 DNS GeoIP 查询增加最小单元测试，覆盖 nil reader 降级和查询并发限制。
- [x] Ping 旧表写入检查错误并记录中文结构化日志。
- [x] 缩小 Ping `pingRWLock` 锁范围，只保护共享 Redis 结果 map 更新。
- [x] 修复 Redis `HDel` 吞错问题，删除失败必须向调用方返回错误。
- [x] 检查 Redis wrapper 中类似 `Get` / `Incr` 的错误处理边界，避免未来误用。
- [x] HTTP 代理 URL 解析失败时明确记录 `http_proxy_config_invalid`，不静默继续。
- [x] DB 初始化处理 `db.engine.DB()` 返回错误，并优化 DSN 构造/转义边界。
- [x] 让 `collector.dns.query_thread` 实际约束单目标 DNS 记录类型查询并发。
- [x] 修正 `server.mode` yaml tag，减少配置加载阶段绕过 zap 的 stdout 输出。
- [x] 为 retention SQL builder 增加表名 allowlist，避免未来误传任意表名。
- [x] 为 observation 写入增加中心化 payload 大小保护，防止后续字段扩展时 Redis/DB 负载失控。

#### 验收标准

- 修复后 `go test ./...`、`go vet ./...` 通过。
- DNS A / AAAA 的 Country / ASN / ISP 字段在测试环境恢复可信。
- Redis 删除失败、Ping 旧表写入失败、HTTP 代理配置错误都能从中文结构化日志中明确定位。
- 未配置新行为时，生产采集频率、并发、旧 Redis key、旧表写入和旧前端展示保持不变。

#### 审计报告

详见 `docs/code-audit-20260524.md`。

---

### v0.3.2 - v2 Observation 低风险信息丰富

- **状态：** 已完成
- **范围：** Ping / HTTP / TLS / DNS
- **目标：** 在不增加默认探测次数和采集强度的前提下，把当前三种协议同一次探测中已经能得到的高价值字段写入 v2 observation，为站点安全参考和访客视角参考提供更丰富依据。

#### 重点

- 只采集“同一次 Ping / HTTP GET / DNS 查询中顺手可得”的字段。
- 字段只作为 observation 信号，不直接包装成健康结论。
- 外部文本继续限长、摘要化、按纯文本处理。
- 默认不启用 HEAD-first、TCP fallback、multi-resolver、额外 `_dmarc` 查询等会改变探测行为的能力。

#### Ping 可扩展字段

- [x] 增加 `min_rtt_ms`、`max_rtt_ms`、`stddev_rtt_ms`。
- [x] 增加 `jitter_ms`，先以 RTT 标准差作为保守近似。
- [x] 增加 `packets_sent`、`packets_recv`、`packets_recv_duplicates`，让 `loss_rate` 更可解释。
- [x] 记录本轮 Ping 实际 `resolved_ip`，辅助对照 DNS 和访客访问路径。

#### HTTP 可扩展字段

- [x] 使用 `httptrace` 记录 `dns_lookup_ms`、`tcp_connect_ms`、`tls_handshake_ms`、`ttfb_ms`、`transfer_ms`。
- [x] 记录 `http_protocol`，区分 HTTP/1.1、HTTP/2 等实际协议。
- [x] 记录最终连接 `remote_ip` / `remote_addr`，辅助和 DNS observation 对照。
- [x] 记录 `content_length`、`body_read_bytes`、`compressed`，作为访客体感大小参考。
- [x] 摘要化 `cache-control`、`etag`、`last-modified`，形成 `cache_policy`。
- [x] 细化常见安全 header 摘要，例如 HSTS `max-age`、CSP 是否包含高风险宽松项。

#### TLS 可扩展字段

- [x] 增加 `cert_not_before`、`cert_not_after`、`cert_chain_length`。
- [x] 增加 `cert_subject_cn`、`cert_san_count`。
- [x] 增加 `cert_signature_algorithm`、`cert_public_key_algorithm`。
- [x] 增加 `ocsp_stapled`、`sct_count`，作为证书透明度和吊销响应参考。
- [x] 增加 `verify_error_category`，把原始校验错误归类为过期、域名不匹配、未知 CA 等。

#### DNS 可扩展字段

- [x] 增加 `rcode`、`authoritative`、`truncated`、`recursion_available`。
- [x] 增加 `answer_count`、`authority_count`、`additional_count`。
- [x] 将已有 TTL 统计整理为 `ttl_min`、`ttl_max`、`ttl_avg` 写入 v2 payload。
- [x] 增加 `cname_chain_depth`，表达解析链复杂度。
- [x] 摘要化 MX 优先级、SOA serial/refresh/retry/expire/minTTL。
- [x] 区分 `dnssec_rrsig_present` 与 `dnssec_ad`，避免把“有签名记录”和“已验证”混为一谈。

#### 验收标准

- 所有新增字段都有单元测试或 payload builder 测试。
- 不增加默认请求次数、DNS 目标数量、Ping 次数、并发或响应体上限。
- v2 latest Redis 和 observation DB 新字段出现，旧 Redis key、旧表和旧前端展示不变化。
- 字段解释写入 `docs/v2-observation-payload.md`，明确它们不是最终健康判断。

---

### v0.3.3 - HTTP 页面语义与展示信息丰富

- **状态：** 已完成
- **范围：** HTTP / 页面语义 / 访客参考 / SEO 线索
- **目标：** 在不增加额外请求和默认探测强度的前提下，让站点详情页能展示更完整的页面语义、分享信息、基础 SEO 信号和访客参考信息。

#### 重点

- 只使用同一次 HTTP GET 已拿到的响应头和响应体，不额外请求 favicon、manifest、截图或子资源。
- 优先补充“前端详情页直接能看懂”的字段，而不是继续堆底层协议细节。
- 外部文本继续限长，HTML 只做纯文本提取，不在 collector 端保留原始标签块。
- 仍然保持旧 Redis key、旧表结构兼容；新增字段同步进入 v2 observation 和旧 HTTP JSON，后端现有透传接口不需要改动。

#### 可扩展 meta / 页面字段

- [x] 增加 `meta.author`、`meta.generator`、`meta.application_name`、`meta.theme_color`、`meta.robots`。
- [x] 增加 Open Graph 摘要：`og.title`、`og.description`、`og.site_name`、`og.type`、`og.image`、`og.url`。
- [x] 增加 Twitter Card 摘要：`twitter.card`、`twitter.title`、`twitter.description`、`twitter.image`、`twitter.site`。
- [x] 增加 `canonical_url`、`html_lang`、`viewport`。
- [x] 增加 `meta_refresh` 摘要，区分正常页面和前端/站点级跳转页。
- [x] 增加 `icon_links` 摘要：只记录页面中声明的 `icon` / `shortcut icon` / `apple-touch-icon` / `manifest` 链接，不额外拉取资源。

#### 可扩展响应头字段

- [x] 扩展保留响应头白名单，增加 `Date`、`Last-Modified`、`ETag`、`Cache-Control`、`Vary`、`Content-Encoding`、`Content-Language`、`X-Robots-Tag`、`Link`、`Alt-Svc`、`X-Powered-By`。
- [x] 增加 `cookie_summary`：`set_cookie_count`、`secure_count`、`http_only_count`、`same_site_lax_count`、`same_site_strict_count`、`same_site_none_count`。
- [x] 增加跨域与隔离相关 header 摘要：`cross_origin_opener_policy`、`cross_origin_embedder_policy`、`cross_origin_resource_policy`、`access_control_allow_origin`。
- [x] 增加 `server_hints`，对 `server`、`x-powered-by`、`generator` 做保守归并，方便详情页展示“可能的技术栈线索”。

#### 访客视角参考字段

- [x] 增加 `content_language_effective`，优先响应头，缺失时回退 `html lang`。
- [x] 增加 `page_text_summary`：页面标题、描述、关键词、站点名的合并纯文本摘要，便于后端或前端做摘要展示。
- [x] 增加 `share_preview`：从 OG/Twitter 中整理可直接用于详情页的小卡片字段。
- [x] 增加 `redirect_hint`，对 meta refresh、canonical、final URL 差异做保守提示，不下结论。

#### 验收标准

- 新增字段全部来自同一次 HTTP GET，不增加额外网络请求。
- `go test ./...`、`go vet ./...` 通过，并补充 payload builder 单元测试。
- 外部文本字段全部有统一限长策略，URL 字段统一做长度限制。
- v2 observation 能查询到新增字段，旧 `request:{domain}` 和旧 `gfn_collector_log_http` 兼容结构不被破坏。
- 规划出的字段能直接支持前端详情页展示更多“页面信息 / 分享信息 / SEO 信号 / 头部策略”。

#### 备注

这一阶段优先解决“详情页信息太少”，所以重点是页面语义和访客参考，不做截图、截图 OCR、favicon 下载、manifest 下载、子资源枚举，也不把技术栈猜测包装成确定事实。

---

### v0.4.0 - 健康状态聚合与展示旁路

- **状态：** 已完成
- **范围：** Backend / Redis summary / Nuxt 展示 / 兼容性
- **目标：** 基于 v2 observation 生成站点健康摘要，但先旁路展示，不替换旧页面主逻辑。

#### 重点

- 多协议综合健康状态。
- target 级和 site 级 summary Redis key。
- 暂不接入后端 v2 summary 接口。
- 暂不改 Nuxt 主展示链路。

#### 任务

- [x] 设计健康状态：`healthy`、`warning`、`degraded`、`unknown`、`down`。
- [x] 定义协议权重：HTTP 高，TLS/DNS 中高，Ping 低。
- [x] 实现 `collector:v2:summary:target:{site_id}:{target}`，保留多采集域名明细。
- [x] 实现 `collector:v2:summary:site:{site_id}`，由 target summary 保守聚合。
- [x] 保留原 `collector:v2:latest:{protocol}:{site_id}`，额外增加 `collector:v2:latest:{protocol}:{site_id}:{target}`。
- [x] 明确规则：HTTP 正常但 Ping 失败不能判 down。
- [x] 明确规则：Ping 失败只作为低权重辅助信号，不单独判定站点 down。
- [x] 明确规则：DNS 失败但 HTTP latest 仍可用时优先 `warning`，不判 down。
- [x] 明确规则：TLS 临期优先 `warning`，已过期或证书校验失败但 HTTP 可访问时优先 `degraded`。
- [x] 明确规则：HTTP observation 缺失或过期时为 `unknown`。

#### 验收标准

- 健康状态有可解释规则和单元测试覆盖。
- 旧页面展示路径可随时恢复。
- 单个协议异常不会误伤站点整体状态。
- 本轮不改 Nuxt，因此不存在外部 HTML 渲染入口。

#### 后续归位

- 后端 `/api/v2/nav` 只读 summary 接口和 Nuxt 灰度展示入口归入下一段实施，避免一次性扩大改动面。

---

### v0.4.1 - Summary 只读接口与灰度展示

- **状态：** 计划中
- **范围：** Backend / Nuxt 展示 / 兼容性
- **目标：** 在 collector summary 生产观察稳定后，再把健康摘要通过只读接口和灰度 UI 暴露出来。

#### 重点

- 后端只读 summary 接口。
- Nuxt 灰度展示入口。
- 旧页面主展示不替换。

#### 任务

- [ ] 后端增加只读 summary 接口，默认关闭。
- [ ] Nuxt 增加灰度展示入口，不影响当前主展示。
- [ ] 明确 summary 外部字段全部按纯文本展示，不渲染 HTML。
- [ ] 为接口增加 Redis 缺失、summary 过期、目标不存在等回退语义。

#### 验收标准

- 未开启灰度时前端展示不变化。
- Redis summary 缺失时接口返回 `unknown` 或空摘要，不影响旧接口。
- 前端能解释 `healthy`、`warning`、`degraded`、`unknown`、`down` 的含义。

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

### v0.7.0 - 可选协议能力评估

- **状态：** 计划中
- **范围：** Ping / HTTP / DNS / 配置 / 灰度
- **目标：** 只在 v2 observation 和 summary 稳定后，再评估会改变探测行为的可选协议能力。

#### 重点

- 默认关闭。
- 可按协议灰度。
- 不增加默认采集强度。
- 任何能力都必须有回滚路径。

#### 任务

- [ ] 评估 TCP connect fallback，但默认关闭。
- [ ] 评估 HEAD-first 模式，必须放在配置开关后，默认继续保持当前 GET 行为。
- [ ] 评估 DNS multi-resolver 对比能力，默认关闭，不进入基础生产路径。
- [ ] 文档化每个可选能力的适用场景、误判风险和关闭方式。

#### 验收标准

- 未配置时生产行为与当前版本一致。
- 任一可选能力都可以单独关闭。
- 不把可选探测结果直接包装成最终健康结论。
- 生产灰度前必须有旧链路对照和回滚说明。

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
- 在生产继续观察 v2 observation 新字段的体积、可读性和查询价值。
- 不切换后端/Nuxt 读取路径，先积累一段时间旁路数据。

### 中期

- 完成 v2 observation、v2 latest、只读 v2 接口。
- 基于已完成的协议语义增强建立 summary 旁路展示，避免单协议误判站点健康。

### 长期

- 只有出现多节点需求时才做 lease 和分片。
- 只在 summary 稳定后评估 HEAD-first、TCP fallback、multi-resolver 等可选能力。
- 完成站点治理、opt-out 和公开采集说明。
- 冻结 v2 schema 后进入 `v1.0.0-alpha.1`。
