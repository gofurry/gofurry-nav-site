# gofurry-nav-web

The source workspace is `apps/cn/web`. The package, Task entrypoints, CI checks
and production Compose project retain their existing `nav-web` identities.

Nuxt 4 frontend for gofurry navigation and game discovery.

This project replaces the legacy Vue SPA frontend for the public site. It uses the versioned Go API path under `/api/v1` and focuses on SSR, prerendering, and SEO-friendly public pages.

## Scripts

Use Node 24 and pnpm 12.6.0 from `packageManager`. This project owns its lockfile
and local script permissions, without a root workspace. Root `task deps:nav-web`,
`task dev:nav-web`, `task test:nav-web` and `task build:nav-web` are convenience
entrypoints; Docker image building is separate via `task build:nav-web-image`.

```bash
pnpm install --frozen-lockfile
pnpm run dev
pnpm run typecheck
pnpm run build
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

## Public Ecosystem

The public Ecosystem product is SSR-rendered at `/insights`, `/insights/sites`, and `/insights/games`, with English equivalents under `/en`. It consumes the stable Nav and Game `/api/v2/*/insights` contracts. Metric and range selections are stored in the URL query.

Primary public system navigation icons use Phosphor. The default layout owns the canvas through `PublicPageBackground.vue`: a fixed solid canvas (Light `#f2e3d0`, Dark `#1d1f22`) and page-top CSS radial dots with nested linear masks. Desktop sides use a 26px grid, 0.8px radius, `min(22%, 340px)` widths and a 520px downward fade; mobile uses a 28px top grid fading over 185px. Dot RGB/maximum alpha are Light `#806c58`/0.24 and Dark `#c1c5cb`/0.22, reduced to 0.16/0.15 on tablets and 0.12/0.10 on phones. No image, script, pattern catalog or page-background preference is involved. Pages must not instantiate or cover their own background. Home Hero local images remain independent; Preferences retains Home and Resource routing.

Run the deterministic route and interaction regressions against the production build:

```bash
pnpm run build
pnpm run test:browser insights
```

The legacy visual report guard is retired. Run `pnpm run test:browser` for runtime
contracts and `pnpm run test:visual` in the pinned Linux environment for accepted
pixel comparisons. See [testing](../../../docs/frontend/testing.md) for all gates;
performance tools retain their separate scope.

## Production

See [DEPLOYMENT.md](./DEPLOYMENT.md) for the Docker-based production migration path and the nginx reverse-proxy snippet for `go-furry.com`.
