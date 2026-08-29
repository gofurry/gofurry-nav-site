# Analytics metric operations

V3 P0.3 metrics are deterministic projections of finalized P0.2 Historical Facts. The existing Game and Nav Collectors own the Metric Engines; no new service, acquisition Job, Run, Task, or queue is created.

The runtime order is:

```text
Acquisition Control -> Historical Facts -> Analytics Metrics
```

Each active metric version reconciles once at startup and every `metrics.reconcile_interval_seconds` (default `600`). One reconcile processes at most one UTC day per metric. Game waits for `game.state_facts`; Nav waits for both `nav.target_facts` and `nav.site_facts`.

## Registry and state contract

Goose owns the Registry. Runtime and Admin read it but do not mutate it. Alpha.4 contains exactly these evaluator contracts:

| Domain | Metric/version | Freshness | Dimensions |
|---|---|---:|---|
| Game | `free_game_share/1` | 72h | `primary_tag_id`, `tag_id` |
| Game | `windows_support/1` | 72h | `primary_tag_id`, `tag_id` |
| Game | `linux_support/1` | 72h | `primary_tag_id`, `tag_id` |
| Nav | `ipv6_adoption/1` | 72h | `site_country`, `group_id`, `nsfw`, `welfare` |
| Nav | `tls13_adoption/1` | 48h | `site_country`, `group_id`, `nsfw`, `welfare` |
| Nav | `security_txt_adoption/1` | 21d | `site_country`, `group_id`, `nsfw`, `welfare` |

Entity state is one of `positive`, `negative`, `stale`, `not_probed`, `probe_failed`, `unknown`, or `not_applicable`. Every row records a reason and provenance. Freshness is relative to `fact_date + 1 day 00:00:00Z`; evidence after that boundary aborts the transaction.

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
gf-nav-collector metrics backfill --config /path/server.yaml --metric ipv6_adoption --version 1 --dry-run
gf-nav-collector metrics rebuild --config /path/server.yaml --metric ipv6_adoption --version 1 --from 2026-08-01 --through 2026-08-07
```

`--to` is an alias for inclusive `--through`; `--max-days=0` is unlimited. Backfill cannot skip the ordered checkpoint's next day. Rebuild locks the metric version, requires the range to have been processed already, leaves `processed_through` unchanged, and supports retired versions when explicitly selected.

## Admin and rollout

The authenticated Admin Metric Center is read-only and exposes Overview, Registry, Checkpoints, Daily Results, and Entity Explorer. Adoption/coverage are calculated by the Admin query layer. Entity Explorer joins the same `fact_date` Game/Site Historical Fact for the historical name and never joins the current catalog.

Deploy Goose migrations for `gfg` and `gfn` before binaries. Then deploy the Collectors/Admin, inspect `metrics status`, run bounded dry-run/backfill, and leave normal runtime reconciliation enabled. Metric projection failures are retried without stopping collection. P0.3 adds no public Game/Nav analytics endpoint and no metric retention.
