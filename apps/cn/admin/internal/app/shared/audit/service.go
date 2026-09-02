package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
	adminsqlc "github.com/gofurry/gofurry-admin/internal/db/admin/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Logger struct {
	queries *adminsqlc.Queries
}

func New(pool *pgxpool.Pool) *Logger {
	return &Logger{queries: adminsqlc.New(pool)}
}

func MetaFromFiber(c fiber.Ctx) Meta {
	meta := Meta{Operator: "system", OperatorName: "system", OperatorRole: "system", RequestID: strings.TrimSpace(c.RequestID()), IPAddress: strings.TrimSpace(c.IP()), UserAgent: strings.TrimSpace(c.UserAgent())}
	if principal, ok := c.Locals(authorization.PrincipalContextKey).(*authorization.Principal); ok && principal != nil {
		return MetaForPrincipal(meta, principal)
	}
	return meta
}

func MetaForPrincipal(meta Meta, principal *authorization.Principal) Meta {
	if principal == nil {
		return meta
	}
	accountID := principal.AccountID
	meta.Operator = principal.Username
	meta.OperatorAccountID = &accountID
	meta.OperatorName = principal.DisplayName
	meta.OperatorRole = string(principal.Role)
	meta.SessionVersion = principal.SessionVersion
	return meta
}

func SystemMeta(source string) Meta {
	source = strings.TrimSpace(source)
	if source == "" {
		source = "system"
	}
	return Meta{Operator: "system", OperatorName: "system", OperatorRole: "system", RequestID: source, IPAddress: "127.0.0.1", UserAgent: source}
}

func (logger *Logger) Log(ctx context.Context, meta Meta, action, resource string, targetID any, before, after any) common.Error {
	return logger.log(ctx, logger.queries, meta, action, resource, targetID, before, after)
}

func (logger *Logger) LogTx(ctx context.Context, tx pgx.Tx, meta Meta, action, resource string, targetID any, before, after any) common.Error {
	return logger.log(ctx, logger.queries.WithTx(tx), meta, action, resource, targetID, before, after)
}

func (logger *Logger) log(ctx context.Context, queries *adminsqlc.Queries, meta Meta, action, resource string, targetID any, before, after any) common.Error {
	action = strings.TrimSpace(action)
	resource = strings.TrimSpace(resource)
	operator := strings.TrimSpace(meta.Operator)
	operatorName := strings.TrimSpace(meta.OperatorName)
	operatorRole := strings.TrimSpace(meta.OperatorRole)
	if action == "" {
		return common.NewValidationError("audit action is required")
	}
	if resource == "" {
		return common.NewValidationError("audit resource is required")
	}
	if operator == "" {
		operator = "system"
	}
	if operatorName == "" {
		operatorName = operator
	}
	if operatorRole == "" {
		operatorRole = "system"
	}
	target := stringifyTargetID(targetID)
	requestID, ipAddress, userAgent := strings.TrimSpace(meta.RequestID), strings.TrimSpace(meta.IPAddress), strings.TrimSpace(meta.UserAgent)
	beforeData, afterData := snapshotJSON(before), snapshotJSON(after)
	err := queries.InsertAdminAuditLog(ctx, adminsqlc.InsertAdminAuditLogParams{
		Action: action, Resource: resource, TargetID: &target, Operator: operator,
		SessionVersion: meta.SessionVersion, RequestID: &requestID, IpAddress: &ipAddress,
		UserAgent: &userAgent, BeforeData: &beforeData, AfterData: &afterData,
		OperatorAccountID: meta.OperatorAccountID, OperatorName: operatorName, OperatorRole: operatorRole,
	})
	if err != nil {
		return common.NewDaoError(err.Error())
	}
	return nil
}

func stringifyTargetID(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func snapshotJSON(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case []byte:
		return strings.TrimSpace(string(v))
	case json.RawMessage:
		return strings.TrimSpace(string(v))
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(data)
	}
}
