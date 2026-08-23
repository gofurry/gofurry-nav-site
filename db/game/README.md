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
