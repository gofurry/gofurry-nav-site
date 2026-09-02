# V3 P1 Insights MVP Acceptance

## Acceptance header

- Final status: **PASS**
- Initial audited commit: `2498e36bda74b101f02b08812d04e2010c128ad7`
- Accepted implementation commit: `7aec0a2900c0ef3b64aaa6170bcd980ffd6c7093`
- Acceptance execution date: `2026-09-01`
- Acceptance environment: repository-configured local PostgreSQL/Redis services, local Nav Backend on `9999`, local Game Backend on `9998`, and the production Nuxt build on `3000`
- Real-data entity cohort: Nav Site `101`, Game `100`
- Isolated browser semantic cohort: Nav Sites `41` and `42`, Games `82` and `83`
- Open P1 blockers: **0**

The implementation consumes the accepted P1-A Public Insights contracts without changing a backend route, database schema, Metric/Detector/Fact contract, cache, worker, or migration. The semantic cohort used browser-local API interception only; it did not mutate the local databases.

## G1 — Site Entity Insights

Status: **PASS**

Evidence:

- `/site/:id` renders an early Site Insights panel after the high-level signal cards and before observation detail.
- Site detail and Site Insights are independent concurrent SSR requests. A failed Insights request renders an unavailable panel without breaking the base Site detail.
- IPv6, TLS 1.3, and security.txt retain the exact seven Public states: `supported`, `unsupported`, `stale`, `not_probed`, `unavailable`, `unknown`, and `not_applicable`.
- A capability omitted by the API is rendered as unavailable rather than fabricated as `unknown` or `unsupported`.
- Each capability renders the API-supplied ecosystem value, coverage, and `as_of`; the isolated cohort verified all three comparisons use the same supplied Fact date.
- Entity changes use a bounded semantic timeline and the panel links back to `/insights/sites`.
- `/site/:id/:domain` and `/en/site/:id/:domain` do not render Site-level Insights.
- Real SSR passed for `/site/101` and `/en/site/101`.

Resolved finding:

- `P1C-001` — **BLOCKER, resolved**: the existing Site parent page classified child routes from raw URL segment count. The locale prefix made `/en/site/:id` look like a target child route, suppressing the English Site detail and Insights panel. Route classification now uses the actual `domain` parameter; zh/en entity SSR and target-route exclusion both pass.

## G2 — Game Entity Insights

Status: **PASS**

Evidence:

- The Insights tab is immediately after Introduction in the existing Game tab system.
- The nullable Game Insights summary is loaded through an independent SSR request and retained in the Nuxt payload. Real SSR passed for `/games/100` and `/en/games/100`.
- Summary rendering covers current players, 30-day peak, CN price, Windows/Linux support, free/paid state, release state, and source dates without converting nulls to zero.
- The Insights view is outside `BlurWrapper`; existing Intro, Gallery, and News NSFW boundaries remain unchanged.
- Player and price histories mount only on first entry into the Insights tab. The mounted tab is retained with `v-show`, so its page-lifecycle range cache survives tab changes.
- A shared `30d` / `90d` / `all` selector loads Player and CN Price independently. Successful ranges are cached; the isolated `30d → 90d → 30d` flow verified one request per 30-day endpoint.
- Loading may retain the previous chart; Player and Price errors have independent unavailable/retry states. The isolated cohort verified each history can render while the other fails.
- The Game entity timeline links back to `/insights/games` and works when `detail` is null.

## G3 — Historical semantics

Status: **PASS**

Evidence:

- Player charts use only returned dates and daily `max` as the plotted series; `avg` appears only in the tooltip. There is no zero-fill, forward-fill, interpolation, or fabricated date generation.
- A real player value of `0` remains visible as `0`; a null current player or failed/no-sample history renders unavailable instead of zero.
- Price histories consume CN only and do not add a region selector, fallback, conversion, historical-low, or prediction behavior.
- `free` is rendered as Free with null amount; `priced` with `final_amount=0` remains a priced `¥0.00`; `unknown` and `unpriced` become chart gaps/unavailable.
- Price lines use `connectNulls: false`, preserving missing and unavailable dates.
- Day-precision changes render the supplied date only. Exact `occurred_at` values render localized date plus time; no midnight is synthesized.
- Unit-level semantic checks and browser-level isolated flows cover priced zero, free, unknown/unpriced, genuine player zero, unavailable histories, same-day Site context, and event precision.

## G4 — Ecosystem to entity product loop

Status: **PASS**

Evidence:

- P1-B Recent Changes continues to route Site events to `/site/:id` and Game events to `/games/:id`.
- Shared event labels and time formatting were extracted only into `app/utils/insightChanges.ts` and are used by both the ecosystem feed and entity timeline.
- Site and Game entity experiences both provide a direct return link to their ecosystem Insights page.
- Browser smoke exercised `/insights → Site change → Site Insights` and `/insights → Game change → Game Insights → history → timeline`.

## G5 — SSR, i18n, responsive, theme, and failure handling

Status: **PASS**

Evidence:

- zh/en labels cover the complete entity experience and exact Public capability state wording.
- Real-data SSR verified both locales for Site and Game entities.
- Isolated failures verified Site Insights, Game summary, Player history, and Price history do not break their base page or sibling data section.
- The visual guard passed. New Game Entity and Site Entity scenarios produced light/dark screenshots at `1440x900` and `390x844` with required structures present and no horizontal overflow.
- Manual inspection of the eight entity screenshots confirmed readable card, chart, timeline, and mobile stacking behavior. Existing visual-guard network-idle, optional-empty-data, and unrelated Game-home screenshot warnings remained non-failing; no entity scenario failed.

## G6 — Regression and CI

Status: **PASS**

Checks executed:

- `npm ci` — PASS.
- `npm run typecheck` — PASS after the final install and final source changes.
- `npm run build` — PASS production client, SSR server, prerender, and Nitro artifact build. Existing source-map, dependency deprecation, and large-chunk warnings remain non-failing.
- `npm run insights:semantics` — PASS for Public price-state and event-precision semantics.
- `npm run insights:smoke -- --entity-site-id 101 --entity-game-id 100` — PASS for P1-B SSR, Workshop 404, URL state, entity SSR, Site target boundaries, exact states, same-day comparison, player zero, priced zero, range caching, timelines, and partial failures.
- `npm run visual:guard -- --entity-site-id 101 --entity-game-id 100` — PASS for source-debt checks and the existing plus new light/dark desktop/mobile matrix.
- `go run ./check-production-policy` from `tools` — PASS for the six active Go modules and repository production boundaries.
- `git diff --check` — PASS before the implementation commit.

Repository CI change detection selects only the `nav-web` job for the implementation paths; that job's `npm ci`, typecheck, and build sequence was reproduced locally. Go/sqlc/Goose/PostgreSQL integration jobs are not selected because no Go, SQL, generated sqlc, schema, migration, or workflow file changed. The repository production policy was additionally run explicitly.

P1-B regressions:

- `/insights`, `/insights/sites`, `/insights/games`, and their English variants SSR successfully.
- All tested `/workshop*` and `/en/workshop*` routes remain ordinary `404` responses without redirects.
- This implementation does not delete or change a Steam asset, Game detail media/resource path, or backend Game capability; Steam-related Game behavior remains outside the removed GoFurry Workshop product route.

## Migration and backend contract

- Migration added: **No**.
- Backend code changed: **No**.
- P1-A Public Insights contract changed: **No**.
- New table/materialized view/cache/worker/service: **No**.

## Open blockers

None.

## Deferred P2 items

- Ranking and Compare.
- Deep dimensions and HTTP/Ping ecosystem analytics.
- Regional price intelligence, historical low, and player ranking.
- Language, developer, and publisher analytics.
- Change Explorer, Data Explorer, export, and composite scoring.
- Any other P2 capability.

## Final decision

All G1–G6 gates pass and Open P1 Blockers = 0.

`V3-P1-C Entity Insights & MVP Acceptance: PASS.`

`V3-P1 Insights MVP: PASS.`
