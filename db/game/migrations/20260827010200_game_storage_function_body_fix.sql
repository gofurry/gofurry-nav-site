-- Recompile the renamed snapshot-pruning function against the canonical table.
-- PostgreSQL table rename does not rewrite relation names embedded in PL/pgSQL text.
-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.gfg_game_prune_detail_snapshots(
    p_appid bigint,
    p_lang text,
    p_region text,
    p_keep_count integer DEFAULT 5
)
RETURNS integer
LANGUAGE plpgsql
AS $function$
DECLARE
    deleted_count integer;
BEGIN
    WITH ranked AS (
        SELECT id,
               row_number() OVER (
                   PARTITION BY appid, lang, region
                   ORDER BY collected_at DESC, id DESC
               ) AS rn
        FROM public.gfg_game_detail_snapshots
        WHERE appid = p_appid
          AND lang = p_lang
          AND region = p_region
    ),
    deleted AS (
        DELETE FROM public.gfg_game_detail_snapshots snapshot
        USING ranked
        WHERE snapshot.id = ranked.id
          AND ranked.rn > GREATEST(p_keep_count, 0)
        RETURNING snapshot.id
    )
    SELECT count(*) INTO deleted_count FROM deleted;

    RETURN deleted_count;
END;
$function$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.gfg_game_prune_detail_snapshots(
    p_appid bigint,
    p_lang text,
    p_region text,
    p_keep_count integer DEFAULT 5
)
RETURNS integer
LANGUAGE plpgsql
AS $function$
DECLARE
    deleted_count integer;
BEGIN
    WITH ranked AS (
        SELECT id,
               row_number() OVER (
                   PARTITION BY appid, lang, region
                   ORDER BY collected_at DESC, id DESC
               ) AS rn
        FROM public.gfg_game_v2_detail_snapshots
        WHERE appid = p_appid
          AND lang = p_lang
          AND region = p_region
    ),
    deleted AS (
        DELETE FROM public.gfg_game_v2_detail_snapshots snapshot
        USING ranked
        WHERE snapshot.id = ranked.id
          AND ranked.rn > GREATEST(p_keep_count, 0)
        RETURNING snapshot.id
    )
    SELECT count(*) INTO deleted_count FROM deleted;

    RETURN deleted_count;
END;
$function$;
-- +goose StatementEnd
