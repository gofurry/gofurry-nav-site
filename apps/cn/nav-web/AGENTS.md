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
`--gf-page-background`; do not reuse legacy `--gf-bg-page` in new code.
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

`<style scoped>` is valid for component-private structure. It must not redefine
shared primitive appearance or invent a local theme. Repeated appearance should
move from component to domain to shared primitive as reuse warrants it.

Do not introduce new raw colors, radii, shadows, visual durations or typography
scales in ordinary selectors. Reuse the semantic token owner first. Theme changes
belong in semantic variables under the existing `html.dark` model, not new
`dark:bg-*`/`dark:text-*` variants or independent page-dark classes.

## Historical debt and phase boundaries

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
`tests/browser`; build production Nitro before running them. P3.2.1 migrates only
Game Detail to Playwright Test. Hero/Resource/Insights and other legacy smokes stay
until their own migration. Chromium is the browser gate; retries are zero and CI
uses one worker. Domain fixtures own servers and reset state per case; Playwright
owns contexts/pages. Failure artifacts are diagnostics, not visual baselines.
Style-policy keeps its separate Node built-in runner; Insights/SEO remain
Contract Guards. Check `package.json` for relevant focused commands. CI runs
Unit and Nuxt tests as independent steps alongside the existing guards, followed
by Build, Chromium installation and Browser tests.
Review the complete diff for accidental production Vue/style changes before commit.

[Frontend documentation](../../../docs/frontend/README.md) routes to the contract,
debt state and later design/testing guidance.
