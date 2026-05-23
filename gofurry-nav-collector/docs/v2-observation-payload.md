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
- `loss_rate`：丢包率。
- `duration_ms`：单目标 Ping 探测墙钟耗时。
- `error_code`：`ping_init_failed`、`ping_run_failed`、`ping_unreachable` 或空字符串。
- 兼容字段：`delay_ms`、`legacy_delay`、`legacy_loss`、`legacy_status`。

Ping 只作为辅助信号；是否影响站点整体健康状态由后续 summary 规则决定。

## HTTP / TLS Payload

HTTP v2 payload 当前包含：

- HTTP 基础字段：`url`、`status_code`、`response_time_ms`、`content_length`、`title`、`server`、`headers`、`meta`。
- 重定向字段：`redirects`、`redirect_chain`、`redirect_count`、`final_url`。
- 响应类型：`content_type`。
- 常见安全响应头是否存在：`security_headers`。
- TLS 兼容字段：`tls_version`、`cipher_suite`、`cert_expiry`、`cert_days_left`、`cert_issuer`、`cert_issuer_org`、`cert_dns_names`、`cert_pub_key_alg`、`cert_sig_alg`、`cert_email`、`cert_is_ca`。
- TLS 语义字段：`cert_collected`、`cert_verified`、`verify_error`、`tls_handshake`。

HTTP 仍使用 GET。v0.3.0 不启用 HEAD-first，也不改变重定向、超时、响应体上限和旧展示结构。

外部文本在 v2 payload 中做了限长：`title`、`server`、`meta`、`headers`、redirect URL、final URL、证书 issuer、证书 DNS names、证书 email、`verify_error` 都按固定上限截断，旧 Redis/旧表不受影响。

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
