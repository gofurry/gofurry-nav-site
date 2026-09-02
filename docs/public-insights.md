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
- `GET /api/v2/game/insights/players/ranking?metric=latest_observed|peak_30d|average_30d&limit=20`
- `GET /api/v2/game/insights/prices/overview?region=CN|US|HK`
- `GET /api/v2/game/insights/prices/discounts?region=CN|US|HK&limit=20`
- `GET /api/v2/game/insights/languages/overview`
- `GET /api/v2/game/games/:gameId/insights`
- `GET /api/v2/game/games/:gameId/insights/players?range=30d|90d|all`
- `GET /api/v2/game/games/:gameId/insights/prices?region=CN|US|HK&range=30d|90d|all`

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
| Game | `mac` | `mac_support/1` |
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

Game public types cover free/paid, Windows, macOS, Linux, release availability/plan, CN price, and discount changes. Mac changes map `mac_support_transition/1` to `game.mac.added` and `game.mac.removed`. Price events must have region scope `CN`.

Overview feeds apply the public whitelist, keep the newest eligible event per entity, order newest first, and return at most eight. Entity timelines do not deduplicate by entity and return at most twenty.

## Change Explorer contract

The domain-specific Explorer endpoints expose the complete approved public stream without overview entity deduplication. `/insights/changes` selects exactly one domain (`site` or `game`) in the UI; the P1 overview remains the only cross-domain recent feed.

Site categories are:

- `capability`: IPv6, TLS 1.3, and security.txt public types;
- `target`: `site.primary_target.changed`;
- `certificate`: `site.tls_certificate.changed`.

Game categories are:

- `pricing_model`: free/paid transitions;
- `platform`: Windows, macOS, and Linux added/removed transitions;
- `release`: availability, withdrawal, and release-plan changes;
- `price`: CN price increase/decrease/state/currency changes;
- `discount`: CN discount start/end/change.

Game `price` and `discount` retain the SQL-level `region/CN` scope guard; P2.2 regional price intelligence does not expand Change Explorer beyond CN. The API accepts exact public `type` filtering even though the first UI exposes category filtering only.

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

Player history reads finalized daily facts and emits only days with at least one successful sample and a valid player value. A genuine collected zero is returned as `0`; failed acquisition or no valid sample is omitted and never converted to zero. Entity `players.current` means the latest successful GoFurry observation in the current tracking period, including a successful manual observation; it is not realtime Steam concurrency.

Player rankings accept `latest_observed`, `peak_30d`, and `average_30d`, default to `latest_observed`, and return at most 100 items. Latest ranking uses the newest successful or partial scheduled all-Game Job slot only; manual and unusable slots never enter its cohort. `latest_slot_scheduled_for` separately discloses a newer failed, skipped, missed, or canceled terminal slot. Population is the explicit tracking cohort at the usable slot and real zero remains rankable.

The 30-day rankings and entity summary use the common `game.player_facts.processed_through` horizon and the current explicit tracking period only. `peak_30d` is the maximum daily peak. `average_30d` is weighted by `successful_samples`, not an average of daily averages. Player quality exposes eligible start, observed days, successful samples, and sample coverage. Sample coverage is `null` unless every applicable daily fact has a reliable expected-sample denominator.

Regional price intelligence supports exactly `CN`, `US`, and `HK`; omitted `region` remains compatible with CN. Entity regional summaries use one common finalized Game State Fact `as_of`, and each region is read on exactly that day. A missing regional fact is `available: false`, while an explicit `unknown` fact is `available: true, state: unknown`; there is no region fallback.

Price states remain independent: `free`, `priced`, `unpriced`, and `unknown`. Free has null monetary amounts. `priced` with `final_amount: 0` remains priced and may carry a 100% discount. History never zero-fills, forward-fills, interpolates, joins currencies, or crosses the current tracking period. The UI breaks chart lines across missing dates, non-monetary states, and currency changes.

`GoFurry Observed Low` / `GoFurry 观测最低价` is the minimum finalized `final_amount` within the current continuous priced, same-currency identity for the selected region and tracking period. Tracking changes, currency changes, free/unpriced/unknown states, and missing finalized calendar days break the identity. Only priced facts participate, including priced zero; free never participates. `first_seen` is the first GoFurry finalized daily fact at that minimum, and makes no claim about Steam's first discount time. A current non-priced or unavailable region has no observed low.

Price overview uses the common finalized State Fact cohort and preserves `population = priced + free + unpriced + unknown + unavailable`; `known = priced + free + unpriced` and zero-denominator coverage is null. Current discounts contain only priced facts with a positive discount, include priced zero at 100%, exclude free, and order by discount percentage descending then Game ID ascending. No endpoint performs FX conversion, cheapest-region judgment, or cross-currency monetary comparison.

## Mac contract

Mac is a peer of Windows and Linux. `gfg_game_daily.mac` is projected through the version-frozen `mac_support/1` Metric with `tracked_game_v1`, 259200-second freshness, and global/primary-tag/tag dimensions. Entity state exposes nullable `state.mac`. The `mac_support_transition/1` detector emits public platform changes only for adjacent reliable `unsupported -> supported` or `supported -> unsupported` states within one tracking period. Unknown evidence and tracking boundaries never create a transition.

## Language analytics

Language overview uses one common finalized Game State Fact snapshot. Evidence is `fresh`, `stale`, or `unobserved` with `freshness_seconds=259200`; coverage is fresh divided by population. Supported-language and Explicit Full-Audio shares use fresh as their denominator and are overlapping prevalence, so shares must not be summed to 100%.

Canonical language codes provide stable identity. Unknown names are never fuzzy-mapped or merged into a fake language; they appear only in `fully_normalized_games`, `unmapped_games`, `unmapped_entries`, and `normalization_coverage`. Full Audio is positive-only evidence: absence of an explicit marker does not mean unsupported. Nullable Interface and Subtitle fields likewise remain unknown rather than false.

## P2.2 product and boundary

The Game Intelligence product routes are `/insights/games`, `/insights/games/players`, `/insights/games/prices`, and `/insights/games/languages`. Player ranking, regional overview/discounts, and language snapshot are SSR-loaded. Game Detail summary is SSR-loaded; player and region-plus-range price histories remain lazy and page-lifecycle cached.

Player, State Fact, Metric, and Change horizons remain independently disclosed; P2.2 does not fabricate a global cross-product `as_of`. Other than the Mac Metric and Change contracts/projectors, P2.2 adds no leaderboard, observed-low, or language aggregate table, materialized view, cache, Analytics service, ORM, or query builder. Mac history is backfilled explicitly outside migrations in Metric-before-Change order. Site capability expansion, Compare, regional Change Explorer, FX, and language trends remain deferred.
