# Managed asset production cutover

This is a maintenance-window cutover, not a rolling deployment. The tools never
load the root `.env`, connect to PostgreSQL, purge a CDN, delete old objects or
execute their generated SQL. Goose remains the sole schema owner.

## Prepare before downtime

1. Configure production COS Primary, R2 Mirror, immutable caching and the two
   public asset origins. Both origins must return CORS headers for **all managed
   assets as well as the probe**: SVG CSS masks require CORS even when ordinary
   `<img>` works. Allow the actual frontend origin for GET/HEAD/OPTIONS, expose
   Content-Length/ETag and provider cache status, and forward Origin correctly.
   Public credential-free resources may use `Access-Control-Allow-Origin: *`.
   The development R2 bucket currently allows `http://localhost:3000`; an
   arbitrary loopback port is not equivalent.
2. Keep the fixed 8192-byte `system/probes/cdn.bin` in both stores. It is not
   content-addressed, has its own approximately one-day CDN cache policy, and
   is not created/replaced by ordinary asset upload or mirror repair.
3. Export every nonempty `gfn_site.icon`, **including soft-deleted sites**, with
   its ID. Download the original icon bytes and both existing Hero pools to a
   private local staging directory. Keep icon extensions; prepare valid AVIF
   Hero files. Do not infer mobile entries from desktop or resize automatically.
4. Prepare `sources.json` beside those files (relative file paths resolve there):

```json
{
  "site_icons": [
    { "site_id": 123, "old": "furrywiki.svg", "file": "icons/furrywiki.svg" }
  ],
  "hero_desktop": [
    { "name": "legacy-desktop-1", "file": "desktop/first.avif" }
  ],
  "hero_mobile": []
}
```

IDs are exact JSON integers, decoded directly to int64. `old` may be NULL when
intentionally adding an icon to an empty row. Empty Hero pools are valid.
Do not check private inventories or generated production artifacts into Git.

5. From `apps/cn/admin`, run the staging tool with an explicit ignored YAML and a
   **new** output directory:

```text
go run ./tools/asset-cutover --config /private/admin/server.yaml --sources /private/staging/sources.json --output-dir /private/staging/prepared
```

All files and mappings are validated before any upload. The tool uploads to COS
and then best-effort R2 using the same adapters, hashes, metadata and validation
as Admin, without changing business references. A COS failure stops preparation;
an R2 failure is recorded in `publication.json` and never changes Primary success.
Resolve mirror warnings with CloudOps repair before the planned cutover if the
mirror is expected to be available at launch. Unreferenced uploads are retained.

Outputs are `manifest.json`, `publication.json`, `cutover.sql`, and `rollback.sql`.
The manifest contains `site_icons: [{site_id,old,new}]` and independent
`hero_desktop/hero_mobile: [{name,object_key}]`. Review both SQL files, compare
the manifest against the DB export, and use Object Inspector plus both CDN GETs
to verify representative/all staged objects. The `--manifest <file>` alternative
regenerates SQL offline from an already-staged manifest; it makes no cloud calls.

## Maintenance window

1. Enable the existing [Nginx maintenance page](../apps/cn/nav-web/DEPLOYMENT.md#maintenance-page).
   Stop old Nav Backend, Admin and frontend/content writers. Keep production
   backups of the old binaries, frontend, explicit configs and a full GFN dump.
2. Apply the Goose migration `20260913010000_nav_managed_assets.sql` with the
   GFN **migrator** role, following [db/README.md](../db/README.md). Applications
   continue using their application role. Goose performs no object I/O.
3. Execute the reviewed `cutover.sql` against GFN with the migrator connection:
   `psql -v ON_ERROR_STOP=1 -f /private/staging/prepared/cutover.sql` (supply
   credentials through the operator's private connection configuration).
   The transaction locks the two affected tables, requires an empty new Hero
   table, verifies every old icon value and rejects missing mappings. It updates
   only mapped icons and inserts the two explicit Hero pools. Patterns start
   empty and can be published through Admin after deployment.
4. Install the new explicit Admin YAML fields from `server.example.yaml`; remove
   Nav Backend `resource` / numbered Hero fields. Build/deploy the embedded React
   Admin, Go Nav Backend and Nuxt together. Nuxt needs only
   `NUXT_PUBLIC_ASSET_PRIMARY_BASE` / `NUXT_PUBLIC_ASSET_MIRROR_BASE` public origins;
   remove the old SiteLogo/Game prefix environment variables. Steam settings
   remain unchanged. Do not deploy only the frontend against the old Home API.
5. With the old Nav Backend stopped, invalidate only the affected derived Nav
   keys: `site:list:v2`, `featured-sites:list`, `nav:home:v3:zh/en`,
   `nav:site-directory:v1:zh/en`, and keys matching `nav:site-group:v1:*`.
   Use SCAN and explicit DEL for that prefix in the configured Redis database;
   do not FLUSHDB/FLUSHALL or delete observation/view counters. The new Backend's
   startup cache warm-up rebuilds the site list and derived views in order.
   Wait until both `/api/v2/nav/home?lang=zh` and `?lang=en` return schema 4 with
   the expected groups and managed icon keys. The home Redis **key name stays
   v3** deliberately; its envelope is now schema 4.
6. Smoke Admin icon upload/clear, each Hero pool, SVG preview/catalog, public
   desktop/mobile home and server/local backgrounds. Verify SSR HTML with no
   cookie and `gf_asset_cdn=mirror`; check real images, CORS and both probes in
   browser DevTools. Test Mirror outage without blocking Primary publication.
7. In Admin Cloud resources, submit the **main hostname** EdgeOne purge shortcut
   and wait for the task result. This is `purge_host(go-furry.com)`, not a zone-wide
   purge. Restore Nginx traffic after smoke checks pass. Retain old COS paths
   throughout the observation period; there is no automated deletion policy.

## Roll back during the same window

Keep maintenance enabled and stop new content writes. Run the paired
`rollback.sql` with `ON_ERROR_STOP=1`; it restores original icon values and removes
only the manifest's initial Hero rows. It rejects changes made to those rows
after cutover, instead of overwriting later operator work. Restore the backed-up
old binaries/frontend/configs, invalidate/rebuild the same derived caches, smoke,
purge the main hostname and restore traffic. The additive empty tables can remain;
if schema rollback is required, use Goose rather than hand-written DROP commands.
For rollback after live content changes, use a reviewed DB backup/reconciliation
plan; this one-time paired script intentionally refuses that case.

## Executable verification

- Admin: `go test ./tools/asset-cutover`; real PG round trip with
  `GOFURRY_ASSET_DEV_CONFIG` and `GOFURRY_NAV_ASSET_TEST_URL` set explicitly:
  `go test ./tools/asset-cutover -run TestRealDevCutoverRoundTrip -count=1 -v`.
  The latter uses session-local temporary tables, never production table writes.
- Nav Web: `npm run build` then `npm run assets:cloud-smoke` loads public dev
  origins from the ignored frontend `.env`. The built Nuxt/API rows are isolated;
  browser requests to real dev COS/R2 CDNs are not mocked. Local requests are
  routed through `http://localhost:3000` inside the test browser to match bucket
  CORS without stopping the developer's existing frontend. Fault cases abort
  selected real requests to exercise runtime fallback.
- The independent [real storage, Admin API and CloudOps suites](managed-assets.md)
  cover real writes, PG/GFA transactions, object repair and provider purge APIs.
  Unit/fixture success is never a replacement for those suites.
