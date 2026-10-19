# Cross-service deployment

For alpha.8 → alpha.9, follow the [release-specific upgrade guide](releases/v3.0.0-alpha.9.md), including the irreversible Game Tag migration and coordinated runtime replacement.

## Build ownership

From a Windows, Linux or macOS release workspace, `task build` builds only the five
active Linux/amd64 Go binaries with `CGO_ENABLED=0`, `-trimpath` and
`-ldflags="-s -w"`. Individual targets use `task build:<target>`:

| Target | Artifact |
|---|---|
| `nav-backend` | `build/gf-nav/gf-nav` |
| `nav-collector` | `build/gf-nav-collector/gf-nav-collector` |
| `game-backend` | `build/gf-game/gf-game` |
| `game-collector` | `build/gf-game-collector/gf-game-collector` |
| `admin` | `build/gofurry-admin/gofurry-admin` |

Admin's embedded web UI is built before its binary. `apps/cn/web` has a separate pnpm/Docker deployment flow described in its local `DEPLOYMENT.md`.

`task build:admin` always installs the frozen React lock, clears/rebuilds the
canonical embed output through Vite, then builds Go and copies the same files to
`build/gofurry-admin/dist/`. `task build:nav-web-image` builds an image only, using
`apps/cn/web` as its context. Production still uses `cd apps/cn/web` and
`./update.sh`; the host needs Docker/Compose, not Node, pnpm or Task. No Task wraps
deployment, migration, systemd or cloud operations.

For an existing `apps/cn/nav-web` installation, follow the maintainer-only
[workspace migration instructions](../apps/cn/web/DEPLOYMENT.md#existing-installation-workspace-rename-maintainer-only)
before using the new directory. Preserve private `.env`/`.env.next`, verify the
rendered Compose project remains `nav-web`, and build elsewhere if production
memory is insufficient. The configurable 4096 MiB heap applies only to the image
build stage; it does not change the runtime environment.

`legacy`, `experimental`, `third-party`, and `apps/intl` are not production build targets.

## Go service deployment

Treat binary deployment and schema migration as separate operator actions:

1. Build and verify release artifacts.
2. Back up the current binary, configuration, and systemd unit.
3. If the release contains Goose migrations, back up the affected database and run Goose explicitly from the matching root `db/*/migrations` directory. For incompatible/destructive schema changes, stop all old runtime clients and writes **before** Goose; the release-specific cutover order takes precedence.
4. Stop the affected service.
5. Apply the release's documented configuration changes and replace the binary.
6. Install or replace systemd registration as documented in [systemd operations](operations/systemd.md).
7. Review the generated unit, then start the service manually.
8. Check status, logs, readiness, API smoke tests, collector schedules, and Admin access.

Applications never run Goose during startup. Do not run a migration merely because a binary changed: first verify that the release actually includes a new migration.

## Availability operation

Public availability has moved to [LibreCosta Uptime](https://github.com/librecosta/uptime) v0.1.0. It runs as `librecosta-uptime.service` on `127.0.0.1:18580`; `status.go-furry.com` already points to that service. Deployment, configuration and upgrades belong to the independent repository.

The former `gf-uptime.service` is stopped and disabled. Its source is preserved in `legacy/gf-uptime`, excluded from GoFurry builds and scans. Old history does not need migration; this repository cleanup performs no production operations.

- Preserve the existing Nav Web status-page links and `NUXT_PUBLIC_UPTIME_URL`.
- Keep backend/Admin health routes and Collector loopback/private readiness listeners unchanged.
- The repository maintenance config keeps the independent status virtual host live while business hosts serve the maintenance page. Its status upstream is `http://127.0.0.1:18580`; certificates and other proxies are unchanged.
- Follow the independent project's endpoint and storage contract rather than the archived Fiber/Bbolt documentation. Never expose Collector listeners publicly merely to make polling work.

## Deployment boundaries

Routine deployments must preserve:

- public API behavior, including existing Nav V1 availability;
- Redis key names, TTLs, and cache semantics;
- collector schedules, intervals, and one-shot behavior;
- PostgreSQL data and schema ownership;
- Admin's PostgreSQL-only three-pool model.

## Rollback

Keep the previous binary and unit backup until post-deployment checks pass. For a binary-only rollback, stop the service, restore the previous binary/unit, reload systemd, and start it. Database rollback must follow the migration-specific runbook; never improvise destructive down migrations against production.

## Admin Collaboration Center (#117)

Back up GFA, coordinate/stop Admin writes, manually apply Goose migrations from `db/admin/migrations` with `gofurry_migrator`, then deploy the new Admin binary with embedded React. Apply through `20260926010000` for the GFA collaboration canvas. This renames the old note table into canvas nodes and adds edges, so deploy a matching Admin binary; rollback requires coordinated GFA restore and binary rollback. No GFG/GFN migration or Game/Nav Backend/Collector/Nav Web deployment is required for #117. Verify ideas, Board, capabilities, Audit and create-success/link-failure recovery using the [acceptance guide](collaboration-center.md).
