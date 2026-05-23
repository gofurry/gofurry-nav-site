# Roadmap

## Current Position

`gofurry-nav-collector` 已经负责 Ping、HTTP/TLS、DNS 三类采集，并将 latest 数据写入 Redis、历史数据写入 PostgreSQL。当前生产风险主要集中在采集重入、网络操作超时、GeoIP 文件打开频率、Redis/DB 峰值和大表日志清理。

## Roadmap Strategy

优先完成不会改变旧接口、旧 Redis key、旧表语义的稳定性工作。后续能力必须保持低频、非侵入、可灰度、可关闭、可回滚；不引入 Prometheus 生态，不做漏洞扫描、端口全扫、目录爆破、弱口令尝试或高强度探测。

## Version Plan

### v0.1.0 - Phase 0 Stability

**Status:** In progress
**Scope:** Stability / Safety / Testing
**Goal:** Make the existing collector safer to run against production-scale data.

#### Focus

- prevent overlapping collector runs
- bound network and Redis operations
- reduce DB and filesystem pressure
- keep old user-facing behavior unchanged

#### Tasks

- [x] Add Ping / HTTP / DNS non-overlap guards.
- [x] Replace global per-protocol `WaitGroup` state with per-run worker state.
- [x] Add configurable probe budget defaults for Redis, HTTP, TLS, DNS, PTR, response size, and DNS record count.
- [x] Open GeoIP databases once for the DNS module lifecycle and tolerate partial GeoIP availability.
- [x] Move retention cleanup out of every collector run and keep it on scheduled cleanup tasks.
- [x] Replace retention delete SQL with stable batched CTE cleanup.
- [x] Add manual concurrent index SQL for retention queries.
- [x] Add focused tests for defaults, retention SQL shape, and nil GeoIP readers.

#### Acceptance Criteria

- Collector runs skip instead of overlapping when a previous run is still active.
- DNS/PTR/HTTP operations have bounded timeouts.
- Redis commands use per-command timeout contexts.
- Existing Redis keys, database tables, and backend APIs keep their current behavior.
- `gofmt`, `go vet`, and `go test` pass for `gofurry-nav-collector` and `gofurry-nav-backend`.

---

### v0.2.0 - v2 Data Plane

**Status:** Planned
**Scope:** Architecture / Backend API / Safety
**Goal:** Add a sidecar v2 data path without switching production reads.

#### Focus

- observation table
- v2 latest Redis keys
- read-only backend v2 endpoints
- v1/v2 comparison

#### Tasks

- [ ] Add `gfn_collector_observation` with JSONB payloads.
- [ ] Add `collector:v2:latest:{protocol}:{site_id}` keys behind feature flags.
- [ ] Keep old table and old key writes unchanged during dual-write.
- [ ] Add backend read-only collector latest / summary endpoints behind feature flags.
- [ ] Add comparison logs for sampled v1/v2 results.

#### Acceptance Criteria

- v2 writes can be fully disabled without changing v1 behavior.
- v2 data volume and write rate are measurable through logs and SQL/Redis checks.
- Backend can read v2 latest data without changing existing `/api/nav` routes.

---

### v0.3.0 - Protocol Semantics

**Status:** Planned
**Scope:** User-facing / Stability / Safety
**Goal:** Improve protocol result quality while keeping probing low intensity.

#### Focus

- TLS verification semantics
- HTTP redirect/header collection
- DNS risk flags
- Ping as auxiliary signal

#### Tasks

- [ ] Split TLS certificate collection from certificate verification.
- [ ] Add HTTP redirect chain and security header presence checks without payload probing.
- [ ] Replace strong DNS hijack wording with risk flags in v2 payloads.
- [ ] Treat Ping as an auxiliary connectivity signal, not a site-down decision.

#### Acceptance Criteria

- New protocol fields are explicitly marked as observations, not final judgments.
- External title, meta, header, TXT, PTR, and certificate strings are treated as untrusted text.
- No protocol change increases probe frequency or probe intensity by default.

---

### v1.0.0-alpha.1 - API Freeze Candidate

**Status:** Planned
**Scope:** Release / Documentation / Compatibility
**Goal:** Freeze collector v2 payload candidates and prepare for stable operation.

#### Focus

- documented payload schemas
- migration and rollback notes
- compatibility checks
- production operating guide

#### Tasks

- [ ] Document v2 payload schema versions and compatibility policy.
- [ ] Document production rollout and rollback procedures.
- [ ] Add regression tests for v1 compatibility and v2 read behavior.
- [ ] Collect production feedback before promoting v2 reads.

#### Acceptance Criteria

- v1 behavior remains available during the alpha period.
- Operators can disable any v2 protocol or read path through configuration.
- Known blocker-level safety issues are resolved before `v1.0.0`.
