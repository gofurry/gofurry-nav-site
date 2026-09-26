# Admin authorization contract

- Admin authentication uses an HTTPOnly cookie JWT containing only `account_id`, `session_version`, and standard registered claims. Role and capabilities are never trusted from the token.
- Every authenticated request resolves the current `gfa_admin_account`, requires `status=active`, and requires an exact session-version match before constructing one request-scoped Principal.
- The fixed roles are `owner`, `developer`, and `operator`. The database stores only the role assignment; the compiled Go policy owns Role-to-Capability mapping. Custom roles, permission tables, policy DSLs, and per-resource ACLs are outside the contract.
- Business routes authorize capabilities, never roles. Missing roles and unknown capabilities fail closed. Authentication failure returns `401`; capability denial returns `403`.
- CloudOps uses `cloudops.read` and `cloudops.manage` for Owner/Developer, while `cloudops.purge_all` belongs only to Owner. Operator keeps `content.write` for business asset management and receives no CloudOps permissions. Full-zone EdgeOne purge has a separate route; ordinary host purge cannot invoke it.
- Bootstrap is available only while the account count is zero and creates one active Owner. Disabled accounts never reopen bootstrap.
- Role changes, status changes, password resets, and explicit revocation increment `session_version`. Display-name changes do not.
- Every authenticated account can change its own username or password through `/api/v1/auth/self/*` after current-password verification, without `account.manage`. Username changes enforce canonical uniqueness, preserve the session, and refresh identity. Password changes increment `session_version`, clear the auth cookie, invalidate prior sessions, and require login again. Both actions write redacted audit snapshots.
- Disabling or demoting the last active Owner is forbidden under transaction-safe PostgreSQL row locking. Concurrent mutations must never leave zero active Owners.
- Account deletion is not exposed. Disable preserves audit identity.
- Audit rows retain the legacy `operator` field and also snapshot `operator_account_id`, `operator_name`, and `operator_role`. Snapshots remain interpretable after later account changes and never contain password hashes, tokens, cookies, or secrets.

- Collaboration uses independent `collaboration.read` and `collaboration.write`; all three fixed roles currently receive both. Source actors always come from Principal. These capabilities do not substitute for `content.write` when entering formal Game/Site creation.
