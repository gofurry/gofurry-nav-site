package workbench

import (
	"context"
	"fmt"
	"sort"

	"github.com/gofurry/gofurry-admin/internal/app/auditadmin"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
	authservice "github.com/gofurry/gofurry-admin/internal/app/auth/service"
	"github.com/gofurry/gofurry-admin/internal/app/changeadmin"
	collectionservice "github.com/gofurry/gofurry-admin/internal/app/collectionadmin/service"
	"github.com/gofurry/gofurry-admin/internal/app/dataops"
	"github.com/gofurry/gofurry-admin/internal/app/metricadmin"
)

type Service struct {
	collection *collectionservice.Service
	metrics    *metricadmin.Service
	changes    *changeadmin.Service
	data       *dataops.Service
	audit      *auditadmin.Service
	accounts   *authservice.AuthService
}

func New(collection *collectionservice.Service, metrics *metricadmin.Service, changes *changeadmin.Service, data *dataops.Service, audit *auditadmin.Service, accounts *authservice.AuthService) *Service {
	return &Service{collection: collection, metrics: metrics, changes: changes, data: data, audit: audit, accounts: accounts}
}

func (service *Service) Summary(ctx context.Context, principal *authorization.Principal) Summary {
	result := Summary{Attention: []AttentionItem{}, RecentChanges: []RecentChange{}, RecentOperations: []RecentOperation{}, SystemStatus: []SystemStatus{}}
	flags := featuresFor(principal)
	if flags.collection {
		service.collectionSummary(ctx, &result)
	}
	if flags.metrics {
		service.metricSummary(ctx, &result, flags.metricTechnical)
	}
	if flags.changes {
		service.changeSummary(ctx, &result, flags.changeTechnical)
	}
	if flags.dataOps {
		service.dataSummary(ctx, &result)
	}
	if flags.audit {
		service.auditSummary(ctx, &result)
	}
	if flags.accounts {
		service.accountSummary(ctx, &result)
	}
	return result
}

type featureFlags struct {
	collection, metrics, metricTechnical, changes, changeTechnical, dataOps, audit, accounts bool
}

func featuresFor(principal *authorization.Principal) featureFlags {
	if principal == nil {
		return featureFlags{}
	}
	return featureFlags{
		collection: principal.Has(authorization.CollectionRead), metrics: principal.Has(authorization.MetricsRead),
		metricTechnical: principal.Has(authorization.MetricsTechnical), changes: principal.Has(authorization.ChangesRead),
		changeTechnical: principal.Has(authorization.ChangesTechnical), dataOps: principal.Has(authorization.DataOpsRead),
		audit: principal.Has(authorization.AuditRead), accounts: principal.Has(authorization.AccountManage),
	}
}

func (service *Service) collectionSummary(ctx context.Context, result *Summary) {
	overview, err := service.collection.Overview(ctx)
	if err != nil {
		result.Attention = append(result.Attention, AttentionItem{Key: "collection-unavailable", Tone: "danger", Title: "采集状态不可用", Summary: "无法读取采集控制面，请检查业务数据库。", Href: "/collection"})
		result.SystemStatus = append(result.SystemStatus, SystemStatus{Key: "collection", Label: "采集控制面", Status: "danger", Summary: "读取失败", Href: "/collection"})
		return
	}
	if overview.Failed24h > 0 {
		result.Attention = append(result.Attention, AttentionItem{Key: "collection-failed", Tone: "danger", Title: "最近 24 小时存在失败采集", Summary: fmt.Sprintf("%d 个采集任务失败", overview.Failed24h), Href: "/collection?tab=history&status=failed"})
	}
	if overview.Missed24h > 0 {
		result.Attention = append(result.Attention, AttentionItem{Key: "collection-missed", Tone: "warning", Title: "最近 24 小时存在错过调度", Summary: fmt.Sprintf("%d 个计划调度未执行", overview.Missed24h), Href: "/collection?tab=schedules"})
	}
	instances, instanceErr := service.collection.Instances(ctx, collectionservice.InstanceFilters{View: "current", Limit: 100})
	degraded := 0
	if instanceErr == nil {
		for _, instance := range instances.List {
			if instance.Health != "healthy" {
				degraded++
			}
		}
	}
	if degraded > 0 {
		result.Attention = append(result.Attention, AttentionItem{Key: "collector-degraded", Tone: "warning", Title: "Collector 需要关注", Summary: fmt.Sprintf("%d 个当前实例离线或心跳异常", degraded), Href: "/collection?tab=overview"})
	}
	status := "success"
	if overview.Failed24h > 0 || overview.Missed24h > 0 || degraded > 0 {
		status = "warning"
	}
	result.SystemStatus = append(result.SystemStatus, SystemStatus{Key: "collection", Label: "采集控制面", Status: status, Summary: fmt.Sprintf("%d 运行中 · %d 排队 · %d Collector", overview.RunningCount, overview.QueuedCount, instances.Total), Href: "/collection"})
}

func (service *Service) metricSummary(ctx context.Context, result *Summary, technical bool) {
	rows, err := service.metrics.Overview(ctx, "")
	if err != nil {
		result.SystemStatus = append(result.SystemStatus, SystemStatus{Key: "metrics", Label: "指标流水线", Status: "danger", Summary: "读取失败", Href: "/metrics"})
		return
	}
	worstLag := 0
	for _, metric := range rows {
		if metric.CoverageRate != nil && *metric.CoverageRate < 0.8 {
			result.Attention = append(result.Attention, AttentionItem{Key: "coverage-" + metric.Domain + "-" + metric.MetricKey, Tone: "warning", Title: metric.Description + "覆盖率偏低", Summary: fmt.Sprintf("当前覆盖率 %.1f%%", *metric.CoverageRate*100), Href: "/metrics"})
		}
		if metric.LagDays != nil && *metric.LagDays > worstLag {
			worstLag = *metric.LagDays
		}
		if technical && metric.LagDays != nil && *metric.LagDays > 0 {
			result.Attention = append(result.Attention, AttentionItem{Key: "metric-lag-" + metric.Domain + "-" + metric.MetricKey, Tone: "warning", Title: metric.Description + "处理滞后", Summary: fmt.Sprintf("落后上游 %d 天", *metric.LagDays), Href: "/metrics?tab=technical"})
		}
	}
	status := "success"
	if worstLag > 0 {
		status = "warning"
	}
	result.SystemStatus = append(result.SystemStatus, SystemStatus{Key: "metrics", Label: "指标流水线", Status: status, Summary: fmt.Sprintf("%d 个活跃指标 · 最大滞后 %d 天", len(rows), worstLag), Href: "/metrics"})
}

func (service *Service) changeSummary(ctx context.Context, result *Summary, technical bool) {
	overview, err := service.changes.Overview(ctx, "")
	if err != nil {
		result.SystemStatus = append(result.SystemStatus, SystemStatus{Key: "changes", Label: "变化流水线", Status: "danger", Summary: "读取失败", Href: "/changes"})
		return
	}
	worstLag := 0
	for _, detector := range overview {
		if detector.LagDays != nil && *detector.LagDays > worstLag {
			worstLag = *detector.LagDays
		}
		if technical && detector.LagDays != nil && *detector.LagDays > 0 {
			result.Attention = append(result.Attention, AttentionItem{Key: "change-lag-" + detector.Domain + "-" + detector.DetectorKey, Tone: "warning", Title: detector.Description + "处理滞后", Summary: fmt.Sprintf("落后上游 %d 天", *detector.LagDays), Href: "/changes?tab=technical"})
		}
	}
	for _, domain := range []string{"game", "nav"} {
		page, pageErr := service.changes.Events(ctx, changeadmin.Filters{Domain: domain, Page: 1, PageSize: 5})
		if pageErr != nil {
			continue
		}
		for _, event := range page.List {
			result.RecentChanges = append(result.RecentChanges, RecentChange{Domain: event.Domain, EventKey: event.EventKey, HistoricalName: event.HistoricalName, EventCode: event.EventCode, EventAt: event.EventAt, ProjectionDate: event.ProjectionDate})
		}
	}
	sort.Slice(result.RecentChanges, func(i, j int) bool {
		return changeSortKey(result.RecentChanges[i]) > changeSortKey(result.RecentChanges[j])
	})
	if len(result.RecentChanges) > 6 {
		result.RecentChanges = result.RecentChanges[:6]
	}
	status := "success"
	if worstLag > 0 {
		status = "warning"
	}
	result.SystemStatus = append(result.SystemStatus, SystemStatus{Key: "changes", Label: "变化流水线", Status: status, Summary: fmt.Sprintf("%d 个活跃检测器 · 最大滞后 %d 天", len(overview), worstLag), Href: "/changes"})
}

func changeSortKey(change RecentChange) string {
	if change.EventAt != nil {
		return *change.EventAt
	}
	return change.ProjectionDate
}

func (service *Service) dataSummary(ctx context.Context, result *Summary) {
	overview := service.data.Overview(ctx)
	for _, database := range overview.Databases {
		status := "success"
		if database.Health == "unavailable" {
			status = "danger"
		} else if database.Health != "healthy" {
			status = "warning"
		}
		result.SystemStatus = append(result.SystemStatus, SystemStatus{Key: "database-" + database.Key, Label: database.Key + " 数据库", Status: status, Summary: database.Migration.Status + " · " + database.DatabaseName, Href: "/system/data-operations"})
		if status != "success" {
			result.Attention = append(result.Attention, AttentionItem{Key: "database-attention-" + database.Key, Tone: status, Title: database.Key + " 数据库需要关注", Summary: "健康或迁移状态不是 current", Href: "/system/data-operations"})
		}
	}
}

func (service *Service) auditSummary(ctx context.Context, result *Summary) {
	page, err := service.audit.List(ctx, auditadmin.Filters{Page: 1, PageSize: 6})
	if err != nil {
		return
	}
	for _, row := range page.List {
		result.RecentOperations = append(result.RecentOperations, RecentOperation{ID: row.ID, Action: row.Action, Resource: row.Resource, TargetID: row.TargetID, OperatorName: row.OperatorName, OperatorRole: row.OperatorRole, CreatedAt: row.CreatedAt})
	}
}

func (service *Service) accountSummary(ctx context.Context, result *Summary) {
	page, err := service.accounts.ListAccounts(ctx, "", 200, 0)
	if err != nil {
		result.SystemStatus = append(result.SystemStatus, SystemStatus{Key: "accounts", Label: "账号治理", Status: "danger", Summary: "读取失败", Href: "/system/accounts"})
		return
	}
	active, disabled := 0, 0
	for _, account := range page.List {
		if account.Status == authorization.StatusActive {
			active++
		} else {
			disabled++
		}
	}
	result.SystemStatus = append(result.SystemStatus, SystemStatus{Key: "accounts", Label: "账号治理", Status: "success", Summary: fmt.Sprintf("%d 启用 · %d 停用", active, disabled), Href: "/system/accounts"})
}
