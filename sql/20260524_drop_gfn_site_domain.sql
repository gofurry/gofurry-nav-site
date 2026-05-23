-- GoFurry Nav Collector v0.2.1 收尾清理。
-- 在 collector_domain.site_id 方案完成上线并验证稳定后，物理移除历史字段 gfn_site.domain。
--
-- 执行前提：
-- 1. 已执行 sql/20260523_collector_v2_observation.sql。
-- 2. 已执行并人工检查 sql/20260523_collector_domain_site_sync.sql。
-- 3. gofurry-nav-collector、gofurry-nav-backend、gofurry-admin 与 Nuxt
--    均已运行基于 collector_domain.site_id 的实现。
-- 4. 生产执行前请先备份数据库。

DO $$
DECLARE
    active_unbound_count bigint;
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

    SELECT count(*)
    INTO active_unbound_count
    FROM gfn_collector_domain
    WHERE deleted IS NOT TRUE
      AND (site_id IS NULL OR site_id <= 0);

    IF active_unbound_count > 0 THEN
        RAISE EXCEPTION '发现 % 个有效采集域名缺少 site_id，请先修复后再移除 gfn_site.domain', active_unbound_count;
    END IF;
END $$;

ALTER TABLE IF EXISTS gfn_site
DROP COLUMN IF EXISTS domain;

-- 验证：字段移除后应返回空结果。
SELECT column_name
FROM information_schema.columns
WHERE table_schema = current_schema()
  AND table_name = 'gfn_site'
  AND column_name = 'domain';
