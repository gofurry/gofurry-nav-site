# Local development

## Prerequisites

- Go 1.26.7 for all six active Go modules and `tools`
- Task >=3.45.3 (built-in file utilities on Windows)
- Node.js 24 and pnpm 12.6.0 for `apps/cn/nav-web` and `apps/cn/admin/react`
- development PostgreSQL and Redis matching the selected app configuration
- sqlc and govulncheck through the pinned Go tools in `tools/go.mod`

The repository has no `go.work` or root pnpm workspace. Each frontend pins
`packageManager: pnpm@12.6.0`, owns its `pnpm-lock.yaml`, and uses its local
`pnpm-workspace.yaml` only for single-project settings. Keep Node/Go/Task/pnpm
installation explicit; `task doctor` only reports versions. For a workstation
bootstrap, `npm install --global pnpm@12.6.0` installs the package manager itself;
all project dependencies and scripts use pnpm. Corepack is not required.

## Repository commands

Run `task` or `task --list` from the root for the public commands. Commands also
work with Task's root-file discovery from a subdirectory.

| Command | Scope |
| --- | --- |
| `task doctor`, `task doctor:docker` | Read-only toolchain check; Docker is optional for the base doctor |
| `task deps`, `task deps:admin`, `task deps:nav-web` | Always run frozen installs; no node_modules existence shortcut |
| `task dev:nav-web`, `task dev:admin-web` | Start one frontend development server |
| `task dev:nav-backend`, `task dev:nav-collector`, `task dev:game-backend`, `task dev:game-collector`, `task dev:admin-api`, `task dev:uptime` | Start one Go service using its existing local YAML; fail if config is absent |
| `task fmt`, `task check:fmt` | Write/check Go formatting for six modules plus tools |
| `task lint`, `task lint:go`, `task lint:admin`, `task lint:nav-web` | Go vet, Admin lint, Nav ESLint/Stylelint/style policy |
| `task typecheck`, `task typecheck:admin`, `task typecheck:nav-web` | Frontend typing |
| `task test`, `task test:go`, `task test:admin`, `task test:nav-web` | Ordinary local tests; no integration configuration is supplied |
| `task test:nav-web:browser` | Explicit serial production Browser suite after a build and `pnpm exec playwright install chromium` |
| `task check:policy`, `task check:sqlc` | Production policy; SQL vet and temporary generation comparison without rewriting repository files |
| `task generate:sqlc` | Explicitly regenerate committed sqlc outputs |
| `task check` | Formatting, lint, typing, policy and sqlc; no source changes |
| `task verify` | Frozen deps → frontend builds → check → test → ordinary Go builds |
| `task verify:frontend-build`, `task verify:go-build` | Individual local build checks |
| `task build` | Exactly six Linux/amd64 Go release artifacts |
| `task build:nav-backend`, `task build:nav-collector`, `task build:game-backend`, `task build:game-collector`, `task build:admin`, `task build:uptime` | Individual release artifacts |
| `task build:nav-web`, `task build:nav-web-image` | Separate Nuxt production build / Docker image build only |
| `task clean` | Root `build/` only; preserves configs, dependencies, lockfiles and package store |

On a fresh checkout use `task verify`: Admin assets must exist before Go vet,
tests or builds compile `go:embed`. For focused `task check` / `task test`, first
run `task deps` and `pnpm --dir apps/cn/admin/react run build` if embed output is
missing. Checks do not silently build it. Typechecks may update ignored compiler
caches; they do not edit source. No Task starts the whole stack, loads root `.env`,
runs Goose, operates systemd/cloud resources or deploys an image.
Browser, pinned Linux Visual, Docker and isolated PostgreSQL/provider acceptance
remain separate explicit gates; `task verify` does not claim those passed.

Frontend dependency changes use `pnpm add/remove/update` in the owning project.
When migrating an existing npm checkout, move aside its old `node_modules` once
before the first frozen pnpm install; stale hoisted modules can mask resolution
errors. Ordinary `task deps` never removes dependency directories.
Do not regenerate locks just to refresh ranges. Nav Web allows only `esbuild`
(platform binary validation) and `vue-demi` (Vue bridge selection) install scripts.
Published optional native binaries serve `@parcel/watcher` and `unrs-resolver`;
their rebuild scripts and optional `fsevents` rebuilds are denied. Admin needs
no dependency build scripts. Unknown scripts still fail installation. Both projects
set `verifyDepsBeforeRun: error` so running a check cannot silently reinstall.

Nav Web keeps narrow overrides for `vue` and `@vue/server-renderer` at the original
production lock's 3.5.33. The old npm lock nested Vue 3.5.43 only for test-utils;
pnpm's isolated resolution otherwise lets two Vue instances enter Nitro and
causes SSR 500 responses. Keep the runtime/renderer pair aligned when a future
Vue upgrade is explicitly scoped. This migration changes no dependency ranges
and introduces no newer frontend dependency versions.

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

After a Nav Web production build and `pnpm exec playwright install chromium`, `pnpm run test:browser --workers=1` runs the formal Shared/Nav/Games/Insights/SEO/background runtime contracts against isolated fixture APIs. It covers SSR/hydration, interactions, failures and responsive layout without a live database or collector. Pinned Linux `test:visual` separately compares accepted pixels; the old `visual:guard` runner is retired. See [frontend testing](frontend/testing.md) for the full gate sequence.

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
pnpm install --frozen-lockfile
pnpm run typecheck
pnpm test
pnpm run build

cd ../../nav-web
pnpm install --frozen-lockfile
pnpm run typecheck
pnpm run build
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
