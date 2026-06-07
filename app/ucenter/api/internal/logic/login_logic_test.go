package logic

import (
	"context"
	"testing"

	"mscoin_go/app/ucenter/api/internal/config"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	loginpb "mscoin_go/app/ucenter/rpc/pb/login"

	"google.golang.org/grpc"
)

type fakeLoginClient struct {
	loginFn func(ctx context.Context, in *loginpb.LoginReq, opts ...grpc.CallOption) (*loginpb.LoginRes, error)
}

func (f *fakeLoginClient) Login(ctx context.Context, in *loginpb.LoginReq, opts ...grpc.CallOption) (*loginpb.LoginRes, error) {
	return f.loginFn(ctx, in, opts...)
}

func TestLoginRejectsNilCaptcha(t *testing.T) {
	logic := NewLoginLogic(context.Background(), &svc.ServiceContext{
		Config: config.Config{},
		LoginClient: &fakeLoginClient{
			loginFn: func(ctx context.Context, in *loginpb.LoginReq, opts ...grpc.CallOption) (*loginpb.LoginRes, error) {
				t.Fatal("Login RPC should not be called when captcha is nil")
				return nil, nil
			},
		},
	})

	if _, err := logic.Login(&types.LoginReq{Username: "13800000000", Password: "secret"}); err == nil {
		t.Fatal("Login() error = nil, want captcha validation error")
	}
}

func TestLoginMapsRPCResponse(t *testing.T) {
	logic := NewLoginLogic(context.Background(), &svc.ServiceContext{
		Config: config.Config{},
		LoginClient: &fakeLoginClient{
			loginFn: func(ctx context.Context, in *loginpb.LoginReq, opts ...grpc.CallOption) (*loginpb.LoginRes, error) {
				if in.Captcha == nil {
					t.Fatal("Login() forwarded nil captcha")
				}
				if in.Captcha.Server != "server-token" || in.Captcha.Token != "captcha-token" {
					t.Fatalf("Login() forwarded unexpected captcha: %+v", in.Captcha)
				}
				return &loginpb.LoginRes{
					Username:      "alice",
					Token:         "jwt-token",
					MemberLevel:   "实名",
					RealName:      "Alice",
					Country:       "CN",
					Avatar:        "avatar.png",
					PromotionCode: "PROMO",
					Id:            88,
					LoginCount:    3,
					SuperPartner:  "1",
					MemberRate:    1,
				}, nil
			},
		},
	})

	resp, err := logic.Login(&types.LoginReq{
		Username: "13800000000",
		Password: "secret",
		Captcha: &types.CaptchaReq{
			Server: "server-token",
			Token:  "captcha-token",
		},
		IP: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if resp.Token != "jwt-token" {
		t.Fatalf("Login().Token = %q, want %q", resp.Token, "jwt-token")
	}
	if resp.MemberRate != 1 {
		t.Fatalf("Login().MemberRate = %d, want 1", resp.MemberRate)
	}
	if resp.Id != 88 {
		t.Fatalf("Login().Id = %d, want 88", resp.Id)
	}
}
