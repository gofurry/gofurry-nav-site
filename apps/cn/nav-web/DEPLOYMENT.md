# Production migration for `gofurry-nav-web`

This frontend is designed to run inside Docker, so the production host does not need Node.js installed.

## Build and run

```bash
cp .env.production.example .env.production
docker compose -f docker-compose.prod.yml up -d --build
```

The container listens on `127.0.0.1:3000` through the published port mapping.

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

The repository also includes `../ops/nginx/gofurry-nav-web.locations.conf`, which can be copied into the `go-furry.com` server block.

Keep the existing `nav.go-furry.com` and `game.go-furry.com` API server blocks unchanged.

## Maintenance page

For planned downtime, switch nginx to the dedicated maintenance config instead of letting requests fail against stopped services:

```bash
cp /usr/local/nginx/conf/nginx.conf /usr/local/nginx/conf/nginx.conf.bak.$(date +%Y%m%d%H%M%S)
cp /home/gofurry/gfs/gofurry-repo/gofurry-nav-site/ops/nginx/nginx.maintenance.conf /usr/local/nginx/conf/nginx.conf
/usr/local/nginx/sbin/nginx -t
/usr/local/nginx/sbin/nginx -s reload
```

The maintenance config serves `../unavailable/index.html` as a self-contained native HTML/CSS/JS page and returns HTTP `503 Service Unavailable` with `Retry-After`.

## Environment variables

The Docker image uses the versioned public API path and CDN values, plus internal API bases for SSR requests.

Required values:

- `NUXT_PUBLIC_SITE_URL=https://go-furry.com`
- `NUXT_PUBLIC_NAV_API_BASE=https://nav.go-furry.com/api/v1`
- `NUXT_PUBLIC_GAME_API_BASE=https://game.go-furry.com/api/v1`
- `NAV_API_INTERNAL_BASE=http://10.6.0.11:9999/api/v1`
- `GAME_API_INTERNAL_BASE=http://10.6.0.11:9998/api/v1`

The CDN and logo URLs stay pointed at the existing `qcdn.go-furry.com` assets.

## Notes

- `robots.txt` and `sitemap.xml` are served from Nuxt `server/routes`.
- Deploy the Go nav/game backends and the Nuxt frontend together during a maintenance window. The public nav/game APIs now live under `/api/v1`, and the old non-versioned API aliases are intentionally removed.
- The old Vue frontend can stay in the repository as a legacy reference, but it is no longer the production entrypoint.
