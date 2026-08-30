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
