package service

import (
	"cmp"
	"context"
	"crypto/md5"
	"errors"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/common/log"
)

const insightVisualQueryTimeout = 500 * time.Millisecond

var insightHeaderPath = regexp.MustCompile(`^/(steam/apps/([1-9][0-9]*)/|store_item_assets/steam/apps/([1-9][0-9]*)/([0-9a-f]{40}/)?)header(_alt_assets_[0-9]{1,4})?(_schinese)?(_2x)?[.]jpg$`)
var insightHeaderQuery = regexp.MustCompile(`^t=[0-9]{1,20}$`)

func (s *InsightsService) featuredVisuals(ctx context.Context, now time.Time) []v2models.InsightFeaturedVisual {
	day := now.UTC().Format(time.DateOnly)
	readCtx, cancel := context.WithTimeout(ctx, insightVisualQueryTimeout)
	defer cancel()
	rows, err := s.store.ListInsightVisualCandidates(readCtx, day)
	if err != nil || readCtx.Err() != nil {
		reason := "read_failed"
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(readCtx.Err(), context.DeadlineExceeded) {
			reason = "timeout"
		}
		if errors.Is(err, context.Canceled) || errors.Is(readCtx.Err(), context.Canceled) {
			reason = "canceled"
		}
		// Never log the raw driver error, SQL parameters, names, URLs or credentials.
		log.Warn("[insights] featured visuals unavailable", "reason", reason)
		return []v2models.InsightFeaturedVisual{}
	}
	return selectInsightVisuals(rows, day)
}

// This validates provenance/identity only, never pixel content or human approval.
// Keep the finite grammar in ListGameInsightVisualCandidates aligned; integration
// tests exercise SQL and this parser with the same accepted/rejected evidence.
func insightSteamHeader(raw string, appID int64) bool {
	if appID <= 0 || strings.ContainsAny(raw, "%\\") || strings.IndexFunc(raw, func(r rune) bool { return r <= 32 || r == 127 }) >= 0 {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Opaque != "" || u.Fragment != "" || u.RawFragment != "" || u.ForceQuery || u.String() != raw {
		return false
	}
	switch u.Host {
	case "shared.steamstatic.com", "shared.akamai.steamstatic.com", "cdn.akamai.steamstatic.com", "cdn.cloudflare.steamstatic.com", "cdn.steamstatic.com", "steamcdn-a.akamaihd.net":
	default:
		return false
	}
	if u.RawQuery != "" && !insightHeaderQuery.MatchString(u.RawQuery) {
		return false
	}
	parts := insightHeaderPath.FindStringSubmatch(u.Path)
	return len(parts) > 3 && (parts[2] == strconv.FormatInt(appID, 10) || parts[3] == strconv.FormatInt(appID, 10))
}

func selectInsightVisuals(rows []v2models.InsightVisualCandidateRecord, day string) []v2models.InsightFeaturedVisual {
	eligible := make([]v2models.InsightVisualCandidateRecord, 0, len(rows))
	for _, row := range rows {
		if row.GameID <= 0 || row.AppID <= 0 || !row.HasClassification || row.HasAdult || row.AssetGameID != row.GameID || row.AssetAppID != row.AppID || row.AssetID <= 0 || (row.Exists != nil && !*row.Exists) {
			continue
		}
		if row.AssetType != "header" && row.AssetType != "header_2x" {
			continue
		}
		if row.Lang != "zh" && row.Lang != "en" && row.Lang != "" {
			continue
		}
		row.Name, row.NameEn = strings.TrimSpace(row.Name), strings.TrimSpace(row.NameEn)
		if row.Name == "" {
			row.Name = row.NameEn
		}
		if row.Name == "" || !insightSteamHeader(row.Asset, row.AppID) {
			continue
		}
		eligible = append(eligible, row)
	}
	// MD5 is a non-security daily ordering key, matching PostgreSQL's query order.
	// Asset tie-breaks also make unordered/duplicate inputs deterministic.
	rank := func(id int64) string {
		sum := md5.Sum([]byte(day + ":" + strconv.FormatInt(id, 10)))
		return string(sum[:])
	}
	langRank := func(lang string) int {
		if lang == "zh" {
			return 0
		}
		if lang == "en" {
			return 1
		}
		return 2
	}
	slices.SortFunc(eligible, func(a, b v2models.InsightVisualCandidateRecord) int {
		for _, order := range []int{cmp.Compare(rank(a.GameID), rank(b.GameID)), cmp.Compare(a.GameID, b.GameID), cmp.Compare(a.AssetType, b.AssetType), cmp.Compare(langRank(a.Lang), langRank(b.Lang)), cmp.Compare(a.AssetFamily, b.AssetFamily), cmp.Compare(a.SortOrder, b.SortOrder), cmp.Compare(a.AssetID, b.AssetID), cmp.Compare(a.Asset, b.Asset), cmp.Compare(a.Name, b.Name), cmp.Compare(a.NameEn, b.NameEn)} {
			if order != 0 {
				return order
			}
		}
		return 0
	})
	result := make([]v2models.InsightFeaturedVisual, 0, 3)
	seen := make(map[int64]bool, 3)
	for _, row := range eligible {
		if seen[row.GameID] {
			continue
		}
		seen[row.GameID] = true
		result = append(result, v2models.InsightFeaturedVisual{GameID: row.GameID, Name: row.Name, NameEn: row.NameEn, Visual: v2models.InsightEntityVisual{Kind: "game_header", Asset: row.Asset}})
		if len(result) == 3 {
			break
		}
	}
	return result
}
