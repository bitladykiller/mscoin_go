package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// GetAddressLogic 获取地址列表的逻辑处理器
type GetAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetAddressLogic 创建逻辑处理器实例
func NewGetAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressLogic {
	return &GetAddressLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetAddress 获取指定币种的所有钱包地址
func (l *GetAddressLogic) GetAddress(req *assetpb.AssetReq) (*assetpb.AddressList, error) {
	addresses, err := l.svcCtx.WalletService.GetAllAddress(l.ctx, req.CoinName)
	if err != nil {
		return nil, err
	}
	return &assetpb.AddressList{List: addresses}, nil
}
