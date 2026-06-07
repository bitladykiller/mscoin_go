package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"mscoin_go/app/jobcenter/internal/model"
	marketpb "mscoin_go/app/market/rpc/pb/market"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"

	goredis "github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
)

type fakeWithdrawRepository struct {
	findByIDFn    func(ctx context.Context, id int64) (*model.WithdrawRecord, error)
	markSuccessFn func(ctx context.Context, id int64, txID string, dealTime int64) (bool, error)
}

func (f *fakeWithdrawRepository) FindByID(ctx context.Context, id int64) (*model.WithdrawRecord, error) {
	return f.findByIDFn(ctx, id)
}

func (f *fakeWithdrawRepository) MarkSuccess(ctx context.Context, id int64, txID string, dealTime int64) (bool, error) {
	return f.markSuccessFn(ctx, id, txID, dealTime)
}

type fakeMarketFinder struct {
	findCoinByIDFn func(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.Coin, error)
}

func (f *fakeMarketFinder) FindCoinById(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.Coin, error) {
	return f.findCoinByIDFn(ctx, in, opts...)
}

type fakeAssetFinder struct {
	findWalletFn func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWallet, error)
}

func (f *fakeAssetFinder) FindWalletBySymbol(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWallet, error) {
	return f.findWalletFn(ctx, in, opts...)
}

type fakeTxCache struct {
	getFn           func(ctx context.Context, key string, value any) error
	setWithExpireFn func(ctx context.Context, key string, value any, ttl time.Duration) error
}

func (f *fakeTxCache) GetCtx(ctx context.Context, key string, value any) error {
	return f.getFn(ctx, key, value)
}

func (f *fakeTxCache) SetWithExpireCtx(ctx context.Context, key string, value any, ttl time.Duration) error {
	return f.setWithExpireFn(ctx, key, value, ttl)
}

type fakeBitcoinSender struct {
	sendFn func(ctx context.Context, fromAddress string, toAddress string, totalAmount float64, arrivedAmount float64) (string, error)
}

func (f *fakeBitcoinSender) Send(ctx context.Context, fromAddress string, toAddress string, totalAmount float64, arrivedAmount float64) (string, error) {
	return f.sendFn(ctx, fromAddress, toAddress, totalAmount, arrivedAmount)
}

func TestProcessAppliedMarksWithdrawSuccess(t *testing.T) {
	t.Parallel()

	var (
		cachedKey string
		sentFrom  string
		sentTo    string
	)

	service := NewWithdrawService(
		&fakeWithdrawRepository{
			findByIDFn: func(context.Context, int64) (*model.WithdrawRecord, error) {
				return &model.WithdrawRecord{
					Id:            9,
					MemberId:      1001,
					CoinId:        1,
					TotalAmount:   1.2,
					ArrivedAmount: 1.15,
					Address:       "target-address",
					Status:        model.WithdrawStatusProcessing,
				}, nil
			},
			markSuccessFn: func(ctx context.Context, id int64, txID string, dealTime int64) (bool, error) {
				if id != 9 {
					t.Fatalf("MarkSuccess() id = %d, want 9", id)
				}
				if txID != "btc-txid" {
					t.Fatalf("MarkSuccess() txID = %q, want btc-txid", txID)
				}
				if dealTime <= 0 {
					t.Fatalf("MarkSuccess() dealTime = %d, want > 0", dealTime)
				}
				return true, nil
			},
		},
		&fakeMarketFinder{
			findCoinByIDFn: func(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.Coin, error) {
				if in.Id != 1 {
					t.Fatalf("FindCoinById() id = %d, want 1", in.Id)
				}
				return &marketpb.Coin{Unit: "BTC"}, nil
			},
		},
		&fakeAssetFinder{
			findWalletFn: func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWallet, error) {
				if in.UserId != 1001 || in.CoinName != "BTC" {
					t.Fatalf("FindWalletBySymbol() req = %+v, want user=1001 coin=BTC", in)
				}
				return &assetpb.MemberWallet{Address: "source-address"}, nil
			},
		},
		&fakeTxCache{
			getFn: func(ctx context.Context, key string, value any) error {
				return goredis.Nil
			},
			setWithExpireFn: func(ctx context.Context, key string, value any, ttl time.Duration) error {
				cachedKey = key
				entry, ok := value.(WithdrawTxCacheEntry)
				if !ok {
					t.Fatalf("cache value type = %T, want WithdrawTxCacheEntry", value)
				}
				if entry.TxID != "btc-txid" || entry.DealTime <= 0 {
					t.Fatalf("cache entry = %+v, want txid and dealTime", entry)
				}
				if ttl != withdrawTxCacheTTL {
					t.Fatalf("cache ttl = %v, want %v", ttl, withdrawTxCacheTTL)
				}
				return nil
			},
		},
		&fakeBitcoinSender{
			sendFn: func(ctx context.Context, fromAddress string, toAddress string, totalAmount float64, arrivedAmount float64) (string, error) {
				sentFrom = fromAddress
				sentTo = toAddress
				if totalAmount != 1.2 || arrivedAmount != 1.15 {
					t.Fatalf("Send() amounts = (%v,%v), want (1.2,1.15)", totalAmount, arrivedAmount)
				}
				return "btc-txid", nil
			},
		},
	)

	err := service.ProcessApplied(context.Background(), &model.WithdrawRecordEvent{
		Id:       9,
		MemberId: 1001,
		CoinId:   1,
		Address:  "target-address",
	})
	if err != nil {
		t.Fatalf("ProcessApplied() error = %v", err)
	}
	if sentFrom != "source-address" || sentTo != "target-address" {
		t.Fatalf("Send() addresses = (%q,%q), want (source-address,target-address)", sentFrom, sentTo)
	}
	if cachedKey != withdrawTxCacheKey(9) {
		t.Fatalf("cache key = %q, want %q", cachedKey, withdrawTxCacheKey(9))
	}
}

func TestProcessAppliedReturnsRetryableWhenRecordNotCommittedYet(t *testing.T) {
	t.Parallel()

	service := NewWithdrawService(
		&fakeWithdrawRepository{
			findByIDFn: func(context.Context, int64) (*model.WithdrawRecord, error) {
				return nil, nil
			},
			markSuccessFn: func(context.Context, int64, string, int64) (bool, error) {
				return false, nil
			},
		},
		&fakeMarketFinder{},
		&fakeAssetFinder{},
		&fakeTxCache{
			getFn: func(ctx context.Context, key string, value any) error { return goredis.Nil },
			setWithExpireFn: func(ctx context.Context, key string, value any, ttl time.Duration) error {
				return nil
			},
		},
		&fakeBitcoinSender{},
	)

	err := service.ProcessApplied(context.Background(), &model.WithdrawRecordEvent{
		Id:       1,
		MemberId: 2,
		CoinId:   3,
		Address:  "addr",
	})
	if err == nil {
		t.Fatal("ProcessApplied() should fail when the record is not committed yet")
	}
	if IsNonRetryable(err) {
		t.Fatalf("ProcessApplied() error = %v, want retryable", err)
	}
}

func TestProcessAppliedFinalizesFromCacheBeforeResending(t *testing.T) {
	t.Parallel()

	var senderCalled bool
	service := NewWithdrawService(
		&fakeWithdrawRepository{
			findByIDFn: func(context.Context, int64) (*model.WithdrawRecord, error) {
				return &model.WithdrawRecord{
					Id:       12,
					MemberId: 8,
					CoinId:   2,
					Address:  "target",
					Status:   model.WithdrawStatusProcessing,
				}, nil
			},
			markSuccessFn: func(ctx context.Context, id int64, txID string, dealTime int64) (bool, error) {
				if txID != "cached-txid" || dealTime != 1710000000000 {
					t.Fatalf("MarkSuccess() = (%q,%d), want cached tx values", txID, dealTime)
				}
				return true, nil
			},
		},
		&fakeMarketFinder{
			findCoinByIDFn: func(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.Coin, error) {
				t.Fatal("FindCoinById() should not be called when cache already has txid")
				return nil, nil
			},
		},
		&fakeAssetFinder{
			findWalletFn: func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWallet, error) {
				t.Fatal("FindWalletBySymbol() should not be called when cache already has txid")
				return nil, nil
			},
		},
		&fakeTxCache{
			getFn: func(ctx context.Context, key string, value any) error {
				entry := value.(*WithdrawTxCacheEntry)
				*entry = WithdrawTxCacheEntry{
					TxID:     "cached-txid",
					DealTime: 1710000000000,
				}
				return nil
			},
			setWithExpireFn: func(ctx context.Context, key string, value any, ttl time.Duration) error {
				t.Fatal("SetWithExpireCtx() should not be called when cache already has txid")
				return nil
			},
		},
		&fakeBitcoinSender{
			sendFn: func(ctx context.Context, fromAddress string, toAddress string, totalAmount float64, arrivedAmount float64) (string, error) {
				senderCalled = true
				return "", nil
			},
		},
	)

	err := service.ProcessApplied(context.Background(), &model.WithdrawRecordEvent{
		Id:       12,
		MemberId: 8,
		CoinId:   2,
		Address:  "target",
	})
	if err != nil {
		t.Fatalf("ProcessApplied() error = %v", err)
	}
	if senderCalled {
		t.Fatal("ProcessApplied() should not resend the bitcoin transaction when cache already has the txid")
	}
}

func TestProcessAppliedRejectsUnsupportedCoinAsNonRetryable(t *testing.T) {
	t.Parallel()

	service := NewWithdrawService(
		&fakeWithdrawRepository{
			findByIDFn: func(context.Context, int64) (*model.WithdrawRecord, error) {
				return &model.WithdrawRecord{
					Id:       7,
					MemberId: 9,
					CoinId:   5,
					Address:  "target",
					Status:   model.WithdrawStatusProcessing,
				}, nil
			},
			markSuccessFn: func(context.Context, int64, string, int64) (bool, error) {
				return false, nil
			},
		},
		&fakeMarketFinder{
			findCoinByIDFn: func(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.Coin, error) {
				return &marketpb.Coin{Unit: "ETH"}, nil
			},
		},
		&fakeAssetFinder{},
		&fakeTxCache{
			getFn: func(ctx context.Context, key string, value any) error { return goredis.Nil },
			setWithExpireFn: func(ctx context.Context, key string, value any, ttl time.Duration) error {
				return nil
			},
		},
		&fakeBitcoinSender{},
	)

	err := service.ProcessApplied(context.Background(), &model.WithdrawRecordEvent{
		Id:       7,
		MemberId: 9,
		CoinId:   5,
		Address:  "target",
	})
	if err == nil {
		t.Fatal("ProcessApplied() should fail for unsupported coins")
	}
	if !IsNonRetryable(err) {
		t.Fatalf("ProcessApplied() error = %v, want non-retryable", err)
	}
}

func TestProcessAppliedReturnsNilWhenCacheCheckpointFailsButDBFinalizes(t *testing.T) {
	t.Parallel()

	service := NewWithdrawService(
		&fakeWithdrawRepository{
			findByIDFn: func(context.Context, int64) (*model.WithdrawRecord, error) {
				return &model.WithdrawRecord{
					Id:            31,
					MemberId:      1002,
					CoinId:        1,
					TotalAmount:   2.0,
					ArrivedAmount: 1.9,
					Address:       "target-address",
					Status:        model.WithdrawStatusProcessing,
				}, nil
			},
			markSuccessFn: func(ctx context.Context, id int64, txID string, dealTime int64) (bool, error) {
				return true, nil
			},
		},
		&fakeMarketFinder{
			findCoinByIDFn: func(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.Coin, error) {
				return &marketpb.Coin{Unit: "BTC"}, nil
			},
		},
		&fakeAssetFinder{
			findWalletFn: func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWallet, error) {
				return &assetpb.MemberWallet{Address: "source-address"}, nil
			},
		},
		&fakeTxCache{
			getFn: func(ctx context.Context, key string, value any) error { return goredis.Nil },
			setWithExpireFn: func(ctx context.Context, key string, value any, ttl time.Duration) error {
				return errors.New("redis unavailable")
			},
		},
		&fakeBitcoinSender{
			sendFn: func(ctx context.Context, fromAddress string, toAddress string, totalAmount float64, arrivedAmount float64) (string, error) {
				return "btc-txid", nil
			},
		},
	)

	if err := service.ProcessApplied(context.Background(), &model.WithdrawRecordEvent{
		Id:       31,
		MemberId: 1002,
		CoinId:   1,
		Address:  "target-address",
	}); err != nil {
		t.Fatalf("ProcessApplied() error = %v, want nil when mysql finalization succeeds", err)
	}
}

func TestProcessAppliedReturnsNonRetryableWhenCheckpointAndDBBothFailAfterBroadcast(t *testing.T) {
	t.Parallel()

	service := NewWithdrawService(
		&fakeWithdrawRepository{
			findByIDFn: func(context.Context, int64) (*model.WithdrawRecord, error) {
				return &model.WithdrawRecord{
					Id:            32,
					MemberId:      1003,
					CoinId:        1,
					TotalAmount:   2.0,
					ArrivedAmount: 1.9,
					Address:       "target-address",
					Status:        model.WithdrawStatusProcessing,
				}, nil
			},
			markSuccessFn: func(ctx context.Context, id int64, txID string, dealTime int64) (bool, error) {
				return false, errors.New("mysql unavailable")
			},
		},
		&fakeMarketFinder{
			findCoinByIDFn: func(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.Coin, error) {
				return &marketpb.Coin{Unit: "BTC"}, nil
			},
		},
		&fakeAssetFinder{
			findWalletFn: func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWallet, error) {
				return &assetpb.MemberWallet{Address: "source-address"}, nil
			},
		},
		&fakeTxCache{
			getFn: func(ctx context.Context, key string, value any) error { return goredis.Nil },
			setWithExpireFn: func(ctx context.Context, key string, value any, ttl time.Duration) error {
				return errors.New("redis unavailable")
			},
		},
		&fakeBitcoinSender{
			sendFn: func(ctx context.Context, fromAddress string, toAddress string, totalAmount float64, arrivedAmount float64) (string, error) {
				return "btc-txid", nil
			},
		},
	)

	err := service.ProcessApplied(context.Background(), &model.WithdrawRecordEvent{
		Id:       32,
		MemberId: 1003,
		CoinId:   1,
		Address:  "target-address",
	})
	if err == nil {
		t.Fatal("ProcessApplied() should fail when checkpoint and mysql finalization both fail after broadcast")
	}
	if !IsNonRetryable(err) {
		t.Fatalf("ProcessApplied() error = %v, want non-retryable", err)
	}
}

func TestProcessAppliedReturnsRetryableWhenCheckpointReadFails(t *testing.T) {
	t.Parallel()

	service := NewWithdrawService(
		&fakeWithdrawRepository{
			findByIDFn: func(context.Context, int64) (*model.WithdrawRecord, error) {
				return &model.WithdrawRecord{
					Id:       33,
					MemberId: 1004,
					CoinId:   1,
					Address:  "target-address",
					Status:   model.WithdrawStatusProcessing,
				}, nil
			},
			markSuccessFn: func(ctx context.Context, id int64, txID string, dealTime int64) (bool, error) {
				return false, nil
			},
		},
		&fakeMarketFinder{},
		&fakeAssetFinder{},
		&fakeTxCache{
			getFn: func(ctx context.Context, key string, value any) error { return errors.New("redis down") },
			setWithExpireFn: func(ctx context.Context, key string, value any, ttl time.Duration) error {
				return nil
			},
		},
		&fakeBitcoinSender{},
	)

	err := service.ProcessApplied(context.Background(), &model.WithdrawRecordEvent{
		Id:       33,
		MemberId: 1004,
		CoinId:   1,
		Address:  "target-address",
	})
	if err == nil {
		t.Fatal("ProcessApplied() should fail when checkpoint read fails")
	}
	if IsNonRetryable(err) {
		t.Fatalf("ProcessApplied() error = %v, want retryable", err)
	}
}
