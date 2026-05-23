package retention

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const DefaultBatchSize = 500

var allowedRetentionTables = map[string]struct{}{
	"gfn_collector_log_ping":    {},
	"gfn_collector_log_http":    {},
	"gfn_collector_log_dns":     {},
	"gfn_collector_observation": {},
}

func BuildDeleteSQL(tableName string) (string, error) {
	if !isAllowedRetentionTable(tableName) {
		return "", errors.New("retention 表名不在允许列表: " + tableName)
	}
	return fmt.Sprintf(`
WITH doomed AS (
	SELECT id
	FROM (
		SELECT
			id,
			ROW_NUMBER() OVER (
				PARTITION BY name
				ORDER BY create_time DESC, id DESC
			) AS rn
		FROM %s
	) ranked
	WHERE ranked.rn > ?
	ORDER BY id
	LIMIT ?
)
DELETE FROM %s target
USING doomed
WHERE target.id = doomed.id;`, tableName, tableName), nil
}

func BuildObservationDeleteSQL(tableName string) (string, error) {
	if !isAllowedRetentionTable(tableName) {
		return "", errors.New("retention 表名不在允许列表: " + tableName)
	}
	return fmt.Sprintf(`
WITH doomed AS (
	SELECT id
	FROM (
		SELECT
			id,
			ROW_NUMBER() OVER (
				PARTITION BY site_id, protocol
				ORDER BY observed_at DESC, id DESC
			) AS rn
		FROM %s
		WHERE protocol = ?
	) ranked
	WHERE ranked.rn > ?
	ORDER BY id
	LIMIT ?
)
DELETE FROM %s target
USING doomed
WHERE target.id = doomed.id;`, tableName, tableName), nil
}

func isAllowedRetentionTable(tableName string) bool {
	_, ok := allowedRetentionTables[tableName]
	return ok
}

func DeleteByNameLimit(db *gorm.DB, tableName string, keepCount int, batchSize int, batchTimeout time.Duration, pause time.Duration) (int64, error) {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}
	if batchTimeout <= 0 {
		batchTimeout = 2 * time.Minute
	}

	var totalDeleted int64
	sql, err := BuildDeleteSQL(tableName)
	if err != nil {
		return 0, err
	}
	for {
		ctx, cancel := context.WithTimeout(context.Background(), batchTimeout)
		result := db.WithContext(ctx).Exec(sql, keepCount, batchSize)
		cancel()

		if result.Error != nil {
			return totalDeleted, result.Error
		}

		deleted := result.RowsAffected
		totalDeleted += deleted
		if deleted < int64(batchSize) {
			break
		}
		if pause > 0 {
			time.Sleep(pause)
		}
	}

	return totalDeleted, nil
}

func DeleteObservationByProtocolLimit(db *gorm.DB, tableName string, protocol string, keepCount int, batchSize int, batchTimeout time.Duration, pause time.Duration) (int64, error) {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}
	if batchTimeout <= 0 {
		batchTimeout = 2 * time.Minute
	}

	var totalDeleted int64
	sql, err := BuildObservationDeleteSQL(tableName)
	if err != nil {
		return 0, err
	}
	for {
		ctx, cancel := context.WithTimeout(context.Background(), batchTimeout)
		result := db.WithContext(ctx).Exec(sql, protocol, keepCount, batchSize)
		cancel()

		if result.Error != nil {
			return totalDeleted, result.Error
		}

		deleted := result.RowsAffected
		totalDeleted += deleted
		if deleted < int64(batchSize) {
			break
		}
		if pause > 0 {
			time.Sleep(pause)
		}
	}

	return totalDeleted, nil
}
