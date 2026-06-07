// Package page 保持 HTTP 面向的分页响应与传统 MSCoin 前端契约兼容。
package page

import "math"

// Result 是列表端点返回的公共 JSON 分页模型。
type Result struct {
	Content       []any `json:"content"`
	TotalElements int64 `json:"totalElements"`
	Number        int64 `json:"number"`
	TotalPages    int64 `json:"totalPages"`
	HasNext       bool  `json:"hasNext"`
	IsLast        bool  `json:"isLast"`
}

// New 从原始项目和总数构造分页信封。
func New(content []any, page int64, pageSize int64, total int64) *Result {
	resp := &Result{
		Content:       content,
		TotalElements: total,
		Number:        page,
	}

	if pageSize <= 0 {
		resp.TotalPages = 1
	} else {
		resp.TotalPages = int64(math.Ceil(float64(total) / float64(pageSize)))
	}
	resp.HasNext = resp.Number+1 < resp.TotalPages
	resp.IsLast = !resp.HasNext
	return resp
}
