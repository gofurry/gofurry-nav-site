# Change Intelligence operations

V3-P0.4 materializes canonical semantic changes after Historical Facts and Analytics Metrics. It does not expose a public change feed and does not replace the existing Redis-backed Nav Change route.

## Data flow and ownership

The existing collectors run `Control -> Facts -> Metrics -> Changes`. Goose owns `gfg_change_registry`, `gfg_change_events`, `gfg_change_checkpoints` and their `gfn_*` equivalents. Runtime and Admin never mutate Registry rows.

Game detectors are `free_game_transition/1`, `windows_support_transition/1`, `mac_support_transition/1`, `linux_support_transition/1`, `game_release_transition/1`, and `game_price_transition/1`. Nav keeps retired `ipv6_transition/1` and `security_txt_transition/1` compiled for historical rebuild, and runs active `ipv6_transition/2`, `security_txt_transition/2`, `tls13_transition/1`, `http2_transition/1`, `hsts_transition/1`, `csp_transition/1`, `tls_certificate_verification_transition/1`, `primary_target_transition/1`, and `tls_certificate_transition/1`. Each Metric detector consumes only its explicitly versioned source contract, so a semantic transition never mixes Metric versions.

Metric detectors read Metric Entity Daily plus same-day Historical Facts. Price and certificate detectors read finalized Facts. Release reads canonical Release History with reliable tracking-period attribution. Primary reads effective-dated Primary Target periods. Raw observations, collection Jobs/Runs/Results, Redis, current catalogs, and legacy Nav Change are excluded.

HTTP/2, HSTS, CSP, and certificate-verification Changes emit only positive/negative transitions within the same Primary Target tracking identity. Unknown, stale, unavailable, initial, and Primary Target switch states never create an event. Fingerprint-only certificate replacement remains separate from verification failed/restored and is not interpreted as renewal success.

## Configuration and runtime

Both Collector configs accept:

```yaml
changes:
  reconcile_interval_seconds: 600
```

Omission keeps the 600-second default. Each pass processes at most one UTC day per active detector. Registry/catalog drift is a startup error; an individual historical projection error is logged and retried without stopping other Collector stages or detector checkpoints.

## CLI

Run against the matching Collector config:

```text
gf-game-collector changes status --config /etc/gf-game-collector/server.yaml
gf-game-collector changes backfill --config /etc/gf-game-collector/server.yaml --dry-run --max-days 10
gf-game-collector changes backfill --config /etc/gf-game-collector/server.yaml --detector free_game_transition --version 1 --through YYYY-MM-DD
gf-game-collector changes rebuild --config /etc/gf-game-collector/server.yaml --detector free_game_transition --version 1 --from YYYY-MM-DD --dry-run
```

Nav uses the same flags on `gf-nav-collector`. Backfill is ordered and cannot skip the next checkpoint day. Rebuild is intentionally forward-propagating: `--from D` recomputes through the locked `processed_through`. An optional `--through` must equal that checkpoint, and `--max-days` may not truncate the required range. Rebuild never moves the checkpoint.

For P2.3 rollout, finish the four new Site capability Metric backfills before starting `http2_transition/1`, `hsts_transition/1`, `csp_transition/1`, or `tls_certificate_verification_transition/1` Change backfills.

## Verification

After migrations and upstream backfill:

1. Check `changes status`; `upstream_processed_through` must cover the next detector day.
2. Run `changes backfill --dry-run`, then a bounded actual backfill.
3. Inspect Admin Change Center Registry, Checkpoints, and Events.
4. Confirm observed events have `event_at = source_after_at`, effective Primary events use the new period start, and day-basis events have no `event_at`.
5. Re-run a bounded `changes rebuild --from ...`; event keys and event count must remain stable while `materialized_at` refreshes.

An `upstream watermark is not ready` result is expected until the relevant Metric checkpoint, Fact checkpoint, or closed UTC day is available. There is no Change retention in P0.4.
