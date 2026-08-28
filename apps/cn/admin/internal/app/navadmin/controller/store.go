package controller

import (
	"context"
	"errors"
	"time"

	"github.com/gofurry/gofurry-admin/internal/app/navadmin/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	navsqlc "github.com/gofurry/gofurry-admin/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	pkgmodels "github.com/gofurry/gofurry-admin/pkg/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type navStore struct {
	pool  *pgxpool.Pool
	q     *navsqlc.Queries
	audit *audit.Logger
}

type navMutation func(*navsqlc.Queries) (targetID int64, before any, after any, err error)

func newNavStore(pool *pgxpool.Pool, auditLogger *audit.Logger) *navStore {
	return &navStore{pool: pool, q: navsqlc.New(pool), audit: auditLogger}
}

func (store *navStore) mutate(ctx context.Context, meta audit.Meta, action, resource string, change navMutation) common.Error {
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return navDAOError(err)
	}
	defer tx.Rollback(ctx)
	targetID, before, after, err := change(store.q.WithTx(tx))
	if err != nil {
		return navDAOError(err)
	}
	// gfn and gfa are separate databases. As before, audit is committed independently
	// while the business transaction is still open; an audit failure rolls gfn back.
	if auditErr := store.audit.Log(ctx, meta, action, resource, targetID, before, after); auditErr != nil {
		return auditErr
	}
	if err := tx.Commit(ctx); err != nil {
		return navDAOError(err)
	}
	return nil
}

func pageArgs(page adminutil.PageQuery) (int32, int32) {
	return int32(page.PageSize), int32((page.PageNum - 1) * page.PageSize)
}

func (store *navStore) listSayings(ctx context.Context, page adminutil.PageQuery) (int64, []models.Saying, common.Error) {
	total, err := store.q.CountSayings(ctx, page.Keyword)
	if err != nil {
		return 0, nil, navDAOError(err)
	}
	limit, offset := pageArgs(page)
	rows, err := store.q.ListSayings(ctx, navsqlc.ListSayingsParams{Keyword: page.Keyword, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return 0, nil, navDAOError(err)
	}
	items := make([]models.Saying, 0, len(rows))
	for _, row := range rows {
		items = append(items, sayingModel(row))
	}
	return total, items, nil
}

func (store *navStore) getSaying(ctx context.Context, id int64) (models.Saying, common.Error) {
	row, err := store.q.GetSaying(ctx, id)
	return sayingModel(row), navDAOError(err)
}

func (store *navStore) createSaying(ctx context.Context, meta audit.Meta, req models.SayingPayload) (models.Saying, common.Error) {
	var result models.Saying
	err := store.mutate(ctx, meta, "create", "gfn_saying", func(q *navsqlc.Queries) (int64, any, any, error) {
		id, err := q.NextSayingID(ctx)
		if err != nil {
			return 0, nil, nil, err
		}
		row, err := q.InsertSaying(ctx, navsqlc.InsertSayingParams{ID: id, Author: req.Author, Saying: req.Saying, Language: req.Language})
		result = sayingModel(row)
		return id, nil, row, err
	})
	return result, err
}

func (store *navStore) updateSaying(ctx context.Context, meta audit.Meta, id int64, req models.SayingPayload) common.Error {
	return store.mutate(ctx, meta, "update", "gfn_saying", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, err := q.GetSaying(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		after, err := q.UpdateSaying(ctx, navsqlc.UpdateSayingParams{ID: id, Author: req.Author, Saying: req.Saying, Language: req.Language})
		return id, before, after, err
	})
}

func (store *navStore) deleteSaying(ctx context.Context, meta audit.Meta, id int64) common.Error {
	return store.mutate(ctx, meta, "delete", "gfn_saying", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, err := q.GetSaying(ctx, id)
		if err == nil {
			_, err = q.DeleteSaying(ctx, id)
		}
		return id, before, nil, err
	})
}

func (store *navStore) listUpdateNotices(ctx context.Context, page adminutil.PageQuery) (int64, []models.UpdateNotice, common.Error) {
	total, err := store.q.CountUpdateNotices(ctx, page.Keyword)
	if err != nil {
		return 0, nil, navDAOError(err)
	}
	limit, offset := pageArgs(page)
	rows, err := store.q.ListUpdateNotices(ctx, navsqlc.ListUpdateNoticesParams{Keyword: page.Keyword, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return 0, nil, navDAOError(err)
	}
	items := make([]models.UpdateNotice, 0, len(rows))
	for _, row := range rows {
		items = append(items, updateNoticeModel(row))
	}
	return total, items, nil
}

func (store *navStore) getUpdateNotice(ctx context.Context, id int64) (models.UpdateNotice, common.Error) {
	row, err := store.q.GetUpdateNotice(ctx, id)
	return updateNoticeModel(row), navDAOError(err)
}

func (store *navStore) createUpdateNotice(ctx context.Context, meta audit.Meta, req models.UpdateNoticePayload, publishedAt time.Time) (models.UpdateNotice, common.Error) {
	var result models.UpdateNotice
	err := store.mutate(ctx, meta, "create", "gfn_nav_update_notice", func(q *navsqlc.Queries) (int64, any, any, error) {
		id, err := q.NextUpdateNoticeID(ctx)
		if err != nil {
			return 0, nil, nil, err
		}
		row, err := q.InsertUpdateNotice(ctx, navsqlc.InsertUpdateNoticeParams{ID: id, Title: req.Title, TitleEn: req.TitleEn, Body: req.Body, BodyEn: req.BodyEn, PublishedAt: navTimestamp(publishedAt)})
		result = updateNoticeModel(row)
		return id, nil, row, err
	})
	return result, err
}

func (store *navStore) updateUpdateNotice(ctx context.Context, meta audit.Meta, id int64, req models.UpdateNoticePayload, publishedAt time.Time) common.Error {
	return store.mutate(ctx, meta, "update", "gfn_nav_update_notice", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, err := q.GetUpdateNoticeAny(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		after, err := q.UpdateUpdateNotice(ctx, navsqlc.UpdateUpdateNoticeParams{ID: id, Title: req.Title, TitleEn: req.TitleEn, Body: req.Body, BodyEn: req.BodyEn, PublishedAt: navTimestamp(publishedAt)})
		return id, before, after, err
	})
}

func (store *navStore) deleteUpdateNotice(ctx context.Context, meta audit.Meta, id int64) common.Error {
	return store.mutate(ctx, meta, "delete", "gfn_nav_update_notice", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, err := q.GetUpdateNoticeAny(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		after, err := q.SoftDeleteUpdateNotice(ctx, id)
		return id, before, after, err
	})
}

func (store *navStore) listCollectorDomains(ctx context.Context, page adminutil.PageQuery) (int64, []models.CollectorDomainDTO, common.Error) {
	total, err := store.q.CountCollectorDomains(ctx, page.Keyword)
	if err != nil {
		return 0, nil, navDAOError(err)
	}
	limit, offset := pageArgs(page)
	rows, err := store.q.ListCollectorDomains(ctx, navsqlc.ListCollectorDomainsParams{Keyword: page.Keyword, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return 0, nil, navDAOError(err)
	}
	items := make([]models.CollectorDomainDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, collectorDomainDTO(row.ID, pointerInt64(row.SiteID), row.SiteName, row.Name, row.Proxy, row.Prefix, row.Tls, row.Deleted))
	}
	return total, items, nil
}

func (store *navStore) getCollectorDomain(ctx context.Context, id int64) (models.CollectorDomainDTO, common.Error) {
	row, err := store.q.GetCollectorDomain(ctx, id)
	return collectorDomainDTO(row.ID, pointerInt64(row.SiteID), row.SiteName, row.Name, row.Proxy, row.Prefix, row.Tls, row.Deleted), navDAOError(err)
}

func validateCollectorDomainQuery(ctx context.Context, q *navsqlc.Queries, req models.CollectorDomainPayload) error {
	count, err := q.CountActiveSiteByID(ctx, req.SiteID)
	if err != nil {
		return err
	}
	if count == 0 {
		return common.NewValidationError("site_id must reference an existing site")
	}
	return nil
}

func collectorTarget(name string, prefix *string) string {
	if prefix == nil {
		return name
	}
	return *prefix + name
}

func validateCollectorTargetIdentity(ctx context.Context, q *navsqlc.Queries, req models.CollectorDomainPayload, excludeDomainID *int64) error {
	conflicts, err := q.CountConflictingActiveTarget(ctx, navsqlc.CountConflictingActiveTargetParams{
		SiteID: req.SiteID, Target: collectorTarget(req.Name, req.Prefix), ExcludeDomainID: excludeDomainID,
	})
	if err != nil {
		return err
	}
	if conflicts > 0 {
		return common.NewValidationError("site already has an active target with the same normalized identity")
	}
	return nil
}

func (store *navStore) createCollectorDomain(ctx context.Context, meta audit.Meta, req models.CollectorDomainPayload) (models.CollectorDomain, common.Error) {
	var result models.CollectorDomain
	err := store.mutate(ctx, meta, "create", "gfn_collector_domain", func(q *navsqlc.Queries) (int64, any, any, error) {
		if err := validateCollectorDomainQuery(ctx, q, req); err != nil {
			return 0, nil, nil, err
		}
		if err := validateCollectorTargetIdentity(ctx, q, req, nil); err != nil {
			return 0, nil, nil, err
		}
		id, err := q.NextCollectorDomainID(ctx)
		if err != nil {
			return 0, nil, nil, err
		}
		row, err := q.InsertCollectorDomain(ctx, navsqlc.InsertCollectorDomainParams{ID: id, SiteID: &req.SiteID, Name: req.Name, Proxy: req.Proxy, Prefix: req.Prefix, Tls: req.TLS})
		if err == nil {
			_, err = q.OpenTargetTrackingPeriod(ctx, navsqlc.OpenTargetTrackingPeriodParams{
				CollectorDomainID: &id, SiteID: req.SiteID,
				Target: collectorTarget(req.Name, req.Prefix), OpenedReason: "created",
			})
		}
		if err == nil {
			_, err = q.AssignPrimaryTargetIfMissing(ctx, req.SiteID)
		}
		if err == nil {
			err = q.RefreshCurrentSiteDaily(ctx, req.SiteID)
		}
		result = collectorDomainModel(row)
		return id, nil, row, err
	})
	return result, err
}

func (store *navStore) updateCollectorDomain(ctx context.Context, meta audit.Meta, id int64, req models.CollectorDomainPayload) common.Error {
	return store.mutate(ctx, meta, "update", "gfn_collector_domain", func(q *navsqlc.Queries) (int64, any, any, error) {
		if err := validateCollectorDomainQuery(ctx, q, req); err != nil {
			return id, nil, nil, err
		}
		before, err := q.LockCollectorDomainForUpdate(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		if before.Deleted {
			return id, before, nil, pgx.ErrNoRows
		}
		if err := validateCollectorTargetIdentity(ctx, q, req, &id); err != nil {
			return id, before, nil, err
		}
		oldSiteID := pointerInt64(before.SiteID)
		identityChanged := oldSiteID != req.SiteID || before.Name != req.Name || !sameOptionalString(before.Prefix, req.Prefix)
		wasPrimary := false
		var oldPeriod navsqlc.GfnTargetTrackingPeriod
		if identityChanged {
			oldPeriod, err = q.GetActiveTargetPeriodByDomain(ctx, &id)
			if err != nil {
				return id, before, nil, err
			}
			primary, primaryErr := q.GetActivePrimaryTarget(ctx, oldSiteID)
			if primaryErr == nil && primary.TargetTrackingPeriodID == oldPeriod.ID {
				wasPrimary = true
				if _, err = q.ClosePrimaryTargetBySite(ctx, oldSiteID); err != nil {
					return id, before, nil, err
				}
			} else if primaryErr != nil && !errors.Is(primaryErr, pgx.ErrNoRows) {
				return id, before, nil, primaryErr
			}
			closedReason := "identity_changed"
			if _, err = q.CloseTargetTrackingPeriod(ctx, navsqlc.CloseTargetTrackingPeriodParams{ID: oldPeriod.ID, ClosedReason: &closedReason}); err != nil {
				return id, before, nil, err
			}
		}
		after, err := q.UpdateCollectorDomain(ctx, navsqlc.UpdateCollectorDomainParams{ID: id, SiteID: &req.SiteID, Name: req.Name, Proxy: req.Proxy, Prefix: req.Prefix, Tls: req.TLS})
		if err == nil && identityChanged {
			var opened navsqlc.GfnTargetTrackingPeriod
			opened, err = q.OpenTargetTrackingPeriod(ctx, navsqlc.OpenTargetTrackingPeriodParams{
				CollectorDomainID: &id, SiteID: req.SiteID,
				Target: collectorTarget(req.Name, req.Prefix), OpenedReason: "identity_changed",
			})
			if err == nil && wasPrimary && oldSiteID == req.SiteID {
				_, err = q.OpenPrimaryTarget(ctx, navsqlc.OpenPrimaryTargetParams{
					SiteID: req.SiteID, TargetTrackingPeriodID: opened.ID, Basis: "explicit",
				})
			}
		}
		if err == nil && identityChanged {
			_, err = q.AssignPrimaryTargetIfMissing(ctx, oldSiteID)
		}
		if err == nil && identityChanged && oldSiteID != req.SiteID {
			_, err = q.AssignPrimaryTargetIfMissing(ctx, req.SiteID)
		}
		if err == nil {
			err = q.RefreshCurrentSiteDaily(ctx, oldSiteID)
		}
		if err == nil && oldSiteID != req.SiteID {
			err = q.RefreshCurrentSiteDaily(ctx, req.SiteID)
		}
		return id, before, after, err
	})
}

func (store *navStore) deleteCollectorDomain(ctx context.Context, meta audit.Meta, id int64) common.Error {
	return store.mutate(ctx, meta, "delete", "gfn_collector_domain", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, err := q.LockCollectorDomainForUpdate(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		if before.Deleted {
			return id, before, nil, pgx.ErrNoRows
		}
		period, err := q.GetActiveTargetPeriodByDomain(ctx, &id)
		if err != nil {
			return id, before, nil, err
		}
		siteID := pointerInt64(before.SiteID)
		primary, primaryErr := q.GetActivePrimaryTarget(ctx, siteID)
		if primaryErr == nil && primary.TargetTrackingPeriodID == period.ID {
			if _, err = q.ClosePrimaryTargetBySite(ctx, siteID); err != nil {
				return id, before, nil, err
			}
		} else if primaryErr != nil && !errors.Is(primaryErr, pgx.ErrNoRows) {
			return id, before, nil, primaryErr
		}
		closedReason := "deleted"
		if _, err = q.CloseTargetTrackingPeriod(ctx, navsqlc.CloseTargetTrackingPeriodParams{ID: period.ID, ClosedReason: &closedReason}); err != nil {
			return id, before, nil, err
		}
		after, err := q.SoftDeleteCollectorDomain(ctx, id)
		if err == nil {
			_, err = q.AssignPrimaryTargetIfMissing(ctx, siteID)
		}
		if err == nil {
			err = q.RefreshCurrentSiteDaily(ctx, siteID)
		}
		return id, before, after, err
	})
}

func (store *navStore) setPrimaryCollectorDomain(ctx context.Context, meta audit.Meta, id int64) common.Error {
	return store.mutate(ctx, meta, "set_primary", "gfn_site_primary_target_periods", func(q *navsqlc.Queries) (int64, any, any, error) {
		domain, err := q.LockCollectorDomainForUpdate(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		if domain.Deleted || domain.SiteID == nil {
			return id, domain, nil, pgx.ErrNoRows
		}
		period, err := q.GetActiveTargetPeriodByDomain(ctx, &id)
		if err != nil {
			return id, domain, nil, err
		}
		var before any
		current, currentErr := q.GetActivePrimaryTarget(ctx, *domain.SiteID)
		if currentErr == nil {
			before = current
			if current.TargetTrackingPeriodID == period.ID {
				return id, before, current, nil
			}
			if _, err = q.ClosePrimaryTargetBySite(ctx, *domain.SiteID); err != nil {
				return id, before, nil, err
			}
		} else if !errors.Is(currentErr, pgx.ErrNoRows) {
			return id, nil, nil, currentErr
		}
		after, err := q.OpenPrimaryTarget(ctx, navsqlc.OpenPrimaryTargetParams{
			SiteID: *domain.SiteID, TargetTrackingPeriodID: period.ID, Basis: "explicit",
		})
		if err == nil {
			err = q.RefreshCurrentSiteDaily(ctx, *domain.SiteID)
		}
		return id, before, after, err
	})
}

func (store *navStore) listSites(ctx context.Context, page adminutil.PageQuery) (int64, []models.SiteDTO, common.Error) {
	total, err := store.q.CountSites(ctx, page.Keyword)
	if err != nil {
		return 0, nil, navDAOError(err)
	}
	limit, offset := pageArgs(page)
	rows, err := store.q.ListSites(ctx, navsqlc.ListSitesParams{Keyword: page.Keyword, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return 0, nil, navDAOError(err)
	}
	items := make([]models.SiteDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, siteDTO(siteModel(row)))
	}
	return total, items, nil
}

func (store *navStore) getSite(ctx context.Context, id int64) (models.SiteDTO, common.Error) {
	row, err := store.q.GetSite(ctx, id)
	return siteDTO(siteModel(row)), navDAOError(err)
}

func (store *navStore) createSite(ctx context.Context, meta audit.Meta, req models.SitePayload) (models.SiteDTO, common.Error) {
	var result models.SiteDTO
	err := store.mutate(ctx, meta, "create", "gfn_site", func(q *navsqlc.Queries) (int64, any, any, error) {
		id, err := q.NextSiteID(ctx)
		if err != nil {
			return 0, nil, nil, err
		}
		row, err := q.InsertSite(ctx, navsqlc.InsertSiteParams{ID: id, Name: req.Name, NameEn: req.NameEn, Info: req.Info, InfoEn: req.InfoEn, Country: req.Country, Nsfw: req.Nsfw, Welfare: req.Welfare, Icon: req.Icon})
		if err == nil {
			err = q.RefreshCurrentSiteDaily(ctx, id)
		}
		result = siteDTO(siteModel(row))
		return id, nil, row, err
	})
	return result, err
}

func (store *navStore) updateSite(ctx context.Context, meta audit.Meta, id int64, req models.SitePayload) common.Error {
	return store.mutate(ctx, meta, "update", "gfn_site", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, err := q.GetSiteAny(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		after, err := q.UpdateSite(ctx, navsqlc.UpdateSiteParams{ID: id, Name: req.Name, NameEn: req.NameEn, Info: req.Info, InfoEn: req.InfoEn, Country: req.Country, Nsfw: req.Nsfw, Welfare: req.Welfare, Icon: req.Icon})
		if err == nil {
			err = q.RefreshCurrentSiteDaily(ctx, id)
		}
		return id, before, after, err
	})
}

func (store *navStore) deleteSite(ctx context.Context, meta audit.Meta, id int64) common.Error {
	return store.mutate(ctx, meta, "delete", "gfn_site", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, err := q.LockSiteForUpdate(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		if before.Deleted {
			return id, before, nil, pgx.ErrNoRows
		}
		if _, err = q.ClosePrimaryTargetBySite(ctx, id); err != nil {
			return id, before, nil, err
		}
		closedReason := "site_deleted"
		if _, err = q.CloseTargetTrackingPeriodsBySite(ctx, navsqlc.CloseTargetTrackingPeriodsBySiteParams{SiteID: id, ClosedReason: &closedReason}); err != nil {
			return id, before, nil, err
		}
		after, err := q.SoftDeleteSite(ctx, id)
		if err == nil {
			err = q.RefreshCurrentSiteDaily(ctx, id)
		}
		return id, before, after, err
	})
}

func (store *navStore) listSiteGroups(ctx context.Context, page adminutil.PageQuery) (int64, []models.SiteGroup, common.Error) {
	total, err := store.q.CountSiteGroups(ctx, page.Keyword)
	if err != nil {
		return 0, nil, navDAOError(err)
	}
	limit, offset := pageArgs(page)
	rows, err := store.q.ListSiteGroups(ctx, navsqlc.ListSiteGroupsParams{Keyword: page.Keyword, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return 0, nil, navDAOError(err)
	}
	items := make([]models.SiteGroup, 0, len(rows))
	for _, row := range rows {
		items = append(items, siteGroupModel(row))
	}
	return total, items, nil
}

func (store *navStore) getSiteGroup(ctx context.Context, id int64) (models.SiteGroup, common.Error) {
	row, err := store.q.GetSiteGroup(ctx, id)
	return siteGroupModel(row), navDAOError(err)
}

func (store *navStore) createSiteGroup(ctx context.Context, meta audit.Meta, req models.SiteGroupPayload) (models.SiteGroup, common.Error) {
	var result models.SiteGroup
	err := store.mutate(ctx, meta, "create", "gfn_site_group", func(q *navsqlc.Queries) (int64, any, any, error) {
		id, err := q.NextSiteGroupID(ctx)
		if err != nil {
			return 0, nil, nil, err
		}
		row, err := q.InsertSiteGroup(ctx, navsqlc.InsertSiteGroupParams{ID: id, Name: req.Name, NameEn: req.NameEn, Info: req.Info, InfoEn: req.InfoEn, Priority: req.Priority})
		result = siteGroupModel(row)
		return id, nil, row, err
	})
	return result, err
}

func (store *navStore) updateSiteGroup(ctx context.Context, meta audit.Meta, id int64, req models.SiteGroupPayload) common.Error {
	return store.mutate(ctx, meta, "update", "gfn_site_group", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, err := q.GetSiteGroup(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		after, err := q.UpdateSiteGroup(ctx, navsqlc.UpdateSiteGroupParams{ID: id, Name: req.Name, NameEn: req.NameEn, Info: req.Info, InfoEn: req.InfoEn, Priority: req.Priority})
		return id, before, after, err
	})
}

func (store *navStore) deleteSiteGroup(ctx context.Context, meta audit.Meta, id int64) common.Error {
	return store.mutate(ctx, meta, "delete", "gfn_site_group", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, err := q.GetSiteGroup(ctx, id)
		if err == nil {
			_, err = q.DeleteSiteGroup(ctx, id)
		}
		return id, before, nil, err
	})
}

func (store *navStore) listSiteGroupMaps(ctx context.Context, page adminutil.PageQuery) (int64, []models.SiteGroupMapDTO, common.Error) {
	total, err := store.q.CountSiteGroupMaps(ctx, page.Keyword)
	if err != nil {
		return 0, nil, navDAOError(err)
	}
	limit, offset := pageArgs(page)
	rows, err := store.q.ListSiteGroupMaps(ctx, navsqlc.ListSiteGroupMapsParams{Keyword: page.Keyword, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return 0, nil, navDAOError(err)
	}
	items := make([]models.SiteGroupMapDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, models.SiteGroupMapDTO{ID: row.ID, SiteID: row.SiteID, GroupID: row.GroupID, SiteName: row.SiteName, GroupName: row.GroupName, Weight: row.Weight, CreateTime: row.CreateTime.Time, UpdateTime: row.UpdateTime.Time})
	}
	return total, items, nil
}

func (store *navStore) getSiteGroupMap(ctx context.Context, id int64) (models.SiteGroupMap, common.Error) {
	row, err := store.q.GetSiteGroupMap(ctx, id)
	return siteGroupMapModel(row), navDAOError(err)
}

func (store *navStore) createSiteGroupMap(ctx context.Context, meta audit.Meta, req models.SiteGroupMapPayload) (models.SiteGroupMap, common.Error) {
	var result models.SiteGroupMap
	err := store.mutate(ctx, meta, "create", "gfn_site_group_map", func(q *navsqlc.Queries) (int64, any, any, error) {
		id, err := q.NextSiteGroupMapID(ctx)
		if err != nil {
			return 0, nil, nil, err
		}
		row, err := q.InsertSiteGroupMap(ctx, navsqlc.InsertSiteGroupMapParams{ID: id, SiteID: req.SiteID, GroupID: req.GroupID, Weight: req.Weight})
		if err == nil {
			err = q.RefreshCurrentSiteDaily(ctx, req.SiteID)
		}
		result = siteGroupMapModel(row)
		return id, nil, row, err
	})
	return result, err
}

func (store *navStore) updateSiteGroupMap(ctx context.Context, meta audit.Meta, id int64, req models.SiteGroupMapPayload) common.Error {
	return store.mutate(ctx, meta, "update", "gfn_site_group_map", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, err := q.GetSiteGroupMap(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		after, err := q.UpdateSiteGroupMap(ctx, navsqlc.UpdateSiteGroupMapParams{ID: id, SiteID: req.SiteID, GroupID: req.GroupID, Weight: req.Weight})
		if err == nil {
			err = q.RefreshCurrentSiteDaily(ctx, before.SiteID)
		}
		if err == nil && before.SiteID != after.SiteID {
			err = q.RefreshCurrentSiteDaily(ctx, after.SiteID)
		}
		return id, before, after, err
	})
}

func (store *navStore) deleteSiteGroupMap(ctx context.Context, meta audit.Meta, id int64) common.Error {
	return store.mutate(ctx, meta, "delete", "gfn_site_group_map", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, err := q.GetSiteGroupMap(ctx, id)
		if err == nil {
			_, err = q.DeleteSiteGroupMap(ctx, id)
		}
		if err == nil {
			err = q.RefreshCurrentSiteDaily(ctx, before.SiteID)
		}
		return id, before, nil, err
	})
}

func (store *navStore) replaceSiteGroupMaps(ctx context.Context, meta audit.Meta, siteID int64, groupIDs []int64) common.Error {
	return store.mutate(ctx, meta, "bulk_replace", "gfn_site_group_map", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, err := q.ListSiteGroupMapsBySite(ctx, siteID)
		if err != nil {
			return siteID, nil, nil, err
		}
		weights := make(map[int64]int64, len(before))
		for _, row := range before {
			weights[row.GroupID] = row.Weight
		}
		if _, err = q.DeleteSiteGroupMapsBySite(ctx, siteID); err != nil {
			return siteID, before, nil, err
		}
		if len(groupIDs) > 0 {
			next, nextErr := q.NextSiteGroupMapID(ctx)
			if nextErr != nil {
				return siteID, before, nil, nextErr
			}
			for index, groupID := range groupIDs {
				if _, err = q.InsertSiteGroupMap(ctx, navsqlc.InsertSiteGroupMapParams{ID: next + int64(index), SiteID: siteID, GroupID: groupID, Weight: weights[groupID]}); err != nil {
					return siteID, before, nil, err
				}
			}
		}
		after, err := q.ListSiteGroupMapsBySite(ctx, siteID)
		if err == nil {
			err = q.RefreshCurrentSiteDaily(ctx, siteID)
		}
		return siteID, before, after, err
	})
}

func (store *navStore) listFeaturedSites(ctx context.Context, page adminutil.PageQuery) (int64, []models.FeaturedSiteDTO, common.Error) {
	total, err := store.q.CountFeaturedSites(ctx, page.Keyword)
	if err != nil {
		return 0, nil, navDAOError(err)
	}
	limit, offset := pageArgs(page)
	rows, err := store.q.ListFeaturedSites(ctx, navsqlc.ListFeaturedSitesParams{Keyword: page.Keyword, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return 0, nil, navDAOError(err)
	}
	items := make([]models.FeaturedSiteDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, featuredSiteDTO(row.ID, row.SiteID, row.SiteName, row.Weight, row.CreateTime.Time, row.UpdateTime.Time))
	}
	return total, items, nil
}

func (store *navStore) getFeaturedSite(ctx context.Context, id int64) (models.FeaturedSiteDTO, common.Error) {
	row, err := store.q.GetFeaturedSite(ctx, id)
	return featuredSiteDTO(row.ID, row.SiteID, row.SiteName, row.Weight, row.CreateTime.Time, row.UpdateTime.Time), navDAOError(err)
}

func validateFeaturedQuery(ctx context.Context, q *navsqlc.Queries, req models.FeaturedSitePayload, currentID int64) error {
	sites, err := q.CountActiveSiteByID(ctx, req.SiteID)
	if err != nil {
		return err
	}
	if sites == 0 {
		return common.NewValidationError("site_id must reference an existing site")
	}
	existing, err := q.CountFeaturedSiteBySite(ctx, navsqlc.CountFeaturedSiteBySiteParams{SiteID: req.SiteID, ExcludeID: currentID})
	if err != nil {
		return err
	}
	if existing > 0 {
		return common.NewValidationError("site_id is already featured")
	}
	return nil
}

func (store *navStore) createFeaturedSite(ctx context.Context, meta audit.Meta, req models.FeaturedSitePayload) (models.FeaturedSite, common.Error) {
	var result models.FeaturedSite
	err := store.mutate(ctx, meta, "create", "gfn_featured_site", func(q *navsqlc.Queries) (int64, any, any, error) {
		if err := validateFeaturedQuery(ctx, q, req, 0); err != nil {
			return 0, nil, nil, err
		}
		id, err := q.NextFeaturedSiteID(ctx)
		if err != nil {
			return 0, nil, nil, err
		}
		row, err := q.InsertFeaturedSite(ctx, navsqlc.InsertFeaturedSiteParams{ID: id, SiteID: req.SiteID, Weight: req.Weight})
		result = featuredSiteModel(row)
		return id, nil, row, err
	})
	return result, err
}

func (store *navStore) updateFeaturedSite(ctx context.Context, meta audit.Meta, id int64, req models.FeaturedSitePayload) common.Error {
	return store.mutate(ctx, meta, "update", "gfn_featured_site", func(q *navsqlc.Queries) (int64, any, any, error) {
		if err := validateFeaturedQuery(ctx, q, req, id); err != nil {
			return id, nil, nil, err
		}
		before, err := q.GetFeaturedSiteAny(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		after, err := q.UpdateFeaturedSite(ctx, navsqlc.UpdateFeaturedSiteParams{ID: id, SiteID: req.SiteID, Weight: req.Weight})
		return id, before, after, err
	})
}

func (store *navStore) deleteFeaturedSite(ctx context.Context, meta audit.Meta, id int64) common.Error {
	return store.mutate(ctx, meta, "delete", "gfn_featured_site", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, err := q.GetFeaturedSiteAny(ctx, id)
		if err == nil {
			_, err = q.DeleteFeaturedSite(ctx, id)
		}
		return id, before, nil, err
	})
}

func navDAOError(err error) common.Error {
	if err == nil {
		return nil
	}
	if appErr, ok := err.(common.Error); ok {
		return appErr
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return common.NewDaoError("record not found")
	}
	return common.NewDaoError(err.Error())
}

func navTimestamp(value time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: value, Valid: !value.IsZero()}
}

func pointerInt64(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func sameOptionalString(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func sayingModel(row navsqlc.GfnSaying) models.Saying {
	return models.Saying{ID: row.ID, Author: row.Author, Saying: row.Saying, Language: row.Language, CreateTime: row.CreateTime.Time, UpdateTime: row.UpdateTime.Time}
}

func updateNoticeModel(row navsqlc.GfnNavUpdateNotice) models.UpdateNotice {
	return models.UpdateNotice{ID: row.ID, Title: row.Title, TitleEn: row.TitleEn, Body: row.Body, BodyEn: row.BodyEn,
		PublishedAt: pkgmodels.LocalTime(row.PublishedAt.Time), CreateTime: pkgmodels.LocalTime(row.CreateTime.Time), UpdateTime: pkgmodels.LocalTime(row.UpdateTime.Time), Deleted: row.Deleted}
}

func collectorDomainModel(row navsqlc.GfnCollectorDomain) models.CollectorDomain {
	return models.CollectorDomain{ID: row.ID, SiteID: pointerInt64(row.SiteID), Name: row.Name, Proxy: row.Proxy, Prefix: row.Prefix, TLS: row.Tls, Deleted: row.Deleted}
}

func collectorDomainDTO(id, siteID int64, siteName, name, proxy string, prefix *string, tls string, deleted bool) models.CollectorDomainDTO {
	return models.CollectorDomainDTO{ID: id, SiteID: siteID, SiteName: siteName, Name: name, Proxy: proxy, Prefix: prefix, TLS: tls, Deleted: deleted}
}

func siteModel(row navsqlc.GfnSite) models.Site {
	return models.Site{ID: row.ID, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn,
		CreateTime: pkgmodels.LocalTime(row.CreateTime.Time), UpdateTime: pkgmodels.LocalTime(row.UpdateTime.Time),
		Country: row.Country, Nsfw: row.Nsfw, Welfare: row.Welfare, Icon: row.Icon, Deleted: row.Deleted}
}

func siteGroupModel(row navsqlc.GfnSiteGroup) models.SiteGroup {
	return models.SiteGroup{ID: row.ID, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, Priority: row.Priority, CreateTime: row.CreateTime.Time, UpdateTime: row.UpdateTime.Time}
}

func siteGroupMapModel(row navsqlc.GfnSiteGroupMap) models.SiteGroupMap {
	return models.SiteGroupMap{ID: row.ID, SiteID: row.SiteID, GroupID: row.GroupID, Weight: row.Weight, CreateTime: row.CreateTime.Time, UpdateTime: row.UpdateTime.Time}
}

func featuredSiteModel(row navsqlc.GfnFeaturedSite) models.FeaturedSite {
	return models.FeaturedSite{ID: row.ID, SiteID: row.SiteID, Weight: row.Weight, CreateTime: row.CreateTime.Time, UpdateTime: row.UpdateTime.Time}
}

func featuredSiteDTO(id, siteID int64, siteName string, weight int64, createdAt, updatedAt time.Time) models.FeaturedSiteDTO {
	return models.FeaturedSiteDTO{ID: id, SiteID: siteID, SiteName: siteName, Weight: weight, CreateTime: createdAt, UpdateTime: updatedAt}
}
