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
