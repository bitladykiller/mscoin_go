// Package result 定义 API 服务使用的统一 HTTP 响应信封。
//
// 旧项目在许多处理器中返回相同的 JSON 格式，因此重构保留该契约，
// 同时将实现移至单个共享包。
//
// 响应格式：
//
//	{
//	    "code": 0,        // 0 表示成功，500 表示失败
//	    "message": "xxx", // 状态描述
//	    "data": {...}     // 业务数据
//	}
//
// 使用场景：
//   - 所有 HTTP API 的统一响应格式
//   - 简化处理器的错误处理逻辑
//   - 前端可以根据 code 统一处理成功/失败
package result

// Result 是标准的 API 响应体。
//
// Code:
//   - 0 表示成功
//   - 500 表示传统行为中的通用业务失败
//   - 其他值可用于表示特定错误类型（如 401 未授权）
//
// Message:
//   - 给调用者的人类可读状态文本
//   - 成功时为 "success"
//   - 失败时为错误描述信息
//
// Data:
//   - 端点返回的任意负载
//   - 成功时包含业务数据
//   - 失败时为 nil
type Result struct {
	Code    int    `json:"code"`    // 状态码：0 成功，500 失败
	Message string `json:"message"` // 状态描述信息
	Data    any    `json:"data"`    // 业务数据
}

// New 创建一个空的结果对象，以便处理器可以一致地构建传统兼容的响应体。
//
// 返回值：
//   - *Result: 空的结果实例
//
// 使用示例：
//
//	resp := result.New()
//	resp.Success(user)
//	return resp
func New() *Result {
	return &Result{}
}

// Success 将响应标记为成功。
//
// 参数：
//   - data: 成功时返回的业务数据，可以是任意类型
//
// 使用示例：
//
//	resp := result.New()
//	resp.Success(map[string]any{
//	    "userId": 123,
//	    "name":   "test",
//	})
func (r *Result) Success(data any) {
	r.Code = 0
	r.Message = "success"
	r.Data = data
}

// Fail 将响应标记为失败。
//
// 参数：
//   - code: 错误码，传统行为使用 500
//   - message: 错误描述信息，将展示给用户
//
// 使用示例：
//
//	resp := result.New()
//	resp.Fail(500, "用户不存在")
//	return resp
func (r *Result) Fail(code int, message string) {
	r.Code = code
	r.Message = message
	r.Data = nil
}

// Deal 将常见的 `(data, err)` 模式转换为传统 MSCoin API 信封。
//
// 这保持处理器简洁，并确保每个端点返回一致的 JSON 契约。
//
// 参数：
//   - data: 业务数据，当 err 为 nil 时返回
//   - err: 错误，当不为 nil 时，使用其消息作为响应
//
// 返回值：
//   - *Result: 构建好的响应实例
//
// 使用示例：
//
//	user, err := userService.GetUser(userId)
//	return result.New().Deal(user, err)
//
// 常见用法：
//
//	func GetUserHandler(w http.ResponseWriter, r *http.Request) {
//	    userId := parseUserId(r)
//	    user, err := service.GetUser(userId)
//	    json.NewEncoder(w).Encode(result.New().Deal(user, err))
//	}
func (r *Result) Deal(data any, err error) *Result {
	if err != nil {
		// 错误情况：返回 500 状态码和错误信息
		r.Fail(500, err.Error())
		return r
	}

	// 成功情况：返回数据
	r.Success(data)
	return r
}