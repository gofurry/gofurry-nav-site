-- V3-P0.2 checkpoint-gated, target-aware Nav observation retention.
-- +goose Up

-- +goose StatementBegin
CREATE FUNCTION public.gfn_prune_observations_batch(
    target_keep_count integer,
    target_batch_size integer
) RETURNS bigint
LANGUAGE plpgsql
AS $function$
DECLARE
    deleted_rows bigint;
BEGIN
    IF target_keep_count < 1 OR target_batch_size < 1 THEN
        RETURN 0;
    END IF;

    WITH ranked AS (
        SELECT observation.id,
               row_number() OVER (
                   PARTITION BY observation.site_id, lower(observation.target), observation.protocol
                   ORDER BY observation.observed_at DESC, observation.id DESC
               ) AS retention_rank
        FROM public.gfn_collector_observation observation
        WHERE EXISTS (
            SELECT 1
            FROM public.gfn_fact_rollup_checkpoints checkpoint
            WHERE checkpoint.pipeline_key = 'nav.target_facts'
              AND checkpoint.processed_through IS NOT NULL
              AND (observation.observed_at AT TIME ZONE 'UTC')::date <= checkpoint.processed_through
        )
    ), eligible AS (
        SELECT id
        FROM ranked
        WHERE retention_rank > target_keep_count
        ORDER BY id
        LIMIT target_batch_size
    )
    DELETE FROM public.gfn_collector_observation observation
    USING eligible
    WHERE observation.id = eligible.id;

    GET DIAGNOSTICS deleted_rows = ROW_COUNT;
    RETURN deleted_rows;
END;
$function$;
-- +goose StatementEnd

-- +goose Down

DROP FUNCTION public.gfn_prune_observations_batch(integer, integer);
