# GoFurry React Admin

This is the target React 19 Admin frontend. During P0.5.2-B it runs independently through Vite; the Vue frontend in `../web` remains the production embed.

~~~text
npm ci
npm run dev
npm run typecheck
npm test
npm run build
~~~

The development server listens on `127.0.0.1:5178` and proxies `/api` and `/csrf` to the Go Admin API on `127.0.0.1:10099`.

Architecture and migration boundaries are defined in `../../../../contracts/admin-frontend.md` and `../../../../docs/admin-react.md`.
