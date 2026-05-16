# gofurry-nav-web

Nuxt 4 frontend for gofurry navigation and game discovery.

This project replaces the legacy Vue SPA frontend for the public site. It keeps the existing Go APIs unchanged and focuses on SSR, prerendering, and SEO-friendly public pages.

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
NAV_API_INTERNAL_BASE=http://127.0.0.1:9999/api
GAME_API_INTERNAL_BASE=http://127.0.0.1:9998/api
NUXT_PUBLIC_NAV_API_BASE=/api
NUXT_PUBLIC_GAME_API_BASE=/api
NUXT_PUBLIC_SITE_URL=http://localhost:3000
```

## Production

See [DEPLOYMENT.md](./DEPLOYMENT.md) for the Docker-based production migration path and the nginx reverse-proxy snippet for `go-furry.com`.
