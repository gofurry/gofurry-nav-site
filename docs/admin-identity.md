# Admin identity and authorization

V3-P0.5.2-A replaces the historical singleton password with a small internal multi-account identity system. Goose owns the schema, sqlc owns normal queries, and the Go application owns the fixed authorization policy.

## Authentication and bootstrap

`POST /api/v1/auth/bootstrap` accepts canonical `username`, human-facing `display_name`, and `password` only while `gfa_admin_account` is empty. It creates an active Owner. Once any account exists, including a disabled one, bootstrap returns conflict.

Login accepts username and password. The cookie JWT contains `account_id`, `session_version`, `iss`, `sub=admin:<id>`, `iat`, `nbf`, and `exp`. Each authenticated request loads that account from `gfa`, verifies active status and session version, and stores one Principal in Fiber locals. Role and capabilities are resolved from current database state and compiled policy, not token claims or a cache.

The alpha.5 singleton migration accepts zero or one legacy account. One row becomes `owner` / `Owner` / `owner` / `active` without changing its password hash, session version, or timestamps. More than one legacy row aborts migration. Historical audit rows are attributed to the legacy Owner when present; otherwise their legacy operator text is retained as a system snapshot.

## Roles and capabilities

| Capability | Owner | Developer | Operator |
|---|:---:|:---:|:---:|
| `content.read` | yes | yes | yes |
| `content.write` | yes | yes | yes |
| `collection.read` | yes | yes | yes |
| `collection.execute` | yes | yes | yes |
| `collection.control` | yes | yes | no |
| `metrics.read` | yes | yes | yes |
| `metrics.technical` | yes | yes | no |
| `changes.read` | yes | yes | yes |
| `changes.technical` | yes | yes | no |
| `dataops.read` | yes | yes | no |
| `audit.read` | yes | yes | no |
| `account.manage` | yes | no | no |
| `system.manage` | yes | no | no |

Unknown roles receive no capabilities and unknown capabilities are denied.

## Route authorization

| Route family | Read/business | Mutation/technical |
|---|---|---|
| Nav, Game, and option APIs | `content.read` | `content.write` for every POST/PUT/DELETE and bulk replacement |
| Collection overview, instances, schedules, jobs, runs, results, charts | `collection.read` | Run Now/manual/retry: `collection.execute`; schedule edit/enable/disable and cancel: `collection.control` |
| Metric overview, daily, entities | `metrics.read` | Registry and checkpoints: `metrics.technical` |
| Change overview and events | `changes.read` | Registry and checkpoints: `changes.technical` |
| Workbench aggregation | `content.read`, shaped by the Principal's other capabilities | — |
| Data Operations metadata | `dataops.read` | read-only; no mutation endpoints |
| Audit history and details | `audit.read` | read-only, snapshot identity and secret redaction |
| Account list/create/display name/role/status/password/revoke | — | `account.manage` |

Account endpoints live under `/api/v1/auth/accounts`. There is no hard-delete endpoint. Username is immutable; display-name-only updates do not revoke sessions. Role, status, password, and explicit revoke operations increment session version and are audited without secret material.

The last active Owner invariant is enforced inside the account transaction by locking the active Owner set before role/status mutation. With two active Owners one can be demoted or disabled; concurrent requests cannot remove both.
