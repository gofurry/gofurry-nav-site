# Game Collector deployment

Game collection keeps the existing V2 data pipeline and public cache contracts, but V3-P0.1.1 moves acquisition control to PostgreSQL.

## Deployment order

1. Stop the old Game Collector.
2. Back up `gfg` and record legacy collection-history row counts.
3. Apply `db/game` Goose migrations through the current version.
4. Deploy Admin, Game Collector, and Game Backend binaries built from the same revision.
5. Start Game Collector and verify its instance heartbeat and the two default schedules in Admin Collection Center.

Do not create compatibility views for retired `gfg_game_v2_*` physical names. Existing `/api/v2/game/*` and valid `game:v2:*` read-cache namespaces remain unchanged.

## Durable schedules

| Job key | Work | Schedule | Misfire | Lane |
|---|---|---|---|---|
| `game.metadata` | details + news | `0 3 * * *`, `Asia/Shanghai` | `catch_up_once` | `steam` |
| `game.players` | online player sample | anchored interval bootstrapped from `game_player_interval` | `skip` | `steam` |

PostgreSQL becomes the runtime schedule source after first bootstrap. `collect_players_on_startup` is accepted for configuration compatibility but ignored: restarting a collector must not enqueue work or shift interval phase. Use Admin Collection Center Run Now for an immediate sample.

Manual full/single jobs have priority 200. `entity_created` and AppID `entity_changed` metadata jobs have priority 300. All Steam-bound work shares one lane and is claimed through PostgreSQL with a renewable lease. An AppID change resets Steam-derived current/canonical state and enqueues a new entity job transactionally.

## State ownership

- PostgreSQL: Schedule, Job, Run, Task Result, collector instance/heartbeat, claim, lease, cancel, and history.
- Redis: expiring `collection:game:run:{run_id}:progress` plus existing public read caches.
- gocron: reconciliation/claim/recovery/heartbeat clock only.

The retired `game:v2:collect:pending`, `game:v2:collect:inflight:*`, and `game-collector:cmd:collect` keys are not read or written. Game Backend no longer exposes collection proxy routes.

## Retention

Destructive pruning of `gfg_game_player_counts` and durable Job/Run history is frozen until P0.2. Operational Task Result pruning remains enabled with a 30-day minimum and 90-day default/recommendation. Raw detail-snapshot retention remains unchanged.

## Verification

~~~powershell
go run . serve --config conf/server.yaml
go test ./...
go vet ./...
go build ./...
~~~

Verify in Admin Collection Center:

- both schedules have stable `next_scheduled_for` across collector restart;
- a single metadata job produces two Task Results;
- a players missed slot has Job `missed` and no Run;
- state-refresh retry creates another Run attempt on the same Job;
- cancel is cooperative and expired leases recover as `worker_lost`;
- final Run counters match Task Result coverage.
