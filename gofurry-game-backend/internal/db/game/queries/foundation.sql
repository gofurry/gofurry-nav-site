-- name: FoundationPing :one
SELECT 1::bigint AS value;

-- name: GetPrizeByID :one
SELECT id, title, "desc", prize, "key", start_time, end_time, create_time, status
FROM gfg_prize WHERE id = sqlc.arg(id);

-- name: GetPrizeMemberByEmail :one
SELECT id, prize_id, name, email, ip, agent, is_winner, prize_key, create_time
FROM gfg_prize_member
WHERE prize_id = sqlc.arg(prize_id) AND email = sqlc.arg(email);

-- name: ListActivePrizes :many
SELECT id, title, "desc", prize, "key", start_time, end_time, create_time, status
FROM gfg_prize WHERE status IS TRUE;

-- name: ListPrizeMembers :many
SELECT id, prize_id, name, email, ip, agent, is_winner, prize_key, create_time
FROM gfg_prize_member WHERE prize_id = sqlc.arg(prize_id);

-- name: ListPrizeHistory :many
SELECT id, title, "desc", end_time, prize
FROM gfg_prize WHERE status IS FALSE ORDER BY end_time DESC;

-- name: CountPrizeMembers :one
SELECT count(*)::bigint FROM gfg_prize_member WHERE prize_id = sqlc.arg(prize_id);

-- name: ListPrizeWinners :many
SELECT id, prize_id, name, email, ip, agent, is_winner, prize_key, create_time
FROM gfg_prize_member
WHERE prize_id = sqlc.arg(prize_id) AND is_winner IS TRUE;

-- name: ListActiveLotteries :many
SELECT id, title, "desc", start_time, end_time, prize
FROM gfg_prize WHERE status IS TRUE ORDER BY end_time DESC;

-- name: InsertPrizeMember :exec
INSERT INTO gfg_prize_member (
    id, prize_id, name, email, ip, agent, is_winner, prize_key, create_time
) VALUES (
    sqlc.arg(id), sqlc.arg(prize_id), sqlc.arg(name), sqlc.arg(email),
    sqlc.arg(ip), sqlc.arg(agent), sqlc.arg(is_winner), sqlc.narg(prize_key),
    sqlc.arg(create_time)
);

-- name: UpdatePrizeMemberWinner :execrows
UPDATE gfg_prize_member
SET is_winner = sqlc.arg(is_winner), prize_key = sqlc.narg(prize_key)
WHERE id = sqlc.arg(id);

-- name: SavePrize :execrows
UPDATE gfg_prize
SET title = sqlc.arg(title), "desc" = sqlc.arg(description), prize = sqlc.arg(prize)::jsonb,
    "key" = sqlc.arg(key), start_time = sqlc.arg(start_time),
    end_time = sqlc.arg(end_time), status = sqlc.arg(status)
WHERE id = sqlc.arg(id);

-- name: GetHotGames :many
SELECT c.game_id::text AS game_id, AVG(c.score)::double precision AS avg_score,
       COUNT(*)::bigint AS comment_count, COALESCE(g.name, '')::text AS name,
       COALESCE(g.name_en, '')::text AS name_en, COALESCE(g.info, '')::text AS info,
       COALESCE(g.info_en, '')::text AS info_en, COALESCE(g.header, '')::text AS header
FROM gfg_game_comment c
LEFT JOIN gfg_game g ON c.game_id = g.id
GROUP BY c.game_id, g.name, g.name_en, g.info, g.info_en, g.header
ORDER BY avg_score DESC
LIMIT sqlc.arg(limit_count);

-- name: GetGameScore :one
SELECT c.game_id::text AS game_id, AVG(c.score)::double precision AS avg_score,
       COUNT(*)::bigint AS comment_count, COALESCE(g.name, '')::text AS name,
       COALESCE(g.name_en, '')::text AS name_en, COALESCE(g.info, '')::text AS info,
       COALESCE(g.info_en, '')::text AS info_en, COALESCE(g.header, '')::text AS header
FROM gfg_game_comment c
LEFT JOIN gfg_game g ON c.game_id = g.id
WHERE c.game_id = sqlc.arg(game_id)
GROUP BY c.game_id, g.name, g.name_en, g.info, g.info_en, g.header;

-- name: GetReviewByIdentity :one
SELECT id, region, content, score, create_time, game_id, ip, name
FROM gfg_game_comment
WHERE ip = sqlc.arg(ip) AND game_id = sqlc.arg(game_id) AND name = sqlc.arg(name);

-- name: ListAnonymousReviews :many
SELECT c.region, c.score, c.content, c.ip, c.create_time AS time,
       COALESCE(CASE WHEN sqlc.arg(lang)::text = 'en' THEN g.name_en ELSE g.name END, '')::text AS game_name,
       COALESCE(g.header, '')::text AS game_cover
FROM gfg_game_comment c
LEFT JOIN gfg_game g ON c.game_id = g.id
ORDER BY c.create_time DESC
LIMIT sqlc.arg(limit_count);

-- name: InsertReview :exec
INSERT INTO gfg_game_comment (id, region, content, score, create_time, game_id, ip, name)
VALUES (sqlc.arg(id), sqlc.arg(region), sqlc.arg(content), sqlc.arg(score),
        sqlc.arg(create_time), sqlc.arg(game_id), sqlc.arg(ip), sqlc.arg(name));

-- name: GetGameViewCount :one
SELECT view_count FROM gfg_game WHERE id = sqlc.arg(id);

-- name: UpdateGameViewCount :execrows
UPDATE gfg_game SET view_count = sqlc.arg(view_count) WHERE id = sqlc.arg(id);
