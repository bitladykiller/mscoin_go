package logic

import (
	"context"
	"errors"
	"time"

	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *RegisterLogic) Register(req *types.Request) (*types.Response, error) {
	if req == nil || req.Captcha == nil {
		return nil, errors.New("captcha verification failed")
	}

	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	_, err := l.svcCtx.RegisterClient.RegisterByPhone(ctx, &registerpb.RegReq{
		Username: req.Username,
		Password: req.Password,
		Captcha: &registerpb.CaptchaReq{
			Server: req.Captcha.Server,
			Token:  req.Captcha.Token,
		},
		Phone:        req.Phone,
		Promotion:    req.Promotion,
		Code:         req.Code,
		Country:      req.Country,
		SuperPartner: req.SuperPartner,
		Ip:           req.IP,
	})
	if err != nil {
		return nil, err
	}

	return &types.Response{}, nil
}
