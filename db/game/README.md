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

Migration `20260916020000` subsequently replaces current Tag storage with
`gfg_tag_category`, `gfg_tag` and `gfg_game_tag`, retaining reviewed leaf IDs and
deriving primary/secondary compatibility fields from relation roles. It removes
the legacy map/prefix/physical role columns, clears old recommendation rows and
advances current projection semantics without rewriting historical facts.
Its Down rejects rollback; recovery requires a verified backup and matching runtimes.

The alpha.9 target is `20260916020000`, not the earlier summary-only
`20260916010000`. Follow the [alpha.9 upgrade guide](../../docs/releases/v3.0.0-alpha.9.md)
for status checks, coordinated shutdown, migration and recommendation rebuild.
Use `gofurry_migrator` only for Goose; applications keep `gofurry_app`.
Already-applied migrations are not executed again. Never rerun baseline adoption
or edit Goose history to replay an applied migration.
