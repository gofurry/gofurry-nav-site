# GoFurry Nav Collector Roadmap

`gofurry-nav-collector` 的 v2 数据面已完成，可以进入 `gofurry-nav-backend /api/v2/nav` 阶段。

## v2.0.1

- [ ] 记录并评估 `gfn_collector_observation` 的保留边界。

当前状态：

- `ping` / `http` / `dns` 三个主协议已经接入 v2 observation retention。
- 保留策略按 `site_id + protocol` 分组保留，保留条数分别复用：
  - `collector.ping.log_count`
  - `collector.request.log_count`
  - `collector.dns.log_count`
- `rdap` / `robots` / `security_txt` / `page_assets` / `port_check` / `waf_canary` 这些 light probe 当前会写入 `gfn_collector_observation`，但尚未接入同样的自动保留清理。

结论：

- 这意味着 `gfn_collector_observation` 目前不是完全无限增长，但 light probe 部分长期仍会持续累积。
- 本项先作为已知边界记录到 roadmap，暂时不修改 collector 代码或生产配置。

后续导航 v2 API、展示层和管理侧演进转入对应模块维护。
