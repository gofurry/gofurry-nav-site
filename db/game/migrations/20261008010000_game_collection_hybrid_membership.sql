-- +goose Up
CREATE TABLE public.gfg_game_collection_tag (
    collection_id bigint NOT NULL REFERENCES public.gfg_game_collection(id) ON DELETE CASCADE,
    tag_id bigint NOT NULL REFERENCES public.gfg_tag(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (collection_id, tag_id)
);
CREATE INDEX gfg_game_collection_tag_tag_index ON public.gfg_game_collection_tag(tag_id, collection_id);

CREATE TABLE public.gfg_game_collection_exclusion (
    collection_id bigint NOT NULL REFERENCES public.gfg_game_collection(id) ON DELETE CASCADE,
    game_id bigint NOT NULL REFERENCES public.gfg_game(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (collection_id, game_id)
);
CREATE INDEX gfg_game_collection_exclusion_game_index ON public.gfg_game_collection_exclusion(game_id, collection_id);

COMMENT ON TABLE public.gfg_game_collection_tag IS '分区自动收录的人工标签规则；任一有效标签命中即可收录，标签或类别归档时规则保留但暂停匹配。';
COMMENT ON COLUMN public.gfg_game_collection_tag.collection_id IS '所属策展分区；删除分区时级联移除规则，不删除标签或游戏。';
COMMENT ON COLUMN public.gfg_game_collection_tag.tag_id IS '自动匹配的正式标签；normal、primary、secondary 关系均参与，删除标签前必须显式解绑。';
COMMENT ON COLUMN public.gfg_game_collection_tag.created_at IS '运营绑定此规则的时刻，不代表自动成员加入时间，也不影响公开时间线排序。';
COMMENT ON TABLE public.gfg_game_collection_exclusion IS '分区人工排除规则；优先于自动命中与人工固定，未命中的游戏也可保留规则供未来匹配时排除。';
COMMENT ON COLUMN public.gfg_game_collection_exclusion.collection_id IS '排除规则所属的策展分区；删除分区时级联清理规则。';
COMMENT ON COLUMN public.gfg_game_collection_exclusion.game_id IS '在本分区排除的游戏；不删除游戏或修改标签，应用写入层禁止同时人工固定。';
COMMENT ON COLUMN public.gfg_game_collection_exclusion.created_at IS '运营设置排除的时刻；解除后按当前自动规则与人工固定重新派生成员。';

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Hybrid Collection rollback would destroy curated rules and exclusions; restore a verified backup or recreate an isolated database instead';
END $$;
-- +goose StatementEnd
