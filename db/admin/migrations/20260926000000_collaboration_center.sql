-- #117: Admin-only content inventory and one shared text board.
-- +goose Up
CREATE TABLE public.gfa_content_idea (
    id bigserial PRIMARY KEY,
    kind text NOT NULL CHECK (kind IN ('game', 'site', 'other')),
    title text,
    source text,
    source_key text,
    note text NOT NULL DEFAULT '',
    priority text NOT NULL DEFAULT 'normal' CHECK (priority IN ('normal', 'high')),
    status text NOT NULL DEFAULT 'idea' CHECK (status IN ('idea', 'researching', 'landed', 'shelved')),
    created_by_account_id bigint NOT NULL REFERENCES public.gfa_admin_account(id) ON DELETE RESTRICT,
    researching_by_account_id bigint REFERENCES public.gfa_admin_account(id) ON DELETE RESTRICT,
    linked_kind text CHECK (linked_kind IN ('game', 'site')),
    linked_resource_id bigint CHECK (linked_resource_id > 0),
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    researching_at timestamp(0),
    landed_at timestamp(0),
    CONSTRAINT idea_content_check CHECK (COALESCE(btrim(title), '') <> '' OR COALESCE(btrim(source), '') <> ''),
    CONSTRAINT idea_research_check CHECK (
        (status = 'researching' AND researching_by_account_id IS NOT NULL AND researching_at IS NOT NULL)
        OR (status <> 'researching' AND researching_by_account_id IS NULL AND researching_at IS NULL)),
    CONSTRAINT idea_link_check CHECK (
        (status <> 'landed' AND linked_kind IS NULL AND linked_resource_id IS NULL AND landed_at IS NULL)
        OR (status = 'landed' AND landed_at IS NOT NULL AND (
            (kind IN ('game', 'site') AND linked_kind IS NOT NULL AND linked_kind = kind AND linked_resource_id IS NOT NULL)
            OR (kind = 'other' AND ((linked_kind IS NULL AND linked_resource_id IS NULL)
                OR (linked_kind IS NOT NULL AND linked_resource_id IS NOT NULL))))))
);
CREATE INDEX idx_idea_status_updated ON public.gfa_content_idea (status, updated_at DESC, id DESC);
CREATE INDEX idx_idea_kind_status_updated ON public.gfa_content_idea (kind, status, updated_at DESC, id DESC);
CREATE INDEX idx_idea_researcher_status ON public.gfa_content_idea (researching_by_account_id, status);
-- Duplicates are warnings, never a uniqueness constraint.
CREATE INDEX idx_idea_source_key ON public.gfa_content_idea (source_key) WHERE source_key IS NOT NULL;
CREATE INDEX idx_idea_landed ON public.gfa_content_idea (landed_at DESC) WHERE status = 'landed';
CREATE INDEX idx_idea_created ON public.gfa_content_idea (created_at DESC, id DESC);

CREATE TABLE public.gfa_collaboration_board_note (
    id bigserial PRIMARY KEY,
    body text NOT NULL CHECK (btrim(body) <> ''),
    x integer NOT NULL DEFAULT 0 CHECK (x >= 0),
    y integer NOT NULL DEFAULT 0 CHECK (y >= 0),
    width integer NOT NULL DEFAULT 280 CHECK (width > 0),
    height integer NOT NULL DEFAULT 200 CHECK (height > 0),
    z_index integer NOT NULL DEFAULT 0 CHECK (z_index >= 0),
    created_by_account_id bigint NOT NULL REFERENCES public.gfa_admin_account(id) ON DELETE RESTRICT,
    updated_by_account_id bigint NOT NULL REFERENCES public.gfa_admin_account(id) ON DELETE RESTRICT,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- No automatic Down: dropping these tables destroys collaboration history.
