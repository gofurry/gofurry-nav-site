package models

/*
 * @Desc: 查询
 * @author: 福狼
 * @version: v1.0.0
 */

type PageReq struct {
	PageNum  int `json:"page_num"`
	PageSize int `json:"page_size"`
}

func (pageReq *PageReq) InitPageIfAbsent() {
	if pageReq.PageNum < 1 {
		pageReq.PageNum = 1
	}
	if pageReq.PageSize < 1 {
		pageReq.PageSize = 10
	}
}

type PageResponse struct {
	Total int64 `json:"total"`
	Data  any   `json:"list"`
}
