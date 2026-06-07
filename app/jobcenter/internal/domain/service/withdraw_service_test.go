// Package service 提供领域服务的单元测试。
//
// 测试覆盖：
//   - 正常流程：验证完整的提现处理流程
//   - 错误处理：验证各种错误场景的处理
//   - 恢复机制：验证从缓存恢复的逻辑
//   - 幂等性：验证重复处理的防护
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

// fakeWithdrawRepository 是 withdrawRepository 的 mock 实现，用于测试。
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

// fakeMarketFinder 是 marketCoinFinder 的 mock 实现，用于测试。
type fakeMarketFinder struct {
	findCoinByIDFn func(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.Coin, error)
}

func (f *fakeMarketFinder) FindCoinById(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.Coin, error) {
	return f.findCoinByIDFn(ctx, in, opts...)
}

// fakeAssetFinder 是 assetWalletFinder 的 mock 实现，用于测试。
type fakeAssetFinder struct {
	findWalletFn func(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWallet, error)
}

func (f *fakeAssetFinder) FindWalletBySymbol(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWallet, error) {
	return f.findWalletFn(ctx, in, opts...)
}

// fakeTxCache 是 txCache 的 mock 实现，用于测试。
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

// fakeBitcoinSender 是 btcx.WithdrawSender 的 mock 实现，用于测试。
type fakeBitcoinSender struct {
	sendFn func(ctx context.Context, fromAddress string, toAddress string, totalAmount float64, arrivedAmount float64) (string, error)
}

func (f *fakeBitcoinSender) Send(ctx context.Context, fromAddress string, toAddress string, totalAmount float64, arrivedAmount float64) (string, error) {
	return f.sendFn(ctx, fromAddress, toAddress, totalAmount, arrivedAmount)
}

// TestProcessAppliedMarksWithdrawSuccess 验证正常提现处理流程。
//
// 测试场景：
//   - 提现记录处于 Processing 状态
//   - 币种为 BTC（支持处理）
//   - 用户钱包地址存在
//   - Bitcoin Core 调用成功
//   - 缓存和数据库更新成功
//
// 验证点：
//   - Bitcoin Core 使用正确的发送方和接收方地址
//   - 金额参数正确传递
//   - 缓存键和值正确写入
//   - 数据库状态正确更新
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

// TestProcessAppliedReturnsRetryableWhenRecordNotCommittedYet 验证记录未提交时的重试行为。
//
// 测试场景：
//   - Kafka 消息已到达，但数据库事务尚未提交
//   - FindByID 返回 nil（记录不存在）
//
// 预期行为：
//   - 返回可重试错误（非 NonRetryableError）
//   - Kafka 消费者会重新投递消息
//   - 等待数据库事务提交后重试成功
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

// TestProcessAppliedFinalizesFromCacheBeforeResending 验证从缓存恢复的逻辑。
//
// 测试场景：
//   - 提现记录处于 Processing 状态
//   - 缓存中已存在 txid 和 dealTime
//
// 预期行为：
//   - 从缓存读取 txid 和 dealTime
//   - 直接更新数据库，不重新调用 Bitcoin Core
//   - Market RPC 和 Asset RPC 不被调用
//   - Bitcoin Sender 不被调用
//
// 这是恢复机制的核心测试，确保：
//   - 交易不会重复广播（防止双重支付）
//   - 已获得的 txid 能被正确复用
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

// TestProcessAppliedRejectsUnsupportedCoinAsNonRetryable 验证不支持币种的处理。
//
// 测试场景：
//   - 币种为 ETH（当前仅支持 BTC）
//
// 预期行为：
//   - 返回 NonRetryableError
//   - 不进行重试
//   - 消息发送到死信队列
//
// 设计原因：
//   - 不支持的币种是业务限制，不会因重试而改变
//   - 避免无限重试阻塞消费者
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

// TestProcessAppliedReturnsNilWhenCacheCheckpointFailsButDBFinalizes 验证缓存失败但数据库成功的场景。
//
// 测试场景：
//   - 交易广播成功
//   - Redis 缓存写入失败
//   - MySQL 更新成功
//
// 预期行为：
//   - 返回 nil（成功）
//   - 缓存失败不影响最终结果
//
// 设计原因：
//   - Redis 是辅助检查点，不是必需依赖
//   - MySQL 更新成功才是最终状态
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

// TestProcessAppliedReturnsNonRetryableWhenCheckpointAndDBBothFailAfterBroadcast 验证关键失败场景。
//
// 测试场景：
//   - 交易已广播成功（txid 已获得）
//   - Redis 缓存写入失败
//   - MySQL 更新失败
//
// 预期行为：
//   - 返回 NonRetryableError
//   - 不进行重试
//   - 需要人工介入
//
// 设计原因：
//   - 交易已广播但无法记录状态是严重问题
//   - 重试可能导致重复广播（双重支付风险）
//   - 发送到死信队列供人工处理
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

// TestProcessAppliedReturnsRetryableWhenCheckpointReadFails 验证缓存读取失败的场景。
//
// 测试场景：
//   - Redis 缓存读取失败（非 redis.Nil）
//   - 无法判断是否存在恢复检查点
//
// 预期行为：
//   - 返回可重试错误
//   - 不继续处理（安全策略）
//
// 设计原因：
//   - 缓存读取失败可能导致漏掉恢复检查点
//   - 重试后缓存可能恢复正常
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
