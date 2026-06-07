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
	// 为什么在这里也做防护：
	// - HTTP handler 已经为真实请求验证了验证码的存在性
	// - 但 logic 层仍可能被测试或未来的内部调用者直接复用
	// - 在 logic 层保持不变性检查可以防止绕过传输层适配器导致的空指针异常
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
