# Production rollout (operator only)

Codex/local automation must not connect to production or execute these steps.

## Initial Goose adoption and pgx/sqlc rollout

1. Freeze schema changes and take verified full backups of `gfg`, `gfn`, and `gfa`.
2. Capture fresh `pg_dump --schema-only` output for all three databases.
3. Build each audited baseline in an empty scratch PostgreSQL 18 database and require normalized structural diff = zero.
4. Run the baseline-adoption utility for `gfg`, `gfn`, and `gfa`; it must reject any drift and must not modify business tables.
5. Verify Goose status is baseline `20260823000000`. The Game/Nav cleanup migration may remain pending at this point; do not run full `goose up` yet.
6. Deploy the five pgx/sqlc binaries in the planned service order, then deploy Admin and frontends as applicable.
7. Run auth, Nav V1/V2, Game, collector, Redis/cache, health, and Admin CRUD smoke tests. Observe errors, pool saturation, and collector outcomes before proceeding.

## Later deprecated-table cleanup

1. After all migrated binaries are stable, repeat the runtime/reference audit for the four deprecated `gfg` tables and `gfn_log_update`.
2. Record each table row count and retain an export that has been restore-tested. Verify the full database backup again.
3. Apply the explicit Game/Nav Goose cleanup migration. Do not add `CASCADE`, `IF EXISTS`, or an invented destructive `Down`.
4. Verify Goose status `20260823000001`, confirm only the five deprecated tables were removed, and rerun production smoke tests.
5. Prefer roll-forward. Use the verified backup restore procedure when destructive recovery is genuinely required.
