-- GoFurry Nav Collector v0.2.1
-- 将 gfn_site.domain 的历史域名关系同步到 gfn_collector_domain.site_id。
--
-- 执行前提：
-- 1. 已先执行 sql/20260523_collector_v2_observation.sql，确保 gfn_collector_domain 已有 site_id / deleted 字段。
-- 2. 建议在低峰期执行；执行期间尽量暂停 gofurry-admin 对“采集域名”的新增操作。
-- 3. 本脚本只回填和补缺，不删除任何已有采集域名。

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'gfn_collector_domain'
          AND column_name = 'site_id'
    ) THEN
        RAISE EXCEPTION 'gfn_collector_domain.site_id 不存在，请先执行 v0.2.0 observation DDL';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'gfn_collector_domain'
          AND column_name = 'deleted'
    ) THEN
        RAISE EXCEPTION 'gfn_collector_domain.deleted 不存在，请先执行 v0.2.0 observation DDL';
    END IF;
END $$;

-- 1. 给已经存在的采集域名回填 site_id。
WITH site_domains AS (
    SELECT
        s.id AS site_id,
        trim(d.domain) AS domain,
        lower(trim(d.domain)) AS domain_key
    FROM gfn_site AS s
    CROSS JOIN LATERAL jsonb_array_elements_text(
        CASE
            WHEN jsonb_typeof((s.domain::jsonb)->'domain') = 'array'
                THEN (s.domain::jsonb)->'domain'
            ELSE '[]'::jsonb
        END
    ) AS d(domain)
    WHERE s.deleted IS NOT TRUE
      AND trim(d.domain) <> ''
),
safe_site_domains AS (
    SELECT
        min(site_id) AS site_id,
        min(domain) AS domain,
        domain_key
    FROM site_domains
    GROUP BY domain_key
    HAVING count(DISTINCT site_id) = 1
)
UPDATE gfn_collector_domain AS cd
SET site_id = sd.site_id
FROM safe_site_domains AS sd
WHERE (cd.site_id IS NULL OR cd.site_id <= 0)
  AND cd.deleted IS NOT TRUE
  AND (
      lower(trim(coalesce(cd.prefix, '') || cd.name)) = sd.domain_key
      OR lower(trim(cd.name)) = sd.domain_key
  );

-- 2. 把 gfn_site.domain 中存在、但采集域名表缺失的域名补入 gfn_collector_domain。
WITH sync_lock AS (
    -- 与 gofurry-admin 的 AllocateSequentialIDs 使用同一张表的 advisory lock key。
    SELECT pg_advisory_xact_lock(1507113902)
),
site_domains AS (
    SELECT
        s.id AS site_id,
        trim(d.domain) AS domain,
        lower(trim(d.domain)) AS domain_key
    FROM gfn_site AS s
    CROSS JOIN LATERAL jsonb_array_elements_text(
        CASE
            WHEN jsonb_typeof((s.domain::jsonb)->'domain') = 'array'
                THEN (s.domain::jsonb)->'domain'
            ELSE '[]'::jsonb
        END
    ) AS d(domain)
    WHERE s.deleted IS NOT TRUE
      AND trim(d.domain) <> ''
),
safe_site_domains AS (
    SELECT
        min(site_id) AS site_id,
        min(domain) AS domain,
        domain_key
    FROM site_domains
    GROUP BY domain_key
    HAVING count(DISTINCT site_id) = 1
),
active_collector_domains AS (
    SELECT lower(trim(coalesce(prefix, '') || name)) AS domain_key
    FROM gfn_collector_domain
    WHERE deleted IS NOT TRUE

    UNION

    SELECT lower(trim(name)) AS domain_key
    FROM gfn_collector_domain
    WHERE deleted IS NOT TRUE
),
missing_domains AS (
    SELECT sd.site_id, sd.domain, sd.domain_key
    FROM safe_site_domains AS sd
    WHERE NOT EXISTS (
        SELECT 1
        FROM active_collector_domains AS cd
        WHERE cd.domain_key = sd.domain_key
    )
),
numbered_missing AS (
    SELECT
        row_number() OVER (ORDER BY site_id ASC, domain_key ASC) AS rn,
        site_id,
        domain
    FROM missing_domains
),
id_base AS (
    SELECT coalesce(max(cd.id), 0) AS max_id
    FROM gfn_collector_domain AS cd, sync_lock
)
INSERT INTO gfn_collector_domain (id, site_id, name, proxy, prefix, tls, deleted)
SELECT
    id_base.max_id + numbered_missing.rn AS id,
    numbered_missing.site_id,
    numbered_missing.domain AS name,
    '0' AS proxy,
    NULL AS prefix,
    '1' AS tls,
    false AS deleted
FROM numbered_missing
CROSS JOIN id_base;

-- 3. 人工检查：仍未绑定站点的采集域名。
SELECT id, site_id, name, prefix, proxy, tls, deleted
FROM gfn_collector_domain
WHERE deleted IS NOT TRUE
  AND (site_id IS NULL OR site_id <= 0)
ORDER BY id ASC;

-- 4. 人工检查：同一个域名如果出现在多个站点中，本脚本不会自动绑定，需要人工决定归属。
WITH site_domains AS (
    SELECT
        s.id AS site_id,
        s.name AS site_name,
        trim(d.domain) AS domain,
        lower(trim(d.domain)) AS domain_key
    FROM gfn_site AS s
    CROSS JOIN LATERAL jsonb_array_elements_text(
        CASE
            WHEN jsonb_typeof((s.domain::jsonb)->'domain') = 'array'
                THEN (s.domain::jsonb)->'domain'
            ELSE '[]'::jsonb
        END
    ) AS d(domain)
    WHERE s.deleted IS NOT TRUE
      AND trim(d.domain) <> ''
)
SELECT domain_key, array_agg(site_id ORDER BY site_id) AS site_ids, array_agg(site_name ORDER BY site_id) AS site_names
FROM site_domains
GROUP BY domain_key
HAVING count(DISTINCT site_id) > 1
ORDER BY domain_key ASC;

-- 5. 人工检查：每个站点当前有效采集域名数量。
SELECT site_id, count(*) AS active_collector_domain_count
FROM gfn_collector_domain
WHERE deleted IS NOT TRUE
  AND site_id > 0
GROUP BY site_id
ORDER BY site_id ASC;
