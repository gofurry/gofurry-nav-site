# Changelog

All notable changes to this project are documented in this file.

Development work that has not been released stays under `Unreleased`. Formal repository-level changelog tracking was introduced before the V3 cycle; earlier project history is reconstructed below as dated milestones from Git history rather than assigned retrospective version numbers.

## Unreleased

### Added

- Add native React Collection, Metrics, and Change Centers with capability-aware controls, count-backed history/explorers, operational charts, and technical contract views.
- Add a capability-shaped Workbench Attention projection, read-only three-database Data Operations health center, filtered/redacted Audit explorer, and Owner account-governance UI.
- Add static PostgreSQL metadata, Goose state, bounded relation-size, audit-pagination, and Workbench aggregation APIs with PostgreSQL integration coverage.
- Add the React 19 Admin foundation with strict TypeScript, Tailwind v4 semantic themes, Base UI/shadcn-style primitives, TanStack Query/Table, React Hook Form/Zod, and capability-aware navigation.
- Add schema-driven React CRUD for sayings, update notices, Site groups, tags, comments, and prizes; add first-class Site and Game content workspaces plus global Site/Game/Tag/Group search.
- Add minimal sqlc-backed Site/Game workspace read models and Site list summaries without changing authentication, collection, Metric, or Change semantics.
- Add multi-account Admin identity, database-backed Principal validation, fixed Owner/Developer/Operator capability RBAC, Owner-only account management APIs, and transaction-safe last-Owner protection.
- Add durable audit identity snapshots plus migration, authorization, session invalidation, route enforcement, and PostgreSQL concurrency coverage.
- Add `ipv6_adoption/2` and `security_txt_adoption/2` with matching v2 Change detectors while retaining published v1 contracts for explicit historical rebuild.
- Add CI-enforced PostgreSQL integration coverage for Game/Nav Change engines and Admin collection operations.

### Changed

- Validate both the target React Admin and the temporary legacy Vue Admin in CI; production embedding and cutover remain owned by P0.5.2-D.
- Change Admin bootstrap/login from a singleton password to canonical username, display name, and password while preserving the legacy account password and timestamps during migration.
- Keep JWTs role-free and resolve current role, status, session version, and capabilities from `gfa` on every authenticated request; minimally update the existing Vue login/setup compatibility surface.
- Preserve per-query AAAA evidence and validate security.txt content before treating either capability as adopted; inconclusive DNS and unrecognized documents remain unknown instead of becoming false or positive.
- Add count-backed pagination and compact detail views to Metric Daily/Entity results, Collection Run/Task history, and Collector lifecycle views.
- Replace manual collection database-ID entry with searchable Game, Site, and current Site Target selectors.

### Fixed

- Correct the React Data Operations capability from `data_ops.read` to canonical `dataops.read` and explicitly reject the legacy alias in regression tests.
- Preserve `schedule_id` and `schedule_version` on Schedule Run Now jobs without creating or moving a scheduled slot, and report unavailable coverage as null rather than 0%.
- Separate current and historical Collector instances, clarify chart timing units and spacing, and keep Task Result details within the viewport.

## v3.0.0-alpha.5 - 2026-08-30

### Added

- Add Goose-owned versioned Game/Nav change registries, deterministic canonical event stores, and independent per-detector-version checkpoints.
- Add ten compiled detectors for Game free/support/release/price transitions and Nav IPv6/TLS 1.3/security.txt/Primary Target/TLS certificate transitions.
- Add in-process Game/Nav Change Engines plus `changes status`, `changes backfill`, and forward-propagating `changes rebuild` commands.
- Add an authenticated read-only Admin Change Center for detector Registry, checkpoints, events, filters, and provenance details.

### Changed

- Run Collector reconciliation in Acquisition -> Facts -> Metrics -> Changes order while preserving the existing Redis-backed Nav Change and all public routes.
- Compare semantic states across unknown gaps only within historical tracking identities, and retain deterministic event time, scope, source keys, versions, and materialization provenance.

## v3.0.0-alpha.4 - 2026-08-29

### Added

- Add Goose-owned versioned Game/Nav metric registries, explainable entity-daily state, global and single-dimension daily counts, and independent per-metric-version checkpoints.
- Add the first six compiled metrics: Game free share and Windows/Linux support plus Nav IPv6, TLS 1.3, and security.txt adoption.
- Add in-process Game/Nav Metric Engines with registry drift guards and `metrics status`, `metrics backfill`, and `metrics rebuild` commands.
- Add an authenticated read-only Admin Metric Center for Registry, checkpoint, daily aggregate, and historical entity inspection.

### Changed

- Run Collector reconciliation in Acquisition -> Facts -> Metrics order and gate metric days on finalized Historical Fact watermarks.
- Compute historical freshness against the UTC Fact day end, persist seven-state reasons and provenance, and calculate adoption/coverage only at query time.

### Fixed

- Use release evidence for `free_game_share` availability outcomes, reject future release/details evidence atomically, and require compiled evaluators for both active and retired Registry versions.

## v3.0.0-alpha.3 - 2026-08-29

### Added

- Add effective-dated Game/AppID, Nav target, and Primary Target periods; UTC hourly/daily Game Player facts; historical Game/Price facts; Nav protocol/target/Site facts; and ordered fact checkpoints.
- Add in-process Game/Nav Fact Engines plus `facts status`, `facts backfill`, and `facts rebuild` commands with dry-run support and shared runtime/backfill projection paths.
- Add minimal Admin Primary Target selection with audit, current-day Game/Site fact write-through, code-whitelisted Nav known-state projection, and checkpoint-gated raw retention feature flags.

### Changed

- Allocate `gfg_game.id` from a PostgreSQL sequence seeded above every durable current, Raw, ledger, tracking, and Fact Game-ID source; add the four-state current price contract, enforce one Player Raw row per durable Run/Game, preserve release history across current Game deletion, and record `gfn_site.deleted_at`.
- Make Game create/AppID-change/delete, Nav target identity mutations, Site deletion, and Primary replacement update historical eligibility in the same business transaction.
- Allow Nav Site Daily `finalized_at` to remain null only for the mutable current-day Admin marker; closed UTC days are finalized by the Site pipeline.
- Show canonical release and planned-release dates on Game homepage cards using the same hover presentation as Game search results.

### Fixed

- Separate scheduled acquisition quality from Fact values, retain unknown as nullable, exclude manual Player samples, keep manual Nav success out of scheduled quality, and prevent failed observations from clearing last-known structured state.
- Replace the P0.1 raw-pruning freeze with disabled-by-default, post-checkpoint Game age retention and target-aware Nav keep-count retention.
- Keep the SVG inside Game home/search review buttons on a fixed pixel box while cards reveal the controls, preventing Chromium hover resampling jitter.
- Treat empty or malformed historical Nav TLS certificate timestamps as unknown so target Fact backfill cannot abort with `SQLSTATE 22007`.
- Align the Nav Backend PostgreSQL integration fixture with the historical Site deletion timestamp invariant.

## v3.0.0-alpha.2 - 2026-08-28

### Added

- Add PostgreSQL-backed Game/Nav schedules, durable jobs, run attempts, per-target results, collector instances, heartbeats, leases, cancellation, recovery, and Redis realtime progress.
- Add the Admin Collection Center for schedule control, queue/history, manual Game/Nav fan-out, retry/cancel, audit, and ECharts outcome/coverage/timing views.

### Changed

- Rename active Game V2 physical storage objects to unversioned canonical names while preserving existing data and legacy collection history.
- Replace process-relative collector scheduling with fixed cron/anchored intervals, explicit `scheduled_for`, `skip`/`catch_up_once` misfire policy, priority lanes, and PostgreSQL `FOR UPDATE SKIP LOCKED` claims.
- Route scheduled, manual, and Game entity-triggered acquisition through the same durable execution path; Redis now holds realtime progress only.
- Freeze destructive player-count and Nav observation retention pending P0.2.
- Retain temporary Game/Nav collection task results for 90 days while preserving durable Job/Run history.

### Removed

- Remove `go-timewheel` from active production modules, old Redis collection command/lease paths, and Game/Nav Backend collection proxy endpoints.

### Fixed

- Preserve stable schedule phase across collector restarts, recover expired worker leases, protect concurrent lanes at the database layer, and keep scoped Nav ping collection from pruning global Redis results.
- Clarify Admin schedule editing with field guidance, validation, database/browser/UTC clock comparison, timezone-aware previews, and a corrected modal layout.
- Let Go CI fall back to direct module downloads when the public module proxy has a transient transport failure.

## v3.0.0-alpha.1 - 2026-08-27

### Added

- Add unversioned canonical Game release-state, first-available, release-history, and language tables with Goose/sqlc contracts.
- Add the `backfill-first-available` Collector command with dry-run reporting and write-once legacy-manual migration.
- Add structured release, first-available, and language fields to the existing `/api/v2/game/*` responses and canonical formatting in Nav Web.
- Add a typed Game Collector option to control whether startup immediately triggers a players collection.

### Changed

- Normalize canonical release and language facts only from the US/English Steam response while preserving non-authoritative observations.
- Make Latest Games and release-range search use First Available calendar windows instead of legacy release strings.
- Keep Recently Collected independent of release status, and add mutually exclusive released/upcoming search with canonical planned-release range filtering and ordering.
- Show canonical release information as a hover overlay on Game search-card artwork without changing the original card layout.
- Give pending single-game collection priority over startup players, scheduled players, and daily/manual full collection, without letting players cooldown block onboarding work.
- Stop Admin and Backend runtime reads/writes of `gfg_game.release_date`; AppID changes now reset Steam-derived state transactionally and re-enqueue collection after commit.

## v2.4.0 - 2026-08-23

### Added

- Add audited Goose baselines and versioned migrations for the `gfg`, `gfn`, and `gfa` PostgreSQL databases, with fresh-database, adoption, drift, upgrade, cleanup, and integration verification.
- Add root sqlc configuration, service-local generated query packages, bounded pgxpool wiring, pinned DB tooling, and generated-code drift checks.
- Add explicit Cobra/Viper CLI lifecycles for active Go services with `serve`, `version`, Linux/systemd `install`, and `uninstall` commands.
- Add the standalone `gf-uptime` service with Fiber uptime monitoring, durable local Bbolt history, `/livez`, `/readyz`, and independent status-page ownership.
- Add optional private `/livez` and `/readyz` endpoints to both Collectors, plus Nav Web `/healthz` and an independent public status URL.
- Add scheduled `govulncheck` scanning for all six active Go applications.
- Add compact Agent-oriented engineering context through `AGENTS.md`, `.agents/*`, stable `contracts/*`, accepted ADRs under `docs/decisions`, and refreshed PR/Issue templates.

### Changed

- Migrate Game Collector, Game Backend, Nav Collector, Nav Backend, and Admin persistence from GORM to `pgx/v5`, bounded `pgxpool`, and static sqlc SQL.
- Make Goose under `db/game`, `db/nav`, and `db/admin` the sole schema source of truth; production applications no longer run schema migration or DDL at startup.
- Make Admin PostgreSQL-only with explicit `gfa`, `gfn`, and `gfg` pools and no implied cross-database transaction.
- Reorganize active production code under `apps/cn`; keep `apps/intl` placeholder-only and exclude `legacy`, `experimental`, and `third-party` from active build, CI, vulnerability scanning, sqlc, and deployment.
- Rewire `build.bat`, CI, sqlc tooling, repository policy checks, documentation, and deployment paths around the active monorepo topology.
- Standardize active Go applications and repository tooling on Go 1.26.7.
- Standardize active Go logging/config conventions around Zap, Lumberjack, Viper, and YAML v3.
- Replace `kardianos/service` with normal foreground processes and deployment-specific systemd units generated from the deployed binary, working directory, runtime user, and explicit config.
- Move availability ownership out of Nav Backend into `gf-uptime`; business services no longer depend on the observer.
- Move operational maintenance/WAF assets under `ops` and keep custom Coraza rules explicit.
- Expand CI to enforce production dependency boundaries, PostgreSQL 18 migration behavior, sqlc generation consistency, integration semantics, and active-module security checks.
- Preserve public API compatibility, Nav V1 reachability, Redis key/cache meaning, Collector scheduling semantics, and established DB-commit-before-cache-refresh behavior during the engineering migration.

### Removed

- Remove GORM, production `AutoMigrate`, global ORM database state, generic ORM DAO infrastructure, and the historical GORM CodeGen tool from active production modules.
- Remove Swagger/swag routes, middleware, annotations, generated API tooling, and related active dependencies.
- Remove Logrus, `github.com/pkg/errors`, YAML v2, Admin MySQL/SQLite support, and other obsolete production dependencies.
- Remove `github.com/kardianos/service` from active Go applications.
- Remove historical executable migration SQL after Goose adoption; Git history remains the archive.
- Remove deprecated Game tables `gfg_game_creator_deprecated_20260614`, `gfg_game_record`, `gfg_game_news`, and `gfg_game_player_count`.
- Remove deprecated Nav table `gfn_log_update`.
- Remove embedded uptime ownership from Nav Backend.
- Remove the vendored OWASP CRS tree, stale CRS auto-discovery, obsolete root-level operational artifacts, and retired build targets.

### Fixed

- Fix Game listing and tag-mapping behavior found during pre-V3 stabilization.
- Align Goose baselines and schema metadata with audited production PostgreSQL dumps.
- Fix generated systemd unit rendering for deployment working directories and command arguments.

### Security

- Raise all active Go applications and tooling to Go 1.26.7.
- Add weekly/manual `govulncheck` coverage for all six active Go services.
- Add executable production policy checks rejecting GORM, startup schema migration, Logrus, `pkg/errors`, YAML v2, Swagger/swag, `kardianos/service`, arbitrary Viper `AutomaticEnv`, and active dependencies on `legacy` or `experimental`.
- Keep Collector health endpoints private by default and separate external-target failure from local process readiness.

---

## Historical milestones

These entries summarize notable project history before repository-level changelog tracking was introduced.

## v2.3.1 - 2026-08-06

### Added

- Add `/livez` and `/readyz` health endpoints to production backend services.
- Add centralized uptime monitoring and a status dashboard inside Nav Backend.
- Add Steam game metadata prefill support to Admin.

### Changed

- Redesign the Admin management workspace and improve Game/Nav operational workflows.
- Continue production dependency updates and CI cleanup around configuration-sensitive tests.

### Fixed

- Resolve Game ingestion, listing, frontend, and related Admin regressions before the August production merge.

## v2.3.0 - 2026-07-11

### Added

- Add single-game collection queuing and additional game statistics/online-state capabilities.
- Expand application observability through the maintained monitor middleware and related operational surfaces.

### Changed

- Rework the Steam-facing frontend area toward the game/workshop direction.
- Refine Nav, Game, site-detail, and general frontend presentation.
- Archive the former Ops Agent / Center services and remove them from the active production path.
- Refresh Admin tag-option behavior and backend dependencies.

## 2026-06-25 - Monitor correctness hardening

### Fixed

- Correct unsafe Fiber monitor integration that could retain request state and cause memory growth.
- Harden monitor middleware usage in the production Nav path.

### Changed

- Continue frontend navigation, theme, image, and locale refinements around the stabilized monitor integration.

## v2.2.0 - 2026-06-20

### Added

- Complete the main Game V2 public API path, including details, search, tags, reviews, recommendations, prize flow, panel data, and Collector status/observability.
- Add Steam-backed Game V2 collection for details, news, player counts, assets, prices, snapshots, and task/run reporting.
- Add application error and maintenance fallback pages.

### Changed

- Cut main Game frontend pages over to V2 APIs and remove legacy Game V1 dynamic packages/routes from the active path.
- Consolidate Nav V2 read models and site-detail data, including summary, latest observations, trends, change events, and light probes.
- Move site ordering/weight behavior into group mappings.
- Improve Game search/detail/home cache behavior and page performance.
- Archive retired legacy modules while keeping production Nav V1 compatibility where still required.

### Fixed

- Correct Game search/news ordering, cache refresh behavior, and test-time config/DB initialization.
- Correct representative page/cache regressions discovered during the V2 cutover.

## 2026-06-08 - Game V2 foundation and frontend performance

### Added

- Introduce `steam-go` into the repository workflow and build the Game Collector V2 Steam client/storage foundation.
- Add Game V2 details, news, player-count, reporting, and rate-control collection paths.
- Add the Game Backend V2 contract and roadmap.
- Add frontend performance regression guards.

### Changed

- Reduce expensive frontend rendering work, lazy-load heavier dependencies, and split initial home rendering.
- Continue the visual migration toward the newer grid/Less-based frontend style system.

## v2.0.0 - 2026-06-06

### Added

- Add Nav V2 home bootstrap, site detail, health summary, search suggestions, and structured bilingual update notices.
- Add Admin management for bilingual update notices.
- Add `llms.txt` and security metadata entrypoints.
- Add spotlight panels and richer public site information.

### Changed

- Promote the redesigned V2 site-detail page as the primary public detail experience.
- Redesign the updates page into a structured bilingual timeline.
- Improve SEO metadata, sitemap output, image alt text, dark mode, and navigation interactions.

## 2026-05-26 - Nav Collector V2 data plane

### Added

- Introduce the Nav Collector V2 observation data plane and domain-based target model.
- Add enriched Ping, HTTP, TLS, and DNS observation payloads.
- Add HTTP security/header/page metadata summaries.
- Add light probes for page assets, ports, edge-provider hints, and WAF canaries.
- Add health summaries, normalized reason contracts, target relation hints, trends, and change events.
- Add the Nav Backend V2 Collector read model and summary/detail endpoints.

### Changed

- Move Nav monitoring from isolated legacy records toward a unified observation/read-model flow while preserving necessary legacy compatibility.
- Add single-instance collection governance and strengthen Collector configuration/test boundaries.

## 2026-05-20 - RAG, lightweight Ops, WAF, and bilingual web work

### Added

- Build the experimental GoFurry RAG service with document ingestion, retrieval debugging, JWT console access, AI chat, citations, source synchronization, and Ollama/Tencent model integration.
- Add the public Archive knowledge-chat experience and contextual Ask entrypoints.
- Add the first lightweight Ops Agent and Ops Center with node/service collection, embedded dashboard, deployment assets, and operational metrics.
- Add Coraza/CRS experimentation and custom SecLang WAF rules.
- Add the placeholder international-site workspace.
- Improve bilingual SEO metadata, theme controls, and Nav interactions.

### Changed

- Normalize the GoFurry namespace and connect Game/Nav content sources into the experimental RAG sync flow.
- Expand root build/CI coverage for the then-active RAG and Ops components.

> The RAG service and former Ops Agent / Center were later retired from the active runtime and are preserved only under `legacy`.

## 2026-05-01 - Nuxt frontend migration

### Added

- Scaffold the Nuxt frontend that became the production public site.
- Add public assets, head/SEO configuration, search/navigation UI, and the initial modern frontend shell.
- Add Docker-oriented production deployment documentation and an update script.
- Add bilingual repository documentation.

### Changed

- Begin replacing the older Vue public frontend with Nuxt for improved SEO and server-rendered public pages.
- Establish the frontend deployment path that later became the production Nav Web workflow.

## 2026-04-12 - Public monorepo foundation

### Added

- Add the first CI workflow and release/build scripts.
- Add repository safety ignores for local configuration.
- Add the public repository README.
- Add the initial `gofurry-admin` service and operations/admin entrypoint.

### Changed

- Format and stabilize the initial Admin sources and bulk-mapping workflow.
- Establish the repository as a public multi-service GoFurry monorepo rather than a single application.

## 2026-03-31 - Project creation

### Added

- Create `gofurry/gofurry-nav-site`.
- Add the initial repository commit that became the base for the public GoFurry navigation-site monorepo.
