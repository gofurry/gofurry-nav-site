# Steam 限流规则校准实验

这个目录用于 `gofurry-game-collector v2.0.0-alpha.7` 的实验性限流校准。它不会进入生产 schedule，也不写数据库，只通过 `steam-go` 请求 Steam 接口并生成实验报告。

## 覆盖接口

- `appdetails`：`Storefront.GetAppDetailsRaw`
- `events`：`Storefront.GetAdjacentPartnerEventsRaw`
- `players`：`SteamUserStats.GetNumberOfCurrentPlayersRaw`

## 默认策略

默认值刻意偏保守：

- Store：`store-interval-ms=2000`
- official API：`api-interval-ms=1000`
- `workers=1`
- `burst=1`
- 出现 `403`、`429`、`5xx`、transport error 或 `block_detected=true` 后，本实验工具会对对应 bucket 进入本地 cooldown。

这与当前经验规则保持一致：

- official API 大约 `100 token / 1 minute`
- Store 大约 `[150, 250] token / 5 minutes`

## 运行示例

```bash
go run . ^
  -tasks appdetails,events,players ^
  -appids 440,570,730 ^
  -regions CN,US,HK ^
  -languages schinese,english ^
  -repeat 2 ^
  -workers 1 ^
  -store-interval-ms 2000 ^
  -api-interval-ms 1000 ^
  -out out
```

使用 appid 文件：

```bash
go run . -appid-file appids.txt -tasks appdetails -repeat 1
```

## 主动摸 official API 上限

`players` 才是 official API；`appdetails` 和 `events` 都是 Store 接口。

如果目标是主动触碰 official API 的限制，可以关闭本地 cooldown 和 SDK retry，并把 API 间隔压低：

```bash
go run . ^
  -tasks players ^
  -appids 440,570,730 ^
  -repeat 200 ^
  -workers 20 ^
  -api-interval-ms 0 ^
  -burst 0 ^
  -retry 0 ^
  -cooldown-on-block-seconds 0 ^
  -timeout-seconds 10 ^
  -print-each ^
  -out out
```

这类实验就是为了撞到上游边界，不要把这组参数用于生产采集。

## 运行时日志

默认每 5 秒会输出一次进度：

```text
progress=18/600 remaining=582 store_cooldown=0s api_cooldown=0s
```

如果需要每个请求都打印，增加 `-print-each`。如果想关闭进度日志，设置 `-progress-seconds 0`。

手动 Ctrl+C 时，工具会尽量写出 partial 报告，便于查看已经完成的样本。

## 输出

每次运行会生成：

- `out/<run-id>/report.json`
- `out/<run-id>/results.csv`
- `out/<run-id>/report.zh-CN.md`

中文报告会包含：

- 实验参数
- 总体成功/失败/风控信号
- 按任务统计
- 按 traffic bucket 统计
- 对生产默认值的建议

## 调参建议

每次只改变一个变量，便于归因：

- 请求间隔：`store-interval-ms` / `api-interval-ms`
- 并发：`workers`
- 突发：`burst`
- 代理：`proxy`
- 样本量：`repeat` / `appid-file`

若观察到 `429`、`403` 或 `block_detected=true`，不要把该轮配置作为生产默认值。
