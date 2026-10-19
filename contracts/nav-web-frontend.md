# Nav Web frontend contract

## Scope and authority

Tooling uses Node 24 and `packageManager: pnpm@12.6.0`, with this frontend's own
`pnpm-lock.yaml` and single-project `pnpm-workspace.yaml`. No repository-root
workspace or shared lock is used. Frozen installs and narrowly reviewed dependency
script permissions apply in local work, CI and Docker. Root Task is the repository
engineering entrypoint; its default tests/verify do not replace Browser/Visual.

This contract governs `apps/cn/web` under
[#124](https://github.com/gofurry/gofurry-nav-site/issues/124#issuecomment-5740012423).
MUST/MUST NOT are requirements; SHOULD permits a reasoned, documented departure;
MAY denotes an allowed choice. P0 established governance and a one-time debt
baseline; P1 added static enforcement. P2.1 formalized token ownership and
annotation; P2.2 separates visual primitives from compound product styles without
changing their contents or visual behavior. P2.3 separates Generic Modal from
Preferences composition and retires its dead cascade/raw-color debt while
preserving effective rendered states.

P4–P6 appearance ownership and P7 runner retirement are implemented. Phase-labelled
creation constraints below record those migrations' boundaries; they do not
authorize reopening completed work or running retired commands. Current commands
and coverage live in [testing](../docs/frontend/testing.md); remote/manual sign-off
status lives in the [closure record](../docs/acceptance/issue-124-frontend-engineering-closure.md).

#145-B supersedes the historical Home Reveal/prewarm/lock and Footer-debt
exceptions below: Home content and Footer MUST be actual SSR DOM. Both locales
MUST hydrate without mismatches and retain the initial Hero node. Scrolling is
native in Normal/Reduced modes; only desktop ToolDock activation observes content
geometry. Its directory remains requested on search-open. See the
[B acceptance record](../docs/acceptance/145-b-hero-scroll-content-readiness.md).

For frontend decisions, resolve evidence in this order:

1. Executable code and tests for actual behavior and compatibility.
2. This contract for intended frontend architecture.
3. `apps/cn/web/AGENTS.md` for operational guidance.
4. `docs/frontend/*` for explanation and usage.
5. Historical `apps/cn/web/docs/*` for past migration context only.

`frontend-style-debt.json` records migration state; it is not design authority.
Historical completion claims such as the `v2.2.x` style migration MUST NOT be
interpreted as completion of #124's engineering foundation.

Contributors MUST report discrepancies. Existing noncompliant code is evidence of
debt, not precedent that overrides the new-code contract. Repository-wide and
[managed asset](assets.md) contracts still apply; this contract does not change
API, SSR, hydration, routing, cookies, local backgrounds or resource fallback.
React Admin follows its separate [frontend contract](admin-frontend.md).

## Site Detail runtime and route ownership (#109 P1)

Site identity, view count and Site Insights belong to `siteId`. Current Target
observations belong to `siteId + selectedDomain + lang`. `useSiteDetailPage`
loads `/nav/sites/:id/detail` under that identity; `SiteDetailPage` loads
`/nav/sites/:id/insights` under Site ID alone. Domain and workspace query MUST NOT
hide Insights or enter its fetch key. A normal hydrated visit has exactly one
Detail GET, one Insights GET, one Recommendations GET (Task E), and one View POST, plus the active lazy slice below.
Task C makes Observation/Performance the default entry, so default hydration adds
exactly one Ping history request; SSR requests Detail, Insights and Recommendations in parallel.
Outside P4 Performance, switching
Target within the same hydrated Site session MUST fetch only Detail; no Target
switch may recount View. UI-only
state MUST NOT introduce trend/history, comparison or other data requests except
the explicit P4 Performance history and P6 ecosystem trend slices below. P1 initially made the legacy Ping
chart sample-triggered. P4 replaces that behavior with hydration-only lazy history
on entering Performance; P6 adds one selected ecosystem trend on hydrated Insights
activation. Overview and Security retain the original budget.

Detail 404 (missing Site or foreign Target) and Detail 503 are authoritative page
failures, including a failed client Target switch. Insights failure is optional:
the Site page remains 200 and the slice is unavailable. Successful empty
capabilities/changes are distinct from unavailable. View failure is a side effect
and MUST NOT replace Site/Target data or fail the page.

`app/utils/siteDetailRouteState.ts` is the sole query parser/normalizer/builder.
`siteRoutes.ts` delegates Target links to it and separately owns the query-free
Entity path. Vue components MUST NOT create competing query preservation rules.

| State | Vocabulary | Default |
| --- | --- | --- |
| `tab` | `overview`, `observation`, `security`, `insights` | `observation` (Task C) |
| Observation `view` | `overview`, `performance`, `http`, `dns`, `web` | `performance` (Task C) |
| Security `view` | `overview`, `tls`, `web`, `exposure` | `overview` |
| Insights `metric` | `ipv6`, `tls13`, `http2`, `hsts`, `csp`, `security_txt`, `certificate_verified` | `ipv6` |
| Insights `range` | `30d`, `90d`, `all` | `30d` |

Invalid UI values fall back; they never cause 404. Query values use the first
array entry, trim whitespace and are not decoded again after Vue Router. Target
membership remains an authoritative backend decision; unknown nonempty domains
MUST NOT silently become the primary Target. Normalization need not rewrite an
incoming URL. Task C builders omit tab/view for Observation/Performance; Site
Overview explicitly keeps `tab=overview`, and Observation Overview explicitly
keeps `tab=observation&view=overview`. Other Observation/Security views keep
Domain/tab/view (Security still omits its default Overview); Insights keeps
Domain/tab/metric/range. Target switches retain the active valid workspace;
primary-tab changes reset the previous workspace's secondary state, even shared
view names. Canonical, hreflang and sitemap always use localized `/site/:id`
Entity URLs without Domain or UI query. Existing SEO copy remains unchanged.

`siteCapabilityRegistry.ts` owns the seven-capability presentation catalog,
stable order, categories and translation keys, never API facts or coverage.
P1 retained an explicit three-item preview for the old panel. P6 retires that
subset metadata together with the old panel; all active consumers use the catalog.

P1 is runtime-only. Its Browser owner is `site-detail-contract.spec.ts`, using
the existing deterministic Nitro/upstream fixture and strict request/error
ledgers; `insights-entity` retains the Ecosystem-to-Entity bridge. P1 MUST NOT
change style-debt budgets, ESLint suppressions or Visual specs/PNGs. No new
workspace UI is implied by route vocabulary. P2 owns Shell/Target Context, P3
Overview, P4 Observation, P5 Security, P6 Insights workspace, P7 appearance/debt
migration and P8 Visual/closure. #108 remains separate.

## Site Detail shell and Target context (#109 P2)

`SiteDetailPage` owns the responsive shell. Its identity-only Hero contains icon,
name, display domain, description, country, SFW/NSFW, welfare, views and Visit.
Task C additionally exposes one reliable, explicitly typed Current Target CDN
hint in Hero; detailed edge evidence and Target selection do not belong there.
The former domain hover popover and large signal-card path are retired.

`siteTargetPresentation.ts` is the shared Current Target adapter for the six-item
Health Strip and the responsive Target Context. It reads only matching Target
summary/latest evidence for health, preserves unknown/absent versus zero/false,
and carries infrastructure hint confidence. Site summary supplies Target catalog
and relation hints only; Site aggregate status MUST NOT stand in for Target health.
Visit uses observed HTTP final URL, falling back to the selected Target over HTTPS.

Four primary tabs use the P1 route-state helpers and router history, with roving
focus and ArrowLeft/ArrowRight/Home/End support. P2 introduced no secondary
navigation; P4 owns Observation's secondary views. After #145-A, primary tabs
stick below the floating NavBar at widths >=768px; on mobile the public NavBar
is an absolute overlay that scrolls away (#145-A1); primary tabs still stick at
the viewport top.
At `xl` the workspace/context tracks are 3:1 with
a bounded minimum sidebar; smaller viewports place compact context before the
workspace, retaining Target/protocol/time while infrastructure detail stays in
the desktop sidebar. Mobile Health Strip is two columns by three rows. One Context/selector
instance serves all widths; it supports pointer/touch, keyboard selection,
Escape/outside dismissal and focus return.

Pending Target requests preserve Hero, tabs and shell, retaining the explicitly
labelled last resolved Target evidence until the next result is ready. A late
response MUST NOT overwrite a newer Target. Site Insights fetching stays Site-owned
across tab/Target changes. P2 initially showed the preview in Overview and Insights;
P3 replaces Overview and confines that legacy panel to Insights. P4 and P5 replace
the Observation and Security transitions respectively.

`app/assets/styles/pages/site-detail.less` owns new appearance under the exact
`.site-detail-page` / `html.dark .site-detail-page` / `--site-detail-*` token roots.
P2 removes only the old Hero/popover/root appearance debt. Other legacy panels,
#108, accepted Visual configuration/specs/PNGs and P8 golden creation are outside
scope. `site-detail-shell.spec.ts` owns interaction/layout/pending/race coverage;
P1 request, failure and SEO contracts remain required. Maintainer visual review
at 1440/390 Light/Dark (and preferably 768 Light) is required before P3; local
Functional passes are not that approval or remote CI acceptance.

The internal `insights` tab value is unchanged; public copy follows the accepted
“生态观测 / Ecosystem” naming contract rather than reviving retired product names.

## Site Overview workspace (#109 P3)

`SiteOverviewWorkspace` replaces the Overview protocol checks and legacy preview
with Site Health, conditional Attention, a full seven-capability snapshot and at
most four recent changes. Current Target evidence remains in the Health Strip,
Target Context and Observation/Security. The outer P2 shell is unchanged; the
Overview's capability/change columns use 3:2 at desktop and stack below `xl`.

`siteOverviewPresentation.ts` is the pure Site-only presentation owner. Health
uses `site_summary`, never Target summary or latest protocols. The first resolved
Site/language summary is retained for the hydrated page session; Target detail
refreshes cannot replace it. Reload or a different Site/language adopts a new
snapshot. Summary `state` (ready/stale/missing) and health `status` are independent;
stale healthy remains visibly stale and missing is not backend unknown. Show
nonzero distribution counts and `generated_at`, never Target observed time.

Attention is absent for a healthy, fresh, reason-free Site. Task B replaces P3's
original message-first presentation: known reason codes use frontend localized
copy, repeated Target reasons group by code, and unknown codes retain a visible
fallback. Without codes, localized status explains affected Targets; backend
Chinese messages must not leak into English UI. Collector remains the health
classification owner. Healthy or unknown Targets without meaningful evidence
remain excluded; Attention does not repeat Current Target cards.

Capability rows come from the P1 registry, including grouping: Network owns IPv6
and HTTP/2, Transport owns TLS 1.3 and certificate verification, Web policy owns
HSTS, CSP and security.txt. P3 corrected HTTP/2's earlier Transport assignment.
All seven rows remain present for success-empty. Backend states are preserved;
frontend `missing` means a successful response lacks that fact. Unsupported is
neutral, not a health failure. No ecosystem percentage or coverage appears here.

Changes use the shared detector labels/order/precision helpers, a four-item cap
and an explicit UTC zone for exact times so SSR and hydration agree. Day-only
events remain dates. The full Ecosystem link uses the P1 route builder and keeps
Target context. Public copy retains the accepted Ecosystem naming.

Site Insights remains fetched once by Site ID, regardless of which tab mounts.
Overview introduces no fetch, retry UI, metric/range control or history/trend
request. Detail failures remain authoritative; Insights failure makes only
capabilities/changes unavailable, distinct from successful empty facts/changes.
View failure remains harmless. `site-overview.spec.ts` owns the additional runtime
contract and `site-overview-presentation.test.ts` owns the pure projection.

Appearance extends `site-detail.less` with existing tokens and no new debt. P4–P6,
#108 and final P8 Visual goldens remain out of scope. P3 needs maintainer review
of 1440/390 Light/Dark and preferably 768 Light before P4; local Functional passes
and review screenshots do not constitute that approval or remote CI acceptance.

## Site Observation workspace (#109 P4)

Observation owns Current Target evidence and five non-sticky secondary tabs:
Overview, Performance, HTTP, DNS and Web. `selectSiteObservationView` extends the
existing route-state owner; it retains Target and clears foreign tab state. Task C
replaces P4's default Overview with Performance, omitting both tab/view for that
state and keeping Overview explicit. Tabs use router history/reload, roving tabindex and
ArrowLeft/ArrowRight/Home/End; mobile navigation scrolls within its own row.

`siteObservationPresentation.ts` is the only raw-payload projection for these
views. Match Target identity before reading summary, core or light-probe evidence.
Overview protocol rows retain status, duration, observed time and freshness as
separate fields. Task B and the Observation refinement supersede P4's original
message-first copy: Attention uses Current Target health reason codes with
frontend localization; raw diagnostic flags stay in DNS evidence. Only observed
endpoint facts appear. HTTP shows summary, conditional redirect chain and common
headers plus native disclosure for all headers, without Security interpretation.
DNS groups actual A/AAAA/CNAME/MX/NS/TXT/CAA/SOA records, follows collected children
for resolution chains, and discloses infrastructure/DNSSEC/PTR and reported risk
evidence. Web is an allowlist: metadata, robots, llms.txt, page assets and RDAP.
Security probes never enter that projection.

Performance's default 20 samples MUST auto-load on first hydrated activation,
including a direct Performance URL. SSR still requests only Detail + Insights.
`useSiteObservationHistory` lives on the page and owns the Site/Target/Ping cache,
loading/ready/empty/unavailable states, explicit slice-local retry and local
20/60/100 slicing of one `limit=100&protocol=ping&payload_mode=preview` request.
All outcomes, including empty/unavailable, stay cached until explicit retry or a
new page session. Captured keys isolate late responses; they may populate their
own cache but MUST NOT replace the selected Target. Browser history requests use
zero automatic client retries; the existing Nitro GET retry on 503 remains counted.

Entering/leaving other views adds no request. Performance Target changes add
Detail plus one history request only for an uncached Target; all other Target
changes add only Detail. No action repeats Site Insights or View. Detail failure
remains authoritative and suppresses history activation; history failure is local.
The P3 Site snapshot stays unchanged across these Target/view transitions.

The timing waterfall displays independent measured stage lengths on a common
scale and explicitly warns that they may overlap; Total is the collected value,
not a stage sum. Ping loss_rate is already a collector percentage (0–100), never
an inferred ratio. Missing RTT/loss is not zero. ECharts mounts only with visible,
ready numeric evidence, shows preparation/no-RTT copy when needed, reads resolved
theme tokens, resizes and disposes its shallow instance, and has no double-update
workaround. History rows expose time/status/RTT/loss without raw payload.

P4 removes consumer-free Observation components and their measured debt only.
Security's remaining renderer, the old light-probe renderer, Site Changes/Insights,
#108 and later-phase cleanup retain their own owners. Appearance stays in
`site-detail.less`; no parent deep overrides, transferred budgets or final P8
goldens are allowed. `site-observation.spec.ts`, the pure presentation unit tests
and real Nuxt history tests own verification. Maintainer visual acceptance of
navigation, timing/chart density, HTTP/DNS/Web and mobile/themes is required before
P5; local tests and screenshots do not grant that approval or remote acceptance.

## Site Security workspace (#109 P5)

Security presents Current Target technical evidence in Overview, TLS & Certificate,
Web Security and Exposure. It cannot assess content safety, overall trustworthiness
or vulnerability. No score, grade, safe/unsafe verdict, risk percentage or inferred
WAF deployment is permitted. The presentation has no `score`, `grade`, `rating`,
`riskPercent`, `overallLevel` or `wafEnabled` fields.

`siteSecurityPresentation.ts` owns Target/envelope matching, raw parsing, nullable
booleans, evidence states and all Security interpretation. Components render that
projection only. HTTP provides TLS, certificate and Header evidence; existing light
probes provide security.txt, Port Check and WAF Canary. P5 adds no endpoint or fetch:
view changes add zero requests; Target changes add only Detail, never Insights/View.
The P1 authoritative Detail and optional-source failure boundaries remain unchanged.

`selectSiteSecurityView` extends the sole route-state owner. Secondary tabs support
router history/reload, roving focus and ArrowLeft/ArrowRight/Home/End, scroll within
their row on mobile and never stick. The raw Header link uses
`selectSiteObservationView(..., 'http')`, preserving explicit Target query or the
implicit primary Target. Materializing the implicit Target into a new query merely
to change view would incorrectly cause a Detail fetch.

Verification is separate from validity. `cert_collected=false` and `not_tls`
prevent Go false/zero defaults from becoming failed certificate evidence. Explicit
verification facts remain meaningful when the older payload lacks the collection
flag. `readSiteCertificateEvidence` is shared with the Target presentation owner,
so the Health Strip cannot contradict Security with uncollected default values.
Remaining days prefer finite `cert_days_left`; when absent or invalid, derive
whole days (floor) from `cert_not_after` minus the HTTP envelope's `observed_at`.
Both timestamps must be valid and timezone-qualified; missing/invalid or Go-zero
times remain not observed. Never use the client clock or a Site/Target summary
timestamp. Health Strip, Security summary and validity facts share the same
expiry projection. Verification remains independent. >30 is normal,
8–30 attention, 1–7 warning and <=0 expired. These tones belong to expiry alone.
SAN, chain and crypto facts use native disclosure; fingerprints wrap in monospace.
False OCSP and zero/missing SCT do not produce a security conclusion.

Header state prefers `security_header_summary.present`, then the collected boolean
map, then raw Header presence. The frontend does not reimplement a Header analyzer.
Observed absence is Missing; failed HTTP is Unavailable; missing evidence is Not
observed. security.txt distinguishes found, found-with-validation-issues, not-found,
unavailable and not-observed. Only collector `validation_errors` drive validation
issues, without a second RFC validator or client-clock expiry check.

Ports retain neutral open/closed/timeout/filtered/skipped evidence; an open port is
not a vulnerability. Collection metadata is disclosed. WAF shows reported counts
and cases; a matching statement requires a complete successful sample and known,
matching counts, never proof of WAF deployment. Truncation is visible outside the
case disclosure. Attention is limited to verification failure, reported expiry,
security.txt validation issues and positive reported Canary mismatch/error counts.

P5 retires the consumer-free SiteObservationMetricGrid and SiteLightProbePanel only.
Appearance extends the existing Site Detail owner, keeps deep selectors at zero
and lowers only actual removed debt. Other cleanup remains P7; Insights remains P6
and #108. `site-security.spec.ts` and pure projection tests own the contract. No
accepted Visual file changes or final P8 goldens; maintainer visual review is still
required before P6, independently of local gates and remote CI status.

## Site Insights workspace (#109 P6)

`SiteInsightsWorkspace` replaces the three-item legacy panel with a lightweight
Site-wide header, seven selectable registry rows, selected Site fact and ecosystem
context, lazy ecosystem adoption trend and the complete returned recent-change set.
There are no secondary tabs, embedded Dimension Explorer or Compare components.
Links lead to localized `/insights/sites` and `/insights/sites/compare?ids=<siteId>`.
The existing public Ecosystem / 生态观测 naming remains authoritative; the P6
brief's Insights / 洞察 examples do not reintroduce the retired product name.

`siteInsightsPresentation.ts` is pure and has no Target input. Registry categories
and order drive the matrix. Slice `ready/empty/unavailable` is separate from the
seven backend fact states and frontend `missing`. A failed slice exposes null fact
state and a dash, never seven invented backend unavailable facts. Successful empty
capabilities yield seven missing rows. P3 follows the same failure distinction.
Adoption value and known/eligible coverage have separate labels; null is not zero.

`useSiteInsights`, instantiated only by the page, owns SSR data/state/retry shared
by P3 and P6. It is keyed by Site ID, validates returned identity, and guards the
published snapshot against a changed Site ID. Retry refreshes only this slice;
the old Site's retry cannot replace a new Site's result. It never refreshes Detail,
View, or trend. Existing transport behavior is unchanged; owners do not schedule
automatic retries. A failed optional slice never becomes an authoritative error.

`selectSiteInsightMetric` and `selectSiteInsightRange` own URL transitions and
retain Target and the other selection. `useSiteInsightTrend` belongs to the page,
activates only after hydration in Insights, and caches by metric + range without
Target or Site identity. It uses the existing `getNavInsightsTrend` service. Every
active identity has loading/ready/empty/unavailable state. Empty and failed results
are cached; only explicit retry reloads them. Late results populate only their own
entry. Normal activation adds exactly one uncached trend; SSR still loads only
Detail and Site Insights. Target changes add only Detail, including while trend or
Site data is pending, and never reset the Site workspace.

Site and trend failures are independent. ECharts is lazy canvas with a shallow
instance, ResizeObserver, disposal and import revision protection. It plots only
ecosystem adoption, uses `connectNulls: false`, and reads Site Detail/global tokens.
Loading includes chart initialization; import failure is classified unavailable.
One-point and all-null samples retain explicit usable/unknown presentation. No
state may leave an unclassified blank chart. #108 chart appearance is not imported.

Recent Site Changes uses all returned items, shared order/date/label/category
helpers, unknown fallback and exact UTC versus day-only precision. It is never
described as complete history. The old SiteInsightsPanel and its sole-consumer
InsightsEntityTimeline are retired after consumer audit. Their isolated styles
are removed; shared Game rules and #108 important/domain debt remain untouched.
New appearance belongs to `site-detail.less`, with deep zero and no added debt.

`site-insights.spec.ts`, pure presentation tests and real Nuxt owner tests verify
this contract. P1–P5 and Entity tests retain their existing invariants while
accounting explicitly for P6's first activated trend. SEO remains Entity-only.
No backend/schema changes, accepted Visual updates or P8 golden are authorized.
Maintainer visual acceptance is required before P7; remote CI is a separate status.

## Site Detail legacy and appearance closure (#109 P7)

P7 deletes the consumer-audited SiteHealthSummaryPanel, SiteOverview,
SiteSignalCards and SiteChangeEvents, their private helpers/styles and all fourteen
retired detailTypes definitions. No explicit import, Nuxt template tag, dynamic
registration or test consumes these components. The unused SiteObservationHistory
type alias is also retired; active history behavior is unchanged.

Site Detail Tailwind appearance, arbitrary appearance and raw visual debt are
zero. Deep selectors and legacy dark entries remain zero. `site-detail.less`
retains the P2–P6 appearance; no healthy selector/token or active UI is redesigned.
Historical Site dark/deep exceptions MUST NOT be treated as current permissions.
The remaining baseline is ambient raw 75 and #108 important 5; no budget transfer,
increase, new exception or unrelated cleanup is authorized.

Retired `site.*` messages are removed from both locales after consumer audit;
`site.siteDnsPanel.none` remains required by Home. The five active Site namespaces
and shared messages retain their values. Existing active ESLint suppressions stay.
All P1–P6 route/request/cache/retry/SSR/SEO/evidence contracts and tests remain
authoritative. P7 adds no feature or final golden. Maintainer Desktop/Mobile smoke
of Overview, Observation, Security and Insights precedes separately authorized P8.

## Site Detail final presentation refinement (#109 Task B)

Task B supersedes the P2–P6 appearance assignments without changing their runtime,
route, API, request, SSR/SEO or Site/Target ownership contracts. The Site Detail
Less owner provides compact translucent Primary/Secondary surfaces and native
disclosures, quiet evidence colors and a shared segmented pattern for Observation,
Security, Ping samples and ecosystem ranges. New/replaced system icons use
Phosphor. Existing public theme tokens remain authoritative; #108 and ambient
appearance are outside this task.

Hero uses a multi-row logo and the full identity content column. The Health Strip
has exactly six primary values, no helper rows, and shares P5's certificate
normalization/expiry presentation with Security. Target Context presents the
selector, Ping/HTTP/DNS and observed time. P2's relation/infrastructure debug
display is retired, while underlying evidence and selector behavior stay intact.

`siteDetailPresentation.ts` owns localized health reason labels and display tones,
never health aggregation. DNS flags remain diagnostic evidence: PTR-empty,
low-TTL and unclassified signals are informational; private-IP and NXDOMAIN with
answers are warning-colored. Security continues to separate verification from
validity; header absence, open ports and canary matches never become safety or
WAF deployment claims. Known security.txt validation codes map to localized copy;
the frontend does not revalidate the file.

Overview keeps its four-change cap and moves its ecosystem/retry actions into the
Capability surface. Insights keeps all returned changes, groups its seven matrix
rows by the shared registry and contains selected fact plus trend in one analysis
surface. Adoption, coverage and dates remain neutral; only fact state uses status
color. No secondary Insights tabs, extra requests or new route vocabulary appear.

`site-detail-refinement.spec.ts` adds focused presentation checks using the shared
deterministic runtime fixture. The six existing P1–P6 Browser owners retain their
request, failure, cache, race and accessibility assertions. Optional temporary
review screenshots are emitted by that same Functional owner, never a new runner
or accepted Visual baseline. Site debt, deep and legacy-dark remain zero. Full
maintainer visual acceptance is required before separately authorized P8 work.

## Site Detail first-round composition completion (pre-P8)

This appearance contract supersedes earlier P2–P6/Task B card, divider and Context
surface descriptions; their Site/Target, route, request, failure and cache
contracts remain unchanged. `site-detail.less` owns shared acrylic Plane A/B/C,
hover, selected and 500ms motion tokens. Content surfaces have no perimeter
border. Facts/list rows use 4–6px gaps, small corners and soft color-only hover;
selected states respond immediately. Keep only meaningful relationship/chart
lines, the primary active indicator, focus rings and the Identity Note separator.
The Target selector may retain necessary control affordances.

Hero and the six Health values share one Identity Note with an inset CSS dashed
separator and three decorative, aria-hidden perforations. Health has no grid
rules. Overview and Observation retain their accepted composition; only materials,
unnecessary borders/dividers and row spacing change. Observation and Security
own their H2 and finite, scrollbar-hidden secondary navigation in one header row,
including on mobile. Preserve roving tabindex, keyboard navigation and route state.

Security keeps four evidence-only views. TLS transport, verification and validity
share one composite while remaining distinct facts; certificate identity is a
secondary plane and crypto a native disclosure. Headers show name/state above
value. security.txt metadata/errors and request cases stay in disclosures. Port
Observation and Request Behavior Check never imply vulnerability or WAF deployment.
Existing certificate normalization, expiry thresholds and explicit Attention
evidence remain authoritative.

Insights has an unboxed header and seven-row Capability Explorer, one primary
Analysis plane and an unboxed Recent Change Stream. Adoption/coverage remain
neutral; the adoption-only trend uses the info accent, not success green. No
secondary tabs, score, filter, pagination or request is introduced. Unavailable
Site Insights retains seven structural rows and the shared inline retry.

The existing Functional Browser owners verify this composition and may export
temporary review screenshots. Full local verification does not grant maintainer
visual acceptance or remote CI acceptance. Stop after first-round completion;
P8 and any second round require separate user authorization. Accepted Visual PNGs
and the zero Site/deep/legacy-dark debt budgets stay unchanged.

## Site Detail Task C consistency and default entry (pre-P8)

Task C supersedes the first round's Plane A/B/C opacity hierarchy, retaining its
composition. `--site-detail-panel` is the sole base material for Identity Note,
Context and all content panels across workspaces, with the same hue and opacity
within a theme. Flows/ghost disclosures are transparent. One shared hover and
same-hue stronger selected token serve rows, segments and the local Target
trigger. Color-only hover takes 500ms; selection is immediate. Do not modify the
global button primitive to style this trigger. Preserve selector behavior/a11y.

Hero view count keeps `Intl.NumberFormat` and an accessible Eye label with stronger
warm brand emphasis. CDN comes only from Current Target `edgeProviderHints`, with
`type=cdn` and medium/high confidence, preferring high and preserving input order
for ties. Provider names cannot imply CDN type. `siteTargetSignals.ts` owns this
selection/localized provider mapping and the protocol status projection used by
Context and Observation Overview. Success has a shared dot and screen-reader
text; stale/failure/unknown stay visibly labelled. Latency tone is independent of
success, and Collector health classification is unchanged.

The route table above records Task C's intentional UX change. Clean and domain-only
URLs render Observation/Performance; entering primary Observation selects
Performance, while selecting the already active primary tab preserves its view.
Invalid UI values fall back without changing authoritative Target validation.
Default history remains hydration-only, one `protocol=ping&limit=100&payload_mode=preview`
request and a local 20-sample slice. Returning uses P4's existing cache. While
Performance is active, a new Target may load its own uncached history; non-Performance
switches still fetch Detail only. Neither recounts View or refetches Site Insights.
P4/P6 cache/race internals, API shape, Site/Target ownership and Entity-only SEO
remain unchanged. No final Visual golden or P8 work is authorized by Task C.

## Site Detail Task D composites, help and focused defects (pre-P8)

Task D retains Task C material/default-entry and all data identities. It groups
Overview Summary/Attention/Capabilities, Observation Current/Endpoint/Attention,
HTTP Response/Redirects/Headers, Security Summary/Attention and Certificate
Identity/Crypto under their respective single panel owners. Inset dashed lines
may separate these related sections; they do not introduce perimeter borders or
Hero perforations. Summary fields have only a label and primary value; TLS
auxiliary evidence remains in a native disclosure. Security and Site Insights
scope copy is removed without changing either ownership contract.

`SiteDetailHelpTooltip` owns active Site Info hover/focus/click, Escape/blur and
ARIA association; native `title` is not its visible help. Waterfall has no extra
heading/help. `SiteChangeStream` adds local Compact/List controls with the same
three-column serpentine geometry as Game Timeline (`serpentineSequence.ts`).
Mobile remains a list; chronology, date precision, Overview's four-event bound
and Insights' complete recent set are unchanged. Redirects use this geometry
without inventing hop status. Hero name/domain share a heading row, while views
share the badge row; Current Target CDN selection is unchanged.

The four touched charts normalize non-finite/non-numeric/missing values to a gap
and `—`, never zero. Game `average_30d` has an independent one-decimal maximum;
current/peak counts stay integer. No new chart/tooltip dependency is introduced.
The Gallery test's expected abort is single-use, exact-request/exact-resource,
`net::ERR_ABORTED` only, and must follow intentional Gallery unmount and actually
occur. Other failed requests and browser errors remain fatal.

Task D explicitly limits local validation to focused cases, typecheck and relevant
lint/style checks, without full build/suites. The opt-in local source fixture
mode is documented in the testing guide and is not production/remote acceptance.
Keep Site/deep/legacy-dark debt zero, #108/ambient budgets and accepted Visual
inventory unchanged. Stop before P8.

## Site Detail Task E — Similar Sites discovery (pre-P8)

`GET /api/v2/nav/sites/:siteId/recommendations?lang=zh|en&limit=8` owns optional
Site Entity discovery. Limit defaults/caps at eight. It returns schema_version 1,
UTC generated_at, state (`ready`/`unavailable`), site_id and SiteVo items, including
view_count. It neither reads display mode nor accepts Target/workspace selectors.
Invalid ID/language/non-integer limit returns 400; unavailable read models return
an unavailable slice with an empty array. Existing Detail remains authoritative.

Backend resolves full Group membership through the existing localized Site/Group
reader, then feeds matched groups into `BuildHomeGroupsForCache`. Each group's
Home Top-8 cutoff precedes union, ID dedupe and self exclusion. SHA-256 of
`siteId|candidateId|UTC YYYY-MM-DD`, sorted lexicographically (ID tie-break), selects
up to eight. Never refill from rank nine, unrelated groups or random global Sites.
No ranking/reason/score, new table, Redis key or refresh job is introduced; existing
read-model cache/DB fallback and Home ordering remain their current owners.

`useSiteRecommendations` loads concurrently with Detail/Insights, keyed only by
Site ID + normalized language. It captures identity and ignores stale Site/locale
responses. Hydration reuses SSR data; Target/query/mode/local-tab changes do not
reload it. Errors/unavailable/empty hide discovery without an error card or retry.
Default Performance still loads Ping after hydration. Task E's shorthand
"Target switch = Detail only" applies outside Performance: existing P4 permits a
new Target's uncached Ping history while Performance is active. Task E does not
change that exception or P4/P6 cache/race/retry behavior.

`SiteSimilarSites` renders a single shared list below Current Target on Desktop,
or as Mobile's local Similar panel. Raw nonempty items control the auxiliary tab;
SSR rendering filters to SFW, then mounted display-mode subscription filters raw
items without fetching. Mobile can show a neutral empty message when all raw
items are hidden; Desktop hides an empty visible list. No hidden-item reason is
shown. Raw optional-slice data remains in Nuxt payload for local mode switching;
NSFW names/icons/links never appear in SSR-rendered recommendation markup.

The four route-owned tabs stay unchanged. Similar is a mobile-only local state,
with tab semantics/keyboard order, no URL/history entry and no independent SEO.
Desktop resize clears the local selection and restores the underlying route
workspace without navigation. Primary tabs scroll horizontally only when needed.
One Site Detail appearance owner supplies borderless rows and 500ms tint; list
content is limited to managed icon, name, one-line info and formatted view
snapshot. P8 preflight removes the display domain only; the recommendation API's
SiteVo domain stays unchanged. Links use localized query-free Entity paths; the Similar
component never increments views. The destination page retains exactly-once View.

Task E keeps verification focused on recommendation/group Go tests, recommendation
SSR/request/discovery Browser cases, typecheck and relevant lint/style checks.
Do not run full frontend tests/build or create final Visual goldens. Stop before
P8 and await maintainer visual acceptance.

P8 preflight aligns Functional assertions with the accepted Task D/E product:
Waterfall has no native-title or Info help; redirects assert ordered nodes and
the semantic connector wrapper, never Phosphor SVG internals. Capability/header/
port/change tests own counts, states, category/date and Compact/List interaction.
Exact gaps, typography sizes and serpentine pixel placement belong to the future
Visual contract. Existing overflow, keyboard, accessible tooltip and request
checks remain. Trend tests must wait for URL, pressed controls and metric/range
identity before ready: tls13|90d → csp|90d → csp|all still requires three requests,
cached Back/Forward adds none, and reload adds a fourth. Preflight grants no
runtime/cache change or final-golden authorization.

## Site Detail final Visual contract and closure (#109 P8)

P8 freezes the current product. The phase-labelled P1–P7 and Task A–E records
remain historical evidence; their intermediate layouts are not the final Visual
contract. Task C supersedes the early default Overview and acrylic-level hierarchy,
Task D supplies the final composites/help/flows, Task E adds Site-level discovery,
and preflight removes Similar's display domain. Task A's Collector health reasons
and the later evidence-derived certificate fix remain in force. P8 adds no route,
feature, API, recommendation rule or runtime change.

The final route owner still has four workspaces: Overview, Observation, Security
and Insights. Blank/domain-only entry means Observation/Performance. Mobile Similar
is local state with no URL/history/SEO mutation. Site ID owns Insights and View;
Site ID/language owns Recommendations and the initial Site summary. Current Target
owns Detail evidence and Ping history. SSR loads Detail + Insights + Recommendations;
hydration counts View once and activates only the current lazy slice. Target changes
never reload Insights/Recommendations or recount View. Active Performance may fetch
the new Target's uncached Ping history; elsewhere the switch reads Detail only.
Insights trend is ecosystem-wide, cached by metric/range without Target. Keep both
lazy/cache/race owners, optional-slice failure isolation and Entity-only SEO intact.

Certificate expiry has one projection shared by Health, Security summary and
validity detail: finite explicit `cert_days_left` wins, otherwise derive whole days
from `cert_not_after - HTTP observed_at`. Never consult the client clock. Missing or
invalid evidence remains not observed; <=0/<=7/<=30/>30 retain expired/warning/
attention/normal meanings. Verification remains distinct from validity.

Recommendations retain full-group membership, per-group Home Top-8 eligibility,
union/dedupe/self exclusion and deterministic UTC-day selection capped at eight;
there is no global backfill or score. Target/mode/tab changes add no recommendation
request. SSR markup is SFW; mode filtering is local. Similar shows managed icon,
Name + Info + Views, no domain line, and only the destination page counts View.

`tests/browser/fixtures/site-detail-visual.ts` and `visual/site-detail.spec.ts` own
exactly eight initial Site Detail PNGs under the latter spec's snapshot directory:

| Viewport | Theme / scene | Golden |
| --- | --- | --- |
| 1440×900 | Light Performance | `site-detail-performance-light-desktop.png` |
| 1440×900 | Dark Performance | `site-detail-performance-dark-desktop.png` |
| 1440×900 | Light Overview | `site-detail-overview-light-desktop.png` |
| 1440×900 | Light TLS | `site-detail-security-tls-light-desktop.png` |
| 1440×900 | Light Insights | `site-detail-insights-light-desktop.png` |
| 1440×900 | Light HTTP | `site-detail-http-light-desktop.png` |
| 390×900 | Light Performance | `site-detail-performance-light-mobile.png` |
| 390×900 | Light Similar | `site-detail-similar-light-mobile.png` |

These are Site Detail root captures at the listed viewports, so image height may
exceed 900px to include the representative workspace/aside. Reuse the existing
digest-pinned Linux Visual runner, not a separate config. Fixed UTC evidence/browser
time, local managed assets, fixed history/trend/changes/recommendations and strict
network diagnostics make the fixture deterministic. Its opt-in diagnostic timestamp
uses that same instant; existing runtime fixtures retain live timestamps. Wait for
hydration, active chart readiness, loaded images/fonts and finite animations;
blur focus, move the pointer away and check overflow. No sleep/networkidle or
diagnostic suppression. HTTP has seven redirect nodes; TLS derives 45 days without
legacy days; Mobile Similar retains the URL and omits domain text.

Visual owns composition, spacing, material, density, typography and responsive
arrangement. Functional owns route/history, requests/cache/race, semantics, a11y,
errors, exactly-once View, mode filtering, certificate derivation and SEO; do not
restore exact pixel-gap assertions there. The eight existing Site Detail regression
owners remain mandatory, including recommendation/refinement owners. Scoped golden
generation MUST be followed by an independent compare with updates disabled.

Site Detail measured appearance/arbitrary/raw/important/deep/legacy-dark debt stays
zero. Unrelated ambient raw 75 and #108 important 5 remain unchanged. Narrow cleanup
may remove proven consumer-free owners only. A real product defect blocks closure;
never update a golden to conceal it. Current-code remote build, all Browser shards,
Visual and repository-policy must pass. The acceptance ledger records actual local
and remote results separately from explicit maintainer approval of the eight PNGs.
Until that approval, the new baselines are review candidates, not accepted design.
After approval #109 is closure-ready; subsequent product work uses a new task, not P9.

## Game Collection Public ownership (#140-C)

Home has exactly three concurrent SSR GETs: Home, Showcase and Collections Home.
Collections Home is always SFW. Both optional slices have a 1000ms SSR budget,
retry 0; unavailable slices get exactly one mounted recovery (8000ms, retry 0,
including the Nitro proxy). Recovery failure stays absent and silent; successful
SSR, including valid empty, never refetches. Language identity/unmount invalidate
late recovery; mode changes never refetch Home. Valid empty slots retain All
Collections on desktop. Shortcuts exist only in the xl Sidebar, sharing Daily
Game's sidebar-action-button material and 2.45rem minimum height; below xl their
DOM is removed. Showcase runtime/tracking remains independent.

Collection Index/Detail SSR must be SFW. Detail retains its mode refresh owner;
Index uses one discovery owner for committed q/phase/sort, mode, snapshot and pagination. It reads
local storage only after mount, issues no extra SFW request, and one NSFW request
when needed. Latest response wins; ready content survives pending/error, old-mode
pagination cannot append, and successful Index mode refresh resets page one.
No URL mode, frontend adult filtering, or frontend chronology sort is allowed.
Detail partitions Backend order without changing it; past/future independently
reuse compactPlacement/connector and retain chronological DOM order at every width.
Only those two groups have connectors. Detail's compact Header groups H1/count with
right-aligned text Back navigation. Timeline cards and NOW share Games Home material,
transparent borders, and fixed two-line title/summary geometry. Optional schema v1
metadata shows at most primary/secondary tags and rating/latest observed online/community
counts with localized accessible labels; missing metadata is omitted, including old cached
payloads. No client filtering or chronology change follows from decorations.
Connectors are 2px dashed CSS gradients over 2.5rem gaps; right/left/down motion follows
chronology, small screens flow down, reduced motion keeps static dashes. Connectors cannot
capture pointer/focus or join undated phases. Date-only facts use UTC/component formatting
and retain precision/inferred; NOW uses the response as_of_date. Missing Detail is
404, initial upstream failure is HTTP 503 with Retry, adult-only SFW is neutral 200.
Index begins visually with Search + Advanced Filter; its H1 is sr-only and SEO remains.
Search is IME-safe with a 350ms debounce. The modal owns a local draft: Cancel has
no side effects, Apply commits both criteria atomically. All criteria/mode requests
share one generation, retain ready cards, reset page on success, and reject stale
load-more responses. Server q/phase/sort results are authoritative; the frontend
never filters adult members or infers phases. Cards use Games Home material and
transparent borders with an inset rounded preview.
Showcase primary CTA alone has a transparent border. Stacked content reserves
context/release/title/summary/tags/note/footer geometry so slide height is stable;
desktop retains lg:h-80 and 64/36. Autoplay/tracking/artwork ownership is unchanged.
Collections use existing Games/global tokens and SteamAssetImage. New copy belongs
to game.collections. Visual acceptance is separate from Functional completion.

## Styling ownership

**Tailwind owns structure; Less owns appearance.**

Tailwind MUST own structural composition only. Visual appearance MUST be
expressed through semantic classes, Less and design tokens.

| Responsibility | Owner |
| --- | --- |
| Placement, flex/grid, alignment, position, responsive composition, overflow, visibility, page/container sizing, outer spacing | Tailwind |
| Text alignment, truncation and whitespace behavior | Tailwind MAY be used |
| Control height/padding, color, background, border appearance, radius, shadow, ring, typography, opacity, visual hover/focus/motion | Owning primitive, compound component or domain Less using semantic tokens |
| Component-private structural geometry not broadly reusable | Scoped style MAY be used |

Allowed examples: `flex`, `grid`, `items-center`, `justify-between`, `relative`,
`absolute`, `fixed`, `overflow-hidden`, `w-full`, `max-w-*`, `px-4`, `gap-4`,
`md:flex`, `lg:grid-cols-3`, `truncate`, `whitespace-nowrap`, `text-center`.
These classes are for composition: `px-4` on a container does not authorize
reconstructing a button's internal padding/height outside its primitive.

New code MUST NOT express appearance through Tailwind utilities such as
`bg-slate-*`, `text-orange-*`, `border-[#…]`, `rounded-xl`, `shadow-*`, `ring-*`,
`font-bold`, `text-sm`, `leading-relaxed`, `tracking-*`, `opacity-70`,
`hover:bg-*`, `focus:ring-*`, `dark:bg-*` or `dark:text-*`.
Responsive/state modifiers do not change the underlying ownership.

New arbitrary visual values (`bg-[#…]`, `text-[#…]`, `border-[#…]`, `shadow-[…]`,
`rounded-[…]`) MUST NOT be introduced. Structural arbitrary values such as
`z-[120]`, `max-w-[2080px]` and `min-h-[calc(…)]` are outside P0's visual counts;
their existing uses MAY remain. Future structural tokens require demonstrated
reuse and a separate scoped change; P0 does not normalize these values.

Typography is visual language: font size, weight, line height, letter spacing
and text color SHOULD migrate from historical Tailwind into semantic/domain
classes. New reusable UI SHOULD consume shared or domain typography rather than
create a local scale. Text alignment/truncation remains structural.

## Current source layout

`nuxt.config.ts` loads `app/assets/css/main.css`, then
`app/assets/styles/index.less`. Contributors MUST retain this working structure
until a separately scoped migration:

- `main.css`: Tailwind bootstrap, reset, base elements/scrollbars and generic
  helpers. It MUST NOT own global GoFurry theme/design-token declarations.
- `styles/index.less`: style composition/import root.
- `styles/tokens.less`: canonical owner of global semantic token declarations,
  including the layout-owned `--gf-page-background` and `--gf-page-pattern*` group.
- `styles/mixins.less`: reusable Less behavior built on tokens.
- `styles/primitives/*.less`: shared domain-neutral visual building blocks.
- `styles/components/*.less`: compound product UI, including preferences,
  shell, navigation and footer; this directory is not deprecated.
- `styles/pages/*.less` and existing subdirectories: current page/domain styles.
- `app/components/common`, `nav`, `game`, `site`, `insights`: existing Vue owners.

`index.less` MUST compose `tokens → mixins → primitives → components → pages`.
The primitive order is button, card, chip, input, modal, pagination, rating; retain
the existing relative page-style order. Preferences MUST load after Modal.
`domains/` and the later test layout remain future migration work; contributors
MUST NOT create empty placeholders. Existing raw values and domain theme islands
remain historical debt. Do not copy them into new code.

## Token ownership, naming and lifecycle

| Level | Meaning and ownership | Examples |
| --- | --- | --- |
| Global/Foundation | Reusable product/theme semantics in `styles/tokens.less`, shared across domains and primitives | `--gf-page-background`, `--gf-surface`, `--gf-text-main`, `--gf-border`, `--gf-accent`, `--gf-focus-ring` |
| Primitive-local | Meaning specific to a reusable primitive, declared in its owning stylesheet | `--gf-rating-empty`, `--gf-rating-fill` in `styles/primitives/rating.less` |
| Compound-local | Product composition/state semantics with an exact approved owner, selector and prefix | `--gf-preferences-*` in `styles/components/preferences.less` |
| Domain | Shared visual meaning within a domain, in that domain's existing Less owner | `--games-*`, `--nav-*`, `--updates-*` |

Contributors MUST search existing tokens before adding one. Primitive-specific
meaning SHOULD stay with its primitive rather than being promoted merely because
the primitive is reusable. Domain tokens SHOULD
alias global semantics by default, for example `--games-border: var(--gf-border)`.
An independent domain value requires a genuine business meaning and a documented
rationale, such as discount emphasis differing from ordinary interaction accent.
Component-private geometry, such as cover ratio or title line count, MAY remain
scoped. It is not another theme-token level and MUST NOT create a separate color,
shadow, radius or typography system.

Token identity MUST follow semantic role, not current literal equality.
`--gf-accent` and `--gf-accent-fill`, or `--gf-surface` and `--gf-input-bg`, MUST
remain distinct: emphasis, action fill, general surface and form surface have
different roles, including different dark-theme behavior. Contributors MUST NOT
deduplicate tokens solely because their current values match.

New global names MUST follow `--gf-<semantic-role>[-<variant-or-state>]`, such as
`--gf-surface-hover` or `--gf-accent-fill-hover`. New literal/color/property-first
names such as `--gf-orange-500`, `--gf-bg-foo` or `--gf-color-bar` require an
explicit compatibility rationale. This is not a mandate to rename active tokens.

A new token MUST have a real consumer and stable semantic meaning. Moving a
literal into a token only to satisfy `style:policy` does not meet that requirement.
Contributors SHOULD remove confirmed unused tokens unless an explicit external
compatibility contract requires retention. P2.1 removed unused `--gf-bg-grid-line`;
do not retain dead tokens for hypothetical future use.

Static/Legal roots MUST remain transparent and MUST NOT own the application
canvas; the default layout and `PublicPageBackground` own page background
semantics through `--gf-page-background`. The retired `--gf-bg-page` compatibility
token MUST NOT be reintroduced or replaced with a Static-specific canvas token.
`--gf-static-panel-shadow` owns the shared About/Legal reading-surface elevation;
Dark intentionally inherits the root value. `static.less` owns the full migrated
color-transition property set, its 500ms duration and existing easing.

Updates global selectors MUST use the `updates-*` domain namespace. State
modifiers MAY use `is-*` only when attached to an Updates-owned base selector.
Updates-owned custom properties MUST use `--updates-*`; domain styles MAY consume
`--gf-*` global semantics. P4.4.2's Timeline selector normalization is historical. #132 P3 retires that
Timeline and its dynamic delay property; the same exact token declaration owners
remain `.updates-page` / `html.dark .updates-page` with the `--updates-` prefix.

P2.1 MUST NOT prebuild typography, spacing, control-height, z-index or container
scales. Later promotion requires repeated real needs and a scoped migration.
Examples in the parent plan do not imply that `--gf-success` or `--gf-warning`
exist today. Actual values belong in their source owner, not duplicated in docs.

### Annotation and theme semantics

New/changed token groups MUST have concise comments explaining semantic meaning,
scope, why they exist and why they belong at that level. Exceptional tokens MUST
also explain the specific exception. Group comments SHOULD carry shared context;
individual comments are for exceptions, not a description of every CSS literal.
Global values stay in the single `tokens.less` owner. Its root groups are ordered
Page & Canvas, Surface, Border, Text, Accent & Action, Form Controls, Modal &
Overlay, Feedback, Focus, Elevation & Blur, Shape, Motion, Browser Chrome.
`html.dark` MUST preserve those meanings and the relative group order; omitted
tokens intentionally inherit root values. Short dark-section labels suffice;
do not repeat the root explanations or invent groups for symmetry.

```less
/* Surface: shared content and interaction surfaces across domains.
 * Global ownership keeps light/dark meaning consistent; consumers should
 * reuse these semantics before introducing domain-specific colors.
 */
```

A comment merely saying "red" is not a rationale. A domain exception comment
should explain its business role and why the global meaning does not fit.
See the [design-system guide](../docs/frontend/design-system.md) for current
repository examples; it is not a second token-value source.

## Typography family and weight contract (#126)

Gate A freezes Primary Sans as Manrope Variable + Noto Sans SC Variable through
`--gf-font-sans`, and Technical Mono through `--gf-font-mono`, in the existing
global token owner. Normal UI SHOULD inherit Sans. Explicit `font-family` values
MUST use one of these tokens; raw business-owner family stacks are prohibited.
`app/assets/styles/primitives/rating.less` alone MAY retain `Arial, sans-serif`
for its star glyph implementation, not as a third text role.

Normal authored `font-weight` MUST be 400 (Body), 500 (Secondary), 600 (UI
Emphasis), 700 (Strong Heading) or 800 (Display only). Choose by actual UI role,
not numeric proximity. 800 MUST NOT be used for ordinary controls, labels,
section/card headings or small KPI values; retained Display owners are the large
Error code and Games rank. Do not introduce weight tokens or locale-specific
weight hierarchies. Tailwind does not own typography appearance; its existing
appearance policy already covers `font-*`, `leading-*` and `tracking-*`.

The existing Stylelint `declaration-property-value-allowed-list` enforces families,
weights and `font: inherit` as the only allowed font shorthand. Do not introduce
another typography scanner, plugin, debt manifest or Style Policy rule family.

Issue #108 F3 retires its temporary legacy-weight override after auditing both
Insights pages and Game Detail consumers. `app/assets/styles/pages/insights.less`
and `app/assets/styles/pages/insights/**/*.less` now follow the standard roles
above, with no family, weight or shorthand exception. Batch B's original exclusion
does not authorize restoring intermediate weights or compensating with metrics.

Batch B changes weights only: the Gate A Fontsource packages/delivery, authored
font size, line height, tracking, spacing/layout, style debt and Visual goldens
remain unchanged. Historical migration pixel/weight notes below describe their
original acceptance; this section supersedes their nonstandard weight values.
Role examples belong in the [design system](../docs/frontend/design-system.md#typography-roles-126).

### Insights style ownership after #108 F3

`insights.less` remains a mixed consumer owner: Game Detail widgets and shared
chart/loading/empty behavior are live, not removable merely because Insights
pages have their own styles. Domain range controls belong to `insights/domain.less`;
the segmented Game controls retain their shared class within `.game-insights-tab`.
Their active/hover/focus states must not depend on competing `!important` rules.
Trend header/chart-shell and disclosure-grid composition belongs to the existing
Vue templates' Tailwind classes. Domain disclosure appearance has one Domain owner;
the compact Sites and Certificate endings are local exceptions, not a global reset.

F3 lowers important debt from five to zero through the existing updater; ambient
raw 75 is unchanged. Historical phase inventories above are not current budgets.
Existing Insights Browser owners plus `game-detail-insights` and `insights-entity`
protect shared consumers. See the [F3 record](../docs/acceptance/108-f3-frontend-engineering-closure.md)
for measured compatibility and pending manual/remote acceptance; no Golden or
runtime ownership change follows from this cleanup.

## Shared visual primitives and compound components

Physical placement in `app/components/common/` MUST NOT imply shared visual
ownership. Classify by responsibility and real consumers: Preferences belongs to
P4, PageScrollDock to P4.5, MobileBottomTabBar to P5, and BlurWrapper/LinkTag to
P6 Game. ManagedAssetImage/SteamAssetImage are runtime infrastructure; their
location does not authorize appearance or routing migration. See the current
[ownership map](../docs/frontend/design-system.md#common-directory-semantic-boundaries).

Before deleting a historical common component, contributors MUST prove zero
production consumers by checking PascalCase, Nuxt path-derived/lazy names,
kebab-case, explicit imports, dynamic components/`resolveComponent` and source
paths. Generated registrations are not consumers. Live components MUST remain;
do not rewrite consumers to force deletion. P4.1 makes no file moves or visual
migration. After deletion, inspect `style:policy` for stale-only debt before
running its downward-only updater; never transfer or raise budgets.

A reusable, domain-neutral visual building block MUST be treated as a primitive.
It has stable appearance semantics, composes into larger UI, and SHOULD NOT
depend on a page root or business token namespace. The seven current owners in
`styles/primitives/` are button, card, chip, input, modal, pagination and rating,
exposing `.gf-button`, `.gf-card`, `.gf-chip`, `.gf-input`, `.gf-modal`,
`.gf-pagination` and `.gf-rating`.
Contributors MUST inspect and reuse the appropriate primitive and its variants
before creating a new selector. New consumers MUST NOT redefine its core
appearance through local overrides; extend the owning primitive deliberately
when a reusable variant is needed.

A product-level UI composition SHOULD remain in `styles/components/`, even when
reused across routes. Navigation, footer and shell are compound owners, not
primitives. Their product-local semantics such as `--gf-nav-*` and `--gf-footer-*`
MAY stay with those owners; reuse alone does not promote them to global tokens.
`primitives/modal.less` owns generic `.gf-modal*` appearance, shared by Preferences
and NSFW confirmation. `components/preferences.less` owns Preferences overrides,
tabs, source selectors, carousel/arrows and the product-only `preferences-toggle`.
Contributors MUST NOT reintroduce `gf-modal__toggle` or promote a control with
one product-specific consumer into a speculative generic primitive.

Preferences input/toggle idle, focus and active semantics MUST stay local under
`--gf-preferences-*`, declared only at `.gf-preferences-modal` and
`html.dark .gf-preferences-modal` in `components/preferences.less`. Theme selects
token values; control selectors select state. Compound-local approval MUST match
the exact file, root selectors and prefix; it is never a whole-file exemption.
Private `.preferences-pages`/`.preferences-page` and Background/Hero/ResourceRoute
editor structure MUST stay scoped, not accumulate in the shared compound owner.

Moving a primitive MUST preserve its selectors, declarations, token values and
variants. A structural move is not authorization to clean up typography,
hover/focus behavior or other historical appearance.

Appearance reuse belongs in a **CSS primitive**. Reused behavior plus
accessibility (state, ARIA, keyboard navigation and focus management) belongs
in a **Vue primitive**: tabs, dialogs, selects, carousels, popovers or segmented
controls when such reuse is demonstrated. Contributors MUST NOT add one-line
wrapper components without behavior/accessibility value.

When a pattern appears in two different business domains, contributors SHOULD
evaluate promotion into the shared foundation. Domain-specific meaning MAY stay
in its domain. Contributors MUST NOT create speculative fields/tabs/notices or
other primitives merely because they appear in the future programme.

## Scoped styles, themes and raw values

`<style scoped>` MAY own component-private structure. It MUST NOT be a place to
invent a new local theme or replace existing primitive appearance. Repeated
appearance SHOULD move upward as reuse warrants it: scoped → domain → primitive.

Theme semantics MUST use CSS variables/semantic tokens, preserving meaning
between light and dark. Contributors MUST reuse the existing `html.dark` theme
entry. They MUST NOT create independent page-dark systems or Tailwind dark
appearance when existing semantics suffice. Legacy theme islands are debt.

Raw colors (`#hex`, `rgb()`, `rgba()`), literal shadows/radii and visual durations
SHOULD live in approved token declarations, not ordinary component/page
selectors, inline styles or script-generated appearance. The same ownership
applies to typography scales. Approved locations are global declarations,
primitive-local declarations, approved compound-local declarations, justified
domain declarations, and precise documented exceptions. The narrow
P0 color count below does not authorize uncatalogued raw typography or geometry
used as control appearance.

## Debt and exceptions

Historical debt MAY remain during staged migration. New debt MUST NOT increase;
after removal, the baseline MUST decrease in the same change. The manifest is
current state, not a design example or a pool of credits for new violations.

`frontend-style-debt.json` uses `schema_version: 1`, a `baseline` object containing
exactly the six P0 rule families below, and an `exceptions` array. Each baseline
maps Nav Web-relative POSIX paths to positive integer occurrence counts. Omitted
file/rule pairs have budget **0**, including new files. Do not store line numbers.
Contributors MUST NOT offset one file's increase with another file's decrease,
inflate counts to pass checks, or evade measurement by moving/renaming debt.

Moving a debt-bearing source file is a **policy migration**, not a mechanical
rename: the old path becomes stale and the new path has budget zero. Future moves
MUST explicitly migrate debt ownership under review as a policy change, proving
no new debt is granted. Neither a rename nor `style:policy:update` transfers a
budget automatically. P2.2's six selected files have no per-file style debt to
transfer, so their move MUST leave the manifest and all six totals unchanged.
P2.3 retires the old `components/modal.less` raw-color budget of 29 through dead
cascade removal and semantic ownership: raw debt decreases from 924 to 895.
Neither new stylesheet receives a debt budget; other rule/file counts stay intact.

An exception intentionally retains a specific pattern temporarily; baseline
debt merely records historical noncompliance. Every exception MUST specify all
of `path`, `rule`, `issue`, `reason`, `remove_when`, using an exact path and a
specific rule. Reasons MUST explain why the existing pattern is necessary and
the removal condition MUST be actionable. Unexplained ignore lists and permanent
directory wildcards are prohibited. Existing visual-guard allowlists MUST NOT
be silently promoted to exceptions. P0 grants no exceptions automatically.

#108 Insights and #109 Site Detail were **migration exclusions** from #124
P4–P6, not style exceptions. #109 P7 now closes Site-specific debt under its own
scope; #108 remains separately owned. Any remaining historical debt MUST still
be measured, and new code there MUST follow this contract.
Any precise exception tied to those redesigns MUST be removed when its stated
replacement lands; issue membership is not a blanket exemption.

`shared-primitive-override` is a normative concern above, but has no numeric P0
baseline: reliable low-false-positive detection belongs to later review.

## P0 audit definition

The initial inventory audits the unchanged application tree at
`f5576f6bfe1981a7c1b3a2d58f0287bf13d2a15a` on 2026-09-19. This section freezes
the one-time counting semantics so P1 can reproduce them; no policy scanner is
installed by P0. Paths below are relative to the frontend workspace (`apps/cn/nav-web`
at that audited SHA; now `apps/cn/web`).

### Source scope and occurrence unit

Inspect Git-tracked `app/**/*.vue`, `*.ts`, `*.js`, `*.css` and `*.less`.
Exclude `tests`, `__tests__`, `fixtures`, `generated` directories, TypeScript
`*.test.ts`, `*.spec.ts`, `*.d.ts`, and the encoded geographic data asset
`app/assets/js/china.js`. The P0 snapshot has 245 in-scope files. Static SVG,
JSON, raster/icon assets, tooling/config, `server/`, dependencies and generated
Nuxt/output directories are not style source for this inventory.
`app/components/experimental/ambient/` IS in scope; it is not the repository's
excluded root `experimental/` tree. #108/#109 receive no scope exclusion.

Count authored source occurrences, not unique spellings, matched lines,
compiled selectors or runtime repetitions. Two copies of a class count twice;
one class in a loop counts once. Ignore comments and non-style content such as
validation messages, slot names and selectors merely mentioning utility classes.
Parse Vue template/style/script regions and CSS/Less declarations separately.

For Tailwind, inspect static `class` and `*-class` attributes (including transition
class props), static choices in bound class expressions, script literals that
actually flow into those bindings, and `@apply` lists. Each branch is authored
source and counts separately. The P0 script-dataflow review found class literals
in `Footer.vue` (`hoverClass`), `site/SiteOverview.vue` (`tagClass`), and
`site/SitePerformance.vue` (`getColor`, `iconColor`), all under `app/components/`.
These are evidence anchors, not permanent scanner allowlists. Do not count CSS
selector references such as `:deep(.rounded-xl)` again as Tailwind use.

### Six rule families

| Rule | P0 count |
| --- | --- |
| `tailwind-appearance` | Each actual Tailwind visual utility occurrence under the classification below |
| `tailwind-arbitrary-appearance` | Each appearance occurrence whose base utility contains an arbitrary `[…]` value/property or `(…)` variable shorthand |
| `raw-visual-value` | Each static hex or numeric `rgb()`/`rgba()` color literal outside the approved declarations below |
| `important` | Each CSS `!important` annotation, including optional whitespace after `!` |
| `deep-selector` | Each authored `:deep(` selector entry, including optional whitespace before `(` |
| `legacy-dark-entry` | Each complete legacy class-name occurrence listed below, plus each style `:global(.dark…)` entry |

Arbitrary visual utility occurrences are a **subset** of appearance and appear
in both rule budgets. They MUST NOT be counted a third time as raw color values.
Totals across rules are rule hits, not disjoint violations.
Bracket-only CSS properties follow the same ownership: `[color:#fff]` is visual,
`[width:10px]` is structural, and an arbitrary transform is visual under the
interaction variants below. None occurred in the initial P0 snapshot.

Tailwind candidates are validated against the lockfile's Tailwind v4 default
theme/utility compiler before classification, so semantic names such as
`blur-wrapper` do not become false positives. Strip variant prefixes at colons
outside brackets/parentheses and leading/trailing utility importance markers;
classify the base utility, keeping the original variant chain as one occurrence.
The P0 appearance families are:

- `rounded`, `shadow`, `inset-shadow`, `drop-shadow`, `text-shadow`, `ring`,
  `inset-ring`, `opacity`, `leading`, `tracking`, `font`, and their utilities.
- `blur`, `brightness`, `contrast`, `grayscale`, `hue-rotate`, `invert`,
  `saturate`, `sepia`, and their `backdrop-*` equivalents, including backdrop opacity.
- `transition`, `duration`, `delay`, `ease`, `animate`; `italic`, `not-italic`,
  `antialiased`, `subpixel-antialiased`, `uppercase`, `lowercase`, `capitalize`,
  `normal-case`, `underline`, `overline`, `line-through`, `no-underline`;
  `decoration`, `underline-offset`, `accent`, `caret`, `fill`, `stroke` utilities.
- `text-*`, excluding alignment (`left`, `center`, `right`, `justify`, `start`,
  `end`) and wrapping/overflow (`wrap`, `nowrap`, `balance`, `pretty`, `ellipsis`, `clip`).
- `bg-*`, excluding attachment (`fixed`, `local`, `scroll`), `clip-*`, `origin-*`,
  sizing (`auto`, `cover`, `contain`), cardinal positioning (`center`, `top`,
  `right`, `bottom`, `left`), `repeat`/`repeat-*` and `no-repeat`.
- `from-*`, `via-*`, `to-*` color stops: standard Tailwind color palette names,
  `black`, `white`, `transparent`, `current`, `inherit`, or arbitrary values.
  Numeric stop positions alone are not P0 color debt.
- `border`, `divide`, `outline` utilities, excluding `border-collapse`,
  `border-separate`, `border-spacing*`, `divide-x-reverse`, `divide-y-reverse`.
- `scale`, `translate`, `rotate`, `skew`, `transform` utilities only under
  `hover`, `focus`, `focus-visible`, `focus-within`, `active`, including group/peer
  variants and named group/peer variants. Ordinary placement transforms are not
  counted here; `hover:scale-[1.05]` is visual in both Tailwind rules.

All other structure is outside these initial counts. Context-sensitive issues
such as a button's internal spacing still require contract review; absence of
a numeric detector is not permission to violate ownership.

Raw colors are `#RGB`, `#RGBA`, `#RRGGBB`, `#RRGGBBAA`, or a fully static numeric
`rgb()`/`rgba()` call, including comma/space/slash numeric syntax. Each function
counts once; a gradient or shadow with multiple color literals counts each.
Inspect CSS/Less declarations, Vue style blocks, literal inline style/SVG paint
attributes, bound style literals and script-generated visual settings/styles.
P0 manually verified script colors in `BackgroundPreferencesEditor.vue`
(`colors`, `defaults`), `game/detail/insights/GamePlayerTrend.vue` and
`GamePriceHistory.vue` (`renderChart`), and `site/SitePerformance.vue`
(`updateChart`), under `app/components/`. Charts are not a blanket exemption.
Color validation messages such as "#9c846a" are not rendered styles.

P0 does not quantify named/HSL colors, dynamically interpolated colors,
token/channel-derived colors, literal radii, shadow geometry, durations or
typography. Their ownership rules still apply. This is a high-confidence color
baseline, not a claim that all raw visual debt has been measured.

Count `!important` and deep entries in authored styles, including CSS emitted by
script strings: the `<noscript>` style in `ErrorExperience.vue` counts. Tailwind's
`!` marker is handled as a class modifier, not another CSS `!important` annotation.
Historical visual-guard deep allowlists do not remove those occurrences from debt.

The legacy list, inherited from the retired visual guard, is enforced by
`scripts/style-policy/css.mjs`: `games-page--dark`,
`search-results--dark`, `is-dark-theme`, `spotlight-panels--dark`,
`about-page--dark`, `legal-page--dark`, `updates-page--dark`,
`nav-home-page--dark`, `gf-static-page--dark`, `lottery-page--dark`.
Match complete names in classes, selectors or class-bearing script literals,
not substrings/comments. Include the guard's prohibited `:global(.dark…)` form
in this family. Canonical `html.dark`/`:global(html.dark …)` is not legacy debt.

### Current approved raw-color declaration locations

The P0/P1 snapshot also approved `app/assets/css/main.css` at `:root` and
`html.dark` for the `--gf-page-` prefix. P2.1 moved those declarations unchanged
into `tokens.less` and removed that obsolete approval. The table below is the
current policy; this ownership migration and removal of the approved dead
`--gf-bg-grid-line` declarations leave all six debt budgets unchanged.

P2.2 moved Rating's approved owner from
`app/assets/styles/components/rating.less` to
`app/assets/styles/primitives/rating.less`, retaining the exact selectors and
`--gf-rating-` prefix. The old path is historical only and is no longer approved.

Only declarations matching **all three** columns are excluded from
`raw-visual-value`. Match the exact rule selector (normalize whitespace and comma
spacing) and property prefix; do not exempt a file, descendant rule or ordinary
property such as `background`. Grouped lottery roots mean the listed selector
group and its corresponding all-dark group.

| File | Exact declaration selector(s) | Custom-property prefix |
| --- | --- | --- |
| `app/assets/styles/tokens.less` | `:root`, `html.dark` | `--gf-` |
| `app/assets/styles/components/nav.less` | `.gf-nav`, `html.dark .gf-nav`, `.mobile-bottom-tabs`, `html.dark .mobile-bottom-tabs` | `--gf-nav-` |
| `app/assets/styles/components/footer.less` | `.gf-footer-shell`, `html.dark .gf-footer-shell` | `--gf-footer-` |
| `app/assets/styles/components/preferences.less` | `.gf-preferences-modal`, `html.dark .gf-preferences-modal` | `--gf-preferences-` |
| `app/assets/styles/components/error.less` | `.error-page`, `html.dark .error-page` | `--gf-error-` |
| `app/assets/styles/components/page-scroll-dock.less` | `.page-scroll-dock` | `--gf-scroll-dock-` |
| `app/assets/styles/primitives/rating.less` | `.gf-rating`, `html.dark .gf-rating` | `--gf-rating-` |
| `app/assets/styles/pages/games.less` | `.games-page`, `html.dark .games-page` | `--games-` |
| `app/assets/styles/pages/games-search.less` | `.games-search-page`, `html.dark .games-search-page` | `--games-search-` |
| `app/assets/styles/pages/nav.less` | `.nav-home-page`, `html.dark .nav-home-page` | `--nav-home-` |
| `app/assets/styles/pages/nav.less` | `.nav-site-card`, `html.dark .nav-site-card` | `--nav-home-card-hover-` |
| `app/assets/styles/pages/nav.less` | `.nav-group-toggle`, `html.dark .nav-group-toggle` | `--nav-home-group-toggle-` |
| `app/assets/styles/pages/nav.less` | `.site-popover`, `html.dark .site-popover` | `--nav-home-popover-` |
| `app/assets/styles/pages/nav.less` | `.group-popover`, `html.dark .group-popover` | `--nav-home-group-popover-` |
| `app/assets/styles/pages/nav.less` | `.nav-transition-bar__author`, `html.dark .nav-transition-bar__author` | `--nav-home-transition-author-` |
| `app/assets/styles/pages/updates.less` | `.updates-page`, `html.dark .updates-page` | `--updates-` |
| `app/assets/styles/pages/lottery.less` | `.lottery-page, .lottery-activation-page, .lottery-modal`; the same group with `html.dark ` before each root | `--lottery-` |
| `app/assets/styles/pages/static.less` | `.about-page`, `html.dark .about-page` | `--about-` |
| `app/assets/styles/pages/static.less` | `.legal-page`, `html.dark .legal-page` | `--legal-` |
| `app/assets/styles/pages/insights/foundation.less` | `.insights-page` | `--insights-` |

This recognizes existing declaration locations only. It does not approve all
current token hierarchy, aliasing or annotation, nor establish another token
level; further consolidation belongs to later phases. `foundation.less` has structural
values/aliases and excludes no raw colors. Ordinary selectors in every listed
file remain measured. Game-detail-specific text overrides, datepicker `--dp-*`
overrides and `SiteDetailPage.vue`'s local theme were not approved declaration
layers in this P0 snapshot and were counted as raw-color debt. The Site theme is
retired under #109; these are historical evidence anchors. P4.5.3 moves Error and Dock declarations into
the exact component owners above; ordinary properties there still count as debt.

P1 MUST compare its initial detector results against this same source snapshot,
investigate discrepancies and document corrections rather than raising budgets
to hide differences. It MUST NOT turn these P0 evidence anchors into exclusions
for newly added files or script-generated styles.

## Verification responsibilities and phase boundary

| Responsibility | Current owner / phase |
| --- | --- |
| TypeScript/Vue typing | Existing Nuxt typecheck |
| JS/TS/Vue engineering rules | Nuxt-compatible ESLint flat config; official bulk suppressions track historical findings |
| CSS/Less correctness and hygiene | Conservative Stylelint recommended rules with CSS/Less/Vue parsers |
| GoFurry-specific ownership and per-file debt budget | Parser-backed `style:policy`; Node built-in tests verify the guard |
| Pure TS/domain/utility logic | Vitest `unit` project, Node environment, `tests/unit/*.test.ts` |
| Nuxt runtime/composables | Vitest `nuxt` project with `@nuxt/test-utils`, happy-dom, `tests/nuxt/*.nuxt.test.ts` |
| Repository/source/config/semantic contracts | Existing Node Insights/SEO Contract Guards, outside Vitest |
| Migrated browser behavior, SSR, hydration and historical regressions | Playwright Test in `tests/browser`: Shared/Nav/Games surfaces, Routing/Hero/Preferences, Insights, SEO HTTP/SSR and background IndexedDB |
| Stable pixel appearance | `playwright.visual.config.ts` / `test:visual`; pinned sentinel and accepted surface golden comparisons |
| Broad route/theme/overflow checks formerly in visual guard | Domain Browser/Visual cases plus P7.2 locale/theme gaps; explicit retirement mapping in the closure record |
| Performance budget | Existing `perf:guard`, separate from visual correctness |
| Real external services | Explicit development acceptance, not a default PR gate |

P0 introduced no runners, dependencies or CI gates. P1's local/CI sequence is
`pnpm install --frozen-lockfile`, `pnpm run lint`, `pnpm run stylelint`, `pnpm run style:policy:test`,
`pnpm run style:policy`, `pnpm run typecheck`, `pnpm run insights:semantics`,
`pnpm run seo:recovery:test`, `pnpm run build`, from `apps/cn/web`.
For runtime changes, run relevant existing focused scripts from `package.json`
as well. A skipped external acceptance test is not a pass.

The old visual guard is retired after P7 responsibility mapping; do not recreate it.
P1's detector reproduces the P0 baseline on the unchanged app tree. P2 owns token/primitive organization;
P3 owns the testing foundation; P4+ owns staged, visually equivalent migrations.
Each phase MUST finish in a stable, independently deployable state.

P7.1 consolidates the eight remaining background/SEO/Insights runtime runners
into this single Browser Gate. Worker fixtures own only stable Nitro/upstream
transport; each test owns state, request/error ledgers and releasable gates.
Fault allowances MUST match injected request instances/counts, never generic
hydration or console-error suppression. Real debounce/timeout behavior remains.
Search-to-Detail acceptance waits for the exact mounted view POST response before
comparing ledgers. [Testing guidance](../docs/frontend/testing.md#p71-runtime-consolidation)
records the coverage mapping. This phase changes no production code, style debt
or Visual baseline. P7.2 completes the separate legacy visual-guard retirement.

P3.1 uses Vitest `projects` with separate `test:unit` and `test:nuxt` commands;
`pnpm test` runs both. CI MUST execute separate Unit tests and Nuxt tests steps
before typecheck, retained Insights/SEO guards and build. Choose the lowest-cost
environment that faithfully represents the tested behavior. Browser-like globals
alone do not require Nuxt; pure tests SHOULD use Vitest-controlled stubs/cleanup.

Nuxt-dependent cases MUST use the real Nuxt runtime and reset cookies/useState
between cases using supported APIs. Mock only business injection boundaries,
not Nuxt's state/cookie/app context. Source-transpile/data-URL import hacks and
handwritten fake Nuxt runtimes MUST NOT replace ordinary imports/runtime tests.
Keep `@nuxt/test-utils/module` out of production Nuxt config. Do not refactor
production behavior merely to accommodate tests. The first four legacy suites
are replaced, not duplicated; [testing guidance](../docs/frontend/testing.md)
records their assertion ownership. Style-policy's Node runner, Contract Guards,
Browser tests and external acceptance retain independent scope.

P3.2.1 establishes a Chromium-only Playwright Test gate against the production
Nitro build, never `nuxt dev`. CI MUST build successfully before running
Browser tests; retries are zero and each CI job has one worker. #134 separates
daily Smoke from Full Regression; Full MUST run all four explicitly inventoried
groups without reducing coverage or enabling retries/fullyParallel.
Domain fixtures own worker-scoped production servers, close them at teardown,
and reset mutable state for every case. Playwright owns per-test browser contexts
and pages; do not introduce a global `webServer` or order-dependent serial suites.
Capture/assert browser exceptions and hydration errors explicitly. Retain traces
and screenshots only on failure, with video off; CI uploads failure diagnostics.
Game Detail and Resource Routing have migrated, including Preferences, Managed
and Steam resource snapshots/fallback, and legacy Steam recommendation-only
migration. Routing probe latency/failure/block/gate state MUST be test-scoped;
the worker owns only stable production Nitro/local API resources. Managed/Steam
requests MUST use deterministic local responses, never real CDN services.
P3.2.3 also migrates Hero lifecycle/Local, Preferences foundation, Fixed/BigInt,
Catalog and staged handoff. The two Hero domain fixtures MUST remain separate;
each test owns fresh catalogs, request evidence, failures and independent gates.
Teardown MUST release API/image/IndexedDB gates and clear the active scenario;
the worker's API resolver MUST reject access without an active scenario.
Only the Hero-specific assertion may acknowledge the existing mobile homepage
Footer hydration debt: width below 768, exactly one mismatch, no SSR footer,
exactly one client footer, retained SSR Hero node, and no other browser errors.
Its default route remains `/`; P7.2 explicitly opts English Home into exact `/en`.
Callers MUST preserve raw initial evidence and fail on subsequent errors. See
[RUNTIME-FOOTER-01](../docs/acceptance/issue-124-frontend-engineering-closure.md#runtime-footer-01)
for the owned follow-up and removal condition; no generic capture exemption exists.
Generic error capture MUST NOT ignore hydration. Same-run clipped paint Buffer
and computed-style equality MAY remain runtime invariants; success screenshots,
computed audit JSON and golden baselines MUST NOT be introduced in P3.2.3.
P7.1 migrated the remaining Insights/background/SEO runtime runners. `pnpm test`
remains Vitest-only; Visual keeps its separate pinned runner.

P0 MUST NOT change production Vue, CSS/Less, runtime behavior, package/lockfiles,
CI, dependency versions or style directories, and MUST NOT implement #108/#109.
It neither replaces Nuxt/Tailwind/Less nor introduces a new UI framework.

### P3.3 Visual environment and baseline governance

Functional `playwright.config.ts` MUST exclude `tests/browser/visual/**` and keep
the existing Smoke/Regression gate unchanged. `playwright.visual.config.ts` owns
Visual tests with Chromium only, one worker, zero retries and a 60-second timeout.
The environment MUST fix headless mode, 1440×900, zh-CN, UTC, DPR 1, light default
and reduced motion. Failure-only trace/screenshots and no video use independent
`playwright-visual-report/` and `visual-test-results/` directories.

Authoritative Visual CI MUST use the official Playwright image matching the
package version, pinned by a verified immutable digest, with `--ipc=host` and
Node 24. The exact identity and local commands live in
[testing guidance](../docs/frontend/testing.md#visual-runner-and-pinned-environment-p331).
The #134 CI tiers supersede the earlier every-push Browser/Visual/Image topology:

- Fast (`checks.yml`) runs for selected dev pushes and PRs: frozen install, lint,
  stylelint, policy tests/scan, Unit/Nuxt, typecheck, semantic/SEO guards, production
  build and all seven Smoke owners in one pinned job. The stable `nav-web` gate
  depends only on change detection and `nav-web-fast`; it MUST report their failure.
- Full (`nav-web-full.yml`) runs on main push, manual selected ref and nightly dev.
  Build once, archive `.output` preserving symlinks/permissions, then restore it
  inside the identical pinned image in four fixed groups. The explicit lists in
  `.github/scripts/nav-web-regression-groups.mjs` MUST cover every regression spec
  exactly once and fail on stale/missing/duplicate membership. Resolve scheduled
  dev HEAD once and pin all consumer checkouts/artifacts to that SHA. No numerical
  sharding or dynamic bin-packing. Its final gate requires every group and build;
  deployment image verification is required on main/manual and skipped on nightly.
- Visual (`nav-web-visual.yml`) is manual-only, building and comparing all existing
  specs in one pinned job with `GOFURRY_VISUAL_ENV=pinned`. It MUST NOT run on daily
  pushes/PRs/main/nightly, install browsers or update snapshots.

All hosted runners use `ubuntu-24.04`; Playwright image/package identity is unchanged.
Docs-only selection remains cheap; workflow/script edits remain conservative.
Failure reports/results retain seven days; only Full's one-day production build
artifact is shared between jobs. Fast's 3–5 minute feedback is a measured target,
not permission to skip checks. Full/Visual remote acceptance remains separate
from Fast and from maintainer visual approval. See [CI tiers](../docs/frontend/testing.md#ci-tiers-134).

Baseline update is an explicit visual-change review action, not a test-fix
command. Agents MUST NOT update baselines just to make CI green. Updates require
explicit maintainer/user approval or a task explicitly authorizing the visual
migration, and MUST use `test:visual:update` in the pinned environment. Its guard
requires Linux, Node 24 and `GOFURRY_VISUAL_ENV=pinned`; the marker is an operator
assertion, not approval or an image fingerprint. Unpinned `test:visual` runs are
diagnostic only. Unexpected differences require investigation and implementation
fixes followed by comparison, not automatic acceptance.

Baselines MUST live in tracked `tests/browser/visual/__snapshots__/`,
using deterministic test-file and explicit snapshot-name paths. No global
`maxDiffPixels` / `maxDiffPixelRatio` or broad threshold widening is allowed;
any future tolerance needs a small, local, justified exception. Playwright
packages, Docker tag/digest, browser revision and visual baselines MUST be reviewed
as one atomic upgrade unit. P3.3.1 MUST create no golden PNG, screenshot assertion,
production test route or UI fixture. The sentinel launches real Chromium and
checks the environment without taking a screenshot; real visual contracts start
in P3.3.2. Keep direct `playwright` for the retained performance/cloud tools.

P3.3.2's `ui-foundation.spec.ts` owns stable shared primitive appearance through
exactly four locator baselines (light/dark × desktop/mobile). Its test-only
markup MUST consume production Nitro CSS loaded from `/about`, preserve the
production head and exclude product JavaScript. Fixture CSS MAY own structure;
it MUST NOT redefine control appearance. Only the canvas MAY set background/text
using production page/text tokens. Generic Modal MUST remain outside the isolated
Preferences Toggle token scope. Verify tokens, scope, fonts, focus and overflow
before capture; do not mask instability or widen tolerances. Approved baseline
creation MUST pass two consecutive pinned comparisons and receive maintainer
review of all four images. Real Preferences/backdrop composition belongs to
P3.3.3 and business surfaces to P4+, not this Foundation fixture.

P3.3.3's real Preferences visual contract MUST exercise production Nuxt/Vue,
NavBar, Teleport, backdrop and child components through the existing Hero
Preferences fixture. Optional visual seeds MUST preserve functional defaults.
Tests MAY hide unrelated underlying page content and use the semantic canvas;
they MUST NOT fake or restyle Preferences. Theme MUST use the production
localStorage/Theme Store path. Routing time and fresh diagnostics MUST be fixed
only in the test context, with no automatic probe traffic before capture.
Real tab selection, finite animations/fonts, backdrop/filter/containment and
active-page overflow MUST settle semantically, without arbitrary sleeps, masks
or tolerance widening. The eight-case matrix is Home light/dark desktop/mobile,
Background light desktop/mobile and Routing light desktop/mobile. It MUST remain
separate from functional regressions and business surfaces; expansion requires
explicit review. Creation requires pinned generation, two consecutive comparisons
and maintainer review of the eight new images; the four Foundation images remain
unchanged. Real product defects MUST be reported without expanding this phase
into production UI changes.

P4.3.1's Static/Legal Visual baselines MUST exercise real production Nitro,
Nuxt SSR/hydration, Theme Store, default layout and `PublicPageBackground`.
Layout owns the canvas; the Static root MUST remain computed-transparent.
Fixtures MAY hide only PageScrollDock and MobileBottomTabBar, never restyle
Static/Legal content. The six viewport baselines cover About light/dark at
desktop/mobile and Terms light/dark on mobile. Readiness MUST use load plus
semantic image/animation/font waits, with zero external/upstream requests and
the current computed color-transition contract. Creation requires pinned
generation, two consecutive comparisons and maintainer review. Appearance
migrations MUST pass these accepted images without snapshot updates or changes
to layout-owned canvas semantics. All eighteen baselines MUST stay intact.

Updates business-surface baselines MUST use the real production SSR/hydration
path with deterministic ready-state `/nav/updates` data. Functional coverage
owns year expansion and load-more persistence; the shared fixture MUST prove
exactly one SSR API call, payload reuse and zero external/unrelated requests.
The four viewport baselines retain Theme Store, PublicPageBackground and real
reduced motion. Creation requires pinned generation, two consecutive comparisons
and maintainer review. Updates selector normalization MUST pass accepted goldens
without snapshot updates; no selector or token cleanup accompanies P4.4.1.

Error Experience baselines MUST exercise the real Nuxt missing-route error
boundary and its own canvas. Hydrated coverage owns production theme/artwork
readiness and normal-motion keyboard focus visibility; a genuinely JS-disabled,
normal-motion regression MUST preserve the readable default Light fallback.
Appearance migration MUST pass the four accepted Error locator goldens and both
behavior contracts without snapshot updates. Initial creation requires pinned
generation, two consecutive comparisons and maintainer review.

PageScrollDock runtime coverage MUST use the real document scroller on a
production page. It owns Desktop render/visibility, displayed/accessibly labelled
progress, quarter-of-total-distance scrolling and Mobile unmount. The two Visual
baselines own normal/hover appearance at 50%, including surrounding shadow and
canvas. Appearance migration MUST pass these contracts without snapshot updates;
unused custom-scroller props are outside the current product contract.

P4.5.3 assigns Error appearance to `components/error.less` with local
`--gf-error-*` semantics and Dock appearance to `components/page-scroll-dock.less`
with local `--gf-scroll-dock-*` semantics. Private geometry/placement MUST remain
in their SFC scoped styles. These single-consumer semantics MUST NOT be promoted
to global `tokens.less`. `--gf-error-step-delay` and `--gf-scroll-dock-progress`
are Vue-supplied runtime channels, not static tokens. The existing Browser and
Visual contracts MUST pass unchanged, without snapshot updates.

Footer/Shell Visual contracts MUST use real production `/terms`, SSR/hydration,
Theme Store and default PublicPageBackground. The two Desktop light/dark goldens
own the stable first/second columns and Footer canvas; DOM-derived clips MUST
exclude dynamic current-year meta without mocking time or masking. Computed
assertions own heading/icon/meta/link appearance, four accessible-name social
hover colors and the exact App Shell color-transition property set/timing.
Creation requires pinned generation, two consecutive comparisons and maintainer
review. Footer/Shell migration MUST pass these contracts without snapshot updates.

P4.6.2 completes Footer appearance in `components/footer.less` and its local
`--gf-footer-*` semantics; structural composition remains Tailwind. App Shell's
full color transition belongs to `components/shell.less`. Global scrollbar colors
use foundation-level `--gf-scrollbar-*` semantics. P4 completion means no
unassigned Stable/Common appearance debt remains: every remaining baseline entry
has an explicit [later owner](../docs/frontend/design-system.md#p4-exit-ownership).
That baseline MUST NOT authorize opportunistic or early cross-phase migration.
Dark Footer link-hover aliases its normal link token to preserve the prior
computed cascade; brand glows retain the same values in both themes.

Nav Shell contracts own ordinary and homepage-overlay NavBar states, Desktop
theme/language integration, Mobile Menu behavior and MobileBottomTabBar scroll,
active-state and breakpoint semantics. P5.1.2 places NavBar and MobileBottomTabBar
appearance together in `components/nav.less`, using the exact Navigation roots
and `--gf-nav-*` declaration prefix above. Vue MUST retain behavior and private
runtime geometry; the physical `common/` location does not create a separate
BottomTab appearance owner. Changes MUST pass the accepted Browser/Visual
contracts without snapshot updates. Hero lifecycle, SearchBox, QuickAccess and
Nav content retain separate ownership.
The P5.1.1 audit found existing Home side effects absent from its initial plan:
null saying triggers one saying request, mounted content creates a weather iframe,
and Mobile Home has the already-guarded Footer hydration mismatch. The approved
fixture exceptions MUST remain exact, locally fulfilled and separately recorded;
reuse the narrow existing hydration evidence check, never broaden generic error
suppression. Terms remains quiet except for one separately recorded empty pattern
catalog requested only after the real BottomTab Mode action mounts Preferences.
All other requests/errors MUST remain zero.

P5.2.1 adds a separate live Home Header contract: two Functional cases own
Search debounce/keyboard/popup/reveal lock, theme-independent computed appearance,
and Quick Sites validation/add/delete/storage. Four clipped Visual cases own
Desktop Header, Desktop/Mobile suggestions and Desktop Quick Sites validation.
The shared fixture keeps real SSR/hydration/Hero rendering, exact local asset
boundaries and one SSR Home request; it does not supersede Hero lifecycle or
Nav Shell contracts. The chip transition's actual Less cascade (background,
box-shadow and color, 500ms each) takes precedence over utility-name inference.
Issue #123 separates Suggestion provider from Search destination. SearchBox calls
`GET /api/v2/nav/search/suggestions?q=...` for non-empty Search-category input,
regardless of platform. Schema 2 has no engine; ready renders suggestions, empty
renders the existing empty state, unavailable/network/429 silently closes the
optional dropdown without retry. Existing doSearch/platform URLs remain unchanged.
Keep 600ms debounce, AbortController and stale-response protection. Composition
updates local display only; compositionend schedules final text once. While the
composition lifecycle or native isComposing flag is active, Enter/Arrow keys/Escape
must not select, navigate or prevent IME confirmation. Render provider text with
interpolation and highlighted segments, never v-html or manually escaped HTML.

P5.2.2 keeps Header, Search, QuickAccess and Quick Sites appearance in
`pages/nav.less`, with `--nav-home-*` declarations on `.nav-home-page`; do not split
out a Header stylesheet. These Hero-backed semantics remain theme-independent,
including mobile states. `SiteIconStrip` and its styles were deleted after a
zero-consumer audit. The accepted Header contracts and all 42 PNGs are unchanged.
Spotlight, ToolDock, TransitionBar, Cards and Popovers remain P5.3-owned; this
Header migration does not authorize their cleanup.
Mobile Home reuses the existing `assertHeroHydration` strict Footer mismatch
contract with SSR Hero retention evidence. Preserve raw errors and reject all
later errors; Desktop retains zero browser errors. This is not a general exception.

P5.3.1's independent `nav-revealed-content` fixture/specs MUST use real production
Home SSR/hydration and wheel reveal. Four runtime cases own Card typography,
Group/Site popovers (including geometric top placement and ping), fixed-time
TransitionBar/weather/author, ToolDock directory/search/cache/popup, Spotlight
paging/visited/view side effects, responsive panel counts and SFW/NSFW filtering.
The actual description cascade is 12px / 16.2px (`line-height: 1.35`), overriding
the plan's inferred 16px; accepted computed appearance remains authority.
Eight Light/Dark clips own Core, SitePopover, ToolDock Search and Spotlight.
Home MUST be SSR-only exactly once per scenario; other requests are limited to
exact fixture assets/weather and explicitly exercised directory/view/popup paths.
No production/debt changes are authorized by this contract stage. Keep
`.nav-content-loading` for P5.3.4 re-audit and the dynamic live
`.nav-tool-button--search`; Site Groups/games-page coupling remains P5.4-owned.
The eight initial goldens require pinned generation, two full comparisons and
maintainer review before P5.3.2. Existing 42 snapshots MUST remain unchanged.

P5.3.2 Core appearance stays in `pages/nav.less`; NavSiteGrid's accepted title
16/24/500 and description 12/16.2 are preserved. Cards/Groups/Popovers and
TransitionBar retire 46 raw occurrences. ToolDock 42, Spotlight 51 and the five
games-page occurrences remain outside this migration. P5.3.1 tests and all 50
accepted PNGs are frozen.
Approved inheritance corrections add only exact local declaration owners:
SitePopover (`--nav-home-popover-*`), GroupPopover (`--nav-home-group-popover-*`)
and the TransitionBar author (`--nav-home-transition-author-*`) on their body
Teleport roots; shared Card hover (`--nav-home-card-hover-*`) and group toggle
(`--nav-home-group-toggle-*`) on their consumer roots, including exact Dark roots.
SitePopover MUST retain its effective transparent background/zero border.
Do not activate unreachable page-root surface tokens or alter Site Groups'
existing cascade. Other Nav Home semantics retain their existing page owner;
detectors and exception rules are unchanged.

P6.1.3 Games Home closure MUST preserve the P6.1.1 core fixture and four Home
goldens. Ten additional regressions protect News empty/single/multiple states,
locale/carousel/resize/popup, populated Reviews SSR time/content/appearance and
seven-breakpoint group clipping in Light zh / Dark en. The isolated SSR/browser
clock is opt-in, fixes only current time and MUST NOT alter production or timers.
Exact requests and raw errors remain accounted for. Six additional pinned Page
clips cover News and Reviews; historical nonempty Reviews equivalence precedes
their acceptance. The equivalent group-layout cases retire only their old runner
and alias, not other legacy smoke ownership. Production and style debt remain
unchanged; P6.1.4 requires maintainer review of these six new images first.

P6.1.4 completes direct Games Home appearance ownership. News typography and its
ten `--games-home-news-*` roles belong to `pages/games.less` under the existing
`.games-page` / `html.dark .games-page` roots. Preserve accepted computed values,
including line-height precision, rather than replacing them with rounded values
or inferred utility intent. Carousel behavior, fixtures and all 64 accepted PNGs
remain unchanged. This completion excludes shared ReviewDialog (P6.2),
SidebarSearch (P6.3), Lottery (P6.4), Detail/Common (P6.5) and legacy Games root
cleanup (P6.6); remaining debt is not authorization to cross those boundaries.

P6.2.1 protects shared GameReviewDialog through real Home, mounted Search and
SSR Detail consumers. Its body Teleport MUST stay outside page-root ownership.
Eight runtime cases cover settled reset, required/score validation, trim/decimal
normalization, gated pending, success/rejection/503 and Mobile focus/bounds.
Submit requests MUST be exact local fulfillments, with raw evidence and no real
backend writes. Successful submission currently keeps the dialog/draft and does
not refetch host data. P6.2.1 recorded the legacy Light panel on Dark hosts; the
maintainer subsequently requested its Dark palette correction for P6.2.2.

Shared Review appearance belongs to `components/game-review-dialog.less`, with
`--games-review-*` declared only on `.review-dialog-backdrop` and its exact Dark
root. Dark aliases global Modal/Text/Form/Action semantics; compound-specific
focus and feedback retain local ownership. Light values, geometry, typography
and runtime stay unchanged. Only the four Review Dark PNGs may change for this
approved correction, subject to renewed visual review; the other 68 MUST remain
identical. The shared SFC's 24 Tailwind / 1 arbitrary / 13 raw debts are retired
without transferring them. Missing close translation/accessibility behavior and
potential pending-close races remain separate work.

P6.3.0 gives Advanced Filter a component-local draft. The parent Search query
MUST represent committed criteria; Cancel MUST NOT affect later pagination.
Apply emits an independent snapshot, explicitly clears inactive/removed date
fields and retains the existing single-query route-replace behavior. Simple,
advanced and category requests MUST recognize cancellation through their captured
AbortController signal, reject canceled/stale writes and preserve latest-request
pending ownership. Do not swallow non-cancellation failures. This repair does not
change Search's zh CSR/en SSR-shell boundary, mounted API loading, shared Sidebar
debounce, CSS or accepted images. Missing failure feedback and overlay stacking
remain separately scoped defects, not behavior to freeze as correct.

P6.3.0a resolves Search failure feedback/recovery without changing its API/SSR
boundary. Route initialization MUST finish independently of tags/results I/O.
Current results, tag options and shared simple suggestions MUST distinguish
pending, success, empty and error. A failed page MUST NOT display an older page's
cards as current results; retry preserves the committed URL/criteria and reloads
only the failed resource. Tag retry preserves the open Filter draft. Simple
results/errors belong to the current input and locale; blur, Escape and clear
MUST NOT be undone by a late response. Expected transport/business failures are
visible UI states, while unexpected application errors remain test failures.
Keep captured signal/token cancellation, existing successful card/track/Review
behavior, debounce and locale lifecycle. The maintainer-approved scope extension
repairs short-page Filter stacking: body Teleport uses `.games-search-overlay-scope`,
sharing existing `--games-search-*` declarations with the page in `games-search.less`.
The overlay scope gets no page canvas. Filter/datepicker selectors retain their
specificity and declaration values; Style Policy approves only the exact shared
Light/Dark declaration roots and prefix. Jump and full focus redesign stay outside
this repair.

P6.3.1 completes Search dialog interaction before appearance migration. Filter
and Jump share the existing body-owned Search overlay scope and a Search-only
focus/scroll lifecycle. Keyboard navigation stays within the dialog; Cancel,
Escape and unmount restore background interaction. Datepicker Escape consumes
only its menu before Filter can close. Jump accepts only safe integer pages in
range, preserves locale/route-replace semantics and never requests an invalid
or unchanged page. Preserve the previous computed typography when replacing
clickable spans/divs with native buttons. Search's root remains transparent under
the layout-owned canvas; SidebarSearch's Home and Search cascades are distinct.
The shared Search fixture owns interaction/consumer/legacy-tag regressions and
fourteen Light/Dark local-surface visual contracts. Earlier lifecycle/failure
assertions and all pre-existing snapshots remain authoritative. Tag smoke's four
locale/viewport cases move into the Browser gate; no separate legacy runner is
needed. The maintainer accepted the new Search snapshots before P6.3.2; this work
does not authorize style-debt migration or changes to Review/Detail/Lottery.

P6.3.2 completes the authorized Search appearance migration while retaining those
runtime contracts. Advanced Search, Filter/Jump, result typography and motion
stay in `pages/games-search.less` with the existing page/body-overlay declaration
roots and `--games-search-*`. Shared SidebarSearch uses `pages/games.less` and
`--games-sidebar-search-*` on existing Games roots; its higher-specificity Dark
selectors MUST retain their priority over Search overrides. SFCs keep private
geometry and unchanged script lifecycles. Preserve the measured inherited
16px/24px inputs, Datepicker adapter, complete transition sets and focus reset
values when removing utility/important debt. Neither policy ownership nor global
tokens expand. All 147 Browser / 87 Visual contracts and 86 PNGs remain unchanged.
The remaining Games raw debt belongs to Detail/BlurWrapper and the legacy root;
P6.3 completion does not authorize Lottery, Detail or P6.6 root removal.

### Lottery runtime boundary (P6.4.1)

Lottery Prize and Activation remain CSR/noindex surfaces. The mounted Prize GET
MUST distinguish loading, ready-empty and unavailable data; Retry is local,
single-flight and canceled on unmount. The live request owner is
`app/utils/api/game.ts`, not the unused duplicate in `app/services/game.ts`.
Prize reads pass `lang=zh|en`. Game Backend selects activity title/description and
prize title/platform independently, falling back to Chinese for missing English.
The same contract applies to active pools, Join and winner history. Locale changes
cancel previous-language reads and retain stale-response protection; participation
payloads, redemption-key privacy, CSR/noindex and the existing retry contract do
not change. The shared history cache retains both languages without redemption
keys, including compatibility with older single-language cache entries.
The local body-mounted Lottery dialog owns focus trapping/restoration, background
inertness and scroll cleanup. Validate a trimmed input snapshot before POST;
pending Close/Cancel remains available, and late responses cannot affect a new
instance. Successful submission clears fields and reports the activation email,
without inventing a member or refreshing Prize data. Activation's immediate and
15-second return both preserve locale, with timer cleanup on exit.

`fixtures/lottery.ts` supports fourteen real runtime cases and twelve initial
Prize/Join/Activation goldens. All participation responses are exact local test
routes; no real email, backend activation or production access is authorized.
The maintainer accepted the twelve initial Lottery goldens before P6.4.2.

P6.4.2 closes Lottery appearance in `pages/lottery.less`, using only the existing
full `.lottery-page, .lottery-activation-page, .lottery-modal` declaration group
and its Dark counterpart for `--lottery-*`. Compound canvas, elevation, action
and status tokens preserve their exact values; the layout continues to own the
transparent page roots. Vue keeps structure and unchanged runtime scripts.
Native controls retain inherited typography; explicit text hierarchy keeps the
accepted unitless line-height precision, complete transition sets and existing
reduced-motion behavior. The seven Lottery debt entries are retired without
policy exceptions or ownership expansion. All 161 Browser / 99 Visual contracts
and 98 PNGs stay unchanged. Detail/Common and final Games cleanup remain P6.5/P6.6.

### Game Detail runtime boundary (P6.5.1)

P6.5.1 establishes populated Game Detail behavior and initial visuals without
appearance migration. Lightbox owns its body Teleport, focus/Escape/scroll
lifetime and the exact `.game-detail-lightbox` / Dark `--games-detail-*` overlay
declaration root in `games.less`. BlurWrapper's locked content is inert; NSFW
confirmation keeps cancel/confirm semantics and local dialog isolation. Tabs use
manual activation, and unavailable reviews/recommendations remain distinct from
successful empty data with generation-safe slice Retry. Keep the original Detail
tests, earlier PNGs and style debt unchanged. Twenty new goldens require review
before P6.5.2; Game-only Insights appearance remains Detail-owned, distinct from
#108 Insights and #109 Site capability styles.
The approved runtime correction fixes Header's reporting zone to Asia/Shanghai
across SSR/client and keeps ECharts instances outside Vue's deep proxy graph;
it does not authorize a chart palette or data-flow migration.

### Game Detail appearance boundary (P6.5.2)

After maintainer approval of the twenty Detail goldens, Detail appearance belongs
to `pages/games.less` with exact `.game-detail-page` / `html.dark .game-detail-page`
declaration roots and `--games-detail-*`. The body Lightbox keeps its separate
exact roots; it MUST NOT depend on page inheritance. BlurWrapper and LinkTag are
Detail consumers, not global visual primitives. Similar keeps private geometry
scoped; its typography and motion belong to the domain Less owner.

Player/Price charts read resolved inherited CSS palette values from their real
elements after theme application, retaining shallow ECharts instances and existing
data/request/tooltip behavior. Game-only discount/platform appearance and the two
Game-specific cascade overrides in `insights.less` stay within this boundary;
Site capability and shared Insights debt remain #109/#108 work. Preserve native
button inheritance and unitless typography ratios rather than utility intent.

The accepted 179 Browser / 119 Visual contracts and all 118 PNGs stay unchanged.
P6.5 removes only its nineteen rule/file debt entries or budgets; P6.6 separately
closes the audited legacy Games root cleanup.

### Games completion boundary (P6.6)

The existing `.games-page` scope MUST remain on Home/Search/Detail while its
audited shared consumers depend on it. Its fourteen retained compatibility roles
and distinct root `--games-text-body` belong to `pages/games.less`; new surfaces
MUST use their explicit domain owner, not extend this compatibility palette.
Preserve Dark root specificity, asymmetric Detail aliases and higher-priority
shared SidebarSearch rules. The layout owns the transparent canvas.

Plain Pagination's transparent border/fill and square corners belong to the
explicit primitive modifier `.gf-pagination--plain`. Search's domain adapter
retains text emphasis, underline and its accepted typography/motion/geometry;
it MUST NOT reset the shared `.gf-pagination__button*` appearance again.

P6 completion means no unassigned Games appearance debt, not removal of every
shared root or private scoped style. Site #109, experimental Ambient and the
five shared/domain Insights `!important` declarations keep their documented
owners. The accepted P6 runtime/Visual contracts remain intact through P7.
Retirement mapping MUST preserve effective checks, not stale class spellings or
uncompared success screenshot reports. No #108/#109 migration follows from closure.

### P7 completion boundary

`visual:guard` and the eight P7.1 runtime runners MUST NOT be restored. Formal
Browser tests own route/theme/overflow and behavior; style-policy owns the old
source scans; pinned Visual owns accepted pixels. Performance tools, reusable
fixture data/transport, Node Contract Guards and policy tooling tests deliberately
remain under `scripts/`. Directory relocation is not a completion requirement.

P7.2 changes no production code, debt or golden. Remaining measured raw debt is
Site #109's 388 (including six Site capability values physically in `insights.less`)
and preserved ambient's 75; five Insights important declarations belong to #108.
The manifest remains authoritative at rule/file granularity, with no wildcard
exceptions. Completion of #124 requires final-SHA local and actual remote gates,
maintainer acceptance and explicit known-issue ownership, not zero excluded debt.

### P1 enforcement and maintenance

ESLint uses the official Nuxt static flat-config factory without a runtime module
or formatting policy. `eslint-suppressions.json` is the one-time historical
bootstrap, not a license to suppress new code. Contributors MUST fix new lint
findings, MUST NOT rerun bulk suppress-all to grant debt, and MUST NOT use
`--pass-on-unpruned-suppressions`. Use `pnpm run lint:prune` when removing debt;
unused suppressions fail the normal lint command.

Stylelint owns correctness only. Its config documents narrow Less/Tailwind/Vue
compatibility decisions and existing cascade patterns; it does not enforce
formatting or duplicate the six architecture-debt rules. The configured current
source must have zero Stylelint violations; no Stylelint debt manifest exists.

`scripts/style-policy.mjs` discovers tracked and non-ignored new application
source, parses SFC/TS/CSS/Less, extracts shared style facts, applies the six
detectors and exact exceptions, then compares every rule/file budget:

- `actual > baseline`: regression, fail.
- `actual < baseline`: stale budget, fail.
- `actual == baseline`: pass.

`pnpm run style:policy:update` can only lower budgets or remove zero entries. If
any pair increased, it refuses every write. It preserves exceptions and refuses
to overwrite a manifest changed during the scan. New files default to zero;
moving debt never transfers its budget automatically.

Class extraction works backward from template/DOM class sinks through bounded
local constants, arrays/objects, refs/computed values and simple function returns.
Script visual settings/palettes and actual `innerHTML`/`v-html` embedded styles
are supported without interpreting arbitrary runtime code. Authored literals
are deduplicated across references. Unresolved runtime expressions are an
explicit static-analysis boundary, not permission for new architecture debt.
Cross-file execution, imported function evaluation and general dynamic string
solving are outside P1. Parse/read/manifest/compiler failures fail closed.

Tests verify extraction, classification, exact exceptions, all three comparison
outcomes, update safety, fail-closed CLI behavior and current per-file parity.
The existing Nav Web CI job runs each guard separately before the preserved
typecheck, Insights semantics, SEO recovery and build steps. P1 adds no Vitest,
Playwright Test migration, production style cleanup or UI behavior change.

## Release Notes final public contract (#132 P3)

The editorial `/updates` index and `/updates/:id` article, including `/en`, replace
the historical Timeline. Index renders only metadata/summary and groups history
by China-site month. Detail owns sanitized SSR body and API-provided older/newer
neighbors. No year toggle, marker animation or Timeline owner remains. P3.1 replaces
P3's all-history index with bounded SSR: Latest plus 20 history items. Subsequent
pages require an explicit More action, never a hidden all-data fetch or infinite
scroll. Errors retain loaded content; retry repeats that page. Pending clicks are
deduplicated, and locale change/unmount protects against stale append responses.

Each direct SSR entry requests its own API exactly once; hydration reuses payload.
Detail 404 is HTTP 404 and upstream failure is 503. Index retains explicit loading,
ready, empty and unavailable states with local retry. Locale preserves release ID.
The public index projection is metadata-only; P1 lifecycle and visibility remain
backend-authoritative. No new schema, permission or media integration is involved.

Public rendering MUST satisfy every shared `contracts/fixtures/update-markdown.json`
case and `contracts/update-markdown.md` using the exact P2 versions. Only sanitized
renderer output may enter `v-html`. Metadata never derives summaries from body.
App-level locale head remains canonical/hreflang owner; detail supplies localized
SEO, article OG metadata and its own OG URL. Sitemap includes both localized
public release IDs and fails closed if any inventory fails, including Updates.
The sitemap's existing unpaged 100-release inventory remains unchanged. P3.1 index
pagination defaults to page 1 / size 100 (max); the public UI explicitly uses 21.
Detail labels are Previous/Next (上一篇/下一篇), while older/newer identity is unchanged.

Functional `regression/updates.spec.ts` and `seo-recovery.spec.ts` own behavior;
`visual/updates-page.spec.ts` owns exactly eight Index/Article × Light/Dark ×
Desktop/Mobile goldens. This expressly supersedes the four Timeline goldens and
historical viewport/marker assumptions above. Generate only this authorized spec
in the pinned runner, then compare separately. No unrelated golden may change.
Remote acceptance uses #134 Fast, Manual Full and Manual Visual independently;
maintainer Admin/Index/Article review remains explicit, not inferred from pixels.

## Games Home Showcase runtime (#119-C)

`/games` and `/en/games` load Home and the independent optional Showcase through
one SSR async-data owner (since #140-C also Collections Home): exactly three
parallel GETs. Home/Showcase use normalized `lang` and `region=CN`. Tests assert
exact unordered sets, not arrival order. Successful Showcase SSR, including empty,
has zero hydration GETs; unavailable SSR gets one mounted recovery as specified
above. Failure hides the entire surface and its margin without affecting Home. Neither the Home DTO/cache nor page SEO changes.

`GameHomeShowcase` renders before Recently Released inside GameInfoPanel, in
Composer order. Desktop uses 64/36 at 320px, Tablet stacks 2:1 artwork, Mobile
stacks 16:9 with two tags/two summary lines and no Editorial Note. Fixed localized
context/action copy surrounds strict backend-locale content. Sponsored never shows
an Editorial Note. Managed/Steam images use their existing routing components;
the managed resolver accepts only the frozen Showcase AVIF key grammar. Mobile
artwork falls back to the desktop key when absent. Exhausted images retain the
slide, count, text and actions on a neutral surface. ManagedAssetImage uses an
optional mobile object key with native picture/source selection, so only the
current viewport variant downloads. The default single-image contract is retained.
Visited Showcase image nodes and their provider/failure state survive until the
snapshot/page ends. A slow new image leaves the previous decoded artwork visible
until ready; stale completions never replace a newer selection. No CDN URLs are
constructed in business components.

One item has no controls/count; two to four use native, circular Prev/Next
(always enabled).
Index lives only in the page instance and resets on snapshot change. Initial SSR
has no animation; switching uses the existing 200ms media/text transition, instant
under reduced motion. Rapid clicks update immediately without animation locks.
There is no swipe, keyboard-shell capture or focus movement. Circular Autoplay below
supersedes the earlier no-autoplay and non-circular restrictions. The polite live region is
initially empty and updates only after a manual switch, never an automatic one.
Internal actions use localized Game routes; external HTTPS
actions use `noopener noreferrer` in a new tab. Artwork/title/actions are separate
links, not one clickable Hero. Appearance belongs to `games.less` under
`--games-home-showcase-*`; Tailwind owns structure. Existing style debt and goldens
cannot be increased or updated by this implementation. Showcase material matches
Games Home cards (Light 40%, Dark 6.5%, matching shadow/text family) through its
own tokens. Narrow layouts place the 44px navigation controls below the CTAs.

Tracking requires the active item, >=50% root intersection, a visible document
and a continuous 1000ms interval. Leaving visibility, switching items/snapshots
or unmounting cancels that timer. Page-local snapshot/item dedupe supplements the
server authority. Pending media cannot generate impressions; the previous artwork
keeps its own destination/token during handoff, while new text/CTA follow the
selected index. Cached hidden frames never become tracking owners. Lazy UUIDv4
identity uses sessionStorage only, with page-memory fallback when blocked;
missing crypto.randomUUID disables tracking. Click sources
are artwork/title/primary/secondary, sent without awaiting or blocking navigation.
Direct best-effort `$fetch<void>` accepts the event endpoint's empty 204, with no
useApi envelope, retry, toast or analytics-driven UI state. Browser accounting
separates these POSTs from forbidden hydration refetches.

The existing Games Home fixture defaults to empty Showcase and owns deterministic
populated/failure scenarios. `games-home-showcase.spec.ts` is explicitly assigned
to the existing Full Games regression group. Runtime/Functional completion is
separate from maintainer Desktop/Tablet/Mobile review, new Visual acceptance and
issue closure. No accepted Visual spec or snapshot changes in this first pass.


### Showcase circular Autoplay (#119-C follow-up)

`useGameShowcaseAutoplay` owns one forward-only timeout with
`AUTO_ADVANCE_MS = 6000`. It owns its own intersection/visibility/hover/focus/media
listeners; it neither imports Tracking nor sends analytics. It runs only for
multiple items, including the last index, when the existing `artworkState.ready` is
true and `artworkState.key === item.key`, root intersection is >=50%, the document
is visible, the mouse is outside, focus is outside, the user has not paused and
reduced motion is off. Every interruption clears the timeout; resumption starts a
new complete 6000ms window. Pending pictures never consume that window.

Manual Prev/Next reset the window without permanently disabling autoplay. Pause
is page-instance memory only, survives snapshot changes, and writes no cookie or
storage. Play restarts a full window when eligible. Both automatic and manual
navigation wrap with modulo arithmetic: last Next returns to first, first Prev
returns to last. This explicitly supersedes V1's stop-at-end/non-circular behavior.
The navigation owner passes the actual move direction to Artwork so wrapping
retains the existing forward/backward transition without inferring from indexes.
Reduced motion disables autoplay completely and hides its control; zero/one item
also hides the control and has no timer. Mount/unmount own and release every
listener, observer and timer. Automatic changes never update aria-live.

The compact Phosphor Pause/Play button reuses Showcase control appearance and the
existing gameShowcase.control namespace. Only the control row may wrap when its
added button requires space. Layout proportions, height, typography, colors,
transitions, crop, CTA and disclosure are unchanged. Existing visited-image
retention, slow/stale handoff and neutral fallback remain Artwork's responsibility.
Tracking still requires its independent continuous one-second visibility window
and snapshot/item dedupe across every loop; an auto switch does not create an
impression itself.

Maintainer approval of the preceding appearance is recorded separately from this
autoplay follow-up's pending dynamic review. No Visual baseline update or final
Visual Closure is authorized by this work.
