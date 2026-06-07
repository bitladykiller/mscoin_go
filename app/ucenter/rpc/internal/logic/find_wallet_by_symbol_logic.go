// Package logic 定义按币种查询钱包业务逻辑处理器。
package logic

import (
	"context"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// FindWalletBySymbolLogic 根据币种查询钱包的逻辑处理器
// 处理按币种查询会员钱包 RPC 请求
type FindWalletBySymbolLogic struct {
	ctx    context.Context     // 请求上下文
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewFindWalletBySymbolLogic 创建逻辑处理器实例
func NewFindWalletBySymbolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindWalletBySymbolLogic {
	return &FindWalletBySymbolLogic{ctx: ctx, svcCtx: svcCtx}
}

// FindWalletBySymbol 根据币种查询会员钱包
// 调用 WalletService.FindWalletBySymbol 处理查询逻辑
//
// 参数：
//   - req: 资产请求，包含用户 ID 和币种名称
//
// 返回：
//   - MemberWallet: 钱包信息，包含余额和币种信息
//   - error: 错误信息
//
// 注意：如果钱包不存在，会自动创建
func (l *FindWalletBySymbolLogic) FindWalletBySymbol(req *assetpb.AssetReq) (*assetpb.MemberWallet, error) {
	// 从 Market RPC 获取币种信息
	coin, err := l.svcCtx.MarketClient.FindCoinInfo(l.ctx, &marketpb.MarketReq{Unit: req.CoinName})
	if err != nil {
		return nil, err
	}

	// 调用 WalletService 查询钱包
	// 如果钱包不存在会自动创建
	return l.svcCtx.WalletService.FindWalletBySymbol(l.ctx, req.UserId, req.CoinName, coin)
}
