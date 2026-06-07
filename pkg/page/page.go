// Package page 保持 HTTP 面向的分页响应与传统 MSCoin 前端契约兼容。
//
// 本包提供统一的分页响应格式，用于列表类 API 的返回。
// 分页信息与前端约定一致，减少前后端沟通成本。
//
// 分页术语说明：
//   - Content: 当前页的数据列表
//   - TotalElements: 总记录数
//   - Number: 当前页码（从 0 开始）
//   - TotalPages: 总页数
//   - HasNext: 是否有下一页
//   - IsLast: 是否最后一页
//
// 使用场景：
//   - 用户列表查询
//   - 交易记录查询
//   - 资产记录查询
package page

import "math"

// Result 是列表端点返回的公共 JSON 分页模型。
//
// JSON 字段名称与传统 MSCoin 前端契约一致，
// 前端可以直接使用这些字段进行分页渲染。
//
// 字段说明：
//   - Content: 当前页的数据列表，使用 []any 以支持多种数据类型
//   - TotalElements: 总记录数，用于前端显示总数和计算总页数
//   - Number: 当前页码（从 0 开始），前端通常需要 +1 显示
//   - TotalPages: 总页数，由后端计算，避免前端重复计算
//   - HasNext: 是否有下一页，简化前端的下一页按钮状态判断
//   - IsLast: 是否最后一页，与 HasNext 互斥
type Result struct {
	Content       []any `json:"content"`       // 当前页数据列表
	TotalElements int64 `json:"totalElements"` // 总记录数
	Number        int64 `json:"number"`        // 当前页码（从 0 开始）
	TotalPages    int64 `json:"totalPages"`    // 总页数
	HasNext       bool  `json:"hasNext"`       // 是否有下一页
	IsLast        bool  `json:"isLast"`        // 是否最后一页
}

// New 从原始项目和总数构造分页信封。
//
// 该函数会自动计算：
//   - TotalPages: 根据总记录数和每页大小计算
//   - HasNext: 根据当前页码和总页数计算
//   - IsLast: 与 HasNext 相反
//
// 参数：
//   - content: 当前页的数据列表
//   - page: 当前页码（从 0 开始）
//   - pageSize: 每页大小，0 或负数会导致 TotalPages = 1
//   - total: 总记录数
//
// 返回值：
//   - *Result: 分页结果实例
//
// 使用示例：
//
//	// 查询第 2 页（page=1），每页 10 条
//	users := db.QueryUsers(page, pageSize)
//	total := db.CountUsers()
//	result := page.New(users, 1, 10, total)
//	// 返回给前端
func New(content []any, page int64, pageSize int64, total int64) *Result {
	resp := &Result{
		Content:       content,
		TotalElements: total,
		Number:        page,
	}

	// 计算总页数
	// pageSize <= 0 时，视为不分页，总页数为 1
	if pageSize <= 0 {
		resp.TotalPages = 1
	} else {
		resp.TotalPages = int64(math.Ceil(float64(total) / float64(pageSize)))
	}

	// 计算是否有下一页和是否最后一页
	// 当前页码 + 1 < 总页数，说明还有下一页
	resp.HasNext = resp.Number+1 < resp.TotalPages
	resp.IsLast = !resp.HasNext

	return resp
}