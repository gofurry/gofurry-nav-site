# Game database (`gfg`)

This directory exclusively owns `gfg` schema migrations. The current-state
baseline contains the audited 22-table schema, six sequences, the `pg_trgm`
extension, the snapshot-pruning function, all constraints, indexes, defaults,
and schema comments.

The baseline deliberately preserves the audited pre-Goose state. Migration
`20260823000001` removes the four deprecated Game tables after all runtime
consumers moved to sqlc/pgx and the final production-module reference audit
returned zero active references.

The baseline has no destructive Down section.

Migration `20260916010000` widens only `gfg_game.info` and `info_en` from
`varchar(300)` to `varchar(400)`. The historical baseline and its adoption
snapshot remain unchanged; the final schema snapshot records 400. Rollback
refuses to narrow columns while values longer than 300 characters exist.

From the repository root on the deployment host, with the intended GFG migrator
URL in the private `.env`, apply only through this release's summary migration:

```bash
set -a
. ./.env
set +a
cd tools
GOOSE_DRIVER=postgres GOOSE_DBSTRING="$GOFURRY_GFG_MIGRATOR_URL" go tool goose -dir ../db/game/migrations status
GOOSE_DRIVER=postgres GOOSE_DBSTRING="$GOFURRY_GFG_MIGRATOR_URL" go tool goose -dir ../db/game/migrations up-to 20260916010000
```

Use `gofurry_migrator`; deploy the schema before the Admin binary that accepts
400-character summaries. Already-applied migrations are not executed again.
