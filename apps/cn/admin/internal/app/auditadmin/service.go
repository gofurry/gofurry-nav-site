package auditadmin

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	sharedaudit "github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	adminsqlc "github.com/gofurry/gofurry-admin/internal/db/admin/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct{ queries *adminsqlc.Queries }

func New(pool *pgxpool.Pool) *Service { return &Service{queries: adminsqlc.New(pool)} }

func (service *Service) List(ctx context.Context, filter Filters) (Page, common.Error) {
	filter = normalizeFilters(filter)
	params := adminsqlc.CountAdminAuditLogsParams{
		OperatorQuery: filter.Operator, OperatorRole: filter.Role, Action: filter.Action,
		Resource: filter.Resource, FromTime: timestamp(filter.From), UntilTime: timestamp(filter.Until),
	}
	total, err := service.queries.CountAdminAuditLogs(ctx, params)
	if err != nil {
		return Page{}, common.NewServiceError("failed to query audit history")
	}
	rows, err := service.queries.ListAdminAuditLogs(ctx, adminsqlc.ListAdminAuditLogsParams{
		OperatorQuery: params.OperatorQuery, OperatorRole: params.OperatorRole, Action: params.Action,
		Resource: params.Resource, FromTime: params.FromTime, UntilTime: params.UntilTime,
		RowLimit: filter.PageSize, RowOffset: (filter.Page - 1) * filter.PageSize,
	})
	if err != nil {
		return Page{}, common.NewServiceError("failed to query audit history")
	}
	page := Page{Total: total, List: make([]sharedaudit.AdminAuditLog, 0, len(rows))}
	for _, row := range rows {
		page.List = append(page.List, sharedaudit.AdminAuditLog{
			ID: row.ID, Action: row.Action, Resource: row.Resource, TargetID: stringValue(row.TargetID),
			Operator: row.Operator, SessionVersion: row.SessionVersion, RequestID: stringValue(row.RequestID),
			IPAddress: stringValue(row.IpAddress), UserAgent: stringValue(row.UserAgent),
			BeforeData: redactSnapshot(row.BeforeData), AfterData: redactSnapshot(row.AfterData),
			OperatorAccountID: row.OperatorAccountID, OperatorName: row.OperatorName,
			OperatorRole: row.OperatorRole, CreatedAt: row.CreatedAt.Time,
		})
	}
	return page, nil
}

func normalizeFilters(filter Filters) Filters {
	filter.Operator = strings.TrimSpace(filter.Operator)
	filter.Role = strings.TrimSpace(filter.Role)
	filter.Action = strings.TrimSpace(filter.Action)
	filter.Resource = strings.TrimSpace(filter.Resource)
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 200 {
		filter.PageSize = 20
	}
	return filter
}

func timestamp(value *time.Time) pgtype.Timestamp {
	if value == nil {
		return pgtype.Timestamp{}
	}
	return pgtype.Timestamp{Time: *value, Valid: true}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func redactSnapshot(raw *string) string {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return "{}"
	}
	var value any
	if json.Unmarshal([]byte(*raw), &value) != nil {
		return "{}"
	}
	redactValue(value)
	encoded, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func redactValue(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if sensitiveKey(key) {
				typed[key] = "[REDACTED]"
				continue
			}
			redactValue(child)
		}
	case []any:
		for _, child := range typed {
			redactValue(child)
		}
	}
}

func sensitiveKey(key string) bool {
	normalized := strings.NewReplacer("-", "", "_", "", ".", "").Replace(strings.ToLower(key))
	for _, fragment := range []string{"password", "passwd", "hash", "token", "jwt", "cookie", "secret", "dsn", "connectionstring"} {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}
