# Database migrations

Goose is the only schema migration source of truth.

| Database | Migration owner | Runtime consumers |
| --- | --- | --- |
| `gfg` | `db/game` | Game Collector, Game Backend, Admin |
| `gfn` | `db/nav` | Nav Collector, Nav Backend, Admin |
| `gfa` | `db/admin` | Admin |

Runtime services must never execute Goose. Run migrations as a serialized
deployment or operator action. Query SQL remains service-local and does not own
DDL.

From `tools/`, use the pinned commands:

```text
go tool goose -dir ../db/game/migrations postgres "$GOFURRY_DATABASE_URL" status
go tool goose -dir ../db/game/migrations postgres "$GOFURRY_DATABASE_URL" up
```

Substitute `nav` or `admin` for the other databases. Never run a baseline on an
existing pre-Goose database; follow the adoption runbook and use
`tools/db-baseline` only after an exact structural match.
