# Admin role operator guide

The backend capability list is authoritative. This guide describes the fixed role responsibilities for acceptance and operations; frontend code must not reconstruct this table.

## Operator

- Manage normal Nav/Game content, Site/Game relationships, and simple resources.
- Use Workbench and global search.
- View Collection, Metrics, and Changes business views.
- Run Now, create manual collection work, and retry eligible runs.
- Cannot edit schedules, cancel work, open technical Metric/Change contracts, Data Operations, Audit, or Accounts.
- May manage Site icons, independent desktop/mobile Hero pools, and SVG patterns through content capabilities, but has no CloudOps capabilities or Cloud Resources access.

## Developer

- Has all Operator workflows.
- May edit/enable/disable schedules and cancel supported collection work.
- May open Metric/Change technical contracts, read-only Data Operations, and Audit.
- Cannot manage accounts or Owner governance.
- Has `cloudops.read` and `cloudops.manage`: may use `/system/cloud` for object inspection, COS-to-R2 repair, scoped EdgeOne/Cloudflare purges, and EdgeOne task history. Has no `cloudops.purge_all` and cannot purge the entire EdgeOne Zone.

## Owner

- Has all Developer workflows.
- Has all 16 capabilities, including Owner-only `cloudops.purge_all`. The full-zone EdgeOne action is separate from ordinary host purge, collapsed by default, and requires explicit scope confirmation.
- May list/create accounts, change display names and roles, enable/disable accounts, reset passwords, and revoke sessions.
- The backend forbids disabling or demoting the last active Owner; the UI displays that server rejection without masking it.

## Production acceptance

Every authenticated role may change its own username or password after current-password verification. Username changes preserve the session and refresh identity; password changes revoke prior sessions and require login again. These self-service actions do not require `account.manage`.

For every role, verify login/logout, visible navigation, direct-route guards, and backend `403` responses for forbidden APIs. Check System/Light/Dark at 1024, 1280, 1440, and 1920 widths. Core pages must show explicit loading, empty, error, success, and permission states. Mutations wait for server confirmation, refresh their queries, and report success or failure; destructive or elevated actions explain impact before submission.
