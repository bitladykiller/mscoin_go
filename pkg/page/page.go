// Package page keeps the HTTP-facing pagination response compatible with the
// legacy MSCoin frontend contract.
package page

import "math"

// Result is the public JSON pagination model returned by list endpoints.
type Result struct {
	Content       []any `json:"content"`
	TotalElements int64 `json:"totalElements"`
	Number        int64 `json:"number"`
	TotalPages    int64 `json:"totalPages"`
	HasNext       bool  `json:"hasNext"`
	IsLast        bool  `json:"isLast"`
}

// New constructs a pagination envelope from raw items and total count.
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
