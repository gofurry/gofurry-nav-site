# React Admin development

The sole Admin frontend is `apps/cn/admin/react`. It owns content workspaces and all operational/system workflows and is the production entrypoint embedded by the Go binary.

Start the existing Go API with the ignored local development config:

~~~text
cd apps/cn/admin
go run . serve --config config/server.yaml
~~~

Then start Vite in another terminal:

~~~text
cd apps/cn/admin/react
pnpm install --frozen-lockfile
pnpm run dev
~~~

Open `http://127.0.0.1:5178`. Vite proxies `/api` and `/csrf` to `http://127.0.0.1:10099`.

Validation:

~~~text
pnpm run typecheck
pnpm test
pnpm run build
~~~

`pnpm run build` clears and writes `apps/cn/admin/internal/transport/http/webui/dist`. The root `task build:admin` target performs this React build before compiling the Go binary and copying the deployment `dist/` companion artifact. No manual asset copy or runtime Node process is used.

The App Shell consumes the current principal from `/api/v1/auth/state`. Missing navigation or actions should first be checked against returned capabilities and backend authorization; never patch around the contract with role comparisons.

Simple resources are defined in `src/features/resources/definitions.tsx`. Site and Game must remain dedicated workspaces. New server reads should be small, explicit sqlc-backed read models rather than a generic frontend BFF.

Collection, Metrics, and Changes are under `src/features/operations`; Cloud Resources, DataOps, Audit, and Accounts are under `src/features/system`. `dataops.read` is the only valid Data Operations capability. Operator/Developer/Owner differences must be expressed through `auth.can(...)`, not client-side role matrices.

`/system/cloud` uses shared Admin controls for storage summaries, object inspection, scoped purges, and task history. `cloudops.read` permits reads; `cloudops.manage` permits mirror repair and scoped EdgeOne/Cloudflare purges; `cloudops.purge_all` permits the separate, initially collapsed full-zone EdgeOne action. The backend grants the first two to Owner/Developer, the last only to Owner, and none to Operator.

Authenticated self-service username/password actions use `/api/v1/auth/self/*` with current-password verification and no `account.manage` requirement. Username changes refresh identity without ending the session; password changes clear authentication and require login again.

See [the cutover parity matrix](admin-frontend-parity.md) and [the role operator guide](operations/admin-roles.md) for production acceptance boundaries.

`src/features/collaboration` owns `/collaboration` (ideas/board), visible pipe-delimited line parsing, version conflicts and shared idea context. Creation pages prefill only Steam AppID or Site name; never auto-fetch Steam or create targets. Link failure preserves successful creation and the `?idea=` recovery banner. Use Vitest/Testing Library for these flows; see [Collaboration Center](collaboration-center.md).
