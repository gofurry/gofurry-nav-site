# #149 / #151 — Dark canvas and edge dots

Date: 2026-10-23. Baseline: `989ccad`, existing clean `dev`.

## Scope

Implemented A → B → C: neutralize the Dark canvas and directly conflicting
surface tokens, add CSS edge dots, then retire the WebP and its contracts.
Light stays `#f2e3d0`; Dark is `#1d1f22`. Shared surfaces keep their alpha and
hierarchy. Dark inputs use RGB 32/34/38; the existing modal layers use RGB
38/40/44 → 25/27/30 over 32/34/38. Dark navigation uses RGB 20/22/25;
Footer retains its deeper independent canvas, `#0d0e10` → `#090a0c`.
Light surface values, text/focus/brand/status colors, navigation geometry,
Hero and all page business logic are unchanged.

Dot RGB is Light `#806c58`, Dark `#c1c5cb`. Maximum alpha before masks is
0.24 / 0.22 on wide desktop, 0.16 / 0.15 at 768–1279px and 0.12 / 0.10 below
768px. Desktop uses 26px cells / 0.8px radius, each edge `min(22%, 340px)`;
the center remains at least 56% clear. Vertical fade ends at 520px. Mobile
uses one full-width 28px grid fading by 185px. Nested linear masks include
WebKit declarations; no mask compositing, blur, animation, image or JS is used.

The solid canvas remains fixed; decoration is absolute at the shell's page
origin and scrolls away. The default layout remains the only instance owner.
Deleted the unused `public/web/background/background.webp` and old image tokens.
The user's original external file is untouched. Fixture startup now probes
`/defaultLogo.svg`. Existing asset smoke/Insights assertions and the affected
Nav/review-dialog color expectations use the new contract without exemptions.

## Verification

- PASS: scoped ESLint for changed Vue, Browser fixtures/specs and scripts.
- PASS: Stylelint for the background Vue and three affected Less owners.
- PASS: `pnpm run style:policy` (no baseline changes).
- PASS: `node scripts/insights-semantics.test.mjs`.
- PASS: 13 focused `static-page-background.spec.ts` cases across local source runs:
  8 viewport/theme + 2 Home + Games in `browser-first-11.log`; Sites and Insights
  in the final `browser-remaining.log` (2 passed). The product source was unchanged
  while correcting the new test setup; the 11 passed cases were not repeated.
- NOT RUN: production build, typecheck, full Browser/Visual, cloud acceptance,
  Docker and GitHub Actions. No Golden was changed.

Browser command: `GOFURRY_FIXTURE_DEV=1 GOFURRY_149_REVIEW=1 pnpm exec playwright
test tests/browser/regression/static-page-background.spec.ts --workers=1
--timeout=120000 --max-failures=1` (environment variables set through PowerShell).
This is local Nuxt source evidence, not production/pinned Visual acceptance.
After setup corrections, only the two remaining cases used the same command
with `--grep '(Sites|Insights) shares'`. An initial anchored grep selected no tests;
it was corrected without changing test selection or assertions.
The owner covers 390/768/1024/1440, both themes, SSR singleton/no hydration errors,
scroll geometry, theme changes, Preferences and old-preference isolation. Home
has separate 1440 Dark / 390 Light cases. Games and Insights cover their normal
data; Sites uses its real Site Groups owner in the empty state to keep background
checks independent of pagination. Their network/error guards remain intact.

Earlier diagnostic attempts are retained under ignored `.cache/149-c1-review/`:
one local `ERR_NETWORK_CHANGED` interruption and new-test assumptions corrected
to the real contract (Terms max scroll 562px, so test the 520px fade boundary;
Preferences closes through Cancel, not Escape; normalize signed zero in the exact
scroll equation). The SEO-only fixture did not supply browser Ping; Sites now
uses the real Site Groups empty-state fixture, without entering its auto-pagination
scenario. No product workarounds or error waivers were added. Screenshots/logs stay
in the same ignored directory.

## Maintainer review — PENDING

Review edge-dot visibility and neutral Dark surfaces on Home, populated Site Group, Games
and Insights at all four widths; inspect navigation, cards, Footer and Preferences
while switching themes and scrolling. The masked decoration should disappear
without affecting the solid canvas. Screenshots and automated assertions do not
constitute maintainer visual sign-off. Existing accepted Goldens require separate
review before any future update. #149 and #151 remain open; no deployment.
