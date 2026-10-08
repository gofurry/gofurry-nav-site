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

Search inputs that update URL state, remote queries/options, or selection/submit
actions must be IME composition-safe. Composition start and intermediate changes
update local displayed text only; composition end commits the final text once,
including browsers that emit a trailing change. External values synchronize only
outside composition. Keyboard selection must not consume IME confirmation or
navigation keys while the composing ref or native `isComposing` is true. Debounce
starts from committed text and is not a substitute for this boundary. Use the small
`useCompositionSafeSearch` hook for these owners; DataTable retains its equivalent
contract. Ordinary local form fields and local-only filters do not need this guard.

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
/nav/update-notices/new
/nav/update-notices/:id
/nav/sayings
/game/games
/game/games/:id
/game/showcase
/game/showcase/:id
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

Site, Game, Release Notes and Showcase are dedicated workspaces. Simple resources use the typed Resource Engine. Persistence mapping tables are managed as relationships inside workspaces, not exposed as primary navigation.

Showcase consumes the frozen [backend contract](../docs/game-showcase.md) through
`content.read/write`. Its routes precede `/game/:resource`; composition, Campaign
list and discovery tabs, locale, filters and pagination use URL state. Creation
opens a real Draft ID workspace with overview/content/assets/schedule/stats.
Read-only users can inspect every tab without mutation controls. Lifecycle actions
and artwork clear require confirmation; dirty forms and staged uploads protect
navigation/unload and disable lifecycle actions. Archived Campaigns are read-only.

Campaign locale editors are independent zh/en RHF fields. Sponsored payloads clear
both Editorial Notes; action types remain fixed. Artwork uploads retain original
AVIF bytes and use the shared session/CSRF transport. Local object URLs are revoked;
persisted artwork shows provider-neutral keys, without guessed CDN URLs. Focal
controls live in Assets but save the complete `/content` payload. A nine-button
grid selects x/y at 0, 0.5 or 1 and highlights the nearest preset; staged local
preview clicks set clamped proportional coordinates and immediately update
object-position. Neither interaction auto-saves or changes image bytes. Native
preview buttons support keyboard centering; the grid is the precise keyboard
alternative. Numeric values are diagnostic TechnicalLabel text, not primary
inputs. Without a local preview the grid remains usable; read-only/busy controls
are disabled. Weight is labeled “选取权重” (1..10000), Pin “固定展示位置”
(自动排序/null or 第 1–4 位), preserving the backend selection/position semantics.
Scheduling uses
the shared DateTimePicker and explicit Asia/Shanghai RFC3339 conversion independent
of the workstation timezone. Backend readiness/diagnostics remain authoritative.

Statistics use backend totals, at most 366 inclusive dates, two count-series
ECharts and same-origin CSV links. HLL session totals are labeled estimates with
possible cross-day duplication. Lists never request per-row statistics. Discovery
shows internal evidence and links to Game Classification; only that existing form
writes `showcase_eligible` with its complete classification payload. The Campaign
Header has an “操作审计” action only for `auth.can('audit.read')`, after 返回活动
and before lifecycle actions. It opens `/system/audit?resource=gfg_showcase_campaign`
without claiming target-ID filtering. History and its intermediary component are
removed; old `tab=history` naturally renders overview without redirecting. Dirty
navigation protection also applies to the Header action.
Public Hero, public tracking, new cloud resolvers and backend changes are outside
this frontend workspace. See [React Admin](../docs/admin-react.md) for acceptance.

Discovery defaults to eligible candidates and exposes pending-approval, blocked,
and all-diagnostics groups with server counts. Keyword, status, exclusion, sorting,
pool, locale and pagination persist in URL state. The internal diagnostic API owns
whole-set filtering/sorting before pagination; React must not filter only a fetched
page or fetch every page to simulate global search. Pool order is explicitly not
homepage position. Technical evidence lives in an accessible diagnostic dialog;
read-only accounts get inspection links without mutation wording. Pending approval
never includes a Game blocked by any other condition. Empty eligible results offer
pending/all views instead of suggesting a service failure.

Lottery remains a simple Resource Engine resource. Activity title/description and
prize title/platform each have Chinese and English fields; participation passwords
and redemption keys are language-independent. String-array controls MUST retain
blank lines and whitespace during editing. Submit validation trims/removes empty
key lines and requires at least one usable key; backend normalization remains
authoritative. Do not normalize a controlled textarea on each keystroke.

Release Notes routes precede `/nav/:resource`; `update-notices` is not a generic
resource. The existing navigation and `content.read/write` capabilities apply.
One RHF form owns both languages and shared metadata. Opening `/new` creates no
record. Publish/Schedule must await a successful save of dirty content; Publish
Now omits the timestamp and Schedule sends the chosen future China-site time.
Dirty Unpublish is disabled; Delete and all publication changes require explicit
confirmation. Browser unload and internal routing protect unsaved work.
Status badges are display-only and interpret unzoned timestamps as Asia/Shanghai.
Markdown preview follows [the shared contract](update-markdown.md); it stores
source text, sanitizes all generated HTML and adds no upload or public-page owner.
The P3.1 writing workspace combines metadata and localized content in one Section.
It has no Markdown toolbar; Desktop editor/preview share equal panes, with a local
Edit/Preview switch on narrow layouts. Publication and dirty guards stay unchanged.

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
nullable primary/secondary IDs, the complete Tag set and independent
`showcase_eligible` in one
`PUT /api/v1/game/games/:id/classification` request; content saves do not overwrite
classification. The returned workspace includes the role union and resets the form.

## Game Collection curation

`/game/collections`, `/game/collections/new`, `/game/collections/:id` and
`/game/collections/home-curation` are native content workspaces, separate from
the singular `/collection` Collector Control Plane. Use backend `content.read/write`
and `audit.read`; no new capability. Existing Code is read-only. Members are configured as
a hybrid Composition through existing Tag/Game RemoteSelect owners.
GET/PUT `/:id/composition` owns full Tag rules, manual pins and exclusions; never
copy effective members into the manual payload or call the legacy members API
from this Workspace. Paused rules remain visible and removable, but cannot be
newly selected. Show automatic/manual/both sources and Backend counts. Pin removes
an exclusion; exclude removes a pin; unpin preserves any known automatic match.
Draft preview overlays explicit overrides on known matches only. Tag changes and
restored exclusions need Backend confirmation after saving; label pending state
and keep saved counts clearly distinct. Adult badges are inspection only.
No upload, NSFW setting, item order, sync scheduling/status/action or cache prose.

Content and Composition share a baseVersion. Own successful writes advance it while
preserving other local edits; background updates cannot overwrite dirty drafts or
silently rebase them. HTTP 409 retains the draft, presents an Alert and requires
explicit reload. All lifecycle writes require ConfirmAction and are disabled while
dirty. Archived content remains inspectable and only Restore is writable. Protect
content, members and Home drafts with useUnsavedChanges. Home sends all five slots
and its original placement revision; the fixed sixth entry is never in the payload.
Home pickers request published + home_eligible=true and exclude duplicate selections.
Keep all search owners IME-safe. Membership search/pagination are local (20 per page);
retain the complete draft and always submit all three canonical ID sets. More than
100 members is an informational hint, never a truncation or save limit. Games,
Tags and local member search remain independent. Existing ineligible Home slots
are visibly marked and kept in their original positions until explicitly changed.
Keep reload in header actions; clean reload is immediate, dirty discard requires
confirmation and reloads both drafts plus the server version. Conflicts never
auto-reload. Success toasts say only “已保存”; Home removal is reflected by home_slot.
Do not render persistent cache-refresh explanations or promise publication timing. Stage C public UI
is outside this contract. See [Game Collections](../docs/game-collections.md).

## Collaboration Center

`/collaboration` has exactly two internal tabs: content ideas and one shared canvas, guarded by independent `collaboration.read/write`. Inventory lives only in GFA; formal content must never be created by batch import. The browser parses explicit `title | source | note` lines, while Go owns canonicalization and soft duplicate checks (maximum 500 candidates). All updates carry the displayed version; HTTP 409 requires explicit reload.

Game/Site create and formal workspaces preserve `?idea=`. Prefill only reliable Steam AppID or Site title/name, with no automatic Steam call or Collector Target/classification guess. Formal success survives GFA link failure; the workspace offers recovery using the existing resource. Existing options APIs own manual link search. Workbench inventory counts are neutral, never Attention backlog. The Board uses React Flow (`@xyflow/react`), lazy-loaded separately, with notes/reference cards, basic annotations and labeled connections. Drag/resize persists at gesture end, multi-selection saves atomically, and 12-second polling preserves unsaved local drafts. Nodes/edges have independent versions; stale updates return 409. Only GFA persists canvas data; geometry changes omit Audit and all other writes share their Audit transaction. See [the domain guide](../docs/collaboration-center.md).

Collaboration idea create/edit share a dialog. Long list cells truncate, narrow containers hide secondary columns, and secondary actions use a row menu. Confirmed idea deletion uses `collaboration.write` and the displayed version, records a GFA audit snapshot, and never deletes linked formal content. Existing-content options render in the link dialog’s scroll flow to avoid clipping. Batch paste uses visible pipe separators with compact input placeholders, separate preview columns and line-numbered format errors; it does not parse Excel/TSV cells.
