# GoFurry React Admin

This is the sole React 19 Admin frontend. Vite is used for local development and writes the production build directly to the Go embed directory.

~~~text
npm ci
npm run dev
npm run typecheck
npm test
npm run build
~~~

`npm run build` clears and prepares `../internal/transport/http/webui/dist`; run it before Go tests or builds that compile the embedded frontend from a clean checkout.

The development server listens on `127.0.0.1:5178` and proxies `/api` and `/csrf` to the Go Admin API on `127.0.0.1:10099`.

Architecture and migration boundaries are defined in `../../../../contracts/admin-frontend.md` and `../../../../docs/admin-react.md`.
