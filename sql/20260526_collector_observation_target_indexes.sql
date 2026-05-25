-- GoFurry Nav Collector v0.6.5 observation 历史派生查询索引
--
-- 用途：
--   trend / change event 会按 site_id + target + protocol + observed_at 查询
--   gfn_collector_observation。该索引用于降低后端 v2 正式消费前的历史查询成本。
--
-- 执行注意：
--   1. CREATE INDEX CONCURRENTLY 不能放在事务块里执行。
--   2. 建议低峰期手动执行，并先在测试库执行 EXPLAIN ANALYZE 确认查询计划。
--   3. 本脚本可重复执行。

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gfn_collector_observation_site_target_protocol_time_id
ON gfn_collector_observation (site_id, target, protocol, observed_at DESC, id DESC);
