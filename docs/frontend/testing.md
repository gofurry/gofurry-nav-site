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
| Playwright Visual | `tests/browser/visual/*.spec.ts`, `npm run test:visual` | Pinned sentinel, four Shared Primitive Foundation and eight real Preferences Modal locator baselines; business surfaces remain separate |
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
