# GoFurry Nav Collector Roadmap

## 当前状态

`gofurry-nav-collector` 既定基础 roadmap 已完成。当前文档不再保留过长历史完成项，只记录后端正式进入 v2 前需要补齐的 collector 能力。

`v0.5.1` 来自 2026-05-25 代码审计，已完成稳定性补丁；`v0.5.2` 已补齐 v2 schema 收口前的低风险字段；`v0.5.3` 已完成默认关闭的域名治理低频轻探测；`v0.5.4` 已完成默认关闭的页面资源声明轻探测；`v0.5.5` 已完成默认关闭的 TCP 端口连通性轻探测；`v0.5.6` 已完成 WAF / CDN 被动指纹推断。后续 `v0.5.7` 用于在不影响旧链路的前提下补齐少量明确授权的轻探测能力。

## 迭代原则

- 不增加默认采集频率、请求次数、DNS 查询目标或并发强度。
- 不改变旧 Redis key、旧日志表、旧后端接口和旧前端展示结构。
- 优先修复会影响数据可信度、错误可见性、运行稳定性的点。
- 对外部站点返回的数据继续按不可信内容处理，所有新增说明都只作为 observation 信号。
- 继续避免不必要复杂度；不引入 Prometheus 生态，不做漏洞扫描，不做高频探测。
- 新增会对外发起额外请求或连接的能力必须默认关闭或低频，并具备按站点、按协议关闭能力。
- 端口探测、WAF 规则验证等主动安全能力只允许对自有或明确授权目标启用，不对普通收录站点默认执行。

## 版本计划

### v0.5.1 - 审计发现修复与稳定性补丁

**状态：** 已完成
**范围：** 稳定性 / 数据可信度 / Redis 错误语义 / 测试
**目标：** 修复 2026-05-25 代码审计发现的问题，让 collector 在现有单实例生产模式下更可靠、更容易排查。

#### Focus

- DNS 递归记录预算和 IPv6 风险信号。
- Redis 写入、lease、Ping 目标刷新错误语义。
- v2 summary 聚合热路径收敛。
- HTTP 与旧 JSON 写入的数据可信度补强。
- 不改变采集强度和旧业务链路。

#### Tasks

- [x] 为 DNS CNAME / MX / NS 递归 children 加入全局记录预算，避免 `max_dns_records_per_query` 只限制父级 Answer。
- [x] 为 DNS v2 `private_ip` 风险标记补齐 IPv6 特殊地址识别，例如 `::1`、`fc00::/7`、`fe80::/10`。
- [x] 优化 site summary 聚合路径，避免每次 observation 写入都用 Redis `SCAN` 全量扫 `collector:v2:summary:target:{site_id}:*`。
- [x] 将 Redis `SetNX` wrapper 改为返回 `created bool` 和 `GFError`，让 lease 获取失败能区分“其他实例持有”和“Redis 命令失败”。
- [x] 修复 Ping 目标列表刷新时忽略 `SetNX` 结果的问题；刷新失败必须记录清晰日志并返回旁路错误。
- [x] HTTP 读取响应体时显式记录是否被 `max_response_bytes` 截断，避免把截断 HTML 当作完整页面语义。
- [x] 检查 Ping / HTTP / DNS 旧 JSON 写入路径中被忽略的 `sonic.Marshal` 错误，至少记录中文结构化日志。
- [x] 为 DNS GeoIP / PTR 缓存增加边界或清理策略，避免长期运行中缓存无限增长。
- [x] 为上述修复补充小范围单元测试，不扩大生产默认行为。

#### Acceptance Criteria

- `gofmt -l .` 为空。
- `go test ./...` 通过。
- `go vet ./...` 通过。
- DNS payload 在异常递归链路下仍受预算保护。
- Redis lease 日志能明确区分锁被占用和 Redis 写入失败。
- Ping 目标刷新失败不再静默吞掉。
- HTTP v2 payload 能标记 body 是否截断，旧 Redis/旧表结构保持兼容。
- 生产未修改配置时，采集频率、并发、旧 Redis key、旧表写入、后端接口和前端展示保持不变。

#### Notes

- 审计报告：`docs/code-audit-20260525.md`。
- 本阶段只做补丁修复，没有增加采集频率、探测次数或默认协议强度。

---

### v0.5.2 - v2 Schema 收口前的低风险字段补充

**状态：** 已完成
**范围：** Ping / DNS / TLS / HTTP / v2 observation
**目标：** 在不增加请求次数和探测强度的前提下，把现有一次采集中已经能拿到的高价值字段补齐，为后端 v2 schema 稳定做准备。

#### Focus

- TLS 证书身份与指纹。
- HTTP 页面/传输的完整性和访客体感字段。
- DNS 解析结构摘要。
- Ping 选择 IP 的解释信息。

#### Tasks

- [x] TLS v2 payload 增加 `cert_serial_number`、`cert_fingerprint_sha256`、`cert_spki_sha256`。
- [x] TLS v2 payload 增加 `cert_public_key_bits`、`cert_subject_org`、`cert_issuer_cn`、`cert_chain_issuers`。
- [x] HTTP v2 payload 增加 `content_length_header`、`transfer_encoding`、`is_chunked`、`html_charset`、`doctype`。
- [x] HTTP v2 payload 增加 `robots_meta_policy`、`canonical_host_matches_final_host`、`compression_ratio_estimated`。
- [x] DNS v2 payload 增加 `has_a`、`has_aaaa`、`ipv4_count`、`ipv6_count`、`cname_terminal`。
- [x] DNS v2 payload 增加 `name_server_hosts`、`mx_hosts`、`ttl_spread`、`mixed_private_public_ip`。
- [x] Ping v2 payload 增加 `ip_family`、`resolved_ips`、`selected_ip`、`resolution_source`、`icmp_blocked_suspected`。
- [x] 更新 v2 observation payload 文档，给出字段边界和样例 JSON。

#### Acceptance Criteria

- 不新增 HTTP 请求、DNS 查询目标、Ping 次数或默认并发。
- 旧 Redis key、旧日志表、旧前端展示结构不变化。
- 新字段全部有 builder 级单元测试。
- 新字段只作为 observation 信号，不作为健康结论。

#### Notes

- 字段说明：`docs/v2-observation-payload.md`。
- 本阶段没有新增 SQL、配置项、请求次数或默认并发。

---

### v0.5.3 - 域名治理低频轻探测

**状态：** 已完成
**范围：** RDAP / robots.txt / security.txt / 低频任务
**目标：** 为站点提供域名生命周期、爬虫策略和安全联系渠道参考，默认低频，失败只记录旁路错误类别。

#### Focus

- 低频域名信息。
- 明确缓存和退避。
- 不抓取大量外部内容。

#### Tasks

- [x] 增加 RDAP 低频探测，默认周期建议 7 天，记录 registrar、domain status、expires_at、nameservers、dnssec delegation 状态。
- [x] RDAP 失败时记录错误类别，不影响 HTTP/DNS/Ping 主链路。
- [x] 增加 robots.txt 低频探测，记录是否存在、状态码、sitemap 链接数量、是否存在全站 disallow。
- [x] 增加 security.txt 低频探测，只检查 `/.well-known/security.txt` 和 `/security.txt`，记录是否存在、Contact、Expires、Policy 摘要。
- [x] 为 RDAP / robots.txt / security.txt 增加独立配置开关、超时、低频 interval 和 latest Redis key。
- [x] 文档化这些结果不是站点安全结论，只是治理参考。

#### Acceptance Criteria

- 默认不会增加现有 Ping / HTTP / DNS / TLS 采集强度。
- 每类轻探测都有独立关闭开关和超时。
- 不抓取 sitemap 明细，不扫描目录，不解析或保存大体积文本。
- v2 observation / latest Redis 能查询到低频探测结果。

#### Notes

- 配置位于 `collector.v2.light_probe`，三类探测默认全部关闭。
- latest Redis key 继续复用 v2 latest 规则：`collector:v2:latest:{protocol}:{site_id}` 与 `collector:v2:latest:{protocol}:{site_id}:{target}`。
- 字段说明：`docs/v2-observation-payload.md`。

---

### v0.5.4 - 页面资源声明轻探测

**状态：** 已完成
**范围：** favicon / manifest / 前端展示参考 / 低频任务
**目标：** 在已有 HTML 声明的基础上低频补充页面资源元信息，为站点详情页提供更完整的视觉和 PWA 参考。

#### Focus

- 只跟随 HTML 中明确声明的资源。
- 限制大小、数量和 Content-Type。
- 不枚举子资源。

#### Tasks

- [x] 基于现有 HTTP HTML 提取到的 icon links，低频拉取最多 1 个优先 favicon，只记录状态、Content-Type、大小、hash，不保存原始图片。
- [x] 基于 HTML 中声明的 manifest link，低频拉取 manifest，记录 `name`、`short_name`、`theme_color`、`background_color`、icons count、display。
- [x] 增加资源响应体大小上限、Content-Type allowlist、redirect 限制和超时。
- [x] 增加 `collector.v2.light_probe.page_assets` 开关，默认关闭或低频。
- [x] 对 manifest/icon URL 做同站或明确允许的跨站限制，避免任意第三方资源跟随。
- [x] 更新文档，说明本阶段不做截图、OCR、页面正文抓取或子资源枚举。

#### Acceptance Criteria

- 默认不拉取页面所有资源。
- 单站点单轮最多新增极少量请求，并可完全关闭。
- 不保存图片、HTML 正文或 manifest 原文大字段。
- 结果可作为站点详情页展示参考，但不参与健康判定。

#### Notes

- `page_assets` 只消费 HTTP v2 latest 中已存在的 `icon_links` / `manifest_link`，不会为了资源声明重新抓取首页。
- 字段说明：`docs/v2-observation-payload.md`。

---

### v0.5.5 - 授权范围内的高价值端口轻探测

**状态：** 已完成
**范围：** TCP connect / 授权目标 / 安全边界
**目标：** 对自有或明确授权站点做少量高价值端口连通性观测，帮助发现意外暴露或服务变更；不对普通收录站点默认执行。

#### Focus

- 默认关闭。
- 开启即视为当前有效目标已获得授权。
- 端口必须显式配置，空列表不探测。
- 只做 TCP connect，不抓 banner，不发协议 payload。

#### Tasks

- [x] 增加 `collector.v2.light_probe.port_check.enabled=false`，默认不产生任何端口探测。
- [x] 端口列表完全由配置 `ports` 显式提供；代码不内置默认扫描端口。
- [x] 每个目标每轮限制端口数量、并发和超时，失败不重试。
- [x] 只记录 `open` / `closed` / `timeout` / `filtered_suspected` / `skipped`、connect 耗时和错误类别。
- [x] 明确禁止 banner grabbing、协议交互、弱口令、目录枚举和漏洞验证。
- [x] 为端口探测增加中文日志、run state、v2 observation payload 和测试。

#### Acceptance Criteria

- 未启用 `port_check` 或未配置端口列表时不会发起 TCP 连接。
- 端口探测失败不影响 Ping / HTTP / DNS / TLS 主链路。
- 可通过配置一键关闭全部端口探测。
- 文档明确适用范围和法律/授权边界。

#### Notes

- `port_check.enabled=true` 表示维护者已确认当前 collector 有效目标处于授权探测范围内。
- Prometheus / Grafana 只作为端口名提示，不接入 Prometheus 生态，不抓取 metrics，不访问 Web 页面。
- 字段说明：`docs/v2-observation-payload.md`。

---

### v0.5.6 - WAF / CDN 被动指纹推断

**状态：** 已完成
**范围：** HTTP headers / DNS / TLS / 被动推断
**目标：** 基于已有采集结果做保守的 WAF/CDN 线索识别，帮助站点理解访问链路，但不把推断包装成确定事实。

#### Focus

- 被动推断优先。
- 证据字段可解释。
- 不新增攻击性请求。

#### Tasks

- [x] 基于已有 HTTP header、DNS CNAME/NS/ASN、TLS issuer/SAN、server hints 建立保守规则。
- [x] 输出 `edge_provider_hints`，包含 provider、hint_type、confidence、evidence，不输出绝对结论。
- [x] 区分 CDN、WAF、反代、对象存储、托管平台等不同 hint 类型。
- [x] 对 Cloudflare、Tencent Cloud、Aliyun、AWS CloudFront、Fastly、Vercel、Netlify、GitHub Pages 等常见服务做基础规则。
- [x] 规则表内置但保持简单，可通过 `collector.v2.edge_hints.enabled=false` 禁用。
- [x] 增加测试，保证证据不足时返回空 hint 或 low confidence。

#### Acceptance Criteria

- 不额外请求目标站点。
- 不影响现有健康状态和前端主展示。
- 每个 hint 都能看到 evidence，避免误读。
- 文档明确“推断不是事实”，需要人工确认。

#### Notes

- 本阶段没有新增 HTTP/DNS/Ping/light probe 请求，只消费已有 v2 latest payload。
- 字段说明：`docs/v2-observation-payload.md`。

---

### v0.5.7 - 自有站点轻量 WAF 规则无害验证

**状态：** 计划中
**范围：** 自有站点 / WAF canary / 安全测试 / 默认关闭
**目标：** 只对 GoFurry 自有站点或明确授权的专用测试目标，验证基础 WAF 规则是否仍能拦截常见无害测试请求；不对导航收录的第三方站点执行。

#### Focus

- 默认关闭。
- 只允许自有 allowlist。
- 建议使用专用 canary 路径。
- 只验证拦截行为，不做漏洞利用。

#### Tasks

- [ ] 增加 `collector.v2.light_probe.waf_canary.enabled=false`，并要求配置 `allowed_hosts` 和 `canary_path`。
- [ ] WAF 测试请求必须带清晰 User-Agent，例如 `GoFurry-Nav-Collector-WAF-Canary`，便于日志识别。
- [ ] 测试请求只打到无业务副作用的 canary 路径，不携带账号、Cookie、Token 或真实用户数据。
- [ ] 按规则类别验证基础拦截：扫描器 UA、JSON 解析错误、参数数量、异常 URI 长度、SQLi 关键词、XSS 标记、命令注入标记、路径穿越标记、非法 HTTP 方法、非法 Content-Type、JSON 危险关键字。
- [ ] 只记录状态码、是否命中预期、规则类别、耗时和错误类别，不保存完整攻击样本。
- [ ] 增加严格频率限制，建议每天或手动触发，不做失败重试。
- [ ] 如果目标 host 不在 allowlist 或 canary_path 为空，任务必须拒绝启动并记录中文错误日志。
- [ ] 文档化该能力只适合自有 WAF 回归测试，不属于导航站点常规采集能力。

#### Acceptance Criteria

- 默认配置下不会发送任何 WAF 测试请求。
- 不能对非 allowlist host 执行。
- 测试结果只进入 v2 observation / latest Redis，不影响旧页面和站点健康结论。
- 所有测试 payload 都保持无害、可解释、可在 WAF 日志中识别。
- 误配置时安全失败，不发起请求。
