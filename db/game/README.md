# Game database (`gfg`)

This directory exclusively owns `gfg` schema migrations. The current-state
baseline contains the audited 22-table schema, six sequences, the `pg_trgm`
extension, the snapshot-pruning function, all constraints, indexes, defaults,
and schema comments.

The baseline deliberately includes the four deprecated Game tables. Their
removal is a later migration that is safe only after all runtime consumers have
moved off GORM and the final reference audit passes.

The baseline has no destructive Down section.
