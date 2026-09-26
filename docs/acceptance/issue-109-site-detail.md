# Issue #109 — Site Detail phase ledger

## P1: Runtime and information architecture

Scope: the [Site Detail runtime contract](../../contracts/nav-web-frontend.md#site-detail-runtime-and-route-ownership-109-p1).
This is P1 evidence only; #109 and P2–P8 are not complete. #108 is unchanged.

### Audit and resolved discrepancies

- The route's `hasTargetQuery` / `showInsights` condition hid Site Insights on
  Target URLs, and `insights-entity` asserted that absence. Both are replaced by
  Site-ID-only Insights ownership; the existing panel appearance remains.
- The old three-key array was the only panel catalog. The shared registry now
  contains all seven capabilities, categories, order and bilingual label keys.
  The legacy three-item preview is explicitly a subset, not the P3/P6 catalog.
- `siteRoutes.ts` previously discarded workspace query when selecting a Target.
  It now delegates to the single route-state parser/normalizer/builder; primary
  tab transitions clear previous secondary state, Target transitions preserve it.
- `buildSiteDetailSeo` copy and Entity-only canonical/hreflang/sitemap behavior
  needed no production rewrite. SEO Browser coverage now exercises all UI keys.
- A populated HTTP detail mounted `SitePerformance`, whose mount hook fetched
  Ping observations immediately. This contradicted the required initial budget;
  the old empty-HTTP fixture hid that request. The #109 fixture now supplies HTTP
  evidence and distinct Target status values. Ping history is fetched only by
  existing sample controls, with pending dedupe/cache and stale-result guards;
  no new control, chart, endpoint or history feature is introduced.

### Executable ownership

- Site Detail requests use Site ID, selected Domain and language. Site Insights
  uses only Site ID; View remains a Site-level mounted side effect. A counted
  View value survives Target Detail replacement.
- Seventeen independent `site-detail-contract.spec.ts` cases use the shared
  `insights-runtime` Nitro/upstream transport and exact request/error ledgers.
  They cover zh/en SSR, real Target clicks from all four workspace states,
  invalid UI fallback, authoritative Detail failure, optional Insights failure,
  success-empty, View failure and explicit legacy Ping-history interaction.
- Successful hydrated Target switches retain Site capability semantics and
  workspace query with exactly one additional Detail request, zero additional
  Insights requests and zero additional View POSTs. The HTTP evidence changes
  from fixture status 200 to 201. The normal initial budget is three requests,
  including when the performance component is present.
- Client Detail 503 reaches the page error. Existing transport retry behavior
  remains explicit: SSR GET failures have two upstream attempts; client failures
  have two browser attempts, each with two Nitro proxy upstream attempts.
- The existing `insights-entity` bridge and 1440/390 × Light/Dark runtime cases
  remain. SEO checks require query-free localized Entity canonical/hreflang and
  sitemap inventory. Unit tests cover route vocabularies/cleanup/encoding and
  complete registry metadata. No runner or browser-error allowance was added.

### Local verification (2026-09-26)

Environment: Windows, Node 24.15.0, pnpm 12.6.0, frozen dependency installation.
Chromium installed through the existing Playwright command; retries remain zero.

Passed: lint, stylelint, style-policy tests (74), exact style-policy baseline,
Unit (83), Nuxt (6), combined Vitest (89), typecheck, Insights semantics, SEO
recovery guard, production build and focused Site Detail Browser (17).
The final full `pnpm run test:browser --workers=1` run passed all **423** cases
in 12.3 minutes, with no failures, skips or retries. The earlier interrupted
run is not acceptance evidence. Local P1 exit criteria are satisfied.

Style debt, ESLint suppressions, Visual specs/configuration and the 118 accepted
PNGs are unchanged. No style-budget updater or snapshot generation was run.
Pinned Linux Visual comparison and remote CI are unverified for this change;
no earlier CI run is used as acceptance evidence.

### Maintainer handoff

Confirm on a real Site with multiple Targets: normal and Target URLs retain
Insights; selecting another Target changes its evidence without routing failure;
canonical remains the localized Entity URL. Existing Ping history now requires
one sample-button interaction. No new visual review is requested by P1.

P2 owns Shell/Hero/Health Strip/Target Context; P3 Overview; P4 Observation;
P5 Security; P6 Insights workspace; P7 appearance/debt cleanup; P8 Visual/closure.
After the lightweight maintainer behavior confirmation, P2 can use the P1
owners. Its implementation remains outside this phase; remote CI/Visual status
above must not be represented as a pass.

## P2: Shell and Target context

Scope: [P2 shell contract](../../contracts/nav-web-frontend.md#site-detail-shell-and-target-context-109-p2).
P1 remains the mandatory runtime baseline. This implementation stays on `dev`;
P3–P8, #108 and final Site Detail Visual golden creation remain out of scope.

### Implementation and ownership

- Rewrote the existing Hero in place as identity-only. Removed its hover domain
  popover, edge/keyword/technical content and the old large signal-card path.
- Added six Target-owned health fields, four router primary tabs, a 3:1 desktop
  workspace/context grid, and compact inline context on tablet/mobile. Only the
  tabs and desktop context stick; the public header/Hero/health remain normal flow.
  Mobile health is 2×3; long descriptions are clamped on mobile.
- A single responsive Context and accessible listbox selector use the shared
  `siteTargetPresentation.ts` projection. It preserves missing/zero/false, refuses
  unrelated Target data, never substitutes Site aggregate health, and retains
  infrastructure hint confidence and existing Target relations.
- Checked backend missing-summary constructors: nil Target arrays and Go zero
  timestamps are legitimate responses. Adapter and tests handle them as missing
  evidence rather than throwing or displaying year 1.
- Tab changes use the P1 query owner, restore through reload/back/forward and
  clear foreign workspace state. Target selection preserves valid secondary
  state. Pending retains the shell and explicitly labels last-resolved evidence;
  held-response coverage proves a late A response cannot overwrite newer B.
- Kept one Site Insights slice mounted. Existing preview appears in Overview and
  Ecosystem; legacy observation panels and minimal transport evidence are thin
  transitional workspace content, not P3–P6 implementations. Public labels retain
  the accepted “生态观测 / Ecosystem” naming while the route value remains `insights`.

### Executable verification

- P1's 17 Browser cases retain SSR, authoritative/optional failure boundaries,
  SEO/request budgets and Target-switch assertions. They now click the selector;
  the existing history sample test opens Observation explicitly.
- Twelve additional `site-detail-shell.spec.ts` cases cover six viewport/theme
  combinations, long hostnames, route/history/reload, tab and selector keyboard,
  touch, outside dismissal/focus return, keyboard visibility in a long list,
  pending/race and missing/sticky behavior.
  The same Nitro/upstream fixture, strict error capture and zero retries apply.
- Five pure adapter cases cover Site/Target ownership, null/zero/false, stale
  evidence, deterministic timestamps, relations/hints and actual missing-summary
  shape. A policy test confines the new token declarations to exact roots.

### Style and Visual boundaries

Only Hero/popover/page-root debt was removed: Tailwind appearance **632 → 572**,
raw visual values **463 → 412**, deep selectors **33 → 25**. Arbitrary appearance
(5), `!important` (5), legacy dark entries (0) and every other per-file budget
remain unchanged. The updater ran only after the policy reported these five
stale entries with zero regressions. ESLint pruning removed only the page's two
obsolete `no-explicit-any` suppressions. No budget transfer/increase occurred.

Existing Visual specs/configuration and accepted PNGs are protected independently
from the local manual-review screenshots. No snapshot update or new runner is
part of P2. Remaining legacy panel debt retains its later-phase owners.

### Local and remote acceptance

On 2026-09-26, Windows / Node 24.15.0 / pnpm 12.6.0 passed frozen installation,
lint, stylelint, style-policy tests (75), exact style-policy baseline, Unit (88),
Nuxt (6), combined Vitest (94), typecheck, Insights semantics, SEO recovery guard
and production build. Chromium was installed through the existing Playwright
command. The final focused Site Detail run passed 29 cases. After the last
selector fix, the complete `pnpm run test:browser --workers=1` run passed all
**435 cases in 13.0 minutes**, with no failures, skips or retries. These are the
current local acceptance results, not the earlier exploratory/failing runs.

The protected Visual file hashes and all **118 accepted PNGs** match the P1
baseline; Visual inventory remains 119 tests. No accepted snapshot was generated
or changed. Remote CI and pinned Linux Visual comparison remain **unverified**
for this local change; no push was performed and earlier remote runs are not
acceptance. Technical P2 criteria are locally verified; manual criterion 30 is
still outstanding.

### Required maintainer review before P3

Review 1440 Light/Dark, 390 Light/Dark and preferably 768 Light: identity clarity,
Hero height, six-field density, sticky tabs, 75/25 balance/sidebar width, long
hostnames, selector operation, mobile 2×3/compact context and theme consistency.
Local fixture screenshots support this review; they are not a final golden or
maintainer approval. Also confirm a real multi-Target Site's Visit URL and Target
switch behavior. P2 exit criterion 30 remains pending until the maintainer accepts
the visual direction. Do not enter P3 before that acceptance.
