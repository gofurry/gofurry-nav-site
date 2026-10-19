# Repository and runtime architecture

## Active topology

Production code is limited to five Go services plus one Nuxt frontend under `apps/cn`:

| Application | Responsibility | Stateful dependencies |
|---|---|---|
| `nav-backend` | Nav HTTP API; existing V1 and V2 behavior remains available | `gfn` PostgreSQL, existing Redis keys |
| `nav-collector` | Nav observation and probe collection | `gfn` PostgreSQL, existing Redis keys |
| `game-backend` | Game HTTP API | `gfg` PostgreSQL, existing Redis keys |
| `game-collector` | Steam/game collection and operator one-shot commands | `gfg` PostgreSQL, existing Redis keys |
| `admin` | Embedded Admin UI and management API | explicit `gfa`, `gfn`, and `gfg` PostgreSQL pools plus Redis |
| `web` | Production Nuxt frontend (Nav Web) | Nav and Game APIs |

`apps/intl` contains placeholders only. It has no build, CI, deployment, database, or runtime ownership.

## Database ownership

The root directories `db/game`, `db/nav`, and `db/admin` own the `gfg`, `gfn`, and `gfa` schemas. Goose is the only schema migration source of truth. Applications never execute migrations at startup.

Normal business SQL is declared in service-local sqlc query files and generated from the root `sqlc.yaml` contract. Generated sqlc code is committed and is never edited manually. Production PostgreSQL access uses pgx/v5 and pgxpool.

Admin writes business data through explicit `gfn` or `gfg` pools and records authentication/audit state in `gfa`. No distributed transaction across the three databases is implied.

Availability is owned by the independent [LibreCosta Uptime](https://github.com/librecosta/uptime) project, outside this repository's Goose/sqlc and runtime graph. The old implementation is archived in `legacy/gf-uptime`.

## Availability boundaries

Planned maintenance keeps `status.go-furry.com` proxying LibreCosta Uptime at `127.0.0.1:18580` while business hosts serve the maintenance page.

- Nav Web exposes dependency-free `GET /healthz`.
- Nav/Game Backend and Admin retain their existing health endpoints.
- Nav/Game Collector expose optional internal `net/http` `/livez` and `/readyz` listeners only during scheduled `serve`.
- Collector readiness covers only local PostgreSQL, required Redis, scheduler initialization, and shutdown state.
- LibreCosta Uptime v0.1.0 owns the public status UI via `librecosta-uptime.service`; the old `gf-uptime.service` is stopped and disabled. Business services retain their health endpoints and never depend on the observer.

## Managed Asset Delivery

Admin publishes to COS Primary, then an R2 best-effort Mirror; a Mirror failure
does not block Primary publication. GFN stores provider-neutral object keys.
Nav Web resolves these through the EdgeOne Primary or Cloudflare Mirror public
origin using an SSR-readable preference cookie, browser background probes, and
failure fallback. CloudOps provides scoped object inspection, COS-to-R2 repair,
and CDN purges with capability checks and audited mutations.

The initial production cutover completed on 2026-09-14. See [Managed assets](managed-assets.md)
and the [asset contract](../contracts/assets.md) for the durable boundaries.

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
