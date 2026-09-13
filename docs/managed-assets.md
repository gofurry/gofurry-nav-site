# Managed assets: development and operations

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
maintenance cutover must clear/rebuild derived caches before traffic resumes.

Nuxt configures `NUXT_PUBLIC_ASSET_PRIMARY_BASE` and
`NUXT_PUBLIC_ASSET_MIRROR_BASE`, with public origins only. SSR uses the valid
`gf_asset_cdn` cookie or Primary. After hydration, idle work performs two rounds
of parallel GET probes, consumes and validates all 8192 bytes, and caches the
choice for 12 hours. Mirror must be at least 20% or 50 ms faster. A real asset
failure tries the other provider, invalidates the cookie and schedules another
probe; terminal fallbacks are the bundled default logo, no Hero image and the
bundled page pattern. Steam asset resolution remains separate.

Run Nav Web `npm run assets:test`, `npm run insights:semantics`,
`npm run seo:recovery:test`, `npm run typecheck`, and `npm run build`.
Nav Backend's opt-in `TestRealDevAppearanceQueries` uses only transaction-local
temporary tables copied from the Goose schema to verify pool isolation,
disabled/deleted filtering and pattern order on real PostgreSQL.

## Page background preferences

Preferences offers bundled default, enabled server patterns and a local image.
SVG uses the same repeating CSS mask as the public page, including light/dark
colors, opacity and tile size. Raster images expose opacity and size only.
Only explicit overrides are stored in `gf_background_preference`; catalog
defaults remain live. Local SVG is validated, and SVG/raster bytes stay in
IndexedDB. No local image upload endpoint exists. Clearing removes the stored
Blob; missing/disabled patterns, missing local files and exhausted CDN retries
fall back to the bundled pattern. SSR always renders the bundled background
before browser-only preferences are loaded.

Run `npm run background:smoke` for real Chromium IndexedDB persistence,
SVG/raster validation, clearing and verification that local images cause no
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

## Development configuration and storage acceptance

The durable boundaries are in [the asset contract](../contracts/assets.md).
Production staging, generated cutover/rollback SQL and the maintenance sequence
are in [the cutover runbook](managed-assets-cutover.md).

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
