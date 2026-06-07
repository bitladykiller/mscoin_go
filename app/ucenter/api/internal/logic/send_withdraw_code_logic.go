package logic

import (
	"context"
	"time"

	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/rpc/pb/member"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

type SendWithdrawCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendWithdrawCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendWithdrawCodeLogic {
	return &SendWithdrawCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *SendWithdrawCodeLogic) SendCode() (string, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	userID := middleware.UserIDFromContext(l.ctx)
	memberInfo, err := l.svcCtx.MemberClient.FindMemberById(ctx, &member.MemberReq{MemberId: userID})
	if err != nil {
		return "", err
	}

	_, err = l.svcCtx.WithdrawClient.SendCode(ctx, &withdrawpb.WithdrawReq{Phone: memberInfo.MobilePhone})
	if err != nil {
		return "", err
	}

	return "success", nil
}
