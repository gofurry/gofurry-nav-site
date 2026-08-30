# 0009: Materialize versioned change intelligence from canonical history

## Status

Accepted

## Context

Historical Facts and Metrics describe entity state at a point in time, but a trustworthy change feed also needs stable event identity, semantic memory across unknown gaps, historical identity boundaries, event-time provenance, and reproducible rebuild behavior. Comparing current rows or adjacent days would create false changes after identity replacement and lose transitions separated by stale or unknown evidence.

## Decision

Game and Nav use separate Goose-owned, versioned change registries and a fixed compiled detector catalog. Runtime validates every active and retired Registry contract before starting. Each detector version owns an ordered PostgreSQL checkpoint and reads only canonical domain history, finalized Historical Facts, Metric Entity Daily, or effective-dated periods. Raw acquisition data, Redis, current catalogs, and the legacy Nav Redis Change feature are not canonical detector sources.

Canonical events use a deterministic text `event_key` and a unique detector/version/source-event identity. They retain projection date, explicit effective/observed/day time basis, semantic code, scope, old/new JSON objects, source keys and timestamps, source versions, and the latest materialization time. They never reference current Game or Site rows.

Metric, price, and certificate detectors retain semantic memory across unknown or failed days and compare only within one historical tracking identity. Release history is compared only where both states map reliably to one Game tracking period. Primary Target changes require a continuous effective-period replacement. Rebuild always propagates from the requested day through the detector checkpoint and never moves that checkpoint.

Collectors run Changes after Metrics. A detector-day transaction locks its checkpoint, verifies its metric/fact/closed-day watermark, replaces that detector day, validates event-code and time provenance, advances the checkpoint, and commits. A detector projection failure leaves its checkpoint unchanged without stopping acquisition, Facts, Metrics, or other detectors.

## Alternatives Considered

Current-row diffs, D-1-only comparison, generic JSON diff rules, random event IDs, Raw or Redis sources, importing or replacing legacy Nav Change, a new service/database, event retention, alerts, and a public change feed were rejected because they violate the frozen P0.4 reproducibility or compatibility boundary.

## Consequences

Game supplies five detectors for free status, Windows/Linux support, release, and regional price. Nav supplies five detectors for IPv6, TLS 1.3, security.txt, Primary Target, and TLS certificate. Existing Collectors expose `changes status`, `changes backfill`, and forward-propagating `changes rebuild`; Admin receives an authenticated read-only Change Center. Legacy Nav Change and all existing public routes remain unchanged. Public Change APIs/UI, alerts, retention, and additional detector families remain outside this decision.
