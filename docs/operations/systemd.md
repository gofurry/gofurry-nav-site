# systemd lifecycle

The five Go applications expose a foreground `serve` command and Linux-only `install` / `uninstall` helpers. systemd supervises the foreground process.

Installation requires the application binary to be run from its intended deployment directory with an explicit configuration file. It writes and enables a unit but does not start it. Uninstall stops, disables, and removes only the unit registration; it does not remove binaries, configuration, logs, or application data.

Exact commands and rollback checks are documented after the CLI migration is complete.
