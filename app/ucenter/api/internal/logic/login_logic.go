package logic

import (
	"context"
	"errors"
	"time"

	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	loginpb "mscoin_go/app/ucenter/rpc/pb/login"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *LoginLogic) Login(req *types.LoginReq) (*types.LoginRes, error) {
	// Why guard here as well:
	// - the HTTP handler already validates captcha presence for real requests
	// - logic can still be reused directly by tests or future internal callers
	// - keeping the invariant inside the logic layer prevents nil-pointer panics
	//   from bypassing the transport adapter
	if req == nil || req.Captcha == nil {
		return nil, errors.New("captcha verification failed")
	}

	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	payload, err := l.svcCtx.LoginClient.Login(ctx, &loginpb.LoginReq{
		Username: req.Username,
		Password: req.Password,
		Captcha: &loginpb.CaptchaReq{
			Server: req.Captcha.Server,
			Token:  req.Captcha.Token,
		},
		Ip: req.IP,
	})
	if err != nil {
		return nil, err
	}

	return &types.LoginRes{
		Username:      payload.Username,
		Token:         payload.Token,
		MemberLevel:   payload.MemberLevel,
		RealName:      payload.RealName,
		Country:       payload.Country,
		Avatar:        payload.Avatar,
		PromotionCode: payload.PromotionCode,
		Id:            payload.Id,
		LoginCount:    int(payload.LoginCount),
		SuperPartner:  payload.SuperPartner,
		MemberRate:    int(payload.MemberRate),
	}, nil
}
