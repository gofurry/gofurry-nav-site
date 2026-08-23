# Operations assets

`ops` contains files consumed directly by production operations:

- `nginx`: frontend locations, maintenance configuration, and the static 503 page;
- `waf/coraza.conf`: the maintained custom Coraza configuration.

Cross-service prose and operator runbooks belong under `docs`, especially [the systemd lifecycle](../docs/operations/systemd.md). Application configuration and service-specific documentation remain under `apps/cn`.

No file in `ops` makes `legacy`, `experimental`, `third-party`, or `apps/intl` a production target.
