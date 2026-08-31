# 0011: Expose Public Insights through backend-owned read models

## Status

Accepted

## Context

Production Accepted P0 facts, metrics, and change events are technical contracts. Their keys, versions, provenance, and raw before/after values are useful to operators but are not a stable public product API. Public Nav and Game clients need bounded overview, trend, entity, player, and price read models without acquiring those implementation details.

## Decision

Nav Backend and Game Backend own their domain's Public Insights anti-corruption layer. It reads finalized P0 facts, metrics, and changes through static SQL, sqlc, and pgx/v5, then maps them to stable public product keys and DTOs.

Metric versions and Change detector versions are explicitly compiled into each backend. Public contracts never follow Registry `active` state automatically. Admin APIs remain technical operator contracts and are not reused as public product APIs.

P1 adds no Analytics service, Insights persistence, materialized view, cache, queue, worker, or projector. Nuxt will own cross-domain composition in P1-B.

## Alternatives Considered

Exposing Admin DTOs, resolving the current active Registry version dynamically, adding an Analytics microservice, and projecting a separate Insights store were rejected because they leak technical versioning, permit unreviewed public contract changes, or add unnecessary operational state.

## Consequences

Nav and Game can evolve internal P0 contracts without silently changing public semantics. A new internal version requires a deliberate mapping change and regression tests. Initial reads use the existing indexed P0 tables directly; future caching or persistence requires separate evidence and a new decision.
