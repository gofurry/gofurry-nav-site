# Cross-service deployment

## Build ownership

From a Windows release workspace, `build.bat all` builds only the six active Linux/amd64 Go binaries:

| Target | Artifact |
|---|---|
| `nav-backend` | `build/gf-nav/gf-nav` |
| `nav-collector` | `build/gf-nav-collector/gf-nav-collector` |
| `game-backend` | `build/gf-game/gf-game` |
| `game-collector` | `build/gf-game-collector/gf-game-collector` |
| `admin` | `build/gofurry-admin/gofurry-admin` |
| `uptime` | `build/gf-uptime/gf-uptime` |

Admin's embedded web UI is built before its binary. `apps/cn/nav-web` has a separate npm/Docker deployment flow described in its local `DEPLOYMENT.md`.

`legacy`, `experimental`, `third-party`, and `apps/intl` are not production build targets.

## Go service deployment

Treat binary deployment and schema migration as separate operator actions:

1. Build and verify release artifacts.
2. Back up the current binary, configuration, and systemd unit.
3. If the release contains Goose migrations, back up the affected database and run Goose explicitly from the matching root `db/*/migrations` directory.
4. Stop the affected service.
5. Apply the release's documented configuration changes and replace the binary.
6. Install or replace systemd registration as documented in [systemd operations](operations/systemd.md).
7. Review the generated unit, then start the service manually.
8. Check status, logs, readiness, API smoke tests, collector schedules, and Admin access.

Applications never run Goose during startup. Do not run a migration merely because a binary changed: first verify that the release actually includes a new migration.

## Availability operation

Prefer running `gf-uptime` on a small host outside the business host's failure domain. It can run on the same host when necessary, but its runtime must still use only its local Bbolt file.

- Keep endpoint IDs stable and store the Bbolt file at a durable absolute path writable by the service user.
- Expose the status UI through Nginx while keeping the application listener local to its host.
- Bind Collector health listeners only to loopback or private addresses; one-shot Game commands remain listener-free.
- Set Nav Web `NUXT_PUBLIC_UPTIME_URL` to the independent public status origin.
- Use an external HTTP monitor for the public `gf-uptime` readiness endpoint.

On the status host, the relevant Nginx locations are intentionally small:

~~~nginx
location = / {
    return 302 /uptime;
}

location / {
    proxy_pass http://127.0.0.1:9980;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
~~~

Terminate TLS using the host's existing certificate workflow. If uptime runs on a separate host, replace collector/backend loopback endpoint URLs with reachable private or HTTPS health URLs; never expose collector listeners publicly merely to make status polling work.

## Deployment boundaries

Routine deployments must preserve:

- public API behavior, including existing Nav V1 availability;
- Redis key names, TTLs, and cache semantics;
- collector schedules, intervals, and one-shot behavior;
- PostgreSQL data and schema ownership;
- Admin's PostgreSQL-only three-pool model.

## Rollback

Keep the previous binary and unit backup until post-deployment checks pass. For a binary-only rollback, stop the service, restore the previous binary/unit, reload systemd, and start it. Database rollback must follow the migration-specific runbook; never improvise destructive down migrations against production.
