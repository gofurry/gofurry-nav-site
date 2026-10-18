# Nav Web engineering

Source lives in `apps/cn/web`; public Task/CI names and the Compose project remain
`nav-web`. Directory naming does not change product or deployment identity.

Read [the frontend contract](../../../contracts/nav-web-frontend.md) first.
Root `AGENTS.md` and `.agents/` apply. This is the working router; detailed
ownership is in the contract, [design system](../../../docs/frontend/design-system.md)
and [testing guide](../../../docs/frontend/testing.md). Managed assets, Hero and
routing also follow [assets](../../../contracts/assets.md) and
[managed-assets](../../../docs/managed-assets.md).

Motion and animation changes MUST follow [Motion Contract V1](../../../docs/frontend/motion.md).
Preserve Reduced Motion, SSR, focus, and cleanup of stale timers/RAFs; prefer the
existing Vue Transition + Less owners. New animation libraries need explicit approval.

## Choose the existing owner

**Tailwind owns structure; Less owns appearance.**

- Tailwind owns placement, flex/grid, alignment, responsive composition, outer
  spacing, position, overflow, visibility and text alignment/truncation.
- Search `styles/tokens.less`, `primitives/`, `components/` and `pages/` before
  adding appearance. Reuse `.gf-button`, `.gf-card`, `.gf-input`, `.gf-chip`,
  `.gf-modal`, `.gf-pagination` and `.gf-rating`, including existing variants.
- Control typography, internal sizing, colors, radius, shadow and hover/focus
  belong to the owning Less/primitive layer. Do not reset a primitive from a consumer.
- Global tokens require real shared meaning. Primitive/domain/compound tokens
  stay with their exact owners. Equal literal values do not imply equal semantics.
  Comments explain role/scope, not CSS syntax; keep literal values in source.
- Scoped style owns private structure/runtime geometry, not another theme system.
  Repeated appearance may move from component to domain to primitive when justified.
- Reuse `html.dark`; do not add page-dark classes, Tailwind dark appearance or
  ordinary raw visual values. Approved token declarations match file + root + prefix.
- Preserve `tokens → mixins → primitives → components → pages`. Do not create an
  empty `domains/` tree or move files merely to match an aspirational directory map.
- `common/` is a location, not a visual owner. CSS primitives provide appearance;
  Vue primitives need reused behavior/accessibility, not a one-line wrapper.

The design-system guide records effective cascade decisions: native control
`font: inherit`, preserved line-height ratios, Dark specificity and Teleport roots.
Accepted computed appearance takes priority over assumptions about utility classes.
The layout/PublicPageBackground owns the canvas; Static/Legal roots are transparent.
Never reintroduce `--gf-bg-page` or `gf-modal__toggle`.

Typography families use `--gf-font-sans` / `--gf-font-mono`; the frozen Sans is
Manrope Variable + Noto Sans SC Variable. Normal authored weights are only
400/500/600/700/800, with 800 reserved for Display. Keep `font: inherit` for native
controls; no other `font:` shorthand may bypass the family/weight guard.
`primitives/rating.less` alone retains the Arial star-glyph exception. #108 F3
retires the Insights legacy-weight override, including its Game Detail consumers.
See the contract and design system for role mapping; do not compensate with metrics.

## Preserve product boundaries

P4 Stable/Common, P5 Nav and P6 Games appearance migrations are complete. Existing
contracts remain authoritative; completion does not authorize redesign or more cleanup.

- Preferences owns `--gf-preferences-*`; private editor layout stays scoped.
- Nav Home keeps `pages/nav.less` and `--nav-home-*`; Header is theme-independent.
  Body popovers and shared Card/toggle states retain their precise token roots.
- Games Home/Search/Detail retain the audited `.games-page` compatibility scope.
  Review, Search Filter and Detail Lightbox have their own body-mounted owners.
  Charts read resolved tokens after theme updates; preserve their shallow instances.
- Keep Hero displayed-resource fallback, per-resource routing snapshots,
  Fixed/Local/BigInt persistence, Save/Cancel and staged handoff contracts intact.
- Preserve Search draft/Apply/Cancel, instance-bound cancellation and slice-local
  Retry; preserve Detail/NSFW/Lightbox keyboard, inertness and scroll cleanup.
- Lottery reads/writes remain isolated in tests; never send real participation/email.
- Managed/Steam image components are runtime infrastructure, not migration targets
  merely because they live in `common/`.

## Debt and exceptions

Site Detail #109 P1 runtime follows the contract's Site/Target ownership section.
Use `siteDetailRouteState.ts` for UI query and `siteRoutes.ts` for links; keep
Insights keyed only by Site ID and View counted once per hydrated Site session.
Use the complete `siteCapabilityRegistry.ts`,
as the catalog. `site-detail-contract.spec.ts` owns Target-switch request counts
and failure isolation; SEO remains Entity-only. P1 grants no appearance/debt or
Visual baseline changes; P2–P8 and #108 require their own scope.

P2 Shell/Target Context follows the contract's separate P2 section. New Target
surfaces consume `siteTargetPresentation.ts`; primary tabs/selector use the P1
route owner. `site-detail.less` owns appearance, `site-detail-shell.spec.ts` owns
responsive/keyboard/pending/race behavior. Only P2's replaced Hero/popover/root
debt may decrease; remaining panels keep their later-phase owners. P2 requires
maintainer visual review before P3 and does not create a Visual golden.

P3 Overview consumes `siteOverviewPresentation.ts`, a pure Site-only projection.
Keep the first Site/language summary for the page session, use all seven registry
capabilities, and preserve empty/unavailable plus day/exact precision. P6 replaced
the legacy Insights panel; Overview owns no fetch or Target
protocol checks. `site-overview.spec.ts` owns this contract using the shared
fixture. P3 adds no appearance debt or Visual golden and requires maintainer
visual acceptance before P4.

P4 Observation uses `siteObservationPresentation.ts` for Target evidence and the
P1 route owner for its five secondary views. Page-owned `useSiteObservationHistory`
auto-loads only hydrated Performance, caches by Site/Target/Ping, slices one 100-row
response and isolates late results. Other views add no fetch. Keep Security probes
out of Web, raw timings unsummed and collector loss percentages unscaled. Use
`site-observation.spec.ts` plus pure/Nuxt contracts; only removed Observation debt
may decrease. P4 requires manual review before P5 and no accepted Visual updates.

[frontend-style-debt.json](frontend-style-debt.json) is measured state, never an
example or permission. After #108 F3, only intentionally preserved ambient raw
debt remains. Earlier phase notes below record their original boundaries, not
current budgets. No opportunistic ambient migration; new code follows the contract.

P5 Security consumes only Current Target Detail evidence through
`siteSecurityPresentation.ts`; four secondary views use `selectSiteSecurityView`.
No view fetch, score/verdict, port risk inference or WAF deployment claim is allowed.
Preserve collector booleans, missing/failure distinctions and verification versus
validity. `site-security.spec.ts` owns requests/states/disclosures/responsive checks.
Only its replaced Metric Grid/Light Probe debt is removed; deep remains zero.
P6 requires maintainer review and no P8 golden is created here.

P6 Insights uses `siteInsightsPresentation.ts` for Site facts and full recent
changes. Page-owned `useSiteInsights` shares SSR data/state/retry with Overview;
`useSiteInsightTrend` activates only hydrated Insights and caches metric + range.
Target never enters either identity. Use route helpers for metric/range, preserve
slice/fact failure distinctions and classify every active chart state. Public
copy keeps Ecosystem naming. `site-insights.spec.ts` and real Nuxt tests own this
contract; Site Detail owns appearance, #108 remains separate. Manual review precedes
P7, and accepted Visual files are unchanged.

P7 retires the consumer-audited SiteHealthSummaryPanel, SiteOverview, SiteSignalCards,
SiteChangeEvents and detailTypes. Site Detail has zero style debt; do not restore
legacy clones or historical dark/deep exceptions. Active P2–P6 components and
`site-detail.less` keep their existing appearance/runtime ownership. Home still
uses `site.siteDnsPanel.none`; other retired `site.*` labels are removed. Preserve
active ESLint suppressions, ambient raw debt and #108 important debt. P8 requires
maintainer Desktop/Mobile smoke and separate authorization for final Visual work.

Task B refines only Site Detail presentation through its existing Less owner.
Keep the three Surface levels, Phosphor system icons and shared segmented pattern.
`siteDetailPresentation.ts` owns localized reason copy and display tones, never
Collector health decisions. Health Strip/Security share P5 certificate expiry
normalization; known reason/validation codes use both locales. Target Context no
longer renders relation/infrastructure dumps. `site-detail-refinement.spec.ts`
adds focused UI assertions and optional temporary review screenshots through the
same Functional fixture. Debt stays zero; full manual review precedes P8 goldens.

First-round Site Detail completion uses shared acrylic Plane A/B/C and borderless
content surfaces, spaced 500ms hover rows, an Identity Note and title-row
Observation/Security subnav. Overview/Observation composition remains intact;
Security and Insights preserve all runtime and evidence contracts. This supersedes
earlier card/divider appearance only. See the frontend contract. Stop before P8
or a second round; final Visual work still needs separate authorization.

Task C explicitly authorizes the second-round consistency/entry refinement and
supersedes that material hierarchy: all content panels use `--site-detail-panel`,
with one shared hover/selected family. `siteTargetSignals.ts` owns reliable typed
CDN hints and shared protocol status display. Blank/domain-only URLs now mean
Observation/Performance; both Overviews are explicit. Default hydration adds one
Ping history call, never SSR. Preserve P4/P6 cache/race internals and all other
runtime/SEO/debt boundaries. Task C still does not authorize P8 or final goldens.

Task D authorizes focused composite/help/changes refinement plus Game average and
one-shot Gallery media-abort fixes. Preserve Task C entry and P4/P6 runtime.
`SiteDetailHelpTooltip` replaces native Info titles; `SiteChangeStream` shares only
serpentine geometry with Game. All appearance stays with its existing owner.
For Task D only, run focused Playwright, typecheck and focused lint/style checks;
do not run full tests or production build. `GOFURRY_FIXTURE_DEV=1` opts the existing
fixture into isolated local Nuxt source checks, never CI/production acceptance.
No P8 or final golden is authorized.

Task E adds one optional Site ID + language recommendation SSR slice. Backend
membership uses full groups; candidates use the existing Home Top-8 builder before
union/dedupe/daily SHA-256 shuffle. Mobile Similar is auxiliary local state, never
a fifth route tab. SSR-rendered recommendations are SFW; mode changes only filter
raw items. Similar links never count views themselves. The focused Task E budget
also forbids full suites/build; reuse the opt-in source fixture. No P8/goldens.

P8 freezes the final Site Detail product. `fixtures/site-detail-visual.ts` and
`visual/site-detail.spec.ts` under `tests/browser/` own eight root captures in the
existing pinned Visual runner. Scope updates to this spec and require a separate
compare; never repair a product defect by updating a golden. Functional retains
the eight Site Detail regression owners, including recommendations/refinement,
with no exact pixel-gap assertions. The contract's final P8 section supersedes
intermediate phase designs; the #109 acceptance ledger separates local evidence,
current-code remote gates and pending explicit maintainer visual approval.
No new feature/design round or P9 follows from closure.

Absent rule/file budgets are zero. Both increases and stale larger budgets fail.
After removing debt, inspect stale-only output before `style:policy:update`; never
raise budgets, move debt between files, or manually rebalance totals. A debt-bearing
rename requires explicit policy review. Exceptions require exact path/rule/issue/
reason/remove_when; no wildcard exemptions. ESLint bulk suppressions also cannot
be regenerated to hide new findings; prune them only after fixing their debt.

## Verification and test ownership

Use Node 24 and pnpm 12.6.0, pinned in `packageManager`. Install this project's
independent lock with `pnpm install --frozen-lockfile`; use `pnpm run` / `pnpm exec`.
Its `pnpm-workspace.yaml` contains only local settings and script permissions;
there is no root workspace. Root Task offers `deps:nav-web`, `lint:nav-web`,
`typecheck:nav-web`, `test:nav-web`, `test:nav-web:browser`, `build:nav-web` and
`build:nav-web-image`. Browser/Visual remain outside default `task test`/`verify`.

Use the [current verification sequence](../../../docs/frontend/testing.md#full-verification).
Pure logic uses Vitest unit; real Nuxt state/composables use the Nuxt project.
Reset cookies/useState per case and mock only business injection boundaries.
No source-transpile/data-URL imports or fake Nuxt runtime.

Browser tests use the real production Nitro build, deterministic local upstream
and exact network/error accounting. Each worker owns stable servers; each test
owns fresh mutable scenario/gates. Playwright owns contexts/pages. Release gates
unconditionally; do not use `unrouteAll(wait)` or fixed sleeps for readiness.
Chromium only, retries zero, CI workers one. Daily dev/PR `nav-web` requires Fast
static/unit/Nuxt/build plus the seven Smoke owners in one job. Full runs on main,
manual dispatch and nightly dev through four explicit inventory-checked groups;
deployment image verification runs on main/manual Full only. Visual is manual-only.
Use `--workers=1` locally; do not run Full locally merely to validate CI wiring.
See the testing guide's #134 tier contract and measured acceptance record.

`assertHeroHydration` is the sole narrow mobile Home Footer-debt check. It defaults
to `/`; only the English Home fixture explicitly opts into `/en`. Require one
initial mismatch, width <768, SSR Footer=0/client Footer=1, retained SSR Hero and
no other errors. Preserve raw evidence and reject later errors. Never change
generic browser-error capture to ignore hydration. See the
[owned runtime follow-up](../../../docs/acceptance/issue-124-frontend-engineering-closure.md#runtime-footer-01).

`visual:guard` and the migrated legacy runtime runners are retired. Their source
checks belong to style-policy; browser behavior to Playwright Test; accepted
pixels to Visual. Do not recreate a second runner. Keep direct `playwright` for
perf/cloud tools, reusable `scripts/fixtures`, Node style-policy tests and the
Insights/SEO Contract Guards. External/cloud acceptance requires explicit scope
and credentials and is not a normal gate.

Visual is separate from Functional Browser. Only the current digest-pinned Linux
Playwright image with Node 24 is authoritative. Fast builds and runs Smoke in one
job. Full archives one build for all four groups, pinning their checkout to its
resolved source SHA (nightly resolves dev once). Manual Visual builds and compares
in its own pinned job. Docker separately verifies main/manual Full deployment images.
Never update snapshots to make a test pass. Approved visual changes require
explicit authorization, pinned generation and maintainer review. Treat package,
image digest, browser revision and baselines as one upgrade unit.

Before committing, review the full diff and run applicable gates. A local pass or
skipped CI job is not remote acceptance. The [#124 closure record](../../../docs/acceptance/issue-124-frontend-engineering-closure.md)
keeps final evidence, known issues and pending maintainer sign-off; historical
phase counts are not the current suite inventory.
