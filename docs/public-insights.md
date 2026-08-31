# Public Insights contract

Public Insights is the stable product read layer over finalized P0 Historical Facts, Analytics Metrics, and Change Intelligence. Nav Backend owns Nav responses; Game Backend owns Game responses. The response body uses each backend's existing `{ "code", "data" }` envelope.

## Endpoints

Nav:

- `GET /api/v2/nav/insights/overview`
- `GET /api/v2/nav/insights/metrics/:metricKey/trend?range=30d|90d|all`
- `GET /api/v2/nav/sites/:siteId/insights`

Game:

- `GET /api/v2/game/insights/overview`
- `GET /api/v2/game/insights/metrics/:metricKey/trend?range=30d|90d|all`
- `GET /api/v2/game/games/:gameId/insights`
- `GET /api/v2/game/games/:gameId/insights/players?range=30d|90d|all`
- `GET /api/v2/game/games/:gameId/insights/prices?range=30d|90d|all`

Normal and empty results return 200. Invalid metric keys or ranges return 400, missing entities return 404, and genuine backend failures return 500.

## Metric contracts

Public keys are deliberately mapped to one reviewed internal version:

| Domain | Public key | Internal contract |
|---|---|---|
| Nav | `ipv6` | `ipv6_adoption/2` |
| Nav | `tls13` | `tls13_adoption/1` |
| Nav | `security_txt` | `security_txt_adoption/2` |
| Game | `free` | `free_game_share/1` |
| Game | `windows` | `windows_support/1` |
| Game | `linux` | `linux_support/1` |

Registry `active` state does not select a public version. Adding or activating an internal version does not change Public Insights until this mapping and its regression tests are deliberately updated.

Metric mathematics preserves the P0 contract:

```text
known    = positive_count + negative_count
value    = positive_count / known
coverage = known / eligible_count
```

A zero denominator produces `null`. `delta_30d` compares the current value with the same public contract exactly 30 calendar days earlier; it is `null` if that exact reliable date or either value is unavailable. Each metric carries its own `as_of` and `available_from`.

Trends accept only `30d`, `90d`, and `all`. Points are real reliable rows for the mapped version. Missing days are not zero-filled, values are not copied backward, and versions are not stitched together. Empty history uses null availability fields and `points: []`.

## Entity states

| Internal state | Public state |
|---|---|
| `positive` | `supported` |
| `negative` | `unsupported` |
| `stale` | `stale` |
| `not_probed` | `not_probed` |
| `probe_failed` | `unavailable` |
| `unknown` | `unknown` |
| `not_applicable` | `not_applicable` |

`unknown`, `unavailable`, `not_probed`, and `stale` never mean `unsupported`. A Site capability and its ecosystem comparison always come from the same `fact_date`.

## Public changes

Public DTOs expose only stable `type`, date/time precision, entity identity, and deliberately typed public detail. They never expose internal Metric or Detector keys/versions, event/source keys, source contracts/versions, or raw `old_value`/`new_value`. P1-A returns `detail: null`. Day-precision events retain `occurred_at: null`; midnight is not fabricated.

Nav public types cover IPv6, TLS 1.3, security.txt, Primary Target, and TLS certificate changes. Only IPv6, TLS 1.3, and security.txt appear in the overview. Primary Target and TLS certificate remain entity-timeline-only.

Game public types cover free/paid, Windows, Linux, release availability/plan, CN price, and discount changes. Price events must have region scope `CN`.

Overview feeds apply the public whitelist, keep the newest eligible event per entity, order newest first, and return at most eight. Entity timelines do not deduplicate by entity and return at most twenty.

## Game facts

Player history reads finalized daily facts and emits only days with at least one successful sample and a valid player value. A genuine collected zero is returned as `0`; failed acquisition or no valid sample is omitted and never converted to zero. The summary's current and 30-day peak fields are nullable independently.

Price summary and history use finalized `CN` facts only. There is no region parameter or fallback. `free` has null amounts and remains distinct from `priced` with `final_amount: 0`; `unpriced` and `unknown` also keep amounts null.

## P1 boundary

P1-A adds no Nuxt Insights page, chart, ranking, compare, deep dimensions, regional price intelligence, historical-low calculation, language/developer/publisher analytics, Change Explorer, data export, or composite score. It adds no Insights schema, registry, materialized view, cache, projector, worker, or Analytics microservice.
