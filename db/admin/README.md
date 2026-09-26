# Admin database (`gfa`)

This directory exclusively owns the PostgreSQL `gfa` schema migrations. The
current-state baseline contains `gfa_admin_account`, `gfa_admin_audit_log`, and
their two sequences.

Admin runtime startup never creates or migrates these tables. The baseline has
no destructive Down section.

Migration `20260830030000` expands the historical singleton account into the
fixed-role multi-account identity model and adds durable audit identity
snapshots. Zero-account databases remain bootstrap-ready. A single legacy row
becomes the initial active Owner without changing its password, session, or
timestamps; more than one legacy row aborts migration rather than guessing
privilege.

Migration `20260926000000` adds GFA-only content ideas and the initial Board notes. `20260926010000` upgrades Board notes to canvas nodes and edges, retaining the earlier migration history. The node table is renamed/extended, so old text-board binaries cannot run against the upgraded schema. Version checks, account foreign keys and status/link constraints protect collaboration state. Source-key duplicates are intentionally allowed. Back up GFA and manually apply Goose before deploying the upgraded Admin; see [Collaboration Center](../../docs/collaboration-center.md).
