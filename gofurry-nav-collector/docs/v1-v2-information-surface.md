# GoFurry Nav Collector v1/v2 信息面统计

本文档统计 `gofurry-nav-collector` v1 与 v2 能采集、写入和供后端消费的信息面，用于后续 `gofurry-nav-backend /api/v2/nav` 站点详情接口设计。字段名以当前代码里的 JSON / DB / Redis key 为准。

## 统计摘要

| 项目 | v1 | v2 |
|---|---|---|
| 主采集协议 | `ping`、`http`、`dns` | `ping`、`http`、`dns`，并保留 v1 写入 |
| 低频旁路协议 | 无 | `rdap`、`robots`、`security_txt`、`page_assets`、`port_check`、`waf_canary` |
| 历史表 | `gfn_collector_log_ping`、`gfn_collector_log_http`、`gfn_collector_log_dns` | `gfn_collector_observation` 统一保存所有协议 observation |
| 最新结果 | `ping:result`、`request:{target}`、`dns:{target}` | `collector:v2:latest:{protocol}:{site_id}`、`collector:v2:latest:{protocol}:{site_id}:{target}` |
| 健康摘要 | 无 | `collector:v2:summary:target:{site_id}:{target}`、`collector:v2:summary:site:{site_id}` |
| 派生信息 | 无 | `collector:v2:trend:target:{site_id}:{target}`、`collector:v2:change:target:{site_id}:{target}` |
| 运行状态 | 主要依赖日志 | `collector:v2:run:{protocol}:latest`、`collector:v2:run:{protocol}:{job_id}`、可选 `collector:v2:lease:{protocol}` |
| 后端当前接入 | v1 已接入 Ping / HTTP / DNS 最新与部分历史 | v2 仅接入 site summary 与 target summary |

## v1 信息面

### 采集目标

| 来源 | 字段 | 说明 |
|---|---|---|
| `gfn_collector_domain` | `id` | 采集域名记录 ID。 |
| `gfn_collector_domain` | `site_id` | 关联 `gfn_site.id`，v2 observation 也依赖该字段。 |
| `gfn_collector_domain` | `name` | 域名主体。 |
| `gfn_collector_domain` | `prefix` | 可选前缀，实际 target 为 `prefix + name`。 |
| `gfn_collector_domain` | `proxy` | 是否使用代理，`1` 表示需要代理。 |
| `gfn_collector_domain` | `tls` | 是否按 HTTPS 构造 URL，`1` 表示 HTTPS。 |
| `gfn_collector_domain` | `deleted` | 软删除标记；采集时排除已删除目标。 |

collector 查询目标时会 join `gfn_site` 并排除 `gfn_site.deleted IS TRUE` 的站点。后端 v1 站点列表会读取 `gfn_site` 并聚合 `gfn_collector_domain` 为站点域名数组。

### Ping v1

| 信息面 | 字段 | 说明 |
|---|---|---|
| `ping:result` hash value | `status` | `up` / `down`，基于丢包率和平均延迟生成。 |
| `ping:result` hash value | `time` | 本次 Ping 时间。 |
| `ping:result` hash value | `loss` | 丢包率字符串。 |
| `ping:result` hash value | `delay` | 平均延迟字符串，例如 `42ms`。 |
| `gfn_collector_log_ping` | `id` | Ping 历史记录 ID。 |
| `gfn_collector_log_ping` | `name` | target 域名。 |
| `gfn_collector_log_ping` | `delay` | 平均延迟字符串。 |
| `gfn_collector_log_ping` | `loss` | 丢包率字符串。 |
| `gfn_collector_log_ping` | `status` | `up` / `down`。 |
| `gfn_collector_log_ping` | `create_time` | 入库时间。 |

### HTTP/TLS v1

| 信息面 | 字段 | 说明 |
|---|---|---|
| `request:{target}` / `gfn_collector_log_http.info` | `domain` | target 域名。 |
| 同上 | `url` | 实际请求 URL。 |
| 同上 | `statusCode` | HTTP 状态码。 |
| 同上 | `responseTime` | 响应耗时字符串。 |
| 同上 | `contentLength` | 页面大小。 |
| 同上 | `title` | 页面标题。 |
| 同上 | `server` | `Server` 响应头。 |
| 同上 | `redirects` | 重定向链。 |
| 同上 | `headers` | 常见响应头映射。 |
| 同上 | `meta` | meta 标签摘要。 |
| 同上 | `openGraph` | OpenGraph meta 摘要。 |
| 同上 | `twitterCard` | Twitter Card meta 摘要。 |
| 同上 | `canonicalUrl` | canonical URL。 |
| 同上 | `htmlLang` | HTML `lang`。 |
| 同上 | `metaRefresh` | `present`、`delay_seconds`、`url`。 |
| 同上 | `iconLinks` | `rel`、`href`、`type`、`sizes`。 |
| 同上 | `manifestLink` | `rel`、`href`、`type`、`sizes`。 |
| 同上 | `cookieSummary` | `set_cookie_count`、`secure_count`、`http_only_count`、`same_site_lax_count`、`same_site_strict_count`、`same_site_none_count`。 |
| 同上 | `serverHints` | `server`、`x_powered_by`、`generator`。 |
| 同上 | `crossOriginSummary` | `cross_origin_opener_policy`、`cross_origin_embedder_policy`、`cross_origin_resource_policy`、`access_control_allow_origin`。 |
| 同上 | `contentLanguageEffective` | 内容语言。 |
| 同上 | `pageTextSummary` | 页面文本摘要。 |
| 同上 | `sharePreview` | `title`、`description`、`site_name`、`image`、`url`。 |
| 同上 | `redirectHint` | `final_url_different`、`canonical_url_different`、`meta_refresh_present`、`meta_refresh_url`。 |
| 同上 | `tlsVersion` | TLS 版本。 |
| 同上 | `cipherSuite` | TLS 加密套件。 |
| 同上 | `certExpiry` | 证书过期时间。 |
| 同上 | `certDaysLeft` | 证书剩余天数。 |
| 同上 | `certIssuer` | 证书签发机构。 |
| 同上 | `certIssuerOrg` | 签发组织。 |
| 同上 | `certDNSNames` | 证书 DNS Names。 |
| 同上 | `certPubKeyAlg` | 公钥算法。 |
| 同上 | `certSigAlg` | 签名算法。 |
| 同上 | `certEmail` | 证书邮箱。 |
| 同上 | `certIsCA` | 是否 CA 证书。 |
| `gfn_collector_log_http` | `id`、`name`、`info`、`status`、`create_time` | HTTP 历史日志外壳，`info` 保存上述 JSON。 |

### DNS v1

| 信息面 | 字段 | 说明 |
|---|---|---|
| `dns:{target}` hash | `A`、`AAAA`、`CNAME`、`TXT`、`MX`、`NS`、`SOA`、`CAA` | 每个字段保存对应记录类型的 JSON 数组。 |
| v1 DNS record | `type` | 记录类型。 |
| v1 DNS record | `value` | 记录值。 |
| v1 DNS record | `ttl` | TTL。 |
| v1 DNS record | `dnssec` | 是否出现 DNSSEC 信号。 |
| v1 DNS record | `asn` | IP ASN。 |
| v1 DNS record | `country` | IP 国家。 |
| v1 DNS record | `city` | IP 城市。 |
| v1 DNS record | `provider_type` | CDN / Origin 等判定。 |
| v1 DNS record | `isp` | ISP。 |
| v1 DNS record | `duration` | 查询耗时。 |
| v1 DNS record | `children` | 递归查询产生的子记录。 |
| v1 DNS record | `reverse_ptr` | PTR 反查结果。 |
| v1 DNS record | `hijacked` | 旧劫持检测标记。 |
| `gfn_collector_log_dns` | `id`、`name`、`a`、`aaaa`、`mx`、`ns`、`soa`、`txt`、`caa`、`cname`、`status`、`create_time` | DNS 历史日志字段。 |

## v2 信息面

### 通用 observation 外壳

| 信息面 | 字段 | 说明 |
|---|---|---|
| `gfn_collector_observation` | `id` | observation 记录 ID。 |
| 同上 | `site_id` | 站点 ID。 |
| 同上 | `target` | 采集目标。 |
| 同上 | `protocol` | `ping`、`http`、`dns`、`rdap`、`robots`、`security_txt`、`page_assets`、`port_check`、`waf_canary`。 |
| 同上 | `status` | `success` / `failure`。 |
| 同上 | `observed_at` | 观测时间。 |
| 同上 | `duration_ms` | 单次观测耗时。 |
| 同上 | `error_code` | 可空错误码。 |
| 同上 | `error_message` | 可空错误消息。 |
| 同上 | `payload` | 协议 payload，JSONB。 |
| 同上 | `schema_version` | 当前为 `1`。 |
| 同上 | `create_time` | 入库时间。 |
| `collector:v2:latest:{protocol}:{site_id}` | `site_id`、`target`、`protocol`、`status`、`observed_at`、`duration_ms`、`error_code`、`error_message`、`payload`、`schema_version`、`collector_id`、`job_id` | site 级协议 latest，保留兼容。 |
| `collector:v2:latest:{protocol}:{site_id}:{target}` | 同上 | target 级协议 latest，后端 v2 应优先读取。 |

`collector_id` 与 `job_id` 也会在可用时注入 payload，用于排查采集实例和运行批次。

### Ping v2 payload

| 字段 | 说明 |
|---|---|
| `icmp_status` | `reachable` / `unreachable`。 |
| `avg_rtt_ms`、`min_rtt_ms`、`max_rtt_ms`、`stddev_rtt_ms`、`jitter_ms` | RTT 与抖动统计；不可达时可为空。 |
| `loss_rate` | 丢包率。 |
| `packets_sent`、`packets_recv`、`packets_recv_duplicates` | 包统计。 |
| `resolved_ip`、`resolved_ips`、`selected_ip`、`ip_family`、`resolution_source` | 解析与选路信息。 |
| `icmp_blocked_suspected` | 已解析 IP 但 ICMP 全丢包时的旁路判断。 |
| `duration_ms` | 单目标探测墙钟耗时。 |
| `error_code` | `ping_unreachable`、`ping_init_failed`、`ping_run_failed` 等。 |
| `delay_ms`、`legacy_delay`、`legacy_loss`、`legacy_status` | v1 兼容字段。 |

### HTTP/TLS v2 payload

| 分类 | 字段 |
|---|---|
| 基本响应 | `domain`、`url`、`status_code`、`response_time_ms`、`content_length`、`title`、`server`、`content_type` |
| HTML / metadata | `headers`、`meta`、`open_graph`、`twitter_card`、`canonical_url`、`html_lang`、`meta_refresh`、`icon_links`、`manifest_link`、`content_language_effective`、`page_text_summary`、`share_preview` |
| redirect | `redirects`、`redirect_chain`、`redirect_count`、`final_url`、`redirect_hint`、`canonical_host_matches_final_host` |
| cookie / server / cross-origin | `cookie_summary`、`server_hints`、`cross_origin_summary` |
| trace / transport | `http_protocol`、`dns_lookup_ms`、`tcp_connect_ms`、`tls_handshake_ms`、`ttfb_ms`、`transfer_ms`、`remote_addr`、`remote_ip` |
| body / transfer | `body_read_bytes`、`body_truncated`、`body_limit_bytes`、`compressed`、`content_encoding`、`content_length_header`、`transfer_encoding`、`is_chunked`、`html_charset`、`doctype`、`robots_meta_policy`、`compression_ratio_estimated` |
| cache | `cache_policy.cache_control`、`cache_policy.etag`、`cache_policy.last_modified` |
| security headers | `security_headers`、`security_header_summary` |
| TLS legacy | `tls_version`、`cipher_suite`、`cert_expiry`、`cert_days_left`、`cert_issuer`、`cert_issuer_org`、`cert_dns_names`、`cert_pub_key_alg`、`cert_sig_alg`、`cert_public_key_algorithm`、`cert_signature_algorithm`、`cert_email`、`cert_is_ca` |
| TLS v2 | `cert_collected`、`cert_verified`、`verify_error`、`verify_error_category`、`tls_handshake`、`cert_not_before`、`cert_not_after`、`cert_chain_length`、`cert_subject_cn`、`cert_san_count`、`ocsp_stapled`、`sct_count`、`cert_serial_number`、`cert_fingerprint_sha256`、`cert_spki_sha256`、`cert_public_key_bits`、`cert_subject_org`、`cert_issuer_cn`、`cert_chain_issuers` |

### DNS v2 payload

| 分类 | 字段 |
|---|---|
| 记录集合 | `A`、`AAAA`、`MX`、`NS`、`TXT`、`CNAME`、`SOA`、`CAA` |
| 记录字段 | `type`、`value`、`ttl`、`dnssec`、`asn`、`country`、`city`、`provider_type`、`isp`、`duration_ms`、`children`、`reverse_ptr`、`risk_flags` |
| 长文本保护 | `value_truncated`、`value_original_length`、`value_sha256`、`text_kind` |
| 顶层摘要 | `risk_flags`、`response_summary`、`cname_chain_depth`、`mx_priorities`、`soa`、`record_budget_exhausted` |
| 结构摘要 | `has_a`、`has_aaaa`、`ipv4_count`、`ipv6_count`、`cname_terminal`、`name_server_hosts`、`mx_hosts`、`ttl_spread`、`mixed_private_public_ip` |
| `response_summary.{type}` | `rcode`、`authoritative`、`truncated`、`recursion_available`、`answer_count`、`authority_count`、`additional_count`、`ttl_min`、`ttl_max`、`ttl_avg`、`dnssec_rrsig_present`、`dnssec_ad` |
| `mx_priorities[]` | `host`、`priority` |
| `soa` | `ns`、`mbox`、`serial`、`refresh`、`retry`、`expire`、`minttl` |

v2 不再输出 v1 的 `hijacked` 字段，改用 `risk_flags` 表达风险信号。

### Summary v2

| 信息面 | 字段 |
|---|---|
| `collector:v2:summary:target:{site_id}:{target}` | `site_id`、`target`、`status`、`reason_codes`、`reason_messages`、`protocols`、`canonical_target_hint`、`target_relation_hints`、`edge_provider_hints`、`observed_at`、`generated_at`、`schema_version` |
| `protocols.{protocol}` | `protocol`、`status`、`observed_at`、`duration_ms`、`stale`、`stale_after_seconds`、`error_code` |
| `canonical_target_hint` | `target_host`、`final_host`、`canonical_host`、`preferred_host`、`relation`、`source`、`final_url`、`canonical_url` |
| target 级 `target_relation_hints[]` | `relation`、`source`、`target_host`、`related_host`、`value` |
| `edge_provider_hints[]` | `provider`、`hint_type`、`confidence`、`evidence` |
| `edge_provider_hints[].evidence[]` | `source`、`field`、`value` |
| `collector:v2:summary:site:{site_id}` | `site_id`、`status`、`reason_codes`、`reason_messages`、`target_count`、`status_counts`、`targets`、`target_relation_hints`、`generated_at`、`schema_version` |
| `targets[]` | `target`、`status`、`reason_codes`、`reason_messages`、`canonical_target_hint`、`target_relation_hints`、`edge_provider_hints`、`observed_at` |
| site 级 `target_relation_hints[]` | `relation`、`host`、`targets` |
| `collector:v2:summary:site_targets:{site_id}` | target summary 索引集合，members 为 target 字符串。 |

`status` 取值：`healthy`、`warning`、`degraded`、`unknown`、`down`。summary 当前只把 Ping / HTTP / DNS / TLS 作为健康聚合信号，light probe、edge hints、trend、change 都是旁路参考。

### Trend v2

| 信息面 | 字段 |
|---|---|
| `collector:v2:trend:target:{site_id}:{target}` | `site_id`、`target`、`windows`、`generated_at`、`schema_version` |
| `windows.24h` / `windows.7d` | `since`、`until`、`protocols` |
| `protocols.{protocol}` | `protocol`、`observation_count`、`success_count`、`failure_count`、`success_rate`、`avg_duration_ms`、`p95_duration_ms`、`last_observed_at`、`last_failure_at`、`http`、`ping`、`dns`、`tls` |
| `http` | `avg_response_time_ms`、`p95_response_time_ms`、`latest_failure_at` |
| `ping` | `avg_rtt_ms`、`avg_loss_rate`、`avg_jitter_ms`、`latest_loss_rate`、`latest_avg_rtt_ms`、`latest_jitter_ms` |
| `dns` | `success_rate`、`latest_ttl_min`、`latest_ttl_max`、`latest_ttl_avg`、`previous_ttl_min`、`previous_ttl_max`、`previous_ttl_avg`、`risk_flag_counts`、`latest_risk_flags` |
| `tls` | `latest_cert_days_left`、`previous_cert_days_left`、`cert_issuer_changed`、`cert_fingerprint_changed`、`latest_cert_issuer`、`latest_cert_fingerprint_sha256`、`latest_cert_not_after`、`latest_cert_observed_at` |

### Change v2

| 信息面 | 字段 |
|---|---|
| `collector:v2:change:target:{site_id}:{target}` | `site_id`、`target`、`events`、`generated_at`、`schema_version` |
| `events[]` | `event_id`、`protocol`、`category`、`field`、`old_value`、`new_value`、`old_observed_at`、`new_observed_at`、`detected_at` |

当前变化检测范围：

| 协议 | 字段 |
|---|---|
| HTTP | `status_code`、`title`、`server`、`x_powered_by`、`final_url`、`security_headers` |
| TLS | `cert_fingerprint_sha256`、`cert_issuer`、`cert_san_count`、`cert_not_after` |
| DNS | `A`、`AAAA`、`cname_terminal`、`mx_hosts`、`name_server_hosts`、`soa_serial` |
| Port check | 每个端口的 `open`、`closed`、`timeout`、`filtered_suspected` 等状态变化 |
| RDAP | `statuses`、`expires_at`、`nameservers` |

### Light probe v2

| 协议 | 字段 |
|---|---|
| `rdap` | `registrable_domain`、`rdap_server`、`status_code`、`registrar`、`statuses`、`expires_at`、`nameservers`、`dnssec_delegation_signed`、`events_summary`、`raw_truncated`、`error_code` |
| `robots` | `exists`、`status_code`、`content_type`、`sitemap_count`、`sitemaps`、`global_disallow_all`、`user_agent_star_present`、`body_truncated`、`error_code` |
| `security_txt` | `exists`、`path_used`、`status_code`、`content_type`、`contact`、`expires`、`policy`、`preferred_languages`、`canonical`、`body_truncated`、`error_code` |
| `page_assets` | `http_latest_found`、`icon`、`manifest` |
| `page_assets.icon` | `exists`、`source_url`、`selected_rel`、`selected_sizes`、`status_code`、`content_type`、`content_length_header`、`body_read_bytes`、`body_truncated`、`sha256`、`skipped_reason`、`error_message` |
| `page_assets.manifest` | `exists`、`source_url`、`selected_rel`、`selected_sizes`、`status_code`、`content_type`、`content_length_header`、`body_read_bytes`、`body_truncated`、`sha256`、`name`、`short_name`、`theme_color`、`background_color`、`display`、`start_url`、`scope`、`icons_count`、`manifest_decode_error`、`skipped_reason`、`error_message` |
| `port_check` | `ports_configured`、`ports_checked`、`open_count`、`closed_count`、`timeout_count`、`filtered_suspected_count`、`skipped_count`、`invalid_port_count`、`duplicate_port_count`、`truncated_port_count`、`truncated`、`results`、`skipped_reason` |
| `port_check.results[]` | `port`、`service_hint`、`status`、`duration_ms`、`error_code`、`error_message` |
| `waf_canary` | `canary_path`、`cases_total`、`cases_executed`、`blocked_count`、`expected_blocked_count`、`expected_blocked_matched_count`、`unexpected_pass_count`、`network_error_count`、`status_code_unexpected_count`、`max_targets_per_run`、`truncated_target_count`、`target_run_truncated`、`cases` |
| `waf_canary.cases[]` | `case_id`、`category`、`method`、`expected_blocked`、`expected_status_codes`、`status_code_expected`、`status_code`、`blocked`、`matched_expected`、`duration_ms`、`error_code`、`error_message` |

### Run state v2

| 信息面 | 字段 |
|---|---|
| `collector:v2:run:{protocol}:latest` | `collector_id`、`job_id`、`protocol`、`status`、`started_at`、`finished_at`、`duration_ms`、`target_count`、`success_count`、`failure_count`、`skipped_count`、`error_count`、`skip_reason` |
| `collector:v2:run:{protocol}:{job_id}` | 同上，默认 168 小时 TTL。 |
| `collector:v2:lease:{protocol}` | `collector_id`、`job_id`、`protocol`、`acquired_at`、`expires_at`；仅 lease 开启时存在。 |

## v1 与 v2 差异

| 差异项 | v1 | v2 |
|---|---|---|
| 数据模型 | 按协议分散在旧表和旧 Redis key。 | 统一 observation 外壳，协议细节进入 payload。 |
| 站点关联 | 多数 v1 最新 key 以 target 字符串为主。 | `site_id + target + protocol` 成为稳定主键语义。 |
| 状态语义 | Ping 使用 `up/down`，HTTP/DNS 使用 `success/failure`，站点健康需要前端或人工理解。 | observation 使用 `success/failure`，summary 输出 `healthy/warning/degraded/unknown/down`。 |
| 历史查询 | Ping / HTTP / DNS 分表，字段形态不同。 | `gfn_collector_observation` 可按 `site_id`、`target`、`protocol`、`observed_at` 查询。 |
| HTTP 深度 | v1 已有页面和证书基本信息。 | 增加 trace、transport、body 读取、cache、安全响应头、证书指纹、SPKI、校验分类等字段。 |
| DNS 深度 | v1 保存记录数组和 `hijacked`。 | 增加 `risk_flags`、响应摘要、TTL 汇总、MX/SOA 摘要、私网/公网混合等信号；不再输出 `hijacked`。 |
| Ping 深度 | v1 只有延迟、丢包、状态。 | 增加 RTT 分布、包统计、解析 IP、IP family、ICMP 阻断疑似信号。 |
| 治理信息 | 无。 | 增加 RDAP、robots.txt、security.txt、page assets、port check、WAF canary。 |
| 派生信息 | 无。 | 增加 target 趋势和变化事件。 |
| 运行排查 | 主要看应用日志。 | Redis run state 记录每轮采集状态、计数和失败原因。 |
| 后端消费方式 | 后端直接读取旧 key / 旧表，部分返回原始字符串。 | 后端应通过只读 read model 聚合 latest、summary、trend、change、light probe。 |

## 后端当前接入统计

### v1 已接入

| 后端接口 | 读取来源 | 已暴露字段 |
|---|---|---|
| `GET /api/v1/nav/page/site/list` | `gfn_site` + `gfn_collector_domain` | `id`、`name`、`domain`、`info`、`country`、`nsfw`、`welfare`、`icon`。其中 `domain` 是聚合后的 target 数组 JSON 字符串。 |
| `GET /api/v1/nav/page/ping/list` | `ping:result` | target 到 Ping v1 JSON 字符串的 map，包含 `status`、`time`、`loss`、`delay`。 |
| `GET /api/v1/nav/site/getSitePingRecord` | `gfn_collector_log_ping` | `twenty`、`sixty`、`hundred` 三个窗口，每个包含 `DelayModel[]`、`avgDelay`、`avgLoss`；`DelayModel[]` 包含 `delay`、`loss`、`status`、`time`。 |
| `GET /api/v1/nav/site/getSiteHttpRecord` | `request:{domain}` | 返回 v1 HTTP/TLS JSON 字符串。 |
| `GET /api/v1/nav/site/getSiteDnsRecord` | `dns:{domain}` | `a`、`AAAA`、`CNAME`、`txt`、`MX`、`ns`、`SOA`、`caa`，值为 v1 DNS record JSON 字符串。 |

### v2 已接入

| 后端接口 | 读取来源 | 已暴露字段 |
|---|---|---|
| `GET /api/v2/nav/sites/:siteId/summary` | `collector:v2:summary:site:{site_id}` | `state`、`site_id`、`status`、`reason_codes`、`reason_messages`、`target_count`、`status_counts`、`targets`、`generated_at`、`schema_version`。 |
| `GET /api/v2/nav/sites/:siteId/targets/:target/summary` | `collector:v2:summary:target:{site_id}:{target}` | `state`、`site_id`、`target`、`status`、`reason_codes`、`reason_messages`、`protocols`、`observed_at`、`generated_at`、`schema_version`。 |

当前 backend summary DTO 尚未保留 collector 已输出的这些字段：

| collector 字段 | 缺口 |
|---|---|
| target summary `canonical_target_hint` | 后端 target summary 响应未暴露。 |
| target summary `target_relation_hints` | 后端 target summary 响应未暴露。 |
| target summary `edge_provider_hints` | 后端 target summary 响应未暴露。 |
| site summary 顶层 `target_relation_hints` | 后端 site summary 响应未暴露。 |
| site summary `targets[].canonical_target_hint` | 后端 site summary target item 未暴露。 |
| site summary `targets[].target_relation_hints` | 后端 site summary target item 未暴露。 |
| site summary `targets[].edge_provider_hints` | 后端 site summary target item 未暴露。 |

### v2 待后端接入

| 待接入信息面 | 来源 | 建议用途 |
|---|---|---|
| raw observation 历史 | `gfn_collector_observation` | 站点详情历史序列、排障、协议分页查询。 |
| target latest | `collector:v2:latest:{protocol}:{site_id}:{target}` | 站点详情当前 Ping / HTTP / DNS / light probe 最新原始观测。 |
| site latest 兼容 key | `collector:v2:latest:{protocol}:{site_id}` | 兼容和排查，不作为详情页首选。 |
| summary hints | `canonical_target_hint`、`target_relation_hints`、`edge_provider_hints` | 详情页目标治理、跳转关系、CDN/WAF/托管线索展示。 |
| trend | `collector:v2:trend:target:{site_id}:{target}` | 24h / 7d 成功率、P95、TTL、证书变化趋势。 |
| change | `collector:v2:change:target:{site_id}:{target}` | 最近稳定变化事件，例如标题、DNS、证书、端口、RDAP。 |
| light probe | `rdap`、`robots`、`security_txt`、`page_assets`、`port_check`、`waf_canary` latest / observation | 低频治理与安全参考，不参与健康状态判断。 |
| run state | `collector:v2:run:{protocol}:latest`、`collector:v2:run:{protocol}:{job_id}` | 后端诊断接口或管理侧排查。 |
| summary target index | `collector:v2:summary:site_targets:{site_id}` | 后端聚合 target 列表和补齐详情页 target selector。 |

## 后端 v2 接入优先级建议

| 优先级 | 内容 | 原因 |
|---|---|---|
| P0 | 补齐 summary hints 字段透传。 | collector 已稳定输出，当前后端 DTO 丢字段；改动小，能立刻提升详情页可用信息。 |
| P0 | 新增 observation read model 与 target latest 读取。 | 这是站点详情 v2 的基础数据层。 |
| P1 | 接入 trend / change。 | 用于详情页趋势与变化解释，缺失时不影响主状态。 |
| P1 | 接入 light probe latest。 | 低频治理信息适合详情页分区展示，但不应阻断主详情接口。 |
| P2 | 接入 run state。 | 更偏管理和排障，可后置到管理侧或诊断接口。 |
| P2 | 设计浏览量独立写接口。 | v2 detail 应保持只读，避免继续把读取详情和写 view count 绑定。 |
