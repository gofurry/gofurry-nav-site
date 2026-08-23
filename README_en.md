<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26.7-00ADD8?logo=go&logoColor=white" alt="Go Version">&nbsp;&nbsp;
  <img src="https://img.shields.io/github/license/gofurry/gofurry-nav-site" alt="License">&nbsp;&nbsp;
  <img src="https://img.shields.io/badge/weekend-project-8B5CF6?style=flat" alt="Weekend Project">&nbsp;&nbsp;
  <img src="https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange" alt="Made with Love">
</p>

<p align="center">
  ⭐🐺 <a href="./README.md">中文说明</a> |
  <a href="https://go-furry.com">GoFurry China</a> |
  <a href="https://gofurry.com">GoFurry International</a> 🐺⭐
</p>

GoFurry is an open-source, multi-service repository for furry culture discovery, site navigation, game information, and availability observation. Active production code lives under `apps/cn`, and each service is developed and deployed independently.

## Project Scope

- `apps/cn/nav-web`: public Nuxt frontend
- `apps/cn/nav-backend`: navigation API
- `apps/cn/nav-collector`: navigation data collection
- `apps/cn/game-backend`: game API
- `apps/cn/game-collector`: Steam/game data collection
- `apps/cn/admin`: management service with an embedded frontend
- `apps/cn/uptime`: independent availability observer backed by local Bbolt

`apps/intl` is placeholder-only. `legacy`, `experimental`, and `third-party` are outside the active build, CI, and production deployment graph. `db/game`, `db/nav`, and `db/admin` own Goose migrations for `gfg`, `gfn`, and `gfa`.

## Stack

- Go / Fiber
- PostgreSQL / Redis
- Nuxt 4 / Vue 3
- Tailwind CSS / Less
- Coraza WAF
- Bbolt

## Quick Start

Frontend development:

```bash
cd apps/cn/nav-web
npm install
npm run dev
```

Go service development:

```bash
cd apps/cn/nav-backend
cp conf/server.example.yaml conf/server.yaml
# Edit the local configuration, then:
go run . serve --config conf/server.yaml
```

The root command of every Go application only displays help. Running a service requires explicit `serve --config <file>`. Copy example configuration files for local use and never commit real keys or credentials.

## Build and Validation

```bat
build.bat all
```

The script builds only the six active Go applications and writes artifacts to the root `build/` directory. Nav Web retains its separate Node/Docker workflow. See [Local development](./docs/development.md) and the [Agent playbook](./.agents/playbook.md) for the complete validation commands.

## Deployment and Operations

Nav Web follows the Docker workflow in its [deployment guide](./apps/cn/nav-web/DEPLOYMENT.md). The six Go binaries provide built-in Linux/systemd `install --config <file>` and `uninstall` commands; installation enables the unit but never starts it.

Database migration and binary deployment are separate operator actions. Applications never run Goose at startup. See:

- [Cross-service deployment](./docs/deployment.md)
- [Linux/systemd operations](./docs/operations/systemd.md)
- [Database contract](./contracts/database.md)
- [Availability contract](./contracts/availability.md)

## Contributing

Issues and pull requests are welcome. Keep changes within the owning service and follow the [collaboration guide](./AGENTS.md), [compatibility contract](./contracts/compatibility.md), and [accepted architecture decisions](./docs/decisions/README.md).

## License

This repository is released under the [MIT License](./LICENSE).
