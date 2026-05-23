# gofurry-nav-collector 代码审计报告

审计日期：2026-05-24

审计范围：`gofurry-nav-collector` 当前 `dev` 分支。重点关注线上稳定性、采集结果可信度、旧链路兼容、v2 observation 数据质量，以及后续单人维护成本。

审计结论：当前 collector 已经具备较清晰的低频旁路采集架构，Phase 0、v0.2、v0.3 的方向是稳妥的；本次没有发现会直接破坏线上旧接口或导致高强度探测的 P0 问题。但存在几个会影响数据可信度、错误暴露和维护预期的问题，建议归入 `v0.3.1` 优先修复。

## 严重程度定义

- **P0：** 会导致线上核心链路不可用、数据大面积破坏、不可控高强度探测或安全事故。
- **P1：** 高概率影响生产可靠性、关键数据可信度或回滚能力，需要优先修复。
- **P2：** 中等风险问题，会造成局部数据失真、错误被吞、配置不符合预期或维护风险。
- **P3：** 低风险问题，主要影响可读性、可维护性、诊断体验或未来扩展安全边界。

## Findings

### P1：DNS GeoIP reader 参数顺序错误，ASN / Country 信息可能失真

位置：

- `collector/dns/service/dnsService.go:534`
- `collector/dns/service/dnsService.go:555`
- `collector/dns/service/dnsService.go:592`
- `collector/dns/service/dnsService.go:653`

`performDNSQuery` 的签名是 `asnDB, cityDB, countryDB`，但调用 `queryDNS` 时传入的是 `countryDB, cityDB, asnDB`。`queryDNS` 内部又按 `lookupGeoASN(v.A, countryDB, cityDB, asnDB)` 查询，实际会把 ASN reader 和 Country reader 对调。

影响：

- DNS A / AAAA 记录的国家、ASN、ISP 字段可能大量降级为 `Unknown`，或者在不同 mmdb 查询结构下得到错误结果。
- v2 observation 的 DNS 访客视角参考价值下降，后续如果基于 ASN/CDN 做安全提示会放大误判。

建议：

- 统一 `queryDNS` 参数顺序，建议改为 `countryDB, cityDB, asnDB`，和 `lookupGeoASN` 保持一致。
- 为 `queryDNS` / `lookupGeoASN` 增加最小单元测试，覆盖 nil reader 降级和 reader 传参顺序。
- 修复后用一个已知 IP 的 mmdb 查询结果做本地验证。

### P2：Ping 旧表写入错误被忽略，且 DB / v2 写入处于全局锁内

位置：

- `collector/ping/service/pingService.go:429`
- `collector/ping/service/pingService.go:436`
- `collector/ping/service/pingService.go:438`

Ping 结果写入旧表时调用 `dao.GetPingDao().Add(pindSaveRecord)`，但没有检查返回错误。同时，`pingRWLock` 包住了 Redis 结果 map 更新、旧表写入和 v2 observation 写入。

影响：

- 旧表写入失败时，日志和上层流程无法感知，容易误以为 Ping 历史记录正常。
- DB 或 Redis v2 写入变慢时，会拖住全局锁，降低 Ping worker 并发实际收益。
- 一次外部存储抖动可能放大成整轮 Ping 收尾变慢。

建议：

- 只在修改共享 `data` map 时持有 `pingRWLock`。
- 旧表写入和 v2 observation 写入移出锁外，并分别记录中文结构化错误日志。
- 补充测试：旧表写入失败时不影响旧 Redis map 更新，但必须有日志或返回路径可观测。

### P2：Redis `HDel` 吞掉删除错误，清理失败会被误报为成功

位置：

- `common/service/redisService.go:204`
- `common/service/redisService.go:211`
- `common/service/redisService.go:215`
- `collector/ping/service/pingService.go:304`

`HDel` 在 Redis 返回错误时记录日志，但最终返回 `intVal, nil`。调用方会认为清理成功，尤其是 Ping 过期 Redis 结果清理会把失败当作正常完成。

影响：

- Redis 旧结果可能长期残留，影响前端或后端读取到的 Ping 展示。
- 生产排查时，日志可能有错误，但业务日志会继续记录“清理完成”，造成判断偏差。

建议：

- `HDel` 在 `err != nil` 时返回 `common.NewServiceError("删除缓存失败.")`。
- 检查类似包装函数，至少让写入、删除类 Redis 命令不吞错。
- 为 `HDel` 增加单测或轻量 fake 测试，验证错误透传。

### P2：HTTP 代理配置解析错误被忽略

位置：

- `collector/http/service/httpService.go:496`
- `collector/http/service/httpService.go:497`
- `collector/http/service/httpService.go:498`

当采集域名配置 `proxy="1"` 时，代码会解析全局代理地址，但忽略 `url.Parse` 错误。代理配置错误时，实际行为不够明确，可能导致请求失败、绕过预期代理或难以诊断。

影响：

- 生产中部分站点依赖代理可达，代理配置错误时会产生大量 HTTP 探测失败。
- 日志缺少“代理配置不可用”的明确字段，排障成本高。

建议：

- 代理 URL 解析失败时，当前目标 HTTP 探测应明确失败，并写入 `http_proxy_config_invalid`。
- 日志包含 `site_id`、`target`、`proxy`，但注意不要输出敏感认证信息。
- 如果未来支持带账号密码的代理，日志中必须脱敏。

### P2：数据库初始化忽略 `gorm.DB()` 错误，DSN 拼接缺少转义边界

位置：

- `roof/db/db.go:40`
- `roof/db/db.go:47`

数据库 DSN 通过 `fmt.Sprintf` 拼接，随后 `sqlDB, _ := db.engine.DB()` 忽略错误。当前生产密码没有触发问题，但一旦密码、用户名或库名包含空格等特殊字符，DSN 解析风险会升高；如果 `DB()` 返回错误，后续 `sqlDB.Set...` 可能直接 panic。

影响：

- 配置变更时可能出现启动失败或难以理解的数据库连接错误。
- 忽略 `DB()` 错误会降低启动阶段诊断质量。

建议：

- 检查并处理 `db.engine.DB()` 错误。
- 用 pgx / postgres driver 推荐的配置构造方式，或至少对 DSN 参数做安全拼接。
- 启动日志中输出 DB host/port/dbname，不输出用户名密码。

### P3：`collector.dns.query_thread` 配置未实际约束单目标 DNS 查询并发

位置：

- `roof/env/config.go:68`
- `collector/dns/service/dnsService.go:550`

配置中存在 `query_thread`，但 `performDNSQuery` 会对 `models.RecordTypes` 的每种记录类型直接启动 goroutine。当前记录类型数量不大，风险有限，但配置含义与实际行为不一致。

影响：

- 运维以为调小 `query_thread` 可以降低 DNS 单目标并发，实际不会生效。
- 后续增加记录类型时，单目标查询并发会随记录类型数量增长。

建议：

- 在 DNS 记录类型查询处加入局部 semaphore，使用 `collector.dns.query_thread` 限制单目标查询并发。
- 当配置小于等于 0 时使用安全默认值。
- 在日志中记录实际 `query_thread`。

### P3：配置字段和启动日志仍有可维护性问题

位置：

- `roof/env/config.go:92`
- `roof/env/config.go:187`
- `roof/env/config.go:198`

`ServerConfig.Mode` 的 yaml tag 写成了 `yaml:"models"`，生产配置里的 `mode` 不会被正确加载。配置查找阶段仍使用 `fmt.Println` 输出路径，绕过 zap 日志体系。

影响：

- `server.mode` 配置语义不可靠。
- 启动日志和正式日志格式不一致，容易污染 systemd 日志和人工排查视线。

建议：

- 修正 yaml tag 为 `yaml:"mode"`。
- 配置查找阶段改为可控的启动日志；如果 logger 初始化顺序不允许，至少减少默认 stdout 输出。
- 增加配置加载测试，覆盖 `mode`、`probe_budget`、`v2` 默认值。

### P3：Retention SQL builder 接受任意表名，建议加 allowlist

位置：

- `common/retention/retention.go:13`
- `common/retention/retention.go:35`

当前调用方使用常量表名，暂未发现用户输入直接进入 tableName。但 SQL builder 通过 `fmt.Sprintf` 拼接表名，未来扩展时容易被误用。

影响：

- 当前风险低，主要是未来维护边界不够清晰。

建议：

- 对 retention 表名做 allowlist，例如只允许三张旧日志表和 `gfn_collector_observation`。
- 或把 builder 改为不暴露任意 tableName 的函数，改用枚举协议生成 SQL。

## 后续可采集的高价值字段

这些字段都应继续遵守现有原则：默认低频、低强度，不做漏洞扫描，不增加默认探测次数，先写 v2 observation，不影响旧 Redis key、旧表和旧展示链路。

### Ping

- `min_rtt_ms`、`max_rtt_ms`、`stddev_rtt_ms`：从现有 Ping 统计直接获得，能表达网络抖动。
- `jitter_ms`：可先用 `stddev_rtt_ms` 近似，作为访客访问稳定性的参考信号。
- `packets_sent`、`packets_recv`：让丢包率更可解释。
- `resolved_ip`：记录本轮 ICMP 实际目标 IP，辅助排查 DNS 与 Ping 目标不一致。

### HTTP

- `dns_lookup_ms`、`tcp_connect_ms`、`tls_handshake_ms`、`ttfb_ms`、`transfer_ms`：使用 `httptrace` 从同一次 GET 中获得，不增加请求数量。
- `http_protocol`：记录 HTTP/1.1、HTTP/2 等实际协议。
- `remote_addr` / `remote_ip`：记录最终连接到的服务端地址，帮助对照 DNS 结果。
- `content_length`、`body_read_bytes`、`compressed`：表达页面对访客的体感大小。
- `cache_policy`：摘要化 `cache-control`、`etag`、`last-modified`，用于访客性能参考。
- 安全 header 细化摘要：例如 HSTS `max-age`、CSP 是否存在 `unsafe-inline`，仍只作为 observation 信号。

### TLS

- `cert_not_before`、`cert_not_after`、`cert_chain_length`：让证书生命周期更可解释。
- `cert_subject_cn`、`cert_san_count`：辅助定位证书覆盖范围。
- `cert_signature_algorithm`、`cert_public_key_algorithm`：提供基础加密配置参考。
- `ocsp_stapled`、`sct_count`：从当前 TLS 连接状态读取，作为证书透明度和吊销响应参考。
- `verify_error_category`：把原始错误归类为 `expired`、`hostname_mismatch`、`unknown_authority` 等，便于后端和前端解释。

### DNS

- `rcode`、`authoritative`、`truncated`、`recursion_available`：记录 DNS 响应语义，辅助判断解析质量。
- `answer_count`、`authority_count`、`additional_count`：表达响应规模。
- `ttl_min`、`ttl_max`、`ttl_avg`：当前已有统计变量，可进入 v2 payload。
- `cname_chain_depth`：辅助判断跳转链过长或解析复杂度。
- `mx_priorities`、`soa_serial`、`soa_refresh`、`soa_retry`、`soa_expire`：从已有记录中摘要，不额外查询。
- `dnssec_rrsig_present`、`dnssec_ad`：区分“响应里有签名记录”和“resolver 声称已验证”。

## 推荐落地顺序

1. `v0.3.1` 先修复本报告中的 P1/P2/P3 可靠性问题，尤其是 DNS GeoIP 参数顺序和错误吞掉问题。
2. `v0.3.2` 再扩展“同一次探测可顺手得到”的高价值字段，避免增加默认探测强度。
3. `v0.4.0` 基于稳定 observation 做健康状态聚合，继续保持旁路和可回滚。
