package logic

import (
	"context"
	"testing"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/api/internal/config"
	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"

	"google.golang.org/grpc"
)

type fakeWithdrawClient struct {
	findAddressByCoinIDFn func(ctx context.Context, in *withdrawpb.WithdrawReq, opts ...grpc.CallOption) (*withdrawpb.AddressSimpleList, error)
	sendCodeFn            func(ctx context.Context, in *withdrawpb.WithdrawReq, opts ...grpc.CallOption) (*withdrawpb.NoRes, error)
	withdrawCodeFn        func(ctx context.Context, in *withdrawpb.WithdrawReq, opts ...grpc.CallOption) (*withdrawpb.NoRes, error)
	withdrawRecordFn      func(ctx context.Context, in *withdrawpb.WithdrawReq, opts ...grpc.CallOption) (*withdrawpb.RecordList, error)
}

func (f *fakeWithdrawClient) FindAddressByCoinId(ctx context.Context, in *withdrawpb.WithdrawReq, opts ...grpc.CallOption) (*withdrawpb.AddressSimpleList, error) {
	return f.findAddressByCoinIDFn(ctx, in, opts...)
}

func (f *fakeWithdrawClient) SendCode(ctx context.Context, in *withdrawpb.WithdrawReq, opts ...grpc.CallOption) (*withdrawpb.NoRes, error) {
	return f.sendCodeFn(ctx, in, opts...)
}

func (f *fakeWithdrawClient) WithdrawCode(ctx context.Context, in *withdrawpb.WithdrawReq, opts ...grpc.CallOption) (*withdrawpb.NoRes, error) {
	return f.withdrawCodeFn(ctx, in, opts...)
}

func (f *fakeWithdrawClient) WithdrawRecord(ctx context.Context, in *withdrawpb.WithdrawReq, opts ...grpc.CallOption) (*withdrawpb.RecordList, error) {
	return f.withdrawRecordFn(ctx, in, opts...)
}

type fakeMarketClientForWithdraw struct {
	findAllCoinFn func(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.CoinList, error)
}

func (f *fakeMarketClientForWithdraw) FindSymbolThumbTrend(context.Context, *marketpb.MarketReq, ...grpc.CallOption) (*marketpb.SymbolThumbRes, error) {
	return nil, nil
}

func (f *fakeMarketClientForWithdraw) FindSymbolInfo(context.Context, *marketpb.MarketReq, ...grpc.CallOption) (*marketpb.ExchangeCoin, error) {
	return nil, nil
}

func (f *fakeMarketClientForWithdraw) FindCoinInfo(context.Context, *marketpb.MarketReq, ...grpc.CallOption) (*marketpb.Coin, error) {
	return nil, nil
}

func (f *fakeMarketClientForWithdraw) FindAllCoin(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.CoinList, error) {
	return f.findAllCoinFn(ctx, in, opts...)
}

func (f *fakeMarketClientForWithdraw) HistoryKline(context.Context, *marketpb.MarketReq, ...grpc.CallOption) (*marketpb.HistoryRes, error) {
	return nil, nil
}

func (f *fakeMarketClientForWithdraw) FindExchangeCoinVisible(context.Context, *marketpb.MarketReq, ...grpc.CallOption) (*marketpb.ExchangeCoinRes, error) {
	return nil, nil
}

func (f *fakeMarketClientForWithdraw) FindCoinById(context.Context, *marketpb.MarketReq, ...grpc.CallOption) (*marketpb.Coin, error) {
	return nil, nil
}

func TestQueryWithdrawCoinMapsWalletsAndAddresses(t *testing.T) {
	ctx := middleware.WithUserID(context.Background(), 77)
	logic := NewQueryWithdrawCoinLogic(ctx, &svc.ServiceContext{
		Config: config.Config{},
		AssetClient: &fakeAssetClient{
			findWalletBySymbolFn: func(context.Context, *assetpb.AssetReq, ...grpc.CallOption) (*assetpb.MemberWallet, error) {
				return nil, nil
			},
			findWalletFn: func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWalletList, error) {
				if in.UserId != 77 {
					t.Fatalf("FindWallet() userId = %d, want 77", in.UserId)
				}
				return &assetpb.MemberWalletList{
					List: []*assetpb.MemberWallet{
						{
							Balance: 12.5,
							Coin:    &assetpb.Coin{Unit: "BTC"},
						},
					},
				}, nil
			},
			findTransactionFn: func(context.Context, *assetpb.AssetReq, ...grpc.CallOption) (*assetpb.MemberTransactionList, error) {
				return nil, nil
			},
		},
		MarketClient: &fakeMarketClientForWithdraw{
			findAllCoinFn: func(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.CoinList, error) {
				return &marketpb.CoinList{
					List: []*marketpb.Coin{
						{
							Id:                9,
							Name:              "Bitcoin",
							NameCn:            "比特币",
							Unit:              "BTC",
							MinTxFee:          0.01,
							MaxTxFee:          1.2,
							MinWithdrawAmount: 0.1,
							MaxWithdrawAmount: 99,
							WithdrawThreshold: 0.5,
							WithdrawScale:     8,
							AccountType:       1,
							CanAutoWithdraw:   0,
						},
					},
				}, nil
			},
		},
		WithdrawClient: &fakeWithdrawClient{
			findAddressByCoinIDFn: func(ctx context.Context, in *withdrawpb.WithdrawReq, opts ...grpc.CallOption) (*withdrawpb.AddressSimpleList, error) {
				if in.UserId != 77 || in.CoinId != 9 {
					t.Fatalf("FindAddressByCoinId() payload = %+v, want userId=77 coinId=9", in)
				}
				return &withdrawpb.AddressSimpleList{
					List: []*withdrawpb.AddressSimple{
						{Remark: "main", Address: "1btc"},
					},
				}, nil
			},
			sendCodeFn: func(context.Context, *withdrawpb.WithdrawReq, ...grpc.CallOption) (*withdrawpb.NoRes, error) {
				return nil, nil
			},
			withdrawCodeFn: func(context.Context, *withdrawpb.WithdrawReq, ...grpc.CallOption) (*withdrawpb.NoRes, error) {
				return nil, nil
			},
			withdrawRecordFn: func(context.Context, *withdrawpb.WithdrawReq, ...grpc.CallOption) (*withdrawpb.RecordList, error) {
				return nil, nil
			},
		},
	})

	resp, err := logic.QueryWithdrawCoin()
	if err != nil {
		t.Fatalf("QueryWithdrawCoin() error = %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("QueryWithdrawCoin() len = %d, want 1", len(resp))
	}
	if resp[0].CanAutoWithdraw != "true" {
		t.Fatalf("QueryWithdrawCoin().CanAutoWithdraw = %q, want true", resp[0].CanAutoWithdraw)
	}
	if len(resp[0].Addresses) != 1 || resp[0].Addresses[0].Address != "1btc" {
		t.Fatalf("QueryWithdrawCoin().Addresses = %+v, want mapped addresses", resp[0].Addresses)
	}
}

func TestSendWithdrawCodeUsesMemberPhone(t *testing.T) {
	ctx := middleware.WithUserID(context.Background(), 99)
	logic := NewSendWithdrawCodeLogic(ctx, &svc.ServiceContext{
		Config: config.Config{},
		MemberClient: &fakeMemberClient{
			findMemberByIDFn: func(ctx context.Context, in *memberpb.MemberReq, opts ...grpc.CallOption) (*memberpb.MemberInfo, error) {
				if in.MemberId != 99 {
					t.Fatalf("FindMemberById() memberId = %d, want 99", in.MemberId)
				}
				return &memberpb.MemberInfo{MobilePhone: "13800000000"}, nil
			},
		},
		WithdrawClient: &fakeWithdrawClient{
			findAddressByCoinIDFn: func(context.Context, *withdrawpb.WithdrawReq, ...grpc.CallOption) (*withdrawpb.AddressSimpleList, error) {
				return nil, nil
			},
			sendCodeFn: func(ctx context.Context, in *withdrawpb.WithdrawReq, opts ...grpc.CallOption) (*withdrawpb.NoRes, error) {
				if in.Phone != "13800000000" {
					t.Fatalf("SendCode() phone = %q, want 13800000000", in.Phone)
				}
				return &withdrawpb.NoRes{}, nil
			},
			withdrawCodeFn: func(context.Context, *withdrawpb.WithdrawReq, ...grpc.CallOption) (*withdrawpb.NoRes, error) {
				return nil, nil
			},
			withdrawRecordFn: func(context.Context, *withdrawpb.WithdrawReq, ...grpc.CallOption) (*withdrawpb.RecordList, error) {
				return nil, nil
			},
		},
	})

	resp, err := logic.SendCode()
	if err != nil {
		t.Fatalf("SendCode() error = %v", err)
	}
	if resp != "success" {
		t.Fatalf("SendCode() = %q, want success", resp)
	}
}

func TestWithdrawRecordDefaultsPageValues(t *testing.T) {
	ctx := middleware.WithUserID(context.Background(), 55)
	logic := NewWithdrawRecordLogic(ctx, &svc.ServiceContext{
		Config: config.Config{},
		WithdrawClient: &fakeWithdrawClient{
			findAddressByCoinIDFn: func(context.Context, *withdrawpb.WithdrawReq, ...grpc.CallOption) (*withdrawpb.AddressSimpleList, error) {
				return nil, nil
			},
			sendCodeFn: func(context.Context, *withdrawpb.WithdrawReq, ...grpc.CallOption) (*withdrawpb.NoRes, error) {
				return nil, nil
			},
			withdrawCodeFn: func(context.Context, *withdrawpb.WithdrawReq, ...grpc.CallOption) (*withdrawpb.NoRes, error) {
				return nil, nil
			},
			withdrawRecordFn: func(ctx context.Context, in *withdrawpb.WithdrawReq, opts ...grpc.CallOption) (*withdrawpb.RecordList, error) {
				if in.UserId != 55 || in.Page != 1 || in.PageSize != 10 {
					t.Fatalf("WithdrawRecord() payload = %+v, want userId=55 page=1 pageSize=10", in)
				}
				return &withdrawpb.RecordList{
					List:  []*withdrawpb.WithdrawRecord{{Id: 1}},
					Total: 1,
				}, nil
			},
		},
	})

	resp, err := logic.Record(&types.WithdrawReq{})
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if resp.TotalElements != 1 || len(resp.Content) != 1 {
		t.Fatalf("Record() = %+v, want one paged item", resp)
	}
}

func TestWithdrawCodeMapsRequestToRPC(t *testing.T) {
	ctx := middleware.WithUserID(context.Background(), 66)
	logic := NewWithdrawCodeLogic(ctx, &svc.ServiceContext{
		Config: config.Config{},
		WithdrawClient: &fakeWithdrawClient{
			findAddressByCoinIDFn: func(context.Context, *withdrawpb.WithdrawReq, ...grpc.CallOption) (*withdrawpb.AddressSimpleList, error) {
				return nil, nil
			},
			sendCodeFn: func(context.Context, *withdrawpb.WithdrawReq, ...grpc.CallOption) (*withdrawpb.NoRes, error) {
				return nil, nil
			},
			withdrawCodeFn: func(ctx context.Context, in *withdrawpb.WithdrawReq, opts ...grpc.CallOption) (*withdrawpb.NoRes, error) {
				if in.UserId != 66 || in.Unit != "BTC" || in.Code != "123456" {
					t.Fatalf("WithdrawCode() payload = %+v, want mapped request", in)
				}
				return &withdrawpb.NoRes{}, nil
			},
			withdrawRecordFn: func(context.Context, *withdrawpb.WithdrawReq, ...grpc.CallOption) (*withdrawpb.RecordList, error) {
				return nil, nil
			},
		},
	})

	resp, err := logic.WithdrawCode(&types.WithdrawReq{
		Unit:       "BTC",
		Address:    "1btc",
		Amount:     12.5,
		Fee:        0.2,
		JyPassword: "secret",
		Code:       "123456",
	})
	if err != nil {
		t.Fatalf("WithdrawCode() error = %v", err)
	}
	if resp != "success" {
		t.Fatalf("WithdrawCode() = %q, want success", resp)
	}
}
