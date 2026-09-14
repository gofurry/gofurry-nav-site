# V3 alpha.8 Managed Assets Production Acceptance

## Scope

This is a non-executable historical acceptance record for the initial production
Managed Assets / dual-CDN cutover. It records the verified production facts
supplied by the maintainer for release closure; it is not a new runbook or a
claim that production checks were rerun during this documentation task.

## Production cutover

- Cutover date: 2026-09-14.
- Maintenance window: `2026-09-14T22:50:56+08:00` → `2026-09-14T23:35:55+08:00`.
- Runtime source commit: `907042a52ce2b029fed52d13f49bdc2a5251c662`.
- Goose migration: `20260913010000_nav_managed_assets.sql`.
- Initial production cutover is complete. Do not rerun initial `cutover.sql`
  against the already-migrated production database; daily content operations use Admin.

## Data and object integrity

The full custom-format GFN pre-cutover dump was created and `pg_restore -l`
validation passed. Its SHA256 is
`18d04e17f484f4feff6edef24488762454d94c78e184076e7c0b614f34418bf4`.
The frozen `cutover.sql` SHA256 is
`63db913778370c482a9b93b8397b7662c4ff8ac3ee08ef6513bd6c7a30a6bb75`.

| Migrated business references | Count |
| --- | ---: |
| Site icons | 236 |
| Desktop Hero rows | 47 |
| Mobile Hero rows | 40 |
| Hero total | 87 |
| Pattern rows at launch | 0 |

Staging contained 323 manifest references and 322 unique managed objects.
EdgeOne CDN verification passed 322/322 objects and Cloudflare CDN verification
passed 322/322 objects: 644 validated CDN object fetches in total.

The fixed CDN probe is 8192 bytes with SHA256
`a2dacdf8cbd7f21e14efa73f83610bc1f40225cae90e691788000ca82cd8004e`.

## Runtime acceptance

- Home API returned `schema_version: 4` and 12 groups.
- Managed icon references and independent Desktop/Mobile Hero object keys were validated.
- The Pattern catalog was empty as designed at launch.
- Admin Object Inspector: PASS.
- Admin Cloud Overview: both stores reported configured, reachable, and probe exists.
- Final public smoke: `go-furry.com`, `nav.go-furry.com`, and `op.go-furry.com` returned 200.
- `status.go-furry.com` showed the expected redirect/healthy behavior.
- `gf-nav`, `gf-nav-collector`, and `gofurry-admin` were active; Nav Web ran the expected cutover image.

## CDN and fallback acceptance

- Primary: COS / EdgeOne, public origin `https://assets.go-furry.com`.
- Mirror: R2 / Cloudflare, public origin `https://assets.gofurry.com`.
- Representative Primary/Mirror byte parity, immutable cache policy, CORS, and both fixed probes: PASS.
- SSR without a preference cookie used the Primary asset origin: PASS.
- SSR with `gf_asset_cdn=mirror` used the Mirror asset origin: PASS.

This production record establishes the delivery and SSR preference results above;
it does not claim an additional production failure-injection test.

## CloudOps acceptance

EdgeOne host purge for `go-furry.com` was submitted through Admin and its task
completed successfully. No zone-wide purge was used.

## Rollback / observation

Rollback artifacts, the DB backup, previous binaries/image, and old asset paths
were retained for a 7–14 day observation window. No immediate legacy object
deletion was performed. The [initial cutover runbook](../managed-assets-cutover.md)
is retained for historical evidence, rollback context, and first-time deployment
of other environments.

## Security and evidence boundary

This record contains no credentials, private DSNs, internal addresses, account
identities, or raw production data. Only the maintainer-supplied aggregate counts,
public origins, timestamps, source revision, checksums, and acceptance outcomes
are archived here; private staging artifacts are not reproduced.

## Result

Result: PASS
