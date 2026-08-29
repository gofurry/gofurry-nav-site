# 0008: Derive versioned analytics metrics from finalized historical facts

## Status

Accepted

## Context

Historical Facts preserve entity state and acquisition quality, but they do not define a stable analytics population, evaluator version, explainable state, or aggregate contract. Computing adoption directly from Raw/current catalogs would make old results change after retention, identity mutation, or evaluator changes.

## Decision

Game and Nav use separate Goose-owned, versioned metric registries. Alpha.4 supports only `state_ratio` at daily grain and a fixed compiled evaluator catalog. Runtime validates every active or retired Registry contract against that catalog before starting so retained versions remain explicitly rebuildable; evaluator predicates remain explicit SQL/code rather than a dynamic DSL.

Each metric version owns an ordered PostgreSQL checkpoint. An atomic metric-day transaction locks that checkpoint, verifies the upstream Fact watermark, replaces entity and aggregate rows, validates state-count conservation, advances the checkpoint, and commits. Rebuild uses the same projector without moving the checkpoint. A projection error leaves the checkpoint unchanged and does not stop acquisition, Fact reconciliation, or other metrics.

Metric population is finalized Historical Facts with `tracked_at_end=true`. Entity rows use the fixed states `positive`, `negative`, `stale`, `not_probed`, `probe_failed`, `unknown`, and `not_applicable`, retain a reason, historical evidence timestamp, dimensions, and source projection versions, and never reference current Game/Site rows. Freshness is evaluated against the historical UTC day end, never runtime time.

Aggregates contain only global and single-dimension state counts. Adoption and coverage ratios are computed at query time and are null for a zero denominator. Registry, entity, aggregate, and checkpoint tables have unversioned physical names; retired versions retain history and remain explicitly rebuildable while runtime advances active versions only.

## Alternatives Considered

Reading Raw/acquisition ledgers/current catalogs, persisting final ratios, treating unknown as false, using runtime `now()` for freshness, building a JSON rule engine, using one global checkpoint, creating a separate analytics service, and materializing a multi-dimensional cube were rejected because they violate reproducibility, explainability, or the frozen P0.3 boundary.

## Consequences

Game supplies `free_game_share`, `windows_support`, and `linux_support`; Nav supplies `ipv6_adoption`, `tls13_adoption`, and `security_txt_adoption`. Existing Collectors run Metrics after Facts and expose status/backfill/rebuild commands. Admin receives an authenticated read-only Metric Center whose entity names come from same-day Historical Facts. Public analytics APIs, dashboards, change intelligence, metric retention, and additional metrics remain outside this decision.
