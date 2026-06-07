// Package handler 提供 ucenter-api 服务的 HTTP 请求处理器。
//
// 该文件包含查询交易记录相关的 HTTP 处理器。
package handler

import (
	"net/http"

	"mscoin_go/app/ucenter/api/internal/logic"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// FindTransactionHandler 处理查询交易记录请求。
//
// 该接口返回用户的资产交易历史，包括：
//   - 充值记录
//   - 提现记录
//   - 转账记录
//   - 其他资产变动记录
//
// 请求路径：POST /uc/asset/transaction/all
// 认证要求：需要 JWT Token（通过 Auth 中间件验证）
//
// 查询参数（表单提交）：
//   - PageNo：页码，默认 1
//   - PageSize：每页数量，默认 10
//   - StartTime：起始时间
//   - EndTime：结束时间
//   - Symbol：币种筛选
//   - Type：交易类型筛选
//
// 用户身份获取：
//   - 通过 middleware.UserIDFromContext 从 context 获取用户 ID
//
// RPC 调用关系：
//   - AssetClient.FindTransaction -> ucenter-rpc (asset pb)
//   - ucenter-rpc 负责：查询交易记录、按条件筛选、分页处理
func FindTransactionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AssetReq

		// 从表单解析查询参数（支持分页和时间筛选）
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 调用 logic 层查询交易记录
		resp, err := logic.NewFindTransactionLogic(r.Context(), svcCtx).FindTransaction(&req)
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
