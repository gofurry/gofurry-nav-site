-- Close the multi-instance claim race: PostgreSQL enforces one running Job per
-- acquisition lane even when separate workers inspect the queue concurrently.
-- +goose Up

CREATE UNIQUE INDEX uq_gfg_collection_jobs_running_lane
    ON public.gfg_collection_jobs(concurrency_key)
    WHERE status = 'running';

-- +goose Down

DROP INDEX public.uq_gfg_collection_jobs_running_lane;
