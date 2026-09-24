# #124 frontend engineering closure

Implementation base: `074f73af58c2702c54a54b3b921f26856894d3d6` (local dev).
Validation dates: 2026-09-24–25. Candidate: the P7.2 commit that introduces this record;
resolve its exact SHA with `git log --diff-filter=A --format=%H -- docs/acceptance/issue-124-frontend-engineering-closure.md`.

Status: P7.2 implementation and local validation complete. The maintainer reported
manual UI acceptance **PASS** on 2026-09-25 for `ea3d4a1`. Final release/merge-SHA
CI and the remaining Issue closure steps are **Pending**. This record does not close #124.

## Scope and immutable evidence

P7.1 migrated eight runtime runners (179 retained +181 migrated cases). P7.2
retains those 360 and adds 46 focused locale/theme/shell cases, without changing
production, style debt, dependency/lockfiles, CI/Playwright configuration or any
of the 118 accepted PNGs. Existing Visual remains 119 checks (one sentinel).

The frozen local audit covers 491 tracked production/server/Visual/policy-owner/
config/lock/debt files. Final verification compares their SHA-256 values and
checks original Browser case identities, independent of changed line numbers.
Local execution logs and hashes are diagnostics under ignored `.cache/p72`.
They are not committed credentials or a replacement for remote CI evidence.

## Legacy visual-guard retirement ledger

Source before deletion: `074f73a:apps/cn/nav-web/scripts/perf/visual-guard.mjs`.
It generated 90 reports: twenty four-cell groups plus five two-cell groups.
`D` =1440px desktop, `M` =390px mobile, `L/K` =Light/Dark. Each row below covers
all listed cells. Named files are under `apps/cn/nav-web/tests/browser`.

| Legacy route group | Cells | Current executable evidence |
| --- | --- | --- |
| `/` | DL/DK/ML/MK | **New** `nav-home-locales`: Localized Home shell and real reveal; existing Nav/Hero contracts remain. Real revealed root/overflow checks complement Nav-only Visual clips |
| `/en` | DL/DK/ML/MK | **New** smoke `nav-home-locales`: Localized Home shell and real reveal /en; retained SSR node, Theme Store, root/Header/Search, real content reveal and overflow; exact Home/saying/weather budget |
| `/games` | DL/DK/ML/MK | Visual `games-home`: Games Home theme/device; real four groups/sidebar/dock; regression `games-home` + closure own data, pagination and clipping |
| `/games/1` | DL/DK/ML/MK | Visual `game-detail`: Detail overview theme/device, deterministic valid ID; existing Detail regressions own hero/tabs/sidebar/regions/local scroll |
| Game entity Insights tab | DL/DK/ML/MK | Visual `game-detail`: Detail Insights theme/device, real tab + loaded player/price ranges; `game-detail-insights` owns caches/partial failures/timeline |
| `/games/prize` | DL/DK/ML/MK | Visual `lottery`: Lottery page theme/device, real CSR state; regression lottery owns empty/error/modal |
| Prize Activation success | ML/MK | **New** regression `lottery`: activation success mobile shell L/K, actual status/link/canvas/overflow and zero API. Existing goldens cover desktop success and mobile failure, not this pair |
| `/games/search` | DL/DK/ML/MK | Visual `games-search`: Search results theme/device; result/page/Filter contracts and geometry remain |
| `/en/games/search` | DL/DK/ML/MK | **New** regression `games-search-contract`: English Search ready shell; actual visible slide, pagination, exact computed appearance, one results/tags request, no overflow |
| `/insights` | DL/DK/ML/MK | Overview SSR/hydration/media/layout zh matrix (L); **new** Overview Dark ready shell D/M |
| `/insights/sites?metric=ipv6&range=30d` | DL/DK/ML/MK | Domain SSR matrix site (L); **new** Domain Dark ready shell site D/M, real chart/metric rail/dimension/info |
| `/insights/games?metric=free&range=30d` | DL/DK/ML/MK | Domain SSR matrix game (L); **new** Domain Dark ready shell game D/M; exact source budgets |
| `/insights/changes?domain=site&range=30d` | DL/DK/ML/MK | Changes SSR site zh matrix (L); **new** Changes Site Dark ready shell D/M, four events/filter/feed |
| Players latest_observed | DL/DK/ML/MK | Workspace zh SSR matrix (L); **new** Workspace Chinese Dark ready shell players D/M, selected option and 20 ranks |
| Prices region=CN | DL/DK/ML/MK | Workspace zh SSR matrix (L); **new** Workspace Chinese Dark ready shell prices D/M, selected region and six discounts |
| Languages | DL/DK/ML/MK | Workspace zh SSR matrix (L); **new** Workspace Chinese Dark ready shell languages D/M, twelve rows and disclosure |
| `/site/<id>` | DL/DK/ML/MK | **New** Site entity ready shell in `insights-entity`, valid deterministic 41; real Site Insights/capability/timeline, exact detail/insights/view requests |
| `/about` | DL/DK/ML/MK | Visual `static-pages`: Static About theme/viewport, nonzero roots/panels/links, transparent canvas and no overflow |
| `/en/about` | DL/DK/ML/MK | **New** smoke `static-locales`, two real panels/links/media, transparent canvas, no APIs/errors/overflow |
| `/updates` | DL/DK/ML/MK | Visual `updates-page` four cells; real SSR nine notices, initial six entries/latest/load-more, theme/overflow |
| `/en/updates` | DL/DK/ML/MK | **New** English Updates runtime in `regression/updates`; exact SSR lang=en, no hydration refetch, six-to-seven entries, no overflow |
| `/terms` | ML/MK | Visual `static-pages`: Static Terms theme/mobile; real Legal panel/section list, media/canvas/overflow |
| `/en/terms` | ML/MK | **New** smoke `static-locales`, real Legal sections and zero business traffic |
| `/privacy` | ML/MK | **New** smoke `static-locales`, real Legal sections and zero business traffic |
| `/en/privacy` | ML/MK | **New** smoke `static-locales`, real English Legal sections and zero business traffic |

Existing Insights matrices use height 1000 at the same width breakpoints; their
root/overflow/layout assertions own the same width-responsive branch rather than
a screenshot at a different width. New cells use 1440×900 /390×844. Optional legacy
empty-data warnings are replaced by deterministic ready/empty/failure contracts,
not by weakening required containers. Old `intelligence-selector`/table/info class
spellings were replaced by real `[data-workspace-option]`, ranking/discount/language
and disclosure semantics. No obsolete production class is reintroduced.

Legacy full-page success screenshots/manifests were reports, not pixel baselines.
The accepted 118 PNGs remain the authoritative bounded Visual inventory. There
are no new #108/#109 goldens; their runtime checks remain, while redesign belongs
to those Issues. Nav active states are already owned by Nav Shell/BottomTab tests.
Playwright failure trace/report/screenshot replaces the old report diagnostics.

| Retired source scan | Durable owner |
| --- | --- |
| Ten legacy dark class names and `:global(.dark` | Parser-backed `legacy-dark-entry`; P7.2 adds all-ten class/selector regression to existing detector tests |
| File-level deep allowlist | Exact rule/file debt: SiteDetailPage 8, SitePerformancePanel 25, other paths zero; existing debt tests reject growth and transfer |
| Complex Tailwind colors in 15 migrated Game/Detail/Lottery files | All-source Tailwind detector + omitted rule/file budget zero; no old file allowlist survives |
| Theme and DOM checks | Domain runtime/Visual evidence above, actual Theme Store and exact request/error accounting |

Direct `playwright`, reusable `scripts/fixtures/**`, `perf:*`, Node policy tests,
Insights/SEO Contract Guards and explicitly authorized cloud acceptance remain.
`scripts/` is not mechanically emptied; no directory move or dependency upgrade
is part of this retirement.

## Remaining debt ownership

| Rule | Files / occurrences | Owner |
| --- | --- | --- |
| Tailwind appearance | 10 /632 | #109 Site |
| Tailwind arbitrary appearance | 4 /5 | #109 Site |
| Raw visual values | 16 /463 | #109 Site 388; preserved ambient 75 |
| Important | 2 /5 | #108 shared/domain Insights |
| Deep selector | 2 /33 | #109 Site |
| Legacy dark | 0 /0 | None |

Audit correction to the P7.2 proposal: six raw literals physically in
`styles/pages/insights.less` belong to `.site-insights-capability`, hence #109,
consistent with the accepted P6.6 semantic audit. They are not #108 debt merely
because of the filename. No values or budgets change. Manifest exceptions remain
empty; exclusions are not wildcard permission to add debt.

## RUNTIME-FOOTER-01

- Owner: Nav Web default-layout / Nav Home maintainers (repository maintainer).
- Status: independently tracked runtime follow-up; remediation deferred with the
  approved P7.2 plan. GitHub issue linkage and final maintainer disposition are
  Pending before #124 closure; this is not a production fix claim.
- Surface: real fresh SSR `/` and `/en` at width <768, both themes.
- Cause: default layout's immediate route watcher initializes Home Footer false
  in SSR but true on the mobile client. Client renders exactly one Footer.
- Evidence: one initial hydration mismatch, SSR Footer absent/client count=1,
  earliest real SSR Hero node retained, all other errors/traffic zero. English
  desktop has no mismatch. Fixture preserves raw errors including the accepted
  initial entry; later errors of any kind still fail.
- Test boundary: `assertHeroHydration` defaults to `/`; only the English Home
  fixture explicitly supplies `/en`. No generic browser-errors ignore was added.
- Remove when: a separately reviewed layout/runtime change preserves SSR/client
  Footer structure and Hero identity on both localized Home routes; replace both
  narrow allowances with zero-error assertions and rerun Hero/Nav/Visual gates.
- Not allowed: broad regex suppression, skipped mobile coverage, modifying DOM
  to hide the Footer, or treating this known error as a passed zero-error case.

## Execution evidence

| Gate | Local result |
| --- | --- |
| npm ci / postinstall prepare | PASS |
| ESLint / Stylelint | PASS |
| Style policy tooling tests | 74/74 PASS |
| Style policy | PASS, exact unchanged rule/file baseline |
| Unit / Nuxt / combined npm test | 36/36, 6/6, 42/42 PASS |
| Typecheck | PASS |
| Insights semantics / SEO recovery guards | PASS |
| Windows production build / Chromium install | PASS |
| New cases, targeted | All 46 PASS; original cases not removed |
| Changed fixtures repeated twice | 52/52 PASS |
| Full Browser run 1 / run 2 | 406/406 PASS (14.1m), 406/406 PASS (14.6m); zero skipped/retries |
| Pinned Linux install/build/Visual | Fresh install/build PASS; Visual 119/119 PASS (4.0m), no snapshot update |
| Frozen source/PNG evidence | 491 hashes unchanged; 118 PNG; original 360 Browser identities retained |

Pinned validation used Linux, Node 24.15.0 and `GOFURRY_VISUAL_ENV=pinned`, with
the workflow's Playwright v1.60.0-noble image digest
`sha256:9bd26ad900bb5e0f4dee75839e957a89ae89c2b7ab1e76050e559790e946b948`.
It installed and built a fresh source copy inside the container; no Windows
`.output` was reused. Local logs include both full Browser runs and pinned checks.

Early targeted development failures were corrected in test expectations: About
has two panels, Site uses the /detail endpoint, and Search's first grid is a hidden
layout spacer. Assertions now target the real surfaces and mounted view POST.
A supplemental run started before build finished failed fixture initialization;
it was rerun successfully against the completed build. No production fix, weaker
assertion, retry or timeout increase was used.

Remote at audit time: dev `620ac7c92dd3e2c086bbb25258c6da65fb17f378`, successful
[run 35894423663](https://github.com/gofurry/gofurry-nav-site/actions/runs/35894423663).
That run predates four local implementation commits and cannot accept P7.2.

Release-preparation check on 2026-09-25: remote dev now matches the implementation
candidate `ea3d4a1402d4edebafc090f2ae65063bcd747d37`.
[Run 36029345965](https://github.com/gofurry/gofurry-nav-site/actions/runs/36029345965)
completed successfully: `nav-web` and `nav-web-visual` both actually ran and passed,
with logs confirming Browser **406/406** (11.4m) and Visual **119/119** (3.2m).
The subsequent alpha.9 PR/merge commit needs
its own affected-service gates. Skipped Go/database jobs on
this frontend-only push are not full-release validation. See the
[release guide](../releases/v3.0.0-alpha.9.md).
No production/cloud operation or Issue closure is implicit in this record.

## Maintainer acceptance — PASS; publication checks pending

On 2026-09-25 the repository maintainer explicitly reported “人工审核通过” after
the frontend acceptance checklist, for candidate
`ea3d4a1402d4edebafc090f2ae65063bcd747d37`. The checked UI items below record that
maintainer attestation; no additional automated run or production deployment is
implied. Existing golden approval is retained; no image changed in P7.2 or the
subsequent documentation-only alpha.9 preparation.

- [x] Implementation candidate `ea3d4a1`: local complete checks and remote
  nav-web / nav-web-visual actually execute and PASS in run 36029345965.
- [ ] Alpha.9 PR/merge commit: its affected-service gates execute and PASS;
  skipped jobs or the earlier candidate's green status do not replace this check.
- [x] Light/Dark and 1440/390: shell/menu/locale/BottomTab/Footer, Static/Legal,
  Updates, real 404 and ScrollDock; keyboard focus remains usable.
- [x] Preferences three tabs, Save/Cancel, Random/Fixed/Local Hero, page background,
  Resource Routing pin/probe/snapshot/fallback without reloading displayed assets.
- [x] Nav Header search/Quick Sites, real reveal, Cards/Popovers/Spotlight/ToolDock,
  Site Groups; English long-text paths included.
- [x] Games Home groups/Stats/News/Sidebar, ReviewDialog, Search drafts/results/
  filters/date/Jump/retry, Lottery and activation in an isolated development flow.
- [x] Detail tabs, NSFW/Lightbox/focus, gallery containment, 75/25 and narrow layouts.
- [x] Insights/Site excluded surfaces navigate/render/theme correctly; existing
  redesign requests remain #108/#109, not silently reassigned to #124.
- [ ] RUNTIME-FOOTER-01 has explicit maintainer follow-up disposition/Issue link;
  no unassigned blocking regression remains.
- [ ] Maintainer explicitly approves #124 closure. Implementation completion alone
  neither closes the Issue nor claims manual acceptance.
