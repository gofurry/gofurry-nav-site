# v2 Observation Payload 字段说明

本文档记录 `gofurry-nav-collector` 写入 `gfn_collector_observation.payload` 与 `collector:v2:latest:{protocol}:{site_id}` 的主要字段。所有字段都只是 observation 信号，用于后端 v2、前端详情页和人工排查参考，不直接等同于站点健康结论。

旧 Redis key、旧日志表和旧前端展示结构仍是兼容路径；新增字段只在 v2 observation/latest Redis 中旁路出现。

## Summary Status

v2 summary 目前只由 Ping / HTTP / DNS / TLS 相关信号参与健康聚合；RDAP、robots.txt、security.txt、page_assets、port_check、waf_canary 和 edge hints 都是旁路参考，不直接改变健康状态。

| status | 中文含义 | 使用边界 |
|---|---|---|
| `healthy` | 当前主要访问链路健康 | HTTP 可用，且没有参与聚合的 warning/degraded/down 信号。 |
| `warning` | 需要关注 | HTTP 可用，但 DNS、Ping、TLS 即将过期或 DNS risk_flags 等出现观察信号。 |
| `degraded` | 降级 | HTTP 失败但尚未满足 down 条件，或 TLS 校验失败等影响访问信任的信号出现。 |
| `unknown` | 状态未知 | 缺少可用 HTTP 观测、观测过期，或站点没有可聚合 target。 |
| `down` | 不可用 | HTTP 与 DNS 均失败，或站点下所有 target 均为 down。 |

## Reason Code 字典

`reason_codes` 是稳定英文 key，`reason_messages` 是中文说明。后端 v2 应优先使用 `reason_codes` 做逻辑判断，用本表或后端自己的 i18n 字典做展示文案。`affects_health=true` 表示该 reason 当前参与 summary 状态聚合。

| code | message_zh | severity | scope | affects_health |
|---|---|---|---|---|
| `http_missing_or_stale` | HTTP 观测缺失或已过期，无法判断访客是否可打开 | unknown | target | true |
| `http_failed` | HTTP 访问失败 | degraded | target | true |
| `dns_failed` | DNS 解析也失败 | down | target | true |
| `dns_missing_or_stale` | DNS 观测缺失或已过期 | warning | target | true |
| `dns_failed_but_http_ok` | DNS 失败但 HTTP 当前仍可访问 | warning | target | true |
| `ping_failed_but_http_ok` | Ping 失败但 HTTP 当前仍可访问 | warning | target | true |
| `dns_risk_private_ip` | DNS observation 出现风险信号: private_ip | warning | target | true |
| `dns_risk_low_ttl` | DNS observation 出现风险信号: low_ttl | warning | target | true |
| `dns_risk_nxdomain_with_answer` | DNS observation 出现风险信号: nxdomain_with_answer | warning | target | true |
| `dns_risk_ptr_empty` | DNS observation 出现风险信号: ptr_empty | warning | target | true |
| `dns_risk_other` | DNS observation 出现未知风险信号 | warning | target | true |
| `tls_verify_expired` | TLS 证书校验未通过: expired | degraded | target | true |
| `tls_verify_not_yet_valid` | TLS 证书校验未通过: not_yet_valid | degraded | target | true |
| `tls_verify_hostname_mismatch` | TLS 证书校验未通过: hostname_mismatch | degraded | target | true |
| `tls_verify_unknown_authority` | TLS 证书校验未通过: unknown_authority | degraded | target | true |
| `tls_verify_incompatible_usage` | TLS 证书校验未通过: incompatible_usage | degraded | target | true |
| `tls_verify_other` | TLS 证书校验未通过: other | degraded | target | true |
| `tls_cert_expired` | TLS 证书已过期 | degraded | target | true |
| `tls_cert_expiring_soon` | TLS 证书将在 30 天内过期 | warning | target | true |
| `no_target_summary` | 没有可用的采集目标健康摘要 | unknown | site | true |
| `all_targets_down` | 所有采集目标都判定为 down | down | site | true |
| `all_targets_unknown` | 所有采集目标状态未知 | unknown | site | true |
| `some_targets_degraded` | 部分采集目标不可用或降级 | degraded | site | true |
| `some_targets_warning` | 部分采集目标存在需要关注的观测信号 | warning | site | true |

## 后端 v2 消费建议

- 站点详情优先读取 `collector:v2:summary:site:{site_id}` 与 `collector:v2:summary:target:{site_id}:{target}`；raw observation 更适合历史详情、排障和趋势计算。
- 展示逻辑优先依赖 `status` 与 `reason_codes`，不要解析 `reason_messages` 文本做判断。
- `light_probe` 系列协议和 `edge_provider_hints` 是治理/安全参考信号，不应直接作为站点上下架、排序或 down 判断依据。
- 后端 v2 对字段应采用兼容读取策略：未知字段忽略，缺失字段使用安全默认值，避免 collector 后续 additive 字段影响接口稳定。

## 字段契约速查

### Summary

| 字段 | 类型 | 可空 | 来源 | 参与 summary |
|---|---|---:|---|---:|
| `site_id` | number | 否 | collector domain/site | 是 |
| `target` | string | target summary 否 | collector target | 是 |
| `status` | string | 否 | summary builder | 是 |
| `reason_codes` | string[] | 否 | reason 字典 | 是 |
| `reason_messages` | string[] | 否 | reason 字典中文文案 | 否 |
| `protocols` | object | target summary 否 | Ping / HTTP / DNS latest | 是 |
| `edge_provider_hints` | object[] | 是 | HTTP / DNS / TLS 被动推断 | 否 |

### Ping

| 字段 | 类型 | 可空 | 来源 | 参与 summary |
|---|---|---:|---|---:|
| `icmp_status` | string | 否 | go-ping 结果 | 否 |
| `avg_rtt_ms` / `min_rtt_ms` / `max_rtt_ms` / `stddev_rtt_ms` / `jitter_ms` | number | 是 | go-ping 统计 | 否 |
| `loss_rate` | number | 否 | go-ping 统计 | 否 |
| `packets_sent` / `packets_recv` / `packets_recv_duplicates` | number | 否 | go-ping 统计 | 否 |
| `resolved_ip` / `resolved_ips` / `selected_ip` / `ip_family` | string / string[] | 是 | go-ping 解析结果 | 否 |
| `icmp_blocked_suspected` | bool | 否 | Ping payload builder | 否 |

### HTTP / TLS

| 字段 | 类型 | 可空 | 来源 | 参与 summary |
|---|---|---:|---|---:|
| `status_code` / `response_time_ms` | number | 否 | HTTP GET | HTTP latest status 参与 |
| `http_protocol` / `remote_addr` / `remote_ip` | string | 是 | HTTP transport | 否 |
| `dns_lookup_ms` / `tcp_connect_ms` / `tls_handshake_ms` / `ttfb_ms` / `transfer_ms` | number | 是 | httptrace | 否 |
| `body_read_bytes` / `body_truncated` | number / bool | 否 | HTTP body reader | 否 |
| `security_headers` / `security_header_summary` | object | 是 | HTTP headers | 否 |
| `tls_handshake` | string | 否 | TLS connection state | 是 |
| `cert_verified` / `verify_error_category` | bool / string | 是 | TLS verify | 是 |
| `cert_not_after` | string | 是 | TLS leaf certificate | 是 |
| `cert_fingerprint_sha256` / `cert_spki_sha256` | string | 是 | TLS leaf certificate | 否 |

### DNS

| 字段 | 类型 | 可空 | 来源 | 参与 summary |
|---|---|---:|---|---:|
| `risk_flags` | string[] | 是 | DNS payload builder | 是 |
| `response_summary` | object | 是 | DNS response | 否 |
| `has_a` / `has_aaaa` / `ipv4_count` / `ipv6_count` | bool / number | 否 | DNS records | 否 |
| `cname_chain_depth` / `cname_terminal` | number / string | 是 | DNS recursive result | 否 |
| `name_server_hosts` / `mx_hosts` / `mx_priorities` / `soa` | array / object | 是 | DNS records | 否 |
| `record_budget_exhausted` | bool | 否 | DNS query budget | 否 |

### Light Probe

| 协议 | 关键字段 | 可空 | 来源 | 参与 summary |
|---|---|---:|---|---:|
| `rdap` | `registrable_domain`、`registrar`、`statuses`、`expires_at`、`nameservers` | 是 | RDAP 官方服务 | 否 |
| `robots` | `exists`、`status_code`、`sitemap_count`、`global_disallow_all` | 否 | `/robots.txt` | 否 |
| `security_txt` | `exists`、`contact`、`expires`、`policy`、`canonical` | 是 | security.txt | 否 |
| `page_assets` | `icon`、`manifest`、`sha256`、`body_truncated` | 是 | HTTP latest 声明资源 | 否 |
| `port_check` | `ports_checked`、`open_count`、`results` | 否 | TCP connect | 否 |
| `waf_canary` | `cases_total`、`blocked_count`、`unexpected_pass_count`、`cases` | 否 | 授权 WAF canary 请求 | 否 |

## Summary Edge Provider Hints

`edge_provider_hints` 是 v0.5.6 新增的被动推断字段，出现在 v2 target summary 和 site summary 的 target item 中。它只基于已有 HTTP / DNS / TLS latest payload，不新增任何请求。

```json
{
  "edge_provider_hints": [
    {
      "provider": "cloudflare",
      "hint_type": "cdn",
      "confidence": "high",
      "evidence": [
        {
          "source": "http",
          "field": "headers.cf-ray",
          "value": "abc-HKG"
        },
        {
          "source": "dns",
          "field": "A.asn",
          "value": "AS13335 (Cloudflare, Inc.)"
        }
      ]
    }
  ]
}
```

- `provider` 是保守推断的服务方线索，例如 `cloudflare`、`tencent_cloud`、`aliyun`、`aws_cloudfront`、`fastly`、`vercel`、`netlify`、`github_pages`。
- `hint_type` 可能为 `cdn`、`waf`、`reverse_proxy`、`object_storage`、`hosting_platform`。
- `confidence` 可能为 `low`、`medium`、`high`。单一 DNS/TLS 弱证据通常只给 `low`，明确 HTTP header 或多源证据才提高置信度。
- `evidence` 中的 value 已限长；这些 hints 是 observation 线索，不是事实判定，不参与健康状态计算。

## Ping

Ping payload 保留旧兼容字段 `delay_ms`、`legacy_delay`、`legacy_loss`、`legacy_status`，并补充同一次 `go-ping` 统计中已经得到的信息。

```json
{
  "icmp_status": "reachable",
  "avg_rtt_ms": 42,
  "min_rtt_ms": 35,
  "max_rtt_ms": 50,
  "stddev_rtt_ms": 4,
  "jitter_ms": 4,
  "loss_rate": 0,
  "packets_sent": 5,
  "packets_recv": 5,
  "packets_recv_duplicates": 0,
  "resolved_ip": "203.0.113.10",
  "resolved_ips": ["203.0.113.10"],
  "selected_ip": "203.0.113.10",
  "ip_family": "ipv4",
  "resolution_source": "go-ping",
  "icmp_blocked_suspected": false,
  "duration_ms": 5100,
  "error_code": ""
}
```

- `avg_rtt_ms`、`min_rtt_ms`、`max_rtt_ms`、`stddev_rtt_ms`、`jitter_ms`：失败或无有效 RTT 时为 `null`。
- `resolved_ips`：当前阶段不额外做 DNS 解析，只从本次 ping 统计中得到的 IP 生成。
- `icmp_blocked_suspected`：只表示“已解析 IP 但 ICMP 全丢包”的观测信号，不代表网站不可访问。

## HTTP

HTTP payload 继续使用同一次 GET 的响应、响应头和已读取 HTML body，不新增请求次数。

```json
{
  "status_code": 200,
  "response_time_ms": 812,
  "http_protocol": "HTTP/2.0",
  "remote_addr": "203.0.113.10:443",
  "remote_ip": "203.0.113.10",
  "dns_lookup_ms": 3,
  "tcp_connect_ms": 8,
  "tls_handshake_ms": 19,
  "ttfb_ms": 120,
  "transfer_ms": 24,
  "body_read_bytes": 1987,
  "body_truncated": false,
  "content_length_header": 512,
  "transfer_encoding": ["chunked"],
  "is_chunked": true,
  "html_charset": "utf-8",
  "doctype": "<!doctype html>",
  "robots_meta_policy": "index,follow",
  "canonical_host_matches_final_host": true,
  "compression_ratio_estimated": 0.25
}
```

- `remote_addr` / `remote_ip` 是实际 TCP 对端；目标走代理时可能是代理地址，不应包装成源站结论。
- `compression_ratio_estimated` 使用响应头 `Content-Length` 和实际读取 body 字节估算；缺少必要信息时为 `0`。
- `html_charset`、`doctype`、`robots_meta_policy` 来自已读取 HTML，不抓取额外资源。

## TLS

TLS 字段基于当前 HTTPS 连接里拿到的证书链，不额外发起第二次请求。

```json
{
  "cert_collected": true,
  "cert_verified": true,
  "verify_error": "",
  "verify_error_category": "",
  "tls_handshake": "collected",
  "cert_serial_number": "123456789",
  "cert_fingerprint_sha256": "hex...",
  "cert_spki_sha256": "hex...",
  "cert_public_key_bits": 2048,
  "cert_subject_cn": "example.com",
  "cert_subject_org": ["Example Org"],
  "cert_issuer_cn": "Example CA",
  "cert_chain_issuers": ["Example CA", "Root CA"],
  "cert_chain_length": 2,
  "cert_san_count": 3,
  "ocsp_stapled": true,
  "sct_count": 2
}
```

- `cert_fingerprint_sha256`：叶子证书 DER 的 SHA256 指纹。
- `cert_spki_sha256`：叶子证书 Subject Public Key Info 的 SHA256 指纹，适合识别同一公钥换证场景。
- `cert_public_key_bits`：RSA/ECDSA/Ed25519 等常见公钥的位数；无法识别时为 `0`。
- `verify_error_category` 仅做错误分类，可能值包括 `expired`、`not_yet_valid`、`hostname_mismatch`、`unknown_authority`、`incompatible_usage`、`other`。

## DNS

DNS payload 仍使用当前记录类型、当前 resolver 和 `MaxDepth=2`，不新增查询目标。旧 `DNSRecord.Hijacked` 不进入 v2 payload，v2 使用 `risk_flags` 表达观测信号。

```json
{
  "risk_flags": ["low_ttl", "ptr_empty"],
  "has_a": true,
  "has_aaaa": false,
  "ipv4_count": 2,
  "ipv6_count": 0,
  "cname_chain_depth": 1,
  "cname_terminal": "edge.example.com.",
  "name_server_hosts": ["ns1.example.com.", "ns2.example.com."],
  "mx_hosts": ["mail.example.com."],
  "ttl_spread": 295,
  "mixed_private_public_ip": false,
  "record_budget_exhausted": false
}
```

- `response_summary` 继续按记录类型提供 `rcode`、`answer_count`、TTL 统计、DNSSEC AD/RRSIG 等摘要。
- `mx_priorities` 和 `soa` 来自本次已有 MX/SOA 响应。
- `mixed_private_public_ip` 只表示本次 payload 中同时出现私网/特殊 IP 与公网 IP，需要人工确认上下文。
- TXT/SPF/DMARC/CAA 等外部文本继续做限长和 SHA256 摘要保护。

## 低频轻探测

`rdap`、`robots`、`security_txt` 是 v0.5.3 新增的低频旁路探测协议。它们默认关闭，只有 `collector.v2.enabled=true` 且对应 `collector.v2.light_probe.*.enabled=true` 时才会注册任务。

这些结果只用于域名治理、访客视角和安全联系渠道参考，不参与当前健康摘要结论，也不影响 Ping / HTTP / DNS / TLS 主采集。

### RDAP

RDAP 按注册域归并查询，同一轮中 `www.example.com` 与 `api.example.com` 只查询一次 `example.com`，再把结果写回对应 target 的 observation/latest Redis。

```json
{
  "registrable_domain": "example.com",
  "rdap_server": "https://rdap.example/",
  "registrar": "Example Registrar",
  "statuses": ["active"],
  "expires_at": "2030-01-01T00:00:00Z",
  "nameservers": ["ns1.example.com", "ns2.example.com"],
  "dnssec_delegation_signed": true,
  "events_summary": {
    "registration": "2020-01-01T00:00:00Z",
    "expiration": "2030-01-01T00:00:00Z"
  },
  "raw_truncated": false
}
```

- RDAP server 来自 IANA bootstrap JSON，进程内低频缓存，不新增 SQL 或 Redis 缓存。
- 失败时 `status=failure`，`error_code` 可能为 `rdap_bootstrap_failed`、`rdap_no_server`、`rdap_request_failed`、`rdap_decode_failed`。
- 不保存 RDAP 原始大 JSON，只提取治理摘要。

### robots.txt

robots.txt 按每个有效采集 target 低频请求一次 `/{robots.txt}`，scheme 复用采集域名的 TLS 配置。

```json
{
  "exists": true,
  "status_code": 200,
  "content_type": "text/plain",
  "sitemap_count": 2,
  "sitemaps": [
    "https://example.com/sitemap.xml"
  ],
  "global_disallow_all": false,
  "user_agent_star_present": true,
  "body_truncated": false
}
```

- 只记录最多 `max_sitemap_links` 条 sitemap 链接，不抓取 sitemap 内容。
- 404 会写入 `exists=false`，不视为 collector 程序错误。
- 响应体受 `max_response_bytes` 限制，不保存全文。

### security.txt

security.txt 按每个有效采集 target 低频请求 `/.well-known/security.txt`，404 时再 fallback 到 `/security.txt`。

```json
{
  "exists": true,
  "path_used": "/.well-known/security.txt",
  "status_code": 200,
  "content_type": "text/plain",
  "contact": ["mailto:security@example.com"],
  "expires": "2030-01-01T00:00:00Z",
  "policy": ["https://example.com/security-policy"],
  "preferred_languages": ["zh,en"],
  "canonical": ["https://example.com/.well-known/security.txt"],
  "body_truncated": false
}
```

- 只提取固定字段，不保存原始 security.txt 全文。
- 字段值最长 512 字符，列表限项，避免外部文本膨胀。
- 请求失败写 `status=failure` 和 `security_txt_request_failed`，不影响主采集。

### Page Assets

`page_assets` 基于 HTTP v2 latest 里已经存在的 HTML 声明工作，不会重新抓取首页。每个 target 单轮最多拉取 1 个 favicon 和 1 个 manifest；默认关闭。

```json
{
  "http_latest_found": true,
  "icon": {
    "exists": true,
    "source_url": "https://example.com/favicon.ico",
    "selected_rel": "icon",
    "selected_sizes": "32x32",
    "status_code": 200,
    "content_type": "image/png",
    "content_length_header": 1024,
    "body_read_bytes": 1024,
    "body_truncated": false,
    "sha256": "hex..."
  },
  "manifest": {
    "exists": true,
    "source_url": "https://example.com/site.webmanifest",
    "status_code": 200,
    "content_type": "application/manifest+json",
    "body_read_bytes": 512,
    "body_truncated": false,
    "sha256": "hex...",
    "name": "Example App",
    "short_name": "Example",
    "theme_color": "#ffffff",
    "background_color": "#000000",
    "display": "standalone",
    "start_url": "https://example.com/start",
    "scope": "https://example.com/",
    "icons_count": 2
  }
}
```

- 只允许 `http` / `https` 资源 URL。
- 默认只跟随 target 同注册域资源；跨站资源必须显式配置在 `allowed_asset_hosts`。
- 发起请求前会解析资源 host，解析到 loopback、private、link-local、multicast、unspecified 等地址时跳过，写入 `skipped_reason=asset_ip_not_allowed`；解析失败写入 `asset_dns_resolve_failed`。
- favicon 只保存元信息和 SHA256，不保存图片内容。
- manifest 只保存摘要字段和 SHA256，不保存原始 manifest JSON。
- 没有 HTTP latest、没有声明、资源 host 不允许、Content-Type 不允许、请求失败等场景会写 `exists=false` 和 `skipped_reason`，不影响主采集。

### Port Check

`port_check` 是 v0.5.5 新增的低频 TCP connect 轻探测协议。它默认关闭，启用后只检查配置中显式列出的端口；不内置默认端口列表，不抓 banner，不读取服务响应，不发送应用层协议 payload。

```json
{
  "ports_configured": 3,
  "ports_checked": 3,
  "open_count": 2,
  "closed_count": 1,
  "timeout_count": 0,
  "filtered_suspected_count": 0,
  "skipped_count": 0,
  "invalid_port_count": 0,
  "duplicate_port_count": 0,
  "truncated_port_count": 0,
  "truncated": false,
  "results": [
    {
      "port": 80,
      "service_hint": "http",
      "status": "open",
      "duration_ms": 8,
      "error_code": "",
      "error_message": ""
    },
    {
      "port": 3306,
      "service_hint": "mysql",
      "status": "closed",
      "duration_ms": 2,
      "error_code": "connection_refused",
      "error_message": "connect: connection refused"
    }
  ]
}
```

- `status` 只表示 TCP connect 结果，可能值为 `open`、`closed`、`timeout`、`filtered_suspected`、`skipped`。
- `service_hint` 只是静态端口名提示，例如 `3306=mysql`、`5432=postgresql`、`6379=redis`、`9090=prometheus`、`3000=grafana`；它不代表真实服务识别。
- Prometheus/Grafana 只作为普通端口提示，不接入 Prometheus 生态，不抓取 metrics，不访问 Web 页面。
- `port_check.enabled=true` 表示维护者已确认当前采集目标处于授权探测范围内；该结果不参与当前健康摘要和站点状态判断。

### WAF Canary

`waf_canary` 是 v0.5.7 新增的低频 WAF 规则无害验证协议。它默认关闭；启用 `collector.v2.light_probe.waf_canary.enabled=true` 即表示维护者确认当前有效采集目标均处于授权验证范围内。默认每 30 天执行一次，不做失败重试，不参与健康摘要。

```json
{
  "canary_path": "/.well-known/gofurry-waf-canary",
  "cases_total": 12,
  "cases_executed": 12,
  "blocked_count": 11,
  "expected_blocked_count": 11,
  "expected_blocked_matched_count": 11,
  "unexpected_pass_count": 0,
  "network_error_count": 0,
  "status_code_unexpected_count": 0,
  "max_targets_per_run": 0,
  "truncated_target_count": 0,
  "target_run_truncated": false,
  "cases": [
    {
      "case_id": "scanner_user_agent",
      "category": "scanner_user_agent",
      "method": "GET",
      "expected_blocked": true,
      "expected_status_codes": [403],
      "status_code": 403,
      "status_code_expected": true,
      "blocked": true,
      "matched_expected": true,
      "duration_ms": 12,
      "error_code": "",
      "error_message": ""
    }
  ]
}
```

- 默认 canary 路径为 `/.well-known/gofurry-waf-canary`，可通过 `canary_path` 覆盖；路径为空、绝对 URL、带 query/fragment 时会安全失败且不发起请求。
- 默认 `run_on_start=false`，启用后只注册周期任务，不会在服务启动瞬间立即跑全量 canary；本地测试可显式设置 `run_on_start=true`。
- 可配置 `max_targets_per_run` 限制单轮最多验证的 target 数量，默认 `0` 表示不限；被截断时写入 `truncated_target_count` 和 `target_run_truncated=true`。
- 每个 target 执行一个 baseline GET 和 11 个基础规则 case：扫描器 UA、参数数量、URI 长度、SQLi 关键词、XSS 标记、命令注入标记、路径穿越标记、非法方法、非法 Content-Type、JSON 解析错误、JSON 危险关键字。
- 请求带 `GoFurry-Nav-Collector-WAF-Canary/1.0` 类 User-Agent，便于 WAF 日志识别；不会携带 Cookie、Authorization、账号、Token 或真实用户数据。
- `blocked=true` 只按响应码判断：`400`、`403`、`405`、`414`、`415` 视为拦截类响应；`404` 不视为拦截。
- `matched_expected` 只表示“是否按预期拦截/放行”；状态码是否等于建议值由 `status_code_expected` 单独表达，差异会计入 `status_code_unexpected_count`，避免把 400/403/405/414/415 之间的拦截状态码差异误判为未拦截。
- payload 只保存 case_id、分类、状态码、耗时和错误类别，不保存完整 query、body、URI 或测试样本原文。
- 该能力是已授权站点的低频 WAF 回归验证，不是漏洞扫描、绕过测试、目录枚举或高频压测。
