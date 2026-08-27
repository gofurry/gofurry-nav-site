-- Close the multi-instance claim race while still allowing different Nav
-- protocol lanes to execute concurrently.
-- +goose Up

CREATE UNIQUE INDEX uq_gfn_collection_jobs_running_lane
    ON public.gfn_collection_jobs(concurrency_key)
    WHERE status = 'running';

-- +goose Down

DROP INDEX public.uq_gfn_collection_jobs_running_lane;
