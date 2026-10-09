package service

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
	"github.com/gofurry/gofurry-nav-backend/common/log"
)

const overviewVisualTimeout = 500 * time.Millisecond

var siteIconKey = regexp.MustCompile(`^nav/sites/([1-9][0-9]*)/icon/[a-f0-9]{32}([.][a-z0-9]{1,16})?$`)

func overviewVisualIDs(changes []models.Change) []int64 {
	ids := make([]int64, 0, 8)
	seen := map[int64]bool{}
	for _, change := range changes {
		id := change.Entity.ID
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
		if len(ids) == 8 {
			break
		}
	}
	return ids
}

func (s *InsightsService) overviewSiteVisuals(ctx context.Context, changes []models.Change) []models.SiteVisual {
	ids := overviewVisualIDs(changes)
	if len(ids) == 0 {
		return []models.SiteVisual{}
	}
	readCtx, cancel := context.WithTimeout(ctx, overviewVisualTimeout)
	defer cancel()
	rows, err := s.store.ListOverviewSiteVisuals(readCtx, ids)
	if err != nil || readCtx.Err() != nil {
		reason := "read_failed"
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(readCtx.Err(), context.DeadlineExceeded) {
			reason = "timeout"
		}
		if errors.Is(err, context.Canceled) || errors.Is(readCtx.Err(), context.Canceled) {
			reason = "canceled"
		}
		log.Warn("[insights] site visuals unavailable", "reason", reason)
		return []models.SiteVisual{}
	}
	return projectOverviewSiteVisuals(ids, rows)
}

func projectOverviewSiteVisuals(ids []int64, rows []models.SiteVisualRecord) []models.SiteVisual {
	eligible := make(map[int64]models.SiteVisual, len(rows))
	for _, row := range rows {
		// GFN uses literal "0" for explicit non-adult; unknown is never false.
		if row.SiteID <= 0 || row.Deleted || row.NSFW != "0" {
			continue
		}
		key := siteIconKey.FindStringSubmatch(row.Asset)
		if len(key) < 2 || key[1] != strconv.FormatInt(row.SiteID, 10) {
			continue
		}
		name := strings.TrimSpace(row.Name)
		if name == "" {
			name = strings.TrimSpace(row.NameEn)
		}
		if name == "" {
			continue
		}
		eligible[row.SiteID] = models.SiteVisual{SiteID: row.SiteID, Name: name, Visual: models.EntityVisual{Kind: "site_icon", Asset: row.Asset}}
	}
	result := make([]models.SiteVisual, 0, 5)
	seen := map[int64]bool{}
	for _, id := range ids {
		if visual, ok := eligible[id]; ok && !seen[id] {
			result = append(result, visual)
			seen[id] = true
		}
		if len(result) == 5 {
			break
		}
	}
	return result
}
