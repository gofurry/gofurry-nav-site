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
2. Before adding appearance, search `app/assets/styles/tokens.less` for an
   existing `--gf-*` semantic token and `app/assets/styles/components/` for an
   existing shared primitive. Search the relevant domain styles under `pages/`.
3. Reuse `.gf-button`, `.gf-card`, `.gf-input`, `.gf-chip`, `.gf-modal`,
   `.gf-pagination` and `.gf-rating`; use their existing variants before adding one.
4. Keep control height/padding, typography, colors, radius, shadow and hover/focus
   in the owning primitive/Less layer, not a new bundle of Tailwind visual classes.
5. Domain tokens should alias global tokens. Add an independent value only for
   a documented domain meaning. Component tokens are for local geometry/behavior,
   not a second color/theme system.
6. Token comments explain semantics, scope and ownership. Use group comments;
   add an individual rationale for exceptional tokens, not comments naming colors.

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
- Exceptions need exact `path`, `rule`, `issue`, `reason`, and `remove_when`.
  A wildcard ignore or an old visual-guard allowlist is not a new style exception.
- #108 Insights and #109 Site Detail are excluded from proactive P4–P6 migration,
  not debt-free areas. Measure their existing debt; new code follows the contract.
- P0 only establishes governance and the one-time baseline. Do not change UI,
  production CSS/Less/Vue, dependencies, CI, scripts or style directories for P0.
  Keep current `components/*.less` and regression harnesses until their own phases.
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
npm run typecheck
npm run insights:semantics
npm run seo:recovery:test
npm run build
```

Check `package.json` for existing focused checks and run those relevant to the
changed area. Keep regression coverage beside the existing `scripts/` harnesses
until the testing migration; do not assume future runners or commands exist.
P1 uses Node built-in tests for the policy guard; Vitest and Playwright Test
migration remain P3 work. Static checks also run in the existing Nav Web CI job.
Review the complete diff for accidental production Vue/style changes before commit.

[Frontend documentation](../../../docs/frontend/README.md) routes to the contract,
debt state and later design/testing guidance.
