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

Production installation uses the same foreground/systemd contract as the other Go services:

~~~bash
sudo install -d -o gofurry -g gofurry /var/lib/gf-uptime /var/log/gf-uptime
cd /srv/gofurry/gf-uptime
sudo ./gf-uptime install --config /etc/gf-uptime/server.yaml
sudo systemd-analyze verify /etc/systemd/system/gf-uptime.service
sudo systemctl start gf-uptime
curl --fail http://127.0.0.1:9980/readyz
~~~

Set `storage.path` to a durable absolute path such as `/var/lib/gf-uptime/uptime.db` and set the log path to a directory writable by the unit's runtime user. `install` enables but does not start. `uninstall` removes only systemd registration and preserves the Bbolt file, configuration, binary, and logs.

Back up the Bbolt file before replacing or moving the status service. This application has no Goose/sqlc ownership and never requires a business database migration.
