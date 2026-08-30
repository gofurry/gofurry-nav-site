-- name: FoundationPing :one
SELECT 1::bigint AS value;

-- name: NextSayingID :one
WITH lock_row AS MATERIALIZED (SELECT pg_advisory_xact_lock(hashtext('gfn_saying')::bigint))
SELECT (COALESCE(MAX(id), 0) + 1)::bigint FROM gfn_saying CROSS JOIN lock_row;

-- name: CountSayings :one
SELECT COUNT(*)::bigint FROM gfn_saying
WHERE sqlc.arg(keyword)::text = ''
   OR COALESCE(author, '') ILIKE '%' || sqlc.arg(keyword) || '%'
   OR language ILIKE '%' || sqlc.arg(keyword) || '%'
   OR saying ILIKE '%' || sqlc.arg(keyword) || '%'
   OR id::text ILIKE '%' || sqlc.arg(keyword) || '%';

-- name: ListSayings :many
SELECT id, author, saying, create_time, update_time, language FROM gfn_saying
WHERE sqlc.arg(keyword)::text = ''
   OR COALESCE(author, '') ILIKE '%' || sqlc.arg(keyword) || '%'
   OR language ILIKE '%' || sqlc.arg(keyword) || '%'
   OR saying ILIKE '%' || sqlc.arg(keyword) || '%'
   OR id::text ILIKE '%' || sqlc.arg(keyword) || '%'
ORDER BY id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetSaying :one
SELECT id, author, saying, create_time, update_time, language FROM gfn_saying WHERE id = sqlc.arg(id);

-- name: InsertSaying :one
INSERT INTO gfn_saying (id, author, saying, create_time, update_time, language)
VALUES (sqlc.arg(id), sqlc.arg(author), sqlc.arg(saying), NOW()::timestamp(0), NOW()::timestamp(0), sqlc.arg(language))
RETURNING id, author, saying, create_time, update_time, language;

-- name: UpdateSaying :one
UPDATE gfn_saying SET author=sqlc.arg(author), saying=sqlc.arg(saying), language=sqlc.arg(language), update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id)
RETURNING id, author, saying, create_time, update_time, language;

-- name: DeleteSaying :execrows
DELETE FROM gfn_saying WHERE id=sqlc.arg(id);

-- name: NextUpdateNoticeID :one
WITH lock_row AS MATERIALIZED (SELECT pg_advisory_xact_lock(hashtext('gfn_nav_update_notice')::bigint))
SELECT (COALESCE(MAX(id), 0) + 1)::bigint FROM gfn_nav_update_notice CROSS JOIN lock_row;

-- name: CountUpdateNotices :one
SELECT COUNT(*)::bigint FROM gfn_nav_update_notice WHERE deleted IS NOT TRUE AND (
    sqlc.arg(keyword)::text = '' OR title ILIKE '%'||sqlc.arg(keyword)||'%' OR title_en ILIKE '%'||sqlc.arg(keyword)||'%'
    OR body ILIKE '%'||sqlc.arg(keyword)||'%' OR body_en ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text ILIKE '%'||sqlc.arg(keyword)||'%');

-- name: ListUpdateNotices :many
SELECT id,title,title_en,body,body_en,published_at,create_time,update_time,deleted FROM gfn_nav_update_notice
WHERE deleted IS NOT TRUE AND (
    sqlc.arg(keyword)::text = '' OR title ILIKE '%'||sqlc.arg(keyword)||'%' OR title_en ILIKE '%'||sqlc.arg(keyword)||'%'
    OR body ILIKE '%'||sqlc.arg(keyword)||'%' OR body_en ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text ILIKE '%'||sqlc.arg(keyword)||'%')
ORDER BY published_at DESC,id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetUpdateNotice :one
SELECT id,title,title_en,body,body_en,published_at,create_time,update_time,deleted FROM gfn_nav_update_notice
WHERE id=sqlc.arg(id) AND deleted IS NOT TRUE;

-- name: GetUpdateNoticeAny :one
SELECT id,title,title_en,body,body_en,published_at,create_time,update_time,deleted FROM gfn_nav_update_notice WHERE id=sqlc.arg(id);

-- name: InsertUpdateNotice :one
INSERT INTO gfn_nav_update_notice (id,title,title_en,body,body_en,published_at,create_time,update_time,deleted)
VALUES (sqlc.arg(id),sqlc.arg(title),sqlc.arg(title_en),sqlc.arg(body),sqlc.arg(body_en),sqlc.arg(published_at),NOW()::timestamp(0),NOW()::timestamp(0),false)
RETURNING id,title,title_en,body,body_en,published_at,create_time,update_time,deleted;

-- name: UpdateUpdateNotice :one
UPDATE gfn_nav_update_notice SET title=sqlc.arg(title),title_en=sqlc.arg(title_en),body=sqlc.arg(body),body_en=sqlc.arg(body_en),published_at=sqlc.arg(published_at),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) AND deleted IS NOT TRUE
RETURNING id,title,title_en,body,body_en,published_at,create_time,update_time,deleted;

-- name: SoftDeleteUpdateNotice :one
UPDATE gfn_nav_update_notice SET deleted=true,update_time=NOW()::timestamp(0) WHERE id=sqlc.arg(id)
RETURNING id,title,title_en,body,body_en,published_at,create_time,update_time,deleted;

-- name: NextCollectorDomainID :one
WITH lock_row AS MATERIALIZED (SELECT pg_advisory_xact_lock(hashtext('gfn_collector_domain')::bigint))
SELECT (COALESCE(MAX(id), 0) + 1)::bigint FROM gfn_collector_domain CROSS JOIN lock_row;

-- name: CountCollectorDomains :one
SELECT COUNT(*)::bigint FROM gfn_collector_domain cd LEFT JOIN gfn_site s ON s.id=cd.site_id
WHERE cd.deleted IS NOT TRUE AND (sqlc.arg(keyword)::text='' OR cd.name ILIKE '%'||sqlc.arg(keyword)||'%'
 OR cd.proxy ILIKE '%'||sqlc.arg(keyword)||'%' OR cd.tls ILIKE '%'||sqlc.arg(keyword)||'%'
 OR COALESCE(s.name,'') ILIKE '%'||sqlc.arg(keyword)||'%' OR COALESCE(s.name_en,'') ILIKE '%'||sqlc.arg(keyword)||'%'
 OR cd.id::text ILIKE '%'||sqlc.arg(keyword)||'%' OR cd.site_id::text ILIKE '%'||sqlc.arg(keyword)||'%');

-- name: ListCollectorDomains :many
SELECT cd.id,cd.site_id,COALESCE(s.name,'')::text AS site_name,cd.name,cd.proxy,cd.prefix,cd.tls,cd.deleted
FROM gfn_collector_domain cd LEFT JOIN gfn_site s ON s.id=cd.site_id
WHERE cd.deleted IS NOT TRUE AND (sqlc.arg(keyword)::text='' OR cd.name ILIKE '%'||sqlc.arg(keyword)||'%'
 OR cd.proxy ILIKE '%'||sqlc.arg(keyword)||'%' OR cd.tls ILIKE '%'||sqlc.arg(keyword)||'%'
 OR COALESCE(s.name,'') ILIKE '%'||sqlc.arg(keyword)||'%' OR COALESCE(s.name_en,'') ILIKE '%'||sqlc.arg(keyword)||'%'
 OR cd.id::text ILIKE '%'||sqlc.arg(keyword)||'%' OR cd.site_id::text ILIKE '%'||sqlc.arg(keyword)||'%')
ORDER BY cd.id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetCollectorDomain :one
SELECT cd.id,cd.site_id,COALESCE(s.name,'')::text AS site_name,cd.name,cd.proxy,cd.prefix,cd.tls,cd.deleted
FROM gfn_collector_domain cd LEFT JOIN gfn_site s ON s.id=cd.site_id WHERE cd.id=sqlc.arg(id) AND cd.deleted IS NOT TRUE;

-- name: GetCollectorDomainAny :one
SELECT id,name,proxy,prefix,tls,site_id,deleted FROM gfn_collector_domain WHERE id=sqlc.arg(id);

-- name: InsertCollectorDomain :one
INSERT INTO gfn_collector_domain (id,name,proxy,prefix,tls,site_id,deleted)
VALUES (sqlc.arg(id),sqlc.arg(name),sqlc.arg(proxy),sqlc.arg(prefix),sqlc.arg(tls),sqlc.arg(site_id),false)
RETURNING id,name,proxy,prefix,tls,site_id,deleted;

-- name: UpdateCollectorDomain :one
UPDATE gfn_collector_domain SET name=sqlc.arg(name),proxy=sqlc.arg(proxy),prefix=sqlc.arg(prefix),tls=sqlc.arg(tls),site_id=sqlc.arg(site_id)
WHERE id=sqlc.arg(id) AND deleted IS NOT TRUE RETURNING id,name,proxy,prefix,tls,site_id,deleted;

-- name: SoftDeleteCollectorDomain :one
UPDATE gfn_collector_domain SET deleted=true WHERE id=sqlc.arg(id) RETURNING id,name,proxy,prefix,tls,site_id,deleted;

-- name: CountActiveSiteByID :one
SELECT COUNT(*)::bigint FROM gfn_site WHERE id=sqlc.arg(id) AND deleted IS NOT TRUE;

-- name: NextSiteID :one
WITH lock_row AS MATERIALIZED (SELECT pg_advisory_xact_lock(hashtext('gfn_site')::bigint))
SELECT (COALESCE(MAX(id), 0) + 1)::bigint FROM gfn_site CROSS JOIN lock_row;

-- name: CountSites :one
SELECT COUNT(*)::bigint FROM gfn_site WHERE deleted IS NOT TRUE AND (sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%'
 OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR info ILIKE '%'||sqlc.arg(keyword)||'%' OR info_en ILIKE '%'||sqlc.arg(keyword)||'%'
 OR id::text ILIKE '%'||sqlc.arg(keyword)||'%');

-- name: ListSites :many
SELECT id,name,name_en,info,info_en,create_time,update_time,country,nsfw,welfare,icon,deleted,view_count,deleted_at FROM gfn_site
WHERE deleted IS NOT TRUE AND (sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%'
 OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR info ILIKE '%'||sqlc.arg(keyword)||'%' OR info_en ILIKE '%'||sqlc.arg(keyword)||'%'
 OR id::text ILIKE '%'||sqlc.arg(keyword)||'%') ORDER BY update_time DESC,id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetSite :one
SELECT id,name,name_en,info,info_en,create_time,update_time,country,nsfw,welfare,icon,deleted,view_count,deleted_at FROM gfn_site
WHERE id=sqlc.arg(id) AND deleted IS NOT TRUE;

-- name: ListSiteWorkspaceTargets :many
SELECT domain.id,domain.site_id,domain.name,domain.proxy,domain.prefix,domain.tls,
       EXISTS (
           SELECT 1 FROM gfn_site_primary_target_periods primary_period
           JOIN gfn_target_tracking_periods target_period ON target_period.id=primary_period.target_tracking_period_id
           WHERE primary_period.site_id=domain.site_id AND primary_period.effective_until IS NULL
             AND target_period.collector_domain_id=domain.id
       ) AS is_primary
FROM gfn_collector_domain domain
WHERE domain.site_id=sqlc.arg(site_id) AND domain.deleted IS NOT TRUE ORDER BY domain.id ASC;

-- name: ListSiteWorkspaceGroups :many
SELECT m.id,m.site_id,m.group_id,COALESCE(g.name,'')::text AS group_name,m.weight
FROM gfn_site_group_map m LEFT JOIN gfn_site_group g ON g.id=m.group_id
WHERE m.site_id=sqlc.arg(site_id) ORDER BY m.id ASC;

-- name: GetSiteWorkspaceFeatured :one
SELECT id,site_id,weight FROM gfn_featured_site WHERE site_id=sqlc.arg(site_id) ORDER BY id ASC LIMIT 1;

-- name: ListSiteWorkspaceSummaries :many
SELECT site.id,site.name,site.name_en,site.update_time,
       COALESCE(primary_target.target,'')::text AS primary_target,
       COALESCE(site_groups.group_names,ARRAY[]::text[])::text[] AS group_names,
       EXISTS (SELECT 1 FROM gfn_featured_site featured WHERE featured.site_id=site.id) AS featured
FROM gfn_site site
LEFT JOIN LATERAL (
    SELECT target_period.target
    FROM gfn_site_primary_target_periods primary_period
    JOIN gfn_target_tracking_periods target_period ON target_period.id=primary_period.target_tracking_period_id
    WHERE primary_period.site_id=site.id AND primary_period.effective_until IS NULL
    LIMIT 1
) primary_target ON true
LEFT JOIN LATERAL (
    SELECT array_agg(site_group.name ORDER BY site_group.priority DESC,site_group.id DESC)::text[] AS group_names
    FROM gfn_site_group_map site_map
    JOIN gfn_site_group site_group ON site_group.id=site_map.group_id
    WHERE site_map.site_id=site.id
) site_groups ON true
WHERE site.deleted IS NOT TRUE AND (sqlc.arg(keyword)::text='' OR site.name ILIKE '%'||sqlc.arg(keyword)||'%'
 OR site.name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR site.info ILIKE '%'||sqlc.arg(keyword)||'%'
 OR site.info_en ILIKE '%'||sqlc.arg(keyword)||'%' OR site.id::text ILIKE '%'||sqlc.arg(keyword)||'%')
ORDER BY site.update_time DESC,site.id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetSiteAny :one
SELECT id,name,name_en,info,info_en,create_time,update_time,country,nsfw,welfare,icon,deleted,view_count,deleted_at FROM gfn_site WHERE id=sqlc.arg(id);

-- name: InsertSite :one
INSERT INTO gfn_site (id,name,name_en,info,info_en,create_time,update_time,country,nsfw,welfare,icon,deleted,view_count)
VALUES (sqlc.arg(id),sqlc.arg(name),sqlc.arg(name_en),sqlc.arg(info),sqlc.arg(info_en),NOW()::timestamp(0),NOW()::timestamp(0),sqlc.arg(country),sqlc.arg(nsfw),sqlc.arg(welfare),sqlc.arg(icon),false,0)
RETURNING id,name,name_en,info,info_en,create_time,update_time,country,nsfw,welfare,icon,deleted,view_count,deleted_at;

-- name: UpdateSite :one
UPDATE gfn_site SET name=sqlc.arg(name),name_en=sqlc.arg(name_en),info=sqlc.arg(info),info_en=sqlc.arg(info_en),country=sqlc.arg(country),nsfw=sqlc.arg(nsfw),welfare=sqlc.arg(welfare),icon=sqlc.arg(icon),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) AND deleted IS NOT TRUE
RETURNING id,name,name_en,info,info_en,create_time,update_time,country,nsfw,welfare,icon,deleted,view_count,deleted_at;

-- name: SoftDeleteSite :one
UPDATE gfn_site
SET deleted=true,deleted_at=transaction_timestamp(),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id)
RETURNING id,name,name_en,info,info_en,create_time,update_time,country,nsfw,welfare,icon,deleted,view_count,deleted_at;

-- name: NextSiteGroupID :one
WITH lock_row AS MATERIALIZED (SELECT pg_advisory_xact_lock(hashtext('gfn_site_group')::bigint))
SELECT (COALESCE(MAX(id), 0) + 1)::bigint FROM gfn_site_group CROSS JOIN lock_row;

-- name: CountSiteGroups :one
SELECT COUNT(*)::bigint FROM gfn_site_group WHERE sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%'
 OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR info ILIKE '%'||sqlc.arg(keyword)||'%' OR info_en ILIKE '%'||sqlc.arg(keyword)||'%'
 OR id::text ILIKE '%'||sqlc.arg(keyword)||'%';

-- name: ListSiteGroups :many
SELECT id,name,name_en,info,info_en,priority,create_time,update_time FROM gfn_site_group
WHERE sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%'
 OR info ILIKE '%'||sqlc.arg(keyword)||'%' OR info_en ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text ILIKE '%'||sqlc.arg(keyword)||'%'
ORDER BY priority DESC,id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetSiteGroup :one
SELECT id,name,name_en,info,info_en,priority,create_time,update_time FROM gfn_site_group WHERE id=sqlc.arg(id);

-- name: InsertSiteGroup :one
INSERT INTO gfn_site_group (id,name,name_en,info,info_en,priority,create_time,update_time)
VALUES (sqlc.arg(id),sqlc.arg(name),sqlc.arg(name_en),sqlc.arg(info),sqlc.arg(info_en),sqlc.arg(priority),NOW()::timestamp(0),NOW()::timestamp(0))
RETURNING id,name,name_en,info,info_en,priority,create_time,update_time;

-- name: UpdateSiteGroup :one
UPDATE gfn_site_group SET name=sqlc.arg(name),name_en=sqlc.arg(name_en),info=sqlc.arg(info),info_en=sqlc.arg(info_en),priority=sqlc.arg(priority),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) RETURNING id,name,name_en,info,info_en,priority,create_time,update_time;

-- name: DeleteSiteGroup :execrows
DELETE FROM gfn_site_group WHERE id=sqlc.arg(id);

-- name: NextSiteGroupMapID :one
WITH lock_row AS MATERIALIZED (SELECT pg_advisory_xact_lock(hashtext('gfn_site_group_map')::bigint))
SELECT (COALESCE(MAX(id), 0) + 1)::bigint FROM gfn_site_group_map CROSS JOIN lock_row;

-- name: CountSiteGroupMaps :one
SELECT COUNT(*)::bigint FROM gfn_site_group_map m LEFT JOIN gfn_site s ON s.id=m.site_id LEFT JOIN gfn_site_group g ON g.id=m.group_id
WHERE sqlc.arg(keyword)::text='' OR m.id::text ILIKE '%'||sqlc.arg(keyword)||'%' OR m.site_id::text ILIKE '%'||sqlc.arg(keyword)||'%'
 OR m.group_id::text ILIKE '%'||sqlc.arg(keyword)||'%' OR COALESCE(s.name,'') ILIKE '%'||sqlc.arg(keyword)||'%' OR COALESCE(g.name,'') ILIKE '%'||sqlc.arg(keyword)||'%';

-- name: ListSiteGroupMaps :many
SELECT m.id,m.site_id,m.group_id,COALESCE(s.name,'')::text AS site_name,COALESCE(g.name,'')::text AS group_name,m.weight,m.create_time,m.update_time
FROM gfn_site_group_map m LEFT JOIN gfn_site s ON s.id=m.site_id LEFT JOIN gfn_site_group g ON g.id=m.group_id
WHERE sqlc.arg(keyword)::text='' OR m.id::text ILIKE '%'||sqlc.arg(keyword)||'%' OR m.site_id::text ILIKE '%'||sqlc.arg(keyword)||'%'
 OR m.group_id::text ILIKE '%'||sqlc.arg(keyword)||'%' OR COALESCE(s.name,'') ILIKE '%'||sqlc.arg(keyword)||'%' OR COALESCE(g.name,'') ILIKE '%'||sqlc.arg(keyword)||'%'
ORDER BY m.id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetSiteGroupMap :one
SELECT id,site_id,group_id,create_time,update_time,weight FROM gfn_site_group_map WHERE id=sqlc.arg(id);

-- name: ListSiteGroupMapsBySite :many
SELECT id,site_id,group_id,create_time,update_time,weight FROM gfn_site_group_map WHERE site_id=sqlc.arg(site_id) ORDER BY id ASC;

-- name: InsertSiteGroupMap :one
INSERT INTO gfn_site_group_map (id,site_id,group_id,create_time,update_time,weight)
VALUES (sqlc.arg(id),sqlc.arg(site_id),sqlc.arg(group_id),NOW()::timestamp(0),NOW()::timestamp(0),sqlc.arg(weight))
RETURNING id,site_id,group_id,create_time,update_time,weight;

-- name: UpdateSiteGroupMap :one
UPDATE gfn_site_group_map SET site_id=sqlc.arg(site_id),group_id=sqlc.arg(group_id),weight=sqlc.arg(weight),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) RETURNING id,site_id,group_id,create_time,update_time,weight;

-- name: DeleteSiteGroupMap :execrows
DELETE FROM gfn_site_group_map WHERE id=sqlc.arg(id);

-- name: DeleteSiteGroupMapsBySite :execrows
DELETE FROM gfn_site_group_map WHERE site_id=sqlc.arg(site_id);

-- name: NextFeaturedSiteID :one
WITH lock_row AS MATERIALIZED (SELECT pg_advisory_xact_lock(hashtext('gfn_featured_site')::bigint))
SELECT (COALESCE(MAX(id), 0) + 1)::bigint FROM gfn_featured_site CROSS JOIN lock_row;

-- name: CountFeaturedSiteBySite :one
SELECT COUNT(*)::bigint FROM gfn_featured_site WHERE site_id=sqlc.arg(site_id) AND id<>sqlc.arg(exclude_id);

-- name: CountFeaturedSites :one
SELECT COUNT(*)::bigint FROM gfn_featured_site f LEFT JOIN gfn_site s ON s.id=f.site_id
WHERE sqlc.arg(keyword)::text='' OR f.id::text ILIKE '%'||sqlc.arg(keyword)||'%' OR f.site_id::text ILIKE '%'||sqlc.arg(keyword)||'%'
 OR COALESCE(s.name,'') ILIKE '%'||sqlc.arg(keyword)||'%' OR COALESCE(s.name_en,'') ILIKE '%'||sqlc.arg(keyword)||'%';

-- name: ListFeaturedSites :many
SELECT f.id,f.site_id,COALESCE(s.name,'')::text AS site_name,f.weight,f.create_time,f.update_time FROM gfn_featured_site f LEFT JOIN gfn_site s ON s.id=f.site_id
WHERE sqlc.arg(keyword)::text='' OR f.id::text ILIKE '%'||sqlc.arg(keyword)||'%' OR f.site_id::text ILIKE '%'||sqlc.arg(keyword)||'%'
 OR COALESCE(s.name,'') ILIKE '%'||sqlc.arg(keyword)||'%' OR COALESCE(s.name_en,'') ILIKE '%'||sqlc.arg(keyword)||'%'
ORDER BY f.weight DESC,f.id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetFeaturedSite :one
SELECT f.id,f.site_id,COALESCE(s.name,'')::text AS site_name,f.weight,f.create_time,f.update_time FROM gfn_featured_site f LEFT JOIN gfn_site s ON s.id=f.site_id WHERE f.id=sqlc.arg(id);

-- name: GetFeaturedSiteAny :one
SELECT id,site_id,weight,create_time,update_time FROM gfn_featured_site WHERE id=sqlc.arg(id);

-- name: InsertFeaturedSite :one
INSERT INTO gfn_featured_site (id,site_id,weight,create_time,update_time)
VALUES (sqlc.arg(id),sqlc.arg(site_id),sqlc.arg(weight),NOW()::timestamp(0),NOW()::timestamp(0))
RETURNING id,site_id,weight,create_time,update_time;

-- name: UpdateFeaturedSite :one
UPDATE gfn_featured_site SET site_id=sqlc.arg(site_id),weight=sqlc.arg(weight),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) RETURNING id,site_id,weight,create_time,update_time;

-- name: DeleteFeaturedSite :execrows
DELETE FROM gfn_featured_site WHERE id=sqlc.arg(id);

-- name: CountSiteOptions :one
SELECT COUNT(*)::bigint FROM gfn_site site WHERE deleted IS NOT TRUE AND (
    sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text ILIKE '%'||sqlc.arg(keyword)||'%'
    OR EXISTS (SELECT 1 FROM gfn_collector_domain target WHERE target.site_id=site.id AND target.deleted IS NOT TRUE AND lower(COALESCE(target.prefix,'')||target.name) ILIKE '%'||lower(sqlc.arg(keyword))||'%')
);

-- name: ListSiteOptions :many
SELECT id,name,name_en FROM gfn_site site WHERE deleted IS NOT TRUE AND (
    sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text ILIKE '%'||sqlc.arg(keyword)||'%'
    OR EXISTS (SELECT 1 FROM gfn_collector_domain target WHERE target.site_id=site.id AND target.deleted IS NOT TRUE AND lower(COALESCE(target.prefix,'')||target.name) ILIKE '%'||lower(sqlc.arg(keyword))||'%')
)
ORDER BY id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: CountSiteTargetOptions :one
SELECT count(*)::bigint
FROM gfn_collector_domain target
JOIN gfn_site site ON site.id = target.site_id AND site.deleted IS NOT TRUE
WHERE target.site_id = sqlc.arg(site_id) AND target.deleted IS NOT TRUE
  AND (sqlc.arg(keyword)::text = '' OR lower(COALESCE(target.prefix,'')||target.name) ILIKE '%'||lower(sqlc.arg(keyword))||'%');

-- name: ListSiteTargetOptions :many
SELECT target.id, lower(COALESCE(target.prefix,'')||target.name)::text AS target, target.proxy, target.tls
FROM gfn_collector_domain target
JOIN gfn_site site ON site.id = target.site_id AND site.deleted IS NOT TRUE
WHERE target.site_id = sqlc.arg(site_id) AND target.deleted IS NOT TRUE
  AND (sqlc.arg(keyword)::text = '' OR lower(COALESCE(target.prefix,'')||target.name) ILIKE '%'||lower(sqlc.arg(keyword))||'%')
ORDER BY target.id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: CountSiteGroupOptions :one
SELECT COUNT(*)::bigint FROM gfn_site_group WHERE sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text ILIKE '%'||sqlc.arg(keyword)||'%';

-- name: ListSiteGroupOptions :many
SELECT id,name,name_en FROM gfn_site_group WHERE sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text ILIKE '%'||sqlc.arg(keyword)||'%'
ORDER BY priority DESC,id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);
