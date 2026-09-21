# Nav Web engineering

Read [the frontend contract](../../../contracts/nav-web-frontend.md) before
changing Nav Web. It owns the detailed rules; this file is the working entry point.
Root `AGENTS.md` and `.agents/` still apply. Managed asset, Hero, routing and
SSR work also follows [the asset contract](../../../contracts/assets.md) and
[managed asset documentation](../../../docs/managed-assets.md).

## Choose the existing owner

**Tailwind owns structure; Less owns appearance.**

1. Use Tailwind for placement: flex/grid, alignment, responsive composition,
   outer spacing, width, position, overflow, visibility and text alignment/truncation.
2. Before adding appearance, search `app/assets/styles/tokens.less`, the canonical
   global token owner, and `app/assets/styles/primitives/` for reusable appearance.
   `styles/components/` owns compound product UI; search `pages/` for domain styles.
3. Reuse `.gf-button`, `.gf-card`, `.gf-input`, `.gf-chip`, `.gf-modal`,
   `.gf-pagination` and `.gf-rating`; use their existing variants before adding one.
4. Keep control height/padding, typography, colors, radius, shadow and hover/focus
   in the owning primitive/Less layer, not a new bundle of Tailwind visual classes.
5. Primitive-local tokens stay with their primitive; domain tokens should alias
   global meaning by default. Independent domain values need a documented role.
   Component-private geometry may stay scoped; it is not a second theme system.
6. Token comments explain semantics, scope and ownership. Use group comments;
   add an individual rationale for exceptional tokens, not comments naming colors.

New global tokens need a real consumer and stable semantic meaning; do not move
literals into tokens just to pass policy or merge tokens because values match.
Keep actual values in source, not docs. New page backgrounds use
`--gf-page-background`. Static/Legal roots stay transparent; the layout and
`PublicPageBackground` own the canvas. Do not reintroduce retired `--gf-bg-page`.
About/Legal panel elevation shares `--gf-static-panel-shadow` in global tokens.
The [design-system guide](../../../docs/frontend/design-system.md) gives examples.

Rating's tokens stay in `primitives/rating.less`. Generic Modal appearance belongs
to `primitives/modal.less`; shared Preferences composition and local
`--gf-preferences-*` semantics belong to `components/preferences.less`. Do not
reintroduce `gf-modal__toggle`: the product-only control is `preferences-toggle`.
Keep private Preferences paging/editor structure scoped unless shared across
editors. Nav, footer and shell stay in `components/`. Preserve the
`tokens → mixins → primitives → components → pages` import order. Do not create
an empty `domains/` directory or use file moves to clean up appearance.

Appearance reuse belongs in CSS primitives. Reused behavior and accessibility
(keyboard, focus, ARIA and state) justify a Vue primitive. Evaluate promotion when
the same pattern appears in two business domains; do not create speculative wrappers.

`app/components/common/` does not confer shared visual ownership. Follow the
[semantic ownership map](../../../docs/frontend/design-system.md#common-directory-semantic-boundaries)
for P4/P4.5/P5/P6 and runtime asset boundaries; do not move files by directory
label. Delete historical components only after proving zero consumers across
template/auto-import names, imports, dynamic resolution and source paths. Keep
live consumers unchanged. Inspect stale-only policy debt before using the updater.

`<style scoped>` is valid for component-private structure. It must not redefine
shared primitive appearance or invent a local theme. Repeated appearance should
move from component to domain to shared primitive as reuse warrants it.

Do not introduce new raw colors, radii, shadows, visual durations or typography
scales in ordinary selectors. Reuse the semantic token owner first. Theme changes
belong in semantic variables under the existing `html.dark` model, not new
`dark:bg-*`/`dark:text-*` variants or independent page-dark classes.

## Historical debt and phase boundaries

After P4, do not opportunistically clear remaining debt: Nav/MobileBottomTabBar
belong to P5, Game plus BlurWrapper/LinkTag to P6, Site to #109, Insights page/domain
to #108, and ambient effects remain intentionally preserved experimental code.

- Existing violations are historical debt, not examples to copy.
- [frontend-style-debt.json](frontend-style-debt.json) is state, not permission.
  Do not increase any rule/file budget; absent rule/file entries have budget **0**.
  Reduce the matching budget when removing debt. Do not transfer debt between files.
- A debt-bearing file move requires an explicit reviewed policy migration:
  renaming leaves a stale old budget and a zero-budget new path. Do not move it casually.
- Exceptions need exact `path`, `rule`, `issue`, `reason`, and `remove_when`.
  A wildcard ignore or an old visual-guard allowlist is not a new style exception.
- #108 Insights and #109 Site Detail are excluded from proactive P4–P6 migration,
  not debt-free areas. Measure their existing debt; new code follows the contract.
- P0 only establishes governance and the one-time baseline. Do not change UI,
  production CSS/Less/Vue, dependencies, CI, scripts or style directories for P0.
  Keep existing source and regression harnesses until their own scoped phases.
- P1 installs static checks, not UI/style migration. ESLint's official bulk
  suppressions hold historical engineering debt; never regenerate suppress-all
  to hide new findings. Use `npm run lint:prune` after removing lint debt.
- `style:policy` requires exact rule/file equality. Both increased debt and a
  stale higher budget fail. After removing debt, run `npm run style:policy:update`;
  it refuses all writes if any rule/file increased. Never raise budgets manually.

## Verify the change

Run from `apps/cn/nav-web`:

```text
npm ci
npm run lint
npm run stylelint
npm run style:policy:test
npm run style:policy
npm run test:unit
npm run test:nuxt
npm test
npm run typecheck
npm run insights:semantics
npm run seo:recovery:test
npm run build
npx playwright install chromium
npm run test:browser:smoke
npm run test:browser:regression
npm run test:browser
```

Choose the lowest-cost faithful environment: pure logic goes in `tests/unit`
(`test:unit`); real Nuxt runtime goes in `tests/nuxt` (`test:nuxt`). Reset cookies
and Nuxt state per case; mock business injections, never the Nuxt runtime itself.
See [testing guidance](../../../docs/frontend/testing.md) for migration coverage
and isolation. Migrated browser SSR/hydration/interaction regressions belong in
`tests/browser`; build production Nitro before running them. Game Detail,
Resource Routing/Managed/Steam, Hero lifecycle/Local, Preferences foundation,
Fixed/BigInt, Catalog and handoff now belong to Playwright Test. Do not recreate
their retired smoke scripts or `assets:routing-smoke`. Keep the two Hero domain
fixtures separate, with fresh test scenarios and gate teardown. Only the narrow
Hero assertion may acknowledge the known mobile Footer hydration debt; never
globally suppress hydration errors. Insights and other smokes keep their runners.
Chromium is the browser gate; retries are zero and CI
uses one worker. Domain fixtures own servers and reset state per case; Playwright
owns contexts/pages. Failure artifacts are diagnostics, not visual baselines.
Style-policy keeps its separate Node built-in runner; Insights/SEO remain
Contract Guards. Check `package.json` for relevant focused commands. CI runs
Unit and Nuxt tests as independent steps alongside the existing guards, followed
by Build, Chromium installation and Browser tests.
Functional `test:browser` excludes `tests/browser/visual/**`. Visual checks use
`playwright.visual.config.ts` / `test:visual`; only the digest-pinned Linux
Playwright container with Node 24 is authoritative. P3.3.1 adds an environment
sentinel; P3.3.2's `ui-foundation.spec.ts` owns four Shared Primitive baselines.
Its fixture uses production `/about` CSS with product JavaScript disabled;
fixture CSS must not redefine product appearance. Keep Generic Modal separate
from the local Preferences Toggle scope. `preferences-modal.spec.ts` owns eight
real Preferences/backdrop baselines, reusing the Hero fixture and production
theme/runtime. Never fake Preferences DOM, restyle it in tests or bind its
snapshots to homepage pixels. Business surfaces remain P4+ scope. Keep
`visual:guard` and direct `playwright` for legacy page/report checks. See testing
guidance for pinned commands, deterministic Routing seeds and maintainer review.

`static-pages.spec.ts` owns the six real Static/Legal viewport baselines, including
the layout-owned canvas and transparent Static root. Appearance migrations must
pass these accepted baselines unchanged, never update them to turn
the gate green. See testing guidance for semantic readiness and network isolation.

`updates-page.spec.ts` owns Updates appearance; `regression/updates.spec.ts`
owns grouped-year/load-more behavior through the shared real SSR fixture.
Selector normalization must pass both without updating accepted snapshots.
`updates.less` is namespace-enforced by Stylelint: use `updates-*` classes and
`--updates-*` owned custom properties; consuming `--gf-*` global semantics is
allowed. Keep `is-*` states attached to Updates-owned bases and preserve specificity.

`visual/error-experience.spec.ts` owns four real Nuxt 404 locator baselines.
`regression/error-experience.spec.ts` owns normal-motion keyboard focus safety
and the genuine no-JS Light fallback. Error appearance migration must pass these
contracts without updating snapshots; new goldens require maintainer review.

`visual/page-scroll-dock.spec.ts` owns Dock normal/hover appearance at 50%.
`regression/page-scroll-dock.spec.ts` owns real document progress, quarter-step
scrolling and Mobile unmount. Appearance migration must pass both unchanged;
do not expand this contract to unused custom scrollers.

Error/Dock appearance belongs to `components/error.less` and
`components/page-scroll-dock.less`; their SFC scoped styles own private structure
only. Reuse local `--gf-error-*` / `--gf-scroll-dock-*` semantics, keeping dynamic
delay/progress channels and the P4.5 behavior/visual contracts unchanged.

`visual/footer-shell.spec.ts` owns the two Desktop Footer baselines and exact
App Shell color-transition semantics. Derive the clip from real columns to exclude
dynamic current-year meta; computed assertions protect meta/link appearance,
heading/icon semantics and social hover colors. Migration must pass unchanged.

`regression/nav-shell.spec.ts` owns Nav Shell interactions; `visual/nav-shell.spec.ts`
owns Standard/Overlay/Mobile Menu/BottomTab appearance through the shared fixture.
Keep its audited Home/Mode side-effect exceptions precise (see testing guidance).
Hero lifecycle, SearchBox, QuickAccess, Nav content and Site Groups remain separate
P5 contracts. The eight new goldens require maintainer review before P5.1.2.

Never update visual baselines merely to make CI pass. An explicit maintainer
approval or a task authorizing the visual migration is required before running
`test:visual:update` in the pinned environment. CI only compares. Review package,
Docker tag/digest, browser revision and baselines together on Playwright upgrades.
Review the complete diff for accidental production Vue/style changes before commit.

[Frontend documentation](../../../docs/frontend/README.md) routes to the contract,
debt state and later design/testing guidance.
