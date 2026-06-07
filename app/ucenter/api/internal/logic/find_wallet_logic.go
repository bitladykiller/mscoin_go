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

type FindWalletLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFindWalletLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindWalletLogic {
	return &FindWalletLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *FindWalletLogic) FindWallet() ([]*types.MemberWallet, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	userID := middleware.UserIDFromContext(l.ctx)
	payload, err := l.svcCtx.AssetClient.FindWallet(ctx, &assetpb.AssetReq{UserId: userID})
	if err != nil {
		return nil, err
	}

	resp := make([]*types.MemberWallet, len(payload.List))
	if err := copier.Copy(&resp, payload.List); err != nil {
		return nil, err
	}
	return resp, nil
}
