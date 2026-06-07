package logic

import (
	"context"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// FindWalletLogic 查询钱包列表的逻辑处理器
type FindWalletLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFindWalletLogic 创建逻辑处理器实例
func NewFindWalletLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindWalletLogic {
	return &FindWalletLogic{ctx: ctx, svcCtx: svcCtx}
}

// FindWallet 查询会员所有钱包
func (l *FindWalletLogic) FindWallet(req *assetpb.AssetReq) (*assetpb.MemberWalletList, error) {
	list, err := l.svcCtx.WalletService.FindWallet(l.ctx, req.UserId, func(ctx context.Context, unit string) (*marketpb.Coin, error) {
		return l.svcCtx.MarketClient.FindCoinInfo(ctx, &marketpb.MarketReq{Unit: unit})
	})
	if err != nil {
		return nil, err
	}
	return &assetpb.MemberWalletList{List: list}, nil
}
