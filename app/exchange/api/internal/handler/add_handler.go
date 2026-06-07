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
//
// 处理流程：
// 1. 解析请求参数到 ExchangeReq 结构体
// 2. 获取客户端 IP 地址（用于风控和审计）
// 3. 调用 AddLogic 执行业务逻辑
// 4. 返回处理结果（订单 ID）
//
// 与 exchange-rpc 的调用关系：
// - 该处理器通过 ServiceContext.OrderClient 调用 exchange-rpc 的 Add 方法
// - exchange-rpc 会进一步调用 ucenter-rpc 和 market-rpc 验证用户状态和交易对信息
//
// 业务规则验证（在 exchange-rpc 中执行）：
// - 用户交易状态是否正常（通过 ucenter-rpc 查询会员信息）
// - 交易对是否可交易（通过 market-rpc 查询交易对配置）
// - 用户钱包是否被锁定（通过 ucenter-rpc 查询钱包信息）
// - 当前委托订单数量是否超限
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
