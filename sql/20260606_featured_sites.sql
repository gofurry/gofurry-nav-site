CREATE TABLE IF NOT EXISTS gfn_featured_site (
    id bigint PRIMARY KEY,
    site_id bigint NOT NULL,
    weight bigint NOT NULL DEFAULT 0,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_gfn_featured_site_site_id
    ON gfn_featured_site (site_id);

CREATE INDEX IF NOT EXISTS idx_gfn_featured_site_weight
    ON gfn_featured_site (weight DESC, id DESC);
