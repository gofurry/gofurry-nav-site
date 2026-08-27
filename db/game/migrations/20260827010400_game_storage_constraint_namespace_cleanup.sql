-- PostgreSQL 18 materializes NOT NULL declarations as named constraints. Table
-- rename does not rename those constraints, so clean their V2 physical prefix
-- explicitly after the canonical table cutover.
-- +goose Up
-- +goose StatementBegin

DO $migration$
DECLARE
    mapping record;
    item record;
    old_count bigint;
BEGIN
    SELECT count(*) INTO old_count
    FROM pg_constraint
    WHERE connamespace = 'public'::regnamespace
      AND conname LIKE 'gfg_game_v2_%\_not_null' ESCAPE '\';
    IF old_count <> 124 THEN
        RAISE EXCEPTION 'expected 124 versioned Game NOT NULL constraints, found %', old_count;
    END IF;

    FOR mapping IN
        SELECT * FROM (VALUES
            ('gfg_game_assets', 'gfg_game_v2_assets_', 'gfg_game_assets_'),
            ('gfg_game_detail_snapshots', 'gfg_game_v2_detail_snapshots_', 'gfg_game_detail_snapshots_'),
            ('gfg_game_details', 'gfg_game_v2_details_', 'gfg_game_details_'),
            ('gfg_game_localized_details', 'gfg_game_v2_localized_details_', 'gfg_game_localized_details_'),
            ('gfg_game_media', 'gfg_game_v2_media_', 'gfg_game_media_'),
            ('gfg_game_news', 'gfg_game_v2_news_', 'gfg_game_news_'),
            ('gfg_game_player_counts', 'gfg_game_v2_player_counts_', 'gfg_game_player_counts_'),
            ('gfg_game_prices', 'gfg_game_v2_prices_', 'gfg_game_prices_'),
            ('gfg_game_recommendations', 'gfg_game_v2_recommendations_', 'gfg_game_recommendations_'),
            ('gfg_game_requirements', 'gfg_game_v2_requirements_', 'gfg_game_requirements_')
        ) AS names(table_name, old_prefix, new_prefix)
    LOOP
        FOR item IN
            SELECT conname
            FROM pg_constraint
            WHERE conrelid = ('public.' || mapping.table_name)::regclass
              AND conname LIKE mapping.old_prefix || '%\_not_null' ESCAPE '\'
            ORDER BY conname
        LOOP
            EXECUTE format(
                'ALTER TABLE public.%I RENAME CONSTRAINT %I TO %I',
                mapping.table_name,
                item.conname,
                mapping.new_prefix || substr(item.conname, length(mapping.old_prefix) + 1)
            );
        END LOOP;
    END LOOP;

    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE connamespace = 'public'::regnamespace
          AND conname LIKE 'gfg_game_v2_%\_not_null' ESCAPE '\'
    ) THEN
        RAISE EXCEPTION 'versioned Game NOT NULL constraints remain after namespace cleanup';
    END IF;
END;
$migration$;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin

DO $migration$
DECLARE
    mapping record;
    item record;
BEGIN
    FOR mapping IN
        SELECT * FROM (VALUES
            ('gfg_game_assets', 'gfg_game_assets_', 'gfg_game_v2_assets_'),
            ('gfg_game_detail_snapshots', 'gfg_game_detail_snapshots_', 'gfg_game_v2_detail_snapshots_'),
            ('gfg_game_details', 'gfg_game_details_', 'gfg_game_v2_details_'),
            ('gfg_game_localized_details', 'gfg_game_localized_details_', 'gfg_game_v2_localized_details_'),
            ('gfg_game_media', 'gfg_game_media_', 'gfg_game_v2_media_'),
            ('gfg_game_news', 'gfg_game_news_', 'gfg_game_v2_news_'),
            ('gfg_game_player_counts', 'gfg_game_player_counts_', 'gfg_game_v2_player_counts_'),
            ('gfg_game_prices', 'gfg_game_prices_', 'gfg_game_v2_prices_'),
            ('gfg_game_recommendations', 'gfg_game_recommendations_', 'gfg_game_v2_recommendations_'),
            ('gfg_game_requirements', 'gfg_game_requirements_', 'gfg_game_v2_requirements_')
        ) AS names(table_name, old_prefix, new_prefix)
    LOOP
        FOR item IN
            SELECT conname
            FROM pg_constraint
            WHERE conrelid = ('public.' || mapping.table_name)::regclass
              AND conname LIKE mapping.old_prefix || '%\_not_null' ESCAPE '\'
            ORDER BY conname
        LOOP
            EXECUTE format(
                'ALTER TABLE public.%I RENAME CONSTRAINT %I TO %I',
                mapping.table_name,
                item.conname,
                mapping.new_prefix || substr(item.conname, length(mapping.old_prefix) + 1)
            );
        END LOOP;
    END LOOP;
END;
$migration$;

-- +goose StatementEnd
