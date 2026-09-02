# Admin authorization contract

- Admin authentication uses an HTTPOnly cookie JWT containing only `account_id`, `session_version`, and standard registered claims. Role and capabilities are never trusted from the token.
- Every authenticated request resolves the current `gfa_admin_account`, requires `status=active`, and requires an exact session-version match before constructing one request-scoped Principal.
- The fixed roles are `owner`, `developer`, and `operator`. The database stores only the role assignment; the compiled Go policy owns Role-to-Capability mapping. Custom roles, permission tables, policy DSLs, and per-resource ACLs are outside the contract.
- Business routes authorize capabilities, never roles. Missing roles and unknown capabilities fail closed. Authentication failure returns `401`; capability denial returns `403`.
- Bootstrap is available only while the account count is zero and creates one active Owner. Disabled accounts never reopen bootstrap.
- Role changes, status changes, password resets, and explicit revocation increment `session_version`. Display-name changes do not. Usernames are immutable after account creation in P0.5.2-A.
- Disabling or demoting the last active Owner is forbidden under transaction-safe PostgreSQL row locking. Concurrent mutations must never leave zero active Owners.
- Account deletion is not exposed. Disable preserves audit identity.
- Audit rows retain the legacy `operator` field and also snapshot `operator_account_id`, `operator_name`, and `operator_role`. Snapshots remain interpretable after later account changes and never contain password hashes, tokens, cookies, or secrets.
