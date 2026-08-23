<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26.7-00ADD8?logo=go&logoColor=white" alt="Go Version">&nbsp&nbsp
  <img src="https://img.shields.io/github/license/gofurry/gofurry-nav-site" alt="License">&nbsp&nbsp
  <img src="https://img.shields.io/badge/weekend-project-8B5CF6?style=flat" alt="Weekend Project">&nbsp&nbsp
  <img src="https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange" alt="Made with Love">
</p>

<p align="center">
  ⭐🐺 <a href="https://github.com/gofurry/gofurry-nav-site/README.md">中文说明</a> | 
  <a href="https://go-furry.com">GoFurry China</a> | 
  <a href="https://gofurry.com">GoFurry</a> 
  🐺⭐
</P>

gofurry is an open-source multi-service repository for furry culture discovery. It contains the public web frontend, navigation and game APIs, data collectors, the operations admin backend, and operations-related modules.

The production public site has now moved to a Nuxt 4 frontend to improve SEO, first-page rendering, and public-content discoverability. The older Vue frontend is kept in the repository as an archived migration reference rather than the active production entrypoint.

```text
                  ░██████             ░██████████                                        
                 ░██   ░██            ░██                                                
                ░██         ░███████  ░██        ░██    ░██ ░██░████ ░██░████ ░██    ░██ 
                ░██  █████ ░██    ░██ ░█████████ ░██    ░██ ░███     ░███     ░██    ░██ 
                ░██     ██ ░██    ░██ ░██        ░██    ░██ ░██      ░██      ░██    ░██ 
                ░██   ░███ ░██    ░██ ░██        ░██   ░███ ░██      ░██      ░██   ░███ 
                  ░█████░█  ░███████  ░██         ░█████░██ ░██      ░██       ░█████░██ 
                                                                                    ░██ 
                                                                              ░███████
```

## Project Scope

- Public website: `https://go-furry.com`
- Audience: developers and maintainers who want to understand, run, extend, or deploy gofurry
- Repository model: a service-oriented monorepo rather than a single runnable app

## Repository Layout

- `apps/cn/nav-web`: Nuxt 4 frontend used by the production public site
- `apps/cn/nav-backend`: navigation API service
- `apps/cn/nav-collector`: navigation data collector
- `apps/cn/game-backend`: game-related API service
- `apps/cn/game-collector`: game-related data collector
- `apps/cn/admin`: operations backend with embedded frontend
- `apps/intl`: placeholders for the future international site; not a production build target
- `db`: Goose migrations for the `gfg`, `gfn`, and `gfa` databases
- `docs`: cross-service architecture, development, deployment, and operations documentation
- `ops`: Nginx, maintenance-page, and WAF deployment assets
- `legacy`: decommissioned modules, including the old Vue frontend and former RAG service, kept only for historical reference
- `third-party`: maintained dependency source mirrors, outside production CI
- `experimental`: experimental code not included in normal release packaging
- `tools`: helper scripts and local tools

## Stack

- Go
- Fiber
- PostgreSQL
- Redis
- Nuxt 4 / Vue 3
- Tailwind CSS

## Quick Start

Services in this repository are developed and run independently. These are the most common entry points today.

Frontend development:

```bash
cd apps/cn/nav-web
npm install
npm run dev
```

Go service development:

```bash
cd apps/cn/nav-backend
go run .
```

If you need root-level packaging artifacts, use:

```bat
build.bat all
```

Current targets include:

- `nav-backend`
- `nav-collector`
- `game-backend`
- `game-collector`
- `admin`

That script writes release artifacts into the root `build/` directory. The Nuxt frontend keeps its own Docker-based production deployment flow under `apps/cn/nav-web`.

## Production Deployment

There are now two main deployment paths in this repository.

The Nuxt frontend ships with its own Docker deployment path. See:

- [apps/cn/nav-web/DEPLOYMENT.md](./apps/cn/nav-web/DEPLOYMENT.md)
- [apps/cn/nav-web/update.sh](./apps/cn/nav-web/update.sh)
- [Cross-service deployment overview](./docs/deployment.md)

Typical production update flow:

```bash
cd apps/cn/nav-web
./update.sh
```

Go services keep their own binary / install workflows. Modules under `legacy/` are not part of the default build or production deployment path.

## Current Status

- The public site frontend has been migrated to Nuxt 4 and is already running in production
- `gofurry-nav-backend` now serves the main public flow through `/api/v2/nav`, and the old `nav/page/*` live routes are no longer part of the active runtime path
- `gofurry-nav-collector` has completed its v2 data-plane work and now provides summary, latest, observations, trend, change-event, and low-frequency side-channel probe outputs
- The former `archive` free-form Q&A page and site-facing RAG integration have been decommissioned; the frontend entry is now the `/steam` Steam Zone
- Former Ops Agent / Center code is archive-only under `legacy/` and is excluded from active build, CI, and deployment tooling
- The updates page has been rebuilt as a structured bilingual timeline backed by `gfn_nav_update_notice`, without relying on CDN-hosted markdown
- Search suggestions are now unified behind the v2 suggestions API with cache, singleflight deduplication, proxy support, and baseline rate limiting
- `robots.txt`, `sitemap.xml`, `llms.txt`, and `/.well-known/security.txt` are available as public site metadata entrypoints
- The old Vue frontend and former RAG service are archived under `legacy/`
- The root `build.bat` builds only the five active Go applications

## Contributing

Issues and pull requests are welcome.

When contributing:

- keep changes scoped to the relevant service whenever possible
- do not commit `.env` files, private keys, database credentials, or other secrets
- update docs or deployment notes when public behavior changes
- preserve existing service boundaries unless cross-service changes are genuinely required

## License

This repository is released under the [MIT License](./LICENSE).
