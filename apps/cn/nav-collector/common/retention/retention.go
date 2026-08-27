package retention

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DefaultBatchSize = 500

// These four fixed statements are a local direct-pgx exception. sqlc 1.31
// cannot resolve the window-column alias in a ranked CTE used by DELETE, while
// PostgreSQL executes the statements correctly. No table identifier is dynamic
// and every runtime value is parameterized.
const (
	deletePingSQL = `WITH doomed AS (
SELECT id FROM (SELECT id, ROW_NUMBER() OVER (PARTITION BY name ORDER BY create_time DESC,id DESC) rn
FROM gfn_collector_log_ping) ranked WHERE rn>$1 ORDER BY id LIMIT $2)
DELETE FROM gfn_collector_log_ping target USING doomed WHERE target.id=doomed.id`
	deleteHTTPSQL = `WITH doomed AS (
SELECT id FROM (SELECT id, ROW_NUMBER() OVER (PARTITION BY name ORDER BY create_time DESC,id DESC) rn
FROM gfn_collector_log_http) ranked WHERE rn>$1 ORDER BY id LIMIT $2)
DELETE FROM gfn_collector_log_http target USING doomed WHERE target.id=doomed.id`
	deleteDNSSQL = `WITH doomed AS (
SELECT id FROM (SELECT id, ROW_NUMBER() OVER (PARTITION BY name ORDER BY create_time DESC,id DESC) rn
FROM gfn_collector_log_dns) ranked WHERE rn>$1 ORDER BY id LIMIT $2)
DELETE FROM gfn_collector_log_dns target USING doomed WHERE target.id=doomed.id`
)

func DeletePingByNameLimit(pool *pgxpool.Pool, keepCount int, batchSize int, batchTimeout time.Duration, pause time.Duration) (int64, error) {
	return deleteBatches(batchSize, batchTimeout, pause, func(ctx context.Context, batchSize int) (int64, error) {
		tag, err := pool.Exec(ctx, deletePingSQL, keepCount, batchSize)
		return tag.RowsAffected(), err
	})
}

func DeleteHTTPByNameLimit(pool *pgxpool.Pool, keepCount int, batchSize int, batchTimeout time.Duration, pause time.Duration) (int64, error) {
	return deleteBatches(batchSize, batchTimeout, pause, func(ctx context.Context, batchSize int) (int64, error) {
		tag, err := pool.Exec(ctx, deleteHTTPSQL, keepCount, batchSize)
		return tag.RowsAffected(), err
	})
}

func DeleteDNSByNameLimit(pool *pgxpool.Pool, keepCount int, batchSize int, batchTimeout time.Duration, pause time.Duration) (int64, error) {
	return deleteBatches(batchSize, batchTimeout, pause, func(ctx context.Context, batchSize int) (int64, error) {
		tag, err := pool.Exec(ctx, deleteDNSSQL, keepCount, batchSize)
		return tag.RowsAffected(), err
	})
}

func deleteBatches(batchSize int, batchTimeout time.Duration, pause time.Duration, deleteBatch func(context.Context, int) (int64, error)) (int64, error) {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}
	if batchTimeout <= 0 {
		batchTimeout = 2 * time.Minute
	}
	var total int64
	for {
		ctx, cancel := context.WithTimeout(context.Background(), batchTimeout)
		deleted, err := deleteBatch(ctx, batchSize)
		cancel()
		total += deleted
		if err != nil {
			return total, err
		}
		if deleted < int64(batchSize) {
			return total, nil
		}
		if pause > 0 {
			time.Sleep(pause)
		}
	}
}
