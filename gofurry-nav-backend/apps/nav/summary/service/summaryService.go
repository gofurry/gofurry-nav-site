package service

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/summary/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
)

type summaryService struct{}

type redisGetter func(key string) (string, common.GFError)

var summarySingleton = new(summaryService)

func GetSummaryService() *summaryService { return summarySingleton }

func SiteSummaryKey(siteID int64) string {
	return fmt.Sprintf("collector:v2:summary:site:%d", siteID)
}

func TargetSummaryKey(siteID int64, target string) string {
	return fmt.Sprintf("collector:v2:summary:target:%d:%s", siteID, target)
}

func (svc *summaryService) GetSiteSummary(siteID int64) (models.SiteSummaryResponse, common.GFError) {
	return readSiteSummary(cs.GetString, siteID, env.GetServerConfig().NavV2.SummaryStaleAfter(), time.Now())
}

func (svc *summaryService) GetTargetSummary(siteID int64, target string) (models.TargetSummaryResponse, common.GFError) {
	return readTargetSummary(cs.GetString, siteID, target, env.GetServerConfig().NavV2.SummaryStaleAfter(), time.Now())
}

func readSiteSummary(get redisGetter, siteID int64, staleAfter time.Duration, now time.Time) (models.SiteSummaryResponse, common.GFError) {
	if siteID <= 0 {
		return models.SiteSummaryResponse{}, common.NewServiceError("siteId 参数非法")
	}
	raw, err := get(SiteSummaryKey(siteID))
	if err != nil {
		return models.SiteSummaryResponse{}, err
	}
	if strings.TrimSpace(raw) == "" {
		return missingSiteSummary(siteID), nil
	}

	var summary models.SiteSummaryResponse
	if jsonErr := sonic.UnmarshalString(raw, &summary); jsonErr != nil {
		logSummaryRedisJSONDecodeFailure(SiteSummaryKey(siteID), siteID, "", jsonErr)
		return models.SiteSummaryResponse{}, common.NewServiceError("站点健康摘要解析失败")
	}
	if summary.SiteID == 0 {
		summary.SiteID = siteID
	}
	summary.State = models.SummaryStateReady
	if summaryIsStale(summary.GeneratedAt, staleAfter, now) {
		return staleSiteSummary(summary), nil
	}
	return summary, nil
}

func readTargetSummary(get redisGetter, siteID int64, target string, staleAfter time.Duration, now time.Time) (models.TargetSummaryResponse, common.GFError) {
	target = strings.TrimSpace(target)
	if siteID <= 0 {
		return models.TargetSummaryResponse{}, common.NewServiceError("siteId 参数非法")
	}
	if target == "" {
		return models.TargetSummaryResponse{}, common.NewServiceError("target 参数不能为空")
	}

	raw, err := get(TargetSummaryKey(siteID, target))
	if err != nil {
		return models.TargetSummaryResponse{}, err
	}
	if strings.TrimSpace(raw) == "" {
		return missingTargetSummary(siteID, target), nil
	}

	var summary models.TargetSummaryResponse
	if jsonErr := sonic.UnmarshalString(raw, &summary); jsonErr != nil {
		logSummaryRedisJSONDecodeFailure(TargetSummaryKey(siteID, target), siteID, target, jsonErr)
		return models.TargetSummaryResponse{}, common.NewServiceError("目标健康摘要解析失败")
	}
	if summary.SiteID == 0 {
		summary.SiteID = siteID
	}
	if summary.Target == "" {
		summary.Target = target
	}
	summary.State = models.SummaryStateReady
	if summaryIsStale(summary.GeneratedAt, staleAfter, now) {
		return staleTargetSummary(summary), nil
	}
	return summary, nil
}

func summaryIsStale(generatedAt time.Time, staleAfter time.Duration, now time.Time) bool {
	return generatedAt.IsZero() || now.Sub(generatedAt) > staleAfter
}

func missingSiteSummary(siteID int64) models.SiteSummaryResponse {
	return models.SiteSummaryResponse{
		State:          models.SummaryStateMissing,
		SiteID:         siteID,
		Status:         models.StatusUnknown,
		ReasonCodes:    []string{"summary_missing"},
		ReasonMessages: []string{"站点健康摘要暂不可用"},
		StatusCounts:   emptyStatusCounts(),
	}
}

func staleSiteSummary(summary models.SiteSummaryResponse) models.SiteSummaryResponse {
	summary.State = models.SummaryStateStale
	summary.Status = models.StatusUnknown
	summary.ReasonCodes = []string{"summary_stale"}
	summary.ReasonMessages = []string{"站点健康摘要已过期"}
	return summary
}

func missingTargetSummary(siteID int64, target string) models.TargetSummaryResponse {
	return models.TargetSummaryResponse{
		State:          models.SummaryStateMissing,
		SiteID:         siteID,
		Target:         target,
		Status:         models.StatusUnknown,
		ReasonCodes:    []string{"summary_missing"},
		ReasonMessages: []string{"目标健康摘要暂不可用"},
		Protocols:      map[string]models.ProtocolSummary{},
	}
}

func staleTargetSummary(summary models.TargetSummaryResponse) models.TargetSummaryResponse {
	summary.State = models.SummaryStateStale
	summary.Status = models.StatusUnknown
	summary.ReasonCodes = []string{"summary_stale"}
	summary.ReasonMessages = []string{"目标健康摘要已过期"}
	return summary
}

func emptyStatusCounts() map[string]int {
	return map[string]int{
		models.StatusHealthy:  0,
		models.StatusWarning:  0,
		models.StatusDegraded: 0,
		models.StatusUnknown:  0,
		models.StatusDown:     0,
	}
}

func logSummaryRedisJSONDecodeFailure(key string, siteID int64, target string, err error) {
	slog.Warn(
		"collector v2 summary redis json parse failed",
		"redis_key", key,
		"site_id", siteID,
		"target", target,
		"protocol", "",
		"error", err.Error(),
	)
}
