package adminutil

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	"github.com/gofiber/fiber/v3"
)

type PageQuery struct {
	PageNum  int    `json:"page_num"`
	PageSize int    `json:"page_size"`
	Keyword  string `json:"keyword"`
}

type OptionItem struct {
	ID    int64  `json:"id,string"`
	Label string `json:"label"`
	Extra string `json:"extra,omitempty"`
}

type BulkReplaceRequest struct {
	OwnerID int64   `json:"owner_id"`
	IDs     []int64 `json:"ids"`
}

func DecodeBody(c fiber.Ctx, target any) common.Error {
	if err := json.Unmarshal(c.Body(), target); err != nil {
		return common.NewValidationError("request body must be valid json")
	}
	return nil
}

func ParseIDParam(c fiber.Ctx) (int64, common.Error) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id", "0")), 10, 64)
	if err != nil || id <= 0 {
		return 0, common.NewValidationError("id must be a positive integer")
	}
	return id, nil
}

func ParsePageQuery(c fiber.Ctx) PageQuery {
	pageNum, _ := strconv.Atoi(strings.TrimSpace(c.Query("page_num", "1")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(c.Query("page_size", "20")))
	query := PageQuery{
		PageNum:  pageNum,
		PageSize: pageSize,
		Keyword:  strings.TrimSpace(c.Query("keyword", "")),
	}
	if query.PageNum < 1 {
		query.PageNum = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	if query.PageSize > 200 {
		query.PageSize = 200
	}
	return query
}
