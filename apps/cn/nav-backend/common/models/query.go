package models

/*
 * @Desc: 查询
 * @author: 福狼
 * @version: v1.0.0
 */

type PageReq struct {
	PageNum  int `json:"pageNum"`
	PageSize int `json:"pageSize"`
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

type TimeRange struct {
	BeginTime LocalTime `json:"beginTime"`
	EndTime   LocalTime `json:"endTime"`
}

type CountVo struct {
	Count int64 `form:"count" json:"count"`
}

type Delete struct {
	ID []int64 `form:"id" json:"id"`
}

type Mapping struct {
	MappingId   int64  `form:"mappingId" json:"mappingId"`
	MappingType string `form:"mappingType" json:"mappingType"`
}
