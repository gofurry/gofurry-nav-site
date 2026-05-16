package audit

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	authmodels "github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/auth/models"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/infra/db"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	"gorm.io/gorm"
)

const claimsContextKey = "auth_claims"

func MetaFromFiber(c fiber.Ctx) Meta {
	meta := Meta{
		Operator:  "admin",
		RequestID: strings.TrimSpace(c.RequestID()),
		IPAddress: strings.TrimSpace(c.IP()),
		UserAgent: strings.TrimSpace(c.UserAgent()),
	}

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
	return Meta{
		Operator:  "admin",
		RequestID: source,
		IPAddress: "127.0.0.1",
		UserAgent: source,
	}
}

func Log(meta Meta, action, resource string, targetID any, before, after any) common.Error {
	return LogTx(nil, meta, action, resource, targetID, before, after)
}

func LogTx(tx *gorm.DB, meta Meta, action, resource string, targetID any, before, after any) common.Error {
	engine := db.Databases.DB(db.Admin)
	if engine == nil {
		return common.NewDaoError("admin database is not initialized")
	}

	entry := AdminAuditLog{
		Action:         strings.TrimSpace(action),
		Resource:       strings.TrimSpace(resource),
		TargetID:       stringifyTargetID(targetID),
		Operator:       strings.TrimSpace(meta.Operator),
		SessionVersion: meta.SessionVersion,
		RequestID:      strings.TrimSpace(meta.RequestID),
		IPAddress:      strings.TrimSpace(meta.IPAddress),
		UserAgent:      strings.TrimSpace(meta.UserAgent),
		BeforeData:     snapshotJSON(before),
		AfterData:      snapshotJSON(after),
	}

	if entry.Action == "" {
		return common.NewValidationError("audit action is required")
	}
	if entry.Resource == "" {
		return common.NewValidationError("audit resource is required")
	}
	if entry.Operator == "" {
		entry.Operator = "admin"
	}

	if err := engine.Create(&entry).Error; err != nil {
		return common.NewDaoError(err.Error())
	}
	return nil
}

func SnapshotByID(tx *gorm.DB, table string, id int64) (map[string]any, common.Error) {
	table = strings.TrimSpace(table)
	if tx == nil {
		return nil, common.NewDaoError("snapshot database transaction is not initialized")
	}
	if table == "" {
		return nil, common.NewValidationError("snapshot table is required")
	}
	if id <= 0 {
		return nil, common.NewValidationError("snapshot id must be greater than 0")
	}

	row := make(map[string]any)
	if err := tx.Table(table).Where("id = ?", id).Take(&row).Error; err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return row, nil
}

func SnapshotRows(tx *gorm.DB, table, condition, order string, args ...any) ([]map[string]any, common.Error) {
	table = strings.TrimSpace(table)
	if tx == nil {
		return nil, common.NewDaoError("snapshot database transaction is not initialized")
	}
	if table == "" {
		return nil, common.NewValidationError("snapshot table is required")
	}

	query := tx.Table(table)
	if condition = strings.TrimSpace(condition); condition != "" {
		query = query.Where(condition, args...)
	}
	if order = strings.TrimSpace(order); order != "" {
		query = query.Order(order)
	}

	rows := make([]map[string]any, 0)
	if err := query.Find(&rows).Error; err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return rows, nil
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
