# Local development

## Prerequisites

- Go 1.26.7 for all five active Go modules and `tools`
- Node.js 24 and npm for `apps/cn/nav-web` and `apps/cn/admin/web`
- development PostgreSQL and Redis matching the selected app configuration
- sqlc and govulncheck through the pinned Go tools in `tools/go.mod`

The repository intentionally has no `go.work` and retains npm rather than pnpm. Run Go commands from the owning module.

## Configuration

Every runtime command requires an explicit YAML file:

| App | Example config |
|---|---|
| Nav Backend | `apps/cn/nav-backend/conf/server.example.yaml` |
| Nav Collector | `apps/cn/nav-collector/conf/server.example.yaml` |
| Game Backend | `apps/cn/game-backend/conf/server.example.yaml` |
| Game Collector | `apps/cn/game-collector/conf/server.example.yaml` |
| Admin | `apps/cn/admin/config/server.example.yaml` |

Copy the example to an ignored local file, replace placeholders, and never commit credentials.

Example:

~~~bash
cd apps/cn/nav-backend
go run . serve --config conf/server.yaml
~~~

The root command is intentionally non-starting:

~~~bash
go run . --help
go run . version
~~~

Game Collector one-shot operations use the same explicit configuration:

~~~bash
cd apps/cn/game-collector
go run . collect --config conf/server.yaml
go run . players --config conf/server.yaml
go run . all --config conf/server.yaml
~~~

## Normal validation

For each of `game-collector`, `game-backend`, `nav-collector`, `nav-backend`, and `admin` under `apps/cn`:

~~~bash
gofmt -w .
go vet ./...
go test ./...
go build ./...
~~~

Frontends:

~~~bash
cd apps/cn/admin/web
npm ci
npm run build

cd ../../nav-web
npm ci
npm run typecheck
npm run build
~~~

Repository tooling from `tools`:

~~~bash
go test ./...
go tool sqlc vet -f ../sqlc.yaml
go tool sqlc generate -f ../sqlc.yaml
go run ./check-sqlc
go run ./check-production-policy
~~~

Run vulnerability scans for each active module using the pinned tool:

~~~bash
cd tools
go tool govulncheck -C ../apps/cn/nav-backend ./...
~~~

Database integration checks require explicitly configured, isolated development databases. Never point them at production.
