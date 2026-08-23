# Nav database (`gfn`)

This directory exclusively owns `gfn` schema migrations. The current-state
baseline contains the audited 12-table schema, including the three legacy but
active collector log tables and `gfn_log_update`.

Nav V1 compatibility is retained. `gfn_log_update` is removed only by the later
post-service-migration cleanup migration.

The baseline has no destructive Down section.
