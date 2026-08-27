# 0006: Use PostgreSQL as the collector control plane

## Status

Accepted

## Context

Game and Nav acquisition previously depended on process-relative timewheel jobs, Redis command/lease state, and Backend proxy endpoints. Restarts shifted interval phase, missed work was ambiguous, and a collector crash could leave no durable execution lineage. Game storage also retained active `gfg_game_v2_*` physical names after the V2 pipeline became canonical.

## Decision

PostgreSQL is the durable source of truth for Game and Nav collection schedules, jobs, runs, task results, collector instances, leases, cancellation, and history. Collectors remain autonomous schedulers and workers; Admin is the control plane and accesses its existing `gfg` and `gfn` pools directly. `go-co-op/gocron/v2` provides only periodic reconciliation/worker clocks, while `FOR UPDATE SKIP LOCKED`, active dedupe, a one-running-job-per-lane database invariant, and renewable leases own execution.

Schedules use fixed cron or anchored intervals. `scheduled_for` is first-class. Point-in-time missed slots are recorded as `missed` without a Run; state-refresh schedules may `catch_up_once`. Retry creates another Run attempt on the same state-refresh Job, while point-in-time work requires a new Run Now Job. Redis stores expiring realtime progress only and never owns durable queue state.

Active Game physical tables, sequences, indexes, constraints, and functions use unversioned names. Existing `/api/v2/game/*`, `/api/v2/nav/*`, and valid cache namespaces remain unchanged. Game player-count and Nav observation destructive retention are frozen until P0.2.

## Alternatives Considered

Using gocron as the queue, retaining Redis pending/lease sets, routing Admin through Backend collection proxies, adopting River/Asynq or a message broker, keeping compatibility views for `gfg_game_v2_*`, and backfilling point-in-time history were rejected because they introduce a second durable owner, preserve restart-relative behavior, or fabricate historical observations.

## Consequences

Collectors can recover expired leases, preserve stable schedule phase across restarts, expose explainable coverage and timing, and retain lineage from Job through Run and Result/Observation. Deployments must apply Game and Nav Goose migrations before starting the new collectors or Admin. Legacy scheduling configuration remains accepted only as initial schedule bootstrap input where documented; restart flags are ignored. P0.2 analytics and retention design remain out of scope.
