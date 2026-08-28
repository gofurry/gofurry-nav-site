# Database contract

- Goose under `db/{game,nav,admin}/migrations` is the sole schema source of truth. Applications must never run Goose or schema DDL at startup.
- `gfg`, `gfn`, and `gfa` are owned by `db/game`, `db/nav`, and `db/admin`. Admin supports PostgreSQL only.
- Production runtimes use pgx/v5 and bounded pgxpool connections. No global database singleton, ORM, generic repository, UnitOfWork, or SQL builder is permitted.
- Normal business SQL is static sqlc SQL. Generated packages are committed, regenerated with the pinned tool, and never manually edited.
- Direct pgx is allowed for pool lifecycle/health, explicit transaction control, and a documented query that sqlc cannot represent without harming correctness. It must not become an alternate query layer.
- Preserve existing transaction boundaries. Redis refresh occurs only after DB commit where that was the established behavior; Redis failure never rolls back committed DB data.
- Cross-database work is not atomic. In Admin, audit failure prevents the open `gfn`/`gfg` transaction from committing, but an already committed `gfa` audit is not presented as a distributed transaction.
- Destructive migrations require a final production-module reference audit, recorded row counts, verified backup, fresh and baseline-upgrade tests, no `CASCADE`, no masking `IF EXISTS`, and no fake reversible `Down`.
- Active Game storage objects use unversioned physical names. Long-lived compatibility views for retired `gfg_game_v2_*` names are not permitted.
- `gfg_collection_*` and `gfn_collection_*` are the durable collector control-plane contracts. Scheduled slots are idempotent, active dedupe is enforced, and a concurrency lane has at most one running Job.
- Historical Fact pipelines use one PostgreSQL checkpoint row per pipeline and `SELECT ... FOR UPDATE` as the ordered singleton. The checkpoint advances in the same transaction as projection and validation; rebuild never moves it backward.
- Destructive raw retention is disabled by default and runs only in a separate batched transaction after checkpoint commit. Game Player Raw is gated by `game.player_facts` plus configured age. Nav Observation Raw is gated by `nav.target_facts` and preserves `keep_count` independently per `(site_id,target,protocol)`. Missing checkpoints always delete zero rows. Historical Facts have no automatic retention.
