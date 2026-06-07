package logic

import (
	"context"
	"testing"

	"mscoin_go/app/ucenter/api/internal/config"
	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"

	"google.golang.org/grpc"
)

type fakeAssetClient struct {
	findWalletBySymbolFn func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWallet, error)
	findWalletFn         func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWalletList, error)
	resetAddressFn       func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.AssetResp, error)
	findTransactionFn    func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberTransactionList, error)
	getAddressFn         func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.AddressList, error)
}

func (f *fakeAssetClient) FindWalletBySymbol(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWallet, error) {
	return f.findWalletBySymbolFn(ctx, in, opts...)
}

func (f *fakeAssetClient) FindWallet(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWalletList, error) {
	return f.findWalletFn(ctx, in, opts...)
}

func (f *fakeAssetClient) ResetAddress(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.AssetResp, error) {
	if f.resetAddressFn != nil {
		return f.resetAddressFn(ctx, in, opts...)
	}
	return &assetpb.AssetResp{}, nil
}

func (f *fakeAssetClient) FindTransaction(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberTransactionList, error) {
	return f.findTransactionFn(ctx, in, opts...)
}

func (f *fakeAssetClient) GetAddress(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.AddressList, error) {
	if f.getAddressFn != nil {
		return f.getAddressFn(ctx, in, opts...)
	}
	return &assetpb.AddressList{}, nil
}

func TestResetAddressUsesAuthenticatedUserAndUnit(t *testing.T) {
	ctx := middleware.WithUserID(context.Background(), 101)
	logic := NewResetAddressLogic(ctx, &svc.ServiceContext{
		Config: config.Config{},
		AssetClient: &fakeAssetClient{
			findWalletBySymbolFn: func(context.Context, *assetpb.AssetReq, ...grpc.CallOption) (*assetpb.MemberWallet, error) {
				return nil, nil
			},
			findWalletFn: func(context.Context, *assetpb.AssetReq, ...grpc.CallOption) (*assetpb.MemberWalletList, error) {
				return nil, nil
			},
			resetAddressFn: func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.AssetResp, error) {
				if in.UserId != 101 {
					t.Fatalf("ResetAddress() userId = %d, want 101", in.UserId)
				}
				if in.CoinName != "BTC" {
					t.Fatalf("ResetAddress() coinName = %q, want BTC", in.CoinName)
				}
				if in.Ip != "127.0.0.1" {
					t.Fatalf("ResetAddress() ip = %q, want 127.0.0.1", in.Ip)
				}
				return &assetpb.AssetResp{}, nil
			},
			findTransactionFn: func(context.Context, *assetpb.AssetReq, ...grpc.CallOption) (*assetpb.MemberTransactionList, error) {
				return nil, nil
			},
		},
	})

	resp, err := logic.ResetAddress(&types.AssetReq{
		Unit: "BTC",
		IP:   "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("ResetAddress() error = %v", err)
	}
	if resp != "" {
		t.Fatalf("ResetAddress() = %q, want empty string", resp)
	}
}

func TestFindTransactionMapsPageRequest(t *testing.T) {
	ctx := middleware.WithUserID(context.Background(), 88)
	logic := NewFindTransactionLogic(ctx, &svc.ServiceContext{
		Config: config.Config{},
		AssetClient: &fakeAssetClient{
			findWalletBySymbolFn: func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWallet, error) {
				return nil, nil
			},
			findWalletFn: func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWalletList, error) {
				return nil, nil
			},
			findTransactionFn: func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberTransactionList, error) {
				if in.UserId != 88 {
					t.Fatalf("FindTransaction() userId = %d, want 88", in.UserId)
				}
				if in.PageNo != 2 || in.PageSize != 20 {
					t.Fatalf("FindTransaction() paging = (%d,%d), want (2,20)", in.PageNo, in.PageSize)
				}
				if in.Symbol != "BTC" || in.Type != "1" {
					t.Fatalf("FindTransaction() filter = (%q,%q), want (BTC,1)", in.Symbol, in.Type)
				}
				return &assetpb.MemberTransactionList{
					List: []*assetpb.MemberTransaction{
						{Id: 1, Type: "WITHDRAW"},
					},
					Total: 25,
				}, nil
			},
		},
	})

	resp, err := logic.FindTransaction(&types.AssetReq{
		PageNo:   2,
		PageSize: 20,
		Symbol:   "BTC",
		Type:     "1",
	})
	if err != nil {
		t.Fatalf("FindTransaction() error = %v", err)
	}
	if resp.TotalElements != 25 {
		t.Fatalf("FindTransaction().TotalElements = %d, want 25", resp.TotalElements)
	}
	if len(resp.Content) != 1 {
		t.Fatalf("FindTransaction().Content len = %d, want 1", len(resp.Content))
	}
}

func TestFindTransactionDefaultsPageValues(t *testing.T) {
	ctx := middleware.WithUserID(context.Background(), 66)
	logic := NewFindTransactionLogic(ctx, &svc.ServiceContext{
		Config: config.Config{},
		AssetClient: &fakeAssetClient{
			findWalletBySymbolFn: func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWallet, error) {
				return nil, nil
			},
			findWalletFn: func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWalletList, error) {
				return nil, nil
			},
			findTransactionFn: func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberTransactionList, error) {
				if in.PageNo != 1 || in.PageSize != 10 {
					t.Fatalf("FindTransaction() default paging = (%d,%d), want (1,10)", in.PageNo, in.PageSize)
				}
				return &assetpb.MemberTransactionList{}, nil
			},
		},
	})

	if _, err := logic.FindTransaction(&types.AssetReq{}); err != nil {
		t.Fatalf("FindTransaction() error = %v", err)
	}
}
