# Engineering playbook

Run commands from the named module unless stated otherwise.

## Active Go checks

For each of `apps/cn/game-collector`, `apps/cn/game-backend`, `apps/cn/nav-collector`, `apps/cn/nav-backend`, and `apps/cn/admin`:

~~~text
gofmt -w .
go vet ./...
go test ./...
go build ./...
go run . --help
go run . serve --help
go run . install --help
go run . uninstall --help
go run . version
~~~

Do not run `install` or `uninstall` against a real host as a routine test. Unit rendering and non-Linux behavior are covered in `internal/systemd` tests.

Frontend checks:

~~~text
cd apps/cn/admin/web && npm ci && npm run build
cd apps/cn/nav-web && npm ci && npm run typecheck && npm run build
~~~

## Repository and database checks

From `tools`:

~~~text
go test ./...
go tool sqlc vet -f ../sqlc.yaml
go tool sqlc generate -f ../sqlc.yaml
go run ./check-sqlc
go run ./check-production-policy
~~~

Fresh PostgreSQL, exact baseline adoption, drift rejection, cleanup upgrade, seeded row-count, and backup-copy checks:

~~~text
GOFURRY_TEST_POSTGRES_ADMIN_URL='postgres://.../postgres' go test ./db-baseline -run TestPostgresFreshAndBaselineAdoption -count=1 -v
~~~

Local per-service integrations use isolated development configs:

~~~text
GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG=/path/config.yaml go test ./collector/game/v2/repository -run TestPostgresRepositorySemantics -count=1
GOFURRY_GAME_BACKEND_INTEGRATION_CONFIG=/path/config.yaml go test ./apps/game/v2/dao -run TestPostgresReadModelSemantics -count=1
GOFURRY_NAV_COLLECTOR_INTEGRATION_CONFIG=/path/config.yaml go test ./collector/observation -run TestPostgresCollectorPersistenceSemantics -count=1
GOFURRY_NAV_BACKEND_INTEGRATION_CONFIG=/path/config.yaml go test ./apps/nav/navPage/dao -run TestPostgresNavBackendPersistenceSemantics -count=1
GOFURRY_ADMIN_INTEGRATION_CONFIG=/path/server.yaml go test ./internal/bootstrap -run TestAdminThreeDatabasePersistence -count=1
~~~

Run these only against explicitly isolated development PostgreSQL. Never use production credentials.

Run the pinned vulnerability scanner from `tools` for every active module. CI performs the same active-only matrix on dependency changes and on a schedule.

The CI dependency matrix remains: `db/game` runs both Game services and Admin; `db/nav` runs both Nav services and Admin; `db/admin` runs Admin; sqlc/tools changes run all SQL consumers. Archive and experiment trees are never production CI inputs.
