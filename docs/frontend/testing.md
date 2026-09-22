# Nav Web testing

Use the lowest-cost environment that faithfully represents the behavior under
test. The [frontend contract](../../contracts/nav-web-frontend.md) owns this rule;
the commands below run from `apps/cn/nav-web`.

## Choose the test owner

| Category | Owner / command | Boundary |
| --- | --- | --- |
| Pure TS/domain/utility logic | `tests/unit/*.test.ts`, `npm run test:unit` | Vitest `unit` project, Node environment; no Nuxt boot in test cases |
| Nuxt composables and runtime | `tests/nuxt/*.nuxt.test.ts`, `npm run test:nuxt` | Vitest `nuxt` project, real Nuxt app/context through `@nuxt/test-utils`, happy-dom |
| Repository Contract Guards | `npm run insights:semantics`, `npm run seo:recovery:test` | Existing Node scripts inspect source/config/docs and semantic contracts |
| Style Policy Tooling Tests | `npm run style:policy:test` | `node --test scripts/style-policy/*.test.mjs`, independent of Vitest |
| Playwright Browser Tests | `tests/browser/{smoke,regression}/*.spec.ts`, `npm run test:browser` | Production SSR/hydration/interactions; Game Detail, Resource Routing/Managed/Steam and Hero/Preferences are migrated |
| Playwright Visual | `tests/browser/visual/*.spec.ts`, `npm run test:visual` | Pinned sentinel; Foundation, Preferences and Error locator baselines; Static/Legal and Updates viewport baselines; Dock expanded clips |
| Legacy visual/report guard | `npm run visual:guard` | Broad page/selector/theme/overflow reports and historical checks; retained independently of Visual and `style:policy` |
| Legacy Browser Smoke | Existing non-migrated `*:smoke` scripts, after `npm run build` | Insights and unrelated domains retain their runners until scoped migration |
| External Acceptance | Explicitly authorized development/provider checks | Real services, separate from deterministic fixtures and normal CI |

`npm test` runs both Vitest projects once. `vitest.config.ts` uses `projects`, not
the deprecated workspace model. Nuxt configuration is loaded by the test-utils
project; the test module is **not** added to production `nuxt.config.ts`.
Dependencies are pinned in package/lockfiles. No coverage provider or Testing
Library is introduced. Playwright Test and the retained direct `playwright` use
matching 1.60.x versions; legacy smoke/performance still import `playwright`.
Nuxt cases have a 30-second budget for the first mount's cold app/router transform;
unit cases retain Vitest's default timeout. Nuxt's generated `.nuxtrc` module setup
marker is local and ignored, like `.nuxt/`.

## Migrated regression knowledge

| Former `scripts/` suite | Current owner | Preserved contracts |
| --- | --- | --- |
| `managed-assets.test.mjs` | `tests/unit/managed-assets.test.ts` | Key validation, selection threshold, fallback/duplicate URLs, full probe body and digest, background preference normalization and explicit overrides |
| `game-tag-domain.test.mjs` | `tests/unit/game-tag-domain.test.ts` | Explicit/empty categories, selected state, expansion preservation, Adult code rather than historical numeric ID |
| `hero-preferences.test.mjs` | `tests/unit/hero-preferences.test.ts` + `tests/nuxt/use-hero-preferences.nuxt.test.ts` | Bigint-safe IDs, normalization, SSR query/key, actual cookie round trip, retained viewport IDs, revision/legacy state and unrelated Save/cookie isolation |
| `resource-routing.test.mjs` | `tests/unit/resource-routing.test.ts` + `tests/nuxt/asset-composables.nuxt.test.ts` | Pin/recommendation separation, Steam groups/sample fallback, diagnostics, mount/idle/deduplication/TTL/cooldown/offline/Save-Data, per-resource snapshots, Hero handoff and real error fallback |

The Nuxt cases also explicitly cover Local-file revision changes and exhausted
fallback handoff. Probe averaging tests retain successful/failed/timeout meaning.
The single `tests/fixtures/cdn-probe.bin` serves Unit and the existing Hero/routing
browser fixtures; its size/digest are checked by the managed-assets suite.

The four legacy test implementations and unused `assets:test` / `game:tags:test`
aliases are removed. Use the two project commands or `npm test`; do not recreate
parallel test implementations. Node Contract Guards and style-policy tests are
intentionally not discovered by Vitest.

## Runtime fidelity and isolation

Import production utilities/composables normally. Do not read/transpile source,
rewrite imports into data URLs, replace global `useState`/`useCookie`/`useNuxtApp`,
or build a custom renderer to simulate Nuxt.

Nuxt tests mount a small behavior harness using `mountSuspended`; these exercise
real composables, auto-imports, cookies, state and i18n. The harness exists to
provide setup/unmount lifecycle, not to assert trivial component markup.
`tests/nuxt/setup.ts` uses Vitest module mocks only for the two business provider
plugins (`$assetCDN`, `$steamAssetRoute`), preventing real background CDN probes.
Nuxt itself, its context, cookie/state APIs and plugins outside those boundaries
remain real.

Every case unmounts its harnesses in `finally`. Shared setup removes Hero/route
cookies, calls public `refreshCookie` / `clearNuxtState`, clears browser storage
and mock calls before/after cases. Provider refs reset per asset test. In
particular, `hero-preference`, `hero-legacy-pending` and
`hero-preference-revision` must never leak between cases. Do not manipulate
hidden Nuxt payload/cookie internals or rely on case ordering.

Browser-like globals alone do not require Nuxt: scheduling tests stub
`navigator` and `window.requestIdleCallback` with `vi.stubGlobal` and restore them
after each case. Pure probes use injected measurements/fetch responses, never
real CDN or backend endpoints.

happy-dom does not prove browser rendering, SSR/hydration or actual image loading.
Keep `game:tags:smoke` and other non-migrated smoke commands.
Unrelated screenshots and `visual:guard` remain; this phase does not introduce
visual baselines or external acceptance runs.

## Game Detail browser gate (P3.2.1)

```sh
npm run build
npx playwright install chromium
npm run test:browser:smoke
npm run test:browser:regression
npm run test:browser
```

`npm test` remains Vitest-only. Browser commands use `playwright.config.ts`:
Chromium only, zero retries, one worker in CI, normal isolated contexts/pages.
There is no global `webServer`: `tests/browser/fixtures/game-detail.ts` starts
one production Nitro app and loopback API per worker via the unchanged
`scripts/fixtures/insights-app.mjs`. `app.close()` runs in worker teardown.
The automatic per-test fixture resets `legacyNull=true`, `failure=false`,
`gallery=false`, `adult=false` and clears recorded requests before/after use.
Its `baseURL` also serves Playwright's request fixture. Specs configure scenarios
through that state; they never launch/close browsers or create contexts themselves.
The page fixture fulfills only the known managed/Steam background probe endpoints
with local binary/SVG fixtures, keeping real plugins active without external CDN
availability affecting this domain gate.
`browser-errors.ts` explicitly captures page exceptions, hydration mismatch
messages and console errors for assertions in each browser scenario.

The former `scripts/game-detail-smoke.mjs` and unused `game:detail:smoke` alias
are retired after this coverage split:

| Former assertion | Current owner |
| --- | --- |
| `/games`, `/en/games`, `/games/82`, `/en/games/82`: HTTP 200, title/description, unique canonical/path, en-US alternate, no noindex, SSR heading/intro/home content | `smoke/game-detail.spec.ts`: four SSR cases |
| Nuxt hydration, core tabs, missing players/prices, rendering errors | Smoke core navigation case |
| `regional_prices.regions` null versus empty array, missing != zero, repeated tabs/labels | `regression/game-detail.spec.ts`: two data cases |
| All major tabs keep main width, viewport containment, desktop 75/25 and sidebar bounds | Regression at 390/768/1440/1920, ordinary/adult content |
| Gallery/media stay inside parent, thumbnails scroll locally with overflow auto, last ordinary thumbnail activates final image | Same eight layout cases |
| NSFW modal contained, Cancel closes without unlocking | Four regression cases: 390/1440 × light/dark |
| Unknown/nonnumeric IDs return 404, upstream failure returns 503 in both locales | Five regression status cases |

The API uses the same no-Facts and 24-image fixtures as the legacy script. This
guards browser behavior, not pixel equality. Do not port success screenshots,
`modal-computed.json` or add `toHaveScreenshot()`; visual baselines belong to P3.3.
Playwright retains trace/screenshot only on failure (video off), writes HTML to
`playwright-report/` and diagnostics to `test-results/`, both ignored. Fixture
server logs are attached to failing cases. CI uploads both directories only on
Browser tests failure. Shared fixture/performance helpers remain unchanged.

## Resource Routing browser regressions (P3.2.2)

The same three browser commands discover Resource Routing automatically through
the existing single CI Browser tests step. No new runner, project or CI step is
needed. `tests/browser/fixtures/resource-routing.ts` reuses
`startInsightsFixtureApp`: the worker owns only production Nitro/local API and
fixed icon/Hero/pattern/Steam payloads. Every test receives fresh context/storage,
probe latencies (primary 160 / mirror 15 / china 180 / global 15 ms), failure flag,
blocked URL, probe request evidence and an independent gate. `releaseProbes()`
lets a test first prove successful resource rendering, then allow recommendations
to change. Teardown always releases the probe gate, then leaves route disposal to
Playwright's test-scoped context close. It must not wait for already-handled probe
routes with `unrouteAll({ behavior: 'wait' })`, which can stall that context close.
The worker always closes its app. No serial-suite dependency.

The fixture owns all network interception: exact Managed fixture origins and
known Steam origins return local binary/SVG responses, and unknown external
requests are aborted. Managed probes use `tests/fixtures/cdn-probe.bin` with
the real digest and CORS path. Real routing plugins remain active. Cookies and
background/optional legacy Steam storage are seeded before navigation/hydration.
The narrow prop bridge simulates new `objectKey`/`src` on the same mounted
resource component; it is test-only and not a shared Vue introspection API.

`browser-errors.ts` remains the error guard. Fault tests supply only URLs they
deliberately failed; only the exact Chromium `net::ERR_FAILED` or injected HTTP
503 console diagnostic for those URLs is expected. Page exceptions, hydration errors, unknown-network
errors and other console errors still fail. Fallback cases also assert that the
injected failure occurred and that the fallback image actually loaded.

The former `scripts/resource-routing-smoke.mjs` is retired with this mapping:

| Former assertion | Current owner under `tests/browser/` |
| --- | --- |
| Managed and Steam SSR explicit pin takes precedence; Routing UI hydrates with two sections | `smoke/resource-routing.spec.ts` (3 cases) |
| Three tabs, Home/End/wrapping arrows and focus; Cancel/Save pins; manual diagnostics/cooldown survive Cancel; 390px no overflow/focus/light/dark; English failed pins with two warnings | `regression/resource-routing-preferences.spec.ts` (4 cases) |
| Loaded icon/Hero/Pattern survive automatic/manual/Save; icon src MutationObserver; new key uses current pin; actual mirror failure falls back to primary without clearing pin | `regression/managed-asset-routing.spec.ts` (1 lifecycle case) |
| Loaded Steam image survives automatic/manual/Save; new src uses Global; first host failure falls back without clearing pin; legacy storage migrates only to recommendation, is deleted, retains hydrated China image and Auto UI | `regression/steam-asset-routing.spec.ts` (2 cases, legacy migration separate) |

No Resource Routing success screenshots or pixel assertions are migrated.
Insights/background/cloud/performance/visual legacy runners retain their owners.
Visual baselines remain P3.3.

## Hero / Preferences browser regressions (P3.2.3)

`hero-lifecycle.ts` and `hero-preferences.ts` remain separate domain fixtures.
Each worker owns stable production Nitro/local API resources and a nullable
active-scenario reference. A test installs a fresh scenario only when that slot
is empty; the resolver rejects calls without an active scenario. Counts, image
requests, failures and gates belong to that scenario. Preferences receives 48
new Desktop items and two new Mobile items per test, so key replacement and
failed-page injection never change a later case. IDs stay strings, including
`9007199254740994`.

Lifecycle gates hydration scripts until the real renderer request fails, traps
auxiliary `new Image()` authority, and can release the two Local readonly opens
separately. Local data is a real cached Blob at
`gofurry-custom-nav-header-bg` / `directoryHandles` / `nav-header-bg-cache`, plus
`customNavHeaderBgFolderName`; no DirectoryHandle is faked. Preferences seeds the
cache once so reload tests exercise actual persistence. Its API and image gates
are independent: cookies can commit while the old frame remains displayed, and
the pending frame promotes only after its real image loads. Teardown releases
all pending gates, awaits unroute and clears the active scenario, even on failure.
The worker always closes its app; Playwright owns browser/context/page lifetime.

Both fixtures reuse `startInsightsFixtureApp`, the managed probe binary and
`captureBrowserErrors`. Hero and known probe URLs receive deterministic local
artwork/binary responses; unknown external requests abort. Intentional image
failure and Catalog 503 URLs enter the exact expected-failure set before the
failure, without suppressing unrelated network or application errors.

The known narrow-screen homepage Footer mismatch remains production debt.
`assertHeroHydration` is Hero-specific and verifies all of: width < 768, homepage,
exactly one hydration mismatch, no `.gf-footer-shell` in SSR, exactly one client
footer, the same SSR Hero node, and zero other browser errors. Only that verified
navigation evidence is consumed; subsequent errors remain guarded, and reload
must pass the assertion again. The generic collector still captures hydration.
`app/layouts/default.vue` is unchanged.

| Retired source / contract | New owner under `tests/browser/` |
| --- | --- |
| Lifecycle: 1440/390 SSR, actual image load, viewport-only request, retained node, one Home and no Hero API refetch | `smoke/hero.spec.ts` |
| Lifecycle: no auxiliary Hero Images, auxiliary failure leaves URL/paint unchanged, recommendation/Save/focus/resize/prewarm stability, explicit data refresh, pre-hydration Primary→Mirror/terminal fallback, empty Mobile pool | `regression/hero-lifecycle.spec.ts` |
| Lifecycle + Preferences: two-stage Local restore and route stability; explicit Local SSR, Cancel, Local→Random, missing cache one-shot fallback, legacy cookie migration/reload | `regression/hero-local-background.spec.ts` |
| Preferences: Quick Access ARIA/Space/Enter/focus and draft/persistence; input focus; 1440/390/320 × light/dark no overflow, equal source widths, shared computed appearance | `regression/preferences-foundation.spec.ts` |
| Preferences: independent Fixed Desktop/Mobile SSR and exact BigInt cookie, saved Local isolation, reload, same ID/new key, invalid Desktop preserves valid Mobile | `regression/hero-preference-modes.spec.ts` |
| Preferences: lazy Fixed initialization, page size 12, boundary/cache/preview laziness, independent viewport state, Cancel, bounded ID lookup, compact tab keyboard, final Next disabled, 503/retry without saving | `regression/hero-catalog.spec.ts` |
| Preferences: independent API/image gates, no blank-first promotion, no unrelated Home request, Mobile-only + Mirror preserves successful Desktop, failed Mirror→Primary stage retains old frame | `regression/hero-handoff.spec.ts` |

`scripts/hero-lifecycle-smoke.mjs`, `scripts/hero-preferences-smoke.mjs` and
`assets:routing-smoke` are retired. Direct `playwright` remains for other legacy
tools. The existing single Browser tests CI step discovers these specs without
runner/browser/retry/worker changes.

The old `hero-*-before.png`, `foundation-*-on/off.png`, Hero Preferences success
screenshots and `foundation-computed.json` writes are retired. Only same-run
clipped screenshot **Buffers** (never files), selected-control computed-style
equality and layout/focus assertions remain. These guard runtime invariants,
not golden visuals. No `toHaveScreenshot()` or committed images are introduced;
P3.3 owns visual baselines.

## Visual runner and pinned environment (P3.3.1)

`playwright.config.ts` explicitly ignores `tests/browser/visual/**`; the 63
functional Browser cases remain Smoke/Regression only. The separate
`playwright.visual.config.ts` owns Chromium Visual tests: headless, one worker,
zero retries, 60-second timeout, 1440×900, zh-CN, UTC, DPR 1, light default and
reduced motion (`use.contextOptions.reducedMotion` in Playwright 1.60).
Traces/screenshots are failure-only, video is off. Reports go to
`playwright-visual-report/` and diagnostics to `visual-test-results/`, both ignored.

The authoritative identity is Linux amd64 / Ubuntu Noble with Node 24 and
`@playwright/test`, `playwright`, `playwright-core` 1.60.0 (the committed lockfile).
The bundled Chromium/headless-shell revision is **1223**, browser **148.0.7778.96**.
Use the official image:

```text
mcr.microsoft.com/playwright:v1.60.0-noble@sha256:9bd26ad900bb5e0f4dee75839e957a89ae89c2b7ab1e76050e559790e946b948
```

The MCR manifest-list digest was verified on 2026-09-19 against the SHA-256 of
the returned manifest bytes and `docker buildx imagetools inspect`. Its linux/amd64
manifest is `sha256:83192064c7510f7ee73dd63dc5f22a5e01a92c81a2e6a9c715d9e3fe55471fd9`.
The image contains Node 24 and browser/system dependencies; CI still explicitly
selects Node 24 with `actions/setup-node@v6`. The [Playwright Docker guidance](https://playwright.dev/docs/docker)
describes the matching-package/image requirement and `--ipc=host`.

`nav-web-visual` runs only for Nav Web changes after `nav-web` succeeds. It uses
that tag **and** immutable digest with `--ipc=host` and `GOFURRY_VISUAL_ENV=pinned`,
then independently runs `npm ci`, `npm run build`, `npm run test:visual`. It neither
installs browsers nor transfers `.output` from the functional job. On failure,
`nav-web-visual-failure` contains the distinct Visual report/results for seven days.

The environment sentinel uses a real Chromium page to check browser type,
viewport, DPR, language, timezone, reduced motion and light theme; CI/pinned runs
also check Linux and Node 24. It takes no screenshot and needs no server or
production test route. **P3.3.1 has zero golden snapshots and no screenshot
assertions.** P3.3.2 adds the Foundation contract below; P3.3.3 adds real
Preferences composition.
Direct `playwright` and `scripts/perf/visual-guard.mjs` / `visual:guard` remain active.
Static architecture/debt remains `style:policy`'s responsibility.

### Comparison and approved updates

Local `npm run test:visual` is diagnostic only outside the pinned environment.
Comparisons use `updateSnapshots: 'none'`, so missing baselines also fail instead
of being created. Snapshots belong in tracked
`tests/browser/visual/__snapshots__/{testFilePath}/{explicit-name}.png`; do not
ignore this source directory. No global pixel tolerance is configured. Future
tolerances must be small, local and justified for an individual snapshot.

**A baseline update is an explicit visual-change review action, not a test-fix
command.** Agents must not update baselines just to make CI green. A maintainer
must explicitly approve the visual change, or the task must explicitly authorize
that migration, before `test:visual:update` may run. Investigate unexpected
differences, fix the implementation and rerun comparison. Review Playwright
packages, Docker tag/digest, browser revision and baselines as one atomic upgrade.

`test:visual:update` first runs `scripts/visual-baseline-env.mjs`, requiring Linux,
Node 24 and `GOFURRY_VISUAL_ENV=pinned`, before invoking `--update-snapshots`. The
marker asserts that the caller chose the pinned container; it does not grant
approval or fingerprint the image. CI must never invoke the update command/flag.

From a disposable checkout's repository root in a POSIX shell, this compares
inside the pinned image. An empty anonymous volume isolates Linux dependencies.
Nuxt must own normal writable `.nuxt` / `.output` directories in that checkout;
do not mount `.output` itself, because Nitro removes and recreates it on build.

```sh
docker run --rm --init --ipc=host --platform linux/amd64 \
  --mount type=bind,source="$(pwd)",target=/work \
  --mount type=volume,destination=/work/apps/cn/nav-web/node_modules,volume-nocopy \
  --workdir /work/apps/cn/nav-web \
  --env GOFURRY_VISUAL_ENV=pinned \
  mcr.microsoft.com/playwright:v1.60.0-noble@sha256:9bd26ad900bb5e0f4dee75839e957a89ae89c2b7ab1e76050e559790e946b948 \
  sh -lc 'npm ci && npm run build && npm run test:visual'
```

Only for an explicitly approved baseline change, use the same command
with its last line replaced by:

```sh
  sh -lc 'npm ci && npm run build && npm run test:visual:update'
```

Review the resulting snapshot diff before committing. P3.3.1 does not run this
update flow or create a `__snapshots__/` directory.

### Shared Primitive Foundation (P3.3.2)

`visual/fixtures/ui-foundation.ts` reuses `startInsightsFixtureApp` with a
worker-owned production Nitro/local upstream. `/about` supplies the real built
stylesheets; the fixture preserves its head and replaces only the body. A
test-only response CSP (`script-src 'none'`) blocks all product scripts, including
inline bootstraps, so hydration/plugins cannot affect the result. Playwright 1.60
with `javaScriptEnabled: false` also blocks style-load callbacks and animation
frames; CSP permits test evaluation and the required two-frame wait without
running the app. External requests are blocked and fail the fixture sanity check.

Fixture CSS owns grid/flex/spacing geometry only. Its sole canvas exception uses
`--gf-page-background` / `--gf-text-main`; controls consume production CSS/tokens.
Button, Card, Input, Chip, Pagination, Rating and Generic Modal are generic;
Preferences Toggle alone receives a local `.gf-preferences-modal` token scope.
Rating mirrors production's empty/fill DOM, including the partial fifth star.
Token availability, scope isolation, focused input and overflow are asserted.
Fonts and two animation frames settle before locator capture; other controls
remain idle. There are no masks, custom fonts or screenshot tolerance overrides.

`ui-foundation.spec.ts` owns exactly these four tracked images under
`tests/browser/visual/__snapshots__/ui-foundation.spec.ts/`:

- `foundation-light-desktop.png` and `foundation-dark-desktop.png`: 1440×900.
- `foundation-light-mobile.png` and `foundation-dark-mobile.png`: 390×844.

Dark uses production `html.dark`, while browser color scheme stays light. The
P3.3.2 gate has five cases: one environment sentinel plus four Foundation
cases. The 63 functional Browser cases retain their existing ownership.
Real Preferences/backdrop composition belongs to P3.3.3; business surfaces to
P4+. Do not turn Foundation into a page gallery.

Initial baseline creation is an approved visual-contract action. Use the pinned
container/update guard above, then run `npm run test:visual` **twice consecutively**
without updating. Both comparisons must pass. Investigate any difference rather
than regenerating it away. Maintainers must review all four PNGs for correct
primitive states, independent Generic Modal/Toggle scope and mobile fit before
accepting the visual contract; no full-site walkthrough is required.

### Real Preferences / Modal (P3.3.3)

`visual/fixtures/preferences-visual.ts` extends the existing `hero-preferences.ts`
fixture. It reuses production Nitro, real Nuxt/Vue hydration, NavBar interaction,
Teleport, backdrop and all real Preferences children/scoped styles. It does not
copy API/network fixtures or fake Modal markup. Optional `height`, `theme`,
`fixedNow` and `seedSteamDiagnostics` open parameters leave functional defaults
unchanged. Theme is seeded in localStorage before navigation and initialized by
the real Theme Store, including Pinia state and `html.dark`.

After opening the Modal, only `#__nuxt` is hidden and the body canvas uses
`--gf-page-background`. No Preferences appearance is overridden. Real tab clicks
settle on ARIA selection, non-inert panel and carousel scroll position. Idle
captures blur focus, move the mouse away, await finite animations/fonts/two frames,
then verify viewport backdrop coverage/filter, Modal containment and active-panel
overflow. The existing browser-error guard and narrow mobile Footer debt check
remain active. Functional keyboard/persistence/Catalog/handoff coverage stays in
`tests/browser/regression`; Visual does not duplicate it.

Modal geometry uses frame bounds plus header/tab/active-panel overflow checks.
Its aggregate `scrollWidth` also includes the offscreen Routing `sr-only` legends:
their absolute containing block is the Modal because of `backdrop-filter`, despite
their clipped pixels. Do not misclassify those accessible labels or the intentional
three-page carousel as visible overflow; do not alter production CSS to remove them.

Routing alone fixes browser `Date.now()` to `VISUAL_FIXED_NOW` and seeds fresh
Managed 10/20 ms and Steam 30/60 ms diagnostics. Existing Primary/Global pins
remain; recommendations are Primary/China. The wrapper checks zero probe traffic,
stable measurements and collapsed Details. There are no probe gates, screenshot
sleeps, masks or tolerance overrides.

`preferences-modal.spec.ts` adds exactly eight images under
`tests/browser/visual/__snapshots__/preferences-modal.spec.ts/`:

- `preferences-home-{light,dark}-{desktop,mobile}.png` (four): Random Hero,
  Quick Access on and the default display-mode input.
- `preferences-background-light-{desktop,mobile}.png` (two): default pattern,
  no artificial pattern gallery.
- `preferences-routing-light-{desktop,mobile}.png` (two): fixed diagnostics.

Desktop is 1440×900 and Mobile 390×844. Each locator captures the real
`.gf-modal-backdrop`, including centering and outer padding. The Visual suite has
**13 cases** (sentinel + 4 Foundation + 8 Preferences); the existing Visual CI job
discovers them. Foundation images remain unchanged. NSFW, catalogs, local pickers
and business pages are outside this matrix; business surface visuals belong to P4+.

The initial eight images are explicitly authorized by the P3.3.3 task. Generate
only with the pinned update guard, then pass two consecutive `test:visual`
comparisons without updating. Maintainers review these eight PNGs for backdrop,
theme, controls and mobile fit. Do not regenerate on unexpected differences;
investigate the fixture/environment/product and report real product defects
without redesigning production in this phase.

### Static / Legal business surfaces (P4.3.1)

`visual/fixtures/static-pages.ts` reuses `startInsightsFixtureApp` for real
production Nitro, Nuxt SSR/hydration, Vue/Pinia and the default layout. Fresh
contexts seed `localStorage.theme`; production NavBar/Theme Store applies
the theme. Both global Resource Routing plugins probe even on Static pages with
empty history, so the fixture also seeds valid Managed/Steam diagnostics. Real
TTL logic skips probes; modes remain Auto, with no pin or plugin override.
`PublicPageBackground` stays real and reports `default`; the Static
root is computed-transparent: **Layout owns canvas; Static owns content surfaces.**
Only PageScrollDock and MobileBottomTabBar are hidden in test CSS.

`static-pages.spec.ts` owns six viewport images in its snapshot directory:
`static-about-{light,dark}-{desktop,mobile}.png` and
`static-terms-{light,dark}-mobile.png`. Desktop is 1440×900; mobile is 390×844.
No Privacy, English, Resume or full-page Legal images are included. The root is
instant-scrolled to viewport top, keeping NavBar/Footer outside this contract.

Navigation uses `load`, never `networkidle`; hydration/theme/root readiness,
successful local images, blur/neutral mouse, finite animations, fonts and two
RAF ticks settle before capture. Tests assert no horizontal overflow, effective
panel shadow/blur and top veil, and the current 0.5-second color-related root
transition. `static-transition.json` retains the computed property/duration/timing;
`static-requests.json` proves zero external, failed and upstream requests, including
during comparison. External traffic is recorded and aborted; no API data is faked.

The prerequisite Resume fix explicitly prerenders `/about/faolan`; a wildcard
rule advertised a payload that the build did not emit. Keep the zero-error check
for real NuxtLink prefetch rather than hiding missing-payload errors in fixtures.

Initial creation is explicitly authorized only for these six images. In the
pinned Linux/Node 24 container above, after `npm ci` and `npm run build`, run:

```sh
npm run test:visual:update -- -- static-pages.spec.ts
npm run test:visual
npm run test:visual
```

The extra separator keeps the file filter out of Playwright's optional update
mode argument, protects the existing twelve images and still runs the guard.
Both comparisons must report **19 passed**; Functional Browser remains **64**.
The six new images require maintainer review for natural pattern/translucency,
top veil, desktop columns/mobile stack, loaded avatar/actions and Legal wrapping.
P4.3.2 starts only after that approval and must pass the accepted images without
updating them. P4.3.1 changes no production source or style debt.

### Updates runtime and appearance (P4.4.1)

`tests/browser/fixtures/updates.ts` serves both `regression/updates.spec.ts` and
`visual/updates-page.spec.ts`. It reuses `startInsightsFixtureApp` for production
Nitro and the real SSR `useAsyncData` → `/api/v2/nav/updates?lang=zh` → payload
hydration path. Nine fixed notices (7×2026 + 2×2025) naturally produce the initial
six entries, one latest tag, Load more and a collapsed older year. Every fresh
scenario records exactly one upstream Updates call, no browser refetch, no other
business API, external traffic, failed resources or browser errors.

One functional case owns keyboard focus appearance, 6→7 loading, collapse to 0,
re-expansion retaining 7 and older-year expansion to 9. Its accessible/domain
locators avoid the generic selectors scheduled for P4.4.2. The four viewport
goldens are `updates-{light,dark}-{desktop,mobile}.png` under
`visual/__snapshots__/updates-page.spec.ts/`, at 1440×900 and 390×844. Real Theme
Store seeds and fresh Managed/Steam diagnostics keep plugins active and quiet.
Real PublicPageBackground stays `default`; production reduced motion settles
Updates animations. This spec also supplies Chromium's native
`--force-prefers-reduced-motion`: isolated SVG image documents copy native
settings rather than the page's CDP media override. The divider's own media
query disables its animation; no product CSS or shared runner is changed.
Only unrelated fixed tools are hidden. Navigation uses load
and semantic hydration/image/animation/font/RAF readiness, never networkidle or
fixed sleeps; viewport capture starts at `.updates-page`, with overflow checks.

Initial creation is authorized only for the four Updates images. After install
and build in the pinned environment, run:

```sh
npm run test:visual:update -- -- updates-page.spec.ts
npm run test:visual
npm run test:visual
```

Both comparisons must report **23 passed**; Functional Browser is **65**.
Existing eighteen PNGs stay byte-identical, giving twenty-two total. Maintainers
must review the four new images for canvas, summary/divider, timeline/latest
hierarchy, typography/wrapping and mobile containment before P4.4.2. That later
selector migration must pass these accepted images without updates. P4.4.1
changes neither production source nor style debt.

### Error Experience runtime and appearance (P4.5.1)

`tests/browser/fixtures/error-experience.ts` reuses `startInsightsFixtureApp`
for real production Nitro and `/__gofurry_error_contract_missing__`: HTTP 404,
SSR Chinese error content, `app/error.vue`, error layout and real Nuxt hydration.
It seeds localStorage theme for NavBar's Theme Store and fresh Managed/Steam TTL
diagnostics. The error owns its canvas; no DOM, artwork or product CSS is faked.
Hydrated readiness requires exactly one active, ready bundled Light/Dark AVIF
and `.error-page.is-enhanced.has-entered`, not a timer delay.

The two `regression/error-experience.spec.ts` cases use normal motion. The first
reads staged hidden actions and focuses Home in one browser turn, proving
immediate opacity 1, transform none and focus indication before the 2550ms delay
expires. The other uses a real `javaScriptEnabled: false` context and checks all
five steps and the loaded active Light artwork are visible through noscript;
it does not invent no-JS button navigation. Playwright owns both contexts.

Every scenario asserts zero upstream business/external requests and unexpected
resource/browser errors. The shared collector remains unchanged: the exact main
document 404 console diagnostic is separately recorded and required once; no-JS
`csp` script blocks must refer to local module preloads advertised in that SSR
document. All other HTTP/resource/application/hydration errors still fail.

`visual/error-experience.spec.ts` captures only `.error-page`, excluding NavBar,
at 1440×900 / 390×844 in Light/Dark Chinese. Real reduced motion, complete images,
blur, neutral mouse, finite animations, fonts and two RAFs settle capture with
overflow checks, no networkidle, fixed sleeps, masks or tolerance changes.
Initial creation is authorized only for these four Error images in the pinned
Linux/Node 24 environment, after install/build:

```sh
npm run test:visual:update -- -- error-experience.spec.ts
npm run test:visual
npm run test:visual
```

Both comparisons must report **27 passed**; Functional Browser is **67**.
The existing 22 PNGs remain byte-identical, for 26 total. Maintainers must review
the four `error-404-{light,dark}-{desktop,mobile}.png` images for loaded artwork,
mask/composition, text hierarchy, dark contrast and mobile stacked buttons before
P4.5.2. P4.5.1 changes no production source or style debt.

### PageScrollDock runtime and appearance (P4.5.2)

`tests/browser/fixtures/page-scroll-dock.ts` reuses `startInsightsFixtureApp`
with real production `/terms`, Nuxt hydration and the default document scroller.
The actual page must naturally exceed the 320px render threshold; do not inject
height or a custom scroll host. Fresh Light theme and Managed/Steam diagnostics
use real Theme Store/TTL paths, with zero upstream/external requests, failed
resources or browser errors. Playwright owns fresh contexts and route disposal.

One `regression/page-scroll-dock.spec.ts` case owns Desktop mount at 1440×900,
the rendered-but-inactive 0% state and accessible labels, real instant scrolling
to 50%, visible opacity 0.78/pointer events, a real click smoothly stepping back
by 25% of maxScroll, the resulting 25% label, then unmount at 390×844. Polling
uses document position and product listener/RAF output, with one pixel of scroll
rounding. No internal CSS property, custom-scroller or override-prop contract is
introduced; there is no fixed sleep or networkidle.

`visual/page-scroll-dock.spec.ts` adds exactly two Light Desktop images at 50%:
`page-scroll-dock-light-50.png` and `page-scroll-dock-light-50-hover.png`. It checks
the real 46×46 button, progress/accessibility, opacity, pointer events and shadow;
real `hover()` additionally checks opacity 0.93 and saturation. Both Page captures
expand the actual button box by 20px per side, clamped to the viewport, retaining
shadow and real canvas. Blur/neutral mouse or hover, computed state, fonts and
two RAFs settle capture without fake classes, masks or tolerance changes.

Initial creation is authorized only for these two images in the pinned
Linux/Node 24 environment, after install/build:

```sh
npm run test:visual:update -- -- page-scroll-dock.spec.ts
npm run test:visual
npm run test:visual
```

Both comparisons must report **29 passed**; Functional Browser is **68**.
Existing 26 PNGs remain byte-identical, giving 28 total. Maintainers must review
ring/track, inner highlight/core/border, percentage text, normal/hover opacity
and shadow containment before P4.5.3. Production and all style debt stay unchanged.

### Footer and App Shell (P4.6.1)

`visual/fixtures/footer-shell.ts` independently reuses `startInsightsFixtureApp`
for real production `/terms`, SSR/hydration, Theme Store and default
PublicPageBackground. Fresh routing diagnostics keep real plugins on the quiet
TTL path. Each Light/Dark Desktop scenario at 1440×900 requires zero upstream,
external, failed-resource and browser errors; only unrelated fixed tools hide.

The two `footer-shell.spec.ts` Page captures scroll the whole Footer into view
and derive their right edge from the gap between its second and third columns.
The meta/currentYear column must lie outside the clip; time/text is never mocked
or masked. Computed assertions cover all four heading/icon styles, themed
meta/link colors/transitions, four accessible-name social hover RGBs and the exact
App Shell color-transition properties, duration and easing. Real hover settles
through polling and all icons return to neutral before capture. Image/font/finite
animation/two-RAF readiness uses no fixed sleeps or networkidle.

Create only `footer-light-desktop.png` and `footer-dark-desktop.png` in the pinned
Linux/Node 24 environment after install/build:

```sh
npm run test:visual:update -- -- footer-shell.spec.ts
npm run test:visual
npm run test:visual
```

Both comparisons must report **31 passed**; Functional Browser stays **68**.
The existing 28 PNGs remain byte-identical, giving 30 total. Maintainers review
both Footer images for canvas, border, typography, icons, spacing and exclusion
of dynamic meta before P4.6.2. No production source or style debt changes here.

### Nav Shell runtime and appearance (P5.1.1)

`tests/browser/fixtures/nav-shell.ts` serves both the Functional and Visual specs,
reusing `startInsightsFixtureApp` with worker-owned production Nitro/local API
and fresh test-owned contexts, diagnostics, request/error evidence and state.
`/terms` hosts ordinary Nav and Mobile Menu with zero upstream/external requests
or errors before opening Mode. `/` supplies one SSR Home response with empty
groups/spotlights and null Hero keys; real Hero frames, Mobile reveal and Desktop
prewarm remain active.
Theme uses only localStorage seeds and real NavBar/Theme Store initialization.

The initial plan's fully quiet Home assumption conflicts with production.
Explicitly approved exceptions are one `/nav/home/saying?lang=zh` response with
null saying and the exact weather iframe, fulfilled locally without external
network access. Both are counted separately; all other upstream/external calls
fail. Mobile Home reuses the existing narrow `assertHeroHydration` evidence
(one mismatch, narrow viewport, absent SSR/present client Footer, retained SSR
Hero node and no other error), without modifying Hero tests or the generic guard.
Raw errors remain in the attachment and any later mismatch fails.
The real Mode modal also mounts Background Editor eagerly: a further approved
exception allows exactly one empty `/nav/appearance/patterns` catalog only after
the Mobile Functional test clicks BottomTab Mode. Other Terms scenarios allow none.

Two Functional cases own Desktop Theme→EN routing to `/en/terms` and persistent
Dark state; Mobile menu theme retention, route-driven close, English active state,
BottomTab hidden-at-top/visible-above-72 behavior, Mode modal active/Cancel and
the mounted-but-hidden 640px breakpoint. Real scrolling polls product listeners
and its 160ms sync path; no internal state mutation or fixed sleep is used.

Eight Page clips under `visual/__snapshots__/nav-shell.spec.ts/` cover
`nav-shell-standard-{light,dark}-desktop.png`,
`nav-shell-overlay-{light,dark}-desktop.png`,
`nav-shell-mobile-menu-{light,dark}.png` and
`nav-shell-bottom-tabs-{light,dark}-mobile.png` (1440×900 / 390×844, zh-CN).
Desktop clips extend real Nav bounds to include shadow/canvas while excluding
SearchBox/QuickAccess; menu clips include the absolute panel; BottomTab clips
expand its bounds by 12px and use genuine Home active state, never hover.
Target-scoped finite transitions, loaded assets, fonts and two RAFs precede
geometry/quiet checks. No global Home animation wait, mask or tolerance is added.

Only these eight initial images are authorized, in the existing pinned environment:

```sh
npm run test:visual:update -- -- nav-shell.spec.ts
npm run test:visual
npm run test:visual
```

Both comparisons must report **39 passed**, Functional Browser **70**. The original
30 PNGs remain byte-identical, for **38** total. Maintainers must review all eight
new images before P5.1.2, which must pass them unchanged. Production, style debt,
existing P4/Hero tests and legacy `visual:guard` remain untouched.

### Nav Home Header (P5.2.1)

`fixtures/nav-home-header.ts` shares a real production Home between Functional
and Visual. The worker owns Nitro/local API; every test gets fresh request/gate
state, storage, context and routes. Fixed desktop/mobile Hero keys still render
through production HeroBackground/Frame and managed assets. Only their exact
URLs and seeded favicon URLs are fulfilled locally; one exact favicon abort
exercises the real fallback. The explicit Add action additionally permits its
known `example.com` favicon. No CDN/favicon-origin wildcard is allowed.

Home is requested exactly once on SSR, with empty groups/spotlights, a non-null
saying and deterministic Hero. Hydration cannot refetch Home. Only real Bing
`wolf`/`noresult` suggestion requests may follow, through the production debounce
and API chain. A response gate exposes loading and releases in failure cleanup.
The production exact Tianqi iframe is isolated locally; external traffic and
unexpected failures/errors must remain zero. Bing popup navigation is allowed
only at the exact selected-suggestion URL after the real Enter action.

Mobile Home reuses the existing `assertHeroHydration` strict Footer mismatch
contract, with an observer recording the earliest real SSR Hero. Raw errors stay
attached; only that verified initial mismatch is consumed. Every later error
still fails, and Desktop keeps zero browser errors.

Two regressions protect Search typography/theme parity, debounce/loading/empty,
keyboard/popup and reveal lock; plus QuickAccess slots/favicon fallback and
Modal validation/add/delete/localStorage. The audited chip transition is
`background, box-shadow, color` at 500ms each: existing unlayered Less overrides
the plan's `transition-all` assumption. Accepted computed appearance is authority.

Four Page clips cover Desktop Header, Desktop/Mobile focused suggestions and
Desktop Quick Sites validation. Geometry comes from live Nav/Search/Modal bounds;
intentional input focus stays intact. Target images, finite transitions, fonts
and two RAFs settle before capture; infinite unrelated Home motion is not awaited.
No Dark golden is added: Functional compares representative Light/Dark styles.
`SiteIconStrip` has no production consumer and remains a P5.2.2 dead-cleanup
candidate; its five raw debts are not live Header coverage.

Only these four initial snapshots may be created in the pinned environment:

```sh
npm run test:visual:update -- -- nav-home-header.spec.ts
npm run test:visual
npm run test:visual
```

Acceptance is **72 Functional / 43 Visual / 42 PNG**, with the existing 38 PNGs,
production and style debt unchanged. Maintainer review of all four new images
precedes P5.2.2; existing Hero/Nav Shell contracts retain their ownership.

### Nav Revealed Content (P5.3.1)

`fixtures/nav-revealed-content.ts` is independent of the Header fixture and
reuses `startInsightsFixtureApp()`. Its worker owns production Nitro/local API;
each test owns context/routes and fresh scenario data, storage and request
evidence. The responsive case first resizes a hydrated Desktop, then closes its
documents and opens a fresh NSFW document/storage scenario before navigation.
No Mobile SSR exception or internal component mutation is involved.

Real wheel input reveals Cards/Popovers, TransitionBar, ToolDock and Spotlight.
The clock fixes only current time to `2026-09-18T12:40:00+08:00` with
`Asia/Shanghai`; browser timers still run normally. Theme/mode/recent storage
and fresh routing diagnostics initialize production consumers. Saved QuickAccess
is disabled to keep the P5.2 favicon boundary out of this content contract.
Managed Hero/site images remain real, fulfilled only at exact seeded URLs.

Every scenario proves one SSR Home request and no hydration refetch. Non-null
saying avoids a secondary saying request. The exact Tianqi iframe is fulfilled
locally. Directory GET is allowed once only after Search; a view POST and exact
popup URL are authorized only by the real site action. ToolDock clicks must not
POST views. Unexpected API/external traffic, request failures and browser errors
fail; raw network/error evidence is attached even when opening fails. Ping refresh
is not allowed: scenarios must finish before the real 60-second interval.

Four regressions cover Cards' computed typography, quote focus/blur, Group hover,
top SitePopover geometry and ping, ToolDock search/empty/Escape/cache/popup,
Spotlight motion/wrap/visited/POST, 1/2/3/4 responsive panels and cross-surface
SFW/NSFW. Audit correction: the real description is **12px / 16.2px**, because
unlayered `line-height: 1.35` overrides `text-xs`; the plan's 16px is not current
appearance. No product style is changed to satisfy a static utility inference.

Eight Light/Dark Page clips cover Core (1000px), top SitePopover (1000px),
ToolDock Search (1440px) and Spotlight (960px). Clips come from target geometry;
intentional focus/hover remains real. Visible images, target-scoped finite
animations, fonts and two RAFs settle before capture. No global animation wait,
sleep, network-idle, mask, tolerance or full-page snapshot is used.

Only the eight new baselines may be created in the existing pinned environment:

```sh
npm run test:visual:update -- -- nav-revealed-content.spec.ts
npm run test:visual
npm run test:visual
```

Acceptance is **76 Functional / 51 Visual / 50 PNG**, with two consecutive full
visual comparisons and existing 42 PNGs, production and style debt unchanged.
Maintainer review of all eight new images precedes P5.3.2. `.nav-content-loading`
stays a P5.3.4 final-sweep candidate; `.nav-tool-button--search` is live dynamic
output. Site Groups and games-page coupling remain P5.4-owned.

### Games Home missing states and closure (P6.1.3)

`fixtures/games-home.ts` preserves the P6.1.1 core dataset and default clock.
The closure specs add ten regressions: News multiple/locale/carousel/resize/popup,
empty/single states, populated Reviews in both locales, and four group-layout
cases covering Light zh / Dark en at 390, 639, 640, 768, 1023, 1024 and 1440px.
Each layout case protects four groups, 17 records, 8/8/1/wrap pagination,
card/rating bounds, spacer height and local overflow clipping. These take over
the retired `game-group-layout-smoke.mjs` and `game:groups:smoke`; its diagnostic
success screenshots are not golden baselines. Other legacy runners stay intact.

Only the new specs enable a worker clock: a child-process-only Node preload
fixes `Date.now()` and UTC through existing runtime overrides; the browser uses
the same fixed time. Real SSR and hydrated Reviews must show 5 minutes, 2 hours
and 3 days, without hydration errors. No timers, production modules or parent
process environment are patched. Exact asset/popup registration and raw evidence
enforce one SSR Home request, zero browser API calls and zero unexpected traffic.

Before accepting Reviews, identical pinned builds of `0894fc7` and `f4c322c`
were compared: 1,584 computed properties and four Light/Dark normal/hover
captures were identical. Six new Page clips cover News Light/Dark at Desktop
and Mobile, plus populated Reviews Light/Dark at Desktop. Target geometry,
loaded images, finite transitions, fonts and two RAFs determine readiness.
Acceptance is **98 Functional / 65 Visual / 64 PNG**, including two consecutive
pinned comparisons and all previous 58 PNGs unchanged. Maintainer review of the
six new images precedes P6.1.4; News appearance debt is deliberately retained.

### Shared GameReviewDialog (P6.2.1)

`fixtures/game-review-dialog.ts` serves eight runtime cases and eight modal
visual cases through the existing production Nitro helper. Home reads once in
SSR; Search reads categories/results once after mount; Detail reads four SSR
resources plus one browser view POST. Every test gets fresh scenario state and
an exact request ledger. The real form's anonymous review POST is locally gated
at the browser boundary, never sent to a development/production backend.
Teardown releases every gate and awaits only its own finite submit handlers;
context teardown owns registrations, without `unrouteAll(wait)`.

Runtime coverage includes all three game identities, settled draft reset,
required/lexical/range validation, inclusive 0/5, trim and decimal normalization,
disabled pending/no duplicate click, success, business rejection/user retry,
one precise 503 diagnostic, Mobile keyboard order/focus and viewport bounds.
Success retains the dialog/draft without refreshing host data. Search's stable
slide omits `aria-hidden`; use `:not([aria-hidden="true"])`. Raw console evidence
remains available; there is no hydration allowance or broad network suppression.

Eight clipped Home-consumer goldens cover default Light/Dark Desktop/Mobile,
validation Light Desktop/Dark Mobile, pending Dark Desktop and success Light
Desktop. Only unrelated `#__nuxt` content is hidden after opening the real body
Teleport; the body uses the production page-background token. Real geometry
expands the panel clip by 48px, protecting shadow/backdrop without owning Home
pixels. Focus, finite motion, fonts and two RAFs determine readiness. Pending
remains gated through capture. P6.2.1 captured the original Light panel in both
themes; P6.2.2 corrects Dark after explicit maintainer feedback.

Acceptance is **106 Functional / 73 Visual / 72 PNG**, two full pinned compares,
and original 64 PNG hashes, production and style debt unchanged for P6.2.1.
P6.2.2 first proves an equivalent style-owner migration against all eight old
Review images, then revises only the four Dark images in the pinned environment.
Four Light and 64 earlier images remain byte-identical. Computed expectations
split only the approved palette by theme, supplementing Dark placeholder,
focus and hover checks; geometry/typography and eight runtime cases stay intact.
Full acceptance remains 106 Functional / 73 Visual (two compares) / 72 PNG.
The revised Dark images require renewed maintainer review. Missing close translation,
accessibility enhancements and potential pending-close response races remain
separate work; no negative assertions freeze those deficiencies as requirements.

## Verification

Fresh `npm ci` runs `nuxt prepare` through `postinstall`, generating `.nuxt`
types/config before either test project runs. Verification must not rely on a
previous dev server or build having generated `.nuxt/tsconfig.json`.

Run `npm ci`, `npm run lint`, `npm run stylelint`, `npm run style:policy:test`,
`npm run style:policy`, `npm run test:unit`, `npm run test:nuxt`, `npm test`,
`npm run typecheck`, `npm run insights:semantics`, `npm run seo:recovery:test`,
and `npm run build`, then the Chromium installation and three browser commands
above. On Linux CI use `npx playwright install --with-deps chromium`.

The existing Nav Web CI job has separate **Unit tests** and **Nuxt tests** steps
before typecheck/contract guards/build. After Build succeeds it installs Chromium
and runs **Browser tests** (`npm run test:browser`). A successful local run of the
same commands is local evidence, not proof of a remote Actions run. The separate
`nav-web-visual` job adds its own container install/build/Visual gate after this.
Run those three commands in the pinned container as well. After an authorized
push verify that both `nav-web` and `nav-web-visual` actually ran and passed; a
skipped Visual job is not acceptance. When instructed not to push, report remote
gate acceptance as unverified.

Ordinary test-only migrations with unchanged production/runtime and smoke
behavior need no manual UI review once guards pass. Creating or changing golden
baselines additionally requires the focused maintainer visual review above.

Configuration references: [Nuxt testing](https://nuxt.com/docs/4.x/getting-started/testing)
and [Vitest projects](https://vitest.dev/guide/projects).
