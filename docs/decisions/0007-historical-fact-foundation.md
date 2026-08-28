# 0007: Build historical facts on eligibility, quality, and checkpoints

## Status

Accepted

## Context

Current Game/Nav rows and raw observations cannot answer historical questions after identity changes, deletion, missed acquisition, or raw retention. Daily tables alone would also blur a measured value with the quality of the acquisition that produced it.

## Decision

Game/AppID and Nav target/Primary identities use effective-dated periods with UTC `[start,end)` semantics. Historical Fact tables reference those periods, never cascade from current entities, use unversioned physical names, and carry `projection_version=1`. Fact values and quality counters are separate; unknown state remains nullable, manual Player samples never enter canonical statistics, and manual successful Nav observations may update whitelisted known state without changing scheduled quality.

The existing Game and Nav Collectors each own two in-process pipelines. They reconcile after acquisition, lock one PostgreSQL checkpoint row with `FOR UPDATE`, finalize at most one closed UTC day per runtime pass, validate/project facts, advance the checkpoint, and commit. Runtime, backfill, and rebuild use the same implementation. No Fact Job, Run, Task, queue entry, or new service is created.

Destructive raw retention is a separate post-checkpoint transaction and is disabled by default. Game Player Raw requires the `game.player_facts` checkpoint and configured age. Nav Observation retention requires the `nav.target_facts` checkpoint and preserves the configured newest rows independently per `(site_id,target,protocol)`. Facts have no automatic retention.

## Alternatives Considered

Deriving history from current entities, storing a final coverage ratio, treating unknown as false, including manual probes in scheduled quality, copying full Nav payloads, creating a rollup queue/service, and pruning raw in the checkpoint transaction were rejected because they fabricate history, hide quality semantics, duplicate durable ownership, or make recovery unsafe.

## Consequences

P0.2 supports reproducible hourly/daily Player facts, historical Game/Price state, protocol quality and last-known state, typed Nav target facts, Site dimensions, Primary Target history, idempotent rebuild, and checkpoint-gated retention. Nav Site Daily permits `finalized_at=NULL` only for the mutable current-day Admin marker; closed-day rows are finalized by the Site pipeline. P0.3 metric registry/coverage analytics and P0.4 health/adoption work remain outside this decision.
