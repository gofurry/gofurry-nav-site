# v2 Observation Payload 说明

本文记录 `gofurry-nav-collector` 当前已经写入 `gfn_collector_observation.payload` 与 v2 latest Redis key 的协议字段。

这些字段都是 observation 信号，只表示一次采集观察到的事实或风险提示，不是站点最终健康结论。旧 Redis key、旧日志表、旧 Nuxt 展示仍是兼容路径：`ping:result`、`request:{domain}`、`dns:{domain}`、`gfn_collector_log_ping`、`gfn_collector_log_http`、`gfn_collector_log_dns` 不因本文字段改变。

## 通用字段

`gfn_collector_observation` 表外层字段包括：

- `site_id`：站点 ID。
- `target`：本次采集目标。
- `protocol`：`ping`、`http`、`dns`。
- `status`：本次协议采集状态。
- `observed_at`：采集时间。
- `duration_ms`：本次协议探测墙钟耗时。
- `error_code` / `error_message`：采集失败或降级原因。
- `payload`：协议细节 JSON。
- `schema_version`：当前为 `1`。

v2 latest Redis key 为：

- `collector:v2:latest:ping:{site_id}`
- `collector:v2:latest:http:{site_id}`
- `collector:v2:latest:dns:{site_id}`

## Ping Payload

Ping v2 payload 当前包含：

- `icmp_status`：`reachable` 或 `unreachable`。
- `avg_rtt_ms`：平均 RTT，失败时为 `null`。
- `min_rtt_ms`、`max_rtt_ms`、`stddev_rtt_ms`：同一次 Ping 的 RTT 范围和抖动统计，失败时为 `null`。
- `jitter_ms`：当前先以 RTT 标准差作为保守近似。
- `loss_rate`：丢包率。
- `packets_sent`、`packets_recv`、`packets_recv_duplicates`：本轮发包、收包、重复包数量。
- `resolved_ip`：本轮 Ping 解析并实际访问到的 IP。
- `duration_ms`：单目标 Ping 探测墙钟耗时。
- `error_code`：`ping_init_failed`、`ping_run_failed`、`ping_unreachable` 或空字符串。
- 兼容字段：`delay_ms`、`legacy_delay`、`legacy_loss`、`legacy_status`。

Ping 只作为辅助信号；是否影响站点整体健康状态由后续 summary 规则决定。

## HTTP / TLS Payload

HTTP v2 payload 当前包含：

- HTTP 基础字段：`url`、`status_code`、`response_time_ms`、`content_length`、`title`、`server`、`headers`、`meta`。
- 重定向字段：`redirects`、`redirect_chain`、`redirect_count`、`final_url`。
- 响应类型：`content_type`。
- 连接和下载阶段字段：`dns_lookup_ms`、`tcp_connect_ms`、`tls_handshake_ms`、`ttfb_ms`、`transfer_ms`。
- 传输上下文字段：`http_protocol`、`remote_addr`、`remote_ip`、`body_read_bytes`、`compressed`、`content_encoding`。
- 缓存摘要：`cache_policy.cache_control`、`cache_policy.etag`、`cache_policy.last_modified`。
- 常见安全响应头是否存在：`security_headers`。
- 常见安全响应头摘要：`security_header_summary`。
- 页面语义字段：`canonical_url`、`html_lang`、`meta_refresh`、`icon_links`。
- 页面 meta 扩展：`meta.author`、`meta.generator`、`meta.application_name`、`meta.theme_color`、`meta.robots`、`meta.viewport`。
- 分享信息摘要：`open_graph`、`twitter_card`、`share_preview`。
- 响应头补充摘要：`cookie_summary`、`server_hints`、`cross_origin_summary`、`content_language_effective`。
- 访客视角提示：`page_text_summary`、`redirect_hint`。
- TLS 兼容字段：`tls_version`、`cipher_suite`、`cert_expiry`、`cert_days_left`、`cert_issuer`、`cert_issuer_org`、`cert_dns_names`、`cert_pub_key_alg`、`cert_sig_alg`、`cert_email`、`cert_is_ca`。
- TLS 语义字段：`cert_collected`、`cert_verified`、`verify_error`、`verify_error_category`、`tls_handshake`。
- TLS 补充字段：`cert_not_before`、`cert_not_after`、`cert_chain_length`、`cert_subject_cn`、`cert_san_count`、`cert_signature_algorithm`、`cert_public_key_algorithm`、`ocsp_stapled`、`sct_count`。

HTTP 仍使用 GET。v0.3.0 不启用 HEAD-first，也不改变重定向、超时、响应体上限和旧展示结构。

`remote_addr` / `remote_ip` 表示本次 TCP 实际连到的对端，目标走代理时它可能是代理地址，不应被直接当成源站结论。

`security_header_summary` 当前会对这些 header 做保守结构化摘要：

- `hsts`：`present`、`max_age`、`include_subdomains`、`preload`
- `content_security_policy`：`present`、`has_default_src`、`unsafe_inline`、`unsafe_eval`、`wildcard_source`
- `x_frame_options`：`present`、`mode`
- `x_content_type_options`：`present`、`nosniff`
- `referrer_policy`：`present`、`policy`
- `permissions_policy`：`present`、`policy`

这些摘要仍然只是 observation 信号，例如 `unsafe_inline=true` 只表示本次采到的 CSP 中出现了该项，不自动等价为站点存在确定风险结论。

v0.3.3 增加的页面语义字段来自同一次 HTTP GET 已经读取到的响应头和 HTML body，不会额外请求 favicon、manifest、图片或子资源：

- `open_graph`：保守提取 `title`、`description`、`site_name`、`type`、`image`、`url`。
- `twitter_card`：保守提取 `card`、`title`、`description`、`image`、`site`。
- `meta_refresh`：记录是否存在 meta refresh、延迟秒数和目标 URL。
- `icon_links`：只保存页面声明的 icon / apple-touch-icon / manifest 链接，不拉取资源内容。
- `cookie_summary`：只保存 `Set-Cookie` 数量和 `Secure`、`HttpOnly`、`SameSite` 计数，不保存 Cookie 原文。
- `server_hints`：只记录响应头或 meta 显式暴露的信息，例如 `server`、`x-powered-by`、`generator`，不做技术栈确定性推断。
- `redirect_hint`：只提示 final URL、canonical、meta refresh 之间是否存在差异，不表示站点异常。

外部文本在 v2 payload 中做了限长：`title`、`server`、`meta`、`headers`、redirect URL、final URL、缓存相关 header、证书 issuer、证书 DNS names、证书 email、`verify_error`、`cert_subject_cn`、Open Graph、Twitter Card、icon link、页面摘要等都按固定上限截断。v0.3.3 新字段也会同步进入旧 `request:{domain}` JSON 和旧 HTTP 日志 JSON，但只作为向后兼容的追加字段。

## DNS Payload

DNS v2 payload 当前仍按记录类型提供顶层结果，例如 `A`、`AAAA`、`CNAME`、`MX`、`NS`、`TXT`、`SOA`、`CAA`。

每条 v2 DNS 记录包含：

- `type`、`value`、`ttl`、`dnssec`。
- `asn`、`country`、`city`、`provider_type`、`isp`。
- `duration_ms`、`children`、`reverse_ptr`。
- `risk_flags`：风险提示数组。

顶层还包含聚合后的 `risk_flags`。当前风险标记包括：

- `private_ip`
- `low_ttl`
- `nxdomain_with_answer`
- `ptr_empty`

顶层还包含：

- `response_summary`：按记录类型保存 `rcode`、`authoritative`、`truncated`、`recursion_available`、`answer_count`、`authority_count`、`additional_count`、`ttl_min`、`ttl_max`、`ttl_avg`、`dnssec_rrsig_present`、`dnssec_ad`。
- `cname_chain_depth`：本次 observation 中最大的递归 children 深度。
- `mx_priorities`：MX 主机和优先级摘要。
- `soa`：SOA 的 `ns`、`mbox`、`serial`、`refresh`、`retry`、`expire`、`minttl` 摘要。

旧 `DNSRecord.hijacked` 仍保留在旧 Redis/旧表兼容路径中；v2 payload 不输出 `hijacked`。

TXT / SPF / DMARC / CAA 相关文本只基于当前已采到的记录识别，不额外查询 `_dmarc` 或其他新目标。超长文本会包含：

- `value_truncated=true`
- `value_original_length`
- `value_sha256`
- `text_kind`：`txt`、`spf`、`dmarc` 或 `caa`

## 后续边界

- 健康状态聚合、协议权重、站点是否 `down` 属于 v0.4.0。
- TCP connect fallback、HTTP HEAD-first、DNS multi-resolver 属于后续可选协议能力，默认不启用。
- v2 observation 仍是旁路数据面；切换后端或前端读取路径需要单独灰度计划。
