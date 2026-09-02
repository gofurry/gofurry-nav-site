-- name: FoundationPing :one
SELECT 1::bigint AS value;

-- name: LockAdminBootstrap :one
SELECT pg_advisory_xact_lock(718210552001::bigint) IS NULL AS locked;

-- name: CountAdminAccounts :one
SELECT COUNT(*)::bigint FROM gfa_admin_account;

-- name: GetAdminAccountByID :one
SELECT id, username, display_name, role, status, password_hash, session_version,
       last_login_at, created_at, updated_at, password_updated_at
FROM gfa_admin_account
WHERE id = sqlc.arg(account_id);

-- name: GetAdminAccountByUsername :one
SELECT id, username, display_name, role, status, password_hash, session_version,
       last_login_at, created_at, updated_at, password_updated_at
FROM gfa_admin_account
WHERE lower(username) = lower(btrim(sqlc.arg(username)::text));

-- name: LockAdminAccountByID :one
SELECT id, username, display_name, role, status, password_hash, session_version,
       last_login_at, created_at, updated_at, password_updated_at
FROM gfa_admin_account
WHERE id = sqlc.arg(account_id)
FOR UPDATE;

-- name: LockActiveOwnerIDs :many
SELECT id
FROM gfa_admin_account
WHERE role = 'owner' AND status = 'active'
ORDER BY id
FOR UPDATE;

-- name: CountAdminAccountsFiltered :one
SELECT COUNT(*)::bigint
FROM gfa_admin_account
WHERE sqlc.arg(keyword)::text = ''
   OR username ILIKE '%' || sqlc.arg(keyword)::text || '%'
   OR display_name ILIKE '%' || sqlc.arg(keyword)::text || '%';

-- name: ListAdminAccounts :many
SELECT id, username, display_name, role, status, password_hash, session_version,
       last_login_at, created_at, updated_at, password_updated_at
FROM gfa_admin_account
WHERE sqlc.arg(keyword)::text = ''
   OR username ILIKE '%' || sqlc.arg(keyword)::text || '%'
   OR display_name ILIKE '%' || sqlc.arg(keyword)::text || '%'
ORDER BY username, id
LIMIT sqlc.arg(row_limit)::int OFFSET sqlc.arg(row_offset)::int;

-- name: InsertAdminAccount :one
INSERT INTO gfa_admin_account
    (username, display_name, role, status, password_hash, session_version,
     created_at, updated_at, password_updated_at)
VALUES
    (sqlc.arg(username), sqlc.arg(display_name), sqlc.arg(role), sqlc.arg(status),
     sqlc.arg(password_hash), sqlc.arg(session_version), NOW()::timestamp(0),
     NOW()::timestamp(0), sqlc.arg(password_updated_at))
RETURNING id, username, display_name, role, status, password_hash, session_version,
          last_login_at, created_at, updated_at, password_updated_at;

-- name: UpdateAdminLastLogin :one
UPDATE gfa_admin_account
SET last_login_at = NOW()::timestamp(0),
    updated_at = NOW()::timestamp(0)
WHERE id = sqlc.arg(account_id)
RETURNING id, username, display_name, role, status, password_hash, session_version,
          last_login_at, created_at, updated_at, password_updated_at;

-- name: UpdateAdminAccountDisplayName :one
UPDATE gfa_admin_account
SET display_name = sqlc.arg(display_name),
    updated_at = NOW()::timestamp(0)
WHERE id = sqlc.arg(account_id)
RETURNING id, username, display_name, role, status, password_hash, session_version,
          last_login_at, created_at, updated_at, password_updated_at;

-- name: UpdateAdminAccountRole :one
UPDATE gfa_admin_account
SET role = sqlc.arg(role),
    session_version = session_version + 1,
    updated_at = NOW()::timestamp(0)
WHERE id = sqlc.arg(account_id)
RETURNING id, username, display_name, role, status, password_hash, session_version,
          last_login_at, created_at, updated_at, password_updated_at;

-- name: UpdateAdminAccountStatus :one
UPDATE gfa_admin_account
SET status = sqlc.arg(status),
    session_version = session_version + 1,
    updated_at = NOW()::timestamp(0)
WHERE id = sqlc.arg(account_id)
RETURNING id, username, display_name, role, status, password_hash, session_version,
          last_login_at, created_at, updated_at, password_updated_at;

-- name: UpdateAdminAccountPassword :one
UPDATE gfa_admin_account
SET password_hash = sqlc.arg(password_hash),
    session_version = session_version + 1,
    updated_at = NOW()::timestamp(0),
    password_updated_at = sqlc.arg(password_updated_at)
WHERE id = sqlc.arg(account_id)
RETURNING id, username, display_name, role, status, password_hash, session_version,
          last_login_at, created_at, updated_at, password_updated_at;

-- name: IncrementAdminSessionVersion :one
UPDATE gfa_admin_account
SET session_version = session_version + 1,
    updated_at = NOW()::timestamp(0)
WHERE id = sqlc.arg(account_id)
RETURNING id, username, display_name, role, status, password_hash, session_version,
          last_login_at, created_at, updated_at, password_updated_at;

-- name: InsertAdminAuditLog :exec
INSERT INTO gfa_admin_audit_log
    (action, resource, target_id, operator, session_version, request_id, ip_address,
     user_agent, before_data, after_data, operator_account_id, operator_name,
     operator_role, created_at)
VALUES
    (sqlc.arg(action), sqlc.arg(resource), sqlc.arg(target_id), sqlc.arg(operator),
     sqlc.arg(session_version), sqlc.arg(request_id), sqlc.arg(ip_address),
     sqlc.arg(user_agent), sqlc.arg(before_data), sqlc.arg(after_data),
     sqlc.arg(operator_account_id), sqlc.arg(operator_name), sqlc.arg(operator_role),
     NOW()::timestamp(0));

-- name: CountAdminAuditLogs :one
SELECT COUNT(*)::bigint
FROM gfa_admin_audit_log
WHERE (sqlc.arg(operator_query)::text = ''
       OR operator ILIKE '%' || sqlc.arg(operator_query)::text || '%'
       OR operator_name ILIKE '%' || sqlc.arg(operator_query)::text || '%')
  AND (sqlc.arg(operator_role)::text = '' OR operator_role = sqlc.arg(operator_role)::text)
  AND (sqlc.arg(action)::text = '' OR action ILIKE '%' || sqlc.arg(action)::text || '%')
  AND (sqlc.arg(resource)::text = '' OR resource ILIKE '%' || sqlc.arg(resource)::text || '%')
  AND (sqlc.narg(from_time)::timestamp IS NULL OR created_at >= sqlc.narg(from_time)::timestamp)
  AND (sqlc.narg(until_time)::timestamp IS NULL OR created_at < sqlc.narg(until_time)::timestamp);

-- name: ListAdminAuditLogs :many
SELECT id, action, resource, target_id, operator, session_version, request_id,
       ip_address, user_agent, before_data, after_data, operator_account_id,
       operator_name, operator_role, created_at
FROM gfa_admin_audit_log
WHERE (sqlc.arg(operator_query)::text = ''
       OR operator ILIKE '%' || sqlc.arg(operator_query)::text || '%'
       OR operator_name ILIKE '%' || sqlc.arg(operator_query)::text || '%')
  AND (sqlc.arg(operator_role)::text = '' OR operator_role = sqlc.arg(operator_role)::text)
  AND (sqlc.arg(action)::text = '' OR action ILIKE '%' || sqlc.arg(action)::text || '%')
  AND (sqlc.arg(resource)::text = '' OR resource ILIKE '%' || sqlc.arg(resource)::text || '%')
  AND (sqlc.narg(from_time)::timestamp IS NULL OR created_at >= sqlc.narg(from_time)::timestamp)
  AND (sqlc.narg(until_time)::timestamp IS NULL OR created_at < sqlc.narg(until_time)::timestamp)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(row_limit)::int OFFSET sqlc.arg(row_offset)::int;
