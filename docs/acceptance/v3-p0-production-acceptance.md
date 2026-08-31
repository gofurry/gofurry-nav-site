# V3 P0 Production Acceptance

## Acceptance header

- Final status: **PASS**
- Initial audited commit: `36ab41f79860c50b00d0c842ca960fb84c9c0da0`
- Acceptance candidate commit: `3ddf8fa46b8d540be49b308be8db62310e16bca9`
- Accepted commit SHA: `9794c4e19a5f025c2777a68e02e14e489cb9a3e6`
- Acceptance execution date: `2026-08-31`
- Main production acceptance window: `2026-08-29` through `2026-08-30` UTC, two consecutive closed UTC days
- Nav repaired v1 comparison window: `2026-08-28` through `2026-08-29` UTC
- Production databases checked: `gfa`, `gfn`, `gfg`
- Production read-only evidence source: a temporary PostgreSQL role supplied out of band; every session set `transaction_read_only=on`, used an explicit read-only transaction, a 15-second statement timeout, and a 2-second lock timeout
- CI result: [`engineering-foundation` run `33400459588`](https://github.com/gofurry/gofurry-nav-site/actions/runs/33400459588) passed for the accepted commit

No host, username, password, DSN, token, cookie, or Admin account identity is stored in this report. No production mutation, migration, DDL, collection control, rebuild, backfill, account, permission, password, session, or task operation was performed. The temporary database role had no table write grants, no database `CREATE`, and no superuser, role-management, database-creation, replication, or row-security-bypass privilege.

Observed database metadata:

| Database | PostgreSQL | Observed size (bytes) | Repository/production Goose version |
|---|---|---:|---:|
| `gfa` | 18.3 | 9,647,807 | 20260830030000 |
| `gfn` | 18.3 | 1,829,205,695 | 20260830020000 |
| `gfg` | 18.3 | 253,679,295 | 20260830010000 |

## Acceptance cohort

- Fixed Nav Site IDs (12): `338, 337, 336, 335, 334, 333, 332, 331, 330, 329, 328, 327`
- Additional seeded Nav sample (seed `503`, excluding the fixed cohort): `310, 291, 283`
- Fixed Game IDs (12): `312, 311, 310, 309, 308, 307, 306, 305, 304, 303, 302, 301`
- Additional seeded Game sample (seed `503`, excluding the fixed cohort): `206, 209, 291`
- Isolated semantic cohort retained from integration acceptance:
  - Nav Site IDs `99200`–`99208` cover positive, negative, stale, not-probed, probe-failed, unknown, not-applicable, confirmed-absence, and unavailable-evidence behavior.
  - Game IDs `99200`–`99205` cover positive, negative, stale, unknown, not-applicable, platform support, and missing evidence.

Selection rationale: production fixed cohorts are stable, bounded latest finalized entities; the separately selected seeded samples reduce hand-selection bias; the isolated cohort explicitly exercises frozen states that are not guaranteed to occur naturally in a two-day production window. Both production cohorts were traced across the two closed days.

## G0 — Schema & Migration

Status: **PASS**

Scope:

- Current and reproducible `gfg`, `gfn`, and `gfa` Goose schemas.
- Critical Collection, Fact, Metric, Change, account, audit, and authorization schema invariants.
- Runtime DDL/startup migration prohibition.

Evidence:

- Production Goose versions exactly matched the repository: `gfg=20260830010000`, `gfn=20260830020000`, and `gfa=20260830030000`; all latest rows were applied and pending migration count was zero.
- Targeted live-catalog checks found every required table and critical unique/active-lane/Registry index for Collection, Facts, Metrics, Changes, Admin accounts, and audit.
- The production role boundary itself was verified before business evidence was read: read-only transaction enabled, no write grants, and no elevated role/database privileges.
- `TestPostgresFreshAndBaselineAdoption` passed on PostgreSQL 18.3, including fresh, baseline adoption, released upgrade boundaries, drift rejection, seeded row preservation, and cleanup for all three databases.
- `go tool sqlc vet`, `go tool sqlc generate`, generated-diff validation, `check-sqlc`, and `check-production-policy` passed.
- Active Go source contains no `AutoMigrate`, startup Goose invocation, or runtime `CREATE/ALTER TABLE` pattern.

Checks:

- PASS — all three production schemas are current with no pending migration.
- PASS — critical live objects and uniqueness guards are present.
- PASS — isolated adoption, upgrade, drift, and backup-copy checks pass.
- PASS — Goose remains the sole schema owner and runtime DDL policy passes.

Resolved findings:

- `PA-001`

Open blockers:

- None.

Deferred:

- None.

## G1 — Acquisition

Status: **PASS**

Scope:

- Schedule → Job → claim → Run → Task Result lineage.
- Run Now, retry, cancel, collector instance history, leases, lanes, and misfire semantics.

Evidence:

- Production exposed 10 enabled Nav schedules and 2 enabled Game schedules with coherent cron/interval, timezone, version, materialization cursor, and next-slot values.
- Recent scheduled Nav and Game Jobs retained schedule ID, schedule version, and scheduled slot; sampled Runs matched expected/attempted/result counts.
- Recent production activity reached Nav Job `272` at `2026-08-31T14:30:00Z` and Game Job `1837` at `2026-08-31T14:00:00Z`; both had complete Job → Run → Result lineage.
- The bounded latest 5,000 Jobs/Runs and 10,000 Results had zero invalid scheduled lineage, invalid running claim, expired running lease, duplicate running lane, orphan Run, orphan Result, or counter imbalance.
- Current Game and Nav collector instances had fresh heartbeats and retained stopped-instance history separately.
- Existing production manual Nav Jobs had no scheduled slot and did not masquerade as scheduled executions. Schedule Run Now lineage and retry/cancel mutations were actively validated only in isolation.
- `TestAdminThreeDatabasePersistence`, schedule phase/misfire tests, `TestGameExpiredLeaseRecoveryIntegration`, and `TestNavExpiredLeaseRecoveryIntegration` passed without required-test skips.

Checks:

- PASS — recent production acquisition is explainable end to end.
- PASS — current/history instance semantics and heartbeat freshness are correct.
- PASS — production invariant scan returned zero violations.
- PASS — isolated manual/retry/cancel/misfire/concurrency/recovery behavior matches the frozen contract.

Open blockers:

- None.

Deferred:

- None.

## G2 — Historical Facts

Status: **PASS**

Scope:

- Closed UTC-day projection, immutable history, tracking identity, evidence/value separation, checkpoints, retention safety, and deterministic rebuild.

Evidence:

- Game `player_facts` and `state_facts`, and Nav `target_facts` and `site_facts`, were all processed through `2026-08-30` with zero lag against the latest closed UTC day.
- Nav finalized exactly 221 Site rows on each of `2026-08-29` and `2026-08-30`; Game finalized exactly 213 Game rows on each day. Projection version was uniformly 1 and no closed-day row was left unfinalized.
- Bounded recent invariant checks found zero future-finalized rows, invalid projection versions, invalid tracking/effective ranges, duplicate active target identities, duplicate active Primary periods, or duplicate active explicit Game tracking periods.
- Fixed and additional seeded cohorts were traced over both days through tracking identity, Fact state, evidence timestamp, acquisition-ledger quality/counts, projection version, and price/player or DNS/security protocol details.
- After the Nav capability repair migration, current production DNS observations contained 203 `present` and 41 `confirmed_absent` AAAA evidence classifications. No unavailable evidence was coerced to negative in the sampled production evidence.
- Game and Nav Fact integration suites passed projection, quality, unavailable-evidence behavior, tracking, idempotent rebuild, missing-source rejection, checkpoint-safe retention, and canonical history preservation.

Checks:

- PASS — the two-day production window is finalized and checkpoint-current.
- PASS — production fixed/random samples retain explainable evidence and identity.
- PASS — production structural/tracking invariants returned zero violations.
- PASS — isolated rebuild and unavailable-evidence tests are deterministic and truthful.

Open blockers:

- None.

Deferred:

- None.

## G3 — Metrics

Status: **PASS**

Scope:

- Every active Metric contract, state/denominator/coverage arithmetic, checkpoints, lineage, and deterministic rebuild.

Evidence:

- Active Game `free_game_share/1`, `windows_support/1`, and `linux_support/1` were processed through `2026-08-30`. For both closed days, every global Daily row passed count arithmetic and exactly reconciled to its Entity rows.
- Active Nav `tls13_adoption/1` was processed through `2026-08-30`; both closed days reconciled 221 population / 220 eligible entities with `positive=193`, `negative=9`, `probe_failed=15`, `unknown=3`, and one not-applicable entity.
- Retired Nav `ipv6_adoption/1` and `security_txt_adoption/1` retained their two expected production days (`2026-08-28` and `2026-08-29`) and reconciled exactly to Entity state counts.
- Repaired active Nav `ipv6_adoption/2` and `security_txt_adoption/2` have the Goose-owned source start `2026-08-31`. At execution time that UTC day had not closed, so `processed_through=NULL` and no v2 Daily row were the correct expected state rather than lag; see `PA-005`.
- Production cross-gate samples preserved `unknown`, `probe_failed`, and not-applicable states distinctly. Isolated v2 PostgreSQL suites passed explicit confirmed-absence/unavailable/invalid-document semantics and deterministic rebuild.

Checks:

- PASS — every metric expected to run in the production windows reconciled Entity → Daily exactly.
- PASS — all active checkpoints were at their valid upstream boundary; no checkpoint advanced into an open/future day.
- PASS — coverage arithmetic never substituted a fake zero for an unavailable denominator.
- PASS — active and retained retired versions remain backed by compiled contracts and isolated deterministic rebuild evidence.

Resolved findings:

- `PA-005` documented the cutover boundary; no defect or fix was required.

Open blockers:

- None.

Deferred:

- None.

## G4 — Change Intelligence

Status: **PASS**

Scope:

- Every active Detector contract, transition semantics, stable identity, time/provenance, checkpoints, duplicate prevention, and deterministic rebuild.

Evidence:

- All five active Game detectors were processed through `2026-08-30`. The two-day window contained three price-decrease/discount-start transitions followed by three price-increase/discount-end transitions.
- Active Nav `tls13_transition/1`, `primary_target_transition/1`, and `tls_certificate_transition/1` were processed through `2026-08-30`. No qualifying Nav transition occurred in the bounded 14-day production event query, which is a valid empty event set.
- Repaired active Nav `ipv6_transition/2` and `security_txt_transition/2` share the truthful `2026-08-31` source start and were not yet eligible for a closed-day run; their pre-repair v1 predecessors remained preserved through their last eligible date.
- Bounded production checks returned zero missing required event timestamps, invalid day-basis timestamps, observed/source-after timestamp mismatches, or undeclared event codes. Primary keys and detector/version/source-event unique constraints were present live.
- Game/Nav PostgreSQL change suites passed semantic memory across gaps, tracking identity, stable event keys, duplicate prevention, old/new provenance, checkpoint bounds, forward-propagating rebuild, and stable event sets.

Checks:

- PASS — production checkpoints match valid upstream boundaries.
- PASS — sampled production events have valid identity, code, time basis, and provenance shape.
- PASS — empty Nav event evidence is consistent with no observed semantic transition, not a missing pipeline run.
- PASS — isolated transition and rebuild suites cover all active/retired detector contracts.

Open blockers:

- None.

Deferred:

- None.

## G5 — Admin & RBAC

Status: **PASS**

Scope:

- Owner, Developer, and Operator capability behavior in backend and React UI.
- Production account/audit state plus safe frontend boundary evidence.

Evidence:

- Production `gfa` contained one active Owner and one active Developer, with no invalid role/status value and an intact non-zero active-Owner invariant.
- Bounded production audit aggregates showed recent login, account creation/display-name/password operations, Collection schedule operations, and content operations without reading password hashes, secret-shaped before/after data, account names, or session material.
- The direct Admin service port was not publicly reachable, consistent with the private-listener deployment guidance. A safe unauthenticated GET through the public host reached the production reverse-proxy boundary; no production login POST or authenticated mutation was attempted.
- `TestAdminIdentityAuthorizationPersistence`, `TestAdminLegacyIdentityUpgrade`, and `TestAdminThreeDatabasePersistence` passed without skipping. They cover bootstrap/login/current Principal, session invalidation, account authorization, audit identity, concurrent last-active-Owner protection, and read models.
- Isolated direct API evidence covers all three personas: Operator business views with technical/governance denials; Developer technical/DataOps/Audit access with Accounts denied; Owner account governance and last-active-Owner enforcement.
- React strict typecheck, 10 test files / 33 tests, production build, embedded nested-route/API boundary tests, capability-aware UX tests, and theme tests passed. CI rebuilt the embedded React Admin for the accepted commit.

Checks:

- PASS — compiled backend capability policy and direct 403 behavior cover Owner, Developer, and Operator.
- PASS — React consumes backend capabilities and does not reconstruct role policy.
- PASS — production account/audit state is structurally valid and consistent with the deployed identity model.
- PASS — no production session/account mutation was used as acceptance evidence.

Open blockers:

- None.

Deferred:

- None.

## G6 — Recovery

Status: **PASS**

Scope:

- Collector restart/lease expiry, PostgreSQL dependency visibility, retry, misfire reconciliation, checkpoint lag, and Facts/Metrics/Changes convergence.

Evidence:

- Game/Nav expired-lease integration tests prove durable terminal recovery, lane release, single replacement claim, and absence of orphan Runs.
- Collector health tests prove readiness transitions from ready → dependency unavailable → ready after recovery, while liveness remains independent.
- Schedule tests prove restart uses the durable cursor and preserves anchored phase; misfire tests preserve existing skip/catch-up policy.
- Admin integration proves failure/cancel/retry lineage in isolated databases.
- Fact/Metric/Change integration suites prove a failed projection does not falsely advance its checkpoint and bounded rebuilds converge idempotently with stable semantic identities.

Checks:

- PASS — representative isolated recovery modes converge without permanent running state, duplicate active lane, orphan Run, false checkpoint advance, or duplicate semantic event.

Resolved findings:

- `PA-003`

Open blockers:

- None.

Deferred:

- None.

## Cross-gate consistency

Status: **PASS**

- Sites `338`, `337`, and `336` were traced on `2026-08-30` from same-day Task Result and source Observation through finalized Site Fact and `tls13_adoption/1` Entity state. The sample included success/positive and failure/probe-failed behavior; none had a qualifying Change Event, which is allowed.
- Games `312`, `311`, and `310` were traced on `2026-08-30` from same-day player Task Result through finalized Game Fact and all three active Metric states. None had a qualifying Change Event.
- The full fixed and additional seeded cohorts were separately traced across both closed days for Fact identity, evidence/quality, and projection versions.
- Production event samples and isolated semantic cohorts complete the transition/event side where the fixed six entities had no natural transition.

## Findings

### PA-001

- Gate: G0–G5 and cross-gate consistency
- Class: **BLOCKER**
- Description: the initial acceptance run had no independent production read-only evidence channel.
- Evidence: the first report could only use local/isolated databases.
- Impact: G0–G5 production evidence could not initially be signed.
- Resolution: a temporary technically read-only PostgreSQL role was supplied out of band; bounded production queries were executed for `gfa`, `gfn`, and `gfg` with explicit read-only transactions and timeouts.
- Fix commit: not applicable; no code defect.
- Revalidation: PASS — all production evidence sections above completed without a write operation.

### PA-002

- Gate: final CI
- Class: **BLOCKER**
- Description: the initial candidate/report commits had not yet produced remote CI evidence.
- Evidence: the initial report recorded only local equivalent checks.
- Impact: the final accepted SHA could not initially be signed.
- Resolution: `9794c4e19a5f025c2777a68e02e14e489cb9a3e6` was pushed and `engineering-foundation` run `33400459588` completed successfully.
- Fix commit: not applicable.
- Revalidation: PASS — Foundation, all production Go modules, React Admin, Nav Web, PostgreSQL integration, repository policy, and active vulnerability jobs were green.

### PA-003

- Gate: G6
- Class: **NOTE**
- Description: expired-lease recovery existed in both collectors but initially lacked direct PostgreSQL integration regression evidence.
- Evidence: implementation called recovery at startup and on the periodic loop, but the earlier CI integration commands did not directly exercise expiry.
- Impact: recovery evidence was incomplete; no correctness defect was observed.
- Resolution: isolated Game/Nav recovery tests, readiness recovery assertions, CI commands, and policy assertions were added.
- Fix commit: `3ddf8fa46b8d540be49b308be8db62310e16bca9`
- Revalidation: PASS locally and in accepted-commit CI.

### PA-004

- Gate: repository validation
- Class: **NOTE**
- Description: Windows `core.autocrlf=true` makes a direct full-tree `gofmt -l .` list CRLF working-tree files.
- Evidence: accepted-commit CI on Ubuntu passed the repository gofmt matrix; modified Go test files were explicitly formatted and `git diff --check` passed.
- Impact: no repository or production correctness impact; a whole-repository line-ending rewrite would be out of scope.
- Resolution: none required.
- Revalidation: remote CI PASS.

### PA-005

- Gate: G3–G4
- Class: **NOTE**
- Description: repaired Nav `ipv6` and `security_txt` v2 Metric/Detector contracts start on `2026-08-31`, after the two closed-day main production window.
- Evidence: live Goose Registry/checkpoints set `source_start_date=2026-08-31` and `processed_through=NULL`; the execution occurred before that UTC day closed. Production DNS observations after cutover already contained the new explicit evidence classifications.
- Impact: no v2 production Daily/Event row was yet expected, so absence is not lag or data loss. Fabricating or backdating evidence would violate the published cutover contract.
- Resolution: retained the correct boundary, reconciled the two eligible predecessor v1 days, verified post-cutover raw evidence, and relied on the non-skipped isolated v2 semantic/rebuild suites for active behavior.
- Fix commit: not applicable; no defect.
- Revalidation: PASS — checkpoint/source boundary and isolated v2 suites agree.

## Validation summary

- Six active Go modules: `go vet ./...`, `go test ./...`, `go build ./...` — PASS.
- React Admin: `npm ci`, strict typecheck, 33 tests, production build — PASS.
- Nav Web: `npm ci`, typecheck, production build — PASS with existing non-fatal sourcemap/chunk/deprecation warnings.
- Tools: `go test ./...`, sqlc vet/generate/idempotence, `check-sqlc`, `check-production-policy` — PASS.
- PostgreSQL integrations: Game repository/Facts/Metrics, Game Changes, Game Backend, Nav Observation/Facts/Metrics, Nav Changes, Nav Backend, Admin identity/legacy/three-database, Game/Nav lease recovery — PASS without required-test skips.
- Goose: fresh, baseline adoption, released upgrades, drift rejection, seeded-row preservation, and cleanup on PostgreSQL 18.3 — PASS for `gfg`, `gfn`, and `gfa`.
- Vulnerability scan: six active Go modules — no called vulnerabilities.
- Production read-only evidence: G0–G5 and cross-gate bounded queries — PASS.
- Remote CI: accepted commit `9794c4e19a5f025c2777a68e02e14e489cb9a3e6`, run `33400459588` — PASS.

## Final summary

```text
G0 Schema & Migration      PASS
G1 Acquisition             PASS
G2 Historical Facts       PASS
G3 Metrics                 PASS
G4 Change Intelligence     PASS
G5 Admin & RBAC            PASS
G6 Recovery                PASS

Open BLOCKER findings: 0
Resolved BLOCKER findings: 2
Deferred findings: 0
Notes: 3

V3 P0 Production Acceptance: PASS
```

V3 P1 Analytics MVP is unblocked. No P1, P2, or Admin polish work was started.
