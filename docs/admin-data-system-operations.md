# Admin Data and System Operations

`/collection`, `/metrics`, `/changes`, `/system/cloud`, `/system/data-operations`, `/system/audit`, and `/system/accounts` are native React routes in the sole production Admin frontend.

Cloud Resources uses `/api/v1/system/cloud`: `cloudops.read` covers storage overview, object inspection, and EdgeOne task history; `cloudops.manage` covers COS-to-R2 repair and scoped EdgeOne/Cloudflare purges. Owner and Developer have both; Operator has neither. Only Owner has `cloudops.purge_all` for the separate full-zone EdgeOne endpoint. Its initially collapsed UI requires explicit scope confirmation. Cloud mutations are audited; UI gates consume capabilities, never reconstruct the role policy. See [Managed assets](managed-assets.md#cloudops).

Collection, Metric, and Change pages consume the existing frozen control-plane and projection APIs. Run Now remains a manual job with schedule lineage and no scheduled slot. Zero expected targets render coverage as unavailable. Metric `unknown` remains distinct from `negative`, and technical Registry/checkpoint views require their technical capabilities.

`/api/v1/dataops/overview` requires `dataops.read` and performs bounded read-only metadata queries against the explicit `gfa`, `gfn`, and `gfg` pools. It reports health, PostgreSQL/database metadata, connection counts, the latest expected repository Goose version, pending state, and the ten largest public relations. Expected versions are compiled from the repository migration set and regression-tested. The endpoint never returns configuration or connection strings and exposes no mutation operation.

`/api/v1/audit/logs` requires `audit.read`, supports operator/role/action/resource/time filters and count-backed pagination, and interprets history using `operator_name` and `operator_role` snapshots. Before/after JSON is recursively redacted for password, hash, token, cookie, secret, DSN, and connection-string shaped keys before it leaves the backend.

`/api/v1/workbench/summary` requires `content.read` and includes only projections allowed by the current Principal's other capabilities. It combines existing collector failures/misses/health, coverage and pipeline lag, recent canonical changes, recent audit operations, database health, and account summaries. It stores no alert state and introduces no notification subsystem.

Accounts reuse `/api/v1/auth/accounts`. Only `account.manage` can create accounts, change display names/roles/status, reset passwords, or revoke sessions. The backend transaction remains authoritative for the last-active-Owner invariant.

Authenticated self-service under `/api/v1/auth/self/*` separately allows username/password changes after current-password verification. Username changes preserve sessions and refresh identity; password changes invalidate prior sessions, clear the cookie, and require login again. Both are audited without granting `account.manage`.
