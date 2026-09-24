# Production deployment for `gofurry-nav-web`

This frontend is designed to run inside Docker, so the production host does not need Node.js installed.

For the alpha.9 release, use the [cross-service upgrade guide](../../../docs/releases/v3.0.0-alpha.9.md).
The frontend-only engineering changes require no new runtime variables, but the
full alpha.8 → alpha.9 upgrade also includes backend/schema changes.

## Build and run

Run from `apps/cn/nav-web`. Existing deployments using `.env` can keep their
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
