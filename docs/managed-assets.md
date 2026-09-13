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

Admin reads `external_services.asset_storage` and `external_services.cloud_ops` from its explicit configuration file. Copy the shape from `apps/cn/admin/config/server.example.yaml` and put actual development values only in the ignored `config/server.yaml`. Nav Backend has no cloud secret; Nav Web needs public CDN origins only.

Run the opt-in storage acceptance from `apps/cn/admin`:

```powershell
$env:GOFURRY_ASSET_DEV_CONFIG = (Resolve-Path config/server.yaml).Path
go test ./internal/infra/assets -run TestRealDevStorage -count=1 -v
```

It requires development bucket/CDN names, calls real COS and R2 Put/Head/Get, checks both CDN responses, performs mirror repair, and verifies the fixed probe length/hash. It leaves a small set of immutable test objects (including a synthetic, unreferenced site ID); runtime credentials deliberately cannot delete objects. Normal unit tests skip live acceptance without this explicit configuration variable. A skipped test is not cloud acceptance.
