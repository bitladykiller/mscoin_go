package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

// FindAddressByCoinIDLogic 根据币种 ID 查询提现地址的逻辑处理器
type FindAddressByCoinIDLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFindAddressByCoinIDLogic 创建逻辑处理器实例
func NewFindAddressByCoinIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindAddressByCoinIDLogic {
	return &FindAddressByCoinIDLogic{ctx: ctx, svcCtx: svcCtx}
}

// FindAddressByCoinID 根据币种 ID 查询会员提现地址
func (l *FindAddressByCoinIDLogic) FindAddressByCoinID(req *withdrawpb.WithdrawReq) (*withdrawpb.AddressSimpleList, error) {
	list, err := l.svcCtx.WithdrawService.FindAddressByCoinID(l.ctx, req.UserId, req.CoinId)
	if err != nil {
		return nil, err
	}
	return &withdrawpb.AddressSimpleList{List: list}, nil
}
