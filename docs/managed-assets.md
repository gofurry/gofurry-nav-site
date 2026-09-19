# Managed assets: development and operations

## Production status

The initial production cutover completed on 2026-09-14. Managed Assets is the
production architecture: COS / EdgeOne is Primary and R2 / Cloudflare is the
best-effort Mirror. Daily content writes go through Admin. See the
[production acceptance record](acceptance/v3-alpha8-managed-assets-production-acceptance.md)
and the [historical cutover runbook](managed-assets-cutover.md), retained for
rollback context and first-time deployment of other environments. Do not rerun
initial `cutover.sql` against the already-migrated production database.

## Admin content management

All paths below are under `/api/v1/nav`. Reads require `content.read`; mutations
require `content.write`, session authentication and the existing CSRF token.

| Resource | API |
| --- | --- |
| Site icon | `POST /sites/:id/icon` multipart `file`; `DELETE` clears to SQL NULL |
| Hero | `GET/POST /hero-assets`, `GET/PUT/DELETE /hero-assets/:id`, `POST /hero-assets/:id/file` |
| Pattern | `GET/POST /background-patterns`, `GET/PUT/DELETE /background-patterns/:id`, `POST /background-patterns/:id/file` |

Creation uses multipart fields matching the metadata JSON names plus `file`.
Hero creation requires `variant=desktop|mobile`; variant and object key cannot
be changed through metadata updates. List endpoints use `page_num/page_size`;
Hero lists additionally select a variant. New asset IDs are JSON strings.
Uploads return `object_key`, `primary`, `mirror`, `warnings`, and creation also
returns `item`. These publication statuses are never stored in PostgreSQL.
Admin reads include public preview links, never credentials.

Pattern metadata: `name`, `name_en`, `light_color`, `dark_color`,
`light_opacity`, `dark_opacity`, `default_size_px`, `enabled`, `sort_order`.
The editor previews selected SVG bytes as repeating CSS masks before uploading.
File replacement and metadata saving are separate explicit actions.

Ordinary site creation/update rejects any `icon` input, including NULL; its SQL
cannot change the icon. Dedicated mutations preserve original icon bytes and
invalidate the existing Nav site cache. Hero/Pattern reads use current database
state, so no new Redis keys or background synchronization jobs are introduced.

`GOFURRY_ASSET_DEV_CONFIG=<ignored Admin YAML> go test ./internal/app/navadmin/controller -run TestRealDevManagedAssetAPI -count=1 -v`
exercises real COS/R2 plus the development GFN/GFA transaction and audit path.
It creates labeled dev records and soft-deletes them afterward. Immutable files
and their audit trail are retained; no production data or object deletion occurs.

## Public appearance and CDN resolution

Home schema v4 replaces `backgrounds` with `hero.desktop` and `hero.mobile`.
Each is either NULL or `{ "id": "...", "object_key": "nav/hero/..." }`.
`GET /api/v2/nav/home/hero` replaces `/home/backgrounds`.
`GET /api/v2/nav/appearance/patterns` returns schema version 1 and an ordered
`patterns` array. Public responses contain keys, never COS/R2/CDN URLs.
The old Nav V1-to-V2 assessment describes the historical background endpoint;
this contract supersedes that endpoint and the old numbered resource config.

The derived home cache still owns navigation content. Hero selection is composed
at request time from independent pools, so disabling or deleting an asset takes
effect without waiting for the content cache. Empty pools stay NULL and never
borrow from the other viewport. The reader rejects pre-v4 home payloads; the
initial maintenance cutover required clearing/rebuilding derived caches before traffic resumed.

### Hero source preferences (#112)

Home preferences now owns **Random cloud / Fixed cloud / Local folder** inside
the existing first tab. Desktop and mobile fixed selections are independent.
`gf_hero_mode` (`random`, `fixed`, `local`), `gf_hero_desktop_id`, and
`gf_hero_mobile_id` are one-year, path `/`, SameSite=Lax cookies. IDs are positive
signed-bigint **strings**, including cookie decoding; object keys are never
persisted. Missing/invalid modes default to Random. Leaving Fixed retains its IDs.

Both `GET /api/v2/nav/home` and `GET /api/v2/nav/home/hero` accept optional
`hero_desktop_id` and `hero_mobile_id`. Each ID resolves against the current
enabled, non-deleted row of that variant before falling back to its own random
pool. Missing, malformed, wrong-variant, disabled or deleted pins cannot affect
the other viewport. Replacing an asset's file under the same ID therefore takes
effect on the next request. `hero_mode=local` skips both cloud pool queries and
returns null Hero entries. Existing clients omitting these parameters retain
the random behavior; navigation-content caches and Home schema version 4 stay
unchanged. No new database table or migration is needed.

`GET /api/v2/nav/appearance/heroes?variant=desktop&page_num=1&page_size=12&selected_id=123`
is a public, read-only metadata catalog. `variant` is required (`desktop` or
`mobile`); page defaults to 1 (maximum 1,000,000), size to 12 (maximum 24), with invalid parameters
returning HTTP 400. Stable ID ordering, `total`, `items` and optional `selected`
all use only enabled/non-deleted rows. `selected_id` resolves a selection outside
the current page directly. Response schema version is 1; items contain string
`id`, `name`, and `object_key`, with no provider URL. Empty results use `[]` and
an unavailable selected row uses `null`.

Nuxt reads cookies before its SSR Home fetch. The async-data identity includes
language and the initial preference tuple; hydration reuses the same response.
Fixed therefore never renders Random first. Local SSR emits only the neutral
surface, no cloud Hero URL. After mount Local restores the folder/cache from
IndexedDB; unavailable or undecodable local data switches to Random and performs
one focused Hero fetch. Stored folder handles/cache are never consulted for
explicit Random/Fixed modes. Only a missing mode cookie permits a one-time legacy
local restoration; it writes the new mode without deleting the handle/cache.
That first upgrade alone may transition from the SSR cloud image to the old
local image, since the server cannot read legacy browser storage.

The editor holds drafts until Save. Source options reuse the page-background
selector, and compact underline tabs keep desktop/mobile browsing independent.
Arrows browse one preview at a time, automatically fetching adjacent metadata
pages at their boundaries. The preview is the draft selection; there is no
separate apply or random action inside the picker. Each viewport keeps its page
position and metadata cache until the editor closes. For a saved ID outside the
first page, the editor uses stable ID ordering to locate its page by binary
search, so the name and ordinal are accurate without scanning the full catalog.
An IntersectionObserver mounts only the visible current preview while the Home
tab and viewport are active; metadata lookup never loads the other AVIFs. An
unavailable saved ID shows a non-blocking warning.
Cancel leaves cookies, local storage and the displayed Hero untouched. A changed
source/selection uses `/home/hero` instead of refreshing Home or reloading the
page. The previous image remains while the request and new DOM image load; the
same new `picture`/`img` is promoted on success with no blank-first transition.
Exhausted new-image fallback retains the old frame. Unchanged settings do not
reselect Hero, and background routing probes/Save retain #121 snapshot semantics.

Hero normalization lives in `tests/unit/hero-preferences.test.ts`; the real Nuxt
cookie/state composable runs in `tests/nuxt/use-hero-preferences.nuxt.test.ts`.
Use `npm run test:unit` and `npm run test:nuxt`. Production Hero/Preferences browser
regressions run in `npm run test:browser` alongside #121. The two Hero fixtures
cover SSR/hydration, bigint and independent pins, current metadata, Local/legacy
behavior, lazy previews and pagination, selected IDs outside the page, Save under
gated API/image latency, Cancel, themes and keyboard focus. Backend controller,
service and real PostgreSQL temporary-table tests cover invalid parameters,
variant isolation, partial failures, disabled/deleted filtering and pagination.

Nuxt configures `NUXT_PUBLIC_ASSET_PRIMARY_BASE` and
`NUXT_PUBLIC_ASSET_MIRROR_BASE`, with public origins only. SSR resolves the saved
mode first, then the recommendation (default Primary). After hydration, idle work
performs two rounds of parallel GET probes and consumes and SHA-256 validates all
8192 bytes of `system/probes/cdn.bin`. Mirror must be at least 20% or 50 ms faster.
Both probes failing retains the previous recommendation. Terminal fallbacks are
the bundled default logo, no Hero image and the bundled page pattern.

## Resource routing preferences (#107 / #121)

Page preferences has three scroll-snap tabs: Home, Page background, and Resource
routing. Home/End and wrapping Left/Right arrows work across all tabs. Resource
routing contains two independent sections:

- **GoFurry assets:** Automatic / EdgeOne (Primary) / Cloudflare (Mirror).
- **Steam game assets:** Automatic / China route / Global route.

Each section shows the preferred route for new resources, the latest test
recommendation, per-route latency or test state, last test time, and a manual
retest button. Steam route details disclose the underlying hosts without making
them six separate modes. Changes to modes are drafts until Save; Cancel discards
them. Tests immediately persist diagnostics and recommendations, even if the
modal is cancelled. Pinned routes still test, but results never overwrite a pin.
A failed probe is diagnostic, not proof that an entire route is unavailable;
pinned routes with failed tests remain saveable with an inline note.

| Storage | Meaning | Lifetime |
| --- | --- | --- |
| `gf_asset_cdn_mode` cookie | `auto`, `primary`, `mirror` | 1 year |
| `gf_asset_cdn` cookie | Auto recommendation, compatible with old cookies | 12 hours |
| `gf_steam_asset_mode` cookie | `auto`, `china`, `global` | 1 year |
| `gf_steam_asset_group` cookie | Steam Auto recommendation | 12 hours |
| `gf_asset_cdn_diagnostics` localStorage | GoFurry timings, states, checked/manual times | 12-hour freshness |
| `gf_steam_asset_diagnostics_v1` localStorage | Versioned Steam timings, states, sample, checked/manual times | 12-hour freshness |

Missing modes default to Auto. The old
`gofurry:steam-shared-cdn-preference:v1` record is imported only if valid and within
its original six-hour lifetime, and only when no new recommendation exists. It
becomes a recommendation, **never a pinned mode**, then the legacy record is
removed. LocalStorage history is read after hydration; SSR and initial client
render share cookie-derived state. Without history, Chinese uses China and
English uses Global as a cold-start hint only.

`useManagedAsset` snapshots its provider per object key. `useSteamAsset`, used by
`SteamAssetImage` and Gallery video posters, snapshots its candidate list per
source URL. Automatic tests, manual tests and Save do not change the URLs or
reload already successful resources. Newly mounted resources and changed
keys/sources use the latest policy. Only real load errors advance the current
snapshot's fallback chain; they may stale diagnostics/schedule a test but never
clear the saved mode. Steam's former global preference-updated event is removed.

Managed Hero uses a responsive `picture`/`img` as its decorative background layer,
with the same centered cover layout. Only that displayed image's real error event
(or an already-failed SSR image detected at mount) may advance its CDN fallback.
Hero does not use the independent `useManagedAsset` hidden-Image preloader: that
auxiliary request could fail after a CSS background had already painted, wrongly
replacing a successful Hero. Desktop/mobile sources remain separate; an empty or
exhausted pool renders transparent content without borrowing the other pool.

`NavHomePage` reuses the SSR home payload during hydration. Mount, focus, content
prewarming, probes and route preference Save do not refetch its random Hero pool.
An explicit home data refresh/new navigation may select a new key; crossing the
768px breakpoint selects the other viewport pool. Local restoration now follows
the explicit Hero source contract above, including its one-time legacy migration.
Saving resource routes never reselects a local image or cloud Hero.

The Hero lifecycle regression in `npm run test:browser` injects a late
failure into independently constructed Hero Images while allowing the actual
renderer to paint, compares the displayed source and pixels, verifies one home
request using a different-key second-response fixture, and tests explicit new
keys, pre-hydration display failures, terminal fallback and local restoration.
It also narrowly verifies the existing narrow-screen Footer hydration mismatch in
`layouts/default.vue` (server-hidden versus initially client-visible), separately
from the Hero assertions. The SSR Hero node is retained even in that case; this
Hero fix does not change the unrelated Footer/reveal lifecycle.

Steam's China group is `shared.st.dl.eccdnx.com`, followed by
`shared.cdn.steamchina.queniuam.com`. Global is `shared.akamai.steamstatic.com`,
`shared.cloudflare.steamstatic.com`, `shared.fastly.steamstatic.com`, then
`shared.steamstatic.com`. Candidate order is the preferred group followed by the
other group, preserving the full original path/query/hash and original-source
fallback where needed.

Steam probes use `Image()` timing against the first host of each group, with two
parallel-group rounds and a 2.8-second timeout per image. The bounded fallback
pool is `/store_item_assets/steam/apps/{570,440,550,730}/capsule_sm_120.jpg` in that
order. A subsequent sample is tried only if both groups failed; normally only
one sample is downloaded. URLs use a minute bucket plus distinct automatic/manual
round slots (at most four per minute/sample/host), so neither the second round nor
an immediate manual test reuses the automatic test's cached image. Lower mean
latency wins; one successful group wins; both failures retain history or the
locale hint. Browser validation on 2026-09-17 decoded all four Global representative
samples and China's 570/440/730 samples as 120×45 images. China needed retries;
its 550 sample still timed out on the workstation network. These are point-in-time
path checks, not a guarantee of route availability or latency for other users.

Both managers schedule automatic tests only after mount and browser idle,
deduplicate concurrent tests within their domain and use a 12-hour TTL. Automatic tests skip offline and
Save-Data browsers. Manual tests start immediately and have a persisted 60-second
cooldown; explicit tests may still be attempted offline. No continuous probing,
backend API, database state, cloud mutation or dependency is introduced.

`npm run test:unit` covers resolver, fixed-probe, legacy migration, scheduling,
TTL and cooldown regressions. `npm run test:nuxt` covers actual Nuxt composable
snapshots/fallback and Hero cookie/state behavior. See
[frontend testing](frontend/testing.md) for the runner boundaries.
After `npm run build`, `npm run test:browser` runs Resource Routing against the production Nuxt
application with isolated API/CDN fixtures to check SSR, actual loaded image
stability, changed resources, failures, Save/Cancel, three-tab keyboard navigation
and mobile/light/dark interaction. Playwright keeps traces/screenshots only on
failure in ignored `test-results/`; no visual baseline is defined here.
Hero lifecycle and Hero Preferences now share the same Browser Gate; their old
scripts and `assets:routing-smoke` alias are retired. This deterministic
suite does not depend on real CDN availability or substitute for cloud acceptance.

Run Nav Web `npm run test:unit`, `npm run test:nuxt`, `npm run insights:semantics`,
`npm run seo:recovery:test`, `npm run typecheck`, and `npm run build`.
Nav Backend's opt-in `TestRealDevAppearanceQueries` uses only transaction-local
temporary tables copied from the Goose schema to verify pool isolation,
disabled/deleted filtering and pattern order on real PostgreSQL.

## Page background preferences

Preferences offers bundled default, enabled server patterns and a local image.
SVG uses the same repeating CSS mask as the public page, including light/dark
colors, opacity and tile size. Raster images expose opacity and size only.
Only explicit overrides are stored in `gf_background_preference`; catalog
defaults remain live. Original SVG/raster bytes stay in
IndexedDB. No local image upload endpoint exists. Clearing removes the stored
Blob; missing/disabled patterns, missing local files and exhausted CDN retries
fall back to the bundled pattern. SSR always renders the bundled background
before browser-only preferences are loaded.

Run `npm run background:smoke` for real Chromium IndexedDB persistence,
SVG preservation/raster decoding, clearing and verification that local images cause no
network mutations.

## CloudOps

System / Cloud resources uses `/api/v1/system/cloud`:

- `GET /overview` reports configured/reachable storage using HEAD of the fixed probe and configured CDN hosts.
- `GET /object?key=...` inspects both stores and compares size, full SHA256, kind, content type and cache policy.
- `POST /object/repair-mirror` accepts `{ "key": "..." }`; COS supplies the validated bytes for R2 repair.
- `POST /edgeone/purge` and `/cloudflare/purge` accept `{ "type": "url|prefix|host", "targets": [...] }`, at most 20 targets. URL and prefix inputs include a scheme. Only configured exact hosts are allowed; the adapter translates provider enums and Cloudflare's scheme-free prefixes.
- `GET /edgeone/purge-tasks?job_id=...&page_num=1` reads the real task state. Empty job ID selects the past 24 hours.
- `POST /edgeone/purge-all` is a separate Owner-only whole-zone action. The main-site shortcut submits ordinary `host` purge.

CDN actions are bounded and never automatically retried after an ambiguous
network failure. Each cloud mutation writes an intent audit before the provider
call and a result audit afterward, under the same action and request ID. Remote
effects cannot be rolled back: if result-audit writing fails after success, the
response preserves success with a warning and the intent audit remains.

Provider contracts: [Cloudflare purge API](https://developers.cloudflare.com/api/resources/cache/methods/purge/)
and [EdgeOne task query](https://edgeone.ai/document/50532).

From Admin, run `GOFURRY_ASSET_DEV_CONFIG=<ignored dev YAML> go test ./internal/app/cloudops -run TestRealDevCloudOps -count=1 -v`.
Acceptance exercises real storage status/HEAD/repair, Cloudflare URL purge,
EdgeOne URL and **development asset host** purge, completion queries and GFA
audits. The development EdgeOne zone also contains the production main host;
the test does not purge that host or the whole zone. Whole-zone authorization
and provider mapping have separate unit coverage.

### Daily EdgeOne main-host purge

Admin optionally submits one EdgeOne `host` purge per day for the configured
`external_services.cloud_ops.edgeone.main_host` (#122). The explicit Admin YAML
adds `edgeone.scheduled_purge: { enabled: false, time: "05:30", timezone: "Asia/Shanghai" }`.
It is disabled by default; enabling it requires a fully configured EdgeOne
provider, a valid main host, strict `HH:mm`, and an explicit IANA timezone.
The production operator can enable `05:30 Asia/Shanghai` after deployment.

The application scheduler calls the same validated `Service.Purge("edgeone", …)`
as manual host operations, never `PurgeAll`, asset-host or Cloudflare scheduling.
It does not run at startup, catch up missed slots, overlap locally, or retry
failed/ambiguous submissions. A dedicated GFA session advisory lock serializes
instances; a same-slot audit lookup under that lock prevents a later instance
from submitting again after the first releases it. The existing audit table is
the intent record; no migration or scheduler table is added.

System audit action `cloud.edgeone.purge.host.scheduled` targets `main_host`, with
`requested` followed by `completed` (provider job submitted) or `failed`, sharing
a deterministic slot request ID. Intent-audit failure prevents submission;
outcome-audit failure is logged and never triggers resubmission. Task history
remains the source for remote completion. Runtime starts the scheduler after
pools/services and drains it before Redis, database and logger shutdown.

See [Admin CloudOps operations](operations/admin-cloudops.md) for configuration,
development acceptance, audit inspection, and the multi-instance prerequisites.

## Development configuration and storage acceptance

The durable boundaries are in [the asset contract](../contracts/assets.md).
Initial staging, generated cutover/rollback SQL and the maintenance sequence
are preserved in [the historical cutover runbook](managed-assets-cutover.md).

Admin reads `external_services.asset_storage` and `external_services.cloud_ops` from its explicit configuration file. Copy the shape from `apps/cn/admin/config/server.example.yaml` and put actual development values only in the ignored `config/server.yaml`. Nav Backend has no cloud secret; Nav Web needs public CDN origins only.

Run the opt-in storage acceptance from `apps/cn/admin`:

```powershell
$env:GOFURRY_ASSET_DEV_CONFIG = (Resolve-Path config/server.yaml).Path
go test ./internal/infra/assets -run TestRealDevStorage -count=1 -v
```

It requires development bucket/CDN names, calls real COS and R2 Put/Head/Get, checks both CDN responses, performs mirror repair, and verifies the fixed probe length/hash. It leaves a small set of immutable test objects (including a synthetic, unreferenced site ID); runtime credentials deliberately cannot delete objects. Normal unit tests skip live acceptance without this explicit configuration variable. A skipped test is not cloud acceptance.

`TestRealDevStorageCORS` separately checks the probe and pattern with the real
development browser Origin. A successful ordinary image GET does not prove CSS
mask usability. On 2026-09-13, acceptance found missing Primary SVG CORS headers;
the maintainer configured EdgeOne response headers for the development asset
hostname and the HTTP plus production-Nuxt browser suites then passed. Runtime
COS CAM permissions stayed unchanged; bucket configuration is not a runtime duty.

The final browser suite (`npm run assets:cloud-smoke` after a production build)
also verifies live catalog-default updates without copying defaults into browser
overrides, local SVG/raster editing and clearing, 12-hour preference cookies,
independent desktop/mobile AVIF reads and terminal icon/Hero/pattern fallbacks.
It exercises actual development CDN URLs; its isolated API row fixtures complement
the separate real PostgreSQL/Admin/cloud suites instead of replacing them.

## Preview follow-up (2026-09-13)

The initial SVG geometry whitelist is superseded by the owner's request to accept
original SVG content without restrictions. Admin and local background selection
no longer inspect declarations, comments, styles, elements or attributes; the
backend keeps the SVG extension and common upload size checks only. This also
allows the bundled `gofurry-pattern.svg`, whose comments the initial browser
validator incorrectly rejected. Rendering remains URL-based CSS masks/images.

Admin explicitly uses `Cross-Origin-Embedder-Policy: unsafe-none`: Helmet's
implicit `require-corp` blocked ordinary public CDN icon/Hero previews despite
successful uploads. Other security headers and authorization remain unchanged.
Rebuild the React bundle and restart Admin to load both changes.
