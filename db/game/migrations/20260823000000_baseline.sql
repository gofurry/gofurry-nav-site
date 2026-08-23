-- Canonical gfg baseline, version 20260823000000.
-- Derived from the audited production schema and normalized to omit owners, ACLs, and data.
-- +goose Up

CREATE EXTENSION IF NOT EXISTS "pg_trgm" WITH SCHEMA "public";

CREATE SEQUENCE public."gfg_game_v2_assets_id_seq"
    AS bigint
    START WITH 1
    INCREMENT BY 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE SEQUENCE public."gfg_game_v2_collect_task_results_id_seq"
    AS bigint
    START WITH 1
    INCREMENT BY 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE SEQUENCE public."gfg_game_v2_detail_snapshots_id_seq"
    AS bigint
    START WITH 1
    INCREMENT BY 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE SEQUENCE public."gfg_game_v2_media_id_seq"
    AS bigint
    START WITH 1
    INCREMENT BY 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE SEQUENCE public."gfg_game_v2_news_id_seq"
    AS bigint
    START WITH 1
    INCREMENT BY 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE SEQUENCE public."gfg_game_v2_player_counts_id_seq"
    AS bigint
    START WITH 1
    INCREMENT BY 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE TABLE public."gfg_game" (
    "id" bigint NOT NULL,
    "name" character varying(255) NOT NULL,
    "name_en" character varying(255) NOT NULL,
    "info" character varying(300) NOT NULL,
    "info_en" character varying(300) NOT NULL,
    "create_time" timestamp(0) without time zone NOT NULL,
    "update_time" timestamp(0) without time zone NOT NULL,
    "resources" jsonb,
    "groups" jsonb,
    "release_date" character varying(255) NOT NULL,
    "developers" jsonb NOT NULL,
    "publishers" jsonb NOT NULL,
    "appid" bigint NOT NULL,
    "header" character varying(255) NOT NULL,
    "links" jsonb,
    "weight" bigint NOT NULL,
    "primary_tag" bigint NOT NULL,
    "secondary_tag" bigint NOT NULL,
    "view_count" bigint DEFAULT 0 NOT NULL
);

CREATE TABLE public."gfg_game_comment" (
    "id" bigint NOT NULL,
    "region" character varying(50) NOT NULL,
    "content" text NOT NULL,
    "score" double precision NOT NULL,
    "create_time" timestamp(0) without time zone NOT NULL,
    "game_id" bigint NOT NULL,
    "ip" character varying(50) NOT NULL,
    "name" character varying(50) NOT NULL
);

CREATE TABLE public."gfg_game_creator_deprecated_20260614" (
    "id" bigint CONSTRAINT "gfg_game_creator_id_not_null" NOT NULL,
    "name" character varying(100) CONSTRAINT "gfg_game_creator_name_not_null" NOT NULL,
    "info" character varying(700) CONSTRAINT "gfg_game_creator_info_not_null" NOT NULL,
    "main_url" character varying(255) CONSTRAINT "gfg_game_creator_main_url_not_null" NOT NULL,
    "links" jsonb,
    "cover" character varying(255) CONSTRAINT "gfg_game_creator_cover_not_null" NOT NULL,
    "contact" jsonb,
    "create_time" timestamp(0) without time zone CONSTRAINT "gfg_game_creator_create_time_not_null" NOT NULL,
    "update_time" timestamp(0) without time zone CONSTRAINT "gfg_game_creator_update_time_not_null" NOT NULL,
    "type" bigint CONSTRAINT "gfg_game_creator_type_not_null" NOT NULL,
    "name_en" character varying(100) CONSTRAINT "gfg_game_creator_name_en_not_null" NOT NULL,
    "info_en" character varying(700) CONSTRAINT "gfg_game_creator_info_en_not_null" NOT NULL,
    "deleted" boolean CONSTRAINT "gfg_game_creator_deleted_not_null" NOT NULL
);

CREATE TABLE public."gfg_game_news" (
    "id" bigint NOT NULL,
    "game_id" bigint NOT NULL,
    "headline" character varying(255) NOT NULL,
    "content" text NOT NULL,
    "index" bigint NOT NULL,
    "post_time" timestamp(0) without time zone NOT NULL,
    "create_time" timestamp(0) without time zone NOT NULL,
    "author" character varying(50) NOT NULL,
    "url" character varying(255) NOT NULL,
    "total" bigint NOT NULL,
    "lang" character varying(30) NOT NULL
);

CREATE TABLE public."gfg_game_player_count" (
    "id" bigint NOT NULL,
    "game_id" bigint NOT NULL,
    "count" bigint NOT NULL,
    "create_time" timestamp(0) without time zone NOT NULL
);

CREATE TABLE public."gfg_game_record" (
    "id" bigint NOT NULL,
    "game_id" bigint NOT NULL,
    "language" text NOT NULL,
    "release_date" character varying(30) NOT NULL,
    "platform" character varying(50) NOT NULL,
    "developer" character varying(100) NOT NULL,
    "publisher" character varying(100) NOT NULL,
    "info" text NOT NULL,
    "cover" character varying(255) NOT NULL,
    "lang" character varying(20) NOT NULL,
    "price_list" jsonb NOT NULL,
    "initial" bigint NOT NULL,
    "final" bigint NOT NULL,
    "discount" bigint NOT NULL
);

CREATE TABLE public."gfg_game_v2_assets" (
    "id" bigint DEFAULT nextval('gfg_game_v2_assets_id_seq'::regclass) NOT NULL,
    "game_id" bigint NOT NULL,
    "appid" bigint NOT NULL,
    "asset_type" text NOT NULL,
    "asset_family" text DEFAULT ''::text NOT NULL,
    "source" text DEFAULT 'steam'::text NOT NULL,
    "lang" text DEFAULT ''::text NOT NULL,
    "media_key" text DEFAULT ''::text NOT NULL,
    "title" text DEFAULT ''::text NOT NULL,
    "url" text DEFAULT ''::text NOT NULL,
    "thumbnail_url" text DEFAULT ''::text NOT NULL,
    "format" text DEFAULT ''::text NOT NULL,
    "exists" boolean,
    "status_code" integer DEFAULT 0 NOT NULL,
    "content_type" text DEFAULT ''::text NOT NULL,
    "content_length" bigint DEFAULT 0 NOT NULL,
    "extra" jsonb DEFAULT '{}'::jsonb NOT NULL,
    "sort_order" integer DEFAULT 0 NOT NULL,
    "checked_at" timestamp with time zone,
    "collected_at" timestamp with time zone DEFAULT now() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public."gfg_game_v2_collect_runs" (
    "id" text NOT NULL,
    "task_type" text NOT NULL,
    "status" text NOT NULL,
    "total_count" integer DEFAULT 0 NOT NULL,
    "success_count" integer DEFAULT 0 NOT NULL,
    "failed_count" integer DEFAULT 0 NOT NULL,
    "skipped_count" integer DEFAULT 0 NOT NULL,
    "error_kind" text DEFAULT ''::text NOT NULL,
    "error_message" text DEFAULT ''::text NOT NULL,
    "started_at" timestamp with time zone DEFAULT now() NOT NULL,
    "ended_at" timestamp with time zone,
    "partial_count" integer DEFAULT 0 NOT NULL,
    "task_summary" jsonb DEFAULT '[]'::jsonb NOT NULL,
    "duration_millis" bigint DEFAULT 0 NOT NULL
);

CREATE TABLE public."gfg_game_v2_collect_task_results" (
    "id" bigint DEFAULT nextval('gfg_game_v2_collect_task_results_id_seq'::regclass) NOT NULL,
    "run_id" text NOT NULL,
    "task_type" text NOT NULL,
    "status" text NOT NULL,
    "game_id" bigint NOT NULL,
    "appid" bigint NOT NULL,
    "upstream_status_code" integer DEFAULT 0 NOT NULL,
    "traffic_bucket" text DEFAULT ''::text NOT NULL,
    "retry_count" integer DEFAULT 0 NOT NULL,
    "duration_millis" bigint DEFAULT 0 NOT NULL,
    "error_kind" text DEFAULT ''::text NOT NULL,
    "error_message" text DEFAULT ''::text NOT NULL,
    "started_at" timestamp with time zone DEFAULT now() NOT NULL,
    "ended_at" timestamp with time zone
);

CREATE TABLE public."gfg_game_v2_detail_snapshots" (
    "id" bigint DEFAULT nextval('gfg_game_v2_detail_snapshots_id_seq'::regclass) NOT NULL,
    "game_id" bigint NOT NULL,
    "appid" bigint NOT NULL,
    "lang" text NOT NULL,
    "region" text NOT NULL,
    "source" text DEFAULT 'steam'::text NOT NULL,
    "payload_hash" text DEFAULT ''::text NOT NULL,
    "raw_payload" jsonb NOT NULL,
    "collected_at" timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public."gfg_game_v2_details" (
    "game_id" bigint NOT NULL,
    "appid" bigint NOT NULL,
    "source" text DEFAULT 'steam'::text NOT NULL,
    "type" text DEFAULT ''::text NOT NULL,
    "name" text DEFAULT ''::text NOT NULL,
    "is_free" boolean DEFAULT false NOT NULL,
    "website" text DEFAULT ''::text NOT NULL,
    "header_url" text DEFAULT ''::text NOT NULL,
    "developers" jsonb DEFAULT '[]'::jsonb NOT NULL,
    "publishers" jsonb DEFAULT '[]'::jsonb NOT NULL,
    "release_coming_soon" boolean DEFAULT false NOT NULL,
    "release_date_text" text DEFAULT ''::text NOT NULL,
    "platforms" jsonb DEFAULT '{}'::jsonb NOT NULL,
    "supported_languages" text DEFAULT ''::text NOT NULL,
    "support_info" jsonb DEFAULT '{}'::jsonb NOT NULL,
    "content_descriptors" jsonb DEFAULT '{}'::jsonb NOT NULL,
    "ratings" jsonb DEFAULT '[]'::jsonb NOT NULL,
    "collected_at" timestamp with time zone DEFAULT now() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public."gfg_game_v2_localized_details" (
    "game_id" bigint NOT NULL,
    "appid" bigint NOT NULL,
    "lang" text NOT NULL,
    "name" text DEFAULT ''::text NOT NULL,
    "short_description" text DEFAULT ''::text NOT NULL,
    "detailed_description" text DEFAULT ''::text NOT NULL,
    "about_the_game" text DEFAULT ''::text NOT NULL,
    "collected_at" timestamp with time zone DEFAULT now() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public."gfg_game_v2_media" (
    "id" bigint DEFAULT nextval('gfg_game_v2_media_id_seq'::regclass) NOT NULL,
    "game_id" bigint NOT NULL,
    "appid" bigint NOT NULL,
    "media_type" text NOT NULL,
    "media_key" text DEFAULT ''::text NOT NULL,
    "title" text DEFAULT ''::text NOT NULL,
    "url" text DEFAULT ''::text NOT NULL,
    "thumbnail_url" text DEFAULT ''::text NOT NULL,
    "extra" jsonb DEFAULT '{}'::jsonb NOT NULL,
    "sort_order" integer DEFAULT 0 NOT NULL,
    "collected_at" timestamp with time zone DEFAULT now() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public."gfg_game_v2_news" (
    "id" bigint DEFAULT nextval('gfg_game_v2_news_id_seq'::regclass) NOT NULL,
    "game_id" bigint NOT NULL,
    "appid" bigint NOT NULL,
    "lang" text NOT NULL,
    "event_gid" text DEFAULT ''::text NOT NULL,
    "announcement_gid" text DEFAULT ''::text NOT NULL,
    "forum_topic_id" text DEFAULT ''::text NOT NULL,
    "headline" text DEFAULT ''::text NOT NULL,
    "raw_body" text DEFAULT ''::text NOT NULL,
    "html" text DEFAULT ''::text NOT NULL,
    "plain_text" text DEFAULT ''::text NOT NULL,
    "summary" text DEFAULT ''::text NOT NULL,
    "url" text DEFAULT ''::text NOT NULL,
    "tags" jsonb DEFAULT '[]'::jsonb NOT NULL,
    "vote_up_count" integer DEFAULT 0 NOT NULL,
    "vote_down_count" integer DEFAULT 0 NOT NULL,
    "comment_count" integer DEFAULT 0 NOT NULL,
    "raw_event" jsonb DEFAULT '{}'::jsonb NOT NULL,
    "published_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "collected_at" timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public."gfg_game_v2_player_counts" (
    "id" bigint DEFAULT nextval('gfg_game_v2_player_counts_id_seq'::regclass) NOT NULL,
    "game_id" bigint NOT NULL,
    "appid" bigint NOT NULL,
    "count" bigint DEFAULT 0 NOT NULL,
    "status" text DEFAULT 'success'::text NOT NULL,
    "upstream_status_code" integer DEFAULT 0 NOT NULL,
    "error_kind" text DEFAULT ''::text NOT NULL,
    "error_message" text DEFAULT ''::text NOT NULL,
    "collected_at" timestamp with time zone DEFAULT now() NOT NULL,
    "run_id" text DEFAULT ''::text NOT NULL
);

CREATE TABLE public."gfg_game_v2_prices" (
    "game_id" bigint NOT NULL,
    "appid" bigint NOT NULL,
    "region" text NOT NULL,
    "is_free" boolean DEFAULT false NOT NULL,
    "currency" text DEFAULT ''::text NOT NULL,
    "initial_amount" bigint DEFAULT 0 NOT NULL,
    "final_amount" bigint DEFAULT 0 NOT NULL,
    "discount_percent" bigint DEFAULT 0 NOT NULL,
    "initial_formatted" text DEFAULT ''::text NOT NULL,
    "final_formatted" text DEFAULT ''::text NOT NULL,
    "collected_at" timestamp with time zone DEFAULT now() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public."gfg_game_v2_recommendations" (
    "source_game_id" bigint NOT NULL,
    "target_game_id" bigint NOT NULL,
    "score" double precision DEFAULT 0 NOT NULL,
    "display_score" double precision DEFAULT 0 NOT NULL,
    "rank" integer DEFAULT 0 NOT NULL,
    "reason_json" jsonb DEFAULT '[]'::jsonb NOT NULL,
    "algorithm_version" text DEFAULT ''::text NOT NULL,
    "computed_at" timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public."gfg_game_v2_requirements" (
    "game_id" bigint NOT NULL,
    "appid" bigint NOT NULL,
    "pc" jsonb DEFAULT '{}'::jsonb NOT NULL,
    "mac" jsonb DEFAULT '{}'::jsonb NOT NULL,
    "linux" jsonb DEFAULT '{}'::jsonb NOT NULL,
    "collected_at" timestamp with time zone DEFAULT now() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public."gfg_prize" (
    "id" bigint NOT NULL,
    "title" character varying(100) NOT NULL,
    "desc" text NOT NULL,
    "prize" jsonb NOT NULL,
    "key" character varying(255) NOT NULL,
    "start_time" timestamp(0) without time zone NOT NULL,
    "end_time" timestamp(0) without time zone NOT NULL,
    "create_time" timestamp(0) without time zone NOT NULL,
    "status" boolean NOT NULL
);

CREATE TABLE public."gfg_prize_member" (
    "id" bigint NOT NULL,
    "prize_id" bigint NOT NULL,
    "name" character varying(50) NOT NULL,
    "email" character varying(255) NOT NULL,
    "ip" character varying(50) NOT NULL,
    "agent" character varying(700) NOT NULL,
    "is_winner" boolean NOT NULL,
    "prize_key" character varying(255),
    "create_time" timestamp(0) without time zone NOT NULL
);

CREATE TABLE public."gfg_tag" (
    "id" bigint NOT NULL,
    "name" character varying(255) NOT NULL,
    "name_en" character varying(255) NOT NULL,
    "info" character varying(255) NOT NULL,
    "info_en" character varying(255) NOT NULL,
    "prefix" bigint NOT NULL,
    "create_time" timestamp(0) without time zone NOT NULL,
    "update_time" timestamp(0) without time zone NOT NULL
);

CREATE TABLE public."gfg_tag_map" (
    "id" bigint NOT NULL,
    "game_id" bigint NOT NULL,
    "tag_id" bigint NOT NULL,
    "create_time" timestamp(0) without time zone NOT NULL,
    "update_time" timestamp(0) without time zone NOT NULL
);


ALTER SEQUENCE public."gfg_game_v2_assets_id_seq" OWNED BY public."gfg_game_v2_assets"."id";
ALTER SEQUENCE public."gfg_game_v2_collect_task_results_id_seq" OWNED BY public."gfg_game_v2_collect_task_results"."id";
ALTER SEQUENCE public."gfg_game_v2_detail_snapshots_id_seq" OWNED BY public."gfg_game_v2_detail_snapshots"."id";
ALTER SEQUENCE public."gfg_game_v2_media_id_seq" OWNED BY public."gfg_game_v2_media"."id";
ALTER SEQUENCE public."gfg_game_v2_news_id_seq" OWNED BY public."gfg_game_v2_news"."id";
ALTER SEQUENCE public."gfg_game_v2_player_counts_id_seq" OWNED BY public."gfg_game_v2_player_counts"."id";

ALTER TABLE ONLY public."gfg_game" ADD CONSTRAINT "gfg_game_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_game_comment" ADD CONSTRAINT "gfg_game_comment_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_game_creator_deprecated_20260614" ADD CONSTRAINT "gfg_game_creator_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_game_news" ADD CONSTRAINT "gfg_game_news_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_game_player_count" ADD CONSTRAINT "gfg_game_player_count_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_game_record" ADD CONSTRAINT "gfg_game_record_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_game_v2_assets" ADD CONSTRAINT "gfg_game_v2_assets_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_game_v2_collect_runs" ADD CONSTRAINT "gfg_game_v2_collect_runs_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_game_v2_collect_task_results" ADD CONSTRAINT "gfg_game_v2_collect_task_results_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_game_v2_collect_task_results" ADD CONSTRAINT "gfg_game_v2_collect_task_results_run_id_fkey" FOREIGN KEY (run_id) REFERENCES gfg_game_v2_collect_runs(id) ON DELETE CASCADE;
ALTER TABLE ONLY public."gfg_game_v2_detail_snapshots" ADD CONSTRAINT "gfg_game_v2_detail_snapshots_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_game_v2_details" ADD CONSTRAINT "gfg_game_v2_details_appid_key" UNIQUE (appid);
ALTER TABLE ONLY public."gfg_game_v2_details" ADD CONSTRAINT "gfg_game_v2_details_pkey" PRIMARY KEY (game_id);
ALTER TABLE ONLY public."gfg_game_v2_localized_details" ADD CONSTRAINT "gfg_game_v2_localized_details_pkey" PRIMARY KEY (game_id, lang);
ALTER TABLE ONLY public."gfg_game_v2_media" ADD CONSTRAINT "gfg_game_v2_media_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_game_v2_news" ADD CONSTRAINT "gfg_game_v2_news_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_game_v2_player_counts" ADD CONSTRAINT "gfg_game_v2_player_counts_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_game_v2_prices" ADD CONSTRAINT "gfg_game_v2_prices_pkey" PRIMARY KEY (game_id, region);
ALTER TABLE ONLY public."gfg_game_v2_recommendations" ADD CONSTRAINT "gfg_game_v2_recommendations_pkey" PRIMARY KEY (source_game_id, target_game_id);
ALTER TABLE ONLY public."gfg_game_v2_requirements" ADD CONSTRAINT "gfg_game_v2_requirements_appid_key" UNIQUE (appid);
ALTER TABLE ONLY public."gfg_game_v2_requirements" ADD CONSTRAINT "gfg_game_v2_requirements_pkey" PRIMARY KEY (game_id);
ALTER TABLE ONLY public."gfg_prize" ADD CONSTRAINT "gfg_prize_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_prize_member" ADD CONSTRAINT "gfg_prize_member_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_tag" ADD CONSTRAINT "gfg_tags_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfg_tag_map" ADD CONSTRAINT "gfg_tag_map_pkey" PRIMARY KEY (id);

CREATE UNIQUE INDEX idx_gfg_game_appid ON public.gfg_game USING btree (appid);
CREATE INDEX idx_gfg_game_create_time ON public.gfg_game USING btree (create_time DESC);
CREATE INDEX idx_gfg_game_developers_jsonb ON public.gfg_game USING gin (developers);
CREATE INDEX idx_gfg_game_name_en_trgm ON public.gfg_game USING gin (name_en gin_trgm_ops);
CREATE INDEX idx_gfg_game_name_trgm ON public.gfg_game USING gin (name gin_trgm_ops);
CREATE INDEX idx_gfg_game_primary_tag ON public.gfg_game USING btree (primary_tag);
CREATE INDEX idx_gfg_game_publishers_jsonb ON public.gfg_game USING gin (publishers);
CREATE INDEX idx_gfg_game_secondary_tag ON public.gfg_game USING btree (secondary_tag);
CREATE INDEX idx_gfg_game_update_time ON public.gfg_game USING btree (update_time DESC);
CREATE INDEX idx_gfg_game_weight ON public.gfg_game USING btree (weight DESC);
CREATE INDEX idx_gfg_game_comment_game_id ON public.gfg_game_comment USING btree (game_id);
CREATE INDEX idx_gfg_game_comment_game_id_create_time ON public.gfg_game_comment USING btree (game_id, create_time DESC);
CREATE INDEX idx_gfg_game_creator_deleted ON public.gfg_game_creator_deprecated_20260614 USING btree (deleted);
CREATE INDEX idx_gfg_game_creator_type ON public.gfg_game_creator_deprecated_20260614 USING btree (type);
CREATE INDEX idx_gfg_game_news_game_id ON public.gfg_game_news USING btree (game_id);
CREATE INDEX idx_gfg_game_news_game_id_post_time ON public.gfg_game_news USING btree (game_id, post_time DESC);
CREATE INDEX idx_gfg_game_player_count_game_id ON public.gfg_game_player_count USING btree (game_id);
CREATE INDEX idx_gfg_game_player_count_game_id_create_time ON public.gfg_game_player_count USING btree (game_id, create_time DESC);
CREATE INDEX idx_gfg_game_record_game_id ON public.gfg_game_record USING btree (game_id);
CREATE INDEX idx_gfg_game_v2_assets_app_type ON public.gfg_game_v2_assets USING btree (appid, asset_type, lang);
CREATE INDEX idx_gfg_game_v2_assets_exists ON public.gfg_game_v2_assets USING btree (game_id, asset_type) WHERE ("exists" IS DISTINCT FROM false);
CREATE INDEX idx_gfg_game_v2_assets_game_family ON public.gfg_game_v2_assets USING btree (game_id, asset_family, sort_order);
CREATE UNIQUE INDEX uq_gfg_game_v2_assets_item ON public.gfg_game_v2_assets USING btree (game_id, asset_type, lang, media_key);
CREATE INDEX idx_gfg_game_v2_collect_runs_status_time ON public.gfg_game_v2_collect_runs USING btree (status, started_at DESC);
CREATE INDEX idx_gfg_game_v2_collect_runs_task_time ON public.gfg_game_v2_collect_runs USING btree (task_type, started_at DESC);
CREATE INDEX idx_gfg_game_v2_collect_task_results_app_time ON public.gfg_game_v2_collect_task_results USING btree (appid, task_type, started_at DESC);
CREATE INDEX idx_gfg_game_v2_collect_task_results_run ON public.gfg_game_v2_collect_task_results USING btree (run_id, status);
CREATE INDEX idx_gfg_game_v2_collect_task_results_started_at ON public.gfg_game_v2_collect_task_results USING btree (started_at);
CREATE INDEX idx_gfg_game_v2_detail_snapshots_hash ON public.gfg_game_v2_detail_snapshots USING btree (appid, lang, region, payload_hash);
CREATE INDEX idx_gfg_game_v2_detail_snapshots_lookup ON public.gfg_game_v2_detail_snapshots USING btree (appid, lang, region, collected_at DESC);
CREATE INDEX idx_gfg_game_v2_localized_details_app_lang ON public.gfg_game_v2_localized_details USING btree (appid, lang);
CREATE INDEX idx_gfg_game_v2_media_app_type ON public.gfg_game_v2_media USING btree (appid, media_type, sort_order);
CREATE UNIQUE INDEX uq_gfg_game_v2_media_item ON public.gfg_game_v2_media USING btree (game_id, media_type, media_key);
CREATE INDEX idx_gfg_game_v2_news_feed ON public.gfg_game_v2_news USING btree (game_id, lang, published_at DESC NULLS LAST, collected_at DESC);
CREATE UNIQUE INDEX uq_gfg_game_v2_news_event_lang ON public.gfg_game_v2_news USING btree (appid, lang, event_gid, announcement_gid);
CREATE INDEX idx_gfg_game_v2_player_counts_collected_at ON public.gfg_game_v2_player_counts USING btree (collected_at);
CREATE INDEX idx_gfg_game_v2_player_counts_latest ON public.gfg_game_v2_player_counts USING btree (game_id, collected_at DESC);
CREATE INDEX idx_gfg_game_v2_player_counts_run ON public.gfg_game_v2_player_counts USING btree (run_id) WHERE (run_id <> ''::text);
CREATE INDEX idx_gfg_game_v2_prices_app_region ON public.gfg_game_v2_prices USING btree (appid, region);
CREATE INDEX idx_gfg_game_v2_recommendations_computed_at ON public.gfg_game_v2_recommendations USING btree (computed_at);
CREATE INDEX idx_gfg_game_v2_recommendations_lookup ON public.gfg_game_v2_recommendations USING btree (source_game_id, algorithm_version, rank, score DESC);
CREATE INDEX idx_gfg_prize_status_time ON public.gfg_prize USING btree (status, start_time, end_time);
CREATE INDEX idx_gfg_prize_member_prize_id ON public.gfg_prize_member USING btree (prize_id);
CREATE INDEX idx_gfg_prize_member_prize_id_is_winner ON public.gfg_prize_member USING btree (prize_id, is_winner);
CREATE INDEX idx_gfg_tag_prefix ON public.gfg_tag USING btree (prefix);
CREATE INDEX idx_gfg_tag_map_game_id ON public.gfg_tag_map USING btree (game_id);
CREATE INDEX idx_gfg_tag_map_tag_id ON public.gfg_tag_map USING btree (tag_id);

COMMENT ON TABLE public."gfg_game" IS '游戏表';
COMMENT ON COLUMN public."gfg_game"."id" IS '游戏表ID';
COMMENT ON COLUMN public."gfg_game"."name" IS '游戏名称';
COMMENT ON COLUMN public."gfg_game"."name_en" IS '游戏英文名称';
COMMENT ON COLUMN public."gfg_game"."info" IS '游戏简介';
COMMENT ON COLUMN public."gfg_game"."info_en" IS '游戏英文简介';
COMMENT ON COLUMN public."gfg_game"."create_time" IS '创建时间';
COMMENT ON COLUMN public."gfg_game"."update_time" IS '更新时间';
COMMENT ON COLUMN public."gfg_game"."resources" IS '游戏相关资源';
COMMENT ON COLUMN public."gfg_game"."groups" IS '游戏相关社群';
COMMENT ON COLUMN public."gfg_game"."release_date" IS '发行日期';
COMMENT ON COLUMN public."gfg_game"."developers" IS '开发商';
COMMENT ON COLUMN public."gfg_game"."publishers" IS '发行商';
COMMENT ON COLUMN public."gfg_game"."appid" IS 'SteamAPI appid';
COMMENT ON COLUMN public."gfg_game"."header" IS '游戏封面图';
COMMENT ON COLUMN public."gfg_game"."links" IS '三方网站链接';
COMMENT ON COLUMN public."gfg_game"."weight" IS '权重';
COMMENT ON COLUMN public."gfg_game"."primary_tag" IS '主标签';
COMMENT ON COLUMN public."gfg_game"."secondary_tag" IS '次标签';
COMMENT ON TABLE public."gfg_game_comment" IS '评论表';
COMMENT ON COLUMN public."gfg_game_comment"."id" IS '评论表ID';
COMMENT ON COLUMN public."gfg_game_comment"."region" IS '地区';
COMMENT ON COLUMN public."gfg_game_comment"."content" IS '评论';
COMMENT ON COLUMN public."gfg_game_comment"."score" IS '评分';
COMMENT ON COLUMN public."gfg_game_comment"."create_time" IS '创建时间';
COMMENT ON COLUMN public."gfg_game_comment"."game_id" IS '游戏表ID';
COMMENT ON COLUMN public."gfg_game_comment"."ip" IS 'ip';
COMMENT ON COLUMN public."gfg_game_comment"."name" IS '评论人名称';
COMMENT ON TABLE public."gfg_game_creator_deprecated_20260614" IS 'Deprecated on 2026-06-14 after removing game creator directory, admin CRUD, public API, and RAG sync source.';
COMMENT ON COLUMN public."gfg_game_creator_deprecated_20260614"."id" IS '相关作者表id';
COMMENT ON COLUMN public."gfg_game_creator_deprecated_20260614"."name" IS '名称';
COMMENT ON COLUMN public."gfg_game_creator_deprecated_20260614"."info" IS '相关描述';
COMMENT ON COLUMN public."gfg_game_creator_deprecated_20260614"."main_url" IS '主链接';
COMMENT ON COLUMN public."gfg_game_creator_deprecated_20260614"."links" IS '其他链接';
COMMENT ON COLUMN public."gfg_game_creator_deprecated_20260614"."cover" IS '封面图';
COMMENT ON COLUMN public."gfg_game_creator_deprecated_20260614"."contact" IS '联系方式';
COMMENT ON COLUMN public."gfg_game_creator_deprecated_20260614"."create_time" IS '创建时间';
COMMENT ON COLUMN public."gfg_game_creator_deprecated_20260614"."update_time" IS '修改时间';
COMMENT ON COLUMN public."gfg_game_creator_deprecated_20260614"."type" IS '类型描述 1=Steam鉴赏家 2=博主 3=开发者 4=发行者 5=汉化者 6=内容创作者';
COMMENT ON COLUMN public."gfg_game_creator_deprecated_20260614"."name_en" IS '英文名';
COMMENT ON COLUMN public."gfg_game_creator_deprecated_20260614"."info_en" IS '英文描述';
COMMENT ON COLUMN public."gfg_game_creator_deprecated_20260614"."deleted" IS '软删除';
COMMENT ON TABLE public."gfg_game_news" IS '游戏更新公告记录表';
COMMENT ON COLUMN public."gfg_game_news"."id" IS '游戏更新公告记录表id';
COMMENT ON COLUMN public."gfg_game_news"."game_id" IS '游戏表id';
COMMENT ON COLUMN public."gfg_game_news"."headline" IS '更新公告标题';
COMMENT ON COLUMN public."gfg_game_news"."content" IS '更新公告内容';
COMMENT ON COLUMN public."gfg_game_news"."index" IS '更新公告编号';
COMMENT ON COLUMN public."gfg_game_news"."post_time" IS '更新公告上传日期';
COMMENT ON COLUMN public."gfg_game_news"."create_time" IS '采集时间';
COMMENT ON COLUMN public."gfg_game_news"."author" IS '公告作者';
COMMENT ON COLUMN public."gfg_game_news"."url" IS '更新公告原始地址';
COMMENT ON COLUMN public."gfg_game_news"."total" IS '公告总数';
COMMENT ON COLUMN public."gfg_game_news"."lang" IS '记录的语言';
COMMENT ON TABLE public."gfg_game_player_count" IS '在线人数表';
COMMENT ON COLUMN public."gfg_game_player_count"."id" IS '在线人数表ID';
COMMENT ON COLUMN public."gfg_game_player_count"."game_id" IS '游戏表ID';
COMMENT ON COLUMN public."gfg_game_player_count"."count" IS '在线人数';
COMMENT ON COLUMN public."gfg_game_player_count"."create_time" IS '创建时间';
COMMENT ON TABLE public."gfg_game_record" IS '游戏记录表';
COMMENT ON COLUMN public."gfg_game_record"."id" IS '游戏记录表id';
COMMENT ON COLUMN public."gfg_game_record"."game_id" IS '游戏表id';
COMMENT ON COLUMN public."gfg_game_record"."language" IS '支持语言';
COMMENT ON COLUMN public."gfg_game_record"."release_date" IS '发行时间';
COMMENT ON COLUMN public."gfg_game_record"."platform" IS '支持平台';
COMMENT ON COLUMN public."gfg_game_record"."developer" IS '开发商';
COMMENT ON COLUMN public."gfg_game_record"."publisher" IS '发行商';
COMMENT ON COLUMN public."gfg_game_record"."info" IS '游戏概述';
COMMENT ON COLUMN public."gfg_game_record"."cover" IS '封面图';
COMMENT ON COLUMN public."gfg_game_record"."lang" IS '记录的语言';
COMMENT ON COLUMN public."gfg_game_record"."price_list" IS '游戏价格列表';
COMMENT ON COLUMN public."gfg_game_record"."initial" IS '游戏价格';
COMMENT ON COLUMN public."gfg_game_record"."final" IS '当前价格';
COMMENT ON COLUMN public."gfg_game_record"."discount" IS '折扣百分比';
COMMENT ON TABLE public."gfg_game_v2_recommendations" IS 'Precomputed game v2 similar recommendations. One algorithm_version represents one scoring contract.';
COMMENT ON COLUMN public."gfg_game_v2_recommendations"."score" IS 'Raw hybrid content similarity score in range 0..1.';
COMMENT ON COLUMN public."gfg_game_v2_recommendations"."display_score" IS 'Presentation score in range 0..1 after non-linear stretching.';
COMMENT ON COLUMN public."gfg_game_v2_recommendations"."reason_json" IS 'Short explainable recommendation reasons for UI display and later tuning.';
COMMENT ON TABLE public."gfg_prize" IS '抽奖活动表';
COMMENT ON COLUMN public."gfg_prize"."id" IS '抽奖活动表id';
COMMENT ON COLUMN public."gfg_prize"."title" IS '标题';
COMMENT ON COLUMN public."gfg_prize"."desc" IS '描述';
COMMENT ON COLUMN public."gfg_prize"."prize" IS '奖品';
COMMENT ON COLUMN public."gfg_prize"."key" IS '参与密钥';
COMMENT ON COLUMN public."gfg_prize"."start_time" IS '开始时间';
COMMENT ON COLUMN public."gfg_prize"."end_time" IS '结束时间';
COMMENT ON COLUMN public."gfg_prize"."create_time" IS '创建时间';
COMMENT ON COLUMN public."gfg_prize"."status" IS '状态';
COMMENT ON TABLE public."gfg_prize_member" IS '抽奖活动参与表';
COMMENT ON COLUMN public."gfg_prize_member"."id" IS '抽奖活动参与表id';
COMMENT ON COLUMN public."gfg_prize_member"."prize_id" IS '抽奖活动id';
COMMENT ON COLUMN public."gfg_prize_member"."name" IS '参与者名称';
COMMENT ON COLUMN public."gfg_prize_member"."email" IS '参与者邮箱';
COMMENT ON COLUMN public."gfg_prize_member"."ip" IS '参与者ip';
COMMENT ON COLUMN public."gfg_prize_member"."agent" IS 'User-Agent';
COMMENT ON COLUMN public."gfg_prize_member"."is_winner" IS '是否获奖';
COMMENT ON COLUMN public."gfg_prize_member"."prize_key" IS '获奖key';
COMMENT ON COLUMN public."gfg_prize_member"."create_time" IS '创建时间';
COMMENT ON TABLE public."gfg_tag" IS '游戏标签表';
COMMENT ON COLUMN public."gfg_tag"."id" IS '标签表id';
COMMENT ON COLUMN public."gfg_tag"."name" IS '标签名称';
COMMENT ON COLUMN public."gfg_tag"."name_en" IS '标签英文名称';
COMMENT ON COLUMN public."gfg_tag"."info" IS '标签简介';
COMMENT ON COLUMN public."gfg_tag"."info_en" IS '标签英文简介';
COMMENT ON COLUMN public."gfg_tag"."prefix" IS '父标签 没有为-1';
COMMENT ON COLUMN public."gfg_tag"."create_time" IS '创建时间';
COMMENT ON COLUMN public."gfg_tag"."update_time" IS '修改时间';
COMMENT ON TABLE public."gfg_tag_map" IS '游戏标签映射表';
COMMENT ON COLUMN public."gfg_tag_map"."id" IS '游戏标签映射表id';
COMMENT ON COLUMN public."gfg_tag_map"."game_id" IS '游戏id';
COMMENT ON COLUMN public."gfg_tag_map"."tag_id" IS '标签id';
COMMENT ON COLUMN public."gfg_tag_map"."create_time" IS '创建时间';
COMMENT ON COLUMN public."gfg_tag_map"."update_time" IS '修改时间';

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.gfg_game_v2_prune_detail_snapshots(
    p_appid bigint,
    p_lang text,
    p_region text,
    p_keep_count integer DEFAULT 5
)
RETURNS integer
LANGUAGE plpgsql
AS $function$
DECLARE
    deleted_count INTEGER;
BEGIN
    WITH ranked AS (
        SELECT id,
               row_number() OVER (
                   PARTITION BY appid, lang, region
                   ORDER BY collected_at DESC, id DESC
               ) AS rn
        FROM gfg_game_v2_detail_snapshots
        WHERE appid = p_appid
          AND lang = p_lang
          AND region = p_region
    ),
    deleted AS (
        DELETE FROM gfg_game_v2_detail_snapshots s
        USING ranked r
        WHERE s.id = r.id
          AND r.rn > GREATEST(p_keep_count, 0)
        RETURNING s.id
    )
    SELECT count(*) INTO deleted_count FROM deleted;

    RETURN deleted_count;
END;
$function$;
-- +goose StatementEnd


-- Baseline intentionally has no Down section: dropping an application schema is not a safe rollback.
