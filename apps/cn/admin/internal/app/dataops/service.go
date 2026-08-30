// Package dataops exposes bounded, read-only PostgreSQL health metadata.
// The catalog queries are intentionally direct pgx: they are identical across
// the three explicit pools, do not express business persistence, and are kept
// static here instead of creating a generic database abstraction.
package dataops

import (
	"context"
	"time"

	"github.com/gofurry/gofurry-admin/internal/infra/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

const databaseMetadataSQL = `
SELECT current_setting('server_version'), current_database(),
       pg_database_size(current_database())::bigint, clock_timestamp(),
       current_setting('max_connections')::bigint,
       (SELECT COUNT(*)::bigint FROM pg_stat_activity WHERE datname=current_database()),
       (SELECT COUNT(*)::bigint FROM pg_stat_activity WHERE datname=current_database() AND state='active')`

const migrationStateSQL = `
SELECT version_id, is_applied
FROM (
    SELECT version_id, is_applied,
           ROW_NUMBER() OVER (PARTITION BY version_id ORDER BY id DESC) AS position
    FROM public.goose_db_version
) latest
WHERE position=1
ORDER BY version_id`

const largestRelationsSQL = `
SELECT quote_ident(namespace.nspname) || '.' || quote_ident(relation.relname) AS relation_name,
       pg_relation_size(relation.oid)::bigint AS table_bytes,
       pg_indexes_size(relation.oid)::bigint AS index_bytes,
       pg_total_relation_size(relation.oid)::bigint AS total_bytes
FROM pg_class relation
JOIN pg_namespace namespace ON namespace.oid=relation.relnamespace
WHERE namespace.nspname='public' AND relation.relkind IN ('r','m','p')
  AND relation.relname <> 'goose_db_version'
ORDER BY total_bytes DESC, relation_name
LIMIT 10`

var expectedRepoMigrations = map[string][]int64{
	"gfa": {20260823000000, 20260830030000},
	"gfn": {20260823000000, 20260823000001, 20260827010000, 20260827010100, 20260828010000, 20260828010100, 20260828010200, 20260828010300, 20260829010000, 20260829020000, 20260830010000, 20260830020000},
	"gfg": {20260823000000, 20260823000001, 20260824000000, 20260827010000, 20260827010100, 20260827010200, 20260827010300, 20260827010400, 20260828010000, 20260828010100, 20260828010200, 20260828010300, 20260829020000, 20260830010000},
}

type Service struct{ pools *db.Pools }

func New(pools *db.Pools) *Service { return &Service{pools: pools} }

func (service *Service) Overview(ctx context.Context) Overview {
	overview := Overview{GeneratedAt: time.Now().UTC(), Databases: make([]DatabaseStatus, 0, 3)}
	if service == nil || service.pools == nil {
		for _, key := range []string{"gfa", "gfn", "gfg"} {
			overview.Databases = append(overview.Databases, unavailableDatabase(key))
		}
		return overview
	}
	for _, target := range []struct {
		key  string
		pool *pgxpool.Pool
	}{{"gfa", service.pools.Admin}, {"gfn", service.pools.Nav}, {"gfg", service.pools.Game}} {
		overview.Databases = append(overview.Databases, inspectDatabase(ctx, target.key, target.pool))
	}
	return overview
}

func inspectDatabase(parent context.Context, key string, pool *pgxpool.Pool) DatabaseStatus {
	result := unavailableDatabase(key)
	if pool == nil {
		return result
	}
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	var databaseTime time.Time
	if err := pool.QueryRow(ctx, databaseMetadataSQL).Scan(
		&result.PostgreSQLVersion, &result.DatabaseName, &result.DatabaseSizeBytes, &databaseTime,
		&result.MaxConnections, &result.TotalConnections, &result.ActiveConnections,
	); err != nil {
		return result
	}
	result.DatabaseTime = &databaseTime
	if result.MaxConnections > 0 {
		usage := float64(result.TotalConnections) / float64(result.MaxConnections)
		result.ConnectionUsage = &usage
	}
	result.Health = "healthy"
	result.Error = ""
	result.Migration = inspectMigrations(ctx, key, pool)
	if result.Migration.Status != "current" {
		result.Health = "warning"
	}
	result.Relations = inspectRelations(ctx, pool)
	return result
}

func inspectMigrations(ctx context.Context, key string, pool *pgxpool.Pool) MigrationStatus {
	expected := expectedRepoMigrations[key]
	status := MigrationStatus{Status: "unavailable"}
	if len(expected) > 0 {
		status.Expected = expected[len(expected)-1]
	}
	rows, err := pool.Query(ctx, migrationStateSQL)
	if err != nil {
		return status
	}
	defer rows.Close()
	applied := make(map[int64]bool, len(expected))
	for rows.Next() {
		var version int64
		var isApplied bool
		if rows.Scan(&version, &isApplied) != nil {
			return status
		}
		applied[version] = isApplied
		if isApplied && version > status.CurrentApplied {
			status.CurrentApplied = version
		}
	}
	if rows.Err() != nil {
		return status
	}
	for _, version := range expected {
		if !applied[version] {
			status.PendingCount++
		}
	}
	status.Status = migrationStatus(status.CurrentApplied, status.Expected, status.PendingCount)
	return status
}

func migrationStatus(current, expected int64, pending int) string {
	if current > expected {
		return "ahead"
	}
	if pending > 0 || current < expected {
		return "behind"
	}
	return "current"
}

func inspectRelations(ctx context.Context, pool *pgxpool.Pool) []RelationSize {
	rows, err := pool.Query(ctx, largestRelationsSQL)
	if err != nil {
		return []RelationSize{}
	}
	defer rows.Close()
	relations := make([]RelationSize, 0, 10)
	for rows.Next() {
		var relation RelationSize
		if rows.Scan(&relation.Name, &relation.TableBytes, &relation.IndexBytes, &relation.TotalBytes) != nil {
			return []RelationSize{}
		}
		relations = append(relations, relation)
	}
	if rows.Err() != nil {
		return []RelationSize{}
	}
	return relations
}

func unavailableDatabase(key string) DatabaseStatus {
	expected := expectedRepoMigrations[key]
	var expectedVersion int64
	if len(expected) > 0 {
		expectedVersion = expected[len(expected)-1]
	}
	return DatabaseStatus{
		Key: key, Health: "unavailable", Error: "database metadata unavailable",
		Migration: MigrationStatus{Expected: expectedVersion, Status: "unavailable"},
		Relations: []RelationSize{},
	}
}
