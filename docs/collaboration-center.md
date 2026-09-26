# Admin Collaboration Center

Issue #117 adds `/collaboration` with two tabs: 内容想法池 and 共享画板. This is
shared content inventory, not a task tracker or a store of unfinished Games/Sites.
Owner, Developer and Operator currently receive `collaboration.read` and
`collaboration.write` from the compiled backend policy. React checks those
capabilities independently from content creation and account management.

## Ownership and persistence

Goose migration `db/admin/migrations/20260926000000_collaboration_center.sql`
adds content ideas and the initial note table. Migration
`20260926010000_collaboration_canvas.sql` renames/extends the note table into
`gfa_collaboration_board_node` and adds `gfa_collaboration_board_edge`. There is no
old-note API or dual persistence path. Account references use
GFA foreign keys with RESTRICT; formal resource references are logical links.
The historical baseline is unchanged; `tools/db-baseline/expected-final/gfa.json`
records the migrated schema. Data Operations expects all four GFA migrations,
ending at `20260926010000`; earlier migrations/baseline stay unchanged.

Collaboration writes only GFA. Its Game/Nav sqlc queries are read-only, checking
Steam AppIDs, Collector Target hostnames, link existence and card reference titles. Site duplicate
detection is best effort: a Site without a Collector Target cannot match by host.
There is no fuzzy name match, network lookup or cross-database transaction.
Formal resource deletion after a link check remains possible; links describe the
resource verified at the time of linking, not a cross-database foreign key.

Each single mutation and its Audit row commit in one GFA transaction. Batch
creation uses one GFA transaction, sqlc CopyFrom and one summary Audit row.
Audit failure rolls back the GFA write. Board position/size/z-index-only changes
increment version/updater/time without Audit. Node create/content/style/reference
updates/delete and edge create/update/delete are audited. Audit actions use
`collaboration.idea.*`, `collaboration.board_node.*` and
`collaboration.board_edge.*`. Node deletion audits its incident edges in the
same transaction before their foreign-key cascade.

## Ideas and batch import

Kinds are Game, Site and Other. Statuses are `idea` (储备), `researching` (整理中),
`landed` (已落地) and `shelved` (搁置). Priorities are normal/high. A title or source
is required; limits are 500 title characters, 2048 source characters and 10000
note characters. The default list includes idea + researching, with URL-backed
search, kind, status, priority, researcher and pagination. Researcher filters are
all/me/unassigned; the API joins display names without opening account management.

Batch input uses a visible pipe separator: `title | source | note`, one idea per
line. The note may be omitted. Empty title/source fields keep their separator,
for example `整理投稿规范 | | 补充说明`. A complete source may also occupy one line
(e.g. `550`, `steam:620`, or an HTTP(S) URL). Spaces belong to the field; they never
split columns. Literal pipes use `\|`, and literal backslashes may use `\\`.
Excel/TSV quoting and multiline cells are no longer parsed. Missing separators,
Tab columns, extra columns and rows with neither title nor source report the
physical input line number and block preview until corrected.

```text
求生之路 2 | steam:550 | 核对介绍
传送门 2 | steam:620 | 补充资料
test | steam:666 | 测试
```

The input placeholder contains a compact kind-specific example. Preview exposes
separate title/source/note columns so mistakes are visible before submission.
The explanatory panels and sample-fill button have been removed.
Select the batch kind explicitly. Preview and create accept 1–500 candidates.
The browser parses structure; Go canonicalizes and performs at most one duplicate
query in each database. Create always repeats the check. “只加入新内容” sends
`skip_known=true`; “全部加入” permits valid duplicates. Server-invalid rows are
skipped. Warnings distinguish duplicates in this batch, GFA matches and formal
resources. `source_key` has a non-unique index: concurrent imports can retain
duplicates. Removing Excel/TSV support follows the updated maintainer requirement;
it supersedes that input-format requirement in the initial #117 implementation brief.

Create and edit share the same dialog. Tables truncate long text; hover reveals
full values and the editor shows all fields. Smaller containers hide secondary
columns and keep edit and the “更多” menu reachable. Manual linking renders the
existing options search results inside the dialog's scroll flow.

“更多 → 删除想法” requires confirmation and the current version. It permanently
removes only the GFA idea, including duplicate/mistaken or landed ideas, and writes
`collaboration.idea.delete` with the before snapshot in the same transaction.
Linked formal Games/Sites are preserved. Use “搁置” to retain an idea for later.

Reliable Game sources are Steam Store `/app/<id>` URLs, case-insensitive
`steam:<id>` and numeric AppIDs; keys use `steam:<id>`. Site keys use `host:<host>`:
lowercase, remove port/trailing dot/one leading `www.`, preserve other subdomains.
Unrecognized sources remain useful text without a duplicate key. No new IDN or
public suffix processing is performed. Source links accept only HTTP(S).

All idea updates/transitions/deletes and Board update/delete require the displayed version.
Stale versions or inapplicable transitions return HTTP 409, never overwrite
silently. Research by the same account with the current version is idempotent;
another researcher returns a readable conflict. “仍然打开录入页面” only navigates.
Members may release another researcher after confirmation and Audit. There are
no leases, deadlines or automatic reassignment. Game/Site must link to a verified
formal resource to land; Other can land without a link. Reopen clears the link.

## API

Under `/api/v1/collaboration`, read capability protects `GET /summary`,
`GET /ideas`, `GET /ideas/:id`, `POST /ideas/batch-preview`, `GET /board`.
Write capability protects `POST /ideas`, `PUT /ideas/:id`, `DELETE /ideas/:id`, `POST /ideas/batch`,
`POST /ideas/:id/{research,release,shelve,restore,link,land,reopen}`, and
`POST /board/nodes`, `PUT /board/nodes/:id`, `DELETE /board/nodes/:id`,
`PUT /board/nodes/layout`, and `POST/PUT/DELETE /board/edges[/:id]`.
`GET /board` returns `{nodes, edges, references}`. Layout accepts
`{nodes: [{id, version, x, y, width, height, z_index}]}` with 1–500 unique IDs.
The entire gesture commits or returns 409/rolls back; unknown fields are rejected.
Node content updates cannot change geometry; edge updates cannot change endpoints.

Batch bodies are `{items: [...], skip_known: true|false}`. Transitions carry
`{version}`; link additionally carries `{kind: "game"|"site", resource_id}`.
Idea and Board deletion carry `{version}` in its JSON body. Clients cannot set actors,
status or link fields through ordinary idea updates. Actor IDs come from the
authenticated Principal. Existing session, CSRF and capability protections apply.

## Formal content handoff

Researching an idea can open `/game/games/new?idea=<id>` or
`/nav/sites/new?idea=<id>`. The context banner shows source, note, creator and
researcher. Game prefills only a reliable Steam AppID; Steam Prefill still needs
an explicit click. Site prefills only the idea title into name. Existing manual
classification/defaults remain; the idea URL never chooses Collector Target,
country, group, NSFW, TLS or proxy.

Formal creation uses the existing GameAPI/NavAPI. Only after success does React
call Collaboration link. If linking fails, it reports that formal content was
created, retains `?idea=` on the formal workspace, and offers “关联到当前内容” after
reloading the idea. It does not delete or recreate the formal resource. Manual
linking uses the existing `/api/v1/options/games` or `/api/v1/options/sites` picker.

Workbench shows neutral reserve/researching/landed-in-30-days counts when readable.
Inventory volume does not create Attention entries or backlog warnings.

## Shared Board

The single global canvas uses `@xyflow/react` 12.11.2 (React Flow), matching the
SagaFlow canvas library. It loads only when the Board tab is opened. The expanded
canvas scope supersedes the initial #117 text-only/no-connectors restriction.

- Nodes: notes, free/reference cards, text labels, rectangles, ellipses and arrows.
  Card references point to an existing idea, Game or Site; they do not land an
  idea or create/change formal content. Existing options APIs choose Game/Site;
  the existing idea list selects an idea. Referenced titles/status are live read
  projections, batched once per database, with a missing-content indication.
- Interaction: select/multi-select, pan/zoom, fit view, minimap, fullscreen,
  drag, resize, bring to front, double-click editor and confirmed deletion.
  Notes/cards have four connection handles; edges support curve/step routing,
  labels, colors and arrowheads. Node deletion removes its incident edges only.
- Saving: movement/resizing stays local until the gesture ends; keyboard node
  movement also persists. Multiple selected nodes save in one GFA transaction.
  Nodes and edges each own an optimistic version; unrelated edits do not share
  a global board revision. HTTP 409 preserves local drafts for explicit reload.
- Refresh: 12-second polling merges untouched nodes; a local movement remains
  protected even if another member removes its node. Referenced content is not
  a cross-database foreign key. No WebSocket, CRDT, image/file upload, multiple
  boards, comments or workflow engine is introduced.

Coordinates allow −100000 to 100000, dimensions 48–1600 × 40–1600, z-index
0–1000000, titles 200 characters and bodies 10000. Shape annotations can have
empty titles; notes/text/free cards need a title or body. Colors are a fixed
palette; arrow annotations rotate by quarter turns. Board node/edge styles,
content and references are audited; geometry is not.

## Verification and maintainer acceptance

Use the existing Admin Go and React checks, root sqlc/policy checks and isolated
PostgreSQL integration. With an explicitly isolated config:

```text
GOFURRY_ADMIN_INTEGRATION_CONFIG=/path/to/isolated.yaml go test ./internal/bootstrap -run TestAdminCollaborationThreeDatabase -count=1 -v
```

The integration covers 100-row inventory creation, duplicate reads through
read-only business connections, batch limits/skip behavior, links, transitions,
two simultaneous editors, HTTP 409, versioned idea deletion without deleting formal resources, Board node/edge versions, grouped movement rollback, bounded reference reads,
cascading edge deletion, Audit inclusion/exclusion and audit-failure rollback.
CI includes this test in its Admin three-database gate. React retains Vitest and
Testing Library; no Admin Playwright system is introduced.

Before closing #117, a maintainer should perform these checks in a non-production
acceptance environment with two accounts:

| Check | Expected result |
| --- | --- |
| Paste 100 Game ideas | GFA +100; GFG/GFN +0; public site has no partial content |
| Game idea → AppID → explicit Steam Prefill → complete create | One formal Game, then landed link |
| Site idea → complete create | One Site, no automatically created Collector Target, then landed link |
| Interrupt only the Collaboration link after formal creation | Clear success/link warning; same formal workspace retains idea; reload and recover without recreating/deleting content |
| Two accounts research/edit the same displayed version | Research ownership hint or HTTP 409; no silent overwrite |
| Delete duplicate or landed idea | Confirmation and Audit; formal resources retained; stale deletion returns 409 |
| Resize the window, edit an idea, search/link existing content | Ellipsis without table expansion; modal editor and visible search results |
| Canvas notes/reference cards/labels/shapes, drag/resize/multi-select, connect/edit/delete, fullscreen across accounts | One layout save per gesture, per-node/edge conflicts, cascading edge removal, formal resources retained, correct Audit inclusion/exclusion |

## Production upgrade

This change upgrades only Admin and GFA. **Back up GFA first**, coordinate/stop
Admin writes, **manually run Goose** from `db/admin/migrations` with the GFA
`gofurry_migrator` credential through `20260926010000`, then **deploy the new Admin binary** containing the
built React frontend. Start Admin and verify capabilities, ideas, board, handoff
and Audit. Do not run migration from Task or application startup. Game/Nav Backend,
Collectors, public Nav Web, GFG and GFN do not need a #117 deployment/migration.
The canvas migration renames the old note table, so the previous text-board binary
is not compatible with the upgraded schema. A rollback needs a coordinated GFA
backup restore and matching binary; do not improvise a destructive Down migration.
