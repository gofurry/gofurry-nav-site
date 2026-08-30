# React Admin development

The target Admin frontend is `apps/cn/admin/react`. It owns content workspaces and the operational/system workflows completed through P0.5.2-C. It still coexists with the embedded Vue frontend and is not yet the production entrypoint.

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

The App Shell consumes the current principal from `/api/v1/auth/state`. Missing navigation or actions should first be checked against returned capabilities and backend authorization; never patch around the contract with role comparisons.

Simple resources are defined in `src/features/resources/definitions.tsx`. Site and Game must remain dedicated workspaces. New server reads should be small, explicit sqlc-backed read models rather than a generic frontend BFF.

Collection, Metrics, and Changes are under `src/features/operations`; DataOps, Audit, and Accounts are under `src/features/system`. `dataops.read` is the only valid Data Operations capability. Operator/Developer/Owner differences must be expressed through `auth.can(...)`, not client-side role matrices.
