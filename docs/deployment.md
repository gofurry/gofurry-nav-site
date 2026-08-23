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

## Go service release flow

Treat binary rollout and schema rollout as separate operator actions:

1. Build and verify release artifacts.
2. Back up the current binary, configuration, and systemd unit.
3. If the release contains Goose migrations, back up the affected database and run Goose explicitly from the matching root `db/*/migrations` directory.
4. Stop the affected service.
5. Apply the release's documented configuration changes and replace the binary.
6. Install or replace systemd registration as documented in [systemd operations](operations/systemd.md).
7. Review the generated unit, then start the service manually.
8. Check status, logs, readiness, API smoke tests, collector schedules, and Admin access.

Applications never run Goose during startup. Do not run a migration merely because a binary changed: first verify that the release actually includes a new migration.

This availability completion adds no business DDL and requires no database migration of its own.

## Availability rollout

Prefer running `gf-uptime` on a small host outside the business host's failure domain. It can run on the same host when necessary, but its runtime must still use only its local Bbolt file.

1. Copy `apps/cn/uptime/conf/server.example.yaml`, keep the five endpoint IDs stable, choose a durable absolute `storage.path`, and make the storage/log directories writable by the install-time runtime user.
2. Deploy `gf-uptime`, run `install --config /etc/gf-uptime/server.yaml`, review the generated unit, start it manually, and reverse proxy `status.go-furry.com` to its local listener. Redirect `/` to `/uptime` if the public hostname should open the UI at its root.
3. Add the optional loopback/private `health` block to each collector config, deploy the collector binaries, reinstall units with `--force` only after review, and restart them one at a time. One-shot Game commands remain listener-free.
4. Before deploying Nav Backend, remove its old `uptime` block and obsolete `waf.crs_root` key. Keep `waf.directives_files` when custom Coraza rules are required.
5. Set Nav Web `NUXT_PUBLIC_UPTIME_URL=https://status.go-furry.com`, redeploy it, and verify public `/healthz`.
6. Verify the status UI and all five monitored endpoints. Configure an external HTTP monitor for `https://status.go-furry.com/readyz`.

No Goose command is part of these six steps.

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

## Compatibility boundaries

A deployment of this cleanup must preserve:

- public API behavior, including existing Nav V1 availability;
- Redis key names, TTLs, and cache semantics;
- collector schedules, intervals, and one-shot behavior;
- PostgreSQL data and schema ownership;
- Admin's PostgreSQL-only three-pool model.

## Rollback

Keep the previous binary and unit backup until post-deployment checks pass. For a binary-only rollback, stop the service, restore the previous binary/unit, reload systemd, and start it. Database rollback must follow the migration-specific runbook; never improvise destructive down migrations against production.
