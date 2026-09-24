package cloudops

import (
	"context"
	"errors"

	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	adminsqlc "github.com/gofurry/gofurry-admin/internal/db/admin/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgScheduledPurgeStore struct {
	pool   *pgxpool.Pool
	logger *audit.Logger
}

type pgScheduledPurgeLease struct {
	conn    *pgxpool.Conn
	queries *adminsqlc.Queries
	logger  *audit.Logger
}

func (store *pgScheduledPurgeStore) Acquire(ctx context.Context) (scheduledPurgeLease, error) {
	conn, err := store.pool.Acquire(ctx)
	if err != nil {
		// Connection errors may contain a private DSN/host/user.
		return nil, errors.New("GFA connection unavailable for scheduled purge")
	}
	queries := adminsqlc.New(conn)
	acquired, err := queries.TryLockEdgeOneScheduledPurge(ctx)
	if err != nil {
		// A canceled query may have acquired the session lock on the server.
		// Never return that uncertain session to the pool.
		_ = conn.Conn().Close(ctx)
		conn.Release()
		return nil, errors.New("GFA scheduled purge lock query failed")
	}
	if !acquired {
		conn.Release()
		return nil, nil
	}
	return &pgScheduledPurgeLease{conn: conn, queries: queries, logger: store.logger}, nil
}

func (lease *pgScheduledPurgeLease) Recorded(ctx context.Context, slot, host string) (bool, error) {
	return lease.queries.HasEdgeOneScheduledPurgeRequest(ctx, adminsqlc.HasEdgeOneScheduledPurgeRequestParams{RequestID: slot, TargetID: host})
}

func (lease *pgScheduledPurgeLease) Audit(ctx context.Context, meta audit.Meta, host string, detail any) error {
	return lease.logger.LogConn(ctx, lease.conn, meta, scheduledPurgeAction, "cloud", host, nil, detail)
}

func (lease *pgScheduledPurgeLease) Release(ctx context.Context) error {
	defer lease.conn.Release()
	released, err := lease.queries.UnlockEdgeOneScheduledPurge(ctx)
	if err != nil || !released {
		_ = lease.conn.Conn().Close(ctx)
		return errors.New("GFA scheduled purge unlock failed; session closed")
	}
	return nil
}
