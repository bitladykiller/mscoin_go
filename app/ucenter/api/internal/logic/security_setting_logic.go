package logic

import (
	"context"
	"time"

	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
)

type SecuritySettingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSecuritySettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SecuritySettingLogic {
	return &SecuritySettingLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *SecuritySettingLogic) FindSecuritySetting(_ *types.ApproveReq) (*types.MemberSecurity, error) {
	userID := middleware.UserIDFromContext(l.ctx)
	memberRes, err := l.svcCtx.MemberClient.FindMemberById(l.ctx, &memberpb.MemberReq{MemberId: userID})
	if err != nil {
		return nil, err
	}

	resp := &types.MemberSecurity{
		Username:        memberRes.Username,
		Id:              memberRes.Id,
		CreateTime:      formatMemberTime(memberRes.RegistrationTime),
		LoginVerified:   "true",
		RealAuditing:    boolString(memberRes.RealNameStatus == 1),
		Avatar:          memberRes.Avatar,
		AccountVerified: boolString(memberRes.Bank != "" || memberRes.AliNo != "" || memberRes.Wechat != ""),
	}

	if memberRes.Email != "" {
		resp.EmailVerified = "true"
		resp.Email = memberRes.Email
	} else {
		resp.EmailVerified = "false"
	}

	if memberRes.JyPassword != "" {
		resp.FundsVerified = "true"
	} else {
		resp.FundsVerified = "false"
	}

	if memberRes.MobilePhone != "" {
		resp.PhoneVerified = "true"
		resp.MobilePhone = memberRes.MobilePhone
	} else {
		resp.PhoneVerified = "false"
	}

	if memberRes.RealName != "" {
		resp.RealVerified = "true"
		resp.RealName = memberRes.RealName
	} else {
		resp.RealVerified = "false"
	}

	if memberRes.IdNumber != "" {
		resp.IdCard = maskIDCard(memberRes.IdNumber)
	}

	return resp, nil
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func formatMemberTime(millis int64) string {
	if millis <= 0 {
		return ""
	}
	return time.UnixMilli(millis).Format("2006-01-02 15:04:05")
}

func maskIDCard(idNumber string) string {
	if len(idNumber) < 2 {
		return idNumber
	}
	return idNumber[:2] + "********"
}
