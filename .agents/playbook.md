# Engineering playbook

Run commands from the named module unless stated otherwise.

## Task lifecycle

### Before editing

- Define the requested scope and explicit exclusions.
- Inspect the current implementation, tests, configuration, contracts, and accepted ADRs before proposing changes.
- Identify the owning source of truth and report any conflict that cannot be resolved from repository evidence.

### During editing

- Keep changes within scope and avoid unrelated refactors.
- Modify the owning source of truth rather than documenting around stale behavior.
- Preserve public behavior and operational contracts unless the task explicitly changes them.

### Before finishing

- Run validation appropriate to every affected area.
- Inspect the final diff for scope, generated artifacts, and secrets.
- Report what was changed, what passed, and anything not verified or still unresolved.

## Active Go checks

For each of `apps/cn/game-collector`, `apps/cn/game-backend`, `apps/cn/nav-collector`, `apps/cn/nav-backend`, `apps/cn/admin`, and `apps/cn/uptime`:

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
cd apps/cn/admin/react && npm ci && npm run typecheck && npm test && npm run build
cd apps/cn/admin/web && npm ci && npm run build
cd apps/cn/nav-web && npm ci && npm run typecheck && npm run build
~~~

For local React Admin work, run the Go Admin API on `127.0.0.1:10099` and `npm run dev` from `apps/cn/admin/react`; Vite proxies `/api` and `/csrf`. The Vue frontend remains the production embed during coexistence.

## Repository and database checks

From `tools`:

~~~text
go test ./...
go tool sqlc vet -f ../sqlc.yaml
go tool sqlc generate -f ../sqlc.yaml
go run ./check-sqlc
go run ./check-production-policy
~~~

Fresh PostgreSQL, schema adoption, drift rejection, upgrade, seeded row-count, and backup-copy checks:

~~~text
GOFURRY_TEST_POSTGRES_ADMIN_URL='postgres://.../postgres' go test ./db-baseline -run TestPostgresFreshAndBaselineAdoption -count=1 -v
~~~

Local per-service integrations use isolated development configs:

~~~text
GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG=/path/config.yaml go test ./collector/game/v2/repository -run TestPostgresRepositorySemantics -count=1
GOFURRY_GAME_BACKEND_INTEGRATION_CONFIG=/path/config.yaml go test ./apps/game/v2/dao -run TestPostgresReadModelSemantics -count=1
GOFURRY_NAV_COLLECTOR_INTEGRATION_CONFIG=/path/config.yaml go test ./collector/observation -run TestPostgresCollectorPersistenceSemantics -count=1
GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG=/path/config.yaml go test ./collector/changes -count=1
GOFURRY_NAV_COLLECTOR_INTEGRATION_CONFIG=/path/config.yaml go test ./collector/changes -count=1
GOFURRY_NAV_BACKEND_INTEGRATION_CONFIG=/path/config.yaml go test ./apps/nav/navPage/dao -run TestPostgresNavBackendPersistenceSemantics -count=1
GOFURRY_ADMIN_INTEGRATION_CONFIG=/path/server.yaml go test ./internal/bootstrap -run 'TestAdmin(ThreeDatabasePersistence|IdentityAuthorizationPersistence|LegacyIdentityUpgrade)' -count=1
~~~

Historical Fact smoke checks use the Collector CLI against the same isolated config: `facts status`, `facts backfill --dry-run`, `facts backfill`, and a bounded `facts rebuild --pipeline ... --from ... --through ...`. Keep `facts.retention_enabled=false` until checkpoints and fact row counts are verified.

Analytics Metric smoke checks run after Fact watermarks are ready: `metrics status`, `metrics backfill --dry-run`, a bounded `metrics backfill --metric ... --version ...`, and `metrics rebuild --metric ... --version ... --from ... --through ...`. Verify Registry/evaluator drift, per-version checkpoints, the all-zero global row, count conservation, historical-day freshness, and same-day Historical Fact names in Admin Metric Center.

For Nav capability semantics, verify that an AAAA subquery failure never becomes a negative IPv6 state and that HTML/empty/malformed security.txt responses never become positive. Published v1 remains rebuildable; active v2 begins at its Goose-owned source-start cutoff and must not be backdated over evidence that cannot be reconstructed.

Collection Center smoke must verify Run Now schedule lineage without `scheduled_for` or phase movement, nullable coverage when `expected_count=0`, Current/Historical Collector lifecycle views, count-backed Run/Result pagination, and searchable Game/Site/Target manual selection.

React operational smoke must verify Operator sees Collection/Metrics/Changes but no schedule controls or technical tabs; Developer additionally sees controls, technical contracts, Data Operations, and Audit; Owner additionally sees Accounts. Data Operations must report `gfa`/`gfn`/`gfg` without exposing DSNs or mutation controls. Audit pagination must preserve historical identity snapshots and redact secret-shaped fields. Workbench must omit capability-ineligible projections.

Admin identity smoke must verify zero-account bootstrap, username/password login, current Principal/capabilities, role/status/password/session revocation, account-management authorization, and transaction-safe last-active-Owner protection. Run the PostgreSQL identity tests explicitly; a skipped integration test is not evidence.

Change Intelligence smoke checks run after Metric/Fact watermarks are ready: `changes status`, `changes backfill --dry-run`, a bounded `changes backfill --detector ... --version ...`, and `changes rebuild --detector ... --version ... --from ... --dry-run` followed by the actual rebuild when intended. Rebuild must propagate through `processed_through`; an optional `--through` must equal that checkpoint and `--max-days` must not truncate it. Verify stable event keys, semantic-memory gaps, tracking identity resets, event time/provenance, and same-day Historical Fact names in Admin Change Center.

Run these only against explicitly isolated development PostgreSQL. Never use production credentials.

Run the pinned vulnerability scanner from `tools` for every active module. CI performs the same active-only matrix on dependency changes and on a schedule.

Availability smoke checks include Nav Web `/healthz`, backend/Admin readiness, both enabled collector readiness listeners, and the standalone uptime service `/livez`, `/readyz`, and `/uptime`. Uptime does not participate in sqlc, Goose, or PostgreSQL integration jobs.

The CI dependency matrix remains: `db/game` runs both Game services and Admin; `db/nav` runs both Nav services and Admin; `db/admin` runs Admin; sqlc/tools changes run all SQL consumers. Archive and experiment trees are never production CI inputs.
