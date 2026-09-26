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
