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

type FindWalletBySymbolLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFindWalletBySymbolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindWalletBySymbolLogic {
	return &FindWalletBySymbolLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *FindWalletBySymbolLogic) FindWalletBySymbol(req *types.AssetReq) (*types.MemberWallet, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	userID := middleware.UserIDFromContext(l.ctx)
	payload, err := l.svcCtx.AssetClient.FindWalletBySymbol(ctx, &assetpb.AssetReq{
		UserId:   userID,
		CoinName: req.CoinName,
	})
	if err != nil {
		return nil, err
	}

	resp := &types.MemberWallet{}
	if err := copier.Copy(resp, payload); err != nil {
		return nil, err
	}
	return resp, nil
}
