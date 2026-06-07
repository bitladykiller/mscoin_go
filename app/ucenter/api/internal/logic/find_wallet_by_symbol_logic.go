// Package logic 提供 ucenter-api 服务的业务逻辑处理。
//
// 该文件包含按币种查询钱包相关的业务逻辑。
package logic

import (
	"context"
	"time"

	"github.com/jinzhu/copier"

	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// FindWalletBySymbolLogic 是按币种查询钱包业务逻辑处理器。
//
// 该结构体负责查询用户指定币种的钱包详情。
type FindWalletBySymbolLogic struct {
	// ctx 是请求上下文，包含已认证的用户 ID。
	ctx    context.Context

	// svcCtx 是服务上下文，提供 RPC 客户端访问能力。
	svcCtx *svc.ServiceContext
}

// NewFindWalletBySymbolLogic 创建按币种查询钱包业务逻辑处理器实例。
func NewFindWalletBySymbolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindWalletBySymbolLogic {
	return &FindWalletBySymbolLogic{ctx: ctx, svcCtx: svcCtx}
}

// FindWalletBySymbol 执行按币种查询钱包业务逻辑。
//
// 查询流程：
//  1. 从 context 获取已认证的用户 ID
//  2. 从请求中获取币种名称
//  3. 调用 ucenter-rpc AssetClient 查询指定币种钱包
//  4. 转换 RPC 响应为 API 响应格式
//
// 使用场景：
//   - 充值页面：展示充值地址和币种信息
//   - 提现页面：展示余额和提现限制
//
// RPC 调用：
//   - AssetClient.FindWalletBySymbol -> ucenter-rpc
//   - ucenter-rpc 负责：查询单个币种钱包、返回余额和充值地址
//
// 参数：
//   - req：资产请求，包含币种名称（通过 URL 路径传递）
//
// 返回：
//   - *types.MemberWallet：单个钱包详情
//   - error：查询失败时的错误信息
func (l *FindWalletBySymbolLogic) FindWalletBySymbol(req *types.AssetReq) (*types.MemberWallet, error) {
	// 设置 RPC 调用超时
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	// 从 context 获取已认证的用户 ID
	userID := middleware.UserIDFromContext(l.ctx)

	// 调用 ucenter-rpc AssetClient 查询指定币种钱包
	payload, err := l.svcCtx.AssetClient.FindWalletBySymbol(ctx, &assetpb.AssetReq{
		UserId:   userID,
		CoinName: req.CoinName,
	})
	if err != nil {
		return nil, err
	}

	// 转换响应格式
	resp := &types.MemberWallet{}
	if err := copier.Copy(resp, payload); err != nil {
		return nil, err
	}
	return resp, nil
}
