# GoFurry Nav Collector Roadmap

## 当前状态

`gofurry-nav-collector` 的 v2 采集核心已经完成：Ping、HTTP、TLS、DNS、RDAP、robots.txt、security.txt、page assets、port check、edge hints、WAF canary、v2 observation、latest Redis、summary Redis 与 run state 均已具备基础能力。

本 roadmap 只记录后端正式进入 `/api/v2/nav` 之前，collector 侧最后一轮收口工作。目标不是继续扩大探测面，而是让现有数据更好解释、更好治理、更适合被后端 v2 API 稳定消费。

## 迭代原则

- 不新增默认采集频率、请求次数、DNS 查询目标、端口数量或并发强度。
- 不改变旧 Redis key、旧日志表、旧后端接口和旧前端展示结构。
- 不引入 Prometheus 生态，不做漏洞扫描，不做目录扫描，不做弱口令或绕过测试。
- 现有主动轻探测继续默认关闭或低频，且只用于自有或明确授权目标。
- v0.6.x 的重点是解释、治理、趋势和变化检测，而不是继续堆字段。
- 本 roadmap 不包含后端 `/api/v2/nav` 实施；后端 v2 API 由 `gofurry-nav-backend/docs/roadmap.md` 单独维护。

## 版本计划

### v0.6.0 - v2 数据契约与解释字典收口

**状态：** 已完成
**范围：** 数据契约 / reason code / 文档 / 测试
**目标：** 固化 collector v2 对外可依赖的数据语义，让后端 v2 API 可以稳定消费 observation、latest 和 summary。

#### Focus

- 明确 observation payload 的 schema 边界。
- 标准化 summary status、reason code 和中文解释。
- 把“观测信号”和“健康结论”继续分开。

#### Tasks

- [x] 为所有 summary `reason_codes` 建立固定字典，包含中英文说明、英文 key、建议严重度、是否影响健康状态。
- [x] 为 Ping / HTTP / TLS / DNS / light probe payload 增加 schema 文档表格，标注字段类型、可空性、来源和版本。
- [x] 在文档中明确哪些字段是 observation 信号，哪些字段参与 summary 聚合。
- [x] 为 summary status 增加统一说明：`healthy`、`warning`、`degraded`、`unknown`、`down`。
- [x] 为 reason code builder 增加单元测试，避免后续改动产生拼写漂移。
- [x] 给 `docs/v2-observation-payload.md` 增加“后端 v2 消费建议”章节，但不实现后端接口。

#### Acceptance Criteria

- 后端开发者不需要读 collector 源码，也能理解每个 reason code 和 payload 字段。
- 新增或修改 reason code 必须有测试覆盖。
- 文档明确所有主动探测能力的授权边界和默认关闭/低频语义。

---

### v0.6.1 - 采集目标治理与 canonical target 收口

**状态：** 已完成
**范围：** 目标治理 / 域名关系 / 数据一致性
**目标：** 降低 `example.com`、`www.example.com`、跳转后域名、备用域名之间的展示割裂，为后端 v2 API 提供更清晰的 target 关系。

#### Focus

- 不改变采集目标来源。
- 不新增 SQL 表，优先基于现有 `gfn_collector_domain` 和 latest 数据生成旁路字段。
- 让 target 关系可解释、可回滚。

#### Tasks

- [x] 在 HTTP v2 payload 或 summary 中补充 `canonical_target_hint`，优先来自 canonical URL、final URL host、HTTP target host 的保守比较。
- [x] 为 target summary 增加 `target_relation_hints`，表达 `same_host`、`redirect_to_www`、`redirect_to_apex`、`redirect_to_external` 等低风险关系。
- [x] 标记 `final_url` 与采集 target host 不一致的情况，只作为提示，不改变采集目标。
- [x] 对同一 site 下多个 target 的 summary 增加重复 host / 重复 final host 检测。
- [x] 文档化 target 关系只是治理参考，不自动合并、不自动删除、不改变 admin 管理数据。

#### Acceptance Criteria

- 不自动修改 `gfn_collector_domain`。
- 不影响旧 Redis、旧表、旧前端。
- 后端 v2 可以用这些 hints 做展示，但不能把它当作强制跳转或自动归并依据。

---

### v0.6.2 - Observation 历史趋势派生

**状态：** 已完成
**范围：** 历史趋势 / 派生数据 / Redis summary
**目标：** 在不增加采集强度的前提下，从已有 observation 中派生近周期趋势，让站点状态更有时间维度。

#### Focus

- 基于已有 `gfn_collector_observation`，不新增探测。
- 先做小窗口、低成本查询。
- 默认不参与旧链路。

#### Tasks

- [x] 为每个 site + target + protocol 派生近 24 小时 / 7 天的基础趋势摘要。
- [x] HTTP 趋势：成功率、平均响应时间、P95 响应时间、最近失败时间。
- [x] Ping 趋势：成功率、平均 RTT、丢包率均值、抖动均值。
- [x] DNS 趋势：成功率、TTL min/max/avg 变化、risk_flags 出现次数。
- [x] TLS 趋势：证书剩余天数变化、证书 issuer / fingerprint 是否变化。
- [x] 将趋势结果写入新的 v2 Redis summary key，避免影响当前 target/site summary。
- [x] 增加查询预算、超时和 batch 限制，避免 observation 历史查询拖垮数据库。

#### Acceptance Criteria

- 不新增 SQL migration；如后续发现需要索引，只先写手动 SQL 建议，不自动执行。
- 趋势派生失败只记录日志，不影响主采集。
- 趋势结果能被后端 v2 API 直接读取，但本阶段不实现后端接口。

---

### v0.6.3 - 变化检测事件

**状态：** 已完成
**范围：** 变化检测 / 事件摘要 / 可解释性
**目标：** 从 latest 与历史 observation 中识别“发生了什么变化”，让站点维护者看到比单次状态更有价值的事件。

#### Focus

- 只检测低噪声、高价值变化。
- 不做告警系统。
- 不影响站点上下架、排序或健康结论。

#### Tasks

- [x] HTTP 变化：title、server、x-powered-by、security headers、final URL、status code 明显变化。
- [x] TLS 变化：证书 fingerprint、issuer、SAN 数量、not_after 变化。
- [x] DNS 变化：A/AAAA、CNAME terminal、MX、NS、SOA serial 变化。
- [x] Port check 变化：端口从 closed/timeout 变 open，或 open 变 closed。
- [x] RDAP 变化：domain status、expires_at、nameserver 列表变化。
- [x] 变化事件只写摘要，不保存外部大文本原文。
- [x] 增加去抖策略，避免同一变化在短时间内重复写入。

#### Acceptance Criteria

- 变化事件可解释，包含 old/new 摘要、source protocol、observed_at。
- 默认不发送通知、不触发前端强提示。
- 不把一次失败当成稳定变化，避免网络抖动造成噪声。

---

### v0.6.4 - Collector v2 收口验收

**状态：** 已完成
**范围：** 质量验收 / 文档 / 测试 / 生产准备
**目标：** 在后端正式推进 `/api/v2/nav` 前，对 collector v2 数据面做一次最终验收，确认可以作为稳定数据来源。

#### Focus

- 数据契约完整。
- 测试覆盖核心 builder 和边界。
- 生产配置保持克制。
- 后续新增能力进入 v0.7.x 或后端 v2 之后再评估。

#### Tasks

- [x] 梳理 v2 Redis key 清单：latest、summary、run state、trend、change event。
- [x] 补齐关键 builder 的单元测试：summary、trend、change event、reason code 字典。
- [x] 增加一份本地/测试环境 smoke test 文档，说明如何验证 collector v2 数据完整性。
- [x] 检查生产推荐配置，确保新能力默认关闭或低频，不扩大当前站点压力。
- [x] 更新 `docs/v2-observation-payload.md`，标记进入后端 v2 API 前的稳定字段集合。
- [x] 明确冻结范围：后端 v2 API 开始后，collector v2 字段只做兼容新增，不做破坏性重命名。

#### Acceptance Criteria

- `gofmt -l .` 为空。
- `go test ./...` 通过。
- `go vet ./...` 通过。
- 文档能支持后端 v2 API 开发，不需要继续追 collector 源码确认字段含义。
- collector v2 后续新增字段必须保持 additive，不破坏已稳定字段。

---

### v0.6.5 - 后端 v2 前审计修复

**状态：** 计划中
**范围：** 安全 / 稳定性 / 性能 / 配置治理 / 测试
**目标：** 修复 `docs/code-audit-v0.6.5-20260525.md` 中进入后端 v2 前必须处理的 collector 风险，确保 v2 数据源可以更稳地被后端消费。

#### Focus

- 先处理可达依赖漏洞和工具链风险。
- 控制 observation 历史派生查询成本。
- 收敛主动 light probe 的启动行为和配置默认值。
- 清理仓库内真实配置和遗留不安全 helper。

#### Tasks

- [ ] 升级 Go toolchain 到 `1.26.3+`，升级 `golang.org/x/net` 到 `v0.55.0+`，重新运行 `govulncheck`。
- [ ] 新增手动 SQL 脚本，为 `gfn_collector_observation` 补充 `(site_id, target, protocol, observed_at DESC, id DESC)` 并发索引。
- [ ] 为 trend / change event 派生增加更明确的查询预算、开关或去抖，避免每次 summary 更新都无条件放大历史查询。
- [ ] 为 RDAP、robots.txt、security.txt、page_assets、port_check 增加统一 `run_on_start` 配置，默认关闭；保留显式开启时的立即运行能力。
- [ ] 将仓库内 `conf/server.yaml` 收敛为安全示例配置或拆出 `conf/server.example.yaml`，真实本地配置放入 ignored local 文件。
- [ ] 删除或加固 `common/util/http.go` 中无调用点的遗留 HTTP helper，移除 `InsecureSkipVerify`、无界 `ReadAll` 和无界 HTML parse 风险。
- [ ] 缩小 RDAP bootstrap 缓存锁范围，避免持锁执行网络 I/O。
- [ ] 为上述修复补充单元测试和配置加载测试。

#### Acceptance Criteria

- `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` 不再报告 collector 可达漏洞；如有残留项必须写明接受风险。
- `gofmt -l .` 为空。
- `go test ./...` 通过。
- `go vet ./...` 通过。
- `git diff --check` 通过。
- 仓库内不再跟踪真实 DB / Redis 密码。
- light probe 默认不会因服务重启立即发起全量低频探测。
- trend / change event 的历史查询有索引或预算保护。

#### Notes

- 本阶段仍不新增采集协议、不新增默认探测强度、不改旧 Redis key、旧表、旧接口或前端展示。
- SQL 索引脚本只手动执行，生产执行时必须确认 `CREATE INDEX CONCURRENTLY` 不在事务块中。
- 修复完成后再进入后端 `/api/v2/nav` 的正式实现会更稳。

## 暂不计划

- 不做新的主动协议扩张。
- 不做高频探测。
- 不做漏洞扫描、目录扫描、弱口令验证或攻击样本库。
- 不把 WAF canary 扩成通用安全测试平台。
- 不新增分布式分片调度；当前站点规模单实例足够。
- 不在 collector roadmap 中安排后端 `/api/v2/nav` 接口实现。
