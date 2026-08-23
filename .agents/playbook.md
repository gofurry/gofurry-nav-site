# Engineering playbook

Run commands from the named module unless stated otherwise.

## Normal checks

For each of `apps/cn/game-collector`, `apps/cn/game-backend`, `apps/cn/nav-collector`, `apps/cn/nav-backend`, and `apps/cn/admin`:

```text
gofmt -w .
go vet ./...
go test ./...
go build ./...
```

Frontend checks:

```text
cd apps/cn/admin/web && npm ci && npm run build
cd apps/cn/nav-web && npm ci && npm run typecheck && npm run build
```

## Database contract checks

From `tools`:

```text
go tool sqlc vet -f ../sqlc.yaml
go tool sqlc generate -f ../sqlc.yaml
go run ./check-sqlc
go run ./check-production-policy
```

Fresh PostgreSQL, exact baseline adoption, drift rejection, cleanup upgrade, seeded row-count, and backup-copy checks:

```text
GOFURRY_TEST_POSTGRES_ADMIN_URL='postgres://.../postgres' go test ./db-baseline -run TestPostgresFreshAndBaselineAdoption -count=1 -v
```

Local per-service integrations use an isolated development PostgreSQL config path:

```text
GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG=/path/config.yaml go test ./collector/game/v2/repository -run TestPostgresRepositorySemantics -count=1
GOFURRY_GAME_BACKEND_INTEGRATION_CONFIG=/path/config.yaml go test ./apps/game/v2/dao -run TestPostgresReadModelSemantics -count=1
GOFURRY_NAV_COLLECTOR_INTEGRATION_CONFIG=/path/config.yaml go test ./collector/observation -run TestPostgresCollectorPersistenceSemantics -count=1
GOFURRY_NAV_BACKEND_INTEGRATION_CONFIG=/path/config.yaml go test ./apps/nav/navPage/dao -run TestPostgresNavBackendPersistenceSemantics -count=1
GOFURRY_ADMIN_INTEGRATION_CONFIG=/path/server.yaml go test ./internal/bootstrap -run TestAdminThreeDatabasePersistence -count=1
```

CI dependency matrix: `db/game` runs Game DB plus both Game services and Admin; `db/nav` runs Nav DB plus both Nav services and Admin; `db/admin` runs Admin DB and Admin; `sqlc.yaml`/`tools` run all SQL consumers; an `apps/cn` service path runs its Go and relevant PostgreSQL integration checks; Admin also builds its web UI; `apps/cn/nav-web` typechecks and builds. Archive and experiment trees are not part of production CI.
