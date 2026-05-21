# Fiber v3 Medium Template

[Chinese](./README_zh.md)

`medium` is the balanced HTTP service edition of this scaffold. It keeps the practical runtime baseline from `heavy`, but removes platform-oriented runtime complexity so day-to-day API development stays straightforward.

## What This Template Includes

- SQLite-first demo experience with no external database required
- Complete `user` CRUD example
- Plain Go-style business structure under `internal/app`
- Lifecycle bootstrap for DB, Redis, logging, WAF, and graceful shutdown
- Service install and uninstall support through `kardianos/service`
- Reusable helper packages such as `pkg/httpkit` and `pkg/abstract`
- Official Fiber middleware baseline:
  request ID, access log, timeout, health probes, security headers, compression, ETag, rate limiting, pprof, Swagger, CSRF, and WAF
- Optional embedded UI support, disabled by default

## Project Positioning

This is the current `medium` version.

- Compared with `heavy`, it removes extra platform runtime burden.
- Compared with `light`, it still keeps Redis, WAF, service install/uninstall, and embedded UI support ready to use.
- Compared with the old over-assembled style, it keeps business code plain: `controller`, `dao`, `service`, and `models`, with routes registered directly in `url.go`.

If you want a backend template that is still production-oriented but avoids platform-style over-design, this is the intended direction.

## Quick Start

Default config file:

```bash
./config/server.yaml
```

Start the service:

```bash
go run . serve
```

On first startup the template will automatically:

- create `./data/app.db`
- auto-migrate registered models when `database.auto_migrate` is enabled
- expose the built-in user demo endpoints

Show the current version:

```bash
go run . version
```

Install or uninstall the service through the service manager integration:

```bash
go run . install
go run . uninstall
```

## Default Endpoints

Health and runtime endpoints:

- `GET /healthz`
- `GET /livez`
- `GET /readyz`
- `GET /startupz`

User CRUD demo:

- `GET /api/v1/user/`
- `POST /api/v1/user/`
- `GET /api/v1/user/:id`
- `PUT /api/v1/user/:id`
- `DELETE /api/v1/user/:id`

Optional endpoints:

- `GET /csrf/token` when CSRF is enabled
- `GET /swagger` when Swagger is enabled in debug mode
- `GET /debug/pprof/...` in debug mode

## CRUD Demo

Create a user:

```bash
curl -X POST http://127.0.0.1:9999/api/v1/user/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice",
    "email": "alice@example.com",
    "age": 24,
    "status": "active"
  }'
```

List users:

```bash
curl "http://127.0.0.1:9999/api/v1/user/?page_num=1&page_size=10&keyword=alice"
```

Update a user:

```bash
curl -X PUT http://127.0.0.1:9999/api/v1/user/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Updated",
    "email": "alice.updated@example.com",
    "age": 25,
    "status": "active"
  }'
```

Delete a user:

```bash
curl -X DELETE http://127.0.0.1:9999/api/v1/user/1
```

## Business Structure

The business layer stays intentionally plain:

- business code lives under `internal/app/<domain>`
- keep `controller`, `dao`, `service`, and `models` as normal packages
- route registration stays in `internal/transport/http/router/url.go`
- bootstrap registers runtime data such as database models directly

Current references:

- `internal/app/user/controller`
- `internal/app/user/dao`
- `internal/app/user/service`
- `internal/app/user/models`
- `internal/transport/http/router/url.go`
- `internal/bootstrap/lifecycle.go`

To add a new business domain:

1. Create `internal/app/<domain>`.
2. Add `controller`, `dao`, `service`, and `models` packages as needed.
3. Register that domain's routes in `internal/transport/http/router/url.go`.
4. Register its database models directly in `internal/bootstrap/lifecycle.go` if needed.

## Auto Migration

This template keeps `database.auto_migrate` because it is useful for the SQLite out-of-the-box experience.

That is the only schema bootstrap behavior kept in this `medium` version.

- no explicit migration command
- no migration directory requirement
- no migration tracking table

If you later need stricter schema management, that can be added in another variant without making the default workflow heavier.

## Middleware Baseline

Enabled by default:

- Request ID
- Access log
- Recover
- CORS
- Security headers
- Compression
- ETag
- Rate limiter
- Health probes
- Route-level timeout

Disabled by default but available:

- CSRF
- Swagger
- Redis
- WAF
- Embedded UI

Debug-only tooling:

- `pprof`

## Configuration Overview

Main config file:

```bash
./config/server.yaml
```

Important sections:

- `server`
- `database`
- `redis`
- `log`
- `middleware`
- `waf`

The `redis` section supports address, username, password, database index, and pool size.

The default config is intentionally runnable with only Go installed.

## Directory Layout

- `cmd`: CLI commands such as `serve`, `install`, `uninstall`, and `version`
- `config`: configuration files
- `internal/app`: business domains
- `internal/bootstrap`: lifecycle and health probes
- `internal/infra`: DB, logging, and cache infrastructure
- `internal/transport`: HTTP router and embedded UI
- `pkg`: shared abstractions and utilities

`pkg/httpkit` and `pkg/abstract` are intentionally kept in `medium` as reusable building blocks, even though the default demo module does not depend on them directly.

## Testing

Run the centralized test suites from `v3/test`:

```bash
cd ../test
go test ./...
```

Current integration coverage includes:

- service bootstrap with SQLite
- automatic database creation
- automatic table creation through `auto_migrate`
- health probes
- request ID and security headers
- ETag and compression behavior
- end-to-end user CRUD flow

## Known Tradeoffs

- Route registration is intentionally centralized in `internal/transport/http/router/url.go`.
- Runtime model registration is still centralized in `internal/bootstrap/lifecycle.go`.
- Request ID is already present in access logs, but business logs are not yet automatically enriched with request context.

## Template Checklist

Before turning this into your own service:

- replace the module path in `go.mod`
- update app identity in `config/server.yaml`
- remove the demo user domain if you no longer need it
- add your own business domains under `internal/app`
