# #149-C2 / #151-B — Background and Dark surfaces

2026-10-23; existing `dev`, clean baseline `f35e404`. Local source-fixture evidence,
not production/CI or maintainer visual approval. Earlier C1 evidence is unchanged.
Result: implementation complete; runtime acceptance PARTIAL / human visual PENDING.

## Implementation and audit

- One layout background owner, both solid and decorative layers fixed at z-index 0,
  below content and pointer-inert. No scroll listener, new image or client state.
- Desktop grid 30px, diameter 2.2px, each edge `min(30%, 460px)`, vertical fade
  720px clamped to `100svh`. Horizontal 15% / vertical 18% opaque plateaus avoid
  premature combined fading; at least 40% of the viewport stays dot-free.
- Existing colors retained: Light `#f2e3d0` + `#806c58` / 0.24, Dark `#1d1f22`
  + `#c1c5cb` / 0.22. Tablet alpha 0.16 / 0.15; mobile grid 28px, diameter 1.6px,
  height 185px and alpha 0.12 / 0.10, with only downward fading.
- Header's opaque theme-color base covers fixed dots while its unchanged Hero
  image loads or fails. Existing image/masks, scroll and navigation remain owned
  by their existing components; content reveals the dots through normal flow.
- Dark-only local tokens neutralized for Home cards, Spotlight, transition bar,
  group cards/popovers; Games cards/news/stats, sidebar suggestions, Search fields,
  Filter/datepicker/Jump surfaces, Detail actions/tooltips; About/Legal reading
  panels and Preferences controls. Shared surface/hover aliases are reused where
  roles match. Local opaque overlays use RGB 40/42/45, 48/50/54 and 55/57/61.
- Site Detail, Insights (Overview/Domain/Workspace/Compare/Changes/Pulse), Updates
  and Collections were audited: shared neutral surfaces or Games aliases already
  own their backgrounds, so no new local overrides were added.
- Retained intentionally: Hero search/Quick Access image-overlay materials, media
  canvases/lightbox scrims, chart/status/brand/accent colors, focus rings, shadows,
  border/text colors, the separate immersive Lottery/error palettes and existing
  navigation/footer materials. Near-neutral transparent layers need no rewrite.
- Prior C1 already removed the unused WebP; no texture asset/reference was added.
  Current README/design/asset/testing contracts updated; no Golden updated.

## Verification

- PASS: file-scoped `pnpm exec eslint` (changed Vue/script/test owners),
  `pnpm exec stylelint` (background plus six Less owners), `pnpm run style:policy`
  and `pnpm run insights:semantics`; policy debt unchanged.
- PASS: source diff check of the five local surface owners confirmed non-Dark
  declarations unchanged, except the explicit common Header base for occlusion.
- PASS: 14 core background cases: ten width/theme cases, two Home fallback
  cases, Games and empty Site Group. These cover 1920/1440/1024/768/390px,
  fixed positions after scrolling, masks, singleton SSR DOM, old preference/IDB
  isolation, theme changes, two-tab Preferences and short-height behavior.
- PASS: existing Games Home Light/Dark SSR/paging/statistics/hover/review cases,
  and Game Detail real chart theme/tooltip/resize case. Initial Detail Tab
  assertion failed during a source-server reconnect; the unchanged case passed
  on the targeted follow-up. No retry configuration or assertion was relaxed.
- FAIL / PENDING: English Search 1440 Light timed out at its source-fixture Vue
  readiness wait (120s), before appearance assertions; later Search and populated
  Site Group cases in that batch were NOT RUN. No further Search rerun/build.
- FAIL / PENDING: Insights domain background geometry/scroll/theme checks completed,
  but viewport changes to 1024/768/390 triggered strict `assertQuiet`, which caught three `[ECharts] cartesian2d cannot be found for
  series.line (index: 0)` errors. Do not count this case as passing. No Insights
  chart registration, data, error accounting or assertion was changed to hide it.
- PASS: desktop 1440 Dark Hero pixel control: hiding dots changes no Hero pixel,
  but does change the uncovered content-edge sample after normal scrolling.
  PASS: mobile 390 Light control also passed. Its initial test sampled y=200..320,
  outside the intentionally retained 185px mobile dot band; the sample was corrected
  to y=70..140; only that case was rerun. No product CSS or assertion of visible
  dots was weakened. Total: 19 distinct cases passed; Search and Insights remain
  unresolved checks, not a green runtime gate.
- NOT RUN: typecheck, production/Nitro/Docker builds, full Browser/Visual, CI
  dispatch/query/wait, production data/assets/deployment. No dependency/Golden
  changes. Runtime checks use existing `GOFURRY_FIXTURE_DEV=1` owners locally,
  Chromium, `--workers=1 --timeout=120000 --max-failures=1`, not production approval.
  Commands/logs and optional screenshots are under ignored `.cache/149-c2-review/`:
  `browser.log`, `browser-final.log`, `background.log`, `hero.log`, `hero-mobile.log`. The Insights
  failing trace/screenshot are preserved in `insights-failure/`.

## Maintainer visual review — PENDING

Review Light/Dark at 390/1024/1440/1920px: first Home Hero, transition/content,
long-scroll dot persistence, low-height fades, card hover/focus, Footer and
Preferences. Also review populated Site groups/Detail; Games Search/date picker,
Detail/Collections; Insights major pages; Updates and About/Legal. Local fixture
screenshots cannot certify production assets, every page state or personal visual
acceptance. Accepted PNG baselines are intentionally unchanged; any future Visual reconciliation
needs separate reviewed authorization. No production deployment or issue closure
belongs to this change.
