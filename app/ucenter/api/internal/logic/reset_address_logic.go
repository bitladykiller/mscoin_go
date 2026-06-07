package logic

import (
	"context"
	"time"

	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

type ResetAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetAddressLogic {
	return &ResetAddressLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ResetAddressLogic) ResetAddress(req *types.AssetReq) (string, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	coinName := req.Unit
	if coinName == "" {
		coinName = req.CoinName
	}

	userID := middleware.UserIDFromContext(l.ctx)
	_, err := l.svcCtx.AssetClient.ResetAddress(ctx, &assetpb.AssetReq{
		UserId:   userID,
		CoinName: coinName,
		Ip:       req.IP,
	})
	if err != nil {
		return "", err
	}
	return "", nil
}
