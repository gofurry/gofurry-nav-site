-- name: FoundationPing :one
SELECT 1::bigint AS value;

-- name: CountAdminAccounts :one
SELECT COUNT(*)::bigint FROM gfa_admin_account;

-- name: GetAdminAccount :one
SELECT id, password_hash, session_version, created_at, updated_at, password_updated_at
FROM gfa_admin_account
WHERE id = 1;

-- name: InsertAdminAccount :one
INSERT INTO gfa_admin_account
    (id, password_hash, session_version, created_at, updated_at, password_updated_at)
VALUES
    (1, sqlc.arg(password_hash), sqlc.arg(session_version), NOW()::timestamp(0), NOW()::timestamp(0), sqlc.arg(password_updated_at))
RETURNING id, password_hash, session_version, created_at, updated_at, password_updated_at;

-- name: UpdateAdminAccountPassword :one
UPDATE gfa_admin_account
SET password_hash = sqlc.arg(password_hash),
    session_version = sqlc.arg(session_version),
    updated_at = NOW()::timestamp(0),
    password_updated_at = sqlc.arg(password_updated_at)
WHERE id = 1
RETURNING id, password_hash, session_version, created_at, updated_at, password_updated_at;

-- name: InsertAdminAuditLog :exec
INSERT INTO gfa_admin_audit_log
    (action, resource, target_id, operator, session_version, request_id, ip_address,
     user_agent, before_data, after_data, created_at)
VALUES
    (sqlc.arg(action), sqlc.arg(resource), sqlc.arg(target_id), sqlc.arg(operator),
     sqlc.arg(session_version), sqlc.arg(request_id), sqlc.arg(ip_address),
     sqlc.arg(user_agent), sqlc.arg(before_data), sqlc.arg(after_data), NOW()::timestamp(0));
