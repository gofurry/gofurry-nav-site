# Admin frontend contract

`apps/cn/admin/react` is the sole development and production Admin frontend. Its frozen foundation is React 19, Vite, strict TypeScript, React Router, Tailwind CSS v4, shadcn-style wrappers over Base UI primitives, TanStack Query/Table, React Hook Form, Zod, ECharts, and Phosphor for primary system UI icons. Untouched legacy screens may retain Lucide during staged polish.

## Architecture boundaries

Node 24 and the exact `packageManager: pnpm@12.6.0` own frontend tooling.
This project has an independent `pnpm-lock.yaml` and single-project
`pnpm-workspace.yaml`; no root workspace or shared lockfile is used. Local/CI
release installs use `pnpm install --frozen-lockfile`. Dependency build scripts
require explicit package-level review. Root `task build:admin` installs and
builds React before Go compilation and preserves the deployment `dist/` companion.

UI layers flow in one direction:

~~~text
components/ui -> components/admin -> features
~~~

TanStack Query owns server state. Route filters, search, pagination, sorting, and workspace tabs use URL state where practical. React local state owns transient UI. Do not introduce a second server-state copy or a role-based client store.

Admin business forms use the shared Base UI-backed Select, DatePicker, and DateTimePicker controls rather than browser-native select/date controls. Shared controls own the common height, focus treatment, and popup behavior.

Themes default to the system preference. The shared header exposes a single Light/Dark toggle; the first manual choice becomes an explicit persisted `light` or `dark` preference. Business components consume semantic tokens: background, surface, surface-muted, foreground, muted-foreground, border, primary, success, warning, danger, and info.

The backend authorization contract remains authoritative. Navigation and actions ask whether the current principal has a capability; frontend code must not reproduce the Role-to-Capability mapping.

Managed assets use content capabilities: separate desktop/mobile Hero pools and
the Pattern catalog live under Navigation content; Site Icon uses upload/clear
inside the site workspace. Pattern selection is previewed locally with the same
repeating CSS mask model as Nav Web before any upload. System / Cloud resources
requires `cloudops.read` for inspection and task history, `cloudops.manage` for
COS-to-R2 repair and scoped CDN purges, and `cloudops.purge_all` for the Owner-only
entire-EdgeOne-zone action. The main-host button and initially collapsed full-zone
section must remain visibly distinct, with explicit purge confirmation. No object delete, Cloudflare purge everything, invented global
sync percentage, or frontend reconstruction of the role policy is exposed.

## Product structure

Top-level groups are Workbench, Collaboration, Nav Content, Game Content, Data Operations, and System. Content and operational routes are native React:

~~~text
/collaboration
/nav/sites
/nav/sites/:id
/nav/site-groups
/nav/site-groups/:id/curation
/nav/hero-assets
/nav/background-patterns
/nav/update-notices
/nav/sayings
/game/games
/game/games/:id
/game/tags
/game/tag-categories
/game/comments
/game/prizes
/collection
/metrics
/changes
/system/cloud
/system/data-operations
/system/audit
/system/accounts
~~~

Site Group exposes a homepage curation page showing the first eight active sites and the remaining members. Operators move sites instead of entering weights. Group-oriented GET/PUT `/api/v1/nav/site-groups/:id/curation` uses `content.read`/`content.write`, a revision-checked complete member order, and the existing Nav transaction/audit/cache invalidation path. It persists only mapping weights; site-level bulk replacement preserves existing weights. Public derived caches refresh on the existing ten-minute schedule, so saving is not an immediate public-cache publication.

Site and Game are dedicated workspaces. Simple resources use the typed Resource Engine. Persistence mapping tables are managed as relationships inside workspaces, not exposed as primary navigation.

Collection, Metric, and Change reuse their existing business APIs and frozen Fact/Metric/Detector/Collection semantics. Operator-facing views require their read capabilities; schedule control, Metric technical contracts, and Change technical contracts additionally require their native capabilities.

Data Operations is a read-only `dataops.read` health center for the explicit `gfa`, `gfn`, and `gfg` pools. It may expose bounded PostgreSQL metadata, Goose state, and Top N relation sizes, but never DSNs, credentials, arbitrary SQL, migrations, rollback, maintenance, connection termination, or configuration mutation. The spelling `data_ops.read` is not an alias.

Audit requires `audit.read`, uses historical identity snapshots, returns real pagination, and redacts secret-shaped fields before responses. Accounts requires `account.manage`, keeps fixed roles, and delegates last-active-Owner protection to the backend transaction. Workbench aggregation is capability-shaped and projects existing durable attention sources; it is not an alerting subsystem.

Every authenticated account may change its own username or password through `/api/v1/auth/self/*` after current-password verification. These routes do not require or reuse `account.manage`; username changes preserve the current session and refresh the identity, while password changes increment `session_version`, clear the auth cookie, invalidate prior sessions, and require a new login. Both mutations write redacted GFA audit snapshots.

The React production build is the only writer of `apps/cn/admin/internal/transport/http/webui/dist`. It clears stale output before building, and the Go binary embeds that directory. Production must not require Node, a Vite server, a separate frontend service, manual asset copying, or a second Admin frontend implementation. The completed parity audit is recorded in `docs/admin-frontend-parity.md`.

Game content `groups` and `links` choose platform keys from Nav Web's data-only
`apps/cn/nav-web/app/data/platform-icons.json`, also consumed by `SiteIconList.vue`.
Keys cannot be chosen twice within one editor; existing key spelling is retained.
`resources` keeps free-text keys. Both Game summary fields accept at most 400
Unicode characters, with the same business validation in Admin and PostgreSQL.
Game primary/secondary tag searches request ten results with a 300 ms debounce;
the complete tag checklist loads paginated options once and filters locally,
independently of selected IDs.

Category and Tag resources create explicit stable codes with database-assigned IDs;
code controls are read-only after creation. Categories are selected by name and
normal removal is visibly archive/restore. Game classification saves weight,
nullable primary/secondary IDs and the complete Tag set in one
`PUT /api/v1/game/games/:id/classification` request; content saves do not overwrite
classification. The returned workspace includes the role union and resets the form.

## Collaboration Center

`/collaboration` has exactly two internal tabs: content ideas and one shared text Board, guarded by independent `collaboration.read/write`. Inventory lives only in GFA; formal content must never be created by batch import. The browser parses explicit `title | source | note` lines, while Go owns canonicalization and soft duplicate checks (maximum 500 candidates). All updates carry the displayed version; HTTP 409 requires explicit reload.

Game/Site create and formal workspaces preserve `?idea=`. Prefill only reliable Steam AppID or Site title/name, with no automatic Steam call or Collector Target/classification guess. Formal success survives GFA link failure; the workspace offers recovery using the existing resource. Existing options APIs own manual link search. Workbench inventory counts are neutral, never Attention backlog. The Board uses native pointer events and persists movement on pointerup, with 12-second polling and no additional board framework. See [the domain guide](../docs/collaboration-center.md).

Collaboration idea create/edit share a dialog. Long list cells truncate, narrow containers hide secondary columns, and secondary actions use a row menu. Confirmed idea deletion uses `collaboration.write` and the displayed version, records a GFA audit snapshot, and never deletes linked formal content. Existing-content options render in the link dialog’s scroll flow to avoid clipping. Batch paste uses visible pipe separators with examples, separate preview columns and line-numbered format errors; it does not parse Excel/TSV cells.
