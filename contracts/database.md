# Database contract

- Goose under `db/{game,nav,admin}/migrations` is the sole schema source of truth. Applications must never run Goose or schema DDL at startup.
- `gfg`, `gfn`, and `gfa` are owned by `db/game`, `db/nav`, and `db/admin`. Admin supports PostgreSQL only.
- Production runtimes use pgx/v5 and bounded pgxpool connections. No global database singleton, ORM, generic repository, UnitOfWork, or SQL builder is permitted.
- Normal business SQL is static sqlc SQL. Generated packages are committed, regenerated with the pinned tool, and never manually edited.
- Direct pgx is allowed for pool lifecycle/health, explicit transaction control, and a documented query that sqlc cannot represent without harming correctness. It must not become an alternate query layer.
- Preserve existing transaction boundaries. Redis refresh occurs only after DB commit where that was the established behavior; Redis failure never rolls back committed DB data.
- Cross-database work is not atomic. In Admin, audit failure prevents the open `gfn`/`gfg` transaction from committing, but an already committed `gfa` audit is not presented as a distributed transaction.
- Destructive migrations require a final production-module reference audit, recorded row counts, verified backup, fresh and baseline-upgrade tests, no `CASCADE`, no masking `IF EXISTS`, and no fake reversible `Down`.
