# Production deployment for `gofurry-nav-web`

This frontend runs inside Docker; the production host needs Docker and Compose,
not Node.js, pnpm or Task. The build context remains `apps/cn/web`.

For the alpha.9 release, use the [cross-service upgrade guide](../../../docs/releases/v3.0.0-alpha.9.md).
The frontend-only engineering changes require no new runtime variables, but the
full alpha.8 → alpha.9 upgrade also includes backend/schema changes.

## #145 frontend handoff (maintainer only)

Use the [C2 acceptance ledger](../../../docs/acceptance/145-c2-integration-final-acceptance.md)
for the actual accepted source and Fast/Repository Policy, Full four groups/image,
and Visual results. Deployment is blocked until every required gate passes.
The repository default branch is `main`; `git pull --ff-only` updates only the
server's current branch, never imports `dev` into `main`. Confirm that the chosen
deployment branch contains the accepted C2 source. If production uses `main`,
merge through the approved process first and confirm that branch's Full gate.

Before pulling, check branch/status, retain a usable prior image plus matching
private configuration, and confirm sufficient build memory. From `apps/cn/web`,
validate Compose quietly (`docker compose -f docker-compose.prod.yml config -q`)
and inspect only its project name: it must remain `nav-web`. Then the maintainer
may run `./update.sh`, inspect Compose status and `/healthz`, and smoke-test `/`,
`/en`, `/games`, `/games/search`, `/site/:id` and `/site-groups/:id` on desktop/mobile.
Check natural scrolling, navigation handoff, utility search, shortcut sorting and
zero hydration/JS errors. C2 includes no database migration or backend restart.

`update.sh` builds and replaces containers; Codex does not execute it. On failure,
do not force a switch or prune. Recover with the retained image and configuration;
pulling old source alone is not image rollback. Close #145 only after maintainer
deployment and production smoke acceptance.

## Existing installation: workspace rename (maintainer only)

The source moved from `apps/cn/nav-web` to `apps/cn/web`. The maintainer confirmed
the existing container's `com.docker.compose.project` label is **`nav-web`**.
Compose now explicitly declares `name: nav-web`; service, container, image,
port, runtime defaults and `update.sh` remain unchanged. Do not start a second
project named `web`. `COMPOSE_PROJECT_NAME` or `docker compose -p` can override
the file, so verify the rendered project name before starting any container.

These are production instructions for the maintainer, not operations for Codex.
Before updating, check the deployed branch and working tree; the accepted commit
must already belong to that deployment branch (do not assume `dev`). Preserve any
local Dockerfile changes. Record the current container's project label, back up
the old `.env` and `.env.next` securely without printing their contents, and check
`free -h` plus container/BuildKit memory limits.

```bash
docker inspect gofurry-nav-web \
  --format '{{ index .Config.Labels "com.docker.compose.project" }}'
# Must be nav-web.

cd /home/gofurry/gfs/gofurry-repo/gofurry-nav-site
# Only after the branch/working-tree checks, use the approved release workflow;
# if appropriate for that deployment branch:
git pull --ff-only || exit 1
test -f apps/cn/web/docker-compose.prod.yml || exit 1

# Stop and review if destination private files already exist; never overwrite them.
test ! -e apps/cn/web/.env && test ! -e apps/cn/web/.env.next || exit 1
install -m 600 apps/cn/nav-web/.env apps/cn/web/.env || exit 1
if test -f apps/cn/nav-web/.env.next; then
  install -m 600 apps/cn/nav-web/.env.next apps/cn/web/.env.next || exit 1
fi
```

Git moves tracked files only; ignored private files can remain in the old
directory. Copy these files individually and retain the old copies for rollback.
Do not move the entire directory, replace private files with examples, or delete
old configuration during this migration. Do not publish complete `compose config`
output because it may expose private values.

```bash
cd /home/gofurry/gfs/gofurry-repo/gofurry-nav-site/apps/cn/web
docker compose -f docker-compose.prod.yml config -q || exit 1
project_name=$(docker compose -f docker-compose.prod.yml config | grep '^name:')
printf '%s\n' "$project_name"
test "$project_name" = 'name: nav-web' || exit 1
# Build first; only proceed to up after a successful build and adequate memory.
docker compose -f docker-compose.prod.yml build || exit 1
docker compose -f docker-compose.prod.yml up -d --no-build --remove-orphans || exit 1
docker inspect gofurry-nav-web \
  --format '{{ index .Config.Labels "com.docker.compose.project" }}'
docker compose -f docker-compose.prod.yml ps
curl -fsS http://127.0.0.1:3000/healthz
```

The label must remain `nav-web`. On config/build/memory failure, do not run `up`,
`down` or `prune`; keep the old container, image and private files. If the replaced
container fails health checks, restore the verified prior image/deployment using
the retained configuration. A Git checkout alone does not restore an overwritten
image. After successful migration, run `./update.sh` only from the new directory;
clean up old private files only after the maintainer confirms stability and that
no operational scripts still depend on their old paths.

## Build memory

Only Docker's `build` stage sets `NODE_OPTIONS=--max-old-space-size=...`, from
`NODE_BUILD_HEAP_MB` (default `4096`). Compose passes this as a **build argument**,
not a runtime environment variable. `runner` starts independently from `deps`
and does not inherit the build heap setting. An adequately provisioned build host
may override it, for example `NODE_BUILD_HEAP_MB=6144 docker compose -f docker-compose.prod.yml build`.

4096 MiB is a V8 heap limit, not reserved memory or proof of sufficient host RAM.
Nuxt/Nitro and other processes also consume memory. Check host and cgroup limits;
when memory is insufficient, build the compatible image elsewhere rather than
repeatedly forcing builds on a low-memory production host.

## Build and run

Run from `apps/cn/web`. Existing deployments using `.env` can keep their
configuration and the original update command; Compose reads `.env` automatically:

```bash
./update.sh
```

Copy an example only on first deployment; never overwrite existing private
values. If you choose the alternate filename `.env.production`, pass it explicitly
instead of using `update.sh` (which intentionally uses Compose's default `.env`):

```bash
test -f .env.production || cp .env.production.example .env.production
# Adjust private deployment values before the first start.
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build
```

The container listens on `127.0.0.1:3000` through the published port mapping.

Docker caches dependency installation before copying source. The `postinstall`
prepare in that partial tree is insufficient: the build stage runs `nuxt prepare`
again after copying the complete app/config, then `nuxt build`. Keep both steps;
otherwise imported Vue prop types may resolve through stale TypeScript aliases.

The dependency stage bootstraps the exact pnpm version from `packageManager`,
then installs `pnpm-lock.yaml` with `--frozen-lockfile` and project-local script
permissions. The complete-source stage runs `pnpm exec nuxt prepare` and
`pnpm run build`. For a local image-only check use root `task doctor:docker` and
`task build:web-image`; neither starts containers nor invokes `update.sh`.

## nginx change for `go-furry.com`

Replace the old static frontend block:

```nginx
root /home/gofurry/gfs/frontend/www/;
index index.html;

location / {
    try_files $uri $uri/ /index.html;
}
```

with a reverse proxy to the Nuxt container:

```nginx
location / {
    proxy_intercept_errors off;
    proxy_pass http://127.0.0.1:3000;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}
```

If the global nginx config keeps `proxy_intercept_errors on`, the frontend locations must set `proxy_intercept_errors off`. Nuxt owns the frontend 404 page through `app/error.vue`; letting nginx intercept upstream 404 responses will replace the app page with nginx's fallback 404.

The repository also includes `../../../ops/nginx/gofurry-nav-web.locations.conf`, which can be copied into the `go-furry.com` server block.

Keep the existing `nav.go-furry.com` and `game.go-furry.com` API server blocks unchanged.

## Maintenance page

Planned maintenance only places GoFurry business hosts into maintenance mode. `status.go-furry.com` remains live and continues proxying the independent `gf-uptime` service. The canonical maintenance config preserves its separate HTTP redirect and HTTPS proxy to `http://127.0.0.1:9980`.

For planned downtime, switch nginx to the dedicated maintenance config instead of letting requests fail against stopped services:

```bash
cp /usr/local/nginx/conf/nginx.conf /usr/local/nginx/conf/nginx.conf.bak.$(date +%Y%m%d%H%M%S)
cp /home/gofurry/gfs/gofurry-repo/gofurry-nav-site/ops/nginx/nginx.maintenance.conf /usr/local/nginx/conf/nginx.conf
/usr/local/nginx/sbin/nginx -t
/usr/local/nginx/sbin/nginx -s reload
```

The maintenance config serves `../../../ops/nginx/static/unavailable/index.html` as a self-contained native HTML/CSS/JS page and returns HTTP `503 Service Unavailable` with `Retry-After`.

## Environment variables

The Docker image uses the versioned public API path and CDN values, plus internal API bases for SSR requests.

Required values:

- `NUXT_PUBLIC_SITE_URL=https://go-furry.com`
- `NUXT_PUBLIC_NAV_API_BASE=https://nav.go-furry.com/api/v1`
- `NUXT_PUBLIC_NAV_V2_API_BASE=https://nav.go-furry.com/api/v2`
- `NUXT_PUBLIC_GAME_API_BASE=https://game.go-furry.com/api/v1`
- `NUXT_PUBLIC_UPTIME_URL=https://status.go-furry.com`
- `NAV_API_INTERNAL_BASE`: Nav V1 address reachable from the container
- `NAV_V2_API_INTERNAL_BASE`: Nav V2 address reachable from the container
- `GAME_API_INTERNAL_BASE`: Game API address reachable from the container

Managed object keys use `NUXT_PUBLIC_ASSET_PRIMARY_BASE=https://assets.go-furry.com` (COS / EdgeOne) and `NUXT_PUBLIC_ASSET_MIRROR_BASE=https://assets.gofurry.com` (R2 / Cloudflare). Fixed platform icons, About portraits, tool covers and error illustrations are bundled with Nuxt. The [managed asset cutover runbook](../../../docs/managed-assets-cutover.md) is only for first deployment from the old model or historical recovery; do not rerun initial cutover SQL on an already-migrated installation.

## Notes

- `robots.txt` and `sitemap.xml` are served from Nuxt `server/routes`.
- `GET /healthz` returns a dependency-free HTTP 200 response for CDN/Nginx/Nuxt reachability checks.
- The Footer status link uses the independent status service; it no longer points at Nav Backend.
- Deploy backend contract changes before the dependent Nuxt frontend; use the release guide to determine which services actually changed. Preserve the existing versioned `/api/v1` and `/api/v2` proxy configuration and authoritative Nuxt 404/503 responses.
- The old Vue frontend can stay in the repository as a legacy reference, but it is no longer the production entrypoint.
