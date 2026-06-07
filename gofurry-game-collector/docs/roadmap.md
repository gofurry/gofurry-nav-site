# GoFurry Game Collector V2 Roadmap

## 当前结论

游戏采集器 v2 不再按“v1/v2 长期双栈兼容”设计，而是按 **v2 主线重构** 设计。

这意味着：

- v2 已成为唯一主线采集路径。
- v1 采集逻辑已从 collector 主流程删除。
- 采集器内部可以采集更多字段、保留更丰富的原始数据和结构化数据，为后端 v2 和前端游戏模块 v2 提供更强可塑性。
- Steam 复杂度优先沉淀到 `github.com/gofurry/steam-go`，而不是继续留在采集器里。

`dry_run`、任务级 enabled 开关和 v1 fallback 只服务于 alpha/beta 开发期，稳定版不再保留在主采集流程中。

## 架构边界

### `steam-go` 负责

- Steam 官方 API / Store Web JSON 请求。
- Traffic class、限流、重试、代理、cooldown、block detection。
- Store appdetails、Store events、player count 等 Steam 数据模型。
- typed model + raw payload fallback。
- BBCode / HTML 清洗、纯文本摘要、Steam 内容格式兼容。
- Steam 上游字段变化的兼容处理。

### `gofurry-game-collector` 负责

- 采集任务调度。
- 游戏 appid 列表和业务优先级。
- 将 `steam-go` 返回模型映射为 GoFurry 游戏域模型。
- 写入 PostgreSQL / Redis。
- 采集报告、失败记录、数据完整性检查。
- 为 backend v2 提供更丰富、可演进的数据契约。

### 不再推荐的方向

- 不在 collector 内继续扩展手写 Steam HTTP client。
- 不在 collector 内继续堆 BBCode/HTML 清洗逻辑。
- 不长期保留 v1/v2 两套采集实现。
- 不为了迁移安全牺牲最终架构简洁度。

## V2 数据目标

v2 不只复制当前字段，还应该为后端预留更多能力。

### 游戏详情

采集并保存：

- 基础信息：名称、简介、发行日期、开发商、发行商、官网、年龄限制。
- 多语言信息：中文、英文，以及后续可扩展语言。
- 多地区价格：CN、US、HK 及后续可扩展地区。
- 媒体资源：header、capsule、screenshots、movies、library hero 等可扩展资源。
- 平台与系统需求：Windows / macOS / Linux、最低配置、推荐配置。
- 内容描述：content descriptors、ratings、support info。
- 原始 Store appdetails raw snapshot，用于后端未来补字段。

### 游戏新闻

采集并保存：

- event gid / announcement gid。
- 标题、正文 HTML、纯文本摘要、原文 raw body。
- 语言、发布时间、更新时间、Steam 原始 URL。
- tags、vote count、comment count、forum topic id。
- event raw payload，便于以后扩展新闻类型和专题页。

### 在线人数

采集并保存：

- 当前在线人数。
- 采集时间。
- 上游响应状态。
- 失败原因。
- 后续可增加峰值、趋势和异常检测。

### 采集运行记录

建议新增采集运行记录能力：

- run id。
- 任务类型：details / news / players。
- appid / game id。
- 开始时间、结束时间、耗时。
- 成功 / 失败 / 跳过。
- 上游状态码、错误类型、重试次数。
- 使用的 traffic bucket。

这会让后端管理页、运维排查和未来自动修复更容易。

## 推荐目标结构

```txt
collector/game/v2/
  steamclient/        # steam-go adapter，只做 collector 侧配置和 observer glue
  domain/             # GoFurry 游戏域模型，不直接等于 Steam model 或 DB model
  tasks/
    details/          # 游戏详情主采集
    news/             # 新闻主采集
    players/          # 在线人数主采集
  mapper/
    steam/            # steam-go model -> domain model
    persistence/      # domain model -> PostgreSQL / Redis model
  repository/         # PostgreSQL / Redis 写入
  runner/             # 统一任务编排、调度、报告
  report/             # 采集结果、错误分类、运行摘要
```

## 配置方向

v2 配置应围绕主线采集，而不是双栈迁移：

```yaml
collector:
  v2:
    steam:
      api_requests_per_5_minutes: 240
      store_requests_per_5_minutes: 180
      burst: 1 # interval 按 requests_per_5_minutes - burst 换算，避免初始令牌突破窗口预算。
      max_workers: 3
      request_timeout_seconds: 10
      retry:
        max_attempts: 2
        base_delay_seconds: 5
        cooldown_on_429_seconds: 300
```

## Version Plan

### v2.0.0-alpha.1 - Steam Client Adapter Foundation

**Status:** Completed
**Scope:** Foundation / Stability / Configuration
**Goal:** 建立基于 `steam-go` 的统一 Steam client adapter，但不把它设计成长期双栈迁移层。

#### Completed

- [x] 接入 `github.com/gofurry/steam-go v1.3.2`。
- [x] 新增 `collector/game/v2/steamclient`。
- [x] 建立 Steam API / Store 双 traffic bucket。
- [x] 支持 proxy、timeout、retry、cooldown。
- [x] 接入 `steam-go` request observer，输出安全元数据。
- [x] 新增 v2 配置结构和 `server.example.yaml` 示例。
- [x] 新增 adapter 单元测试。
- [x] `go test ./...` 通过。

#### Adjustment

- [x] 后续阶段移除“长期 dry-run / 长期 fallback”思路。
- [x] v2 稳定版删除 `enabled=false` 和任务级开关。
- [x] v2 上线前完成全量替换，而不是上线后长期并行。

---

### v2.0.0-alpha.2 - Steam-Go Capability Closure

**Status:** Completed
**Scope:** Library-first / Maintainability
**Goal:** 先审计 collector v2 所需 Steam 能力，把复杂 Steam 解析和边界补进 `steam-go`，避免 collector 继续变厚。

#### Focus

- Store appdetails typed 字段完整度。
- Store events 新闻字段完整度。
- player count official API 使用边界。
- BBCode / HTML 清洗边界。
- raw payload strategy。

#### Tasks

- [x] 梳理当前 collector v1 从 Steam 读取的全部字段。
- [x] 对照 `steam-go v1.3.2`，列出缺失 typed 字段。
- [x] 若 Store appdetails 字段不足，优先在 `steam-go` 补 typed model。
- [x] Store events 字段已满足当前 v2 新闻采集目标，继续保留 raw payload 扩展。
- [x] 当前 BBCode / HTML 清洗能力已满足 v1 替换目标，后续新规则继续补 `steam-go/addons/markup`。
- [x] 在 `steam-go` 中保留 raw payload，collector 只读取必要字段。
- [x] 为 `steam-go` 新能力补 fixture tests。

#### Completed

- [x] 新增 `docs/steam-go-capability-closure.md`，记录 v1 字段读取面和 `steam-go` 覆盖情况。
- [x] 在 `steam-go` 补齐 appdetails ratings helper：
  - `StoreRatings`
  - `StoreRating`
  - `AppDetailsData.DecodeRatings()`
  - `AppDetailsData.SteamGermanyRequiredAge()`
- [x] `steam-go` appdetails fixture / request decode 测试覆盖德国年龄限制。
- [x] 更新 `steam-go` 英文与中文 Web reference。

#### Acceptance Criteria

- [x] collector v2 不需要手写 Steam HTTP 请求。
- [x] collector v2 不需要新增本地 BBCode parser。
- [x] collector v2 mapper 主要消费 `steam-go` typed model。
- [x] Steam 上游字段波动优先由 `steam-go` 承接。

---

### v2.0.0-alpha.3 - Collector V2 Domain Model and Storage Contract

**Status:** Completed
**Scope:** Domain / Storage / Backend readiness
**Goal:** 设计 collector v2 自己的游戏域模型和存储契约，为后端 v2 提供更丰富数据，而不是继续受 v1 表结构限制。

#### Focus

- domain model。
- PostgreSQL / Redis 写入契约。
- raw snapshot。
- 采集运行记录。

#### Tasks

- [x] 新增 `collector/game/v2/domain`。
- [x] 定义 `GameDetails`、`GamePrice`、`GameMedia`、`GameNews`、`PlayerCount`、`RawSnapshot`。
- [x] 新增 `collector/game/v2/report`，定义 `CollectRun`、`TaskResult` 和错误分类。
- [x] 设计 v2 存储迁移方案，明确 PostgreSQL 是事实来源，Redis 是热缓存。
- [x] 为 Store appdetails raw snapshot 设计 PostgreSQL JSONB 保存位置。
- [x] 为 Store events raw payload 设计 PostgreSQL JSONB 保存位置。
- [x] 为采集运行记录设计 PostgreSQL 表。
- [x] 明确后端 v2 可消费的数据字段。
- [x] 写出 schema migration 草案。
- [x] 新增 `docs/game-collector-v2-storage-contract.md`。

#### Acceptance Criteria

- [x] 后端可以基于 v2 数据契约设计新接口。
- [x] collector 不再只围绕旧 `GfgGameRecord` / `GfgGameNews` 做字段搬运。
- [x] raw snapshot 不影响普通接口性能。
- [x] 采集失败有可查询的结构化记录。

---

### v2.0.0-alpha.4 - News Collector V2 Mainline

**Status:** Completed
**Scope:** News / Content quality
**Goal:** 用 v2 主线新闻采集替代 v1 新闻采集，不长期保留双实现。

#### Focus

- `GetAdjacentPartnerEvents`。
- 多语言新闻。
- sanitized HTML。
- plain text summary。
- event raw payload。

#### Tasks

- [x] 新增 `collector/game/v2/tasks/news`。
- [x] 使用 `steam-go` `GetAdjacentPartnerEvents` 采集新闻。
- [x] 使用 `addons/markup.CleanSteamContent` 输出 sanitized HTML。
- [x] 使用 `addons/markup.PlainText` / `Summary` 输出可索引摘要。
- [x] 保存 event gid、announcement gid、tags、votes、comment count、forum topic id。
- [x] 保存 raw event payload。
- [x] 写入 v2 存储契约。
- [x] v2 news enabled 时停用 v1 新闻采集路径。

#### Acceptance Criteria

- [x] 新闻采集不再依赖 v1 `performGameNewsCollect`。
- [x] 新闻内容不再残留常见 Steam BBCode。
- [x] 后端可以同时拿到 HTML、plain text、summary 和 raw 扩展数据。
- [x] 采集失败不污染已有数据。

---

### v2.0.0-alpha.5 - Player Count Collector V2 Mainline

**Status:** Completed
**Scope:** Players / Stability
**Goal:** 用 `steam-go` official API 替代 v1 在线人数采集。

#### Tasks

- [x] 新增 `collector/game/v2/tasks/players`。
- [x] 使用 `steam-go` `GetNumberOfCurrentPlayers`。
- [x] 写入 v2 player count 存储。
- [x] 保存上游状态、失败原因和采集 run id。
- [x] v2 players enabled 时停用 v1 在线人数 HTTP 拼接逻辑。

#### Acceptance Criteria

- [x] 在线人数采集不再依赖 `util.GetByHttpWithParams`。
- [x] 上游失败不会误写正常数据。
- [x] 后端可以查询当前值和最近采集状态。

---

### v2.0.0-alpha.6 - Game Details Collector V2 Mainline

**Status:** Completed
**Scope:** Details / Pricing / Media / Rich data
**Goal:** 重写游戏详情采集，保留更多 Steam 字段，支撑后端和前端 v2。

#### Focus

- Store appdetails。
- 多地区价格。
- 多语言详情。
- 媒体资源。
- raw snapshot。

#### Tasks

- [x] 新增 `collector/game/v2/tasks/details`。
- [x] 使用 `steam-go` Storefront appdetails。
- [x] 采集 CN / HK / US 价格，并设计可扩展地区列表。
- [x] 采集 zh / en 详情，并设计可扩展语言列表。
- [x] 保存 screenshots、movies、header、capsule、background 等媒体数据。
- [x] 保存 detailed description、about the game、pc requirements、support info。
- [x] 保存 raw appdetails snapshot。
- [x] v2 details enabled 时停用 v1 详情采集路径。

#### Acceptance Criteria

- [x] 详情采集不再依赖手写 `gjson` 大段解析。
- [x] 免费、锁区、coming soon、缺价格、缺媒体资源不会导致 panic。
- [x] 后端 v2 能拿到更丰富的结构化详情。

---

### v2.0.0-alpha.7 - Steam Store Rate-Limit Experiment

**Status:** Completed
**Scope:** Stability / Operations / Testing
**Goal:** 用实验数据校准 Steam Store / official API 风控规则，避免凭感觉扩大 v2 并发。

#### Focus

- Store appdetails 请求间隔。
- Store events 请求间隔。
- official API player count 请求间隔。
- proxy / no-proxy 场景差异。
- 429、403、HTML block、timeout 和 5xx 分类。

#### Background

当前 v1 注释中的经验规则来自大量人工尝试，应继续尊重：

- official API 大约 `100 token / 1 minute`。
- Store 接口大约 `[150, 250] token / 5 minutes`。

alpha.6 不扩大默认并发。details v2 仍按每个游戏顺序执行 CN / US / HK 三个 Store appdetails 请求，并继续走 `steamclient` 的 Store traffic bucket、retry 和 cooldown。

#### Tasks

- [x] 新增实验性 benchmark / smoke 工具，不进入生产 schedule：`experimental/steam-rate-limit`。
- [x] 支持指定 appid 列表、请求间隔、burst、worker 数和 proxy。
- [x] 分别测试 appdetails、events、player count。
- [x] 记录状态码、错误类型、block detection、耗时、重试次数和 cooldown。
- [x] 输出 CSV / JSON / 中文 Markdown 实验报告。
- [x] 基于真实实验结果更新 `collector.v2.steam` 推荐默认值。
- [x] 明确生产安全阈值和测试环境阈值的判断口径。

#### Acceptance Criteria

- [x] 能复现实验并得到可比较报告。
- [x] 能解释当前默认限流配置是否保守。
- [x] 能给出 details / news / players 三类任务的阶段性推荐并发和间隔。
- [x] 不会在生产采集流程中自动运行实验。

#### Implementation

- 工具目录：`experimental/steam-rate-limit`。
- 默认偏保守：Store `180 requests / 5 minutes`、official API `240 requests / 5 minutes`、`workers=3`、`burst=1`，interval 按 `budget - burst` 换算。
- 每轮输出：
  - `report.json`
  - `results.csv`
  - `report.zh-CN.md`
- 出现 `429`、`403`、`5xx`、transport error 或 `block_detected=true` 后，对对应 bucket 进入本地 cooldown。
- 真实实验结论：
  - Store appdetails 在 `workers=10`、无本地 interval / cooldown / retry 时，两轮复测均为 `360` 请求中成功 `228`、失败 `132`。
  - Store `429` 大约从 `220-230` 次请求附近开始，继续请求会快速转为 `403` / block-detected。
  - 等待约 `5` 分钟后可复测到相近边界，支持 Store 约 `[150,250] requests / 5 minutes / egress identity` 的保守预算。
  - production 建议 Store appdetails 先按约 `180 requests / 5 minutes`、`burst=1` 设计，details / news 共用 Store bucket 时继续预留余量。
  - 同一类接口共享限流器：details / news 共享 Store bucket，players 共享 official API bucket。
  - official API 短时 `players` 未观察到 `429`，但 developer key 每日预算才是生产主约束，例如 `10,000 requests / day`。

---

### v2.0.0-beta.1 - Unified V2 Runner

**Status:** Completed
**Scope:** Operations / Reliability
**Goal:** 统一 v2 三类任务编排，形成可上线的主线采集器。

#### Tasks

- [x] 新增 `collector/game/v2/runner`。
- [x] 支持按任务类型执行：details / news / players。
- [x] 支持传入全量 appid 列表或显式 appid 列表。
- [x] 输出 run report：成功数、失败数、跳过数、partial 数和按任务汇总。
- [x] 将现有 details / news / players v2-enabled 路径切到 v2 runner。
- [x] v2-enabled task 不再进入 v1 单任务入口。

#### Acceptance Criteria

- [x] v2 runner 是 v2-enabled task 的唯一主采集入口。
- [x] 生产日志能看到每次采集 run 的完整摘要。
- [x] 单个 appid 失败不影响整个批次。
- [x] `go test ./...` 通过。

#### Implementation

- `collector/game/v2/runner` 只做编排，不直接处理 Steam API、mapper 或 DB 写入。
- `Collect()` 会把 details / news 统一交给 runner。
- `CollectCurrentPlayers()` 会把 players 统一交给 runner。
- runner 使用 `collector.v2.steam.max_workers` 控制批量任务并发；Steam 请求速率仍由 shared Store / official API bucket 控制。
- runner 捕获每个 appid/task 的结果，失败只计入 report，不中断其他 appid。

---

### v2.0.0-rc.1 - Backend Contract Preparation

**Status:** Completed
**Scope:** Backend readiness / API design
**Goal:** 在 collector v2 上线前，为 backend v2 明确数据消费契约。

#### Tasks

- [x] 输出 collector v2 数据契约文档。
- [x] 标注哪些字段适合进入公开 API。
- [x] 标注哪些字段仅用于后台、搜索、推荐或调试。
- [x] 准备 backend v2 API 草案。
- [x] 准备前端 games 模块 v2 所需字段清单。

#### Acceptance Criteria

- [x] backend 不需要理解 Steam 原始 payload 才能使用 v2 数据。
- [x] 前端视觉体验改造有稳定数据来源。
- [x] 数据字段命名、语言、价格、媒体结构清晰。

#### Implementation

- 后端消费契约位于 `gofurry-game-backend/docs/game-v2-backend-contract.md`。
- 公开 API 只暴露稳定业务字段，raw payload / raw event / 采集错误 / traffic bucket 等字段仅用于后台、调试、搜索或推荐。
- backend v2 建议新增 `/api/v2/game/*`，保留 `/api/v1/game/*` 直到前端 games v2 完成切换。
- 后端 v2 继续复用 `gfg_game` 作为站内主档案入口，动态详情、新闻、价格、媒体、在线人数改为消费 collector v2 表和 Redis key。
- 在线人数公开展示只读取最近成功结果，上游失败记录只进入后台观测，不覆盖当前在线人数。

---

### v2.0.0 - V2 Mainline Stable

**Status:** Completed
**Scope:** Stable / Cleanup / Maintenance
**Goal:** v2 成为唯一主线，删除 v1 历史包袱。

#### Tasks

- [x] 删除 v1 Steam HTTP 拼接采集逻辑。
- [x] 删除 collector 内部旧 BBCode parser，改用 `steam-go/addons/markup`。
- [x] 删除长期 dry-run / fallback 分支。
- [x] 清理旧配置项。
- [x] 更新 docs / deployment notes。
- [x] 完成上线前编译与单元测试。

#### Acceptance Criteria

- [x] collector 只保留 v2 主线采集路径。
- [x] Steam 复杂度主要位于 `steam-go`。
- [x] collector 代码职责收敛为调度、映射、写入和报告。
- [x] 后端具备基于 v2 数据继续演进的空间。

#### Implementation

- `collector/game/service/gameService.go` 只保留 v2 runner 主线：details / news / players。
- 删除旧的 v1 Store appdetails、official players、official/store news 手写采集路径。
- 删除旧 MongoDB intro DAO、MongoDB service、MongoDB 配置结构和示例配置。
- 删除 collector 内部 `ParseBBCode`，新闻内容清洗统一由 `steam-go/addons/markup` 负责。
- 删除 v2 `enabled`、`dry_run`、任务级 enabled 开关和旧秒级限流配置，稳定版默认主线即 v2。
- 采集 run summary / task result 会写入 `gfg_game_v2_collect_runs` 和 `gfg_game_v2_collect_task_results`，并刷新 `game:v2:collect:last:*` Redis key。
- `gfg_game_v2_player_counts`、`gfg_game_v2_collect_runs`、`gfg_game_v2_collect_task_results` 已支持可配置保留策略。
- CLI 支持 `go run . collect` / `go run . players` / `go run . all` 进行一次性验证采集。
- `go mod tidy` 已移除 MongoDB 等不再需要的依赖。
- `go test ./...` 通过。

## 上线策略

由于 v2 会在上线前完整写完，本 roadmap 不再设计长期灰度。

推荐上线方式：

1. 本地完成 v2 全链路。
2. 测试环境用真实 appid 列表跑全量采集。
3. 检查 PostgreSQL / Redis 数据完整性。
4. 后端接入 v2 数据契约。
5. 前端 games v2 使用新接口或新字段。
6. 生产发布 v2。
7. 发布后观察 v2 采集 run summary、PostgreSQL / Redis 数据完整性和 Steam cooldown 触发情况。

## 下一步建议

下一步建议进入后端游戏模块 v2 实现。

优先顺序：

1. 在 `gofurry-game-backend` 新增 `/api/v2/game/*` model / dao / service / controller。
2. 后端优先消费 collector v2 PostgreSQL 表，Redis 作为热缓存。
3. 前端 games v2 切到新接口。
4. 生产观察稳定后，再清理后端 v1 动态数据消费路径。
