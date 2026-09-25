# GoFurry React Admin

This is the sole React 19 Admin frontend. Vite is used for local development and writes the production build directly to the Go embed directory.

Use Node 24 and pnpm 12.6.0, pinned by `packageManager`. This project owns its
lockfile and local script permissions; it is independent of Nav Web and has no
root workspace. Root `task deps:admin`, `task dev:admin-web`, `task test:admin`
and `task build:admin` provide the corresponding engineering operations.

~~~text
pnpm install --frozen-lockfile
pnpm run dev
pnpm run typecheck
pnpm test
pnpm run build
~~~

`pnpm run build` clears and prepares `../internal/transport/http/webui/dist`; run it before Go tests or builds that compile the embedded frontend from a clean checkout.

The development server listens on `127.0.0.1:5178` and proxies `/api` and `/csrf` to the Go Admin API on `127.0.0.1:10099`.

Architecture and migration boundaries are defined in `../../../../contracts/admin-frontend.md` and `../../../../docs/admin-react.md`.
