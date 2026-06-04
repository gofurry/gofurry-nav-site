CREATE TABLE IF NOT EXISTS gfn_nav_update_notice (
    id bigint PRIMARY KEY,
    title varchar(120) NOT NULL,
    title_en varchar(120) NOT NULL,
    body text NOT NULL,
    body_en text NOT NULL,
    published_at timestamp(0) without time zone NOT NULL,
    create_time timestamp(0) without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp(0) without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted boolean NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_gfn_nav_update_notice_public
    ON gfn_nav_update_notice (deleted, published_at DESC, id DESC);
