package logic

import (
	"context"
	"time"

	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
)

type SendCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendCodeLogic {
	return &SendCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *SendCodeLogic) SendCode(req *types.CodeRequest) (*types.CodeResponse, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	_, err := l.svcCtx.RegisterClient.SendCode(ctx, &registerpb.CodeReq{
		Phone:   req.Phone,
		Country: req.Country,
	})
	if err != nil {
		return nil, err
	}

	return &types.CodeResponse{}, nil
}
