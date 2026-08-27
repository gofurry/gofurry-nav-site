# Changelog

All notable changes to this project are documented in this file.

Development work that has not been released stays under `Unreleased`. Formal repository-level changelog tracking was introduced before the V3 cycle; earlier project history is reconstructed below as dated milestones from Git history rather than assigned retrospective version numbers.

## Unreleased

### Added

- Add PostgreSQL-backed Game/Nav schedules, durable jobs, run attempts, per-target results, collector instances, heartbeats, leases, cancellation, recovery, and Redis realtime progress.
- Add the Admin Collection Center for schedule control, queue/history, manual Game/Nav fan-out, retry/cancel, audit, and ECharts outcome/coverage/timing views.

### Changed

- Rename active Game V2 physical storage objects to unversioned canonical names while preserving existing data and legacy collection history.
- Replace process-relative collector scheduling with fixed cron/anchored intervals, explicit `scheduled_for`, `skip`/`catch_up_once` misfire policy, priority lanes, and PostgreSQL `FOR UPDATE SKIP LOCKED` claims.
- Route scheduled, manual, and Game entity-triggered acquisition through the same durable execution path; Redis now holds realtime progress only.
- Freeze destructive player-count and Nav observation retention pending P0.2.

### Removed

- Remove `go-timewheel` from active production modules, old Redis collection command/lease paths, and Game/Nav Backend collection proxy endpoints.

### Fixed

- Preserve stable schedule phase across collector restarts, recover expired worker leases, protect concurrent lanes at the database layer, and keep scoped Nav ping collection from pruning global Redis results.

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
