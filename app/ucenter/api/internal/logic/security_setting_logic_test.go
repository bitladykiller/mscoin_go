package logic

import (
	"context"
	"testing"

	"mscoin_go/app/ucenter/api/internal/config"
	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"

	"google.golang.org/grpc"
)

type fakeMemberClient struct {
	findMemberByIDFn func(ctx context.Context, in *memberpb.MemberReq, opts ...grpc.CallOption) (*memberpb.MemberInfo, error)
}

func (f *fakeMemberClient) FindMemberById(ctx context.Context, in *memberpb.MemberReq, opts ...grpc.CallOption) (*memberpb.MemberInfo, error) {
	return f.findMemberByIDFn(ctx, in, opts...)
}

func TestFindSecuritySettingMapsMemberInfo(t *testing.T) {
	logic := NewSecuritySettingLogic(
		middleware.WithUserID(context.Background(), 99),
		&svc.ServiceContext{
			Config: config.Config{},
			MemberClient: &fakeMemberClient{
				findMemberByIDFn: func(ctx context.Context, in *memberpb.MemberReq, opts ...grpc.CallOption) (*memberpb.MemberInfo, error) {
					if in.MemberId != 99 {
						t.Fatalf("FindSecuritySetting() memberId = %d, want 99", in.MemberId)
					}
					return &memberpb.MemberInfo{
						Id:               99,
						Username:         "alice",
						RegistrationTime: 1710000000000,
						Email:            "alice@example.com",
						JyPassword:       "hashed-jy-password",
						MobilePhone:      "13800000000",
						RealName:         "Alice",
						IdNumber:         "110101199001011234",
						RealNameStatus:   1,
						Avatar:           "avatar.png",
						Bank:             "ICBC",
					}, nil
				},
			},
		},
	)

	resp, err := logic.FindSecuritySetting(&types.ApproveReq{})
	if err != nil {
		t.Fatalf("FindSecuritySetting() error = %v", err)
	}
	if resp.Username != "alice" {
		t.Fatalf("FindSecuritySetting().Username = %q, want %q", resp.Username, "alice")
	}
	if resp.EmailVerified != "true" || resp.FundsVerified != "true" || resp.PhoneVerified != "true" {
		t.Fatalf("FindSecuritySetting() verification flags are unexpected: %+v", resp)
	}
	if resp.IdCard != "11********" {
		t.Fatalf("FindSecuritySetting().IdCard = %q, want %q", resp.IdCard, "11********")
	}
	if resp.AccountVerified != "true" {
		t.Fatalf("FindSecuritySetting().AccountVerified = %q, want %q", resp.AccountVerified, "true")
	}
}

func TestFindSecuritySettingMapsMissingFieldsToFalse(t *testing.T) {
	logic := NewSecuritySettingLogic(
		middleware.WithUserID(context.Background(), 11),
		&svc.ServiceContext{
			Config: config.Config{},
			MemberClient: &fakeMemberClient{
				findMemberByIDFn: func(ctx context.Context, in *memberpb.MemberReq, opts ...grpc.CallOption) (*memberpb.MemberInfo, error) {
					return &memberpb.MemberInfo{
						Id:       11,
						Username: "bob",
					}, nil
				},
			},
		},
	)

	resp, err := logic.FindSecuritySetting(&types.ApproveReq{})
	if err != nil {
		t.Fatalf("FindSecuritySetting() error = %v", err)
	}
	if resp.EmailVerified != "false" || resp.FundsVerified != "false" || resp.PhoneVerified != "false" || resp.RealVerified != "false" {
		t.Fatalf("FindSecuritySetting() false flags are unexpected: %+v", resp)
	}
	if resp.AccountVerified != "false" {
		t.Fatalf("FindSecuritySetting().AccountVerified = %q, want %q", resp.AccountVerified, "false")
	}
}
