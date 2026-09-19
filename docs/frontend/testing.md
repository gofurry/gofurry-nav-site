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
| Legacy Browser Smoke | Existing `*:smoke` scripts, after `npm run build` | Real browser/production SSR, hydration, interactions and resource loading; runner migration remains P3.2 |
| External Acceptance | Explicitly authorized development/provider checks | Real services, separate from deterministic fixtures and normal CI |

`npm test` runs both Vitest projects once. `vitest.config.ts` uses `projects`, not
the deprecated workspace model. Nuxt configuration is loaded by the test-utils
project; the test module is **not** added to production `nuxt.config.ts`.
Dependencies are pinned in package/lockfiles. No coverage provider, Testing Library
or Playwright Test runner is introduced. Existing `playwright` remains for smoke.
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
Keep `assets:routing-smoke`, `game:tags:smoke`, `game:detail:smoke` and the other
existing smoke commands until P3.2. Current screenshots and `visual:guard` remain;
this phase does not introduce visual baselines or external acceptance runs.

## Verification

Run `npm ci`, `npm run lint`, `npm run stylelint`, `npm run style:policy:test`,
`npm run style:policy`, `npm run test:unit`, `npm run test:nuxt`, `npm test`,
`npm run typecheck`, `npm run insights:semantics`, `npm run seo:recovery:test`,
and `npm run build`.

The existing Nav Web CI job has separate **Unit tests** and **Nuxt tests** steps
before typecheck/contract guards/build. A successful local run of the same
commands is local evidence, not proof of a remote Actions run. Verify remote
results after an authorized push.

No manual UI review is required for a test-only change with unchanged production
Vue/styles/runtime and unchanged browser smoke behavior, once the migrated tests
and existing guards pass.

Configuration references: [Nuxt testing](https://nuxt.com/docs/4.x/getting-started/testing)
and [Vitest projects](https://vitest.dev/guide/projects).
