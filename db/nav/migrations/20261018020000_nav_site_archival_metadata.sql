-- +goose Up
ALTER TABLE gfn_site
    ADD COLUMN deletion_reason_code text,
    ADD COLUMN deletion_reason text,
    ADD CONSTRAINT gfn_site_deletion_reason_code_check CHECK (
        deletion_reason_code IS NULL OR deletion_reason_code IN (
            'unreachable', 'closed', 'migrated', 'duplicate', 'owner_request', 'other'
        )
    ),
    ADD CONSTRAINT gfn_site_deletion_reason_length_check CHECK (
        deletion_reason IS NULL OR char_length(deletion_reason) <= 1000
    ),
    ADD CONSTRAINT gfn_site_active_archival_metadata_check CHECK (
        deleted IS TRUE OR (deletion_reason_code IS NULL AND deletion_reason IS NULL)
    );

COMMENT ON COLUMN gfn_site.deletion_reason_code IS '网站软归档原因：unreachable 无法访问、closed 确认停运、migrated 网站迁移、duplicate 重复收录、owner_request 站长要求移除、other 其他原因；历史归档允许 NULL，未归档必须为 NULL，新归档由 Admin 强制显式选择。';
COMMENT ON COLUMN gfn_site.deletion_reason IS '网站软归档的可选人工说明，Admin 去除首尾空白后最多 1000 个字符；NULL 表示未填写，不回填历史原因，不作为 Target 时段关闭原因。';

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Site archival metadata is retained history; restore a backup instead of destructive rollback';
END $$;
-- +goose StatementEnd
