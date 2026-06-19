# GoFurry Nginx Config Notes

## Keep Nuxt 404 Pages Visible

The production nginx config currently enables `proxy_intercept_errors on` at the `http` level. That is fine for API upstreams, but the frontend server must override it for Nuxt routes. Otherwise Nuxt's 404 response is intercepted by nginx and falls back to nginx's own 404 handling.

Use `ops/nginx/gofurry-nav-web.locations.conf` inside the `server_name go-furry.com` block, or copy its three frontend locations into the existing server block. Do not keep this frontend-only pair unless it serves a real file:

```nginx
error_page 404 /404.html;
location = /404.html { internal; }
```

Nuxt now provides `gofurry-nav-web/app/error.vue`, so unknown frontend routes render the GoFurry `Page Not Found` page while preserving the HTTP 404 status.

## Maintenance Mode

`ops/nginx/nginx.maintenance.conf` is a full nginx config for planned downtime. It serves the static page at:

```text
/home/gofurry/gfs/gofurry-repo/gofurry-nav-site/unavailable/index.html
```

Suggested switch flow on the server:

```bash
cp /usr/local/nginx/conf/nginx.conf /usr/local/nginx/conf/nginx.conf.bak.$(date +%Y%m%d%H%M%S)
cp /home/gofurry/gfs/gofurry-repo/gofurry-nav-site/ops/nginx/nginx.maintenance.conf /usr/local/nginx/conf/nginx.conf
/usr/local/nginx/sbin/nginx -t
/usr/local/nginx/sbin/nginx -s reload
```

The maintenance page intentionally returns HTTP `503 Service Unavailable` with `Retry-After: 3600`, which is healthier for SEO and crawlers than returning a broken 502/404 or a fake 200 during downtime.
