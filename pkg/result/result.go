// Package result 定义 API 服务使用的统一 HTTP 响应信封。
// 旧项目在许多处理器中返回相同的 JSON 格式，因此重构保留该契约，
// 同时将实现移至单个共享包。
package result

// Result 是标准的 API 响应体。
//
// Code:
//
//	0   表示成功
//	500 表示传统行为中的通用业务失败
//
// Message:
//
//	给调用者的人类可读状态文本。
//
// Data:
//
//	端点返回的任意负载。
type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// New 创建一个空的结果对象，以便处理器可以一致地构建传统兼容的响应体。
func New() *Result {
	return &Result{}
}

// Success 将响应标记为成功。
func (r *Result) Success(data any) {
	r.Code = 0
	r.Message = "success"
	r.Data = data
}

// Fail 将响应标记为失败。
func (r *Result) Fail(code int, message string) {
	r.Code = code
	r.Message = message
	r.Data = nil
}

// Deal 将常见的 `(data, err)` 模式转换为传统 MSCoin API 信封。
// 这保持处理器简洁，并确保每个端点返回一致的 JSON 契约。
func (r *Result) Deal(data any, err error) *Result {
	if err != nil {
		r.Fail(500, err.Error())
		return r
	}

	r.Success(data)
	return r
}
