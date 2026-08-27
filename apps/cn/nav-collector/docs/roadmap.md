# GoFurry Nav Collector Roadmap

## V3-P0.1.1 — completed foundation

- PostgreSQL owns durable schedules, jobs, run attempts, task results, collector instances, leases, cancellation, and recovery.
- Fixed anchored schedules cover ping, HTTP, DNS, RDAP, robots, security.txt, llms.txt, page assets, port checks, and WAF canary.
- Admin supports all/site/target scopes and multi-protocol fan-out without a Nav Backend proxy.
- Observation rows carry optional Job/Run/collector-instance lineage.
- Redis stores realtime progress only; existing public/latest caches retain their meanings.

Legacy interval values are bootstrap-only. `run_on_start`, legacy Redis lease, and legacy Redis run-state settings are accepted only for configuration compatibility and do not control runtime scheduling.

## Deferred to P0.2

Destructive pruning of `gfn_collector_observation` is frozen. P0.2 must define fact retention, aggregation, and long-term analytics before any automatic deletion is reintroduced. Legacy ping/HTTP/DNS log-table retention remains operational and is not an observation-fact policy.
