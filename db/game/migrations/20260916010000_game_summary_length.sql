-- +goose Up
ALTER TABLE public.gfg_game
    ALTER COLUMN info TYPE varchar(400),
    ALTER COLUMN info_en TYPE varchar(400);

-- +goose Down
-- PostgreSQL rejects the narrowing if any summary exceeds 300 characters.
-- Never truncate published content during rollback.
ALTER TABLE public.gfg_game
    ALTER COLUMN info TYPE varchar(300),
    ALTER COLUMN info_en TYPE varchar(300);
