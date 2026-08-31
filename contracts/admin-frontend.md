# Admin frontend contract

`apps/cn/admin/react` is the sole development and production Admin frontend. Its frozen foundation is React 19, Vite, strict TypeScript, React Router, Tailwind CSS v4, shadcn-style wrappers over Base UI primitives, TanStack Query/Table, React Hook Form, Zod, ECharts, and Lucide.

## Architecture boundaries

UI layers flow in one direction:

~~~text
components/ui -> components/admin -> features
~~~

TanStack Query owns server state. Route filters, search, pagination, sorting, and workspace tabs use URL state where practical. React local state owns transient UI. Do not introduce a second server-state copy or a role-based client store.

Themes are `system`, `light`, and `dark`, defaulting to `system`. Business components consume semantic tokens: background, surface, surface-muted, foreground, muted-foreground, border, primary, success, warning, danger, and info.

The backend authorization contract remains authoritative. Navigation and actions ask whether the current principal has a capability; frontend code must not reproduce the Role-to-Capability mapping.

## Product structure

Top-level groups are Workbench, Nav Content, Game Content, Data Operations, and System. Content and operational routes are native React:

~~~text
/nav/sites
/nav/sites/:id
/nav/site-groups
/nav/update-notices
/nav/sayings
/game/games
/game/games/:id
/game/tags
/game/comments
/game/prizes
/collection
/metrics
/changes
/system/data-operations
/system/audit
/system/accounts
~~~

Site and Game are dedicated workspaces. Simple resources use the typed Resource Engine. Persistence mapping tables are managed as relationships inside workspaces, not exposed as primary navigation.

Collection, Metric, and Change reuse their existing business APIs and frozen Fact/Metric/Detector/Collection semantics. Operator-facing views require their read capabilities; schedule control, Metric technical contracts, and Change technical contracts additionally require their native capabilities.

Data Operations is a read-only `dataops.read` health center for the explicit `gfa`, `gfn`, and `gfg` pools. It may expose bounded PostgreSQL metadata, Goose state, and Top N relation sizes, but never DSNs, credentials, arbitrary SQL, migrations, rollback, maintenance, connection termination, or configuration mutation. The spelling `data_ops.read` is not an alias.

Audit requires `audit.read`, uses historical identity snapshots, returns real pagination, and redacts secret-shaped fields before responses. Accounts requires `account.manage`, keeps fixed roles, and delegates last-active-Owner protection to the backend transaction. Workbench aggregation is capability-shaped and projects existing durable attention sources; it is not an alerting subsystem.

The React production build is the only writer of `apps/cn/admin/internal/transport/http/webui/dist`. It clears stale output before building, and the Go binary embeds that directory. Production must not require Node, a Vite server, a separate frontend service, manual asset copying, or a second Admin frontend implementation. The completed parity audit is recorded in `docs/admin-frontend-parity.md`.
