-- Canonical gfn baseline, version 20260823000000.
-- Derived from the audited production schema and normalized to omit owners, ACLs, and data.
-- +goose Up

CREATE TABLE public."gfn_collector_domain" (
    "id" bigint NOT NULL,
    "name" character varying(255) NOT NULL,
    "proxy" character varying(4) NOT NULL,
    "prefix" character varying(255),
    "tls" character varying(4) NOT NULL,
    "site_id" bigint,
    "deleted" boolean DEFAULT false NOT NULL
);

CREATE TABLE public."gfn_collector_log_dns" (
    "id" bigint NOT NULL,
    "name" character varying(255) NOT NULL,
    "a" jsonb,
    "aaaa" jsonb,
    "mx" jsonb,
    "ns" jsonb,
    "soa" jsonb,
    "txt" jsonb,
    "caa" jsonb,
    "cname" jsonb,
    "status" character varying(20) NOT NULL,
    "create_time" timestamp(0) without time zone NOT NULL
);

CREATE TABLE public."gfn_collector_log_http" (
    "id" bigint NOT NULL,
    "name" character varying(255) NOT NULL,
    "info" jsonb NOT NULL,
    "status" character varying(20) NOT NULL,
    "create_time" timestamp(0) without time zone NOT NULL
);

CREATE TABLE public."gfn_collector_log_ping" (
    "id" bigint NOT NULL,
    "name" character varying(255) NOT NULL,
    "delay" character varying(20) NOT NULL,
    "loss" character varying(20) NOT NULL,
    "status" character varying(20) NOT NULL,
    "create_time" timestamp(0) without time zone NOT NULL
);

CREATE TABLE public."gfn_collector_observation" (
    "id" bigint NOT NULL,
    "site_id" bigint NOT NULL,
    "target" character varying(255) NOT NULL,
    "protocol" character varying(16) NOT NULL,
    "status" character varying(32) NOT NULL,
    "observed_at" timestamp with time zone NOT NULL,
    "duration_ms" bigint,
    "error_code" character varying(64),
    "error_message" text,
    "payload" jsonb DEFAULT '{}'::jsonb NOT NULL,
    "schema_version" integer DEFAULT 1 NOT NULL,
    "create_time" timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public."gfn_featured_site" (
    "id" bigint NOT NULL,
    "site_id" bigint NOT NULL,
    "weight" bigint DEFAULT 0 NOT NULL,
    "create_time" timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "update_time" timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE public."gfn_log_update" (
    "id" bigint NOT NULL,
    "title" character varying(100) NOT NULL,
    "url" character varying(255) NOT NULL,
    "create_time" timestamp(0) without time zone NOT NULL,
    "update_time" timestamp(0) without time zone NOT NULL,
    "deleted" boolean NOT NULL
);

CREATE TABLE public."gfn_nav_update_notice" (
    "id" bigint NOT NULL,
    "title" character varying(120) NOT NULL,
    "title_en" character varying(120) NOT NULL,
    "body" text NOT NULL,
    "body_en" text NOT NULL,
    "published_at" timestamp(0) without time zone NOT NULL,
    "create_time" timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "update_time" timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "deleted" boolean DEFAULT false NOT NULL
);

CREATE TABLE public."gfn_saying" (
    "id" bigint NOT NULL,
    "author" character varying(255),
    "saying" character varying(255) NOT NULL,
    "create_time" timestamp(0) without time zone NOT NULL,
    "update_time" timestamp(0) without time zone NOT NULL,
    "language" character varying(8) DEFAULT 'zh'::character varying NOT NULL
);

CREATE TABLE public."gfn_site" (
    "id" bigint NOT NULL,
    "name" character varying(255) NOT NULL,
    "name_en" character varying(255) NOT NULL,
    "info" text NOT NULL,
    "info_en" text NOT NULL,
    "create_time" timestamp(0) without time zone NOT NULL,
    "update_time" timestamp(0) without time zone NOT NULL,
    "country" character varying(20),
    "nsfw" character varying(4) DEFAULT ''::character varying NOT NULL,
    "welfare" character varying(4) NOT NULL,
    "icon" character varying(255),
    "deleted" boolean NOT NULL,
    "view_count" bigint DEFAULT 0 NOT NULL
);

CREATE TABLE public."gfn_site_group" (
    "id" bigint NOT NULL,
    "name" character varying(255) NOT NULL,
    "name_en" character varying(255) NOT NULL,
    "info" character varying(255) NOT NULL,
    "info_en" character varying(255) NOT NULL,
    "priority" bigint NOT NULL,
    "create_time" timestamp(0) without time zone NOT NULL,
    "update_time" timestamp(0) without time zone NOT NULL
);

CREATE TABLE public."gfn_site_group_map" (
    "id" bigint NOT NULL,
    "site_id" bigint NOT NULL,
    "group_id" bigint NOT NULL,
    "create_time" timestamp(0) without time zone NOT NULL,
    "update_time" timestamp(0) without time zone NOT NULL,
    "weight" bigint DEFAULT 0 NOT NULL
);

ALTER TABLE ONLY public."gfn_collector_domain" ADD CONSTRAINT "gfn_collector_domain_deleted_not_null" NOT NULL deleted;
ALTER TABLE ONLY public."gfn_collector_domain" ADD CONSTRAINT "gfn_collector_domain_id_not_null" NOT NULL id;
ALTER TABLE ONLY public."gfn_collector_domain" ADD CONSTRAINT "gfn_collector_domain_name_not_null" NOT NULL name;
ALTER TABLE ONLY public."gfn_collector_domain" ADD CONSTRAINT "gfn_collector_domain_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfn_collector_domain" ADD CONSTRAINT "gfn_collector_domain_primary" UNIQUE (id, name);
ALTER TABLE ONLY public."gfn_collector_domain" ADD CONSTRAINT "gfn_collector_domain_proxy_not_null" NOT NULL proxy;
ALTER TABLE ONLY public."gfn_collector_domain" ADD CONSTRAINT "gfn_collector_domain_tls_not_null" NOT NULL tls;
ALTER TABLE ONLY public."gfn_collector_log_dns" ADD CONSTRAINT "gfn_collector_log_dns_create_time_not_null" NOT NULL create_time;
ALTER TABLE ONLY public."gfn_collector_log_dns" ADD CONSTRAINT "gfn_collector_log_dns_id_not_null" NOT NULL id;
ALTER TABLE ONLY public."gfn_collector_log_dns" ADD CONSTRAINT "gfn_collector_log_dns_name_not_null" NOT NULL name;
ALTER TABLE ONLY public."gfn_collector_log_dns" ADD CONSTRAINT "gfn_collector_log_dns_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfn_collector_log_dns" ADD CONSTRAINT "gfn_collector_log_dns_status_not_null" NOT NULL status;
ALTER TABLE ONLY public."gfn_collector_log_http" ADD CONSTRAINT "gfn_collector_log_http_create_time_not_null" NOT NULL create_time;
ALTER TABLE ONLY public."gfn_collector_log_http" ADD CONSTRAINT "gfn_collector_log_http_id_not_null" NOT NULL id;
ALTER TABLE ONLY public."gfn_collector_log_http" ADD CONSTRAINT "gfn_collector_log_http_info_not_null" NOT NULL info;
ALTER TABLE ONLY public."gfn_collector_log_http" ADD CONSTRAINT "gfn_collector_log_http_name_not_null" NOT NULL name;
ALTER TABLE ONLY public."gfn_collector_log_http" ADD CONSTRAINT "gfn_collector_log_http_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfn_collector_log_http" ADD CONSTRAINT "gfn_collector_log_http_status_not_null" NOT NULL status;
ALTER TABLE ONLY public."gfn_collector_log_ping" ADD CONSTRAINT "gfn_collector_log_ping_create_time_not_null" NOT NULL create_time;
ALTER TABLE ONLY public."gfn_collector_log_ping" ADD CONSTRAINT "gfn_collector_log_ping_delay_not_null" NOT NULL delay;
ALTER TABLE ONLY public."gfn_collector_log_ping" ADD CONSTRAINT "gfn_collector_log_ping_id_not_null" NOT NULL id;
ALTER TABLE ONLY public."gfn_collector_log_ping" ADD CONSTRAINT "gfn_collector_log_ping_loss_not_null" NOT NULL loss;
ALTER TABLE ONLY public."gfn_collector_log_ping" ADD CONSTRAINT "gfn_collector_log_ping_name_not_null" NOT NULL name;
ALTER TABLE ONLY public."gfn_collector_log_ping" ADD CONSTRAINT "gfn_collector_log_ping_status_not_null" NOT NULL status;
ALTER TABLE ONLY public."gfn_collector_log_ping" ADD CONSTRAINT "gfn_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfn_collector_observation" ADD CONSTRAINT "gfn_collector_observation_create_time_not_null" NOT NULL create_time;
ALTER TABLE ONLY public."gfn_collector_observation" ADD CONSTRAINT "gfn_collector_observation_id_not_null" NOT NULL id;
ALTER TABLE ONLY public."gfn_collector_observation" ADD CONSTRAINT "gfn_collector_observation_observed_at_not_null" NOT NULL observed_at;
ALTER TABLE ONLY public."gfn_collector_observation" ADD CONSTRAINT "gfn_collector_observation_payload_not_null" NOT NULL payload;
ALTER TABLE ONLY public."gfn_collector_observation" ADD CONSTRAINT "gfn_collector_observation_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfn_collector_observation" ADD CONSTRAINT "gfn_collector_observation_protocol_not_null" NOT NULL protocol;
ALTER TABLE ONLY public."gfn_collector_observation" ADD CONSTRAINT "gfn_collector_observation_schema_version_not_null" NOT NULL schema_version;
ALTER TABLE ONLY public."gfn_collector_observation" ADD CONSTRAINT "gfn_collector_observation_site_id_not_null" NOT NULL site_id;
ALTER TABLE ONLY public."gfn_collector_observation" ADD CONSTRAINT "gfn_collector_observation_status_not_null" NOT NULL status;
ALTER TABLE ONLY public."gfn_collector_observation" ADD CONSTRAINT "gfn_collector_observation_target_not_null" NOT NULL target;
ALTER TABLE ONLY public."gfn_featured_site" ADD CONSTRAINT "gfn_featured_site_create_time_not_null" NOT NULL create_time;
ALTER TABLE ONLY public."gfn_featured_site" ADD CONSTRAINT "gfn_featured_site_id_not_null" NOT NULL id;
ALTER TABLE ONLY public."gfn_featured_site" ADD CONSTRAINT "gfn_featured_site_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfn_featured_site" ADD CONSTRAINT "gfn_featured_site_site_id_not_null" NOT NULL site_id;
ALTER TABLE ONLY public."gfn_featured_site" ADD CONSTRAINT "gfn_featured_site_update_time_not_null" NOT NULL update_time;
ALTER TABLE ONLY public."gfn_featured_site" ADD CONSTRAINT "gfn_featured_site_weight_not_null" NOT NULL weight;
ALTER TABLE ONLY public."gfn_log_update" ADD CONSTRAINT "gfn_log_update_create_time_not_null" NOT NULL create_time;
ALTER TABLE ONLY public."gfn_log_update" ADD CONSTRAINT "gfn_log_update_deleted_not_null" NOT NULL deleted;
ALTER TABLE ONLY public."gfn_log_update" ADD CONSTRAINT "gfn_log_update_id_not_null" NOT NULL id;
ALTER TABLE ONLY public."gfn_log_update" ADD CONSTRAINT "gfn_log_update_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfn_log_update" ADD CONSTRAINT "gfn_log_update_title_not_null" NOT NULL title;
ALTER TABLE ONLY public."gfn_log_update" ADD CONSTRAINT "gfn_log_update_update_time_not_null" NOT NULL update_time;
ALTER TABLE ONLY public."gfn_log_update" ADD CONSTRAINT "gfn_log_update_url_not_null" NOT NULL url;
ALTER TABLE ONLY public."gfn_nav_update_notice" ADD CONSTRAINT "gfn_nav_update_notice_body_en_not_null" NOT NULL body_en;
ALTER TABLE ONLY public."gfn_nav_update_notice" ADD CONSTRAINT "gfn_nav_update_notice_body_not_null" NOT NULL body;
ALTER TABLE ONLY public."gfn_nav_update_notice" ADD CONSTRAINT "gfn_nav_update_notice_create_time_not_null" NOT NULL create_time;
ALTER TABLE ONLY public."gfn_nav_update_notice" ADD CONSTRAINT "gfn_nav_update_notice_deleted_not_null" NOT NULL deleted;
ALTER TABLE ONLY public."gfn_nav_update_notice" ADD CONSTRAINT "gfn_nav_update_notice_id_not_null" NOT NULL id;
ALTER TABLE ONLY public."gfn_nav_update_notice" ADD CONSTRAINT "gfn_nav_update_notice_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfn_nav_update_notice" ADD CONSTRAINT "gfn_nav_update_notice_published_at_not_null" NOT NULL published_at;
ALTER TABLE ONLY public."gfn_nav_update_notice" ADD CONSTRAINT "gfn_nav_update_notice_title_en_not_null" NOT NULL title_en;
ALTER TABLE ONLY public."gfn_nav_update_notice" ADD CONSTRAINT "gfn_nav_update_notice_title_not_null" NOT NULL title;
ALTER TABLE ONLY public."gfn_nav_update_notice" ADD CONSTRAINT "gfn_nav_update_notice_update_time_not_null" NOT NULL update_time;
ALTER TABLE ONLY public."gfn_saying" ADD CONSTRAINT "chk_gfn_saying_language" CHECK (language::text = ANY (ARRAY['zh'::character varying, 'en'::character varying]::text[]));
ALTER TABLE ONLY public."gfn_saying" ADD CONSTRAINT "gfn_saying_create_time_not_null" NOT NULL create_time;
ALTER TABLE ONLY public."gfn_saying" ADD CONSTRAINT "gfn_saying_id_not_null" NOT NULL id;
ALTER TABLE ONLY public."gfn_saying" ADD CONSTRAINT "gfn_saying_language_not_null" NOT NULL language;
ALTER TABLE ONLY public."gfn_saying" ADD CONSTRAINT "gfn_saying_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfn_saying" ADD CONSTRAINT "gfn_saying_primary" UNIQUE (id, saying);
ALTER TABLE ONLY public."gfn_saying" ADD CONSTRAINT "gfn_saying_saying_not_null" NOT NULL saying;
ALTER TABLE ONLY public."gfn_saying" ADD CONSTRAINT "gfn_saying_update_time_not_null" NOT NULL update_time;
ALTER TABLE ONLY public."gfn_site" ADD CONSTRAINT "gfn_site_create_time_not_null" NOT NULL create_time;
ALTER TABLE ONLY public."gfn_site" ADD CONSTRAINT "gfn_site_deleted_not_null" NOT NULL deleted;
ALTER TABLE ONLY public."gfn_site" ADD CONSTRAINT "gfn_site_id_not_null" NOT NULL id;
ALTER TABLE ONLY public."gfn_site" ADD CONSTRAINT "gfn_site_info_en_not_null" NOT NULL info_en;
ALTER TABLE ONLY public."gfn_site" ADD CONSTRAINT "gfn_site_info_not_null" NOT NULL info;
ALTER TABLE ONLY public."gfn_site" ADD CONSTRAINT "gfn_site_name_en_not_null" NOT NULL name_en;
ALTER TABLE ONLY public."gfn_site" ADD CONSTRAINT "gfn_site_name_not_null" NOT NULL name;
ALTER TABLE ONLY public."gfn_site" ADD CONSTRAINT "gfn_site_nsfw_not_null" NOT NULL nsfw;
ALTER TABLE ONLY public."gfn_site" ADD CONSTRAINT "gfn_site_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfn_site" ADD CONSTRAINT "gfn_site_update_time_not_null" NOT NULL update_time;
ALTER TABLE ONLY public."gfn_site" ADD CONSTRAINT "gfn_site_view_count_not_null" NOT NULL view_count;
ALTER TABLE ONLY public."gfn_site" ADD CONSTRAINT "gfn_site_welfare_not_null" NOT NULL welfare;
ALTER TABLE ONLY public."gfn_site_group" ADD CONSTRAINT "gfn_site_group_create_time_not_null" NOT NULL create_time;
ALTER TABLE ONLY public."gfn_site_group" ADD CONSTRAINT "gfn_site_group_id_not_null" NOT NULL id;
ALTER TABLE ONLY public."gfn_site_group" ADD CONSTRAINT "gfn_site_group_info_en_not_null" NOT NULL info_en;
ALTER TABLE ONLY public."gfn_site_group" ADD CONSTRAINT "gfn_site_group_info_not_null" NOT NULL info;
ALTER TABLE ONLY public."gfn_site_group" ADD CONSTRAINT "gfn_site_group_name_en_not_null" NOT NULL name_en;
ALTER TABLE ONLY public."gfn_site_group" ADD CONSTRAINT "gfn_site_group_name_not_null" NOT NULL name;
ALTER TABLE ONLY public."gfn_site_group" ADD CONSTRAINT "gfn_site_group_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfn_site_group" ADD CONSTRAINT "gfn_site_group_primary" UNIQUE (id, name, name_en, priority);
ALTER TABLE ONLY public."gfn_site_group" ADD CONSTRAINT "gfn_site_group_priority_not_null" NOT NULL priority;
ALTER TABLE ONLY public."gfn_site_group" ADD CONSTRAINT "gfn_site_group_update_time_not_null" NOT NULL update_time;
ALTER TABLE ONLY public."gfn_site_group_map" ADD CONSTRAINT "gfn_site_group_map_create_time_not_null" NOT NULL create_time;
ALTER TABLE ONLY public."gfn_site_group_map" ADD CONSTRAINT "gfn_site_group_map_group_id_not_null" NOT NULL group_id;
ALTER TABLE ONLY public."gfn_site_group_map" ADD CONSTRAINT "gfn_site_group_map_id_not_null" NOT NULL id;
ALTER TABLE ONLY public."gfn_site_group_map" ADD CONSTRAINT "gfn_site_group_map_pkey" PRIMARY KEY (id);
ALTER TABLE ONLY public."gfn_site_group_map" ADD CONSTRAINT "gfn_site_group_map_site_id_not_null" NOT NULL site_id;
ALTER TABLE ONLY public."gfn_site_group_map" ADD CONSTRAINT "gfn_site_group_map_update_time_not_null" NOT NULL update_time;
ALTER TABLE ONLY public."gfn_site_group_map" ADD CONSTRAINT "gfn_site_group_map_weight_not_null" NOT NULL weight;

CREATE INDEX idx_gfn_collector_domain_deleted ON public.gfn_collector_domain USING btree (deleted);
CREATE INDEX idx_gfn_collector_domain_name ON public.gfn_collector_domain USING btree (name);
CREATE INDEX idx_gfn_collector_domain_prefix ON public.gfn_collector_domain USING btree (prefix);
CREATE INDEX idx_gfn_collector_domain_proxy_tls ON public.gfn_collector_domain USING btree (proxy, tls);
CREATE INDEX idx_gfn_collector_domain_site_id ON public.gfn_collector_domain USING btree (site_id);
CREATE INDEX idx_gfn_collector_log_dns_create_time ON public.gfn_collector_log_dns USING btree (create_time DESC);
CREATE INDEX idx_gfn_collector_log_dns_name_create_time_id ON public.gfn_collector_log_dns USING btree (name, create_time DESC, id);
CREATE INDEX idx_gfn_collector_log_http_create_time ON public.gfn_collector_log_http USING btree (create_time DESC);
CREATE INDEX idx_gfn_collector_log_http_name_create_time_id ON public.gfn_collector_log_http USING btree (name, create_time DESC, id);
CREATE INDEX idx_gfn_collector_log_ping_create_time ON public.gfn_collector_log_ping USING btree (create_time DESC);
CREATE INDEX idx_gfn_collector_log_ping_name_create_time_id ON public.gfn_collector_log_ping USING btree (name, create_time DESC, id);
CREATE INDEX idx_gfn_collector_observation_protocol_site_time_id ON public.gfn_collector_observation USING btree (protocol, site_id, observed_at DESC, id DESC);
CREATE INDEX idx_gfn_collector_observation_protocol_time_id ON public.gfn_collector_observation USING btree (protocol, observed_at DESC, id DESC);
CREATE INDEX idx_gfn_collector_observation_site_protocol_time_id ON public.gfn_collector_observation USING btree (site_id, protocol, observed_at DESC, id DESC);
CREATE INDEX idx_gfn_collector_observation_site_target_protocol_time_id ON public.gfn_collector_observation USING btree (site_id, target, protocol, observed_at DESC, id DESC);
CREATE UNIQUE INDEX idx_gfn_featured_site_site_id ON public.gfn_featured_site USING btree (site_id);
CREATE INDEX idx_gfn_featured_site_weight ON public.gfn_featured_site USING btree (weight DESC, id DESC);
CREATE INDEX idx_gfn_nav_update_notice_public ON public.gfn_nav_update_notice USING btree (deleted, published_at DESC, id DESC);
CREATE INDEX idx_gfn_saying_author ON public.gfn_saying USING btree (author);
CREATE INDEX idx_gfn_saying_create_time ON public.gfn_saying USING btree (create_time DESC);
CREATE INDEX idx_gfn_saying_language ON public.gfn_saying USING btree (language);
CREATE INDEX idx_gfn_saying_saying ON public.gfn_saying USING btree (saying);
CREATE INDEX idx_gfn_saying_update_time ON public.gfn_saying USING btree (update_time DESC);
CREATE INDEX idx_gfn_site_country_nsfw_welfare ON public.gfn_site USING btree (country, nsfw, welfare);
CREATE INDEX idx_gfn_site_create_time ON public.gfn_site USING btree (create_time DESC);
CREATE INDEX idx_gfn_site_deleted ON public.gfn_site USING btree (deleted);
CREATE INDEX idx_gfn_site_name ON public.gfn_site USING btree (name);
CREATE INDEX idx_gfn_site_name_en ON public.gfn_site USING btree (name_en);
CREATE INDEX idx_gfn_site_update_time ON public.gfn_site USING btree (update_time DESC);
CREATE INDEX idx_gfn_site_group_create_time ON public.gfn_site_group USING btree (create_time DESC);
CREATE INDEX idx_gfn_site_group_name ON public.gfn_site_group USING btree (name);
CREATE INDEX idx_gfn_site_group_name_en ON public.gfn_site_group USING btree (name_en);
CREATE INDEX idx_gfn_site_group_priority ON public.gfn_site_group USING btree (priority DESC);
CREATE INDEX idx_gfn_site_group_update_time ON public.gfn_site_group USING btree (update_time DESC);
CREATE INDEX idx_gfn_site_group_map_create_time ON public.gfn_site_group_map USING btree (create_time DESC);
CREATE INDEX idx_gfn_site_group_map_group_id ON public.gfn_site_group_map USING btree (group_id);
CREATE INDEX idx_gfn_site_group_map_group_weight ON public.gfn_site_group_map USING btree (group_id, weight DESC, update_time DESC, id DESC, site_id);
CREATE INDEX idx_gfn_site_group_map_site_group ON public.gfn_site_group_map USING btree (site_id, group_id);
CREATE INDEX idx_gfn_site_group_map_site_id ON public.gfn_site_group_map USING btree (site_id);

COMMENT ON TABLE public."gfn_collector_domain" IS '域名请求表';
COMMENT ON COLUMN public."gfn_collector_domain"."id" IS '域名请求表id';
COMMENT ON COLUMN public."gfn_collector_domain"."name" IS '域名';
COMMENT ON COLUMN public."gfn_collector_domain"."proxy" IS '是否需要代理加速 1 0';
COMMENT ON COLUMN public."gfn_collector_domain"."prefix" IS '是否有前缀';
COMMENT ON COLUMN public."gfn_collector_domain"."tls" IS '是否 https 1 0';
COMMENT ON CONSTRAINT "gfn_collector_domain_primary" ON public."gfn_collector_domain" IS 'id 域名 唯一';
COMMENT ON TABLE public."gfn_collector_log_dns" IS 'DNS日志表';
COMMENT ON COLUMN public."gfn_collector_log_dns"."id" IS 'DNS日志表 id';
COMMENT ON COLUMN public."gfn_collector_log_dns"."name" IS '域名';
COMMENT ON COLUMN public."gfn_collector_log_dns"."a" IS 'A记录';
COMMENT ON COLUMN public."gfn_collector_log_dns"."aaaa" IS 'AAAA记录';
COMMENT ON COLUMN public."gfn_collector_log_dns"."mx" IS 'MX记录';
COMMENT ON COLUMN public."gfn_collector_log_dns"."ns" IS 'NS记录';
COMMENT ON COLUMN public."gfn_collector_log_dns"."soa" IS 'SOA记录';
COMMENT ON COLUMN public."gfn_collector_log_dns"."txt" IS 'TXT记录';
COMMENT ON COLUMN public."gfn_collector_log_dns"."caa" IS 'CAA记录';
COMMENT ON COLUMN public."gfn_collector_log_dns"."cname" IS 'CNAME记录';
COMMENT ON COLUMN public."gfn_collector_log_dns"."status" IS '采集状态 success failure';
COMMENT ON COLUMN public."gfn_collector_log_dns"."create_time" IS '采集时间';
COMMENT ON TABLE public."gfn_collector_log_http" IS 'HTTP请求日志表';
COMMENT ON COLUMN public."gfn_collector_log_http"."id" IS 'http请求日志表';
COMMENT ON COLUMN public."gfn_collector_log_http"."name" IS '域名';
COMMENT ON COLUMN public."gfn_collector_log_http"."info" IS '日志内容';
COMMENT ON COLUMN public."gfn_collector_log_http"."status" IS '请求状态 success failure';
COMMENT ON COLUMN public."gfn_collector_log_http"."create_time" IS '请求时间';
COMMENT ON TABLE public."gfn_collector_log_ping" IS 'Ping日志表';
COMMENT ON COLUMN public."gfn_collector_log_ping"."id" IS 'ping记录表id';
COMMENT ON COLUMN public."gfn_collector_log_ping"."name" IS '域名';
COMMENT ON COLUMN public."gfn_collector_log_ping"."delay" IS '延迟';
COMMENT ON COLUMN public."gfn_collector_log_ping"."loss" IS '丢包';
COMMENT ON COLUMN public."gfn_collector_log_ping"."status" IS '可达性 up down';
COMMENT ON COLUMN public."gfn_collector_log_ping"."create_time" IS '日志时间';
COMMENT ON TABLE public."gfn_log_update" IS '更新公告表';
COMMENT ON COLUMN public."gfn_log_update"."id" IS '更新公告表id';
COMMENT ON COLUMN public."gfn_log_update"."title" IS '更新公告标题';
COMMENT ON COLUMN public."gfn_log_update"."url" IS '更新公告文档地址';
COMMENT ON COLUMN public."gfn_log_update"."create_time" IS '创建时间';
COMMENT ON COLUMN public."gfn_log_update"."update_time" IS '更新时间';
COMMENT ON COLUMN public."gfn_log_update"."deleted" IS '软删除';
COMMENT ON TABLE public."gfn_saying" IS '金句表';
COMMENT ON COLUMN public."gfn_saying"."id" IS '金句表ID';
COMMENT ON COLUMN public."gfn_saying"."author" IS '金句提供者';
COMMENT ON COLUMN public."gfn_saying"."saying" IS '金句';
COMMENT ON COLUMN public."gfn_saying"."create_time" IS '创建时间';
COMMENT ON COLUMN public."gfn_saying"."update_time" IS '修改时间';
COMMENT ON TABLE public."gfn_site" IS '导航站点表';
COMMENT ON COLUMN public."gfn_site"."id" IS '站点表id';
COMMENT ON COLUMN public."gfn_site"."name" IS '站点名称';
COMMENT ON COLUMN public."gfn_site"."name_en" IS '站点名称-英文';
COMMENT ON COLUMN public."gfn_site"."info" IS '站点描述';
COMMENT ON COLUMN public."gfn_site"."info_en" IS '站点描述-英文';
COMMENT ON COLUMN public."gfn_site"."create_time" IS '创建时间';
COMMENT ON COLUMN public."gfn_site"."update_time" IS '修改时间';
COMMENT ON COLUMN public."gfn_site"."country" IS '站点所属国家';
COMMENT ON COLUMN public."gfn_site"."nsfw" IS '是否NSFW 1 0';
COMMENT ON COLUMN public."gfn_site"."welfare" IS '是否公益项目 1 0';
COMMENT ON COLUMN public."gfn_site"."icon" IS '站点图标';
COMMENT ON COLUMN public."gfn_site"."deleted" IS '软删除';
COMMENT ON TABLE public."gfn_site_group" IS '导航站点分组表';
COMMENT ON COLUMN public."gfn_site_group"."id" IS '分组表id';
COMMENT ON COLUMN public."gfn_site_group"."name" IS '分组名称';
COMMENT ON COLUMN public."gfn_site_group"."name_en" IS '分组名称-英文';
COMMENT ON COLUMN public."gfn_site_group"."info" IS '分组简介';
COMMENT ON COLUMN public."gfn_site_group"."info_en" IS '分组简介-英文';
COMMENT ON COLUMN public."gfn_site_group"."priority" IS '分组优先级';
COMMENT ON COLUMN public."gfn_site_group"."create_time" IS '创建时间';
COMMENT ON COLUMN public."gfn_site_group"."update_time" IS '修改时间';
COMMENT ON CONSTRAINT "gfn_site_group_primary" ON public."gfn_site_group" IS 'id 名字 优先级 唯一';
COMMENT ON TABLE public."gfn_site_group_map" IS '导航站点分组映射表';
COMMENT ON COLUMN public."gfn_site_group_map"."id" IS '分组映射表id';
COMMENT ON COLUMN public."gfn_site_group_map"."site_id" IS '站点id';
COMMENT ON COLUMN public."gfn_site_group_map"."group_id" IS '分组id';
COMMENT ON COLUMN public."gfn_site_group_map"."create_time" IS '创建时间';
COMMENT ON COLUMN public."gfn_site_group_map"."update_time" IS '修改时间';


-- Baseline intentionally has no Down section: dropping an application schema is not a safe rollback.
