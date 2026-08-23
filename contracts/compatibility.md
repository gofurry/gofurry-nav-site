# Compatibility contract

- Public API routes, response envelopes, status/error semantics, pagination/order, NULL handling, zero values, and not-found behavior remain compatible unless separately versioned.
- Redis key names and cache meanings are stable. Database migration must not silently redefine cache freshness or ownership.
- Collector task selection, timing, retry, retention, V1 legacy writes, V2 writes, and after-commit refresh behavior remain unchanged.
- Nav V1 remains supported until an explicit, accepted deprecation decision defines its replacement and removal conditions.
- Schema and data changes must preserve active data, reject unexpected drift, and provide verified upgrade and recovery evidence appropriate to their risk.
- Admin remains a three-database PostgreSQL service. Auth and audit live in `gfa`; Nav CRUD in `gfn`; Game CRUD in `gfg`.
- Cross-service changes must validate auth, representative Nav/Game behavior, collector runs, cache refresh, availability endpoints, and affected frontends.
