# Nav Web engineering

Read [the frontend contract](../../../contracts/nav-web-frontend.md) first.
Root `AGENTS.md` and `.agents/` apply. This is the working router; detailed
ownership is in the contract, [design system](../../../docs/frontend/design-system.md)
and [testing guide](../../../docs/frontend/testing.md). Managed assets, Hero and
routing also follow [assets](../../../contracts/assets.md) and
[managed-assets](../../../docs/managed-assets.md).

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
Use the complete `siteCapabilityRegistry.ts`, not its existing three-item preview,
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
capabilities, and preserve empty/unavailable plus day/exact precision. The legacy
Insights panel now belongs only to its tab; Overview owns no fetch or Target
protocol checks. `site-overview.spec.ts` owns this contract using the shared
fixture. P3 adds no appearance debt or Visual golden and requires maintainer
visual acceptance before P4.

[frontend-style-debt.json](frontend-style-debt.json) is measured state, never an
example or permission. Remaining style debt belongs to Site Detail #109, Insights
#108 and intentionally preserved ambient effects. No opportunistic migration of
these areas; their new code still follows the contract.

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
Chromium only, retries zero, CI workers one per shard. All three Browser shards
are required; the stable `nav-web` check also requires Visual and Docker. Use `--workers=1` locally for the
complete acceptance run; focused smoke/regression commands remain available.

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
Playwright image with Node 24 is authoritative. CI builds once in that image;
Browser and Visual consume the same commit's archived output in matching containers
and only compare. Docker separately builds the deployment image with its original context.
Never update snapshots to make a test pass. Approved visual changes require
explicit authorization, pinned generation and maintainer review. Treat package,
image digest, browser revision and baselines as one upgrade unit.

Before committing, review the full diff and run applicable gates. A local pass or
skipped CI job is not remote acceptance. The [#124 closure record](../../../docs/acceptance/issue-124-frontend-engineering-closure.md)
keeps final evidence, known issues and pending maintainer sign-off; historical
phase counts are not the current suite inventory.
