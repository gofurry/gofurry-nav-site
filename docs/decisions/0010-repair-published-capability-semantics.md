# 0010: Repair published capability semantics by versioning contracts

## Status

Accepted

## Context

The first Nav adoption contracts treated a successful composite DNS observation with no returned AAAA record as confirmed absence, even when the AAAA subquery failed. The first security.txt contract treated any 2xx response as adoption, including HTML application fallbacks, empty responses, and malformed documents. Both behaviors could turn unavailable evidence into a definitive state. The affected metric and detector versions were already released.

Schedule Run Now also used the generic manual-job insertion path without durable schedule lineage, which prevented Collection Center from relating that execution to its originating schedule.

## Decision

Keep `ipv6_adoption/1`, `security_txt_adoption/1`, `ipv6_transition/1`, and `security_txt_transition/1` immutable and compiled for explicit rebuild. Retire them through Goose and activate version 2 contracts from the first complete UTC day after migration.

IPv6 v2 records per-query AAAA evidence as present, confirmed absent, or unavailable. Only a successful NOERROR AAAA query with no AAAA answers is negative. Security.txt v2 requires a successful plain-text response with a non-empty, non-truncated body, at least one absolute Contact URI, and a valid future Expires timestamp. Invalid/unrecognized 2xx documents remain known observations but evaluate as unknown.

Run Now remains a manual trigger, retains the selected schedule ID/version, uses schedule-version-specific active deduplication, leaves `scheduled_for` null, and does not change schedule materialization phase.

## Alternatives Considered

Silently rewriting v1 projections, treating every ambiguous result as negative, destructively rebuilding old production history, importing legacy results into v2, and manufacturing a scheduled slot for Run Now were rejected because they break published semantics or fabricate lineage/evidence.

## Consequences

Existing v1 facts, metrics, and changes remain intact. Version 2 begins at an explicit cutover because historical observations do not reliably contain the new evidence fields. Operators deploy the Nav Goose migration before the Nav Collector/Admin binaries, then use normal Facts, Metrics, and Changes reconciliation for post-cutover days. No historical destructive rewrite is required.
