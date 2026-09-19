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
to change. Teardown releases pending handlers and awaits unroute before Playwright
closes the context; the worker always closes its app. No serial-suite dependency.

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
same commands is local evidence, not proof of a remote Actions run. Verify both
browser steps after an authorized push; when instructed not to push, report that
remote gate acceptance remains unverified.

No manual UI review is required for a test-only change with unchanged production
Vue/styles/runtime and unchanged browser smoke behavior, once the migrated tests
and existing guards pass.

Configuration references: [Nuxt testing](https://nuxt.com/docs/4.x/getting-started/testing)
and [Vitest projects](https://vitest.dev/guide/projects).
