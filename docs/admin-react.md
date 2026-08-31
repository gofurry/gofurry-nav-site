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
npm ci
npm run dev
~~~

Open `http://127.0.0.1:5178`. Vite proxies `/api` and `/csrf` to `http://127.0.0.1:10099`.

Validation:

~~~text
npm run typecheck
npm test
npm run build
~~~

`npm run build` clears and writes `apps/cn/admin/internal/transport/http/webui/dist`. The root `build.bat admin` target performs this React build before compiling the Go binary and copying the deployment `dist/` companion artifact. No manual asset copy or runtime Node process is used.

The App Shell consumes the current principal from `/api/v1/auth/state`. Missing navigation or actions should first be checked against returned capabilities and backend authorization; never patch around the contract with role comparisons.

Simple resources are defined in `src/features/resources/definitions.tsx`. Site and Game must remain dedicated workspaces. New server reads should be small, explicit sqlc-backed read models rather than a generic frontend BFF.

Collection, Metrics, and Changes are under `src/features/operations`; DataOps, Audit, and Accounts are under `src/features/system`. `dataops.read` is the only valid Data Operations capability. Operator/Developer/Owner differences must be expressed through `auth.can(...)`, not client-side role matrices.

See [the cutover parity matrix](admin-frontend-parity.md) and [the role operator guide](operations/admin-roles.md) for production acceptance boundaries.
