# 0001: Active monorepo boundaries

## Status

Accepted

## Context

The repository contains production applications, placeholders, archived systems, experiments, and maintained third-party source. Treating every directory as active makes builds, dependency checks, and ownership ambiguous.

## Decision

Active production code is limited to the six Go modules and Nav Web under `apps/cn`. `apps/intl` is placeholder-only. `legacy`, `experimental`, and `third-party` are excluded from active builds, CI, vulnerability scans, deployment tooling, and dependency graphs.

## Alternatives Considered

Modernizing the entire tree, creating a repository-wide `go.work`, and splitting every application into a separate repository were rejected because they either expand active scope or weaken the explicit application boundaries.

## Consequences

Tooling must list active modules deliberately. Archived code can remain available for reference, but reviving it requires a separately accepted, scoped decision.
