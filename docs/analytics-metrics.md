# Analytics metric operations

V3 P0.3 metrics are deterministic projections of finalized P0.2 Historical Facts. The existing Game and Nav Collectors own the Metric Engines; no new service, acquisition Job, Run, Task, or queue is created.

The runtime order is:

```text
Acquisition Control -> Historical Facts -> Analytics Metrics
```

Each active metric version reconciles once at startup and every `metrics.reconcile_interval_seconds` (default `600`). One reconcile processes at most one UTC day per metric. Game waits for `game.state_facts`; Nav waits for both `nav.target_facts` and `nav.site_facts`.

## Registry and state contract

Goose owns the Registry. Runtime and Admin read it but do not mutate it. Every active or retired Registry version must retain a matching compiled evaluator and full contract so explicit historical rebuild remains valid. Current evaluator contracts:

| Domain | Metric/version | Freshness | Dimensions |
|---|---|---:|---|
| Game | `free_game_share/1` | 72h | `primary_tag_id`, `tag_id` |
| Game | `windows_support/1` | 72h | `primary_tag_id`, `tag_id` |
| Game | `linux_support/1` | 72h | `primary_tag_id`, `tag_id` |
| Nav | `ipv6_adoption/1` (retired) | 72h | `site_country`, `group_id`, `nsfw`, `welfare` |
| Nav | `ipv6_adoption/2` | 72h | `site_country`, `group_id`, `nsfw`, `welfare` |
| Nav | `tls13_adoption/1` | 48h | `site_country`, `group_id`, `nsfw`, `welfare` |
| Nav | `http2_adoption/1` (retired) | 48h | `site_country`, `group_id`, `nsfw`, `welfare` |
| Nav | `http2_adoption/2` | 48h | `site_country`, `group_id`, `nsfw`, `welfare` |
| Nav | `hsts_adoption/1` | 48h | `site_country`, `group_id`, `nsfw`, `welfare` |
| Nav | `csp_adoption/1` | 48h | `site_country`, `group_id`, `nsfw`, `welfare` |
| Nav | `tls_certificate_verification/1` | 48h | `site_country`, `group_id`, `nsfw`, `welfare` |
| Nav | `security_txt_adoption/1` (retired) | 21d | `site_country`, `group_id`, `nsfw`, `welfare` |
| Nav | `security_txt_adoption/2` | 21d | `site_country`, `group_id`, `nsfw`, `welfare` |

Entity state is one of `positive`, `negative`, `stale`, `not_probed`, `probe_failed`, `unknown`, or `not_applicable`. Every row records a reason and provenance. Freshness is relative to `fact_date + 1 day 00:00:00Z`; evidence after that boundary aborts the transaction.

Nav capability v2 preserves evidence uncertainty. `ipv6_adoption/2` is positive for an AAAA answer, negative only for a successful NOERROR AAAA query with no AAAA answer, and unknown when that subquery failed or its historical evidence is unavailable. `security_txt_adoption/2` is positive only for a recognized plain-text document with a meaningful Contact and future Expires; an absent document is negative, while an invalid/unrecognized 2xx document is unknown. Scheduled failure/no-attempt states continue to come from acquisition-ledger quality, and carried-forward known state keeps its original observation timestamp.

P2.3 adds HTTP/2 negotiated-protocol, HSTS presence, enforcement-CSP presence, and X.509 certificate-verification contracts. Missing protocol/header/verification fields remain unknown, plain HTTP is HSTS `not_applicable`, handshake failure is not certificate-negative, and Report-Only CSP is not enforcement CSP. These contracts are registered by Goose without an in-migration history rewrite; operators backfill Metrics before their dependent Changes.

### HTTP/2 negotiation v2 (#92-A)

`http2_adoption/2` measures negotiation from the GoFurry egress and actual proxy path
to the **final response endpoint**, not universal website HTTP/2 support. The HTTP
Transport explicitly enables `ForceAttemptHTTP2`; the static `Connection` header is
removed. Successful responses carry `http2_probe_profile=go_h2_enabled_v1`,
`http_protocol` from `resp.Proto`, `final_url`, derived `final_scheme`, and `tls_alpn`
from the final `resp.TLS.NegotiatedProtocol`. Empty ALPN is recorded, distinct from a
missing field. The HTTP Known State whitelist preserves these fields, not proxy
URLs, credentials, Authorization or Cookie.
The profile is emitted only after the production Transport has actually registered
its H2 handler; an override such as `GODEBUG=http2client=0` leaves it untrusted,
so a disabled-client H1 response remains unknown rather than negative.

| Final evidence | State / reason |
| --- | --- |
| Trusted profile, HTTPS, HTTP/2.0 and ALPN `h2` | positive / `http2_negotiated` |
| Trusted profile, HTTPS, HTTP/1.0 or HTTP/1.1 and empty / `http/1.1` ALPN | negative / `http1_negotiated_h2_enabled` |
| Trusted profile, HTTP, H1 and empty ALPN | not_applicable / `http_plaintext_not_applicable` |
| Missing/old profile, incomplete or contradictory endpoint/protocol/ALPN, unsupported protocol | unknown / `http2_probe_profile_unknown`, `http2_evidence_incomplete`, `http2_endpoint_inconsistent`, `http2_protocol_alpn_inconsistent`, `http2_protocol_unknown` |
| Original successful evidence older than 48h at UTC day end | stale / `http_state_stale` |
| No usable successful state; ledger has failed attempts / no attempts | probe_failed / `http_probe_failed`; not_probed / `http_not_probed` |

Primary identity and finalized Fact requirements remain mandatory; missing identity,
target Fact or historical acquisition evidence stays unknown. Carried success keeps
its original `known_state_observed_at`, even after later failed/unattempted probes.
Future evidence aborts projection; status codes and independent certificate verification
do not decide protocol support. Adoption uses only positive + negative; zero known
means null adoption, including a nonempty eligible population. Plain HTTP is excluded
from eligible. IPv6/security.txt v2 and all older evaluators keep their semantics.

Migration `20261018010000_nav_http2_probe_metric_v2.sql` retires `http2_adoption/1`
and, with explicit maintainer authorization, its dependent `http2_transition/1`.
Registry rows, checkpoints, Metric/Change history and old evaluators are retained;
explicit v1 rebuild remains available, automatic reconciliation/backfill stops.
There is **no `http2_transition/2`** in this phase. Admin already preserves version
and registry status and selects active versions for Overview; v1/v2 are not stitched.

The independent v2 checkpoint starts no earlier than the later of both Fact pipeline
source dates and the next UTC day after migration; missing upstream start dates keep
it unavailable. This is a lower bound only. Production T is the first complete UTC
day after the new Collector has been deployed and its profile/protocol validated,
also satisfying that lower bound. Production migration/deployment needs separate
authorization: pause the old Collector, migrate, deploy and verify the new Collector,
then process Facts before Metrics for T. If rollout crosses UTC days or validation
fails, stop reconciliation and obtain an explicitly approved forward checkpoint/
cutover correction; never backdate v2 or destructively roll back. Historical v1
results are not reinterpreted. See [development acceptance](acceptance/92-a-http2-probe-metric-v2.md).

`gfg_metric_daily` and `gfn_metric_daily` store state counts only. Query-time consumers derive:

```text
known    = positive_count + negative_count
adoption = positive_count / known
coverage = known / eligible_count
```

A zero denominator yields null. Global `global/all` is always materialized, including an all-zero row for an empty population. Multi-membership tag/group values are deduplicated per entity and slice; no multi-dimensional cube is generated.

## CLI

Commands use the normal explicit Collector config file:

```text
gf-game-collector metrics status --config /path/server.yaml
gf-game-collector metrics backfill --config /path/server.yaml --dry-run
gf-game-collector metrics backfill --config /path/server.yaml --metric free_game_share --version 1 --from 2026-08-01 --through 2026-08-07 --max-days 7
gf-game-collector metrics rebuild --config /path/server.yaml --metric free_game_share --version 1 --from 2026-08-01 --to 2026-08-07 --dry-run

gf-nav-collector metrics status --config /path/server.yaml
gf-nav-collector metrics backfill --config /path/server.yaml --metric ipv6_adoption --version 2 --dry-run
gf-nav-collector metrics backfill --config /path/server.yaml --metric tls_certificate_verification --version 1 --through YYYY-MM-DD
gf-nav-collector metrics rebuild --config /path/server.yaml --metric ipv6_adoption --version 2 --from 2026-08-31 --through 2026-09-07
```

`--to` is an alias for inclusive `--through`; `--max-days=0` is unlimited. Backfill cannot skip the ordered checkpoint's next day. Rebuild locks the metric version, requires the range to have been processed already, leaves `processed_through` unchanged, and supports retired versions when explicitly selected.

## Admin and rollout

The authenticated Admin Metric Center is read-only and exposes Overview, Registry, Checkpoints, Daily Results, and Entity Explorer. Adoption/coverage are calculated by the Admin query layer. Daily and Entity results are count-backed pages; long dimensions and provenance use a bounded details view. Entity Explorer joins the same `fact_date` Game/Site Historical Fact for the historical name and never joins the current catalog.

Deploy Goose migrations for `gfg` and `gfn` before binaries. Then deploy the Collectors/Admin, inspect `metrics status`, run bounded dry-run/backfill, and leave normal runtime reconciliation enabled. Metric projection failures are retried without stopping collection. P0.3 adds no public Game/Nav analytics endpoint and no metric retention.

For the P0.5.1 repair, only Nav has a new migration. The migration retires the affected v1 Metric/Detector contracts and creates v2 checkpoints at the first full UTC day after migration. Pre-cutover observations do not reliably contain AAAA query outcome or security.txt recognition evidence, so production history is not destructively rewritten and v2 must not be backdated. Published v1 remains available for explicit rebuild under its original semantics.
