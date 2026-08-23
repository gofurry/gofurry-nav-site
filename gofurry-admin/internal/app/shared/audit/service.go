package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	authmodels "github.com/gofurry/gofurry-admin/internal/app/auth/models"
	adminsqlc "github.com/gofurry/gofurry-admin/internal/db/admin/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const claimsContextKey = "auth_claims"

type Logger struct {
	queries *adminsqlc.Queries
}

func New(pool *pgxpool.Pool) *Logger {
	return &Logger{queries: adminsqlc.New(pool)}
}

func MetaFromFiber(c fiber.Ctx) Meta {
	meta := Meta{Operator: "admin", RequestID: strings.TrimSpace(c.RequestID()), IPAddress: strings.TrimSpace(c.IP()), UserAgent: strings.TrimSpace(c.UserAgent())}
	if claims, ok := c.Locals(claimsContextKey).(*authmodels.AdminClaims); ok && claims != nil {
		if subject := strings.TrimSpace(claims.Subject); subject != "" {
			meta.Operator = subject
		}
		meta.SessionVersion = claims.SessionVersion
	}
	return meta
}

func SystemMeta(source string) Meta {
	source = strings.TrimSpace(source)
	if source == "" {
		source = "system"
	}
	return Meta{Operator: "admin", RequestID: source, IPAddress: "127.0.0.1", UserAgent: source}
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
	if action == "" {
		return common.NewValidationError("audit action is required")
	}
	if resource == "" {
		return common.NewValidationError("audit resource is required")
	}
	if operator == "" {
		operator = "admin"
	}
	target := stringifyTargetID(targetID)
	requestID, ipAddress, userAgent := strings.TrimSpace(meta.RequestID), strings.TrimSpace(meta.IPAddress), strings.TrimSpace(meta.UserAgent)
	beforeData, afterData := snapshotJSON(before), snapshotJSON(after)
	err := queries.InsertAdminAuditLog(ctx, adminsqlc.InsertAdminAuditLogParams{
		Action: action, Resource: resource, TargetID: &target, Operator: operator,
		SessionVersion: meta.SessionVersion, RequestID: &requestID, IpAddress: &ipAddress,
		UserAgent: &userAgent, BeforeData: &beforeData, AfterData: &afterData,
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
