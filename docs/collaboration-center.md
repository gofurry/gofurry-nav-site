# Admin Collaboration Center

Issue #117 adds `/collaboration` with two tabs: 内容想法池 and 共享画板. This is
shared content inventory, not a task tracker or a store of unfinished Games/Sites.
Owner, Developer and Operator currently receive `collaboration.read` and
`collaboration.write` from the compiled backend policy. React checks those
capabilities independently from content creation and account management.

## Ownership and persistence

Goose migration `db/admin/migrations/20260926000000_collaboration_center.sql`
adds `gfa_content_idea` and `gfa_collaboration_board_note`. Account references use
GFA foreign keys with RESTRICT; formal resource references are logical links.
The historical baseline is unchanged; `tools/db-baseline/expected-final/gfa.json`
records the migrated schema. Data Operations expects this migration version.

Collaboration writes only GFA. Its Game/Nav sqlc queries are read-only, checking
Steam AppIDs, Collector Target hostnames and link existence. Site duplicate
detection is best effort: a Site without a Collector Target cannot match by host.
There is no fuzzy name match, network lookup or cross-database transaction.
Formal resource deletion after a link check remains possible; links describe the
resource verified at the time of linking, not a cross-database foreign key.

Each single mutation and its Audit row commit in one GFA transaction. Batch
creation uses one GFA transaction, sqlc CopyFrom and one summary Audit row.
Audit failure rolls back the GFA write. Board position/size/z-index-only changes
increment version/updater/time without Audit; create, body changes and delete
are audited. Audit actions use `collaboration.idea.*` and
`collaboration.board_note.*`, with the corresponding resource names.

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

“填入示例” supplies a kind-specific example when input is empty. Preview exposes
separate title/source/note columns so mistakes are visible before submission.
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
`GET /ideas`, `GET /ideas/:id`, `POST /ideas/batch-preview`, `GET /board/notes`.
Write capability protects `POST /ideas`, `PUT /ideas/:id`, `DELETE /ideas/:id`, `POST /ideas/batch`,
`POST /ideas/:id/{research,release,shelve,restore,link,land,reopen}`, and
`POST /board/notes`, `PUT /board/notes/:id`, `DELETE /board/notes/:id`.

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

One global scrollable canvas starts at 4000 × 3000. Text notes support create,
explicit body save, drag, resize, bring-to-front and confirmed delete. Native
pointer events keep movement local until pointerup; pointer cancellation sends
no write. Arrow keys on the move/resize handles support 10-pixel steps. The board
polls every 12 seconds; local edits retain their original version until saved or
explicitly reloaded. Conflicts require reloading instead of automatic retries.
Positions are bounded to 0–20000, dimensions to 160–1600 × 120–1600, z-index to
0–1000000 and bodies to 10000 characters. There are no files, images, connectors,
multiple boards, comments, realtime protocol or external board framework.

## Verification and maintainer acceptance

Use the existing Admin Go and React checks, root sqlc/policy checks and isolated
PostgreSQL integration. With an explicitly isolated config:

```text
GOFURRY_ADMIN_INTEGRATION_CONFIG=/path/to/isolated.yaml go test ./internal/bootstrap -run TestAdminCollaborationThreeDatabase -count=1 -v
```

The integration covers 100-row inventory creation, duplicate reads through
read-only business connections, batch limits/skip behavior, links, transitions,
two simultaneous editors, HTTP 409, versioned idea deletion without deleting formal resources, Board versions/Audit and audit rollback (including idea deletion).
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
| Board create/body edit/drag/resize/raise/delete across accounts | Pointerup persistence, version conflicts, correct Audit inclusion/exclusion |

## Production upgrade

This change upgrades only Admin and GFA. **Back up GFA first**, coordinate/stop
Admin writes, **manually run Goose** from `db/admin/migrations` with the GFA
`gofurry_migrator` credential, then **deploy the new Admin binary** containing the
built React frontend. Start Admin and verify capabilities, ideas, board, handoff
and Audit. Do not run migration from Task or application startup. Game/Nav Backend,
Collectors, public Nav Web, GFG and GFN do not need a #117 deployment/migration.
Rollback of the binary should retain the additive GFA tables and their data;
do not improvise a destructive Down migration.
