# 0002: PostgreSQL engineering foundation

## Status

Accepted

## Context

Multiple services share three PostgreSQL databases. Schema ownership and runtime query behavior need one auditable path without a repository-specific persistence framework.

## Decision

Goose is the sole schema migration source of truth. Normal business SQL is declared through sqlc, generated code is committed and never hand-edited, and production runtimes use pgx/v5 with bounded pgxpool connections. Admin uses explicit `gfa`, `gfn`, and `gfg` pools without claiming cross-database atomicity.

## Alternatives Considered

Application startup migrations, an ORM, a generic repository or UnitOfWork layer, a SQL builder, and a global database singleton were rejected because they obscure schema or transaction ownership.

## Consequences

Schema changes are explicit operator actions. Direct pgx remains limited to lifecycle, health, transaction control, and documented cases sqlc cannot represent without harming correctness.
