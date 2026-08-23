# Compatibility contract

- Public API routes, response envelopes, status/error semantics, pagination/order, NULL handling, zero values, and not-found behavior remain compatible unless separately versioned.
- Redis key names and cache meanings are stable. Database migration must not silently redefine cache freshness or ownership.
- Collector task selection, timing, retry, retention, V1 legacy writes, V2 writes, and after-commit refresh behavior remain unchanged.
- Nav V1 remains reachable; this engineering migration does not deprecate it. The three legacy-active Nav collector log tables remain present.
- No migration may lose active data. Baseline adoption records schema history only after an exact structural match; intentional drift must fail without creating the Goose version table.
- Admin remains a three-database PostgreSQL service. Auth and audit live in `gfa`; Nav CRUD in `gfn`; Game CRUD in `gfg`.
- Frontends should require no broad database-driven rewrite. Manual regression focuses on auth, representative Nav/Game CRUD, public Nav V1/V2, collector runs, cache refresh, and frontend build/smoke behavior.
