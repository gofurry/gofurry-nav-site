ALTER TABLE gfn_site
    ADD COLUMN IF NOT EXISTS weight BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_gfn_site_weight_update_time
    ON gfn_site (weight DESC, update_time DESC, id DESC)
    WHERE deleted IS NOT TRUE;
