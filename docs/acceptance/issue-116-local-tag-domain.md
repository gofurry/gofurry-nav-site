# Issue 116 local acceptance evidence (2026-09-16)

Scope: local dev, tag domain only. Shared development gfg and production were not
modified. The disposable shared development clone acceptance is pending the
operator checkpoint in `docs/operations/tag-domain-development-acceptance.md`.

## Root causes and decisions

Legacy parent rows were assignable Tags; prefix and numeric IDs encoded category
semantics. Primary/secondary physical columns and a separate synthetic-ID map
could disagree. Recommendation weights compared the wrong numeric prefixes, and
Nav Web detected Adult by numeric ID. The new three-table model eliminates those
current ambiguities; role relations include the primary/secondary union.

The reviewed Appendix A mapping was compared to the SQL: exactly 214 unique ID/code
entries match. No existing baseline or historical migration was edited. The final
schema snapshot and all affected sqlc models/queries were regenerated.

## Automated evidence

- PostgreSQL 18.6 on loopback only. User-authorized local disposable databases.
- Baseline suite: fresh gfg/gfn/gfa, adoption, drift rejection, upgrades, sequence
  floors and summary migration regression passed.
- Tag migration: 3 retained leaves, 2 old map pairs, 1 missing primary + 2 missing
  secondary relations yield 5 relations. Existing map timestamps and permanent
  IDs/codes survive; past and current facts are unchanged by migration. New
  projection/finalization uses version 2 while past facts retain version 1.
- Six negative migration cases passed: unmapped leaf, orphan prefix, parent
  assignment, duplicate map, invalid primary, equal primary/secondary. Down rejects
  rollback explicitly. Future generated IDs exceed the retained identity floor.
- Game Backend: go vet/test/build passed, including real PostgreSQL public detail,
  search, tag categories, cache reads and recommendation rebuild. CLI on a local
  fixture reported total=2 rebuilt=2 failed=0, algorithm similar-v2.4.0-hybrid-cbf.
  Unit tests cover role/category weights, arbitrary IDs, unknown categories and
  partial rebuild failure reporting.
- Game Collector: go vet/test/build passed. Repository, Changes, Metrics and lease
  recovery integrations ran against isolated local databases, not skipped. A
  dedicated full Collector YAML was required by its strict config decoder.
- Admin: go vet/test/build passed, including three-database and identity integration.
  Tag API tests cover code validation/immutability, archive/restore constraints,
  primary/secondary union, timestamp preservation, invalid-input rollback, active
  options, global cache invalidation and preserving scores for label-only edits.
- Admin React: npm ci/typecheck/build passed; 23 files, 115 tests passed. Includes
  Category/Tag create/edit/archive/restore and atomic classification with preserved
  tag search selections.
- Nav Web: npm ci/typecheck/build passed; tag-domain, insight-semantics and managed
  asset contracts passed. Game detail browser smoke passed SSR metadata/content,
  authoritative 404/503, missing Facts, repeated tab switches, and 390/768/1440/1920
  layouts for both visible and Adult-blurred galleries. The Adult fixture uses ID
  777777 to prove code-based behavior. Advanced search browser smoke also passed
  in Chinese/English at 390/1440px: explicit categories, leaf-only chips and
  selection persistence after searching. This is the focused visual guard; no claim
  is made that the unrelated full-site visual suite was rerun.
- tools go test, sqlc vet, and production-policy checks passed.

## Residual-reference audit

Current application code no longer refers to gfg_tag_map, gfg_tag.prefix, numeric
Tag category ranges, or Adult ID 1014. The final schema has no old current columns.
Legacy references remain intentionally in immutable migration/baseline history
and in upgrade fixtures that exercise the old schema. Historical Facts, Analytics
primary_tag_id/tag_id dimensions, and derived public/admin primary/secondary
response fields are legitimate and retained. Nav collector DNS prefix fields are
unrelated to Tag hierarchy and unchanged.

## Remaining operator work

Create/restore the development disposable clone and provide ignored clone runtime
configs. Actual clone row counts, backup/restore evidence, populated-catalog
rebuild and live Admin/Nav Web acceptance are not yet verified. Follow the linked
runbook; production migration/restarts are outside this task.
