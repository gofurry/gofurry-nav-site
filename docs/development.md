# Local development

## Prerequisites

- Go 1.26.7 for all six active Go modules and `tools`
- Node.js 24 and npm for `apps/cn/nav-web` and `apps/cn/admin/react`
- development PostgreSQL and Redis matching the selected app configuration
- sqlc and govulncheck through the pinned Go tools in `tools/go.mod`

The repository intentionally has no `go.work` and retains npm rather than pnpm. Run Go commands from the owning module.

## Configuration

Every runtime command requires an explicit YAML file:

| App | Example config |
|---|---|
| Nav Backend | `apps/cn/nav-backend/conf/server.example.yaml` |
| Nav Collector | `apps/cn/nav-collector/conf/server.example.yaml` |
| Game Backend | `apps/cn/game-backend/conf/server.example.yaml` |
| Game Collector | `apps/cn/game-collector/conf/server.example.yaml` |
| Admin | `apps/cn/admin/config/server.example.yaml` |
| Uptime | `apps/cn/uptime/conf/server.example.yaml` |

Copy the example to an ignored local file, replace placeholders, and never commit credentials.

Example:

~~~bash
cd apps/cn/nav-backend
go run . serve --config conf/server.yaml
~~~

The root command is intentionally non-starting:

~~~bash
go run . --help
go run . version
~~~

Game Collector one-shot operations use the same explicit configuration:

~~~bash
cd apps/cn/game-collector
go run . collect --config conf/server.yaml
go run . players --config conf/server.yaml
go run . all --config conf/server.yaml
~~~

## Shared Development Infrastructure

The shared development server provides infrastructure only. Developers run Go services, collectors, Nuxt, React, tests, and repository tools on their own workstations, then connect through Tailscale to shared PostgreSQL on port `5432` and Redis on port `6379`. Do not use the server as a remote workstation or clone the repository there merely to run applications or Goose.

The ignored root `.env` is the private local credential source for Codex, Goose, and generation or maintenance of ignored runtime YAML. Start from `.env.example` and keep real hostnames, addresses, passwords, and DSNs out of Git. Applications do not load the root `.env`; every runtime continues to use an explicit `serve --config <file>` YAML file.

When an ignored local `server.yaml` already exists, update only its PostgreSQL and Redis fields and preserve all other local choices. If it does not exist, copy the corresponding `server.example.yaml`. The mappings are:

| Runtime | PostgreSQL database | PostgreSQL role |
|---|---|---|
| Nav Backend, Nav Collector | `gfn` | `gofurry_app` |
| Game Backend, Game Collector | `gfg` | `gofurry_app` |
| Admin primary / Nav / Game pools | `gfa` / `gfn` / `gfg` | `gofurry_app` |

All Redis runtime configurations use the `gofurry_app` ACL username and the corresponding application password. `redis_username` remains optional in configuration for compatibility with existing password-only Redis deployments.

Keep PostgreSQL responsibilities separate:

- `gofurry_app` is the only normal runtime account and performs business DML.
- `gofurry_migrator` is reserved for Goose and must never appear in runtime YAML.
- `gofurry_readonly` is for GUI and manual read-only inspection.
- `gofurry_admin` is reserved for infrastructure administration and does not belong in the root `.env` or ordinary local runtime configuration.

Run pinned Goose from the local `tools/` module with the database-specific migrator URL from the root `.env`: `GOFURRY_GFN_MIGRATOR_URL`, `GOFURRY_GFG_MIGRATOR_URL`, or `GOFURRY_GFA_MIGRATOR_URL`. Load the selected value into the local shell without printing it, and never substitute an application or infrastructure-admin account. Local migrator URLs use `sslmode=disable` only because the private connection is transported inside Tailscale's encrypted tunnel; reassess that setting if the network boundary changes.

The shared database state and operator boundaries are documented in [Development infrastructure operations](./operations/dev-infrastructure.md).

## Game pages over remote development infrastructure

Game Backend may opt into `server.development_home_cache_seconds` in its explicit local YAML (for example `10`, maximum `30`). It defaults to `0` (disabled) and a nonzero value is rejected outside `debug` mode. The two CN homepage language snapshots are held briefly in process to avoid repeatedly transferring the large Redis payload over Tailscale. Other processes' cache updates can take up to that interval to appear locally. Expired entries never mask a failed Redis read, and the existing Redis keys, scheduled refreshes, and production cache behavior stay the same. Restart the local Game Backend after changing this setting.

After a Nav Web production build, `npm run game:detail:smoke` uses isolated fixture APIs to verify full homepage/detail SSR, title/description/canonical/hreflang, empty historical data, tab switching, and authoritative 404/503 responses. No live database or collector is required.

## Normal validation

For each of `game-collector`, `game-backend`, `nav-collector`, `nav-backend`, `admin`, and `uptime` under `apps/cn`:

~~~bash
gofmt -w .
go vet ./...
go test ./...
go build ./...
~~~

Frontends:

~~~bash
cd apps/cn/admin/react
npm ci
npm run typecheck
npm test
npm run build

cd ../../nav-web
npm ci
npm run typecheck
npm run build
~~~

Repository tooling from `tools`:

~~~bash
go test ./...
go tool sqlc vet -f ../sqlc.yaml
go tool sqlc generate -f ../sqlc.yaml
go run ./check-sqlc
go run ./check-production-policy
~~~

Run vulnerability scans for each active module using the pinned tool:

~~~bash
cd tools
go tool govulncheck -C ../apps/cn/nav-backend ./...
~~~

Database integration checks require explicitly configured, isolated development databases. Never point them at production.

The standalone uptime service has no PostgreSQL integration suite. Its tests use temporary Bbolt files. Collector health listeners are disabled when `health` is omitted; example configs bind them to loopback addresses.
