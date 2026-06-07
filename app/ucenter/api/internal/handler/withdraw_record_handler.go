// Package handler 提供 ucenter-api 服务的 HTTP 请求处理器。
//
// 该文件包含查询提现记录相关的 HTTP 处理器。
package handler

import (
	"net/http"

	"mscoin_go/app/ucenter/api/internal/logic"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// WithdrawRecordHandler 处理查询提现记录请求。
//
// 该接口返回用户的历史提现工单列表，用于：
//   - 提现记录页面展示
//   - 查看提现状态和处理进度
//
// 请求路径：POST /uc/withdraw/record
// 认证要求：需要 JWT Token（通过 Auth 中间件验证）
//
// 请求参数（表单提交）：
//   - Page：页码，默认 1
//   - PageSize：每页数量，默认 10
//
// 用户身份获取：
//   - 通过 middleware.UserIDFromContext 从 context 获取用户 ID
//
// RPC 调用关系：
//   - WithdrawClient.WithdrawRecord -> ucenter-rpc (withdraw pb)
//   - ucenter-rpc 负责：查询用户提现工单、分页处理、组装币种信息
//
// 提现状态说明：
//   - 0：待审核
//   - 1：审核通过/处理中
//   - 2：已完成/已打款
//   - 3：已拒绝
//   - 其他状态根据业务定义
func WithdrawRecordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WithdrawReq

		// 从表单解析分页参数
		if err := httpx.ParseForm(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 调用 logic 层查询提现记录
		resp, err := logic.NewWithdrawRecordLogic(r.Context(), svcCtx).Record(&req)
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
