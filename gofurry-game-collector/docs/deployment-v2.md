# GoFurry Game Collector V2 Deployment Notes

## 配置

v2.0.0 stable 后，游戏采集器默认只保留 v2 主线，不再提供 v1 fallback、dry-run 或任务级 enabled 开关。

必需配置：

```yaml
collector:
  proxy: ""
  game:
    game_player_interval: 24
  v2:
    steam:
      api_requests_per_5_minutes: 240
      store_requests_per_5_minutes: 180
      burst: 1
      max_workers: 3
      request_timeout_seconds: 10
      retry:
        max_attempts: 2
        base_delay_seconds: 5
        cooldown_on_429_seconds: 300
    retention:
      player_counts_days: 90
      collect_runs_days: 90
      collect_task_results_days: 7
```

说明：

- `api_requests_per_5_minutes` 用于 official API 共享限流器。
- `store_requests_per_5_minutes` 用于 Store 共享限流器，details / news 共用。
- interval 由 `budget - burst` 计算，避免初始令牌突破 5 分钟预算。
- `max_workers` 控制 runner 并发，不等于 Steam 请求速率。
- MongoDB 已从 collector v2 主线移除。
- `retention.player_counts_days` 默认建议 90 天，便于未来做在线人数趋势图。
- `retention.collect_runs_days` 默认建议 90 天。
- `retention.collect_task_results_days` 默认建议 7 天，因为单游戏任务结果增长最快。

## 数据库

上线前需要确认已执行：

```txt
sql/20260607_game_collector_v2_alpha3.sql
sql/20260607_game_collector_v2_alpha5.sql
sql/20260607_game_collector_v2_stable_collect_runs.sql
```

v2 不删除旧 v1 表。旧表可以等 backend / frontend 完成 v2 切换并稳定观察后再清理。

## 验证

上线前建议：

1. 运行 `go test ./...`。
2. 用开发配置跑一次 details / news / players 采集。
3. 检查 `gfg_game_v2_details`、`gfg_game_v2_localized_details`、`gfg_game_v2_prices`、`gfg_game_v2_media`、`gfg_game_v2_news`、`gfg_game_v2_player_counts` 是否有数据。
4. 检查 Redis v2 key 是否刷新。
5. 检查日志中的 run summary，确认 failed / partial 数量符合预期。
6. 若出现 429 / 403 / block-detected，观察 cooldown 是否触发，并降低 Store 预算或 worker 数。

一次性运行命令：

```powershell
go run . collect # 只跑 details/news 全量采集
go run . players # 只跑当前在线人数采集
go run . all     # 先跑 players，再跑 details/news
```

观测表验证：

```sql
select * from gfg_game_v2_collect_runs order by started_at desc limit 5;
select task_type, status, count(*) from gfg_game_v2_collect_task_results group by task_type, status;
select run_id, count(*) from gfg_game_v2_player_counts where run_id <> '' group by run_id order by run_id desc limit 5;
```

## 发布后观察

重点观察：

- Store 请求是否持续触发 429 / 403。
- official API 是否接近每日 key 预算。
- raw snapshot 是否按最近 5 次保留。
- Redis 写入失败是否影响 PostgreSQL 主数据。
- 后端 v2 接入前，旧前端接口仍不会直接消费这些 v2 表。
