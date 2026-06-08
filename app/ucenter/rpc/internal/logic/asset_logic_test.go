package logic

import (
	"context"
	"errors"
	"testing"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/domain/service"
	"mscoin_go/app/ucenter/rpc/internal/model"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/lock"

	"github.com/alicebob/miniredis/v2"
	"google.golang.org/grpc"
)

type fakeWalletRepo struct {
	findByMemberIDFn            func(ctx context.Context, memberID int64) ([]*model.MemberWallet, error)
	findByMemberIDAndCoinNameFn func(ctx context.Context, memberID int64, coinName string) (*model.MemberWallet, error)
	findAllAddressFn            func(ctx context.Context, coinName string) ([]string, error)
	updateAddressFn             func(ctx context.Context, wallet *model.MemberWallet) error
	saveFn                      func(ctx context.Context, wallet *model.MemberWallet) error
}

func (f *fakeWalletRepo) FindByMemberID(ctx context.Context, memberID int64) ([]*model.MemberWallet, error) {
	return f.findByMemberIDFn(ctx, memberID)
}

func (f *fakeWalletRepo) FindByMemberIDAndCoinName(ctx context.Context, memberID int64, coinName string) (*model.MemberWallet, error) {
	return f.findByMemberIDAndCoinNameFn(ctx, memberID, coinName)
}

func (f *fakeWalletRepo) FindAllAddress(ctx context.Context, coinName string) ([]string, error) {
	return f.findAllAddressFn(ctx, coinName)
}

func (f *fakeWalletRepo) UpdateAddress(ctx context.Context, wallet *model.MemberWallet) error {
	return f.updateAddressFn(ctx, wallet)
}

func (f *fakeWalletRepo) Save(ctx context.Context, wallet *model.MemberWallet) error {
	return f.saveFn(ctx, wallet)
}

type fakeMarketClientForAsset struct {
	findCoinInfoFn func(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.Coin, error)
}

type fakeAddressAllocator struct {
	allocateFn func(ctx context.Context, label string) (string, error)
}

func (f *fakeAddressAllocator) Allocate(ctx context.Context, label string) (string, error) {
	return f.allocateFn(ctx, label)
}

func (f *fakeMarketClientForAsset) FindSymbolThumbTrend(context.Context, *marketpb.MarketReq, ...grpc.CallOption) (*marketpb.SymbolThumbRes, error) {
	return nil, nil
}

func (f *fakeMarketClientForAsset) FindSymbolInfo(context.Context, *marketpb.MarketReq, ...grpc.CallOption) (*marketpb.ExchangeCoin, error) {
	return nil, nil
}

func (f *fakeMarketClientForAsset) FindCoinInfo(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.Coin, error) {
	return f.findCoinInfoFn(ctx, in, opts...)
}

func (f *fakeMarketClientForAsset) FindAllCoin(context.Context, *marketpb.MarketReq, ...grpc.CallOption) (*marketpb.CoinList, error) {
	return nil, nil
}

func (f *fakeMarketClientForAsset) HistoryKline(context.Context, *marketpb.MarketReq, ...grpc.CallOption) (*marketpb.HistoryRes, error) {
	return nil, nil
}

func (f *fakeMarketClientForAsset) FindExchangeCoinVisible(context.Context, *marketpb.MarketReq, ...grpc.CallOption) (*marketpb.ExchangeCoinRes, error) {
	return nil, nil
}

func (f *fakeMarketClientForAsset) FindCoinById(context.Context, *marketpb.MarketReq, ...grpc.CallOption) (*marketpb.Coin, error) {
	return nil, nil
}

func TestResetAddressGeneratesBTCAddressWhenEmpty(t *testing.T) {
	wallet := &model.MemberWallet{
		Id:       1,
		MemberId: 8,
		CoinId:   9,
		CoinName: "BTC",
	}
	updated := false
	repo := &fakeWalletRepo{
		findByMemberIDFn: func(context.Context, int64) ([]*model.MemberWallet, error) { return nil, nil },
		findByMemberIDAndCoinNameFn: func(ctx context.Context, memberID int64, coinName string) (*model.MemberWallet, error) {
			return wallet, nil
		},
		findAllAddressFn: func(context.Context, string) ([]string, error) { return nil, nil },
		updateAddressFn: func(ctx context.Context, payload *model.MemberWallet) error {
			updated = true
			if payload.Address == "" {
				t.Fatal("UpdateAddress() got empty address")
			}
			if payload.AddressPrivateKey != "" {
				t.Fatal("UpdateAddress() should not persist a local private key when Bitcoin Core allocates the address")
			}
			return nil
		},
		saveFn: func(context.Context, *model.MemberWallet) error { return nil },
	}

	logic := NewResetAddressLogic(context.Background(), &svc.ServiceContext{
		WalletService: service.NewWalletService(repo),
		MarketClient: &fakeMarketClientForAsset{
			findCoinInfoFn: func(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.Coin, error) {
				if in.Unit != "BTC" {
					t.Fatalf("FindCoinInfo() unit = %q, want BTC", in.Unit)
				}
				return &marketpb.Coin{Id: 9, Unit: "BTC"}, nil
			},
		},
		AddressAllocator: &fakeAddressAllocator{
			allocateFn: func(ctx context.Context, label string) (string, error) {
				if label != "member-8-btc" {
					t.Fatalf("Allocate() label = %q, want member-8-btc", label)
				}
				return "mnodeAllocatedAddress", nil
			},
		},
	})

	if _, err := logic.ResetAddress(&assetpb.AssetReq{UserId: 8, CoinName: "BTC"}); err != nil {
		t.Fatalf("ResetAddress() error = %v", err)
	}
	if !updated {
		t.Fatal("ResetAddress() did not update wallet address")
	}
	if wallet.Address != "mnodeAllocatedAddress" {
		t.Fatalf("ResetAddress() wallet.Address = %q, want mnodeAllocatedAddress", wallet.Address)
	}
}

func TestGetAddressReturnsAllCoinAddresses(t *testing.T) {
	repo := &fakeWalletRepo{
		findByMemberIDFn:            func(context.Context, int64) ([]*model.MemberWallet, error) { return nil, nil },
		findByMemberIDAndCoinNameFn: func(context.Context, int64, string) (*model.MemberWallet, error) { return nil, nil },
		findAllAddressFn: func(ctx context.Context, coinName string) ([]string, error) {
			if coinName != "BTC" {
				t.Fatalf("GetAllAddress() coinName = %q, want BTC", coinName)
			}
			return []string{"addr-1", "addr-2"}, nil
		},
		updateAddressFn: func(context.Context, *model.MemberWallet) error { return nil },
		saveFn:          func(context.Context, *model.MemberWallet) error { return nil },
	}

	logic := NewGetAddressLogic(context.Background(), &svc.ServiceContext{
		WalletService: service.NewWalletService(repo),
	})

	resp, err := logic.GetAddress(&assetpb.AssetReq{CoinName: "BTC"})
	if err != nil {
		t.Fatalf("GetAddress() error = %v", err)
	}
	if len(resp.List) != 2 {
		t.Fatalf("GetAddress().List len = %d, want 2", len(resp.List))
	}
}

func TestResetAddressFailsWhenDistributedLockIsHeld(t *testing.T) {
	t.Parallel()

	miniRedis, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer miniRedis.Close()

	cache := redisx.New(redisx.Config{Addrs: []string{miniRedis.Addr()}})

	lockKey := resetAddressLockKeyPrefix + "8:BTC"
	competingLock, err := lock.NewRedisLock(
		cache.Raw(),
		lockKey,
		lock.WithTTL(resetAddressLockTTL),
		lock.WithRetry(0, 0),
		lock.WithWatchdog(true),
	)
	if err != nil {
		t.Fatalf("NewRedisLock() error = %v", err)
	}
	defer competingLock.Close()

	if err := competingLock.Lock(context.Background()); err != nil {
		t.Fatalf("competingLock.Lock() error = %v", err)
	}

	allocateCalled := false
	repo := &fakeWalletRepo{
		findByMemberIDFn: func(context.Context, int64) ([]*model.MemberWallet, error) { return nil, nil },
		findByMemberIDAndCoinNameFn: func(ctx context.Context, memberID int64, coinName string) (*model.MemberWallet, error) {
			return &model.MemberWallet{
				Id:       1,
				MemberId: memberID,
				CoinId:   9,
				CoinName: coinName,
			}, nil
		},
		findAllAddressFn: func(context.Context, string) ([]string, error) { return nil, nil },
		updateAddressFn:  func(context.Context, *model.MemberWallet) error { return nil },
		saveFn:           func(context.Context, *model.MemberWallet) error { return nil },
	}

	logic := NewResetAddressLogic(context.Background(), &svc.ServiceContext{
		Cache:         cache,
		WalletService: service.NewWalletService(repo),
		MarketClient: &fakeMarketClientForAsset{
			findCoinInfoFn: func(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.Coin, error) {
				return &marketpb.Coin{Id: 9, Unit: "BTC"}, nil
			},
		},
		AddressAllocator: &fakeAddressAllocator{
			allocateFn: func(ctx context.Context, label string) (string, error) {
				allocateCalled = true
				return "mnodeAllocatedAddress", nil
			},
		},
	})

	_, err = logic.ResetAddress(&assetpb.AssetReq{UserId: 8, CoinName: "BTC"})
	if err == nil {
		t.Fatal("ResetAddress() should fail when distributed lock is held")
	}
	if !errors.Is(err, lock.ErrLockNotAcquired) {
		t.Fatalf("ResetAddress() error = %v, want ErrLockNotAcquired", err)
	}
	if allocateCalled {
		t.Fatal("ResetAddress() should not allocate address when distributed lock acquisition fails")
	}
}
