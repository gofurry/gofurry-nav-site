# Architecture decisions

Accepted Architecture Decision Records define durable choices that affect more than one active application or repository boundary.

- [0001: Active monorepo boundaries](0001-active-monorepo-boundaries.md)
- [0002: PostgreSQL engineering foundation](0002-postgresql-engineering-foundation.md)
- [0003: CLI and systemd lifecycle](0003-cli-and-systemd-lifecycle.md)
- [0004: Standalone availability observer](0004-standalone-availability-observer.md)
- [0005: Normalize the Game release and language domain](0005-game-domain-normalization.md)
- [0006: Use PostgreSQL as the collector control plane](0006-collector-control-plane.md)
- [0007: Build historical facts on eligibility, quality, and checkpoints](0007-historical-fact-foundation.md)
- [0008: Derive versioned analytics metrics from finalized historical facts](0008-analytics-metric-foundation.md)
- [0009: Materialize versioned change intelligence from canonical history](0009-change-intelligence-foundation.md)
- [0010: Repair published capability semantics by versioning contracts](0010-repair-published-capability-semantics.md)

New decisions should be short, use the same section structure, and supersede rather than rewrite an accepted record when the decision changes.
