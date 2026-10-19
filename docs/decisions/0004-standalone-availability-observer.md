# 0004: Standalone availability observer

## Status

Accepted; implementation updated 2026-10-20: public availability is now owned by [LibreCosta Uptime](https://github.com/librecosta/uptime) v0.1.0. The former module is archived in `legacy/gf-uptime`; the current active set contains five Go modules. The original decision below is retained as historical context.

## Context

An availability observer coupled to a business API or business datastore can disappear with the system it is intended to observe and can turn monitoring into a runtime dependency.

## Decision

`gf-uptime` is a standalone Fiber service with local Bbolt history and no business PostgreSQL or Redis dependency. Applications expose narrowly scoped health endpoints; Collector endpoints remain internal. External monitoring probes `gf-uptime`, which independently probes configured targets.

## Alternatives Considered

Hosting uptime in Nav Backend, storing status history in business PostgreSQL or Redis, relying only on an external monitor, and building a general operations platform were rejected because they increase coupling or scope.

## Consequences

The observer can be placed outside the business host failure domain and retains only local state. Business services never depend on it, and monitored-target failures do not redefine local readiness.
