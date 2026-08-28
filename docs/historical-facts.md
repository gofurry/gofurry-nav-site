# Historical facts operations

V3 P0.2 facts are owned by the existing Game and Nav Collectors. Acquisition reconciliation starts first; Fact reconciliation starts second, runs once at startup and then every configured 5–10 minutes, and processes at most one UTC day per pipeline pass.

Pipelines:

- Game: `game.player_facts`, `game.state_facts`
- Nav: `nav.target_facts`, `nav.site_facts`

Use the same explicit config lifecycle as every collector command:

```text
gf-game-collector facts status --config /path/server.yaml
gf-game-collector facts backfill --config /path/server.yaml --dry-run
gf-game-collector facts backfill --config /path/server.yaml --pipeline player --from 2026-08-01 --to 2026-08-07 --max-days 7
gf-game-collector facts backfill --config /path/server.yaml
gf-game-collector facts rebuild --config /path/server.yaml --pipeline player --from 2026-08-01 --to 2026-08-07

gf-nav-collector facts status --config /path/server.yaml
gf-nav-collector facts backfill --config /path/server.yaml --dry-run
gf-nav-collector facts rebuild --config /path/server.yaml --pipeline target --from 2026-08-01 --to 2026-08-07
```

Status includes the projection version, source/cutover/checkpoint dates, latest closed UTC day, and checkpoint lag. Backfill range flags are optional; without them it advances every pipeline from its checkpoint, while an explicit range cannot skip a preceding unprocessed day. `--max-days=0` is unlimited. All dates are UTC and `--to`/`--through` are inclusive CLI dates projected internally as `[day_start,day_end)`. Rebuild locks the pipeline, leaves `processed_through` unchanged, and fails—including in dry-run—if required source evidence has already been pruned.

`facts.retention_enabled` defaults to `false`. When enabled, Game uses the existing Player Raw age plus the `game.player_facts` checkpoint. Nav uses `facts.observation_keep_count` per normalized `(site_id,target,protocol)` plus the `nav.target_facts` checkpoint. Both delete one configured batch only after the projection transaction commits. Removing or clearing a checkpoint makes destructive pruning delete zero rows.

Deployment order is Goose up for `gfg` and `gfn`, sqlc-generated application deployment, Admin/Collector rollout, status/dry-run inspection, then normal Collector startup. Do not enable retention until backfill/checkpoints and row-count validation are complete.
