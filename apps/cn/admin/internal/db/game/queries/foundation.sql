-- name: FoundationPing :one
SELECT 1::bigint AS value;

-- name: NextGameID :one
SELECT nextval('gfg_game_id_seq')::bigint;

-- name: CountGames :one
SELECT COUNT(*)::bigint FROM gfg_game WHERE sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%'
 OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR info ILIKE '%'||sqlc.arg(keyword)||'%' OR info_en ILIKE '%'||sqlc.arg(keyword)||'%'
 OR id::text ILIKE '%'||sqlc.arg(keyword)||'%';

-- name: ListGames :many
SELECT id,name,name_en,info,info_en,create_time,update_time,resources,groups,developers,publishers,appid,header,links,weight,primary_tag,secondary_tag,view_count FROM gfg_game
WHERE sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%'
 OR info ILIKE '%'||sqlc.arg(keyword)||'%' OR info_en ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text ILIKE '%'||sqlc.arg(keyword)||'%'
ORDER BY id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetGame :one
SELECT id,name,name_en,info,info_en,create_time,update_time,resources,groups,developers,publishers,appid,header,links,weight,primary_tag,secondary_tag,view_count FROM gfg_game WHERE id=sqlc.arg(id);

-- name: FindGameByAppIDExcluding :one
SELECT id,name,appid FROM gfg_game WHERE appid=sqlc.arg(appid) AND id<>sqlc.arg(exclude_id) LIMIT 1;

-- name: InsertGame :one
INSERT INTO gfg_game (id,name,name_en,info,info_en,create_time,update_time,resources,groups,developers,publishers,appid,header,links,weight,primary_tag,secondary_tag,view_count)
VALUES (sqlc.arg(id),sqlc.arg(name),sqlc.arg(name_en),sqlc.arg(info),sqlc.arg(info_en),NOW()::timestamp(0),NOW()::timestamp(0),sqlc.arg(resources),sqlc.arg(groups),sqlc.arg(developers),sqlc.arg(publishers),sqlc.arg(appid),sqlc.arg(header),sqlc.arg(links),sqlc.arg(weight),sqlc.arg(primary_tag),sqlc.arg(secondary_tag),0)
RETURNING id,name,name_en,info,info_en,create_time,update_time,resources,groups,developers,publishers,appid,header,links,weight,primary_tag,secondary_tag,view_count;

-- name: UpdateGame :one
UPDATE gfg_game SET name=sqlc.arg(name),name_en=sqlc.arg(name_en),info=sqlc.arg(info),info_en=sqlc.arg(info_en),resources=sqlc.arg(resources),groups=sqlc.arg(groups),developers=sqlc.arg(developers),publishers=sqlc.arg(publishers),appid=sqlc.arg(appid),header=sqlc.arg(header),links=sqlc.arg(links),weight=sqlc.arg(weight),primary_tag=sqlc.arg(primary_tag),secondary_tag=sqlc.arg(secondary_tag),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id)
RETURNING id,name,name_en,info,info_en,create_time,update_time,resources,groups,developers,publishers,appid,header,links,weight,primary_tag,secondary_tag,view_count;

-- name: DeleteGame :execrows
DELETE FROM gfg_game WHERE id=sqlc.arg(id);

-- name: CountGameComments :one
SELECT COUNT(*)::bigint FROM gfg_game_comment WHERE sqlc.arg(keyword)::text='' OR content ILIKE '%'||sqlc.arg(keyword)||'%'
 OR region ILIKE '%'||sqlc.arg(keyword)||'%' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR ip ILIKE '%'||sqlc.arg(keyword)||'%'
 OR id::text ILIKE '%'||sqlc.arg(keyword)||'%' OR game_id::text ILIKE '%'||sqlc.arg(keyword)||'%';

-- name: ListGameComments :many
SELECT id,region,content,score,create_time,game_id,ip,name FROM gfg_game_comment
WHERE sqlc.arg(keyword)::text='' OR content ILIKE '%'||sqlc.arg(keyword)||'%' OR region ILIKE '%'||sqlc.arg(keyword)||'%'
 OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR ip ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text ILIKE '%'||sqlc.arg(keyword)||'%'
 OR game_id::text ILIKE '%'||sqlc.arg(keyword)||'%' ORDER BY id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetGameComment :one
SELECT id,region,content,score,create_time,game_id,ip,name FROM gfg_game_comment WHERE id=sqlc.arg(id);

-- name: InsertGameComment :one
INSERT INTO gfg_game_comment (id,region,content,score,create_time,game_id,ip,name)
VALUES (sqlc.arg(id),sqlc.arg(region),sqlc.arg(content),sqlc.arg(score),NOW()::timestamp(0),sqlc.arg(game_id),sqlc.arg(ip),sqlc.arg(name))
RETURNING id,region,content,score,create_time,game_id,ip,name;

-- name: UpdateGameComment :one
UPDATE gfg_game_comment SET region=sqlc.arg(region),content=sqlc.arg(content),score=sqlc.arg(score),game_id=sqlc.arg(game_id),ip=sqlc.arg(ip),name=sqlc.arg(name)
WHERE id=sqlc.arg(id) RETURNING id,region,content,score,create_time,game_id,ip,name;

-- name: DeleteGameComment :execrows
DELETE FROM gfg_game_comment WHERE id=sqlc.arg(id);

-- name: NextPrizeID :one
WITH lock_row AS MATERIALIZED (SELECT pg_advisory_xact_lock(hashtext('gfg_prize')::bigint))
SELECT (COALESCE(MAX(id),0)+1)::bigint FROM gfg_prize CROSS JOIN lock_row;

-- name: CountPrizes :one
SELECT COUNT(*)::bigint FROM gfg_prize WHERE sqlc.arg(keyword)::text='' OR title ILIKE '%'||sqlc.arg(keyword)||'%'
 OR "desc" ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text ILIKE '%'||sqlc.arg(keyword)||'%';

-- name: ListPrizes :many
SELECT id,title,"desc",prize,"key",start_time,end_time,create_time,status FROM gfg_prize
WHERE sqlc.arg(keyword)::text='' OR title ILIKE '%'||sqlc.arg(keyword)||'%' OR "desc" ILIKE '%'||sqlc.arg(keyword)||'%'
 OR id::text ILIKE '%'||sqlc.arg(keyword)||'%' ORDER BY id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetPrize :one
SELECT id,title,"desc",prize,"key",start_time,end_time,create_time,status FROM gfg_prize WHERE id=sqlc.arg(id);

-- name: InsertPrize :one
INSERT INTO gfg_prize (id,title,"desc",prize,"key",start_time,end_time,create_time,status)
VALUES (sqlc.arg(id),sqlc.arg(title),sqlc.arg(description),sqlc.arg(prize),sqlc.arg(key),sqlc.arg(start_time),sqlc.arg(end_time),NOW()::timestamp(0),sqlc.arg(status))
RETURNING id,title,"desc",prize,"key",start_time,end_time,create_time,status;

-- name: UpdatePrize :one
UPDATE gfg_prize SET title=sqlc.arg(title),"desc"=sqlc.arg(description),prize=sqlc.arg(prize),"key"=sqlc.arg(key),start_time=sqlc.arg(start_time),end_time=sqlc.arg(end_time),status=sqlc.arg(status)
WHERE id=sqlc.arg(id) RETURNING id,title,"desc",prize,"key",start_time,end_time,create_time,status;

-- name: DeletePrize :execrows
DELETE FROM gfg_prize WHERE id=sqlc.arg(id);

-- name: CountTags :one
SELECT COUNT(*)::bigint FROM gfg_tag WHERE sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%'
 OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR info ILIKE '%'||sqlc.arg(keyword)||'%' OR info_en ILIKE '%'||sqlc.arg(keyword)||'%'
 OR id::text ILIKE '%'||sqlc.arg(keyword)||'%';

-- name: ListTags :many
SELECT id,name,name_en,info,info_en,prefix,create_time,update_time FROM gfg_tag
WHERE sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%'
 OR info ILIKE '%'||sqlc.arg(keyword)||'%' OR info_en ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text ILIKE '%'||sqlc.arg(keyword)||'%'
ORDER BY id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetTag :one
SELECT id,name,name_en,info,info_en,prefix,create_time,update_time FROM gfg_tag WHERE id=sqlc.arg(id);

-- name: InsertTag :one
INSERT INTO gfg_tag (id,name,name_en,info,info_en,prefix,create_time,update_time)
VALUES (sqlc.arg(id),sqlc.arg(name),sqlc.arg(name_en),sqlc.arg(info),sqlc.arg(info_en),sqlc.arg(prefix),NOW()::timestamp(0),NOW()::timestamp(0))
RETURNING id,name,name_en,info,info_en,prefix,create_time,update_time;

-- name: UpdateTag :one
UPDATE gfg_tag SET name=sqlc.arg(name),name_en=sqlc.arg(name_en),info=sqlc.arg(info),info_en=sqlc.arg(info_en),prefix=sqlc.arg(prefix),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) RETURNING id,name,name_en,info,info_en,prefix,create_time,update_time;

-- name: DeleteTag :execrows
DELETE FROM gfg_tag WHERE id=sqlc.arg(id);

-- name: NextTagMapID :one
WITH lock_row AS MATERIALIZED (SELECT pg_advisory_xact_lock(hashtext('gfg_tag_map')::bigint))
SELECT (COALESCE(MAX(id),0)+1)::bigint FROM gfg_tag_map CROSS JOIN lock_row;

-- name: CountTagMaps :one
SELECT COUNT(*)::bigint FROM gfg_tag_map m LEFT JOIN gfg_game g ON g.id=m.game_id LEFT JOIN gfg_tag t ON t.id=m.tag_id
WHERE sqlc.arg(keyword)::text='' OR m.id::text ILIKE '%'||sqlc.arg(keyword)||'%' OR m.game_id::text ILIKE '%'||sqlc.arg(keyword)||'%'
 OR m.tag_id::text ILIKE '%'||sqlc.arg(keyword)||'%' OR COALESCE(g.name,'') ILIKE '%'||sqlc.arg(keyword)||'%'
 OR COALESCE(t.name,'') ILIKE '%'||sqlc.arg(keyword)||'%';

-- name: ListTagMaps :many
SELECT m.id,m.game_id,m.tag_id,m.create_time,m.update_time,COALESCE(g.name,'')::text AS game_name,COALESCE(t.name,'')::text AS tag_name
FROM gfg_tag_map m LEFT JOIN gfg_game g ON g.id=m.game_id LEFT JOIN gfg_tag t ON t.id=m.tag_id
WHERE sqlc.arg(keyword)::text='' OR m.id::text ILIKE '%'||sqlc.arg(keyword)||'%' OR m.game_id::text ILIKE '%'||sqlc.arg(keyword)||'%'
 OR m.tag_id::text ILIKE '%'||sqlc.arg(keyword)||'%' OR COALESCE(g.name,'') ILIKE '%'||sqlc.arg(keyword)||'%'
 OR COALESCE(t.name,'') ILIKE '%'||sqlc.arg(keyword)||'%' ORDER BY m.id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetTagMap :one
SELECT id,game_id,tag_id,create_time,update_time FROM gfg_tag_map WHERE id=sqlc.arg(id);

-- name: ListTagMapsByGame :many
SELECT id,game_id,tag_id,create_time,update_time FROM gfg_tag_map WHERE game_id=sqlc.arg(game_id) ORDER BY id ASC;

-- name: ListTagMapsByTag :many
SELECT id,game_id,tag_id,create_time,update_time FROM gfg_tag_map WHERE tag_id=sqlc.arg(tag_id) ORDER BY id ASC;

-- name: ListGameIDsByTag :many
SELECT game_id FROM gfg_tag_map WHERE tag_id=sqlc.arg(tag_id) ORDER BY id ASC;

-- name: InsertTagMap :one
INSERT INTO gfg_tag_map (id,game_id,tag_id,create_time,update_time)
VALUES (sqlc.arg(id),sqlc.arg(game_id),sqlc.arg(tag_id),NOW()::timestamp(0),NOW()::timestamp(0))
RETURNING id,game_id,tag_id,create_time,update_time;

-- name: UpdateTagMap :one
UPDATE gfg_tag_map SET game_id=sqlc.arg(game_id),tag_id=sqlc.arg(tag_id),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) RETURNING id,game_id,tag_id,create_time,update_time;

-- name: DeleteTagMap :execrows
DELETE FROM gfg_tag_map WHERE id=sqlc.arg(id);

-- name: DeleteTagMapsByGame :execrows
DELETE FROM gfg_tag_map WHERE game_id=sqlc.arg(game_id);

-- name: DeleteTagMapsByTagExceptGames :execrows
DELETE FROM gfg_tag_map WHERE tag_id=sqlc.arg(tag_id)
AND (cardinality(sqlc.arg(game_ids)::bigint[]) = 0 OR NOT (game_id = ANY(sqlc.arg(game_ids)::bigint[])));

-- name: CountGameOptions :one
SELECT COUNT(*)::bigint FROM gfg_game WHERE sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR appid::text ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text ILIKE '%'||sqlc.arg(keyword)||'%';

-- name: ListGameOptions :many
SELECT id,name,name_en,appid FROM gfg_game WHERE sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR appid::text ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text ILIKE '%'||sqlc.arg(keyword)||'%'
ORDER BY id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: ListTagOptions :many
SELECT id,name,name_en FROM gfg_tag WHERE sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text ILIKE '%'||sqlc.arg(keyword)||'%'
ORDER BY id DESC;
