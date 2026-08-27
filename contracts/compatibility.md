# Compatibility contract

- Public API routes, response envelopes, status/error semantics, pagination/order, NULL handling, zero values, and not-found behavior remain compatible unless separately versioned.
- Redis key names and cache meanings are stable. Database migration must not silently redefine cache freshness or ownership.
- Existing collector data writes and after-commit cache refresh behavior remain compatible. Acquisition timing, retry, cancellation, and retention now follow the accepted PostgreSQL control-plane contract; process-relative startup collection and Redis command queues are retired.
- Nav V1 remains supported until an explicit, accepted deprecation decision defines its replacement and removal conditions.
- Schema and data changes must preserve active data, reject unexpected drift, and provide verified upgrade and recovery evidence appropriate to their risk.
- Admin remains a three-database PostgreSQL service. Auth and audit live in `gfa`; Nav CRUD in `gfn`; Game CRUD in `gfg`.
- Admin collection operations use the existing `gfg`/`gfn` pools directly. Game/Nav Backend collection proxy endpoints are not a compatibility surface.
- Cross-service changes must validate auth, representative Nav/Game behavior, collector runs, cache refresh, availability endpoints, and affected frontends.
