# Public Insights contract

Public Insights is the stable product read layer over finalized P0 Historical Facts, Analytics Metrics, and Change Intelligence. Nav Backend owns Nav responses; Game Backend owns Game responses. The response body uses each backend's existing `{ "code", "data" }` envelope.

## Endpoints

Nav:

- `GET /api/v2/nav/insights/overview`
- `GET /api/v2/nav/insights/metrics/:metricKey/trend?range=30d|90d|all`
- `GET /api/v2/nav/insights/metrics/:metricKey/breakdown?dimension=country|group|nsfw|public_interest`
- `GET /api/v2/nav/insights/metrics/:metricKey/breakdown/:dimension/:value/trend?range=30d|90d|all`
- `GET /api/v2/nav/insights/changes?range=7d|30d|90d|all&category=&type=&cursor=&limit=`
- `GET /api/v2/nav/sites/:siteId/insights`

Game:

- `GET /api/v2/game/insights/overview`
- `GET /api/v2/game/insights/metrics/:metricKey/trend?range=30d|90d|all`
- `GET /api/v2/game/insights/metrics/:metricKey/breakdown?dimension=primary_tag|tag`
- `GET /api/v2/game/insights/metrics/:metricKey/breakdown/:dimension/:value/trend?range=30d|90d|all`
- `GET /api/v2/game/insights/changes?range=7d|30d|90d|all&category=&type=&cursor=&limit=`
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

## Dimension contract

Public dimension names are an anti-corruption mapping; internal keys never appear in a Public Insights response or URL.

| Domain | Public dimension | Internal dimension | Slice mode |
|---|---|---|---|
| Nav | `country` | `site_country` | `partition` |
| Nav | `group` | `group_id` | `overlapping` |
| Nav | `nsfw` | `nsfw` | `partition` |
| Nav | `public_interest` | `welfare` | `partition` |
| Game | `primary_tag` | `primary_tag_id` | `partition` |
| Game | `tag` | `tag_id` | `overlapping` |

`overlapping` means one entity may appear in multiple slices. Overlapping slice populations must never be added to derive the global population. Breakdown responses state `slice_mode` explicitly.

Every breakdown first resolves the mapped metric version's newest `global/all` row and freezes its `fact_date` as `as_of`. It then reads all requested dimension rows on exactly that date; individual slices never choose their own latest date. Each slice returns `population`, `eligible`, `known`, `metric_value`, and `coverage`, using the same public mathematics as the global metric. Zero denominators remain `null`, small slices are not hidden, and the `unknown` slice is retained.

Country identity is its stored country code and has no backend dictionary. Public boolean identities are:

| Public dimension | Internal value | Public value |
|---|---|---|
| `nsfw` | `true` | `nsfw` |
| `nsfw` | `false` | `sfw` |
| `public_interest` | `true` | `public_interest` |
| `public_interest` | `false` | `standard` |

Internal `unknown` remains public `unknown`. Tag and Group identity is the stable numeric ID; labels are current display metadata loaded with a `LEFT JOIN`. Missing metadata does not remove a historical slice and falls back to `标签 #ID` / `Tag #ID` or `分组 #ID` / `Group #ID`.

Selected-slice trends use the current global metric horizon as their range anchor, not the slice's newest row. `as_of` is that global horizon. `available_from` and `available_through` separately describe the selected slice's actual history. A valid slice with no rows returns 200 with `points: []`; missing dates remain absent and no retired/current versions are combined.

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

## Change Explorer contract

The domain-specific Explorer endpoints expose the complete approved public stream without overview entity deduplication. `/insights/changes` selects exactly one domain (`site` or `game`) in the UI; the P1 overview remains the only cross-domain recent feed.

Site categories are:

- `capability`: IPv6, TLS 1.3, and security.txt public types;
- `target`: `site.primary_target.changed`;
- `certificate`: `site.tls_certificate.changed`.

Game categories are:

- `pricing_model`: free/paid transitions;
- `platform`: Windows and Linux added/removed transitions;
- `release`: availability, withdrawal, and release-plan changes;
- `price`: CN price increase/decrease/state/currency changes;
- `discount`: CN discount start/end/change.

There is no Mac category. Game `price` and `discount` retain the SQL-level `region/CN` scope guard. The API accepts exact public `type` filtering even though the first UI exposes category filtering only.

Ranges are evaluated by `projection_date`, never `event_at`. Defaults are `range=30d` and `limit=20`; the maximum limit is 50. Items expose only `domain`, `category`, public `type`, entity, `date`, precision-preserving `occurred_at`, and `detail`. P2.1 keeps `detail: null`. Day precision remains `occurred_at: null`.

Pagination is keyset-only. The stable descending position is:

```text
projection_date
precision_rank (exact before day)
event_sort_at
md5(event_key) opaque tie
```

The DAO requests `limit + 1`; the Service trims and creates `next_cursor`. No offset, page number, `COUNT(*)`, or `total_count` is used. The base64url-encoded JSON cursor contains a version, bound range/category/type, frozen `range_through`, and the opaque position. A changed filter or malformed cursor returns 400. Raw event, detector, and source keys are never placed in the cursor or public DTO.

## Game facts

Player history reads finalized daily facts and emits only days with at least one successful sample and a valid player value. A genuine collected zero is returned as `0`; failed acquisition or no valid sample is omitted and never converted to zero. The summary's current and 30-day peak fields are nullable independently.

Price summary and history use finalized `CN` facts only. There is no region parameter or fallback. `free` has null amounts and remains distinct from `priced` with `final_amount: 0`; `unpriced` and `unknown` also keep amounts null.

## P2.1 boundary

P2.1 adds no schema, migration, registry, table, view, materialized view, cache, projector, worker, Analytics service, query builder, or ORM. It does not add Mac support, regional price intelligence, observed historical low, player rankings, language analytics, expanded Site capabilities, compare, maps, treemaps, heatmaps, or multi-series comparison.

The Nuxt product keeps global metric/range state in `useInsightsDomain` and dimension/slice state in the independent `useInsightsDimensions`. Breakdown and 30d/90d selected-slice data are SSR-loaded; `all` selected-slice history is client-deferred. Change Explorer first pages are SSR-loaded, later pages are client-loaded, and cursors never enter the URL.
