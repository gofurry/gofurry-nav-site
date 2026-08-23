# Nav database (`gfn`)

This directory exclusively owns `gfn` schema migrations. The current-state
baseline contains the audited 12-table schema, including the three legacy but
active collector log tables and `gfn_log_update`.

Nav V1 compatibility is retained. Migration `20260823000001` removes only the
unused `gfn_log_update` table after the final runtime reference audit; the three
legacy-active collector log tables remain intact.

The baseline has no destructive Down section.
