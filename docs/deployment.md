# Cross-service deployment

## Build ownership

From a Windows release workspace, `build.bat all` builds only the five active Linux/amd64 Go binaries:

| Target | Artifact |
|---|---|
| `nav-backend` | `build/gf-nav/gf-nav` |
| `nav-collector` | `build/gf-nav-collector/gf-nav-collector` |
| `game-backend` | `build/gf-game/gf-game` |
| `game-collector` | `build/gf-game-collector/gf-game-collector` |
| `admin` | `build/gofurry-admin/gofurry-admin` |

Admin's embedded web UI is built before its binary. `apps/cn/nav-web` has a separate npm/Docker deployment flow described in its local `DEPLOYMENT.md`.

`legacy`, `experimental`, `third-party`, and `apps/intl` are not production build targets.

## Go service release flow

Treat binary rollout and schema rollout as separate operator actions:

1. Build and verify release artifacts.
2. Back up the current binary, configuration, and systemd unit.
3. If the release contains Goose migrations, back up the affected database and run Goose explicitly from the matching root `db/*/migrations` directory.
4. Stop the affected service.
5. Replace the binary without changing the service's configuration schema.
6. Install or replace systemd registration as documented in [systemd operations](operations/systemd.md).
7. Review the generated unit, then start the service manually.
8. Check status, logs, readiness, API smoke tests, collector schedules, and Admin access.

Applications never run Goose during startup. Do not run a migration merely because a binary changed: first verify that the release actually includes a new migration.

This V3-preparation repository/runtime cleanup adds no business DDL and requires no database migration of its own.

## Compatibility boundaries

A deployment of this cleanup must preserve:

- public API behavior, including existing Nav V1 availability;
- Redis key names, TTLs, and cache semantics;
- collector schedules, intervals, and one-shot behavior;
- PostgreSQL data and schema ownership;
- Admin's PostgreSQL-only three-pool model.

## Rollback

Keep the previous binary and unit backup until post-deployment checks pass. For a binary-only rollback, stop the service, restore the previous binary/unit, reload systemd, and start it. Database rollback must follow the migration-specific runbook; never improvise destructive down migrations against production.
