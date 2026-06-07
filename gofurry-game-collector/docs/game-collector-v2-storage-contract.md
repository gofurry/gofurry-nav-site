# Game Collector V2 Storage Contract

## 结论

游戏采集器 v2 使用全新的 v2 存储契约，不再复用 v1 的 `gfg_game_record`、`gfg_game_news`、`gfg_game_player_count` 作为主线写入目标。

当前存储边界：

- PostgreSQL：结构化详情、新闻、在线人数、采集运行记录、raw snapshot。
- Redis：前端高频读取的当前详情、当前新闻列表、当前在线人数和最近采集状态。
- MongoDB：不进入 v2 设计。

v2 domain model 不等于数据库 model，也不等于 `steam-go` model。后续分层为：

- `steam-go model -> mapper -> collector v2 domain`
- `collector v2 domain -> repository -> PostgreSQL / Redis`

这样可以让 Steam 字段波动留在 `steam-go` 和 mapper 内，后端与前端消费稳定的 v2 契约。

## Domain Model

本阶段新增：

```txt
collector/game/v2/domain/
  common.go
  details.go
  news.go
  players.go
  snapshot.go

collector/game/v2/report/
  error.go
  run.go
```

核心模型：

- `GameDetails`：游戏基础详情。
- `GameLocalizedDetails`：语言相关详情，当前仅 `zh` / `en`，后续可扩展。
- `GamePrice`：地区价格，当前 `CN` / `HK` / `US` 起步。
- `GameMedia`：header、capsule、background、screenshots、movies。
- `SystemRequirements`：Windows / macOS / Linux 配置需求。
- `GameNews`：Store events 新闻，保存 `raw_body`、`html`、`plain_text`、`summary`。
- `PlayerCount`：在线人数采集结果，区分真实 0 和上游失败。
- `RawSnapshot`：PostgreSQL JSONB raw payload。
- `CollectRun` / `TaskResult`：采集运行记录。

## PostgreSQL Contract

SQL 草案位于：

```txt
sql/20260607_game_collector_v2_alpha3.sql
```

### 游戏详情

`gfg_game_v2_details`

保存跨语言的基础详情：

- `game_id`
- `appid`
- `type`
- `name`
- `is_free`
- `website`
- `header_url`
- `developers jsonb`
- `publishers jsonb`
- `release_coming_soon`
- `release_date_text`
- `platforms jsonb`
- `supported_languages`
- `support_info jsonb`
- `content_descriptors jsonb`
- `ratings jsonb`
- `collected_at`
- `updated_at`

`gfg_game_v2_localized_details`

保存语言相关文本：

- `game_id`
- `appid`
- `lang`
- `name`
- `short_description`
- `detailed_description`
- `about_the_game`

主键：`game_id + lang`。

### 地区价格

`gfg_game_v2_prices`

保存每个地区的当前价格：

- `game_id`
- `appid`
- `region`
- `is_free`
- `currency`
- `initial_amount`
- `final_amount`
- `discount_percent`
- `initial_formatted`
- `final_formatted`

主键：`game_id + region`。

### 媒体资源

`gfg_game_v2_media`

统一保存图片和视频资源：

- `media_type`：`header`、`capsule`、`background`、`screenshot`、`movie` 等。
- `media_key`：Steam media id 或稳定 key。
- `url`
- `thumbnail_url`
- `extra jsonb`：DASH、HLS、WebM、MP4 等扩展字段。
- `sort_order`

唯一索引：`game_id + media_type + media_key`。

### 系统需求

`gfg_game_v2_requirements`

保存平台配置需求：

- `pc jsonb`
- `mac jsonb`
- `linux jsonb`

配置文本保留 Steam 原始 HTML-ish 内容，由后端或前端按展示场景清洗。

### Raw Snapshot

`gfg_game_v2_detail_snapshots`

保存 Store appdetails raw payload：

- `game_id`
- `appid`
- `lang`
- `region`
- `source`
- `payload_hash`
- `raw_payload jsonb`
- `collected_at`

默认保留策略：

- 每组 `appid + lang + region` 保留最近 5 次。
- SQL 草案内提供 `gfg_game_v2_prune_detail_snapshots(appid, lang, region, keep_count)`。
- repository 在成功写入新 snapshot 后调用 prune function。

这样主详情表保持轻量，raw payload 可追溯但不会无限增长。

### 新闻

`gfg_game_v2_news`

新闻主线统一使用 Store events，不再把 official news 作为 v2 主线依赖。official news 后续只作为备用研究方向。

保存字段：

- `event_gid`
- `announcement_gid`
- `forum_topic_id`
- `headline`
- `raw_body`
- `html`
- `plain_text`
- `summary`
- `url`
- `tags jsonb`
- `vote_up_count`
- `vote_down_count`
- `comment_count`
- `raw_event jsonb`
- `published_at`
- `updated_at`
- `collected_at`

当前只采 `zh` / `en`，后续扩展语言时继续复用 `lang` 字段。

如果 Store events 返回 URL 为空，由 mapper 使用：

```txt
https://store.steampowered.com/news/app/{appid}/view/{announcement_gid}
```

### 在线人数

`gfg_game_v2_player_counts`

保存在线人数历史和上游状态：

- `count`
- `status`
- `upstream_status_code`
- `error_kind`
- `error_message`
- `collected_at`

v2 不再把上游失败写成正常 `0`。真实 0 和失败 0 必须通过 `status` 区分。

### 采集运行记录

`gfg_game_v2_collect_runs`

保存批次级运行摘要：

- `id`
- `task_type`
- `status`
- `total_count`
- `success_count`
- `failed_count`
- `skipped_count`
- `error_kind`
- `error_message`
- `started_at`
- `ended_at`

`gfg_game_v2_collect_task_results`

保存单个 appid 的采集结果：

- `run_id`
- `task_type`
- `status`
- `game_id`
- `appid`
- `upstream_status_code`
- `traffic_bucket`
- `retry_count`
- `duration_millis`
- `error_kind`
- `error_message`

这会支撑后台排查：

- 哪个 appid 失败。
- 是 Steam 429 / block / decode / storage 问题。
- 重试了几次。
- 哪个 traffic bucket 触发 cooldown。

## Redis Contract

Redis 只作为热数据缓存，不作为事实来源。

建议 key：

| Key | Value | TTL |
| --- | --- | --- |
| `game:v2:details:{game_id}:{lang}` | 当前游戏详情聚合 JSON | 7 天 |
| `game:v2:prices:{game_id}` | 当前地区价格列表 JSON | 7 天 |
| `game:v2:media:{game_id}` | 当前媒体资源 JSON | 7 天 |
| `game:v2:news:{game_id}:{lang}` | 当前新闻列表 JSON | 7 天 |
| `game:v2:players:{game_id}:current` | 当前在线人数 JSON | 3 小时 |
| `game:v2:collect:last:{task_type}` | 最近一次采集摘要 JSON | 7 天 |

Redis 写入失败不应回滚 PostgreSQL 写入，但需要记录在 task result 中。

## Alpha.3 Scope Boundary

本阶段只完成：

- domain model。
- report model。
- PostgreSQL / Redis 存储契约文档。
- SQL 草案。
- roadmap 更新。

本阶段不做：

- 不接真实 DAO。
- 不改 schedule。
- 不删除 v1 表。
- 不删除 v1 MongoDB 相关旧代码。
- 不实现 mapper。
- 不实现 repository。

## 后续阶段

建议 alpha.4 先做 news v2：

- Store events 数据结构最适合验证 domain / repository。
- 新闻可直接体现 `raw_body / html / plain_text / summary` 的价值。
- 比详情采集改动面小，比在线人数更能验证 v2 内容质量。
