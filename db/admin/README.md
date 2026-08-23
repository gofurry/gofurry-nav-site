# Admin database (`gfa`)

This directory exclusively owns the PostgreSQL `gfa` schema migrations. The
current-state baseline contains `gfa_admin_account`, `gfa_admin_audit_log`, and
their two sequences.

Admin runtime startup never creates or migrates these tables. The baseline has
no destructive Down section.
