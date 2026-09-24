# Issue 116 local acceptance evidence (2026-09-16)

Scope: local and shared development, tag domain only. Production was not accessed.
After the operator provisioned `gfg_tag_refactor_test`, the workstation
completed the clone migration and the database/API checks below. The provisioning
and cleanup boundaries remain in `docs/operations/tag-domain-development-acceptance.md`.

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

## Populated development clone evidence (2026-09-16)

- Operator-created clone was empty and owned by `gofurry_migrator`. Workstation
  `pg_dump`/`pg_restore` used only the verified development host. Goose targeted
  only `gfg_tag_refactor_test`; runtime checks used `gofurry_app`.
- The initial full dump was rejected for lack of access to the infrastructure
  `infra_backup.restore_canary`. Nothing had been restored. After confirming all
  business objects were in `public`, the successful business-domain backup excluded
  only `infra_backup`, retaining the required `pg_trgm` extension. No permissions
  were elevated. Custom archive catalog verification and restore both succeeded.
  Archive size: 9,048,246 bytes; SHA-256:
  `3439f751806d1ac743c15f64f28bb083cf59b0e66898bcb52115cc0abb793615`.
  The archive and private runtime configuration remain outside Git on the workstation.
- Migrated `20260916010000` to `20260916020000`: 213 Games, 218 legacy Tag rows
  became 4 Categories plus 214 retained leaf Tags. All 214 permanent codes match
  the explicit mapping. 3,882 legacy map pairs plus 18 missing primary and 31
  missing secondary pairs became 3,931 relations: 212 primary, 212 secondary,
  3,507 normal. Exact memberships, roles and timestamp rules passed comparison.
- All 34 unrelated business tables retained their row counts and content hashes;
  Game content excluding removed columns also matched. Removed current columns,
  map table and function references were absent. Runtime grants were verified.
  Final schema matched `tools/db-baseline/expected-final/gfg.json`; fingerprint:
  `278ecb518b7b2ac6ea3358be48cae3f1b2be63bb8adbe736c533bd6734043946`.
- This source contains zero Game Daily rows and zero tracking periods. Consequently
  its unchanged empty history is not evidence for populated historical preservation.
  Existing isolated migration regressions cover that case. A clone-only transaction
  additionally seeded tracking/current/past fixtures, verified exact projected
  membership and roles, current/finalized version 2, preserved past version 1, and
  generated Tag ID above the retained maximum 9010, then rolled all fixture rows
  back. Identity sequence advancement from this test is expected.
- Populated Game Backend checks passed in Chinese and English: 4 explicit
  Categories, 214 leaves, Adult code and role/category metadata, and actual HTTP
  handlers for Tags, Categories, detail, search and recommendations.
- Private recommendation CLI completed on the populated clone:
  `algorithm=similar-v2.4.0-hybrid-cbf total=213 rebuilt=213 failed=0`.
  SQL verification found 13,632 cache rows across all 213 sources, only the new
  algorithm version, and maximum rank 64. No HTTP service, collector or Redis was
  started for the rebuild.
- Admin HTTP checks passed against the clone Game pool: Categories, Tags, searchable
  options, workspace, atomic no-op classification and invalid-role rejection without
  changing Game weight/membership. Authentication/audit used a local disposable
  Admin database, which was removed afterward; shared `gfa`/`gfn` were not used.
  These targeted probes ran through temporary workstation test harnesses; they do
  not replace the committed regression suites or claim live browser acceptance.

## Remaining operator work

Live Admin/Nav Web browser acceptance against this populated clone remains manual:
Category/Tag creation and archive/restore flows, category labels, preserved search
selections, Adult display, and bilingual desktop/mobile Gallery. Earlier browser
smokes used isolated fixtures, not this populated clone.

At the end of clone acceptance, shared development `gfg` was still at
`20260916010000`. The subsequently authorized shared-development migration is
recorded below. Clone deletion remains an operator checkpoint. Production
migration/restarts are outside this task.

## Authorized shared development cutover (2026-09-16)

After reviewing the successful clone acceptance, the user explicitly authorized
migration of shared development `gfg`. The workstation verified the configured
development host, database name and migrator role; there were no other connections
before backup or immediately before migration. No credentials, ports, Compose,
Tailscale configuration or infrastructure roles were changed.

A fresh business-domain backup excluded only the previously audited
`infra_backup` schema. Its catalog was verified, and the archive was successfully
restored into a local disposable PostgreSQL database: 213 Games, 218 legacy Tags,
3,882 pairs, version `20260916010000`. That local verification database was then
removed. A verified backup copy is retained in the workstation's private backup
directory, outside Git. SHA-256:
`16f8718a2bd851ca742be6a25c1cefd7fca078b84bdff828a02b3a24a7673063`.

Goose `up-to 20260916020000` completed against shared development `gfg`, with a
10-second lock timeout and a bounded statement timeout. The resulting 4 Categories,
214 original leaf identities/codes and 3,931 role relations matched the saved
pre-migration data and timestamp rules exactly. All 34 unrelated table hashes and
Game content excluding removed columns remained unchanged. Removed objects were
absent; `gofurry_app` inherited the required new-table permissions from the existing
default privileges. No shared database fixture rows were inserted for acceptance.

The shared database's complete public schema matched the committed final snapshot
(fingerprint `278ecb518b7b2ac6ea3358be48cae3f1b2be63bb8adbe736c533bd6734043946`).
Game Backend's Chinese/English Tags, Categories, detail, search and recommendation
HTTP handler checks passed using `gofurry_app` with read-only transactions enforced.
The private recommendation rebuild then completed against shared development:
`algorithm=similar-v2.4.0-hybrid-cbf total=213 rebuilt=213 failed=0`.
Final SQL verification found 13,632 rows, 213 sources, maximum rank 64 and no old
algorithm versions. Final data comparisons still preserved every migrated
membership/timestamp and all unrelated table hashes. Both shared `gfg` and the
retained clone had zero other connections after the workstation checks finished.

Developers must restart current Game Backend, Game Collector, Admin (including the
updated embedded React build) and Nav Web on their workstations. Old binaries are
incompatible with the migrated schema. Application services were not started or
restarted by this database operation; the infrastructure server remains infrastructure-only.
