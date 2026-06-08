ALTER TABLE gfn_saying
    ADD COLUMN IF NOT EXISTS language varchar(8) NOT NULL DEFAULT 'zh';

UPDATE gfn_saying
SET language = 'zh'
WHERE language IS NULL OR trim(language) = '';

ALTER TABLE gfn_saying
    DROP CONSTRAINT IF EXISTS chk_gfn_saying_language;

ALTER TABLE gfn_saying
    ADD CONSTRAINT chk_gfn_saying_language CHECK (language IN ('zh', 'en'));

CREATE INDEX IF NOT EXISTS idx_gfn_saying_language
    ON gfn_saying (language);
