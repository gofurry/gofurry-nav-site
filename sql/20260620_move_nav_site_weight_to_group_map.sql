BEGIN;

ALTER TABLE gfn_site_group_map
    ADD COLUMN IF NOT EXISTS weight BIGINT NOT NULL DEFAULT 0;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'gfn_site'
          AND column_name = 'weight'
    ) THEN
        UPDATE gfn_site_group_map AS m
        SET weight = s.weight
        FROM gfn_site AS s
        WHERE m.site_id = s.id;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_gfn_site_group_map_group_weight
    ON gfn_site_group_map (group_id, weight DESC, update_time DESC, id DESC, site_id ASC);

CREATE INDEX IF NOT EXISTS idx_gfn_site_group_map_site_group
    ON gfn_site_group_map (site_id, group_id);

DROP INDEX IF EXISTS idx_gfn_site_weight_update_time;

ALTER TABLE gfn_site
    DROP COLUMN IF EXISTS weight;

COMMIT;
