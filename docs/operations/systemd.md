# Linux/systemd lifecycle

The six active Go binaries generate their deployment-specific unit at install time. There are no canonical static production `.service` files in this repository.

## Service names

| Binary | Unit |
|---|---|
| `gf-nav` | `gf-nav.service` |
| `gf-nav-collector` | `gf-nav-collector.service` |
| `gf-game` | `gf-game.service` |
| `gf-game-collector` | `gf-game-collector.service` |
| `gofurry-admin` | `gofurry-admin.service` |
| `gf-uptime` | `gf-uptime.service` |

## Install contract

Run installation from the directory that must become `WorkingDirectory`, using the final deployed binary and an explicit configuration:

~~~bash
cd /srv/gofurry/gf-nav
sudo ./gf-nav install --config /etc/gf-nav/server.yaml
~~~

Do not use `go run . install` for a durable installation because `go run` executes a temporary binary.

Before writing anything, the command parses and validates the same typed configuration used by `serve`. It then resolves:

- the canonical current executable from `os.Executable()`;
- the absolute install-time current directory;
- the absolute existing configuration file;
- `SUDO_USER` when invoked through sudo, otherwise the current OS user.

Invoke `sudo` from the account that is intended to run the service. A root login without `SUDO_USER` intentionally produces `User=root`.

The installer verifies Linux, a usable running systemd manager, and root privileges. It atomically writes `/etc/systemd/system/<name>.service`, runs `systemctl daemon-reload`, and enables the unit. It never starts or restarts the service.

Expected output:

~~~text
Installed: /etc/systemd/system/gf-nav.service
Enabled:   yes
Started:   no

Start manually with:
  sudo systemctl start gf-nav
~~~

An existing unit is rejected by default. `install --force` is the only overwrite path; it still validates first and never starts the service.

Generated units use `Type=simple` and execute:

~~~text
<absolute-binary> serve --config <absolute-config>
~~~

with the resolved user/current directory, `Restart=on-failure`, `RestartSec=5`, and `LimitNOFILE=65535`. `ExecStart` arguments use systemd command-line quoting, while scalar directives such as `WorkingDirectory` are emitted without surrounding quotes; systemd `%` specifiers are escaped in both forms. No shell command is constructed.

## Review and manual start

Always review before starting:

~~~bash
sudo systemctl cat gf-nav
sudo systemd-analyze verify /etc/systemd/system/gf-nav.service
sudo systemctl start gf-nav
sudo systemctl status gf-nav --no-pager
journalctl -u gf-nav -n 100 --no-pager
~~~

Repeat with the relevant service name. Confirm `User`, `WorkingDirectory`, `ExecStart`, and `--config` point to the intended production values.

## Uninstall contract

~~~bash
sudo ./gf-nav uninstall
~~~

Uninstall stops the unit if active, disables it if enabled, removes only `/etc/systemd/system/gf-nav.service`, reloads systemd, and resets failed state where applicable. Repeating it for an already stopped, disabled, or absent unit is safe.

It never deletes the binary, configuration, logs, working directory, PostgreSQL state, Redis state, the uptime Bbolt file, or other application data.

## Migration from kardianos-generated units

Migrate one service at a time:

1. Record `systemctl status <name>` and save `systemctl cat <name>` to an operator-controlled backup.
2. Save the current binary and config; record the current runtime user and working directory.
3. Deploy the new binary to its final path, but do not start it.
4. Stop and disable the old registration: `sudo systemctl stop <name>` then `sudo systemctl disable <name>`.
5. Remove only the old `/etc/systemd/system/<name>.service` and run `sudo systemctl daemon-reload`. Confirm no old process remains.
6. As the intended runtime account, `cd` to the intended working directory and invoke the new binary through sudo:
   `sudo ./<binary> install --config /absolute/path/server.yaml`.
7. Do not use `--force` unless the old unit was deliberately stopped, backed up, and the remaining file was reviewed.
8. Review the new unit and optionally run `systemd-analyze verify`.
9. Start manually with `sudo systemctl start <name>`.
10. Verify status, logs, readiness, application smoke tests, and collector activity before moving to the next service.

For rollback, stop the new unit, uninstall its registration, restore the saved old binary/unit, run `daemon-reload`, enable it, and start it. This systemd migration does not require or perform database migration.
