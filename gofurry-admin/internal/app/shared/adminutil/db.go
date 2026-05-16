package adminutil

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"regexp"
	"strings"

	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/models"
	"gorm.io/gorm"
)

var tableNamePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

func BuildPageResponse[T any](total int64, list []T) models.PageResponse {
	if list == nil {
		list = []T{}
	}
	return models.PageResponse{
		Total: total,
		Data:  list,
	}
}

func ApplyKeyword(query *gorm.DB, keyword string, columns ...string) *gorm.DB {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" || len(columns) == 0 {
		return query
	}

	like := "%" + keyword + "%"
	parts := make([]string, 0, len(columns))
	args := make([]any, 0, len(columns))
	for _, column := range columns {
		parts = append(parts, fmt.Sprintf("%s ILIKE ?", column))
		args = append(args, like)
	}
	return query.Where(strings.Join(parts, " OR "), args...)
}

func Paginate(base *gorm.DB, page PageQuery, out any) (int64, common.Error) {
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return 0, common.NewDaoError(err.Error())
	}
	if total == 0 {
		return 0, nil
	}

	offset := (page.PageNum - 1) * page.PageSize
	if err := base.Offset(offset).Limit(page.PageSize).Find(out).Error; err != nil {
		return 0, common.NewDaoError(err.Error())
	}
	return total, nil
}

func AllocateSequentialIDs(tx *gorm.DB, table string, count int) ([]int64, common.Error) {
	if count <= 0 {
		return nil, common.NewValidationError("count must be greater than 0")
	}
	if !tableNamePattern.MatchString(table) {
		return nil, common.NewServiceError("invalid table name")
	}

	lockKey := int64(crc32.ChecksumIEEE([]byte(table)))
	if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", lockKey).Error; err != nil {
		return nil, common.NewDaoError(err.Error())
	}

	var next int64
	if err := tx.Raw(fmt.Sprintf("SELECT COALESCE(MAX(id), 0) + 1 FROM %s", table)).Scan(&next).Error; err != nil {
		return nil, common.NewDaoError(err.Error())
	}

	result := make([]int64, count)
	for i := range result {
		result[i] = next + int64(i)
	}
	return result, nil
}

func MustJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func ParseStringArray(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	var result []string
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return []string{}
	}
	return result
}

func ParseKVArray(raw *string) []models.KvModel {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return []models.KvModel{}
	}
	var result []models.KvModel
	if err := json.Unmarshal([]byte(strings.TrimSpace(*raw)), &result); err != nil {
		return []models.KvModel{}
	}
	return result
}

func ToJSONStringPtr[T any](value T) *string {
	data, _ := json.Marshal(value)
	result := string(data)
	return &result
}
