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

// CurrentHandler 返回处理查询当前订单请求的 HTTP 处理器函数。
//
// 处理流程：
// 1. 解析请求参数到 ExchangeReq 结构体
// 2. 获取客户端 IP 地址（用于风控和审计）
// 3. 调用 CurrentLogic 执行业务逻辑
// 4. 返回处理结果（分页的订单列表）
//
// 与 exchange-rpc 的调用关系：
// - 该处理器通过 ServiceContext.OrderClient 调用 exchange-rpc 的 FindOrderCurrent 方法
// - 返回用户当前正在委托中的订单（状态为 TRADING）
//
// 查询条件：
// - 用户 ID（从 JWT 令牌中提取）
// - 交易对符号（可选，为空时查询所有交易对）
// - 分页参数
func CurrentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ExchangeReq
		if err := httpx.ParseForm(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 获取客户端 IP 并设置到请求中
		req.IP = httputil.ClientIP(r)
		// 调用业务逻辑层处理请求
		resp, err := logic.NewCurrentLogic(r.Context(), svcCtx).Current(&req)
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
