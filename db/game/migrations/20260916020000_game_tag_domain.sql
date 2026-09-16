-- #116: permanent codes are the reviewed Appendix A mapping, never generated from names.
-- This destructive domain migration is intentionally irreversible. Restore a verified
-- backup or recreate the disposable development database to roll back.
-- +goose Up
CREATE TEMP TABLE tag_code_mapping (id bigint PRIMARY KEY, code text NOT NULL UNIQUE) ON COMMIT DROP;
INSERT INTO tag_code_mapping (id, code) VALUES
    (1001, 'fvn'),
    (1002, 'act'),
    (1003, 'competitive'),
    (1004, 'stg'),
    (1005, 'fps'),
    (1006, 'tps'),
    (1007, 'fig'),
    (1008, 'avg'),
    (1009, 'rpg'),
    (1010, 'arpg'),
    (1011, 'jrpg'),
    (1012, 'mmorpg'),
    (1013, 'rpg-maker'),
    (1014, 'adult'),
    (1015, 'soulslike'),
    (1016, 'pixel'),
    (1017, 'grotesque'),
    (1018, 'party'),
    (1019, 'musou'),
    (1020, 'bullet-hell'),
    (1021, 'level-based'),
    (1022, 'desktop-pet'),
    (1023, 'metroidvania'),
    (1024, 'hidden-object'),
    (1025, 'stealth'),
    (1026, 'idle'),
    (1027, 'story-exploration'),
    (1028, 'battle-royale'),
    (1029, 'walking-simulator'),
    (1030, 'platformer'),
    (1031, 'mystery'),
    (1032, 'lighthearted'),
    (1033, 'roguelike'),
    (1034, 'multiplayer'),
    (1035, 'strategy'),
    (1036, 'rts'),
    (1037, 'rtt'),
    (1038, 'moba'),
    (1039, 'tower-defense'),
    (1040, 'simulation'),
    (1041, 'management'),
    (1042, 'raising'),
    (1043, 'flight'),
    (1044, 'rhythm-game'),
    (1045, 'sports'),
    (1046, 'racing'),
    (1047, 'parkour'),
    (1048, 'card-game'),
    (1049, 'deck-building'),
    (1050, 'board-and-card'),
    (1051, 'sandbox'),
    (1052, 'survival'),
    (1053, 'afk'),
    (1054, 'fishing'),
    (1055, 'dress-up'),
    (1056, 'match-3'),
    (1057, 'dating'),
    (1058, 'puzzle'),
    (1059, 'puzzle-solving'),
    (1060, 'tactical-rpg'),
    (1061, 'serious'),
    (1062, 'text-based'),
    (1063, 'adventure-puzzle'),
    (1064, 'open-world'),
    (1065, 'strategy-management'),
    (1066, 'casual'),
    (1067, 'vr'),
    (1068, 'motion-control'),
    (1069, 'side-scrolling'),
    (1070, 'vertical-scrolling'),
    (1071, 'turn-based'),
    (1072, 'real-time'),
    (1074, 'simulation-construction'),
    (1075, 'social'),
    (1076, '3d'),
    (1077, '2d'),
    (1078, 'horror'),
    (1079, 'suspense'),
    (1080, 'decryption'),
    (1081, 'detective'),
    (1082, 'loot-farming'),
    (1083, 'medieval'),
    (1084, 'ancient-style'),
    (1085, 'third-person'),
    (1086, 'free'),
    (1087, 'sci-fi'),
    (1088, 'cyberpunk'),
    (1089, 'steampunk'),
    (1090, 'rich-narrative'),
    (1091, 'fantasy'),
    (1092, 'arcade'),
    (1093, 'retro'),
    (1094, 'violent'),
    (1095, 'dark'),
    (1096, 'linear'),
    (1097, 'magic'),
    (1098, 'tactical'),
    (1099, 'futuristic'),
    (1100, 'crafting'),
    (1101, 'difficult'),
    (1102, 'post-apocalyptic'),
    (1103, 'educational'),
    (1104, 'war'),
    (1105, 'collecting'),
    (1106, 'cooking'),
    (1107, 'auto-chess'),
    (1108, 'mini-games'),
    (1109, 'warmth'),
    (1110, 'collapse'),
    (1111, 'thriller'),
    (1112, 'cruel'),
    (1113, 'healing'),
    (1114, 'rhythm'),
    (1115, 'female-oriented'),
    (1116, 'male-oriented'),
    (1117, 'farming'),
    (1118, 'co-op'),
    (1119, 'pvp'),
    (1120, 'exploration'),
    (1121, 'stylized'),
    (1122, 'slow-burn'),
    (1123, 'lab'),
    (1124, 'physics'),
    (1125, 'cinematic'),
    (1126, 'level-editor'),
    (1127, 'hunting'),
    (1129, 'precision-platformer'),
    (1130, 'psychological-horror'),
    (1131, 'top-down'),
    (1132, 'interactive-novel'),
    (1133, 'hack-and-slash'),
    (1134, 'hakoniwa'),
    (1135, '2-5d'),
    (1136, 'qte'),
    (1137, 'nonlinear'),
    (1138, 'gore'),
    (1139, 'destruction'),
    (1140, 'military'),
    (1141, 'comedy'),
    (1142, 'incremental'),
    (1143, 'quick-reaction'),
    (1144, 'cute'),
    (1145, 'hand-drawn'),
    (1146, 'cartoon'),
    (1147, 'emotional'),
    (1148, 'atmospheric'),
    (1149, 'utopian'),
    (1150, 'dystopian'),
    (1151, 'surreal'),
    (1152, 'mechanical'),
    (1153, 'robots'),
    (2001, 'feral'),
    (2002, 'anthro'),
    (2003, 'scalie'),
    (2004, 'marine'),
    (2005, 'avian'),
    (2006, 'mammal'),
    (2007, 'reptile'),
    (2008, 'digitigrade'),
    (2009, 'canine'),
    (2010, 'feline'),
    (2011, 'dragon'),
    (2012, 'tiger'),
    (2013, 'lion'),
    (2014, 'wolf'),
    (2015, 'fox'),
    (2016, 'dog'),
    (2017, 'cat'),
    (2018, 'crocodile'),
    (2019, 'shark'),
    (2020, 'bear'),
    (2021, 'lizard'),
    (2022, 'dino'),
    (2023, 'crow'),
    (2024, 'orca'),
    (2025, 'bovine'),
    (2026, 'equine'),
    (2027, 'mouse'),
    (2028, 'boar'),
    (2029, 'goo-creature'),
    (2030, 'furred-dragon'),
    (2031, 'sergal'),
    (2032, 'werewolf'),
    (2033, 'charr'),
    (2034, 'worgen'),
    (2035, 'hrothgar'),
    (2036, 'human'),
    (2037, 'leopard'),
    (2038, 'pokemon-like'),
    (2039, 'deer'),
    (2040, 'lynx'),
    (2041, 'kobold'),
    (2042, 'rabbit'),
    (2043, 'hyena'),
    (3001, 'windows'),
    (3002, 'mac'),
    (3003, 'linux'),
    (3004, 'afdian'),
    (3005, 'kickstarter'),
    (3006, 'patreon'),
    (3007, 'ko-fi'),
    (3008, 'google-play'),
    (3009, 'itch-io'),
    (3010, 'modian'),
    (9001, 'homosexual'),
    (9002, 'futanari'),
    (9003, 'foot-fetish'),
    (9004, 'vore'),
    (9005, 'transfur'),
    (9006, 'shota'),
    (9007, 'femboy'),
    (9008, 'therian'),
    (9009, 'dragon-fetish'),
    (9010, 'kemoryona');

-- Serialize the cutover with current-domain writers.
LOCK TABLE public.gfg_game, public.gfg_tag, public.gfg_tag_map IN ACCESS EXCLUSIVE MODE;

-- +goose StatementBegin
DO $preflight$
BEGIN
    IF EXISTS (SELECT 1 FROM public.gfg_tag) AND (SELECT count(*) FROM public.gfg_tag WHERE prefix=-1) <> 4 THEN
        RAISE EXCEPTION 'tag refactor: expected four legacy parents in a populated catalog';
    END IF;
    IF EXISTS (SELECT 1 FROM public.gfg_tag WHERE prefix = -1 AND id NOT IN (100,200,300,900))
       OR EXISTS (SELECT 1 FROM public.gfg_tag WHERE id IN (100,200,300,900) AND prefix <> -1) THEN
        RAISE EXCEPTION 'tag refactor: unsupported legacy parent';
    END IF;
    IF EXISTS (SELECT 1 FROM public.gfg_tag t LEFT JOIN public.gfg_tag p ON p.id=t.prefix
               WHERE t.prefix <> -1 AND (p.id IS NULL OR p.prefix <> -1 OR p.id NOT IN (100,200,300,900))) THEN
        RAISE EXCEPTION 'tag refactor: orphan or unsupported prefix';
    END IF;
    IF EXISTS (SELECT 1 FROM public.gfg_tag t LEFT JOIN tag_code_mapping m USING(id)
               WHERE t.prefix <> -1 AND m.id IS NULL) THEN
        RAISE EXCEPTION 'tag refactor: unmapped legacy leaf; update the reviewed mapping';
    END IF;
    IF EXISTS (SELECT 1 FROM public.gfg_tag_map m LEFT JOIN public.gfg_game g ON g.id=m.game_id
               LEFT JOIN public.gfg_tag t ON t.id=m.tag_id WHERE g.id IS NULL OR t.id IS NULL OR t.prefix=-1) THEN
        RAISE EXCEPTION 'tag refactor: invalid map or parent assigned to game';
    END IF;
    IF EXISTS (SELECT 1 FROM public.gfg_tag_map GROUP BY game_id,tag_id HAVING count(*) > 1) THEN
        RAISE EXCEPTION 'tag refactor: duplicate legacy game/tag pair';
    END IF;
    IF EXISTS (SELECT 1 FROM public.gfg_game g
               WHERE (g.primary_tag<>0 AND NOT EXISTS (SELECT 1 FROM public.gfg_tag t WHERE t.id=g.primary_tag AND t.prefix<>-1))
                  OR (g.secondary_tag<>0 AND NOT EXISTS (SELECT 1 FROM public.gfg_tag t WHERE t.id=g.secondary_tag AND t.prefix<>-1))) THEN
        RAISE EXCEPTION 'tag refactor: invalid primary or secondary';
    END IF;
    IF EXISTS (SELECT 1 FROM public.gfg_game WHERE primary_tag<>0 AND primary_tag=secondary_tag) THEN
        RAISE EXCEPTION 'tag refactor: primary equals secondary';
    END IF;
END;
$preflight$;
-- +goose StatementEnd

CREATE TEMP TABLE tag_refactor_counts ON COMMIT DROP AS
SELECT (SELECT count(*) FROM public.gfg_tag WHERE prefix<>-1) AS leaves,
       (SELECT count(*) FROM public.gfg_tag_map) AS map_pairs,
       (SELECT count(*) FROM public.gfg_game WHERE primary_tag<>0) AS primaries,
       (SELECT count(*) FROM public.gfg_game WHERE secondary_tag<>0) AS secondaries,
       (SELECT count(*) FROM public.gfg_game g WHERE primary_tag<>0 AND NOT EXISTS
           (SELECT 1 FROM public.gfg_tag_map m WHERE m.game_id=g.id AND m.tag_id=g.primary_tag)) AS missing_primary,
       (SELECT count(*) FROM public.gfg_game g WHERE secondary_tag<>0 AND NOT EXISTS
           (SELECT 1 FROM public.gfg_tag_map m WHERE m.game_id=g.id AND m.tag_id=g.secondary_tag)) AS missing_secondary;

CREATE TABLE public.gfg_tag_category (
    id bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    code text NOT NULL UNIQUE,
    name varchar(255) NOT NULL,
    name_en varchar(255) NOT NULL,
    info varchar(255) NOT NULL,
    info_en varchar(255) NOT NULL,
    sort_order integer NOT NULL,
    archived_at timestamp(0) without time zone,
    create_time timestamp(0) without time zone NOT NULL,
    update_time timestamp(0) without time zone NOT NULL,
    CONSTRAINT gfg_tag_category_code_check CHECK (length(code) BETWEEN 1 AND 64 AND code ~ '^[a-z0-9]+(-[a-z0-9]+)*$')
);

-- Empty fresh databases have no legacy parent rows; the explicit category contract
-- also supplies their initial display values. Existing displays are preserved.
INSERT INTO public.gfg_tag_category (id,code,name,name_en,info,info_en,sort_order,create_time,update_time)
SELECT c.id,c.code,COALESCE(t.name,c.name),COALESCE(t.name_en,c.name_en),
       COALESCE(t.info,''),COALESCE(t.info_en,''),c.sort_order,
       COALESCE(t.create_time,transaction_timestamp() AT TIME ZONE 'UTC'),
       COALESCE(t.update_time,transaction_timestamp() AT TIME ZONE 'UTC')
FROM (VALUES (1,'classification','分类','Categories',10,100),
             (2,'species','物种','Species',20,200),
             (3,'platform','平台','Platforms',30,300),
             (4,'other','其他','Other',40,900)) c(id,code,name,name_en,sort_order,old_id)
LEFT JOIN public.gfg_tag t ON t.id=c.old_id;
SELECT setval(pg_get_serial_sequence('public.gfg_tag_category','id'),4,true);

ALTER TABLE public.gfg_tag ADD COLUMN code text,
    ADD COLUMN category_id bigint,
    ADD COLUMN archived_at timestamp(0) without time zone;
UPDATE public.gfg_tag t SET code=m.code,
    category_id=CASE t.prefix WHEN 100 THEN 1 WHEN 200 THEN 2 WHEN 300 THEN 3 WHEN 900 THEN 4 END
FROM tag_code_mapping m WHERE m.id=t.id;
DELETE FROM public.gfg_tag WHERE prefix=-1;
ALTER TABLE public.gfg_tag ALTER COLUMN code SET NOT NULL,
    ALTER COLUMN category_id SET NOT NULL,
    ADD CONSTRAINT gfg_tag_code_key UNIQUE(code),
    ADD CONSTRAINT gfg_tag_code_check CHECK (length(code) BETWEEN 1 AND 64 AND code ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    ADD CONSTRAINT gfg_tag_category_id_fkey FOREIGN KEY(category_id) REFERENCES public.gfg_tag_category(id) ON DELETE RESTRICT,
    ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY;
-- Include durable historical identities so future allocation cannot reuse an old Tag.
SELECT setval(pg_get_serial_sequence('public.gfg_tag','id'), GREATEST(1,
    COALESCE((SELECT max(id) FROM public.gfg_tag),0),
    COALESCE((SELECT max(tag_id) FROM public.gfg_game_daily d CROSS JOIN LATERAL unnest(d.tag_ids) tag_id),0),
    COALESCE((SELECT max(primary_tag_id) FROM public.gfg_game_daily),0),
    COALESCE((SELECT max(secondary_tag_id) FROM public.gfg_game_daily),0)),true);
CREATE INDEX idx_gfg_tag_category_id ON public.gfg_tag(category_id);

CREATE TABLE public.gfg_game_tag (
    game_id bigint NOT NULL REFERENCES public.gfg_game(id) ON DELETE RESTRICT,
    tag_id bigint NOT NULL REFERENCES public.gfg_tag(id) ON DELETE RESTRICT,
    role text NOT NULL,
    create_time timestamp(0) without time zone NOT NULL,
    update_time timestamp(0) without time zone NOT NULL,
    PRIMARY KEY(game_id,tag_id),
    CONSTRAINT gfg_game_tag_role_check CHECK(role IN ('normal','primary','secondary'))
);
CREATE UNIQUE INDEX idx_gfg_game_tag_primary ON public.gfg_game_tag(game_id) WHERE role='primary';
CREATE UNIQUE INDEX idx_gfg_game_tag_secondary ON public.gfg_game_tag(game_id) WHERE role='secondary';
CREATE INDEX idx_gfg_game_tag_tag_game ON public.gfg_game_tag(tag_id,game_id);
INSERT INTO public.gfg_game_tag(game_id,tag_id,role,create_time,update_time)
SELECT game_id,tag_id,'normal',create_time,update_time FROM public.gfg_tag_map;
INSERT INTO public.gfg_game_tag(game_id,tag_id,role,create_time,update_time)
SELECT g.id,roles.tag_id,roles.role,g.update_time,g.update_time
FROM public.gfg_game g CROSS JOIN LATERAL
    (VALUES (g.primary_tag,'primary'),(g.secondary_tag,'secondary')) roles(tag_id,role)
WHERE roles.tag_id<>0
ON CONFLICT(game_id,tag_id) DO UPDATE SET role=EXCLUDED.role,
    update_time=GREATEST(public.gfg_game_tag.update_time,EXCLUDED.update_time);

-- +goose StatementBegin
DO $validation$
DECLARE counts record;
BEGIN
    SELECT * INTO STRICT counts FROM tag_refactor_counts;
    IF (SELECT count(*) FROM public.gfg_tag) <> counts.leaves
       OR (SELECT count(*) FROM public.gfg_game_tag) <> counts.map_pairs+counts.missing_primary+counts.missing_secondary
       OR (SELECT count(*) FROM public.gfg_game_tag WHERE role='primary') <> counts.primaries
       OR (SELECT count(*) FROM public.gfg_game_tag WHERE role='secondary') <> counts.secondaries
       OR EXISTS (SELECT game_id,tag_id FROM public.gfg_tag_map EXCEPT SELECT game_id,tag_id FROM public.gfg_game_tag)
       OR EXISTS (SELECT id,primary_tag FROM public.gfg_game WHERE primary_tag<>0
                  EXCEPT SELECT game_id,tag_id FROM public.gfg_game_tag WHERE role='primary')
       OR EXISTS (SELECT id,secondary_tag FROM public.gfg_game WHERE secondary_tag<>0
                  EXCEPT SELECT game_id,tag_id FROM public.gfg_game_tag WHERE role='secondary') THEN
        RAISE EXCEPTION 'tag refactor: relation preservation validation failed';
    END IF;
    RAISE NOTICE 'tag refactor: leaves=%, legacy pairs=%, primary=%, secondary=%, synthesized primary=%, synthesized secondary=%',
        counts.leaves,counts.map_pairs,counts.primaries,counts.secondaries,counts.missing_primary,counts.missing_secondary;
END;
$validation$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.gfg_refresh_current_game_daily(
    target_game_id bigint,
    source_kind text
) RETURNS void
LANGUAGE plpgsql
AS $function$
BEGIN
    IF source_kind NOT IN ('bootstrap', 'observed', 'carried_forward') THEN
        RAISE EXCEPTION 'invalid Game Daily materialization source: %', source_kind;
    END IF;

    WITH current_game AS (
        SELECT g.*, p.id AS tracking_period_id
        FROM public.gfg_game g
        JOIN public.gfg_game_tracking_periods p
          ON p.game_id = g.id
         AND p.appid = g.appid
         AND p.tracking_basis = 'explicit'
         AND p.tracked_until IS NULL
        WHERE g.id = target_game_id
    ), languages AS (
        SELECT l.game_id,
               array_agg(l.language_code ORDER BY l.sort_order, l.id)
                   FILTER (WHERE l.language_code IS NOT NULL) AS language_codes,
               array_agg(l.steam_name ORDER BY l.sort_order, l.id)
                   FILTER (WHERE l.language_code IS NULL) AS unknown_language_names,
               array_agg(l.language_code ORDER BY l.sort_order, l.id)
                   FILTER (WHERE l.language_code IS NOT NULL AND l.full_audio_supported) AS full_audio_language_codes,
               max(l.observed_at) AS languages_observed_at
        FROM public.gfg_game_languages l
        WHERE l.game_id = target_game_id
        GROUP BY l.game_id
    ), tags AS (
        SELECT m.game_id, array_agg(m.tag_id ORDER BY m.tag_id) AS tag_ids,
               max(m.tag_id) FILTER (WHERE m.role='primary') AS primary_tag_id,
               max(m.tag_id) FILTER (WHERE m.role='secondary') AS secondary_tag_id
        FROM public.gfg_game_tag m
        WHERE m.game_id = target_game_id
        GROUP BY m.game_id
    ), snapshot AS (
        SELECT g.id AS game_id,
               (transaction_timestamp() AT TIME ZONE 'UTC')::date AS fact_date,
               g.tracking_period_id,
               g.appid,
               transaction_timestamp() AS snapshot_at,
               g.name::text AS name,
               g.name_en::text AS name_en,
               g.view_count,
               NULLIF(d.type, '') AS game_type,
               d.is_free,
               CASE WHEN d.game_id IS NULL THEN NULL ELSE (d.platforms ->> 'windows')::boolean END AS windows,
               CASE WHEN d.game_id IS NULL THEN NULL ELSE (d.platforms ->> 'mac')::boolean END AS mac,
               CASE WHEN d.game_id IS NULL THEN NULL ELSE (d.platforms ->> 'linux')::boolean END AS linux,
               r.availability AS release_availability,
               r.precision AS release_precision,
               r.exact_date AS release_exact_date,
               r.release_year,
               r.release_month,
               r.release_quarter,
               r.window_start AS release_window_start,
               r.window_end AS release_window_end,
               r.observed_at AS release_observed_at,
               f.precision AS first_available_precision,
               f.exact_date AS first_available_exact_date,
               f.release_year AS first_available_year,
               f.release_month AS first_available_month,
               f.release_quarter AS first_available_quarter,
               f.window_start AS first_available_window_start,
               f.window_end AS first_available_window_end,
               f.source AS first_available_source,
               f.inferred AS first_available_inferred,
               l.language_codes,
               l.unknown_language_names,
               l.full_audio_language_codes,
               l.languages_observed_at,
               CASE WHEN jsonb_typeof(g.developers) = 'array'
                    THEN ARRAY(SELECT jsonb_array_elements_text(g.developers))
                    ELSE ARRAY[]::text[] END AS developers,
               CASE WHEN jsonb_typeof(g.publishers) = 'array'
                    THEN ARRAY(SELECT jsonb_array_elements_text(g.publishers))
                    ELSE ARRAY[]::text[] END AS publishers,
               t.primary_tag_id AS primary_tag_id,
               t.secondary_tag_id AS secondary_tag_id,
               COALESCE(t.tag_ids, ARRAY[]::bigint[]) AS tag_ids,
               d.collected_at AS details_observed_at
        FROM current_game g
        LEFT JOIN public.gfg_game_details d ON d.game_id = g.id AND d.appid = g.appid
        LEFT JOIN public.gfg_game_release_state r ON r.game_id = g.id
        LEFT JOIN public.gfg_game_first_available f ON f.game_id = g.id
        LEFT JOIN languages l ON l.game_id = g.id
        LEFT JOIN tags t ON t.game_id = g.id
    )
    INSERT INTO public.gfg_game_daily (
        game_id, fact_date, tracking_period_id, appid, snapshot_at, tracked_at_end,
        name, name_en, view_count, game_type, is_free, windows, mac, linux,
        release_availability, release_precision, release_exact_date, release_year,
        release_month, release_quarter, release_window_start, release_window_end,
        release_observed_at, first_available_precision, first_available_exact_date,
        first_available_year, first_available_month, first_available_quarter,
        first_available_window_start, first_available_window_end,
        first_available_source, first_available_inferred, language_codes,
        unknown_language_names, full_audio_language_codes, languages_observed_at,
        developers, publishers, primary_tag_id, secondary_tag_id, tag_ids,
        details_observed_at, materialization_source, projection_version,
        finalized_at, created_at, updated_at
    )
    SELECT game_id, fact_date, tracking_period_id, appid, snapshot_at, true,
           name, name_en, view_count, game_type, is_free, windows, mac, linux,
           release_availability, release_precision, release_exact_date, release_year,
           release_month, release_quarter, release_window_start, release_window_end,
           release_observed_at, first_available_precision, first_available_exact_date,
           first_available_year, first_available_month, first_available_quarter,
           first_available_window_start, first_available_window_end,
           first_available_source, first_available_inferred, language_codes,
           unknown_language_names, full_audio_language_codes, languages_observed_at,
           developers, publishers, primary_tag_id, secondary_tag_id, tag_ids,
           details_observed_at, source_kind, 2,
           NULL, transaction_timestamp(), transaction_timestamp()
    FROM snapshot
    ON CONFLICT (game_id, fact_date) DO UPDATE
    SET tracking_period_id = EXCLUDED.tracking_period_id,
        appid = EXCLUDED.appid,
        snapshot_at = EXCLUDED.snapshot_at,
        tracked_at_end = EXCLUDED.tracked_at_end,
        name = EXCLUDED.name,
        name_en = EXCLUDED.name_en,
        view_count = EXCLUDED.view_count,
        game_type = EXCLUDED.game_type,
        is_free = EXCLUDED.is_free,
        windows = EXCLUDED.windows,
        mac = EXCLUDED.mac,
        linux = EXCLUDED.linux,
        release_availability = EXCLUDED.release_availability,
        release_precision = EXCLUDED.release_precision,
        release_exact_date = EXCLUDED.release_exact_date,
        release_year = EXCLUDED.release_year,
        release_month = EXCLUDED.release_month,
        release_quarter = EXCLUDED.release_quarter,
        release_window_start = EXCLUDED.release_window_start,
        release_window_end = EXCLUDED.release_window_end,
        release_observed_at = EXCLUDED.release_observed_at,
        first_available_precision = EXCLUDED.first_available_precision,
        first_available_exact_date = EXCLUDED.first_available_exact_date,
        first_available_year = EXCLUDED.first_available_year,
        first_available_month = EXCLUDED.first_available_month,
        first_available_quarter = EXCLUDED.first_available_quarter,
        first_available_window_start = EXCLUDED.first_available_window_start,
        first_available_window_end = EXCLUDED.first_available_window_end,
        first_available_source = EXCLUDED.first_available_source,
        first_available_inferred = EXCLUDED.first_available_inferred,
        language_codes = EXCLUDED.language_codes,
        unknown_language_names = EXCLUDED.unknown_language_names,
        full_audio_language_codes = EXCLUDED.full_audio_language_codes,
        languages_observed_at = EXCLUDED.languages_observed_at,
        developers = EXCLUDED.developers,
        publishers = EXCLUDED.publishers,
        primary_tag_id = EXCLUDED.primary_tag_id,
        secondary_tag_id = EXCLUDED.secondary_tag_id,
        tag_ids = EXCLUDED.tag_ids,
        details_observed_at = EXCLUDED.details_observed_at,
        materialization_source = EXCLUDED.materialization_source,
        projection_version = EXCLUDED.projection_version,
        updated_at = transaction_timestamp()
    WHERE public.gfg_game_daily.finalized_at IS NULL;
END;
$function$;
-- +goose StatementEnd

-- Carry-forward retains the source fact version, including pre-cutover facts.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.gfg_project_state_fact_day(target_date date)
RETURNS TABLE(game_rows bigint, price_rows bigint)
LANGUAGE plpgsql
AS $function$
DECLARE
    day_start timestamp with time zone := target_date::timestamp AT TIME ZONE 'UTC';
    day_end timestamp with time zone := (target_date + 1)::timestamp AT TIME ZONE 'UTC';
    eligible_count bigint;
BEGIN
    WITH eligible AS (
        SELECT period.*,
               LEAST(day_end, COALESCE(period.tracked_until, day_end)) AS terminal_at,
               period.tracked_until IS NULL OR period.tracked_until > day_end AS tracked_at_end
        FROM public.gfg_game_tracking_periods period
        WHERE period.tracked_from < day_end
          AND (period.tracked_until IS NULL OR period.tracked_until > day_start)
          AND period.tracking_basis = 'explicit'
    ), materialized AS (
        SELECT eligible.*,
               prior.game_id AS prior_game_id
        FROM eligible
        LEFT JOIN LATERAL (
            SELECT daily.*
            FROM public.gfg_game_daily daily
            WHERE daily.tracking_period_id = eligible.id
              AND daily.fact_date <= target_date
            ORDER BY daily.fact_date DESC
            LIMIT 1
        ) prior ON true
    )
    SELECT count(*), count(prior_game_id) INTO eligible_count, game_rows
    FROM materialized;

    IF game_rows <> eligible_count THEN
        RAISE EXCEPTION 'Game state source unavailable for %, expected %, materialized %', target_date, eligible_count, game_rows;
    END IF;

    WITH eligible AS (
        SELECT period.*,
               LEAST(day_end, COALESCE(period.tracked_until, day_end)) AS terminal_at,
               period.tracked_until IS NULL OR period.tracked_until > day_end AS tracked_at_end
        FROM public.gfg_game_tracking_periods period
        WHERE period.tracked_from < day_end
          AND (period.tracked_until IS NULL OR period.tracked_until > day_start)
          AND period.tracking_basis = 'explicit'
    ), latest AS (
        SELECT eligible.id AS target_tracking_period_id,
               eligible.appid AS target_appid,
               eligible.terminal_at AS terminal_snapshot_at,
               eligible.tracked_at_end AS terminal_tracked_at_end,
               prior.*
        FROM eligible
        JOIN LATERAL (
            SELECT daily.*
            FROM public.gfg_game_daily daily
            WHERE daily.tracking_period_id = eligible.id
              AND daily.fact_date <= target_date
            ORDER BY daily.fact_date DESC
            LIMIT 1
        ) prior ON true
    )
    INSERT INTO public.gfg_game_daily (
        game_id, fact_date, tracking_period_id, appid, snapshot_at, tracked_at_end,
        name, name_en, view_count, game_type, is_free, windows, mac, linux,
        release_availability, release_precision, release_exact_date, release_year,
        release_month, release_quarter, release_window_start, release_window_end,
        release_observed_at, first_available_precision, first_available_exact_date,
        first_available_year, first_available_month, first_available_quarter,
        first_available_window_start, first_available_window_end,
        first_available_source, first_available_inferred, language_codes,
        unknown_language_names, full_audio_language_codes, languages_observed_at,
        developers, publishers, primary_tag_id, secondary_tag_id, tag_ids,
        details_observed_at, materialization_source, projection_version,
        finalized_at, created_at, updated_at
    )
    SELECT latest.game_id, target_date, latest.target_tracking_period_id, latest.target_appid,
           latest.terminal_snapshot_at, latest.terminal_tracked_at_end,
           latest.name, latest.name_en, latest.view_count, latest.game_type,
           latest.is_free, latest.windows, latest.mac, latest.linux,
           latest.release_availability, latest.release_precision,
           latest.release_exact_date, latest.release_year, latest.release_month,
           latest.release_quarter, latest.release_window_start,
           latest.release_window_end, latest.release_observed_at,
           latest.first_available_precision, latest.first_available_exact_date,
           latest.first_available_year, latest.first_available_month,
           latest.first_available_quarter, latest.first_available_window_start,
           latest.first_available_window_end, latest.first_available_source,
           latest.first_available_inferred, latest.language_codes,
           latest.unknown_language_names, latest.full_audio_language_codes,
           latest.languages_observed_at, latest.developers, latest.publishers,
           latest.primary_tag_id, latest.secondary_tag_id, latest.tag_ids,
           latest.details_observed_at,
           CASE WHEN latest.fact_date = target_date THEN latest.materialization_source
                ELSE 'carried_forward' END,
           latest.projection_version, transaction_timestamp(), latest.created_at, transaction_timestamp()
    FROM latest
    ON CONFLICT (game_id, fact_date) DO UPDATE
    SET tracking_period_id = EXCLUDED.tracking_period_id,
        appid = EXCLUDED.appid,
        snapshot_at = EXCLUDED.snapshot_at,
        tracked_at_end = EXCLUDED.tracked_at_end,
        name = EXCLUDED.name,
        name_en = EXCLUDED.name_en,
        view_count = EXCLUDED.view_count,
        game_type = EXCLUDED.game_type,
        is_free = EXCLUDED.is_free,
        windows = EXCLUDED.windows,
        mac = EXCLUDED.mac,
        linux = EXCLUDED.linux,
        release_availability = EXCLUDED.release_availability,
        release_precision = EXCLUDED.release_precision,
        release_exact_date = EXCLUDED.release_exact_date,
        release_year = EXCLUDED.release_year,
        release_month = EXCLUDED.release_month,
        release_quarter = EXCLUDED.release_quarter,
        release_window_start = EXCLUDED.release_window_start,
        release_window_end = EXCLUDED.release_window_end,
        release_observed_at = EXCLUDED.release_observed_at,
        first_available_precision = EXCLUDED.first_available_precision,
        first_available_exact_date = EXCLUDED.first_available_exact_date,
        first_available_year = EXCLUDED.first_available_year,
        first_available_month = EXCLUDED.first_available_month,
        first_available_quarter = EXCLUDED.first_available_quarter,
        first_available_window_start = EXCLUDED.first_available_window_start,
        first_available_window_end = EXCLUDED.first_available_window_end,
        first_available_source = EXCLUDED.first_available_source,
        first_available_inferred = EXCLUDED.first_available_inferred,
        language_codes = EXCLUDED.language_codes,
        unknown_language_names = EXCLUDED.unknown_language_names,
        full_audio_language_codes = EXCLUDED.full_audio_language_codes,
        languages_observed_at = EXCLUDED.languages_observed_at,
        developers = EXCLUDED.developers,
        publishers = EXCLUDED.publishers,
        primary_tag_id = EXCLUDED.primary_tag_id,
        secondary_tag_id = EXCLUDED.secondary_tag_id,
        tag_ids = EXCLUDED.tag_ids,
        details_observed_at = EXCLUDED.details_observed_at,
        materialization_source = EXCLUDED.materialization_source,
        projection_version = EXCLUDED.projection_version,
        finalized_at = transaction_timestamp(),
        updated_at = transaction_timestamp();

    GET DIAGNOSTICS game_rows = ROW_COUNT;

    WITH eligible AS (
        SELECT period.*
        FROM public.gfg_game_tracking_periods period
        WHERE period.tracked_from < day_end
          AND (period.tracked_until IS NULL OR period.tracked_until > day_start)
          AND period.tracking_basis = 'explicit'
    ), candidates AS (
        SELECT period.id AS tracking_period_id,
               period.game_id,
               period.appid,
               region.region,
               COALESCE(today.price_state, prior.price_state, current.price_state, 'unknown') AS price_state,
               COALESCE(today.currency, prior.currency, NULLIF(current.currency, '')) AS currency,
               COALESCE(today.initial_amount, prior.initial_amount, current.initial_amount) AS initial_amount,
               COALESCE(today.final_amount, prior.final_amount, current.final_amount) AS final_amount,
               COALESCE(today.discount_percent, prior.discount_percent, current.discount_percent::integer) AS discount_percent,
               COALESCE(today.observed_at, prior.observed_at, current.collected_at) AS observed_at,
               CASE WHEN today.tracking_period_id IS NOT NULL THEN today.materialization_source
                    WHEN prior.tracking_period_id IS NOT NULL THEN 'carried_forward'
                    ELSE 'bootstrap' END AS source_kind
        FROM eligible period
        CROSS JOIN (VALUES ('CN'::text), ('US'::text), ('HK'::text)) region(region)
        LEFT JOIN public.gfg_game_price_daily today
          ON today.tracking_period_id = period.id
         AND today.region = region.region
         AND today.fact_date = target_date
        LEFT JOIN LATERAL (
            SELECT daily.*
            FROM public.gfg_game_price_daily daily
            WHERE daily.tracking_period_id = period.id
              AND daily.region = region.region
              AND daily.fact_date < target_date
            ORDER BY daily.fact_date DESC
            LIMIT 1
        ) prior ON today.tracking_period_id IS NULL
        LEFT JOIN public.gfg_game_prices current
          ON current.game_id = period.game_id
         AND current.appid = period.appid
         AND current.region = region.region
         AND today.tracking_period_id IS NULL
         AND prior.tracking_period_id IS NULL
    )
    INSERT INTO public.gfg_game_price_daily (
        tracking_period_id, game_id, appid, region, fact_date, price_state,
        currency, initial_amount, final_amount, discount_percent, observed_at,
        materialization_source, projection_version, finalized_at,
        created_at, updated_at
    )
    SELECT tracking_period_id, game_id, appid, region, target_date, price_state,
           CASE WHEN price_state = 'priced' THEN currency END,
           CASE WHEN price_state = 'priced' THEN initial_amount END,
           CASE WHEN price_state = 'priced' THEN final_amount END,
           CASE WHEN price_state = 'priced' THEN discount_percent END,
           observed_at, source_kind, 1, transaction_timestamp(),
           transaction_timestamp(), transaction_timestamp()
    FROM candidates
    ON CONFLICT (tracking_period_id, region, fact_date) DO UPDATE
    SET game_id = EXCLUDED.game_id,
        appid = EXCLUDED.appid,
        price_state = EXCLUDED.price_state,
        currency = EXCLUDED.currency,
        initial_amount = EXCLUDED.initial_amount,
        final_amount = EXCLUDED.final_amount,
        discount_percent = EXCLUDED.discount_percent,
        observed_at = EXCLUDED.observed_at,
        materialization_source = EXCLUDED.materialization_source,
        projection_version = EXCLUDED.projection_version,
        finalized_at = transaction_timestamp(),
        updated_at = transaction_timestamp();

    GET DIAGNOSTICS price_rows = ROW_COUNT;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

-- Current recommendations were generated from different feature semantics.
DELETE FROM public.gfg_game_recommendations;
DROP INDEX public.idx_gfg_tag_prefix;
ALTER TABLE public.gfg_tag DROP COLUMN prefix;
DROP TABLE public.gfg_tag_map;
DROP INDEX public.idx_gfg_game_primary_tag;
DROP INDEX public.idx_gfg_game_secondary_tag;
ALTER TABLE public.gfg_game DROP COLUMN primary_tag, DROP COLUMN secondary_tag;

-- +goose Down
-- +goose StatementBegin
DO $irreversible$
BEGIN
    RAISE EXCEPTION 'Tag domain migration is irreversible; restore a verified backup or recreate the disposable database';
END;
$irreversible$;
-- +goose StatementEnd
