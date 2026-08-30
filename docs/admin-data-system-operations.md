# Admin Data and System Operations

P0.5.2-C makes `/collection`, `/metrics`, `/changes`, `/system/data-operations`, `/system/audit`, and `/system/accounts` native React routes. The embedded Vue build remains the production fallback until P0.5.2-D; this stage does not cut over production routing.

Collection, Metric, and Change pages consume the existing frozen control-plane and projection APIs. Run Now remains a manual job with schedule lineage and no scheduled slot. Zero expected targets render coverage as unavailable. Metric `unknown` remains distinct from `negative`, and technical Registry/checkpoint views require their technical capabilities.

`/api/v1/dataops/overview` requires `dataops.read` and performs bounded read-only metadata queries against the explicit `gfa`, `gfn`, and `gfg` pools. It reports health, PostgreSQL/database metadata, connection counts, the latest expected repository Goose version, pending state, and the ten largest public relations. Expected versions are compiled from the repository migration set and regression-tested. The endpoint never returns configuration or connection strings and exposes no mutation operation.

`/api/v1/audit/logs` requires `audit.read`, supports operator/role/action/resource/time filters and count-backed pagination, and interprets history using `operator_name` and `operator_role` snapshots. Before/after JSON is recursively redacted for password, hash, token, cookie, secret, DSN, and connection-string shaped keys before it leaves the backend.

`/api/v1/workbench/summary` requires `content.read` and includes only projections allowed by the current Principal's other capabilities. It combines existing collector failures/misses/health, coverage and pipeline lag, recent canonical changes, recent audit operations, database health, and account summaries. It stores no alert state and introduces no notification subsystem.

Accounts reuse `/api/v1/auth/accounts`. Only `account.manage` can create accounts, change display names/roles/status, reset passwords, or revoke sessions. The backend transaction remains authoritative for the last-active-Owner invariant.
