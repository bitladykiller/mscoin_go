// Package result defines the uniform HTTP response envelope used by API
// services. The old project returned the same JSON shape in many handlers, so
// the refactor keeps that contract while moving the implementation to a single
// shared package.
package result

// Result is the standard API response body.
//
// Code:
//
//	0   means success
//	500 means generic business failure in the legacy behavior
//
// Message:
//
//	Human-readable status text for the caller.
//
// Data:
//
//	Arbitrary payload returned by the endpoint.
type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// New creates an empty result object so handlers can consistently build a
// legacy-compatible response body.
func New() *Result {
	return &Result{}
}

// Success marks the response as successful.
func (r *Result) Success(data any) {
	r.Code = 0
	r.Message = "success"
	r.Data = data
}

// Fail marks the response as failed.
func (r *Result) Fail(code int, message string) {
	r.Code = code
	r.Message = message
	r.Data = nil
}

// Deal translates the common `(data, err)` pattern into the legacy MSCoin API
// envelope. This keeps handlers small and ensures every endpoint returns a
// consistent JSON contract.
func (r *Result) Deal(data any, err error) *Result {
	if err != nil {
		r.Fail(500, err.Error())
		return r
	}

	r.Success(data)
	return r
}
