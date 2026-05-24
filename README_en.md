<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white" alt="Go Version">&nbsp&nbsp
  <img src="https://img.shields.io/github/license/gofurry/gofurry-nav-site" alt="License">&nbsp&nbsp
  <img src="https://img.shields.io/badge/weekend-project-8B5CF6?style=flat" alt="Weekend Project">&nbsp&nbsp
  <img src="https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange" alt="Made with Love">
</p>

<p align="center">
  ⭐🐺 <a href="https://github.com/gofurry/gofurry-nav-site/README.md">中文说明</a> 🐺⭐
</P>

gofurry is an open-source multi-service repository for furry culture discovery. It contains the public web frontend, navigation and game APIs, data collectors, a RAG service, and operations-related modules.

The production public site has now moved to a Nuxt 4 frontend to improve SEO and first-page rendering. The older Vue frontend is kept in the repository as an archived migration reference rather than the active production entrypoint.

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

- `gofurry-nav-web`: Nuxt 4 frontend, now used by the production public site
- `gofurry-nav-frontend-legacy`: archived Vue frontend kept for migration reference
- `gofurry-nav-backend`: navigation API service
- `gofurry-nav-collector`: navigation data collector
- `gofurry-game-backend`: game-related API service
- `gofurry-game-collector`: game-related data collector
- `gofurry-rag`: lightweight RAG service with an embedded admin console, public Q&A endpoints, and content sync
- `gofurry-admin`: operations backend with embedded frontend
- `gofurry-intl`: internationalization-related service and frontend experiments
- `ops`: lightweight operations probes and monitoring center, including `gofurry-ops-agent`, `gofurry-ops-center`, the audit template, and the roadmap
- `sql`: database scripts and schema-related files
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
cd gofurry-nav-web
npm install
npm run dev
```

Go service development:

```bash
cd gofurry-nav-backend
go run .
```

RAG service development:

```bash
cd gofurry-rag
go run . --config ./config/server.yaml serve
```

Ops local checks:

```bash
cd ops/gofurry-ops-center
go run ./cmd/center check-config --config ./configs/center.example.yaml

cd ../gofurry-ops-agent
go run ./cmd/agent --config ./configs/agent.example.yaml check-config
```

If you need root-level packaging artifacts, use:

```bat
build.bat all
```

Current targets include:

- `gofurry-nav-backend`
- `gofurry-nav-collector`
- `gofurry-game-backend`
- `gofurry-game-collector`
- `gofurry-admin`
- `gofurry-rag`
- `gofurry-ops-agent`
- `gofurry-ops-center`

That script writes traditional release artifacts into the root `build/` directory. The Nuxt frontend still uses its own Docker-based production deployment flow from `gofurry-nav-web`.

## Production Deployment

There are now two main deployment paths in this repository.

The Nuxt frontend ships with its own Docker deployment path. See:

- [gofurry-nav-web/DEPLOYMENT.md](./gofurry-nav-web/DEPLOYMENT.md)
- [gofurry-nav-web/update.sh](./gofurry-nav-web/update.sh)

Typical production update flow:

```bash
cd gofurry-nav-web
./update.sh
```

Go services and the RAG module keep their own binary / install workflows. For the current RAG deployment path, see:

- [gofurry-rag/README.md](./gofurry-rag/README.md)
- [gofurry-rag/docs/deployment.md](./gofurry-rag/docs/deployment.md)

Ops Agent / Center include lightweight systemd, Docker Compose, and Nginx examples, and are intended to run with a dedicated PostgreSQL schema or database:

- [ops/gofurry-ops-agent/README.md](./ops/gofurry-ops-agent/README.md)
- [ops/gofurry-ops-center/README.md](./ops/gofurry-ops-center/README.md)
- [ops/roadmap.md](./ops/roadmap.md)

## Current Status

- The public site frontend has been migrated to Nuxt 4 and is already running in production
- The `archive` free-form Q&A page and site-facing RAG integration are now part of the live stack
- `gofurry-rag` now includes public Q&A guardrails, lightweight multi-turn context, and the first production content sync framework
- `ops/gofurry-ops-agent` and `ops/gofurry-ops-center` now provide a lightweight monitoring loop for node samples, service health, alert state, peer summaries, sync/deploy events, and the embedded console
- `ops/code-audit.md` is kept as the next Go code audit template, and `ops/roadmap.md` records the future Ops evolution path
- `robots.txt` and `sitemap.xml` are generated by Nuxt server routes
- The old Vue frontend is archived as a legacy reference
- The root `build.bat` now covers the main Go services, admin backend, RAG, and Ops Agent / Center build artifacts

## Contributing

Issues and pull requests are welcome.

When contributing:

- keep changes scoped to the relevant service whenever possible
- do not commit `.env` files, private keys, database credentials, or other secrets
- update docs or deployment notes when public behavior changes
- preserve existing service boundaries unless cross-service changes are genuinely required

## License

This repository is released under the [MIT License](./LICENSE).
