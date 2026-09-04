-- name: FoundationPing :one
SELECT 1::bigint AS value;

-- name: ListGameTargets :many
SELECT id, appid
FROM gfg_game;

-- name: UpsertDetails :exec
INSERT INTO gfg_game_details (
    game_id, appid, source, type, name, is_free, website, header_url,
    developers, publishers, release_coming_soon, release_date_text,
    platforms, supported_languages, support_info, content_descriptors, ratings,
    collected_at, updated_at
) VALUES (
    sqlc.arg(game_id), sqlc.arg(appid), 'steam', sqlc.arg(type), sqlc.arg(name),
    sqlc.arg(is_free), sqlc.arg(website), sqlc.arg(header_url),
    sqlc.arg(developers)::jsonb, sqlc.arg(publishers)::jsonb,
    sqlc.arg(release_coming_soon), sqlc.arg(release_date_text),
    sqlc.arg(platforms)::jsonb, sqlc.arg(supported_languages),
    sqlc.arg(support_info)::jsonb, sqlc.arg(content_descriptors)::jsonb,
    sqlc.arg(ratings)::jsonb, sqlc.arg(collected_at), now()
)
ON CONFLICT (game_id) DO UPDATE SET
    appid = EXCLUDED.appid,
    source = EXCLUDED.source,
    type = EXCLUDED.type,
    name = EXCLUDED.name,
    is_free = EXCLUDED.is_free,
    website = EXCLUDED.website,
    header_url = EXCLUDED.header_url,
    developers = EXCLUDED.developers,
    publishers = EXCLUDED.publishers,
    release_coming_soon = EXCLUDED.release_coming_soon,
    release_date_text = EXCLUDED.release_date_text,
    platforms = EXCLUDED.platforms,
    supported_languages = EXCLUDED.supported_languages,
    support_info = EXCLUDED.support_info,
    content_descriptors = EXCLUDED.content_descriptors,
    ratings = EXCLUDED.ratings,
    collected_at = EXCLUDED.collected_at,
    updated_at = now();

-- name: UpsertLocalizedDetails :exec
INSERT INTO gfg_game_localized_details (
    game_id, appid, lang, name, short_description, detailed_description,
    about_the_game, collected_at, updated_at
) VALUES (
    sqlc.arg(game_id), sqlc.arg(appid), sqlc.arg(lang), sqlc.arg(name),
    sqlc.arg(short_description), sqlc.arg(detailed_description),
    sqlc.arg(about_the_game), sqlc.arg(collected_at), now()
)
ON CONFLICT (game_id, lang) DO UPDATE SET
    appid = EXCLUDED.appid,
    name = EXCLUDED.name,
    short_description = EXCLUDED.short_description,
    detailed_description = EXCLUDED.detailed_description,
    about_the_game = EXCLUDED.about_the_game,
    collected_at = EXCLUDED.collected_at,
    updated_at = now();

-- name: UpsertPrice :exec
INSERT INTO gfg_game_prices (
    game_id, appid, region, is_free, currency, initial_amount, final_amount,
    discount_percent, initial_formatted, final_formatted, price_state,
    collected_at, updated_at
) VALUES (
    sqlc.arg(game_id), sqlc.arg(appid), sqlc.arg(region), sqlc.arg(is_free),
    sqlc.arg(currency), sqlc.arg(initial_amount), sqlc.arg(final_amount),
    sqlc.arg(discount_percent), sqlc.arg(initial_formatted),
    sqlc.arg(final_formatted), sqlc.arg(price_state), sqlc.arg(collected_at), now()
)
ON CONFLICT (game_id, region) DO UPDATE SET
    appid = EXCLUDED.appid,
    is_free = EXCLUDED.is_free,
    currency = EXCLUDED.currency,
    initial_amount = EXCLUDED.initial_amount,
    final_amount = EXCLUDED.final_amount,
    discount_percent = EXCLUDED.discount_percent,
    initial_formatted = EXCLUDED.initial_formatted,
    final_formatted = EXCLUDED.final_formatted,
    price_state = EXCLUDED.price_state,
    collected_at = EXCLUDED.collected_at,
    updated_at = now();

-- name: DeleteMediaByGame :exec
DELETE FROM gfg_game_media WHERE game_id = sqlc.arg(game_id);

-- name: UpsertMedia :exec
INSERT INTO gfg_game_media (
    game_id, appid, media_type, media_key, title, url, thumbnail_url, extra,
    sort_order, collected_at, updated_at
) VALUES (
    sqlc.arg(game_id), sqlc.arg(appid), sqlc.arg(media_type), sqlc.arg(media_key),
    sqlc.arg(title), sqlc.arg(url), sqlc.arg(thumbnail_url), sqlc.arg(extra)::jsonb,
    sqlc.arg(sort_order), sqlc.arg(collected_at), now()
)
ON CONFLICT (game_id, media_type, media_key) DO UPDATE SET
    appid = EXCLUDED.appid,
    title = EXCLUDED.title,
    url = EXCLUDED.url,
    thumbnail_url = EXCLUDED.thumbnail_url,
    extra = EXCLUDED.extra,
    sort_order = EXCLUDED.sort_order,
    collected_at = EXCLUDED.collected_at,
    updated_at = now();

-- name: DeleteAssetsByGameSourceLang :exec
DELETE FROM gfg_game_assets
WHERE game_id = sqlc.arg(game_id)
  AND source = sqlc.arg(source)
  AND lang = sqlc.arg(lang);

-- name: ListAssetsByGame :many
SELECT *
FROM gfg_game_assets
WHERE game_id = sqlc.arg(game_id)
ORDER BY asset_family, sort_order, id;

-- name: UpsertAsset :exec
INSERT INTO gfg_game_assets (
    game_id, appid, asset_type, asset_family, source, lang, media_key, title,
    url, thumbnail_url, format, exists, status_code, content_type, content_length,
    extra, sort_order, checked_at, collected_at, updated_at
) VALUES (
    sqlc.arg(game_id), sqlc.arg(appid), sqlc.arg(asset_type), sqlc.arg(asset_family),
    sqlc.arg(source), sqlc.arg(lang), sqlc.arg(media_key), sqlc.arg(title),
    sqlc.arg(url), sqlc.arg(thumbnail_url), sqlc.arg(format), sqlc.narg(exists),
    sqlc.arg(status_code), sqlc.arg(content_type), sqlc.arg(content_length),
    sqlc.arg(extra)::jsonb, sqlc.arg(sort_order), sqlc.narg(checked_at),
    sqlc.arg(collected_at), now()
)
ON CONFLICT (game_id, asset_type, lang, media_key) DO UPDATE SET
    appid = EXCLUDED.appid,
    asset_family = EXCLUDED.asset_family,
    source = EXCLUDED.source,
    title = EXCLUDED.title,
    url = EXCLUDED.url,
    thumbnail_url = EXCLUDED.thumbnail_url,
    format = EXCLUDED.format,
    exists = EXCLUDED.exists,
    status_code = EXCLUDED.status_code,
    content_type = EXCLUDED.content_type,
    content_length = EXCLUDED.content_length,
    extra = EXCLUDED.extra,
    sort_order = EXCLUDED.sort_order,
    checked_at = EXCLUDED.checked_at,
    collected_at = EXCLUDED.collected_at,
    updated_at = now();

-- name: UpsertRequirements :exec
INSERT INTO gfg_game_requirements (
    game_id, appid, pc, mac, linux, collected_at, updated_at
) VALUES (
    sqlc.arg(game_id), sqlc.arg(appid), sqlc.arg(pc)::jsonb,
    sqlc.arg(mac)::jsonb, sqlc.arg(linux)::jsonb, sqlc.arg(collected_at), now()
)
ON CONFLICT (game_id) DO UPDATE SET
    appid = EXCLUDED.appid,
    pc = EXCLUDED.pc,
    mac = EXCLUDED.mac,
    linux = EXCLUDED.linux,
    collected_at = EXCLUDED.collected_at,
    updated_at = now();

-- name: InsertDetailSnapshot :exec
INSERT INTO gfg_game_detail_snapshots (
    game_id, appid, lang, region, source, payload_hash, raw_payload, collected_at
) VALUES (
    sqlc.arg(game_id), sqlc.arg(appid), sqlc.arg(lang), sqlc.arg(region),
    sqlc.arg(source), sqlc.arg(payload_hash), sqlc.arg(raw_payload)::jsonb,
    sqlc.arg(collected_at)
);

-- name: PruneDetailSnapshots :one
SELECT gfg_game_prune_detail_snapshots(
    sqlc.arg(appid), sqlc.arg(lang), sqlc.arg(region), 5
)::integer;

-- name: UpsertNews :exec
INSERT INTO gfg_game_news (
    game_id, appid, lang, event_gid, announcement_gid, forum_topic_id,
    headline, raw_body, html, plain_text, summary, url, tags, vote_up_count,
    vote_down_count, comment_count, raw_event, published_at, updated_at, collected_at
) VALUES (
    sqlc.arg(game_id), sqlc.arg(appid), sqlc.arg(lang), sqlc.arg(event_gid),
    sqlc.arg(announcement_gid), sqlc.arg(forum_topic_id), sqlc.arg(headline),
    sqlc.arg(raw_body), sqlc.arg(html), sqlc.arg(plain_text), sqlc.arg(summary),
    sqlc.arg(url), sqlc.arg(tags)::jsonb, sqlc.arg(vote_up_count),
    sqlc.arg(vote_down_count), sqlc.arg(comment_count), sqlc.arg(raw_event)::jsonb,
    sqlc.narg(published_at), sqlc.narg(updated_at), sqlc.arg(collected_at)
)
ON CONFLICT (appid, lang, event_gid, announcement_gid) DO UPDATE SET
    game_id = EXCLUDED.game_id,
    forum_topic_id = EXCLUDED.forum_topic_id,
    headline = EXCLUDED.headline,
    raw_body = EXCLUDED.raw_body,
    html = EXCLUDED.html,
    plain_text = EXCLUDED.plain_text,
    summary = EXCLUDED.summary,
    url = EXCLUDED.url,
    tags = EXCLUDED.tags,
    vote_up_count = EXCLUDED.vote_up_count,
    vote_down_count = EXCLUDED.vote_down_count,
    comment_count = EXCLUDED.comment_count,
    raw_event = EXCLUDED.raw_event,
    published_at = EXCLUDED.published_at,
    updated_at = EXCLUDED.updated_at,
    collected_at = EXCLUDED.collected_at;

-- name: InsertPlayerCount :exec
INSERT INTO gfg_game_player_counts (
    run_id, game_id, appid, count, status, upstream_status_code,
    error_kind, error_message, collected_at
) VALUES (
    sqlc.arg(run_id), sqlc.arg(game_id), sqlc.arg(appid), sqlc.arg(count),
    sqlc.arg(status), sqlc.arg(upstream_status_code), sqlc.arg(error_kind),
    sqlc.arg(error_message), sqlc.arg(collected_at)
);
