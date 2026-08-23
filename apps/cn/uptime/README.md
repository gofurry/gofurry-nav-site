# GoFurry Uptime

`gf-uptime` is the independent, intentionally small availability service for the active GoFurry stack. It uses Fiber's uptime middleware and persists history in a local Bbolt file. It has no PostgreSQL or external Redis dependency.

~~~bash
go run . serve --config conf/server.example.yaml
~~~

Endpoints:

- `/livez` — process liveness
- `/readyz` — Fiber uptime and local Bbolt storage readiness
- `/uptime` — public status UI
- `/uptime/api/status` — status JSON used by the UI

Use stable endpoint IDs after production rollout. Deploy the service behind a reverse proxy, preferably on a failure domain separate from the business host, and monitor its public `/readyz` with an external HTTP monitor.
