// Package handler 提供 ucenter-api 服务的 HTTP 请求处理器。
//
// 该文件包含查询钱包列表相关的 HTTP 处理器。
package handler

import (
	"net/http"

	"mscoin_go/app/ucenter/api/internal/logic"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// FindWalletHandler 处理查询用户所有钱包请求。
//
// 该接口返回用户持有的所有币种钱包信息，包括：
//   - 钱包余额（可用余额、冻结余额、待释放余额）
//   - 充值地址
//   - 关联的币种信息
//
// 请求路径：POST /uc/asset/wallet
// 认证要求：需要 JWT Token（通过 Auth 中间件验证）
//
// 用户身份获取：
//   - 通过 middleware.UserIDFromContext 从 context 获取用户 ID
//
// RPC 调用关系：
//   - AssetClient.FindWallet -> ucenter-rpc (asset pb)
//   - ucenter-rpc 负责：查询会员所有钱包、组装币种信息、计算总价值
func FindWalletHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 调用 logic 层查询钱包列表
		resp, err := logic.NewFindWalletLogic(r.Context(), svcCtx).FindWallet()
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
