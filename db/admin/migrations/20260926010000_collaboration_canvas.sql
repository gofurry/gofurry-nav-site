-- #117: replace the text-note surface with a single shared node/edge canvas.
-- The prior migration has already been applied in development; evolve its table
-- without rewriting migration history, copying data, or retaining a legacy API.
-- +goose Up
ALTER TABLE public.gfa_collaboration_board_note RENAME TO gfa_collaboration_board_node;
ALTER SEQUENCE public.gfa_collaboration_board_note_id_seq RENAME TO gfa_collaboration_board_node_id_seq;
ALTER TABLE public.gfa_collaboration_board_node
    RENAME CONSTRAINT gfa_collaboration_board_note_pkey TO gfa_collaboration_board_node_pkey;
ALTER TABLE public.gfa_collaboration_board_node
    RENAME CONSTRAINT gfa_collaboration_board_note_created_by_account_id_fkey TO gfa_collaboration_board_node_created_by_account_id_fkey;
ALTER TABLE public.gfa_collaboration_board_node
    RENAME CONSTRAINT gfa_collaboration_board_note_updated_by_account_id_fkey TO gfa_collaboration_board_node_updated_by_account_id_fkey;
ALTER TABLE public.gfa_collaboration_board_node
    RENAME CONSTRAINT gfa_collaboration_board_note_version_check TO gfa_collaboration_board_node_version_check;
ALTER TABLE public.gfa_collaboration_board_node
    DROP CONSTRAINT gfa_collaboration_board_note_body_check,
    DROP CONSTRAINT gfa_collaboration_board_note_x_check,
    DROP CONSTRAINT gfa_collaboration_board_note_y_check,
    DROP CONSTRAINT gfa_collaboration_board_note_width_check,
    DROP CONSTRAINT gfa_collaboration_board_note_height_check,
    DROP CONSTRAINT gfa_collaboration_board_note_z_index_check,
    ALTER COLUMN body SET DEFAULT '',
    ADD COLUMN kind text NOT NULL DEFAULT 'note' CHECK (kind IN ('note','card','text','rectangle','ellipse','arrow')),
    ADD COLUMN title text NOT NULL DEFAULT '' CHECK (char_length(title) <= 200),
    ADD COLUMN color text NOT NULL DEFAULT 'sand' CHECK (color IN ('sand','blue','green','rose','slate')),
    ADD COLUMN rotation integer NOT NULL DEFAULT 0 CHECK (rotation IN (0,90,180,270)),
    ADD COLUMN reference_kind text CHECK (reference_kind IN ('idea','game','site')),
    ADD COLUMN reference_id bigint CHECK (reference_id > 0),
    ADD CONSTRAINT board_node_reference_check CHECK ((reference_kind IS NULL AND reference_id IS NULL) OR (kind='card' AND reference_kind IS NOT NULL AND reference_id IS NOT NULL)),
    ADD CONSTRAINT board_node_body_check CHECK (char_length(body) <= 10000),
    ADD CONSTRAINT board_node_content_check CHECK (kind IN ('rectangle','ellipse','arrow') OR btrim(title) <> '' OR btrim(body) <> '' OR reference_id IS NOT NULL),
    ADD CONSTRAINT board_node_geometry_check CHECK (x BETWEEN -100000 AND 100000 AND y BETWEEN -100000 AND 100000 AND width BETWEEN 48 AND 1600 AND height BETWEEN 40 AND 1600 AND z_index BETWEEN 0 AND 1000000);
CREATE INDEX idx_board_node_order ON public.gfa_collaboration_board_node(z_index, id);

CREATE TABLE public.gfa_collaboration_board_edge (
    id bigserial PRIMARY KEY,
    source_id bigint NOT NULL REFERENCES public.gfa_collaboration_board_node(id) ON DELETE CASCADE,
    target_id bigint NOT NULL REFERENCES public.gfa_collaboration_board_node(id) ON DELETE CASCADE,
    source_handle text NOT NULL CHECK (source_handle IN ('top','right','bottom','left')),
    target_handle text NOT NULL CHECK (target_handle IN ('top','right','bottom','left')),
    routing text NOT NULL DEFAULT 'curve' CHECK (routing IN ('curve','step')),
    label text NOT NULL DEFAULT '' CHECK (char_length(label) <= 200),
    color text NOT NULL DEFAULT 'slate' CHECK (color IN ('sand','blue','green','rose','slate')),
    arrow boolean NOT NULL DEFAULT true,
    created_by_account_id bigint NOT NULL REFERENCES public.gfa_admin_account(id) ON DELETE RESTRICT,
    updated_by_account_id bigint NOT NULL REFERENCES public.gfa_admin_account(id) ON DELETE RESTRICT,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT board_edge_endpoints_check CHECK (source_id <> target_id),
    CONSTRAINT board_edge_connection_key UNIQUE (source_id, source_handle, target_id, target_handle)
);
CREATE INDEX idx_board_edge_target ON public.gfa_collaboration_board_edge(target_id);
-- No automatic Down: canvas contents are durable collaboration data.
