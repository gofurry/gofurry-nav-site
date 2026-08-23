# Baseline adoption utility

This one-time utility marks an existing, exact pre-Goose database as having the
audited baseline. It does not execute business DDL or attempt repair.

It rejects the operation unless:

- the selected contract is `gfg`, `gfn`, or `gfa`;
- version `20260823000000` is explicitly selected;
- the connected database name matches the selected contract;
- the standard Goose version table does not exist;
- tables, columns, types, nullability, defaults, sequences, constraints,
  indexes, extensions, functions, triggers, and relevant comments exactly match
  the embedded audited snapshot.

Only after all checks pass does it use Goose's public PostgreSQL version-store
API to record version `0` and the baseline version in one transaction.

Operator usage from `tools/`:

```text
GOFURRY_DATABASE_URL='postgres://...' go run ./db-baseline \
  -database gfg \
  -baseline-version 20260823000000 \
  -confirm-adopt
```

Do not use this command for an empty database. Empty databases use normal
`goose up`. Do not run it until the production schema-only dump has been
structurally compared with a scratch database built from the baseline.
