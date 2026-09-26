-- name: ListBoardNodes :many
SELECT * FROM gfa_collaboration_board_node ORDER BY z_index, id;

-- name: GetBoardNodeForUpdate :one
SELECT * FROM gfa_collaboration_board_node WHERE id=$1 FOR UPDATE;

-- name: InsertBoardNode :one
INSERT INTO gfa_collaboration_board_node (kind,title,body,color,rotation,reference_kind,reference_id,x,y,width,height,z_index,created_by_account_id,updated_by_account_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,sqlc.arg(account_id),sqlc.arg(account_id)) RETURNING *;

-- name: UpdateBoardNodeVersioned :one
UPDATE gfa_collaboration_board_node SET title=$1,body=$2,color=$3,rotation=$4,reference_kind=$5,reference_id=$6,
updated_by_account_id=$7,version=version+1,updated_at=CURRENT_TIMESTAMP WHERE id=$8 AND version=$9 RETURNING *;

-- name: MoveBoardNodeVersioned :one
UPDATE gfa_collaboration_board_node SET x=$1,y=$2,width=$3,height=$4,z_index=$5,
updated_by_account_id=$6,version=version+1,updated_at=CURRENT_TIMESTAMP WHERE id=$7 AND version=$8 RETURNING *;

-- name: DeleteBoardNodeVersioned :one
DELETE FROM gfa_collaboration_board_node WHERE id=$1 AND version=$2 RETURNING *;

-- name: ListBoardEdges :many
SELECT * FROM gfa_collaboration_board_edge ORDER BY id;

-- name: GetBoardEdgeForUpdate :one
SELECT * FROM gfa_collaboration_board_edge WHERE id=$1 FOR UPDATE;

-- name: ListIncidentBoardEdgesForUpdate :many
SELECT * FROM gfa_collaboration_board_edge WHERE source_id=$1 OR target_id=$1 ORDER BY id FOR UPDATE;

-- name: LockBoardEndpoints :many
SELECT id FROM gfa_collaboration_board_node WHERE id=ANY(sqlc.arg(ids)::bigint[]) AND kind IN ('note','card') ORDER BY id FOR KEY SHARE;

-- name: InsertBoardEdge :one
INSERT INTO gfa_collaboration_board_edge (source_id,target_id,source_handle,target_handle,routing,label,color,arrow,created_by_account_id,updated_by_account_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,sqlc.arg(account_id),sqlc.arg(account_id)) RETURNING *;

-- name: UpdateBoardEdgeVersioned :one
UPDATE gfa_collaboration_board_edge SET routing=$1,label=$2,color=$3,arrow=$4,
updated_by_account_id=$5,version=version+1,updated_at=CURRENT_TIMESTAMP WHERE id=$6 AND version=$7 RETURNING *;

-- name: DeleteBoardEdgeVersioned :one
DELETE FROM gfa_collaboration_board_edge WHERE id=$1 AND version=$2 RETURNING *;

-- name: ListBoardIdeaReferences :many
SELECT id,kind,COALESCE(title,source,'')::text AS title,status FROM gfa_content_idea WHERE id=ANY(sqlc.arg(ids)::bigint[]);
