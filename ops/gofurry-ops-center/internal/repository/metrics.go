package repository

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/model"
	"github.com/jackc/pgx/v5"
)

type avgBucket struct {
	sum   float64
	count int
}

type maxBucket struct {
	value float64
	seen  bool
}

type networkRaw struct {
	bytesSent   int64
	bytesRecv   int64
	packetsSent int64
	packetsRecv int64
	reportedAt  time.Time
}

func (r *Repository) OverviewMetrics(ctx context.Context, since time.Time, bucket time.Duration) (model.OverviewMetrics, error) {
	result := model.OverviewMetrics{
		ServiceStatusCounts: []model.StatusCount{},
		AlertLevelCounts:    []model.StatusCount{},
		TopCPU:              []model.TopResource{},
		TopMemory:           []model.TopResource{},
		TopDisk:             []model.TopResource{},
		CPUTrend:            []model.MetricPoint{},
		MemoryTrend:         []model.MetricPoint{},
		LoadTrend:           []model.MetricPoint{},
		DiskTrend:           []model.MetricPoint{},
		LatencyTrend:        []model.MetricPoint{},
	}
	if err := r.loadOverviewLatestSystem(ctx, &result); err != nil {
		return result, err
	}
	if err := r.loadOverviewLatestDisks(ctx, &result); err != nil {
		return result, err
	}
	if err := r.loadOverviewCounts(ctx, &result); err != nil {
		return result, err
	}
	if err := r.loadOverviewTrends(ctx, since, bucket, &result); err != nil {
		return result, err
	}
	return result, nil
}

func (r *Repository) NodeMetrics(ctx context.Context, nodeID string, since time.Time, bucket time.Duration) (model.NodeMetrics, error) {
	result := model.NodeMetrics{
		Latest: model.NodeLatestMetrics{
			Disks:         []model.DiskSnapshot{},
			Networks:      []model.NetworkSnapshot{},
			Docker:        []model.DockerSnapshot{},
			HTTPChecks:    []model.HTTPCheckSnapshot{},
			ServiceChecks: []model.ServiceCheckSnapshot{},
			Certs:         []model.CertSnapshot{},
		},
		Trends: model.NodeTrendMetrics{
			CPU:            []model.MetricPoint{},
			Memory:         []model.MetricPoint{},
			Load:           []model.MetricPoint{},
			NetworkRx:      []model.MetricPoint{},
			NetworkTx:      []model.MetricPoint{},
			DiskUsage:      []model.MetricSeries{},
			ServiceLatency: []model.MetricSeries{},
		},
	}
	if err := r.loadNodeLatestSystem(ctx, nodeID, &result); err != nil {
		return result, err
	}
	if err := r.loadNodeLatestDisks(ctx, nodeID, &result); err != nil {
		return result, err
	}
	if err := r.loadNodeLatestNetworks(ctx, nodeID, &result); err != nil {
		return result, err
	}
	if err := r.loadNodeLatestDocker(ctx, nodeID, &result); err != nil {
		return result, err
	}
	if err := r.loadNodeLatestHTTP(ctx, nodeID, &result); err != nil {
		return result, err
	}
	if err := r.loadNodeLatestServices(ctx, nodeID, &result); err != nil {
		return result, err
	}
	if err := r.loadNodeLatestCerts(ctx, nodeID, &result); err != nil {
		return result, err
	}
	if err := r.loadNodeTrends(ctx, nodeID, since, bucket, &result); err != nil {
		return result, err
	}
	return result, nil
}

func (r *Repository) loadOverviewLatestSystem(ctx context.Context, result *model.OverviewMetrics) error {
	rows, err := r.pool.Query(ctx, `
SELECT node_id, COALESCE(cpu_usage, 0), COALESCE(memory_usage, 0), COALESCE(load1, 0), reported_at
FROM (
    SELECT node_id, cpu_usage, memory_usage, load1, reported_at,
           row_number() OVER (PARTITION BY node_id ORDER BY reported_at DESC) AS rn
    FROM system_samples
) latest
WHERE rn = 1
`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var cpuSum, memorySum, loadSum float64
	for rows.Next() {
		var nodeID string
		var cpu, memory, load float64
		var reportedAt time.Time
		if err := rows.Scan(&nodeID, &cpu, &memory, &load, &reportedAt); err != nil {
			return err
		}
		result.LatestSystem.NodesReported++
		cpuSum += cpu
		memorySum += memory
		loadSum += load
		updateLatestTime(&result.LatestSystem.ReportedAt, reportedAt)
		updateLatestTime(&result.LastSampleAt, reportedAt)
		result.TopCPU = append(result.TopCPU, model.TopResource{NodeID: nodeID, Value: cpu, Unit: "%", UpdatedAt: reportedAt})
		result.TopMemory = append(result.TopMemory, model.TopResource{NodeID: nodeID, Value: memory, Unit: "%", UpdatedAt: reportedAt})
	}
	if result.LatestSystem.NodesReported > 0 {
		count := float64(result.LatestSystem.NodesReported)
		result.LatestSystem.CPUAvg = cpuSum / count
		result.LatestSystem.MemoryAvg = memorySum / count
		result.LatestSystem.Load1Avg = loadSum / count
	}
	result.TopCPU = limitTopResources(result.TopCPU, 5)
	result.TopMemory = limitTopResources(result.TopMemory, 5)
	return rows.Err()
}

func (r *Repository) loadOverviewLatestDisks(ctx context.Context, result *model.OverviewMetrics) error {
	rows, err := r.pool.Query(ctx, `
SELECT node_id, mount, COALESCE(usage, 0), reported_at
FROM (
    SELECT node_id, mount, usage, reported_at,
           row_number() OVER (PARTITION BY node_id, mount ORDER BY reported_at DESC) AS rn
    FROM disk_samples
) latest
WHERE rn = 1
`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var nodeID, mount string
		var usage float64
		var reportedAt time.Time
		if err := rows.Scan(&nodeID, &mount, &usage, &reportedAt); err != nil {
			return err
		}
		item := model.TopResource{NodeID: nodeID, Name: mount, Value: usage, Unit: "%", UpdatedAt: reportedAt}
		result.TopDisk = append(result.TopDisk, item)
		if result.HighestDisk == nil || usage > result.HighestDisk.Value {
			next := item
			result.HighestDisk = &next
		}
		updateLatestTime(&result.LastSampleAt, reportedAt)
	}
	result.TopDisk = limitTopResources(result.TopDisk, 5)
	return rows.Err()
}

func (r *Repository) loadOverviewCounts(ctx context.Context, result *model.OverviewMetrics) error {
	serviceRows, err := r.pool.Query(ctx, `SELECT status, count(*) FROM service_status GROUP BY status ORDER BY status`)
	if err != nil {
		return err
	}
	defer serviceRows.Close()
	for serviceRows.Next() {
		var item model.StatusCount
		if err := serviceRows.Scan(&item.Name, &item.Count); err != nil {
			return err
		}
		result.ServiceStatusCounts = append(result.ServiceStatusCounts, item)
	}
	if err := serviceRows.Err(); err != nil {
		return err
	}
	alertRows, err := r.pool.Query(ctx, `SELECT level, count(*) FROM alert_states WHERE status = 'active' GROUP BY level ORDER BY level`)
	if err != nil {
		return err
	}
	defer alertRows.Close()
	for alertRows.Next() {
		var item model.StatusCount
		if err := alertRows.Scan(&item.Name, &item.Count); err != nil {
			return err
		}
		result.AlertLevelCounts = append(result.AlertLevelCounts, item)
	}
	return alertRows.Err()
}

func (r *Repository) loadOverviewTrends(ctx context.Context, since time.Time, bucket time.Duration, result *model.OverviewMetrics) error {
	cpuBuckets := map[int64]*avgBucket{}
	memoryBuckets := map[int64]*avgBucket{}
	loadBuckets := map[int64]*avgBucket{}
	rows, err := r.pool.Query(ctx, `
SELECT reported_at, COALESCE(cpu_usage, 0), COALESCE(memory_usage, 0), COALESCE(load1, 0)
FROM system_samples
WHERE reported_at >= $1
ORDER BY reported_at
`, since)
	if err != nil {
		return err
	}
	for rows.Next() {
		var reportedAt time.Time
		var cpu, memory, load float64
		if err := rows.Scan(&reportedAt, &cpu, &memory, &load); err != nil {
			rows.Close()
			return err
		}
		addAvg(cpuBuckets, since, bucket, reportedAt, cpu)
		addAvg(memoryBuckets, since, bucket, reportedAt, memory)
		addAvg(loadBuckets, since, bucket, reportedAt, load)
		updateLatestTime(&result.LastSampleAt, reportedAt)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	result.CPUTrend = avgPoints(cpuBuckets)
	result.MemoryTrend = avgPoints(memoryBuckets)
	result.LoadTrend = avgPoints(loadBuckets)

	diskBuckets := map[int64]*maxBucket{}
	rows, err = r.pool.Query(ctx, `
SELECT reported_at, COALESCE(usage, 0)
FROM disk_samples
WHERE reported_at >= $1
ORDER BY reported_at
`, since)
	if err != nil {
		return err
	}
	for rows.Next() {
		var reportedAt time.Time
		var usage float64
		if err := rows.Scan(&reportedAt, &usage); err != nil {
			rows.Close()
			return err
		}
		addMax(diskBuckets, since, bucket, reportedAt, usage)
		updateLatestTime(&result.LastSampleAt, reportedAt)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	result.DiskTrend = maxPoints(diskBuckets)

	latencyBuckets := map[int64]*avgBucket{}
	rows, err = r.pool.Query(ctx, `
SELECT reported_at, COALESCE(latency_ms, 0)
FROM http_check_results
WHERE reported_at >= $1
UNION ALL
SELECT reported_at, COALESCE(latency_ms, 0)
FROM service_check_results
WHERE reported_at >= $1
ORDER BY reported_at
`, since)
	if err != nil {
		return err
	}
	for rows.Next() {
		var reportedAt time.Time
		var latency int64
		if err := rows.Scan(&reportedAt, &latency); err != nil {
			rows.Close()
			return err
		}
		addAvg(latencyBuckets, since, bucket, reportedAt, float64(latency))
		updateLatestTime(&result.LastSampleAt, reportedAt)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	result.LatencyTrend = avgPoints(latencyBuckets)
	return nil
}

func (r *Repository) loadNodeLatestSystem(ctx context.Context, nodeID string, result *model.NodeMetrics) error {
	var item model.SystemSnapshot
	err := r.pool.QueryRow(ctx, `
SELECT COALESCE(cpu_usage, 0), COALESCE(memory_usage, 0), COALESCE(memory_used, 0), COALESCE(memory_total, 0),
       COALESCE(load1, 0), COALESCE(load5, 0), COALESCE(load15, 0), COALESCE(uptime_seconds, 0), reported_at
FROM system_samples
WHERE node_id = $1
ORDER BY reported_at DESC
LIMIT 1
`, nodeID).Scan(&item.CPUUsage, &item.MemoryUsage, &item.MemoryUsed, &item.MemoryTotal, &item.Load1, &item.Load5, &item.Load15, &item.UptimeSeconds, &item.ReportedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		return err
	}
	result.Latest.System = &item
	updateLatestTime(&result.LastSampleAt, item.ReportedAt)
	return nil
}

func (r *Repository) loadNodeLatestDisks(ctx context.Context, nodeID string, result *model.NodeMetrics) error {
	rows, err := r.pool.Query(ctx, `
SELECT mount, COALESCE(usage, 0), COALESCE(inode_usage, 0), COALESCE(used, 0), COALESCE(total, 0), reported_at
FROM (
    SELECT mount, usage, inode_usage, used, total, reported_at,
           row_number() OVER (PARTITION BY mount ORDER BY reported_at DESC) AS rn
    FROM disk_samples
    WHERE node_id = $1
) latest
WHERE rn = 1
ORDER BY mount
`, nodeID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item model.DiskSnapshot
		if err := rows.Scan(&item.Mount, &item.Usage, &item.InodeUsage, &item.Used, &item.Total, &item.ReportedAt); err != nil {
			return err
		}
		result.Latest.Disks = append(result.Latest.Disks, item)
		updateLatestTime(&result.LastSampleAt, item.ReportedAt)
	}
	return rows.Err()
}

func (r *Repository) loadNodeLatestNetworks(ctx context.Context, nodeID string, result *model.NodeMetrics) error {
	rows, err := r.pool.Query(ctx, `
SELECT name, COALESCE(bytes_sent, 0), COALESCE(bytes_recv, 0), COALESCE(packets_sent, 0), COALESCE(packets_recv, 0), reported_at
FROM (
    SELECT name, bytes_sent, bytes_recv, packets_sent, packets_recv, reported_at,
           row_number() OVER (PARTITION BY name ORDER BY reported_at DESC) AS rn
    FROM network_samples
    WHERE node_id = $1
) latest
WHERE rn <= 2
ORDER BY name, reported_at DESC
`, nodeID)
	if err != nil {
		return err
	}
	defer rows.Close()
	latest := map[string]*model.NetworkSnapshot{}
	previous := map[string]networkRaw{}
	for rows.Next() {
		var name string
		var item networkRaw
		if err := rows.Scan(&name, &item.bytesSent, &item.bytesRecv, &item.packetsSent, &item.packetsRecv, &item.reportedAt); err != nil {
			return err
		}
		if _, ok := latest[name]; !ok {
			latest[name] = &model.NetworkSnapshot{
				Name:        name,
				BytesSent:   item.bytesSent,
				BytesRecv:   item.bytesRecv,
				PacketsSent: item.packetsSent,
				PacketsRecv: item.packetsRecv,
				ReportedAt:  item.reportedAt,
			}
			updateLatestTime(&result.LastSampleAt, item.reportedAt)
			continue
		}
		previous[name] = item
	}
	if err := rows.Err(); err != nil {
		return err
	}
	names := sortedNetworkNames(latest)
	for _, name := range names {
		item := latest[name]
		if prev, ok := previous[name]; ok {
			seconds := item.ReportedAt.Sub(prev.reportedAt).Seconds()
			if seconds > 0 {
				item.TxBytesPerSec = positiveRate(item.BytesSent-prev.bytesSent, seconds)
				item.RxBytesPerSec = positiveRate(item.BytesRecv-prev.bytesRecv, seconds)
			}
		}
		result.Latest.Networks = append(result.Latest.Networks, *item)
	}
	return nil
}

func (r *Repository) loadNodeLatestDocker(ctx context.Context, nodeID string, result *model.NodeMetrics) error {
	rows, err := r.pool.Query(ctx, `
SELECT name, running, status, health_status, restart_count, error_message, reported_at
FROM (
    SELECT name, running, status, health_status, restart_count, error_message, reported_at,
           row_number() OVER (PARTITION BY name ORDER BY reported_at DESC) AS rn
    FROM docker_container_samples
    WHERE node_id = $1
) latest
WHERE rn = 1
ORDER BY name
`, nodeID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item model.DockerSnapshot
		if err := rows.Scan(&item.Name, &item.Running, &item.Status, &item.HealthStatus, &item.RestartCount, &item.ErrorMessage, &item.ReportedAt); err != nil {
			return err
		}
		result.Latest.Docker = append(result.Latest.Docker, item)
		updateLatestTime(&result.LastSampleAt, item.ReportedAt)
	}
	return rows.Err()
}

func (r *Repository) loadNodeLatestHTTP(ctx context.Context, nodeID string, result *model.NodeMetrics) error {
	rows, err := r.pool.Query(ctx, `
SELECT name, url, status, COALESCE(status_code, 0), COALESCE(latency_ms, 0), error_message, reported_at
FROM (
    SELECT name, url, status, status_code, latency_ms, error_message, reported_at,
           row_number() OVER (PARTITION BY name ORDER BY reported_at DESC) AS rn
    FROM http_check_results
    WHERE node_id = $1
) latest
WHERE rn = 1
ORDER BY name
`, nodeID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item model.HTTPCheckSnapshot
		if err := rows.Scan(&item.Name, &item.URL, &item.Status, &item.StatusCode, &item.LatencyMS, &item.ErrorMessage, &item.ReportedAt); err != nil {
			return err
		}
		result.Latest.HTTPChecks = append(result.Latest.HTTPChecks, item)
		updateLatestTime(&result.LastSampleAt, item.ReportedAt)
	}
	return rows.Err()
}

func (r *Repository) loadNodeLatestServices(ctx context.Context, nodeID string, result *model.NodeMetrics) error {
	rows, err := r.pool.Query(ctx, `
SELECT service_type, name, status, COALESCE(latency_ms, 0), error_message,
       COALESCE(database_size, 0), COALESCE(connections, 0), COALESCE(memory_used, 0), COALESCE(key_count, 0), reported_at
FROM (
    SELECT service_type, name, status, latency_ms, error_message, database_size, connections, memory_used, key_count, reported_at,
           row_number() OVER (PARTITION BY service_type, name ORDER BY reported_at DESC) AS rn
    FROM service_check_results
    WHERE node_id = $1
) latest
WHERE rn = 1
ORDER BY service_type, name
`, nodeID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item model.ServiceCheckSnapshot
		if err := rows.Scan(&item.ServiceType, &item.Name, &item.Status, &item.LatencyMS, &item.ErrorMessage, &item.DatabaseSize, &item.Connections, &item.MemoryUsed, &item.KeyCount, &item.ReportedAt); err != nil {
			return err
		}
		result.Latest.ServiceChecks = append(result.Latest.ServiceChecks, item)
		updateLatestTime(&result.LastSampleAt, item.ReportedAt)
	}
	return rows.Err()
}

func (r *Repository) loadNodeLatestCerts(ctx context.Context, nodeID string, result *model.NodeMetrics) error {
	rows, err := r.pool.Query(ctx, `
SELECT name, host, status, expires_at, COALESCE(days_remaining, 0), COALESCE(matched_name, false), error_message, reported_at
FROM (
    SELECT name, host, status, expires_at, days_remaining, matched_name, error_message, reported_at,
           row_number() OVER (PARTITION BY name ORDER BY reported_at DESC) AS rn
    FROM cert_check_results
    WHERE node_id = $1
) latest
WHERE rn = 1
ORDER BY name
`, nodeID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item model.CertSnapshot
		var expires sql.NullTime
		if err := rows.Scan(&item.Name, &item.Host, &item.Status, &expires, &item.DaysRemaining, &item.MatchedName, &item.ErrorMessage, &item.ReportedAt); err != nil {
			return err
		}
		if expires.Valid {
			t := expires.Time
			item.ExpiresAt = &t
		}
		result.Latest.Certs = append(result.Latest.Certs, item)
		updateLatestTime(&result.LastSampleAt, item.ReportedAt)
	}
	return rows.Err()
}

func (r *Repository) loadNodeTrends(ctx context.Context, nodeID string, since time.Time, bucket time.Duration, result *model.NodeMetrics) error {
	cpuBuckets := map[int64]*avgBucket{}
	memoryBuckets := map[int64]*avgBucket{}
	loadBuckets := map[int64]*avgBucket{}
	rows, err := r.pool.Query(ctx, `
SELECT reported_at, COALESCE(cpu_usage, 0), COALESCE(memory_usage, 0), COALESCE(load1, 0)
FROM system_samples
WHERE node_id = $1 AND reported_at >= $2
ORDER BY reported_at
`, nodeID, since)
	if err != nil {
		return err
	}
	for rows.Next() {
		var reportedAt time.Time
		var cpu, memory, load float64
		if err := rows.Scan(&reportedAt, &cpu, &memory, &load); err != nil {
			rows.Close()
			return err
		}
		addAvg(cpuBuckets, since, bucket, reportedAt, cpu)
		addAvg(memoryBuckets, since, bucket, reportedAt, memory)
		addAvg(loadBuckets, since, bucket, reportedAt, load)
		updateLatestTime(&result.LastSampleAt, reportedAt)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	result.Trends.CPU = avgPoints(cpuBuckets)
	result.Trends.Memory = avgPoints(memoryBuckets)
	result.Trends.Load = avgPoints(loadBuckets)

	diskSeries := map[string]map[int64]*avgBucket{}
	rows, err = r.pool.Query(ctx, `
SELECT mount, reported_at, COALESCE(usage, 0)
FROM disk_samples
WHERE node_id = $1 AND reported_at >= $2
ORDER BY mount, reported_at
`, nodeID, since)
	if err != nil {
		return err
	}
	for rows.Next() {
		var mount string
		var reportedAt time.Time
		var usage float64
		if err := rows.Scan(&mount, &reportedAt, &usage); err != nil {
			rows.Close()
			return err
		}
		addSeriesAvg(diskSeries, mount, since, bucket, reportedAt, usage)
		updateLatestTime(&result.LastSampleAt, reportedAt)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	result.Trends.DiskUsage = seriesFromAvgBuckets(diskSeries, "%", 12)

	if err := r.loadNodeNetworkTrends(ctx, nodeID, since, bucket, result); err != nil {
		return err
	}
	if err := r.loadNodeServiceLatencyTrends(ctx, nodeID, since, bucket, result); err != nil {
		return err
	}
	return nil
}

func (r *Repository) loadNodeNetworkTrends(ctx context.Context, nodeID string, since time.Time, bucket time.Duration, result *model.NodeMetrics) error {
	rxBuckets := map[int64]*avgBucket{}
	txBuckets := map[int64]*avgBucket{}
	rows, err := r.pool.Query(ctx, `
SELECT name, COALESCE(bytes_sent, 0), COALESCE(bytes_recv, 0), reported_at
FROM network_samples
WHERE node_id = $1 AND reported_at >= $2
ORDER BY name, reported_at
`, nodeID, since)
	if err != nil {
		return err
	}
	defer rows.Close()
	previous := map[string]networkRaw{}
	for rows.Next() {
		var name string
		var item networkRaw
		if err := rows.Scan(&name, &item.bytesSent, &item.bytesRecv, &item.reportedAt); err != nil {
			return err
		}
		if prev, ok := previous[name]; ok {
			seconds := item.reportedAt.Sub(prev.reportedAt).Seconds()
			if seconds > 0 {
				addAvg(txBuckets, since, bucket, item.reportedAt, positiveRate(item.bytesSent-prev.bytesSent, seconds))
				addAvg(rxBuckets, since, bucket, item.reportedAt, positiveRate(item.bytesRecv-prev.bytesRecv, seconds))
			}
		}
		previous[name] = item
		updateLatestTime(&result.LastSampleAt, item.reportedAt)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	result.Trends.NetworkRx = avgPoints(rxBuckets)
	result.Trends.NetworkTx = avgPoints(txBuckets)
	return nil
}

func (r *Repository) loadNodeServiceLatencyTrends(ctx context.Context, nodeID string, since time.Time, bucket time.Duration, result *model.NodeMetrics) error {
	series := map[string]map[int64]*avgBucket{}
	rows, err := r.pool.Query(ctx, `
SELECT 'http' AS service_type, name, reported_at, COALESCE(latency_ms, 0)
FROM http_check_results
WHERE node_id = $1 AND reported_at >= $2
UNION ALL
SELECT service_type, name, reported_at, COALESCE(latency_ms, 0)
FROM service_check_results
WHERE node_id = $1 AND reported_at >= $2
ORDER BY service_type, name, reported_at
`, nodeID, since)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var serviceType, name string
		var reportedAt time.Time
		var latency int64
		if err := rows.Scan(&serviceType, &name, &reportedAt, &latency); err != nil {
			return err
		}
		addSeriesAvg(series, serviceType+"/"+name, since, bucket, reportedAt, float64(latency))
		updateLatestTime(&result.LastSampleAt, reportedAt)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	result.Trends.ServiceLatency = seriesFromAvgBuckets(series, "ms", 8)
	return nil
}

func addAvg(buckets map[int64]*avgBucket, since time.Time, bucket time.Duration, at time.Time, value float64) {
	key := bucketKey(since, bucket, at)
	item := buckets[key]
	if item == nil {
		item = &avgBucket{}
		buckets[key] = item
	}
	item.sum += value
	item.count++
}

func addSeriesAvg(series map[string]map[int64]*avgBucket, name string, since time.Time, bucket time.Duration, at time.Time, value float64) {
	buckets := series[name]
	if buckets == nil {
		buckets = map[int64]*avgBucket{}
		series[name] = buckets
	}
	addAvg(buckets, since, bucket, at, value)
}

func addMax(buckets map[int64]*maxBucket, since time.Time, bucket time.Duration, at time.Time, value float64) {
	key := bucketKey(since, bucket, at)
	item := buckets[key]
	if item == nil {
		item = &maxBucket{value: value, seen: true}
		buckets[key] = item
		return
	}
	if !item.seen || value > item.value {
		item.value = value
		item.seen = true
	}
}

func avgPoints(buckets map[int64]*avgBucket) []model.MetricPoint {
	keys := sortedBucketKeys(buckets)
	points := make([]model.MetricPoint, 0, len(keys))
	for _, key := range keys {
		item := buckets[key]
		if item == nil || item.count == 0 {
			continue
		}
		points = append(points, model.MetricPoint{Timestamp: time.Unix(0, key).UTC(), Value: item.sum / float64(item.count)})
	}
	return points
}

func maxPoints(buckets map[int64]*maxBucket) []model.MetricPoint {
	keys := sortedMaxBucketKeys(buckets)
	points := make([]model.MetricPoint, 0, len(keys))
	for _, key := range keys {
		item := buckets[key]
		if item == nil || !item.seen {
			continue
		}
		points = append(points, model.MetricPoint{Timestamp: time.Unix(0, key).UTC(), Value: item.value})
	}
	return points
}

func seriesFromAvgBuckets(series map[string]map[int64]*avgBucket, unit string, limit int) []model.MetricSeries {
	names := make([]string, 0, len(series))
	for name := range series {
		names = append(names, name)
	}
	sort.Strings(names)
	if limit > 0 && len(names) > limit {
		names = names[:limit]
	}
	items := make([]model.MetricSeries, 0, len(names))
	for _, name := range names {
		items = append(items, model.MetricSeries{Name: name, Unit: unit, Points: avgPoints(series[name])})
	}
	return items
}

func bucketKey(since time.Time, bucket time.Duration, at time.Time) int64 {
	if bucket <= 0 {
		bucket = time.Minute
	}
	diff := at.Sub(since)
	if diff < 0 {
		diff = 0
	}
	return since.Add((diff / bucket) * bucket).Truncate(time.Second).UnixNano()
}

func sortedBucketKeys(buckets map[int64]*avgBucket) []int64 {
	keys := make([]int64, 0, len(buckets))
	for key := range buckets {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortedMaxBucketKeys(buckets map[int64]*maxBucket) []int64 {
	keys := make([]int64, 0, len(buckets))
	for key := range buckets {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func updateLatestTime(target **time.Time, value time.Time) {
	if value.IsZero() {
		return
	}
	if *target == nil || value.After(**target) {
		t := value
		*target = &t
	}
}

func limitTopResources(items []model.TopResource, limit int) []model.TopResource {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Value == items[j].Value {
			return items[i].NodeID < items[j].NodeID
		}
		return items[i].Value > items[j].Value
	})
	if len(items) > limit {
		return items[:limit]
	}
	return items
}

func sortedNetworkNames(items map[string]*model.NetworkSnapshot) []string {
	names := make([]string, 0, len(items))
	for name := range items {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func positiveRate(delta int64, seconds float64) float64 {
	if delta <= 0 || seconds <= 0 {
		return 0
	}
	return float64(delta) / seconds
}
