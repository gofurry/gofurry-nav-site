# V3 P0 Production Acceptance

## Acceptance header

- Final status: **BLOCKED**
- Initial audited commit: `36ab41f79860c50b00d0c842ca960fb84c9c0da0`
- Acceptance candidate commit: `3ddf8fa46b8d540be49b308be8db62310e16bca9`
- Accepted commit SHA: **Not assigned while a BLOCKER remains open**
- Execution date: `2026-08-31`
- Production acceptance window: **Not established; no production read-only connection was available**
- Production databases checked: **None**
- Production read-only evidence sources used: **None**
- Remote CI evidence: `engineering-foundation #150` passed for the initial commit `36ab41f`
- Candidate CI evidence: full local equivalent checks passed; remote CI is unavailable because `3ddf8fa` has not been pushed

This report does not treat the configured private-LAN `debug` databases as production. No production mutation, migration, DDL, collection control, rebuild, backfill, account, permission, password, or session operation was performed.

## Acceptance cohort

The following candidate cohort was fixed from bounded, read-only local Admin GET responses so the same identities can be used when production read-only access is supplied. It is not production evidence.

- Fixed Nav Site IDs (12): `296, 295, 294, 293, 292, 291, 290, 289, 288, 287, 286, 285`
- Fixed Game IDs (12): `307, 306, 305, 304, 303, 302, 301, 300, 299, 298, 297, 296`
- Random Nav Site sample (seed `503`, 3): `285, 265, 292`
- Random Game sample (seed `503`, 3): `296, 276, 303`
- Local integration semantic cohort:
  - Nav Site IDs `99200`–`99208` cover positive, negative, stale, not-probed, probe-failed, unknown, not-applicable, confirmed-absence, and unavailable-evidence behavior.
  - Game IDs `99200`–`99205` cover positive, negative, stale, unknown, not-applicable, platform support, and missing evidence.

Selection rationale: the fixed cohort is stable and bounded; the seeded random sample reduces hand-selection bias; the synthetic integration cohort explicitly exercises frozen state semantics. Production membership, two-day evidence, and cross-gate traces remain unverified under `PA-001`.

## G0 — Schema & Migration

Status: **BLOCKED**

Scope:

- Current and reproducible `gfg`, `gfn`, and `gfa` Goose schemas.
- Critical Collection, Fact, Metric, Change, account, audit, and authorization schema invariants.
- Runtime DDL/startup migration prohibition.

Evidence:

- `TestPostgresFreshAndBaselineAdoption` passed on PostgreSQL `18.3`, including fresh, baseline adoption, released upgrade boundaries, drift rejection, seeded row preservation, and cleanup for all three databases.
- Repository expected versions are `gfg=20260830010000`, `gfn=20260830020000`, and `gfa=20260830030000`.
- Bounded local Data Operations GET evidence reported all three development databases healthy, at the expected version, with `pending_count=0`.
- `go tool sqlc vet`, `go tool sqlc generate`, generated-diff validation, `check-sqlc`, and `check-production-policy` passed.
- Active Go source contains no `AutoMigrate`, startup Goose invocation, or runtime `CREATE/ALTER TABLE` pattern.

Checks:

- PASS — isolated fresh/adoption/upgrade/drift and critical schema assertions.
- PASS — Goose remains the sole schema owner; sqlc generation is idempotent.
- BLOCKED — production PostgreSQL versions, sizes, Goose state, pending count, and critical live invariants were not read.

Open blockers:

- `PA-001`

Deferred:

- None.

## G1 — Acquisition

Status: **BLOCKED**

Scope:

- Schedule → Job → claim → Run → Task Result lineage.
- Run Now, retry, cancel, collector instance history, leases, lanes, and misfire semantics.

Evidence:

- `TestAdminThreeDatabasePersistence` passed without skipping and verified schedule Run Now lineage, manual semantics, cancel/retry eligibility, one active dedupe entry, collector current/history views, count-backed Run/Result pagination, and null coverage when `expected_count=0`.
- Game/Nav schedule tests passed for anchored phase across restart, fixed cron, point-in-time miss, and `catch_up_once` behavior.
- `TestGameExpiredLeaseRecoveryIntegration` and `TestNavExpiredLeaseRecoveryIntegration` passed against isolated randomly named databases.
- The recovery tests prove expired Job/Run pairs become terminal `failed/worker_lost`, leases are cleared, the lane is released, a replacement instance claims exactly one queued Job, and no orphan Run is created.

Checks:

- PASS — isolated acquisition, retry/cancel, misfire, concurrency, and lease recovery semantics.
- BLOCKED — no recent production Schedule/Job/Run/Result or collector heartbeat evidence was read.

Open blockers:

- `PA-001`

Deferred:

- None.

## G2 — Historical Facts

Status: **BLOCKED**

Scope:

- Closed UTC-day projection, immutable history, tracking identity, evidence/value separation, checkpoints, retention safety, and deterministic rebuild.

Evidence:

- `TestPostgresRepositorySemantics` passed for Game Facts, including projection, idempotent rebuild, missing-source rejection, checkpoint-safe retention, and preservation of canonical history.
- `TestPostgresCollectorPersistenceSemantics` passed for Nav Facts, including scheduled quality lineage, immutable target/site history, idempotent target/site rebuild, and checkpoint-gated retention.
- Unit and integration evidence confirms AAAA query failure remains unavailable/unknown and malformed, HTML, empty, or invalid security.txt evidence does not become positive.

Checks:

- PASS — isolated Fact truthfulness, tracking, checkpoint, retention, and rebuild invariants.
- BLOCKED — no production two-closed-day cohort trace or production Fact invariant query was executed.

Open blockers:

- `PA-001`

Deferred:

- None.

## G3 — Metrics

Status: **BLOCKED**

Scope:

- Every active Metric version, state/denominator/coverage arithmetic, checkpoints, lineage, and deterministic rebuild.

Evidence:

- Actual local Registry status exposed active Game `free_game_share/1`, `windows_support/1`, `linux_support/1`; active Nav `ipv6_adoption/2`, `tls13_adoption/1`, `security_txt_adoption/2`; and retained retired Nav v1 contracts.
- Game and Nav PostgreSQL semantic suites passed state classification, provenance, count conservation, all-zero/global behavior, null denominator behavior, checkpoint serialization, future-evidence rollback, retired-version explicit rebuild, and deterministic active-version rebuild.
- `unknown`, `probe_failed`, `not_probed`, `stale`, and `not_applicable` remained distinct from `negative` in the integration cohort.

Checks:

- PASS — isolated Registry/catalog agreement, arithmetic, checkpoint, and rebuild invariants for active versions.
- BLOCKED — no production daily/entity reconciliation for two consecutive closed UTC days was performed.

Open blockers:

- `PA-001`

Deferred:

- None.

## G4 — Change Intelligence

Status: **BLOCKED**

Scope:

- Every active Detector version, transition semantics, stable identity, time/provenance, checkpoints, no duplicates, and deterministic rebuild.

Evidence:

- Actual local Registry status exposed five active Game detectors and active Nav `ipv6_transition/2`, `security_txt_transition/2`, `tls13_transition/1`, `primary_target_transition/1`, and `tls_certificate_transition/1`, while retaining the retired Nav v1 contracts.
- `TestGameChangeProjectorsIntegration`, `TestGameChangeCheckpointLockSerializes`, `TestGameChangeEngineBackfillRebuildIntegration`, `TestNavChangeProjectorsIntegration`, and `TestNavChangeEngineBackfillRebuildIntegration` all passed without skipping.
- The suites cover semantic memory across unknown gaps, tracking identity, deterministic event keys, duplicate prevention, event time/provenance, checkpoint bounds, forward-propagating rebuild, and stable event sets.

Checks:

- PASS — isolated active Detector contract, identity, time, checkpoint, and rebuild invariants.
- BLOCKED — no bounded production duplicate/event trace or production detector lag evidence was read.

Open blockers:

- `PA-001`

Deferred:

- None.

## G5 — Admin & RBAC

Status: **BLOCKED**

Scope:

- Owner, Developer, and Operator capability behavior in backend and React UI.
- Read-only production Admin smoke and agreement with database evidence.

Evidence:

- `TestAdminIdentityAuthorizationPersistence`, `TestAdminLegacyIdentityUpgrade`, and `TestAdminThreeDatabasePersistence` passed without skipping.
- Integration evidence covers bootstrap/login/current Principal, session invalidation, account authorization, audit identity, concurrent last-active-Owner protection, Collection mutations in isolation, and Metric/Change/Admin read models.
- Local read-only API smoke returned:
  - Developer: Workbench, Collection, Metrics, Metric Registry, Changes, Change Registry, DataOps, Audit `200`; Accounts `403`.
  - Operator: Workbench, Collection, Metrics, Changes `200`; Metric Registry, Change Registry, DataOps, Audit, Accounts `403`.
- React strict typecheck, 10 test files / 33 tests, production build, embedded nested-route/API boundary tests, capability UX tests, and theme tests passed.

Checks:

- PASS — isolated backend capability policy, direct API denials, account governance, and React capability-aware behavior.
- BLOCKED — no production React/auth/read-page smoke or production Admin-to-DB evidence comparison was performed.

Open blockers:

- `PA-001`

Deferred:

- None.

## G6 — Recovery

Status: **PASS**

Scope:

- Collector restart/lease expiry, PostgreSQL dependency visibility, retry, misfire reconciliation, checkpoint lag, and Facts/Metrics/Changes convergence.

Evidence:

- New Game/Nav expired-lease integration tests prove durable terminal recovery, lane release, single replacement claim, and absence of orphan Runs.
- Collector health tests prove readiness transitions from ready → dependency unavailable → ready after dependency recovery, while liveness remains independent.
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

Status: **BLOCKED**

- The synthetic integration suites trace acquisition quality into Facts, Facts into Metric Entity/Daily outputs, and Metric/Fact/domain history into deterministic Change Events; Admin integration separately verifies historical names and read-model presentation.
- A production trace for at least three Sites and three Games across the full chain was not possible without `PA-001`.

## Findings

### PA-001

- Gate: G0–G5 and cross-gate consistency
- Class: **BLOCKER**
- Description: no independent production read-only DSN, bounded query channel, or authenticated read-only Admin session was available in the execution environment.
- Evidence: only ignored local `debug` configurations were present; all database endpoints were private-LAN development endpoints. No production-specific environment variable was configured.
- Impact: production schema agreement, two consecutive closed UTC days, production acquisition lineage, Fact/Metric/Change reconciliation, fixed/random production cohort, Admin production smoke, and cross-gate production traces cannot be signed.
- Resolution: provide a technically read-only production connection for `gfa`, `gfn`, and `gfg`, or a reviewed bounded read-only evidence export plus safe Admin GET access. Then execute G0–G5 production checks without mutations.
- Fix commit: not applicable.
- Revalidation: pending.

### PA-002

- Gate: final CI
- Class: **BLOCKER**
- Description: the acceptance candidate `3ddf8fa46b8d540be49b308be8db62310e16bca9` exists only on local `dev`; pushing was not authorized, so GitHub Actions has not run for that SHA.
- Evidence: local branch was one commit ahead of `origin/dev` after the recovery-test commit.
- Impact: the required “final commit CI green” condition is not met even though the complete local equivalent passed.
- Resolution: push the candidate/report commits, wait for `engineering-foundation`, and record the successful run before final sign-off.
- Fix commit: not applicable.
- Revalidation: pending.

### PA-003

- Gate: G6
- Class: **NOTE**
- Description: expired-lease recovery existed in both collectors but lacked direct PostgreSQL integration regression evidence.
- Evidence: implementation called `RecoverExpired` at startup and every 30 seconds, but the prior CI integration commands did not execute a lease-expiry recovery test.
- Impact: G6 evidence was incomplete; no product correctness defect was observed.
- Resolution: added isolated Game/Nav recovery tests, readiness recovery assertions, CI commands, and policy assertions.
- Fix commit: `3ddf8fa46b8d540be49b308be8db62310e16bca9`
- Revalidation: PASS locally for both new non-skipped integration tests, both collector suites, builds, and repository policy.

### PA-004

- Gate: repository validation
- Class: **NOTE**
- Description: Windows `core.autocrlf=true` makes a direct full-tree `gofmt -l .` list CRLF working-tree files.
- Evidence: GitHub Actions `engineering-foundation #150` on Ubuntu passed the repository gofmt matrix for `36ab41f`; all modified Go test files were explicitly formatted and `git diff --check` passed.
- Impact: no repository or production correctness impact; avoid a scope-expanding whole-repository line-ending rewrite during acceptance.
- Resolution: none required in P0.5.3.
- Revalidation: remote initial CI PASS; targeted candidate formatting PASS.

## Validation summary

- Six active Go modules: `go vet ./...`, `go test ./...`, `go build ./...` — PASS.
- React Admin: `npm ci`, strict typecheck, 33 tests, production build — PASS.
- Nav Web: `npm ci`, typecheck, production build — PASS with existing non-fatal sourcemap/chunk/deprecation warnings.
- Tools: `go test ./...`, sqlc vet/generate/idempotence, `check-sqlc`, `check-production-policy` — PASS.
- PostgreSQL integrations: Game repository/Facts/Metrics, Game Changes, Game Backend, Nav Observation/Facts/Metrics, Nav Changes, Nav Backend, Admin identity/legacy/three-database, Game/Nav lease recovery — PASS without required-test skips.
- Goose: fresh, baseline adoption, released upgrades, drift rejection, seeded-row preservation, cleanup on PostgreSQL 18.3 — PASS for `gfg`, `gfn`, and `gfa`.
- Vulnerability scan: six active Go modules — no called vulnerabilities.
- Remote CI: `engineering-foundation #150` — PASS for initial audited commit `36ab41f`.
- Candidate/final remote CI: BLOCKED by `PA-002`.

## Final summary

```text
G0 Schema & Migration      BLOCKED
G1 Acquisition             BLOCKED
G2 Historical Facts       BLOCKED
G3 Metrics                 BLOCKED
G4 Change Intelligence     BLOCKED
G5 Admin & RBAC            BLOCKED
G6 Recovery                PASS

Open BLOCKER findings: 2
Resolved BLOCKER findings: 0
Deferred findings: 0
Notes: 2

V3 P0 Production Acceptance: BLOCKED
```

P1 is not unblocked. Re-run the production-only evidence portions after resolving `PA-001`, then push and verify the final candidate under `PA-002`. No P1, P2, or Admin polish work was started.
