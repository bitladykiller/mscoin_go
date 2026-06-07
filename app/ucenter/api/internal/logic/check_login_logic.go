package logic

import (
	"context"

	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/pkg/auth"
)

type CheckLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckLoginLogic {
	return &CheckLoginLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *CheckLoginLogic) CheckLogin(token string) (bool, error) {
	_, err := auth.ParseUserID(token, l.svcCtx.Config.JWT.AccessSecret)
	if err != nil {
		return false, nil
	}
	return true, nil
}
