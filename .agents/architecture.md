# Runtime architecture

| Runtime | PostgreSQL | Other state |
|---|---|---|
| Game Collector | `gfg` | existing Redis keys; refresh only after DB commit |
| Game Backend | `gfg` | existing Redis cache and scheduling semantics |
| Nav Collector | `gfn` | existing Redis keys and unchanged scheduler/timewheel behavior |
| Nav Backend | `gfn` | reachable Nav V1 and V2 remain supported |
| Admin | explicit `gfa`, `gfn`, `gfg` pools | `gfa` auth/audit; no cross-DB transaction assumption |

`db/game`, `db/nav`, and `db/admin` exclusively own their schemas through Goose. Applications open bounded pgxpool connections and never execute migrations. Root `sqlc.yaml` generates a service-local package for every database consumer; normal business SQL is static sqlc SQL. Direct pgx is reserved for transactions, health checks, and documented cases sqlc cannot express cleanly.

Admin writes business data in `gfn` or `gfg` while audit rows are independently written to `gfa`. An audit failure rolls back the still-open business transaction; there is no distributed commit claim.
