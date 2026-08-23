package adminutil

import (
	"encoding/json"
	"strings"

	"github.com/gofurry/gofurry-admin/pkg/models"
)

func BuildPageResponse[T any](total int64, list []T) models.PageResponse {
	if list == nil {
		list = []T{}
	}
	return models.PageResponse{
		Total: total,
		Data:  list,
	}
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
