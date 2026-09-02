# GoFurry Admin

[中文](README_zh.md)

GoFurry Admin is the active operations application for the China-site stack. Its sole frontend lives in `react` and is embedded into the production Go binary. The service uses `gfa` for Admin authentication/audit state plus explicit `gfn` and `gfg` pools for Nav and Game operations. Redis supports existing runtime behavior.

Schema is owned exclusively by the root Goose migrations. Admin does not create or migrate tables during startup.

Admin authentication is a database-backed multi-account system with fixed `owner`, `developer`, and `operator` roles. Cookie JWTs carry account/session identity only; each request reloads the current active account and authorizes a compiled capability policy. See [Admin identity and authorization](../../../docs/admin-identity.md).

The Collection Center manages durable Schedule / Job / Run / Result / Collector Instance state directly through the existing `gfg` and `gfn` pools. It supports schedule enable/disable and Run Now, manual Game/Nav collection, queue/history, cancellation, constrained retry, audit, and ECharts outcome/coverage/timing. It does not proxy collection through either Backend, and Admin downtime does not stop autonomous Collector scheduling or workers.

React natively provides Collection, Metrics, Changes, Workbench attention, read-only Data Operations, Audit, and account governance. UI behavior consumes backend capabilities only. DataOps exposes safe metadata, Goose state, and bounded Top N storage information for the three pools; it never executes SQL or database maintenance.

## Development

Requirements: Go 1.26.7, Node.js/npm, PostgreSQL, and Redis.

~~~bash
# React Admin (development server on 5178, API proxy to 10099)
cd react
npm ci
npm run dev
cd ..

cp config/server.example.yaml config/server.yaml
# Edit the ignored local config before running.
go run . serve --config config/server.yaml
~~~

Other commands:

~~~bash
go run . --help
go run . version
go run . reset-password --config config/server.yaml --username owner --password '<new-password>'
~~~

The root command only displays help. `serve` runs in the foreground and shuts down Fiber, Redis, all PostgreSQL pools, and logging on SIGINT/SIGTERM.

## Production build and systemd

The root `build.bat admin` target builds React into the embed directory, then builds the self-contained Linux binary and deployment `dist/` companion. Install only from the final deployed binary and intended working directory:

~~~bash
cd /srv/gofurry/gofurry-admin
sudo ./gofurry-admin install --config /etc/gofurry-admin/server.yaml
sudo systemctl cat gofurry-admin
sudo systemctl start gofurry-admin
~~~

Install enables `gofurry-admin.service` but never starts it. It selects `SUDO_USER` as the runtime user and refuses to replace an existing unit unless `--force` is explicit.

~~~bash
sudo ./gofurry-admin uninstall
~~~

Uninstall removes only systemd registration; it does not delete the binary, config, logs, database, Redis data, or working directory. See [the shared systemd runbook](../../../docs/operations/systemd.md).

## Validation

~~~bash
go vet ./...
go test ./...
go build ./...

cd react
npm ci
npm run typecheck
npm test
npm run build
~~~

See [React Admin development](../../../docs/admin-react.md), [frontend parity](../../../docs/admin-frontend-parity.md), [role operations](../../../docs/operations/admin-roles.md), [Data and System Operations](../../../docs/admin-data-system-operations.md), and the [frontend contract](../../../contracts/admin-frontend.md).
