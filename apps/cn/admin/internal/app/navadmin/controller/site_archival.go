package controller

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofurry/gofurry-admin/internal/app/navadmin/models"
	navsqlc "github.com/gofurry/gofurry-admin/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func normalizeSiteArchival(req models.SiteArchivalPayload) (models.SiteArchivalPayload, common.Error) {
	switch req.DeletionReasonCode {
	case "unreachable", "closed", "migrated", "duplicate", "owner_request", "other":
	default:
		return req, common.NewValidationError("deletion_reason_code must be a supported explicit archival reason")
	}
	req.DeletionReason = normalizeStringPtr(req.DeletionReason)
	if req.DeletionReason != nil && (!utf8.ValidString(*req.DeletionReason) || strings.ContainsRune(*req.DeletionReason, 0) || utf8.RuneCountInString(*req.DeletionReason) > 1000) {
		return req, common.NewValidationError("deletion_reason must be valid text of at most 1000 characters")
	}
	return req, nil
}

// Lock the Site through commit so an archival cannot race target mutations.
func requireActiveSite(ctx context.Context, q *navsqlc.Queries, id int64) error {
	site, err := q.LockSiteForUpdate(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return common.NewValidationError("site_id must reference an existing site")
	}
	if err != nil {
		return err
	}
	if site.Deleted {
		return common.NewValidationError("archived site is read-only")
	}
	return nil
}

func siteDeletedAt(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

// Match archival's Site-before-period lock order. A concurrently moved domain
// must be reloaded instead of mutating an unchecked former owner.
func lockActiveCollectorDomain(ctx context.Context, q *navsqlc.Queries, id int64, additionalSites ...int64) (navsqlc.GfnCollectorDomain, error) {
	before, err := q.GetCollectorDomainAny(ctx, id)
	if err != nil {
		return before, err
	}
	ids := append(additionalSites, pointerInt64(before.SiteID))
	slices.Sort(ids)
	for _, siteID := range slices.Compact(ids) {
		if err := requireActiveSite(ctx, q, siteID); err != nil {
			return before, err
		}
	}
	locked, err := q.LockCollectorDomainForUpdate(ctx, id)
	if err == nil && pointerInt64(locked.SiteID) != pointerInt64(before.SiteID) {
		return locked, common.NewConflictError("collector domain owner changed; reload before editing")
	}
	return locked, err
}
