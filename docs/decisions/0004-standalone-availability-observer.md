# 0004: Standalone availability observer

## Status

Accepted

## Context

An availability observer coupled to a business API or business datastore can disappear with the system it is intended to observe and can turn monitoring into a runtime dependency.

## Decision

`gf-uptime` is a standalone Fiber service with local Bbolt history and no business PostgreSQL or Redis dependency. Applications expose narrowly scoped health endpoints; Collector endpoints remain internal. External monitoring probes `gf-uptime`, which independently probes configured targets.

## Alternatives Considered

Hosting uptime in Nav Backend, storing status history in business PostgreSQL or Redis, relying only on an external monitor, and building a general operations platform were rejected because they increase coupling or scope.

## Consequences

The observer can be placed outside the business host failure domain and retains only local state. Business services never depend on it, and monitored-target failures do not redefine local readiness.
