# Repository and runtime architecture

## Active topology

Production code is limited to six China-site applications under `apps/cn`:

| Application | Responsibility | Stateful dependencies |
|---|---|---|
| `nav-backend` | Nav HTTP API; existing V1 and V2 behavior remains available | `gfn` PostgreSQL, existing Redis keys |
| `nav-collector` | Nav observation and probe collection | `gfn` PostgreSQL, existing Redis keys |
| `game-backend` | Game HTTP API | `gfg` PostgreSQL, existing Redis keys |
| `game-collector` | Steam/game collection and operator one-shot commands | `gfg` PostgreSQL, existing Redis keys |
| `admin` | Embedded Admin UI and management API | explicit `gfa`, `gfn`, and `gfg` PostgreSQL pools plus Redis |
| `nav-web` | Production Nuxt frontend | Nav and Game APIs |

`apps/intl` contains placeholders only. It has no build, CI, deployment, database, or runtime ownership in this release.

## Database ownership

The root directories `db/game`, `db/nav`, and `db/admin` own the `gfg`, `gfn`, and `gfa` schemas. Goose is the only schema migration source of truth. Applications never execute migrations at startup.

Normal business SQL is declared in service-local sqlc query files and generated from the root `sqlc.yaml` contract. Generated sqlc code is committed and is never edited manually. Production PostgreSQL access uses pgx/v5 and pgxpool.

Admin writes business data through explicit `gfn` or `gfg` pools and records authentication/audit state in `gfa`. No distributed transaction across the three databases is implied.

## Process lifecycle

Each active Go binary has an explicit Cobra CLI:

- the root command displays help and never starts a service implicitly;
- `serve --config <absolute-or-relative-file>` runs in the foreground;
- `version` prints version information without loading runtime dependencies;
- `install --config <file> [--force]` is a Linux/systemd deployment helper;
- `uninstall` removes only systemd registration;
- Game Collector retains `collect`/`full`, `players`, and `all` one-shot commands.

Configuration is loaded through a per-app `viper.New()` instance from an explicit file and decoded into the existing typed configuration. Arbitrary environment-variable binding is intentionally disabled.

## Archive boundary

`legacy`, `experimental`, and `third-party` are outside the active production graph. Default build, CI, vulnerability scanning, sqlc generation, deployment tooling, and active application dependencies must not enter those trees. Reviving archived code requires a separate scoped migration.
