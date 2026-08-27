# GoFurry Admin

[中文](README_zh.md)

GoFurry Admin is the active operations application for the China-site stack. It embeds its Vue frontend and uses PostgreSQL only: `gfa` for Admin authentication/audit state plus explicit `gfn` and `gfg` pools for Nav and Game operations. Redis supports existing runtime behavior.

Schema is owned exclusively by the root Goose migrations. Admin does not create or migrate tables during startup.

The Collection Center manages durable Schedule / Job / Run / Result / Collector Instance state directly through the existing `gfg` and `gfn` pools. It supports schedule enable/disable and Run Now, manual Game/Nav collection, queue/history, cancellation, constrained retry, audit, and ECharts outcome/coverage/timing. It does not proxy collection through either Backend, and Admin downtime does not stop autonomous Collector scheduling or workers.

## Development

Requirements: Go 1.26.7, Node.js/npm, PostgreSQL, and Redis.

~~~bash
cd web
npm ci
npm run build
cd ..

cp config/server.example.yaml config/server.yaml
# Edit the ignored local config before running.
go run . serve --config config/server.yaml
~~~

Other commands:

~~~bash
go run . --help
go run . version
go run . reset-password --config config/server.yaml --password '<new-password>'
~~~

The root command only displays help. `serve` runs in the foreground and shuts down Fiber, Redis, all PostgreSQL pools, and logging on SIGINT/SIGTERM.

## Production build and systemd

The root `build.bat admin` target builds the web UI and Linux binary. Install only from the final deployed binary and intended working directory:

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

cd web
npm ci
npm run build
~~~
