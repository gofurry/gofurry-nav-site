# Changelog

All notable changes to this project are documented in this file.

Development work that has not been released stays under `Unreleased`. Formal repository-level changelog tracking was introduced before the V3 cycle; earlier project history is reconstructed below as dated milestones from Git history rather than assigned retrospective version numbers.

## Unreleased

### Added

- Add an opt-in Admin daily EdgeOne main-host purge configured by YAML time/timezone, with system intent/outcome audit, GFA session-lock and same-slot audit deduplication, bounded runtime shutdown, and no startup catch-up or automatic retry (#122).
- Add Random cloud, independent desktop/mobile Fixed cloud, and Local folder Hero sources in Home preferences, with SSR-readable ID cookies, a paginated public metadata catalog, lazy single-image previews, and Save/Cancel with no blank-first transition (#112).
- Add a bilingual Resource routing tab to page preferences with Auto/EdgeOne/Cloudflare modes for GoFurry assets and Auto/China/Global modes for Steam assets, per-route diagnostics, manual retesting, route details, and keyboard navigation across all three tabs (#107).
- Add explicit Admin Tag Category/Tag management with immutable codes, database-generated Tag IDs, archive/restore confirmations and usage guards, plus atomic Game classification saves (#116).
- Add a public Game tag-category endpoint and a private `recommendations rebuild --config <file>` command that rebuilds up to 64 recommendations per eligible Game without starting HTTP, collectors, or Redis (#116).

### Changed

- Relicense GoFurry-owned source code and project documentation from MIT to BSD-3-Clause.
- Add repository licensing, contribution, conduct, and security guidance.
- Route ideas and architecture proposals to GitHub Discussions and clarify website content-license boundaries.
- Separate saved resource routing modes from probe recommendations with SSR-readable cookies, 12-hour probe freshness, a 60-second manual cooldown, and fixed Valve thumbnail probes; migrate old Steam automatic preferences as recommendations rather than pinned modes, and retain runtime fallback for pinned routes (#107, #121).
- Remove ordinary Game content hover shadows and share/load-more hover movement while preserving original hover colors and shadows on actual floating UI (#118).
- Replace the legacy Game tag hierarchy with `gfg_tag_category`, `gfg_tag`, and `gfg_game_tag`; preserve existing leaf IDs using the reviewed 214-entry ID/code mapping and backfill primary/secondary assignments missing from the old map. Relation roles become the sole current classification source, with compatibility response fields derived from them (#116).
- Upgrade recommendations to `similar-v2.4.0-hybrid-cbf`, using category codes and relation roles instead of numeric Tag IDs/prefixes for weighting; serialize recomputation with classification changes to prevent stale cache writes (#116).
- Use explicit categories for Nav Web tag filtering and `code=adult` for Adult detection, preserving selected leaf tags through search changes (#116).
- Advance current Game Daily projection to version 2 and preserve source projection versions during daily finalization, without rewriting historical Facts or Analytics dimensions (#116).

### Fixed

- Snapshot CDN providers and Steam candidate chains per resource so automatic/manual probes and saved route changes never replace loaded image URLs; new keys/sources adopt the latest policy, and real loading failures still advance fallback. Include Gallery video posters in the same behavior (#121).
- Match Game group placeholder rating heights to real cards so accumulated row-height differences no longer clip the bottom cards on small and medium screens; preserve pagination overflow and existing hover styling.
- Restore debounced primary/secondary tag search and selection-preserving local tag filtering in React Admin; honor tag option pagination and keyword filtering (#113).
- Use the public frontend platform icon key catalog for Admin Game group/link selectors, prevent duplicate choices, and preserve free-text resource keys (#114).
- Widen Chinese and English Game summaries to 400 characters through a new Goose migration, with matching backend/Admin validation and character counts (#115).
- Constrain the Game detail flex column and Gallery media to their parent width while preserving local thumbnail scrolling, the desktop sidebar, and mobile tabs (#120).
- Keep Admin classification audit snapshots compatible with both Go 1.26.7 and 1.27.1 formatting rules, fixing the pinned CI `gofmt` failure without changing audit payloads.

### Removed

- Remove News carousel edge decorations while retaining overflow clipping, pagination, and progress indicators (#118).
- Remove current `gfg_tag.prefix`, `gfg_tag_map`, and physical `gfg_game.primary_tag` / `secondary_tag` columns, along with the legacy Admin Tag Map editor/API; retain historical tag dimensions and derive supported primary/secondary response fields from the new relations (#116).

### Upgrade notes

- Rebuild and redeploy Admin for scheduled EdgeOne main-host purge (#122). Scheduling remains disabled until explicitly enabled in YAML; production may use `05:30 Asia/Shanghai`. No frontend update, new dependency or database migration is required. See [Admin CloudOps operations](docs/operations/admin-cloudops.md).
- Deploy Nav Backend before Nav Web for Hero source selection (#112). This adds read-only Hero catalog and request-time fixed-ID resolution; no database migration, new dependency, or environment variable is required.
- The #118 visual cleanup, responsive Game card clipping fix, and #107/#121 resource routing changes require only rebuilding and redeploying Nav Web. They introduce no backend/Admin update, database migration, dependency, or environment-variable requirement; earlier Tag domain upgrade requirements below remain separate.
- The irreversible `20260916020000_game_tag_domain.sql` migration requires a verified backup and coordinated updates to Game Backend, Game Collector, Admin with its embedded React build, and Nav Web. Stop old runtime clients before migrating; rollback requires restoring the database and compatible binaries rather than running Goose Down. The migration clears old recommendation rows; rebuild them with the new private command (#116).
- Record isolated regressions, populated development-clone acceptance, and the authorized shared-development migration in [the Tag domain acceptance record](docs/acceptance/issue-116-local-tag-domain.md).

## v3.0.0-alpha.8 - 2026-09-15

### Added

- Add group-centric Admin homepage curation with an ordered Top 8 preview and remaining members; persist existing group-map weights with revision checks, transaction locking, audit, and cache invalidation while preserving site-level group editing and bulk-replacement weights (#101).
- Add managed Site icons, independent desktop/mobile AVIF Hero pools, and a bilingual SVG pattern catalog with Admin publishing, previews, capabilities, and audit. COS is Primary; R2 is a best-effort Mirror whose failure returns a warning without blocking Primary publication (#104, #85).
- Add the Goose-owned `gfn_home_hero_asset` and `gfn_background_pattern` tables, generated sqlc queries, managed object keys in `gfn_site.icon`, Home schema v4, and the public pattern catalog API.
- Add the Admin Cloud Resources workspace for COS/R2 status, object inspection, COS-to-R2 repair, scoped EdgeOne/Cloudflare cache purges, and EdgeOne task history; reserve full-zone purge for its separate Owner capability and endpoint (#93).
- Add an SSR-aware managed-asset CDN resolver with verified background probes, a 12-hour CDN preference cookie, request-failure fallback, and bundled defaults while preserving the independent Steam CDN system (#104).
- Add public background preferences for bundled/server patterns and browser-local images, with theme-specific appearance overrides and IndexedDB file persistence; local files never upload and SSR starts from the bundled default (#85).
- Add explicit cloud configuration examples, opt-in real development cloud acceptance suites, and asset staging with manifests and reviewable cutover/rollback SQL. The initial production maintenance-window cutover completed on 2026-09-14; Goose never contacts object storage. Archive the verified production results in [the acceptance record](docs/acceptance/v3-alpha8-managed-assets-production-acceptance.md).

### Changed

- Recompose player, price, language, and certificate analytical workspaces with shared selectors, rankings, risk lists, coverage summaries, and expandable methodology; improve Site/Game comparison with searchable entity pickers, media, and responsive matrices.
- Redesign the Change Explorer as a date-grouped entity timeline while retaining authoritative cursor ordering, repeated events, filtering, and localized presentation.
- Split public preferences into Home and Page Background underline tabs; replace background dropdowns with source controls and a preview carousel, align appearance inputs, and preserve explicit Save/Cancel behavior.
- Compact Admin asset editors with shared SVG file-picker buttons, full-card Hero previews at desktop 16:9 and mobile 1:1 ratios, split light/dark pattern previews, and confirmation before clearing icons, replacing files, or deleting entries.
- Organize Cloud Resources into storage summaries, a full-width object inspector, paired CDN controls, and task history using shared Admin components and custom Selects; keep full-zone operations in a separate, initially collapsed section with explicit confirmation.
- Migrate Game email delivery from `gomail` to `github.com/wneessen/go-mail`, retaining all five templates, attachments, retry/timeout behavior, and CC/BCC recipients (#72).
- Migrate direct YAML usage in active Go services and tools to `go.yaml.in/yaml/v4` through its v3-compatible API while preserving explicit YAML loading and configuration decoding semantics (#72).
- Migrate Nav Backend monitoring to native `github.com/gofiber/contrib/v3/monitor` middleware with `Next`-mode global request accounting, retaining `/monitor` without the HTTP adaptor, manual request counters, or a `Stop()` lifecycle hook (#102).
- Differentiate the Site and Game observatories with capability rails, contextual trends, selectable dimension bars and complete raw tables; reuse existing entity media and Game Pulse while preserving query and SEO contracts.
- Recompose the Ecosystem overview as an editorial entry with independent Site/Game snapshots, recent entity activity, existing Game Panel highlights, and optional Overview image references; preserve domain pages and SEO contracts.
- Establish the Ecosystem Observatory visual foundation with scoped layout tokens and stacked, accessible primary/domain text navigation while preserving existing page content and URLs.
- Redesign the public Nuxt error experience with immersive theme-aware artwork, staged accessible transitions, localized 404/5xx copy, and dedicated recovery actions.
- Redesign the Game detail ecosystem tab around a unified player/price/state overview, dual-series player history, reliable price tooltips, and responsive compact/list change timelines.

### Fixed

- Preserve the independent `status.go-furry.com` uptime service while planned maintenance places the main GoFurry application hosts behind the 503 maintenance page.
- Recognize the managed-asset migration in Admin Data Operations and align three-database integration coverage with dedicated icon mutation endpoints, including preservation of existing icons during ordinary Site edits.
- Make Admin Steam prefill best-effort across Chinese/English details and auxiliary assets; apply available nonempty fields even when another source fails, and fail only when no meaningful data remains (#103).
- Reuse the cached Game Home panel for Ecosystem pages instead of rebuilding the uncached panel during SSR.
- Keep Game detail tabs working when regional price data is missing; batch independent detail reads and provide an opt-in, debug-only homepage cache for remote development infrastructure without changing production cache defaults or SEO/SSR semantics.
- Replace obsolete workstation-specific Nuxt backend and development monitor defaults with loopback addresses while retaining environment overrides.
- Reuse the shared Insights selector styles for Game detail compact/list timeline controls, preserving their existing responsive behavior (#100).
- Restore Admin CDN image previews through compatible response headers and accept original SVG content without sanitizing or rejecting declarations; retain upload size limits and render SVG through image URLs/CSS masks rather than inline markup.
- Stabilize Admin sidebar width transitions with a shared rail/workspace width, nonwrapping hidden labels, and collapsed group separators; use distinct semantic Phosphor menu icons and hide only the sidebar scrollbar while retaining scrolling.
- Prevent browser history autocomplete and IME composition rerenders from disrupting Admin search, filter, and remote-option input experiences.
- Align the Site detail background with the public shell, refine the responsive navigation toggle, initialize Admin datetime drafts from local time, and close Sheet footers cleanly at the viewport edge.
- Close the remaining SEO recovery edge cases with strict invalid-entity 404 semantics, CI recovery guards, unique Site Group metadata, Prize noindex headers, and removal of the retired Steam-page performance scenario.
- Consolidate public Site entity SEO identity on `/site/:id`, permanently redirect legacy Site detail aliases, and keep target selection as a non-canonical query view.
- Return real 404/503 semantics for authoritative Site/Game detail failures instead of indexable HTTP 200 error shells.
- Make sitemap generation fail closed on inventory errors and publish canonical entity URLs only.
- Upgrade steam-go to v1.3.10 to restore StoreBrowse asset decoding after Steam introduced non-string asset metadata.
- Confirm vertical library-cover acquisition through the local single-game canary.
- Preserve Steam StoreBrowse vertical-cover Last Known Good with explicit source/language replacement scopes, observable partial failures, Official API traffic classification, and post-commit merged cache refreshes.
- Keep authoritative hashed Steam asset pathnames unchanged while selecting real 1x/2x library-cover rows independently in Game Backend and Nav Web CDN fallback.

### Removed

- Remove active hardcoded `qcdn.go-furry.com` URLs, legacy `SiteLogo`/`GamePrefix` configuration, `nav/static/SiteLogos` and `nav/bg` path conventions, numbered Hero counts, and `GetImageUrl()`; bundle fixed artwork and use managed keys without a permanent legacy compatibility layer (#105).
- Remove unused Nav Web Axios helpers, `md-editor-v3`, and `highlight.js`, plus dead copied Go HTTP helpers and their unused dependency edges; retain actively used HTML parsing, media, date-picker, and chart dependencies (#72).
- Remove the obsolete cross-framework monitor experiment and `third-party/monitor` submodule after the production middleware migration; retain the Steam reference submodule (#72).

## v3.0.0-alpha.7 - 2026-09-04

### Added

- Establish Phosphor as the primary system UI icon family for the Admin and public navigation surfaces touched by Stable Polish Batch 1.
- Add an application-level public background foundation with the supplied mask-friendly, infinitely tiled `gofurry-pattern.svg`, theme-controlled color/opacity, and a replaceable `--gf-page-pattern` asset slot; keep default-layout page roots transparent and preserve the previous grid and falling-leaf effects under an experimental namespace.
- Add authenticated Admin self-service username and password APIs with current-password verification, uniqueness enforcement, audit history, identity refresh, and session-version revocation.

### Changed

- Rename the public `洞察 / Insights` product to `生态观测 / Ecosystem` while preserving all `/insights/*` routes and internal contracts.
- Remove repetitive Ecosystem hero blocks, combine size-consistent primary and domain navigation into responsive left/right or two-row layouts, and add Ecosystem to the localized mobile bottom navigation.
- Make the Admin shell own the viewport, sidebar/workspace scrolling, subtle shared scrollbars, and table-local horizontal overflow; simplify the text-only sidebar brand to `GoFurry` expanded and `GF` collapsed.
- Polish the Admin product shell with a 208px expanded sidebar, warm-gray Light surfaces, a persisted single-button Light/Dark switch, and an explicit capability-aware Data Operations shortcut.
- Recompose the Workbench into flat list cards, simplify every shared Admin PageHeader to title plus actions, align shared DataTable toolbar controls, and replace split remote search/select controls with one searchable metadata-aware combobox.
- Present active Collection jobs with attempted/expected progress bars and 2.5-second Running-tab refreshes, including explicit queued and waiting-for-progress states.
- Tighten shared Admin PageHeader spacing, replace native select/date controls with reusable Base UI controls, add the capability-aware Collection header shortcut, and constrain long comment previews.
- Recompose Data Operations around selectable database summary cards, two operational status regions, a single-layer relation table, and on-demand technical details while preserving read-only 60-second refresh behavior.
- Standardize every Admin workspace on a compact shared page rhythm, anchor shared Select popups below their triggers, and make visible DataTable columns immediately identifiable with check indicators.
- Recompose Collection overview and schedules plus Metrics and Changes into compact, single-level operational layouts with denser KPIs, unified chart/filter toolbars, separator-based lists, and on-demand technical details.
- Further compact the Collection KPI row by removing its collector-view subtitle.

### Fixed

- Bind all six schema-driven Admin resource routes to an explicit Nav or Game domain so valid Site Group, Update Notice, Saying, Tag, Comment, and Prize pages resolve their definitions.
- Keep shared Admin dialogs centered and viewport-safe with portal backdrops, bounded height, and internal scrolling without document-level horizontal shift.
- Keep Collection ECharts lines visible while axis tooltips are active by disabling hover emphasis replacement.
- Clear protected Query/Mutation state on logout before replacing the current route with `/login`, preventing protected-shell error remnants.
- Preserve Game Comment IDs as strings across Go JSON and React CRUD routes so 64-bit detail, update, and delete operations never lose precision.
- Remove the invisible DataTable search-label spacer so page headings and table toolbars follow the shared compact vertical rhythm.

## v3.0.0-alpha.6 - 2026-09-03

### Added

- Add V3-P2.4 order-preserving Site and Game Entity Compare APIs and SSR product pages for bounded 2–4 entity cohorts, without scores, rankings, winners, or recommendations.
- Add common-snapshot Site capability/certificate comparison and separately disclosed Game state, scheduled-player, 30-day player-quality, regional-price, observed-low, and language horizons.
- Add Compare unit, route, PostgreSQL, URL-state, semantic, SSR, and CI regression coverage, including real-zero versus unavailable players and free versus priced-zero safeguards.
- Add V3-P2.3 Site Capability Intelligence with version-frozen HTTP/2, HSTS, enforcement CSP, and TLS certificate-verification Metrics plus same-Primary-Target semantic Change detectors.
- Add the bounded certificate overview API and bilingual `/insights/sites/certificates` product surface with common fact-day horizon, deterministic expiry buckets, attention lists, and public verification-issue whitelisting.
- Add focused PostgreSQL projection, Metric-before-Change backfill/rebuild, certificate read-model, route, SSR/typecheck/build, Goose, sqlc, and query-plan regression coverage.
- Add V3-P2.2 Game Intelligence: scheduled Player rankings and quality-aware 30-day peak/weighted-average read models, CN/US/HK price overview/discount APIs, region-aware histories, and bounded GoFurry Observed Low semantics.
- Add `mac_support/1` and `mac_support_transition/1` Goose contracts, compiled Metric/Change projectors, global/primary-tag/tag slices, entity state, public platform changes, and explicit Metric-before-Change backfill support.
- Add common-snapshot Supported Languages and Explicit Full-Audio distributions with freshness and normalization-quality disclosure, without fuzzy-mapping unknown language names.
- Add SSR Game Intelligence Players, Prices, and Languages pages plus local navigation, regional Game Detail summaries, region+range price caching, currency-separated chart history, and bilingual responsive coverage.
- Add Public Insights dimension breakdown and selected-slice trend APIs for Site country/group/content/public-interest and Game primary/all-tag dimensions, with frozen global horizons, null-safe public mathematics, overlapping-slice disclosure, and metadata fallbacks.
- Add the domain-specific Change Explorer APIs and `/insights/changes` product flow with public category/type filters, projection-date ranges, precision-aware opaque keyset cursors, CN-only Game price/discount scope, SSR first pages, and client Load More.
- Add independent Nuxt dimension URL state, SSR/deferred slice loading, responsive bilingual dimension tables/charts, Explorer filters, semantic regression coverage, and visual-guard scenarios.
- Add SSR-isolated Site Entity Insights with exact seven-state capability rendering, same-day ecosystem context, semantic timelines, and links back to website ecosystem Insights.
- Add a Game Detail Insights tab with SSR summary data, lazy player/CN-price histories, shared 30d/90d/all selection, page-lifecycle range caching, independent retry states, and entity timelines.
- Add regression coverage for nullable/zero player semantics, free versus priced-zero versus unavailable prices, day/exact event precision, entity failure isolation, target-route boundaries, and entity visual scenarios.
- Add SSR-rendered bilingual Public Insights overview, website ecosystem, and game ecosystem pages with metric cards, an ECharts trend view, coverage disclosure, and recent public changes.
- Add focused Insights route, Workshop 404, URL interaction, null/error semantic, and responsive smoke coverage plus Insights visual-guard scenarios.
- Add Nav and Game Public Insights overview, trend, entity, player, and CN price APIs with explicit public-to-internal version mappings and correctness regression coverage.
- Add the final Admin Vue-to-React functional parity matrix, embedded SPA routing regression coverage, and Owner/Developer/Operator guide.
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

- Replace the legacy GoFurry Workshop navigation and pages with Insights, leaving old zh/en Workshop URLs as ordinary 404 responses.
- Make React the sole Admin production frontend, write its clean Vite build directly to the Go embed directory, and remove the Vue build/cache path from CI.
- Restore searchable Resource Engine remote options plus Collection chart/history/result filters found by the final parity audit.
- Change Admin bootstrap/login from a singleton password to canonical username, display name, and password while preserving the legacy account password and timestamps during migration.
- Keep JWTs role-free and resolve current role, status, session version, and capabilities from `gfa` on every authenticated request; minimally update the existing Vue login/setup compatibility surface.
- Preserve per-query AAAA evidence and validate security.txt content before treating either capability as adopted; inconclusive DNS and unrecognized documents remain unknown instead of becoming false or positive.
- Add count-backed pagination and compact detail views to Metric Daily/Entity results, Collection Run/Task history, and Collector lifecycle views.
- Replace manual collection database-ID entry with searchable Game, Site, and current Site Target selectors.

### Fixed

- Align Admin migration catalogs and Nav Metric registry/checkpoint integration expectations with the P2.2/P2.3 Goose contracts.
- Detect Site target child routes from the `domain` route parameter so localized `/en/site/:id` pages retain their Site-level Insights SSR panel.
- Require explicit impact confirmation for collection cancellation and schedule state changes while preserving the existing control-plane semantics.
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
