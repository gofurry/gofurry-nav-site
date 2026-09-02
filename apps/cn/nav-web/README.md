# gofurry-nav-web

Nuxt 4 frontend for gofurry navigation and game discovery.

This project replaces the legacy Vue SPA frontend for the public site. It uses the versioned Go API path under `/api/v1` and focuses on SSR, prerendering, and SEO-friendly public pages.

## Scripts

```bash
npm install
npm run dev
npm run typecheck
npm run build
```

## Runtime Config

Use UTF-8 for all source files and environment files.

```bash
NAV_API_INTERNAL_BASE=http://127.0.0.1:9999/api/v1
NAV_V2_API_INTERNAL_BASE=http://127.0.0.1:9999/api/v2
GAME_API_INTERNAL_BASE=http://127.0.0.1:9998/api/v1
GAME_V2_API_INTERNAL_BASE=http://127.0.0.1:9998/api/v2
NUXT_PUBLIC_NAV_API_BASE=/api/v1
NUXT_PUBLIC_NAV_V2_API_BASE=/api/v2
NUXT_PUBLIC_GAME_API_BASE=/api/v1
NUXT_PUBLIC_GAME_V2_API_BASE=/api/v2
NUXT_PUBLIC_SITE_URL=http://localhost:3000
```

## Public Insights

The public Insights product is SSR-rendered at `/insights`, `/insights/sites`, and `/insights/games`, with English equivalents under `/en`. It consumes the stable Nav and Game `/api/v2/*/insights` contracts. Metric and range selections are stored in the URL query.

After starting a production preview, run the focused route and interaction smoke plus the visual guard:

```bash
npm run insights:smoke -- --base-url http://localhost:3000
npm run visual:guard -- --base-url http://localhost:3000
```

## Production

See [DEPLOYMENT.md](./DEPLOYMENT.md) for the Docker-based production migration path and the nginx reverse-proxy snippet for `go-furry.com`.
