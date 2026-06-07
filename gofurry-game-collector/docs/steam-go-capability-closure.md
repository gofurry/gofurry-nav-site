# Steam-Go Capability Closure

## 结论

`v2.0.0-alpha.2` 的目标已经收敛为：collector v2 不再自己维护 Steam HTTP 拼接、`gjson` 深路径解析和 BBCode/HTML 清洗，Steam 上游复杂度优先进入 `github.com/gofurry/steam-go`。

基于当前 v1 采集逻辑审计，`steam-go` 已经具备 collector v2 主线所需的核心能力：

- Store appdetails：可覆盖详情、价格、平台、发行日期、开发商、发行商、媒体资源、系统需求、支持信息、内容描述和 raw snapshot。
- Store events：可覆盖 `store.steampowered.com/events/ajaxgetadjacentpartnerevents` 新闻事件、公告正文、语言、评论、投票、标签、时间字段和 raw payload。
- Player count：可通过官方 `ISteamUserStats/GetNumberOfCurrentPlayers` typed client 获取。
- 内容清洗：可通过 `addons/markup` 输出 sanitized HTML、plain text 和 summary。
- ratings：已在 `steam-go` 新增 typed helper，collector 不再需要读取 `ratings.steam_germany.required_age` 的 raw JSON 路径。

因此后续 alpha.3 起可以进入 collector v2 domain model 和存储契约设计，不需要继续在 collector 内补 Steam 解析工具。

## V1 字段读取面

### 游戏详情

v1 当前从 Store appdetails 读取：

| 字段 | v1 来源 | v2 处理方式 |
| --- | --- | --- |
| 免费状态 | `data.is_free` | `storefront.AppDetailsData.IsFree` |
| 价格 | `data.price_overview` | `storefront.StorePrice` |
| 支持语言 | `data.supported_languages` | `AppDetailsData.SupportedLanguages` |
| 发行日期 | `data.release_date` | `StoreReleaseDate` |
| 平台 | `data.platforms` | `StorePlatforms` |
| 开发商 | `data.developers` | `AppDetailsData.Developers` |
| 发行商 | `data.publishers` | `AppDetailsData.Publishers` |
| Header 图 | `data.header_image` | `AppDetailsData.HeaderImage` |
| 简介 | `data.short_description` | `AppDetailsData.ShortDescription` |
| 德国年龄限制 | `data.ratings.steam_germany.required_age` | `AppDetailsData.SteamGermanyRequiredAge()` |
| 支持信息 | `data.support_info` | `StoreSupportInfo` |
| 官网 | `data.website` | `AppDetailsData.Website` |
| 内容描述 | `data.content_descriptors.notes` | `StoreContentDescriptors.Notes` |
| 截图 | `data.screenshots` | `[]StoreScreenshot` |
| 视频 | `data.movies` | `[]StoreMovie`，包含 WebM、MP4、DASH、HLS |
| 详细描述 | `data.detailed_description` | `AppDetailsData.DetailedDescription` |
| 关于游戏 | `data.about_the_game` | `AppDetailsData.AboutTheGame` |
| PC 配置 | `data.pc_requirements` | `StoreRequirements` |
| raw snapshot | v1 未系统保存 | v2 存储契约新增 raw snapshot |

### 新闻

v1 当前使用两个来源：

| 数据 | v1 来源 | v2 处理方式 |
| --- | --- | --- |
| 英文新闻 | `ISteamNews/GetNewsForApp/v2` | 可继续使用 `api/steamnews`，但主线新闻建议统一走 Store events |
| 中文新闻 | `events/ajaxgetadjacentpartnerevents` | `Web.Storefront.GetAdjacentPartnerEvents` |
| 新闻标题 | `announcement_body.headline` | `PartnerAnnouncementBody.Headline` |
| 新闻正文 | `announcement_body.body` | `PartnerAnnouncementBody.Body` + `markup.CleanSteamContent` |
| 发布时间 | `posttime` / official `date` | `PartnerAnnouncementBody.PostTime` |
| 更新时间 | `updatetime` | `PartnerAnnouncementBody.UpdateTime` |
| URL | official `url` 或 event body `url` | 优先 typed URL，缺失时 mapper 基于 appid/gid 构建 |
| tags / votes / comments | v1 未完整保存 | `PartnerAnnouncementBody.Tags`、`VoteUpCount`、`VoteDownCount`、`CommentCount` |
| raw event | v1 未保存 | `PartnerEvent.Raw`、`PartnerAnnouncementBody.Raw` |

### 在线人数

| 数据 | v1 来源 | v2 处理方式 |
| --- | --- | --- |
| 当前在线人数 | 手写 URL + `gjson response.player_count` | `api/steamuserstats.GetNumberOfCurrentPlayers` |
| 上游失败 | 返回 `0`，容易和真实 0 混淆 | v2 domain model 保存状态和失败原因 |

## Steam-Go 补齐项

本阶段已在 `steam-go` 子模块补充：

- `storefront.StoreRatings`
- `storefront.StoreRating`
- `AppDetailsData.DecodeRatings()`
- `AppDetailsData.SteamGermanyRequiredAge()`
- appdetails fixture / request decode 测试
- 英文与中文 Web reference 文档说明

保留策略：

- `AppDetailsData.Ratings` 仍保持 `json.RawMessage`，避免破坏既有公开 API。
- typed helper 只覆盖通用 rating 字段，rating board 特有字段继续通过 raw payload 承接。

## 后续实现约束

后续 collector v2 mapper 应遵守：

- 不新增 Steam 专用 HTTP 拼接。
- 不新增 collector 本地 BBCode parser。
- 不新增 `gjson` 深路径读取 Steam payload。
- Store appdetails 和 Store events 的 raw payload 可以保存，但 raw 解析优先封装进 `steam-go`。
- 如果发现新的 Steam 字段或 BBCode 规则，先补 `steam-go`，再消费。

## 下一步

进入 `v2.0.0-alpha.3`：

- 定义 collector v2 domain model。
- 设计 PostgreSQL / Redis / MongoDB 的 v2 存储契约。
- 明确 raw snapshot 保存位置。
- 明确采集运行记录结构。
- 为后端 v2 和前端 games v2 准备稳定字段清单。
