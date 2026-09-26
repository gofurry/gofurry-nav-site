-- name: ListGamesByAppIDsForCollaboration :many
SELECT id, name, name_en, appid FROM gfg_game WHERE appid = ANY(sqlc.arg(appids)::bigint[]) ORDER BY id;

-- name: GetGameForCollaborationLink :one
SELECT id, name, name_en, appid FROM gfg_game WHERE id = $1;
