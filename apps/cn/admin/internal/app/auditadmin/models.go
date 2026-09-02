package auditadmin

import (
	"time"

	sharedaudit "github.com/gofurry/gofurry-admin/internal/app/shared/audit"
)

type Filters struct {
	Operator string
	Role     string
	Action   string
	Resource string
	From     *time.Time
	Until    *time.Time
	Page     int32
	PageSize int32
}

type Page struct {
	Total int64                       `json:"total"`
	List  []sharedaudit.AdminAuditLog `json:"list"`
}
