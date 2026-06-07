// Package handler 定义了 exchange-api 的 HTTP 请求处理器。
// 每个处理器负责解析请求、调用业务逻辑、返回响应。
package handler

import (
	"net/http"

	"mscoin_go/app/exchange/api/internal/logic"
	"mscoin_go/app/exchange/api/internal/svc"
	"mscoin_go/app/exchange/api/internal/types"
	"mscoin_go/pkg/httputil"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// AddHandler 返回处理新增订单请求的 HTTP 处理器函数。
// 处理流程：
// 1. 解析请求参数到 ExchangeReq 结构体
// 2. 获取客户端 IP 地址
// 3. 调用 AddLogic 执行业务逻辑
// 4. 返回处理结果
func AddHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ExchangeReq
		if err := httpx.ParseForm(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 获取客户端 IP 并设置到请求中
		req.IP = httputil.ClientIP(r)
		// 调用业务逻辑层处理请求
		resp, err := logic.NewAddLogic(r.Context(), svcCtx).Add(&req)
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
