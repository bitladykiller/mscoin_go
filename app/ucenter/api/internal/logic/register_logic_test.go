package logic

import (
	"context"
	"testing"

	"mscoin_go/app/ucenter/api/internal/config"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"

	"google.golang.org/grpc"
)

type fakeRegisterClient struct {
	registerByPhoneFn func(ctx context.Context, in *registerpb.RegReq, opts ...grpc.CallOption) (*registerpb.RegRes, error)
	sendCodeFn        func(ctx context.Context, in *registerpb.CodeReq, opts ...grpc.CallOption) (*registerpb.NoRes, error)
}

func (f *fakeRegisterClient) RegisterByPhone(ctx context.Context, in *registerpb.RegReq, opts ...grpc.CallOption) (*registerpb.RegRes, error) {
	return f.registerByPhoneFn(ctx, in, opts...)
}

func (f *fakeRegisterClient) SendCode(ctx context.Context, in *registerpb.CodeReq, opts ...grpc.CallOption) (*registerpb.NoRes, error) {
	return f.sendCodeFn(ctx, in, opts...)
}

func TestRegisterRejectsNilCaptcha(t *testing.T) {
	logic := NewRegisterLogic(context.Background(), &svc.ServiceContext{
		Config: config.Config{},
		RegisterClient: &fakeRegisterClient{
			registerByPhoneFn: func(ctx context.Context, in *registerpb.RegReq, opts ...grpc.CallOption) (*registerpb.RegRes, error) {
				t.Fatal("Register RPC should not be called when captcha is nil")
				return nil, nil
			},
			sendCodeFn: func(ctx context.Context, in *registerpb.CodeReq, opts ...grpc.CallOption) (*registerpb.NoRes, error) {
				return nil, nil
			},
		},
	})

	if _, err := logic.Register(&types.Request{Phone: "13800000000"}); err == nil {
		t.Fatal("Register() error = nil, want captcha validation error")
	}
}

func TestRegisterMapsRequestToRPC(t *testing.T) {
	logic := NewRegisterLogic(context.Background(), &svc.ServiceContext{
		Config: config.Config{},
		RegisterClient: &fakeRegisterClient{
			registerByPhoneFn: func(ctx context.Context, in *registerpb.RegReq, opts ...grpc.CallOption) (*registerpb.RegRes, error) {
				if in.Phone != "13800000000" {
					t.Fatalf("Register() phone = %q, want %q", in.Phone, "13800000000")
				}
				if in.Captcha == nil || in.Captcha.Server != "captcha-server" || in.Captcha.Token != "captcha-token" {
					t.Fatalf("Register() forwarded unexpected captcha: %+v", in.Captcha)
				}
				if in.SuperPartner != "1" {
					t.Fatalf("Register() superPartner = %q, want %q", in.SuperPartner, "1")
				}
				return &registerpb.RegRes{}, nil
			},
			sendCodeFn: func(ctx context.Context, in *registerpb.CodeReq, opts ...grpc.CallOption) (*registerpb.NoRes, error) {
				return nil, nil
			},
		},
	})

	resp, err := logic.Register(&types.Request{
		Username:     "alice",
		Password:     "secret",
		Captcha:      &types.CaptchaReq{Server: "captcha-server", Token: "captcha-token"},
		Phone:        "13800000000",
		Promotion:    "PROMO",
		Code:         "1234",
		Country:      "CN",
		SuperPartner: "1",
		IP:           "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if resp == nil {
		t.Fatal("Register() response = nil, want non-nil")
	}
}

func TestSendCodeMapsRequestToRPC(t *testing.T) {
	logic := NewSendCodeLogic(context.Background(), &svc.ServiceContext{
		Config: config.Config{},
		RegisterClient: &fakeRegisterClient{
			registerByPhoneFn: func(ctx context.Context, in *registerpb.RegReq, opts ...grpc.CallOption) (*registerpb.RegRes, error) {
				return nil, nil
			},
			sendCodeFn: func(ctx context.Context, in *registerpb.CodeReq, opts ...grpc.CallOption) (*registerpb.NoRes, error) {
				if in.Phone != "13800000000" || in.Country != "CN" {
					t.Fatalf("SendCode() forwarded unexpected payload: %+v", in)
				}
				return &registerpb.NoRes{}, nil
			},
		},
	})

	resp, err := logic.SendCode(&types.CodeRequest{
		Phone:   "13800000000",
		Country: "CN",
	})
	if err != nil {
		t.Fatalf("SendCode() error = %v", err)
	}
	if resp == nil {
		t.Fatal("SendCode() response = nil, want non-nil")
	}
}
