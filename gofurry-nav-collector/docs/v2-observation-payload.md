# v2 Observation Payload 字段说明

本文档记录 `gofurry-nav-collector` 写入 `gfn_collector_observation.payload` 与 `collector:v2:latest:{protocol}:{site_id}` 的主要字段。所有字段都只是 observation 信号，用于后端 v2、前端详情页和人工排查参考，不直接等同于站点健康结论。

旧 Redis key、旧日志表和旧前端展示结构仍是兼容路径；新增字段只在 v2 observation/latest Redis 中旁路出现。

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
