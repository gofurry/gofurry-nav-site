# GoFurry Game Backend V2 Contract

## 结论

游戏后端 v2 不直接理解 Steam 原始 payload，也不再把 v1 的 `gfg_game_record`、`gfg_game_news`、`gfg_game_player_count` 作为游戏模块动态数据的主消费来源。

v2 后端消费边界：

- PostgreSQL 是事实来源，读取 collector v2 写入的结构化表。
- 当前可执行详情读模型直接从 PostgreSQL 聚合；Redis 用于首页、在线峰值和 collector 保留的兼容缓存键。
- Steam raw snapshot 只用于排查、回放和后续 mapper 改进，不进入公开 API。
- official news 暂不进入 v2 主线，新闻统一消费 Store events。
- MongoDB 不进入游戏模块 v2。

后端 v2 的目标不是把 collector 的所有字段透传给前端，而是提供稳定、可维护、适合前端 games v2 视觉改造的业务响应模型。

## 数据来源

collector v2 已定义的 PostgreSQL 表：

| 表 | 后端用途 |
| --- | --- |
| `gfg_game` | 仍作为站内游戏主档案、权重、标签映射、人工维护资源的入口 |
| `gfg_game_details` | 游戏跨语言基础详情 |
| `gfg_game_localized_details` | zh/en 文案、简介、详情正文 |
| `gfg_game_prices` | CN/HK/US 等地区价格 |
| `gfg_game_media` | header、capsule、background、screenshots、movies |
| `gfg_game_requirements` | PC/macOS/Linux 系统需求 |
| `gfg_game_news` | Store events 新闻 |
| `gfg_game_player_counts` | 在线人数历史和当前状态回源 |
| `gfg_game_v2_collect_runs` | 后台采集批次观测 |
| `gfg_game_v2_collect_task_results` | 后台单游戏采集结果观测 |
| `gfg_game_detail_snapshots` | 调试和回放，不进入公开 API |
| `gfg_game_release_state` | 当前 normalized US/English Steam release state |
| `gfg_game_first_available` | write-once 首次正式可购买/可玩日期或区间 |
| `gfg_game_languages` | normalized language capabilities |

`gfg_game_release_history` 在 P0.1 只保存，不公开 API。`gfg_game.release_date`、`gfg_game_details.release_date_text` 和 `supported_languages` 仅保留兼容/调试价值；业务排序、过滤和新 UI 不解析这些字符串。

建议 Redis key：

| Key | 后端用途 |
| --- | --- |
| `game:v2:details:{game_id}:{lang}` | 游戏详情页聚合缓存 |
| `game:v2:prices:{game_id}` | 地区价格缓存 |
| `game:v2:media:{game_id}` | 媒体资源缓存 |
| `game:v2:news:{game_id}:{lang}` | 游戏新闻列表缓存 |
| `game:v2:players:{game_id}:current` | 当前在线人数缓存 |
| `game:v2:players:{game_id}:peak` | 配置窗口内在线峰值缓存；热数据不过期，随在线人数读模型刷新 |
| `game:v2:collect:last:{task_type}` | 后台最近采集状态 |

详情 API 始终以 PostgreSQL 聚合结果为事实。Collector 只允许用提交后重新读取的完整合并资产集刷新兼容详情/媒体缓存，StoreBrowse 失败不得把 Last Known Good 竖版封面覆盖为空。

## 字段分层

### 公开 API 字段

适合进入前端公开 API：

- 游戏身份：`game_id`、`appid`、`name`、`type`、`is_free`。
- 多语言文案：`lang`、`short_description`、`detailed_description`、`about_the_game`。
- 发行信息：structured `release`、`first_available`、`developers`、`publishers`；旧 `release.coming_soon/date` 暂时兼容。
- 平台信息：`platforms`、structured `languages`、`website`、`support_info`；旧 `supported_languages` 暂时兼容。
- 价格信息：`region`、`currency`、`initial_amount`、`final_amount`、`discount_percent`、`initial_formatted`、`final_formatted`。
- 媒体信息：`header_url`、`capsule_url`、`background_url`、`screenshots`、`movies`。
- 新闻信息：`headline`、`summary`、`plain_text`、`html`、`url`、`published_at`、`tags`。
- 在线人数：`count`、`status`、`collected_at`。
- 站内增强：标签、资源、社群、三方链接、浏览量、评论统计。

### 后台或调试字段

只适合后台、搜索、推荐或调试：

- `raw_event`
- `raw_body`
- `raw_payload`
- `payload_hash`
- `upstream_status_code`
- `traffic_bucket`
- `retry_count`
- `error_kind`
- `error_message`
- `collect_runs`
- `collect_task_results`

这些字段可以进入后台接口，但不应该进入公开详情页响应，避免泄露上游异常细节和扩大响应体。

### 搜索与推荐字段

适合用于搜索或推荐，但不一定直接展示：

- `short_description`
- `plain_text`
- `tags`
- `developers`
- `publishers`
- `platforms`
- `is_free`
- `discount_percent`
- `final_amount`
- `player_count`
- `first_available`

搜索索引建议使用 `plain_text` 和 cleaned text，不使用 `html`。

## 公开响应模型草案

### GameV2ListItem

用于列表、推荐、搜索结果、首页轻量卡片。

```json
{
  "id": "1",
  "appid": "440",
  "name": "Team Fortress 2",
  "summary": "Short localized summary.",
  "header_url": "https://...",
  "capsule_url": "https://...",
  "release_date": "Oct 10, 2007",
  "developers": ["Valve"],
  "publishers": ["Valve"],
  "platforms": {
    "windows": true,
    "mac": true,
    "linux": true
  },
  "price": {
    "region": "CN",
    "currency": "CNY",
    "final_amount": 0,
    "final_formatted": "Free To Play",
    "discount_percent": 0,
    "is_free": true
  },
  "online_count": {
    "count": 12345,
    "peak_count": 23456,
    "status": "success",
    "collected_at": "2026-06-07T10:00:00Z"
  },
  "tags": []
}
```

### GameV2Detail

用于游戏详情页。

```json
{
  "id": "1",
  "appid": "440",
  "lang": "zh",
  "name": "Team Fortress 2",
  "short_description": "",
  "detailed_description": "",
  "about_the_game": "",
  "release": {
    "coming_soon": false,
    "date": "Oct 10, 2007",
    "availability": "available",
    "precision": "day",
    "exact_date": "2007-10-10",
    "year": 2007,
    "month": 10,
    "quarter": null,
    "window_start": "2007-10-10",
    "window_end": "2007-10-10",
    "raw_text": "Oct 10, 2007"
  },
  "first_available": {
    "precision": "day",
    "exact_date": "2007-10-10",
    "year": 2007,
    "month": 10,
    "quarter": null,
    "window_start": "2007-10-10",
    "window_end": "2007-10-10",
    "source": "legacy_manual",
    "inferred": false
  },
  "languages": [],
  "developers": ["Valve"],
  "publishers": ["Valve"],
  "website": "https://...",
  "platforms": {
    "windows": true,
    "mac": true,
    "linux": true
  },
  "supported_languages": "",
  "support_info": {
    "url": "",
    "email": ""
  },
  "prices": [],
  "media": {
    "header_url": "",
    "capsule_url": "",
    "background_url": "",
    "screenshots": [],
    "movies": []
  },
  "requirements": {
    "pc": {
      "minimum": "",
      "recommended": ""
    },
    "mac": {
      "minimum": "",
      "recommended": ""
    },
    "linux": {
      "minimum": "",
      "recommended": ""
    }
  },
  "news": [],
  "online_count": {
    "count": 0,
    "peak_count": 0,
    "status": "unknown",
    "collected_at": ""
  },
  "site": {
    "resources": [],
    "groups": [],
    "links": [],
    "tags": [],
    "view_count": 0
  },
  "updated_at": "2026-06-07T10:00:00Z"
}
```

### GameV2NewsItem

用于详情页新闻、首页更新列表、同步接口。

```json
{
  "id": "123456",
  "game_id": "1",
  "appid": "440",
  "lang": "zh",
  "game_name": "Team Fortress 2",
  "headline": "",
  "summary": "",
  "plain_text": "",
  "html": "",
  "url": "https://store.steampowered.com/news/app/440/view/123456",
  "tags": [],
  "published_at": "2026-06-07T10:00:00Z"
}
```

## API 草案

游戏动态主线已经切到 `/api/v2/game/*`。v1 只保留尚未迁移的运营流程，已稳定切换的公开接口应直接移除 v1 路由，避免长期双栈。

| Method | Path | 用途 |
| --- | --- | --- |
| `GET` | `/api/v2/game/list` | 游戏轻量列表，支持 `lang`、`limit`、`offset`、`sort` |
| `GET` | `/api/v2/game/info` | 游戏详情，参数 `id` 或 `appid`、`lang` |
| `GET` | `/api/v2/game/news` | 单游戏新闻列表，参数 `id` 或 `appid`、`lang`、`limit`、`offset` |
| `GET` | `/api/v2/game/news/latest` | 全站最新游戏新闻 |
| `GET` | `/api/v2/game/panel/main` | 首页或游戏页聚合面板 |
| `GET` | `/api/v2/game/prizes` | 抽奖页展示数据 |
| `POST` | `/api/v2/game/prizes/participation` | 提交抽奖参与申请 |
| `GET` | `/api/v2/game/prizes/participation/activation` | 邮件激活抽奖参与申请 |
| `GET` | `/api/v2/game/collect/status` | 后台采集状态，需要后台鉴权 |

查询参数建议：

- `lang`：默认 `zh`，当前支持 `zh` / `en`。
- `region`：默认 `CN`，可支持 `CN` / `HK` / `US`。
- `id`：站内游戏 ID。
- `appid`：Steam appid。
- `limit`：默认按接口限制，公开接口需要上限。
- `offset` 或 `cursor`：列表分页。

抽奖参与临时 Redis key 统一使用 v2 命名空间：

- 邮箱短期频率限制：`game:v2:prize:email:{prize_id}:{email}`，TTL 3 分钟。
- 待激活参与记录：`game:v2:prize:participation:{prize_id}:{activation_key}`，TTL 10 分钟。

邮件激活链接必须指向 `/api/v2/game/prizes/participation/activation`，不再发送 v1 API URL。

## PostgreSQL 查询策略

详情页建议聚合：

1. 从 `gfg_game` 获取站内基础信息、权重、资源、社群、链接。
2. 从 `gfg_game_details` 获取跨语言基础详情。
3. 从 `gfg_game_localized_details` 按 `game_id + lang` 获取当前语言文案。
4. 从 `gfg_game_prices` 获取地区价格。
5. 从 `gfg_game_media` 获取媒体资源。
6. 从 `gfg_game_requirements` 获取系统需求。
7. 从 `gfg_game_news` 获取最近新闻。
8. 从 `gfg_game_player_counts` 获取最近一条 `status='success'` 的在线人数，并按后端 `game.online_peak_cache_days` 配置窗口计算 `peak_count`。
9. 从站内标签、评论、浏览量表补齐站内增强数据。

列表批量读取 `release_state` 与 `first_available`，不为 languages 增加 N+1。Latest Games 只包含存在 First Available 的游戏并按 `window_end DESC, game_id DESC` 排序；Recently Collected 只按 `gfg_game.create_time DESC, game_id DESC` 排序，不因 `available`、`upcoming` 或 `unknown` 改变收录资格。

高级搜索接受可选 `availability=available|upcoming`。省略时保持兼容，不增加状态过滤；Nav Web 默认明确发送 `available`。`pub_start_time/pub_end_time` 始终表示 First Available `[window_start, window_end]` overlap，`planned_start_time/planned_end_time` 只在 `availability=upcoming` 时表示 Current Release State planned window overlap，两组字段不得互换语义。Upcoming 默认按有效未来 `window_end ASC`、`window_start ASC`、precision 排序，过期 observation 随后，TBA 最后。搜索响应同时提供轻量 `release` 与 `first_available`。

新闻列表建议用 `published_at DESC NULLS LAST, collected_at DESC` 排序。

在线人数公开展示只读取最近成功结果。失败结果可进入后台采集状态，但不能覆盖前端当前在线人数。

## Redis 策略

详情页读取顺序：

1. 尝试读取 `game:v2:details:{game_id}:{lang}`。
2. 命中则返回。
3. 未命中则 PostgreSQL 聚合。
4. 聚合成功后写回 Redis。

缓存 TTL 建议：

- 详情、价格、媒体、新闻：`7d`。
- 当前在线人数：`3h`。
- 采集状态：`7d`。

后台手动刷新、采集器成功写入、或管理端更新站内资源后，可以主动删除对应详情缓存。

## 前端 games v2 字段清单

游戏列表页需要：

- `id`
- `appid`
- `name`
- `summary`
- `header_url`
- `capsule_url`
- `release_date`
- `first_available`
- `developers`
- `publishers`
- `platforms`
- `price`
- `online_count`
- `tags`

游戏详情页需要：

- 列表页全部字段。
- `detailed_description`
- `about_the_game`
- `website`
- `languages`
- compatibility-only `supported_languages`
- `support_info`
- `prices`
- `media.screenshots`
- `media.movies`
- `requirements`
- `news`
- `site.resources`
- `site.groups`
- `site.links`
- `site.view_count`

首页或游戏模块聚合面板需要：

- 在线人数排行。
- 最新新闻。
- 最高折扣。
- 免费游戏。
- 近期更新或近期入库。

## 当前切换状态

Game 前端已调用 `/api/v2/game/*`。V3-P0.1 原地演进 V2 model/dao/service/controller，不新增 `/api/v3`，也不复制版本包。兼容 raw 字段暂时返回用于 rollback/debug，但 Nav Web 主展示只消费 canonical typed fields。

优先替换：

- 游戏详情页动态字段。
- 游戏新闻列表。
- 在线人数。
- 价格和媒体。

暂时继续复用：

- `gfg_game` 站内主档案。
- 标签、评论、浏览量、资源、社群、链接。
- 推荐和搜索模块，直到它们有明确的 v2 索引策略。

## 风险点

- v2 表和 v1 表并存期间，字段来源容易混用，需要在 service 层明确 v2 聚合边界。
- Redis key 名称必须和 collector v2 契约一致，否则会出现缓存未命中但难以发现。
- 新闻 `html` 可以展示，但搜索和摘要应使用 `plain_text` / `summary`。
- 在线人数失败不能按 `0` 展示，必须尊重 `status`。
- raw snapshot 不能进入公开响应，避免响应体过大和泄露调试细节。
- `lang` 当前只支持 `zh` / `en`，后续扩展语言时不要新增硬编码分支。

## V3-P0.1 交付边界

本阶段完成 canonical Release/Languages read model、First Available list/search semantics、structured API 和 Nav Web cutover。Release History 只保存不公开；daily facts、Analytics Metric、通用 Change Intelligence 和 `/api/v3` 留在后续里程碑。
