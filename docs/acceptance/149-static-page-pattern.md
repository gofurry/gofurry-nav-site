# #149 — Static page pattern

2026-10-22. Implementation follows the maintainer's amendment: ordinary CSS
`background-image`, **not** an Alpha Mask. The supplied 512×512 RGBA WebP has
transparent outer edges; wolf details live in RGB. Original asset SHA-256:
`f2f59f8c44bcc77986f7757fde4b32aa127e5d5606089c962b57cca5557042c1`.
The shipped bytes are unchanged. A 512px repeat retains the source density;
only the decoration receives Light 0.055 / Dark 0.075 opacity.

## Scope

- `PublicPageBackground.vue` remains the single canvas owner, now without any
  script, storage, catalog, object URL or runtime selection. Mask/color tokens
  and the old SVG are removed. Old preferences and IndexedDB data are ignored.
- Preferences keeps Home/Hero and Resource routing. Background editors, presets,
  upload/storage utilities, public catalog, Admin UI/routes/CRUD and SQL queries
  are removed. SQLC output was generated, never edited manually.
- Hero local storage, icon/Hero publication, CDN snapshots/fallback, Audit and
  CloudOps are preserved. Shared Hero Audit helpers formerly placed in the
  pattern controller are retained in `assets.go`; the live Audit rollback case
  now verifies icon replacement instead of the removed pattern mutation.
- No migration, schema snapshot or database data changed. SQLC's generated
  `GfnBackgroundPattern` schema types still reflect the retained table; they are
  not catalog business DTOs. CloudOps still recognizes historical pattern keys
  for inspection/repair, while publishing new `pattern` objects is rejected.

## Validation

| Result | Evidence |
| --- | --- |
| PASS | `go tool sqlc generate -f ../sqlc.yaml` from `tools`; only the two managed-assets query outputs changed |
| PASS | Go 1.26.10: Admin `go test -mod=readonly ./internal/app/navadmin/controller ./internal/transport/http/router ./internal/infra/assets`; Nav Backend Home service/controller and routers; retired public route explicitly returns 404 |
| PASS | Admin `pnpm run typecheck`; `pnpm exec vitest run src/features/assets/asset-pages.test.tsx` (4 tests); scoped Oxlint has warnings but no errors |
| PASS | Web `pnpm run typecheck`; scoped ESLint/Stylelint; `pnpm run style:policy` (zero new debt); scoped managed-assets/Hero unit tests (25); `node scripts/insights-semantics.test.mjs`; focused style-detector tests (29) |
| PASS | `node --test .github/scripts/nav-web-regression-groups.test.mjs` (8); explicit owner list replaces retired storage spec with static texture spec; no workflow changed |
| PASS | 18 distinct focused development-runtime Browser cases: managed-asset routing (1), resource-routing Preferences (4), Preferences foundation (9), static texture (4). Completed as 8 passing cases plus a narrowed final 10-case run; no automatic retries |
| FAIL (development readiness) | Initial broad Insights owner attempt: two unrelated navigation-clearance cases stopped in `insights-runtime.ts`'s existing 5s production-chunk readiness assertion while Vite modules were still loading, before clearance assertions. Run cancelled and narrowed; no assertion, retry or timeout guard was relaxed. Contexts: `.cache/149-review/initial-development-run/` |
| NOT RUN | Production build, full suites, Visual compare/update, CI workflows, database/cloud integrations and deployment |

Browser uses the existing `GOFURRY_FIXTURE_DEV=1` opt-in, `--workers=1`,
`--timeout=120000`, no retries, and isolated API fixtures; it is source-runtime
evidence, **not** production/cloud acceptance. The focused owners are static page
background, Preferences foundation, resource-routing Preferences and managed-asset
routing. New assertions cover SSR, exact asset bytes, theme/tile/opacity,
retired storage isolation and zero catalog/IndexedDB reads. Existing assertions
cover keyboard/Save/Cancel, Reduced Motion interruption, focus, icon/Hero route
snapshots and real injected image failures.

Actual Browser commands (from `apps/cn/web`, with the two environment flags
`GOFURRY_FIXTURE_DEV=1` and `GOFURRY_149_REVIEW=1`):

```text
pnpm exec playwright test tests/browser/regression/static-page-background.spec.ts tests/browser/regression/preferences-foundation.spec.ts tests/browser/regression/resource-routing-preferences.spec.ts tests/browser/regression/managed-asset-routing.spec.ts --workers=1 --timeout=120000 --max-failures=1
pnpm exec playwright test tests/browser/regression/static-page-background.spec.ts tests/browser/regression/preferences-foundation.spec.ts -g 'Preferences retain Hero|static page background' --workers=1 --timeout=120000 --max-failures=1
```

The first command passed 8 cases before an added count assertion accidentally
included Hero's nested viewport Tabs. Its selector was corrected to the direct
Modal Tab bar (`.gf-preferences-modal > .preferences-tabs`); the final second
command passed all 10 remaining cases, including six responsive/theme variants.
No product behavior or test expectation was relaxed to obtain those results.

## Visual review / remaining work

PENDING maintainer review: Light/Dark at 390/1440px, wolf detail readability,
repeat joins/density and unobtrusive contrast behind text; two-tab Preferences
and unchanged Hero/Resource routing. Representative screenshots, when enabled,
are in ignored `.cache/149-review/`: `background-{1440,390}-{light,dark}.png`.
Inspection confirms visible RGB wolf details without obvious hard tile edges;
screenshots are not human approval.

The two obsolete background-tab goldens are deleted with their cases. No remaining
golden is refreshed. The remaining Preferences tab-bar and public-canvas goldens
need separately reviewed reconciliation in the existing pinned Linux environment;
do not claim Visual green or hide the intended differences with masks/tolerances.
Physical table/object cleanup requires a separate data audit and safe migration.
No production operation or Issue closure was performed.
