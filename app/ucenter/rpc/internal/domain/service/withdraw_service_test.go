package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"mscoin_go/app/ucenter/rpc/internal/model"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
	"mscoin_go/pkg/db/mysqlx"
)

type fakeWithdrawMemberRepo struct {
	findByIDFn func(ctx context.Context, memberID int64) (*model.Member, error)
}

func (f *fakeWithdrawMemberRepo) FindByID(ctx context.Context, memberID int64) (*model.Member, error) {
	return f.findByIDFn(ctx, memberID)
}

type fakeWithdrawWalletRepo struct {
	findByMemberIDAndCoinNameFn          func(ctx context.Context, memberID int64, coinName string) (*model.MemberWallet, error)
	findByMemberIDAndCoinNameForUpdateFn func(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string) (*model.MemberWallet, error)
	freezeBalanceFn                      func(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string, amount float64) error
}

func (f *fakeWithdrawWalletRepo) FindByMemberIDAndCoinName(ctx context.Context, memberID int64, coinName string) (*model.MemberWallet, error) {
	return f.findByMemberIDAndCoinNameFn(ctx, memberID, coinName)
}

func (f *fakeWithdrawWalletRepo) FindByMemberIDAndCoinNameForUpdate(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string) (*model.MemberWallet, error) {
	return f.findByMemberIDAndCoinNameForUpdateFn(ctx, exec, memberID, coinName)
}

func (f *fakeWithdrawWalletRepo) FreezeBalance(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string, amount float64) error {
	return f.freezeBalanceFn(ctx, exec, memberID, coinName, amount)
}

type fakeWithdrawAddressRepo struct {
	findByMemberIDAndCoinIDFn func(ctx context.Context, memberID int64, coinID int64) ([]*model.MemberAddress, error)
}

func (f *fakeWithdrawAddressRepo) FindByMemberIDAndCoinID(ctx context.Context, memberID int64, coinID int64) ([]*model.MemberAddress, error) {
	return f.findByMemberIDAndCoinIDFn(ctx, memberID, coinID)
}

type fakeWithdrawRecordRepo struct {
	findByMemberIDFn func(ctx context.Context, memberID int64, page int64, pageSize int64) ([]*model.WithdrawRecord, int64, error)
	saveFn           func(ctx context.Context, exec mysqlx.ExtContext, record *model.WithdrawRecord) error
}

func (f *fakeWithdrawRecordRepo) FindByMemberID(ctx context.Context, memberID int64, page int64, pageSize int64) ([]*model.WithdrawRecord, int64, error) {
	return f.findByMemberIDFn(ctx, memberID, page, pageSize)
}

func (f *fakeWithdrawRecordRepo) Save(ctx context.Context, exec mysqlx.ExtContext, record *model.WithdrawRecord) error {
	return f.saveFn(ctx, exec, record)
}

type fakeWithdrawCache struct {
	getCtxFn           func(ctx context.Context, key string, value any) error
	setWithExpireCtxFn func(ctx context.Context, key string, value any, ttl time.Duration) error
}

func (f *fakeWithdrawCache) GetCtx(ctx context.Context, key string, value any) error {
	return f.getCtxFn(ctx, key, value)
}

func (f *fakeWithdrawCache) SetWithExpireCtx(ctx context.Context, key string, value any, ttl time.Duration) error {
	return f.setWithExpireCtxFn(ctx, key, value, ttl)
}

type fakeTxManager struct {
	withinTxFn func(ctx context.Context, fn func(exec mysqlx.ExtContext) error) error
}

func (f *fakeTxManager) WithinTx(ctx context.Context, fn func(exec mysqlx.ExtContext) error) error {
	return f.withinTxFn(ctx, fn)
}

type fakeKafkaProducer struct {
	pushWithKeyFn func(ctx context.Context, key string, value string) error
	closeFn       func() error
}

func (f *fakeKafkaProducer) PushWithKey(ctx context.Context, key string, value string) error {
	return f.pushWithKeyFn(ctx, key, value)
}

func (f *fakeKafkaProducer) Close() error {
	if f.closeFn != nil {
		return f.closeFn()
	}
	return nil
}

func TestWithdrawServiceApplySuccess(t *testing.T) {
	t.Parallel()

	var frozeAmount float64
	var savedRecord *model.WithdrawRecord
	var publishedRecord model.WithdrawRecord
	txCalls := 0

	svc := NewWithdrawService(
		&fakeWithdrawMemberRepo{
			findByIDFn: func(ctx context.Context, memberID int64) (*model.Member, error) {
				return &model.Member{Id: memberID, MobilePhone: "13800000000", JyPassword: "secret"}, nil
			},
		},
		&fakeWithdrawWalletRepo{
			findByMemberIDAndCoinNameFn: func(context.Context, int64, string) (*model.MemberWallet, error) {
				return nil, nil
			},
			findByMemberIDAndCoinNameForUpdateFn: func(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string) (*model.MemberWallet, error) {
				if memberID != 9 || coinName != "BTC" {
					t.Fatalf("FindByMemberIDAndCoinNameForUpdate() got memberID=%d coinName=%q", memberID, coinName)
				}
				return &model.MemberWallet{CoinId: 5, Balance: 15}, nil
			},
			freezeBalanceFn: func(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string, amount float64) error {
				frozeAmount = amount
				return nil
			},
		},
		&fakeWithdrawAddressRepo{
			findByMemberIDAndCoinIDFn: func(context.Context, int64, int64) ([]*model.MemberAddress, error) {
				return nil, nil
			},
		},
		&fakeWithdrawRecordRepo{
			findByMemberIDFn: func(context.Context, int64, int64, int64) ([]*model.WithdrawRecord, int64, error) {
				return nil, 0, nil
			},
			saveFn: func(ctx context.Context, exec mysqlx.ExtContext, record *model.WithdrawRecord) error {
				savedRecord = record
				record.Id = 88
				return nil
			},
		},
		&fakeWithdrawCache{
			getCtxFn: func(ctx context.Context, key string, value any) error {
				if key != withdrawCacheKey+"13800000000" {
					t.Fatalf("GetCtx() key = %q, want withdraw cache key", key)
				}
				target, ok := value.(*string)
				if !ok {
					t.Fatalf("GetCtx() target type = %T, want *string", value)
				}
				*target = "123456"
				return nil
			},
			setWithExpireCtxFn: func(context.Context, string, any, time.Duration) error {
				return nil
			},
		},
		&fakeTxManager{
			withinTxFn: func(ctx context.Context, fn func(exec mysqlx.ExtContext) error) error {
				txCalls++
				return fn(nil)
			},
		},
		&fakeKafkaProducer{
			pushWithKeyFn: func(ctx context.Context, key string, value string) error {
				if key != "9" {
					t.Fatalf("PushWithKey() key = %q, want 9", key)
				}
				if err := json.Unmarshal([]byte(value), &publishedRecord); err != nil {
					t.Fatalf("PushWithKey() payload unmarshal error = %v", err)
				}
				return nil
			},
		},
	)

	err := svc.Apply(context.Background(), &withdrawpb.WithdrawReq{
		UserId:     9,
		Unit:       "BTC",
		Address:    "tb1-destination",
		Amount:     10.2,
		Fee:        0.2,
		JyPassword: "secret",
		Code:       "123456",
	})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if txCalls != 1 {
		t.Fatalf("Apply() txCalls = %d, want 1", txCalls)
	}
	if frozeAmount != 10.2 {
		t.Fatalf("Apply() frozeAmount = %v, want 10.2", frozeAmount)
	}
	if savedRecord == nil {
		t.Fatal("Apply() did not save withdraw record")
	}
	if savedRecord.Status != model.WithdrawStatusProcessing {
		t.Fatalf("Apply() saved status = %d, want %d", savedRecord.Status, model.WithdrawStatusProcessing)
	}
	if savedRecord.ArrivedAmount != 10 {
		t.Fatalf("Apply() arrived amount = %v, want 10", savedRecord.ArrivedAmount)
	}
	if publishedRecord.Id != 88 || publishedRecord.MemberId != 9 {
		t.Fatalf("Apply() published record = %+v, want persisted record payload", publishedRecord)
	}
}

func TestWithdrawServiceApplyVerificationCodeMismatch(t *testing.T) {
	t.Parallel()

	txCalled := false
	queueCalled := false

	svc := NewWithdrawService(
		&fakeWithdrawMemberRepo{
			findByIDFn: func(ctx context.Context, memberID int64) (*model.Member, error) {
				return &model.Member{Id: memberID, MobilePhone: "13800000000", JyPassword: "secret"}, nil
			},
		},
		&fakeWithdrawWalletRepo{},
		&fakeWithdrawAddressRepo{},
		&fakeWithdrawRecordRepo{},
		&fakeWithdrawCache{
			getCtxFn: func(ctx context.Context, key string, value any) error {
				*(value.(*string)) = "999999"
				return nil
			},
			setWithExpireCtxFn: func(context.Context, string, any, time.Duration) error {
				return nil
			},
		},
		&fakeTxManager{
			withinTxFn: func(ctx context.Context, fn func(exec mysqlx.ExtContext) error) error {
				txCalled = true
				return fn(nil)
			},
		},
		&fakeKafkaProducer{
			pushWithKeyFn: func(ctx context.Context, key string, value string) error {
				queueCalled = true
				return nil
			},
		},
	)

	err := svc.Apply(context.Background(), &withdrawpb.WithdrawReq{
		UserId:     9,
		Unit:       "BTC",
		Address:    "tb1-destination",
		Amount:     10,
		Fee:        0.1,
		JyPassword: "secret",
		Code:       "123456",
	})
	if err == nil || !strings.Contains(err.Error(), "verification code mismatch") {
		t.Fatalf("Apply() error = %v, want verification code mismatch", err)
	}
	if txCalled {
		t.Fatal("Apply() should not enter transaction on code mismatch")
	}
	if queueCalled {
		t.Fatal("Apply() should not publish on code mismatch")
	}
}

func TestWithdrawServiceApplyInsufficientBalance(t *testing.T) {
	t.Parallel()

	freezeCalled := false
	queueCalled := false

	svc := NewWithdrawService(
		&fakeWithdrawMemberRepo{
			findByIDFn: func(ctx context.Context, memberID int64) (*model.Member, error) {
				return &model.Member{Id: memberID, MobilePhone: "13800000000", JyPassword: "secret"}, nil
			},
		},
		&fakeWithdrawWalletRepo{
			findByMemberIDAndCoinNameFn: func(context.Context, int64, string) (*model.MemberWallet, error) {
				return nil, nil
			},
			findByMemberIDAndCoinNameForUpdateFn: func(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string) (*model.MemberWallet, error) {
				return &model.MemberWallet{CoinId: 5, Balance: 1}, nil
			},
			freezeBalanceFn: func(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string, amount float64) error {
				freezeCalled = true
				return nil
			},
		},
		&fakeWithdrawAddressRepo{},
		&fakeWithdrawRecordRepo{
			saveFn: func(ctx context.Context, exec mysqlx.ExtContext, record *model.WithdrawRecord) error {
				t.Fatal("Save() should not be called when balance is insufficient")
				return nil
			},
		},
		&fakeWithdrawCache{
			getCtxFn: func(ctx context.Context, key string, value any) error {
				*(value.(*string)) = "123456"
				return nil
			},
			setWithExpireCtxFn: func(context.Context, string, any, time.Duration) error {
				return nil
			},
		},
		&fakeTxManager{
			withinTxFn: func(ctx context.Context, fn func(exec mysqlx.ExtContext) error) error {
				return fn(nil)
			},
		},
		&fakeKafkaProducer{
			pushWithKeyFn: func(ctx context.Context, key string, value string) error {
				queueCalled = true
				return nil
			},
		},
	)

	err := svc.Apply(context.Background(), &withdrawpb.WithdrawReq{
		UserId:     9,
		Unit:       "BTC",
		Address:    "tb1-destination",
		Amount:     10,
		Fee:        0.1,
		JyPassword: "secret",
		Code:       "123456",
	})
	if err == nil || !strings.Contains(err.Error(), "insufficient balance") {
		t.Fatalf("Apply() error = %v, want insufficient balance", err)
	}
	if freezeCalled {
		t.Fatal("Apply() should not freeze balance when funds are insufficient")
	}
	if queueCalled {
		t.Fatal("Apply() should not publish when funds are insufficient")
	}
}

func TestWithdrawServiceApplyPublishError(t *testing.T) {
	t.Parallel()

	saveCalled := false

	svc := NewWithdrawService(
		&fakeWithdrawMemberRepo{
			findByIDFn: func(ctx context.Context, memberID int64) (*model.Member, error) {
				return &model.Member{Id: memberID, MobilePhone: "13800000000", JyPassword: "secret"}, nil
			},
		},
		&fakeWithdrawWalletRepo{
			findByMemberIDAndCoinNameFn: func(context.Context, int64, string) (*model.MemberWallet, error) {
				return nil, nil
			},
			findByMemberIDAndCoinNameForUpdateFn: func(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string) (*model.MemberWallet, error) {
				return &model.MemberWallet{CoinId: 5, Balance: 15}, nil
			},
			freezeBalanceFn: func(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string, amount float64) error {
				return nil
			},
		},
		&fakeWithdrawAddressRepo{},
		&fakeWithdrawRecordRepo{
			saveFn: func(ctx context.Context, exec mysqlx.ExtContext, record *model.WithdrawRecord) error {
				saveCalled = true
				return nil
			},
		},
		&fakeWithdrawCache{
			getCtxFn: func(ctx context.Context, key string, value any) error {
				*(value.(*string)) = "123456"
				return nil
			},
			setWithExpireCtxFn: func(context.Context, string, any, time.Duration) error {
				return nil
			},
		},
		&fakeTxManager{
			withinTxFn: func(ctx context.Context, fn func(exec mysqlx.ExtContext) error) error {
				return fn(nil)
			},
		},
		&fakeKafkaProducer{
			pushWithKeyFn: func(ctx context.Context, key string, value string) error {
				return errors.New("boom")
			},
		},
	)

	err := svc.Apply(context.Background(), &withdrawpb.WithdrawReq{
		UserId:     9,
		Unit:       "BTC",
		Address:    "tb1-destination",
		Amount:     10,
		Fee:        0.1,
		JyPassword: "secret",
		Code:       "123456",
	})
	if err == nil || !strings.Contains(err.Error(), "publish withdraw event") {
		t.Fatalf("Apply() error = %v, want publish withdraw event failure", err)
	}
	if !saveCalled {
		t.Fatal("Apply() should save record before publish attempt")
	}
}

func TestWithdrawServiceSendCodeStoresSixDigitCode(t *testing.T) {
	t.Parallel()

	var capturedKey string
	var capturedCode string
	var capturedTTL time.Duration

	svc := NewWithdrawService(
		&fakeWithdrawMemberRepo{},
		&fakeWithdrawWalletRepo{},
		&fakeWithdrawAddressRepo{},
		&fakeWithdrawRecordRepo{},
		&fakeWithdrawCache{
			getCtxFn: func(context.Context, string, any) error {
				return nil
			},
			setWithExpireCtxFn: func(ctx context.Context, key string, value any, ttl time.Duration) error {
				capturedKey = key
				capturedTTL = ttl
				capturedCode = value.(string)
				return nil
			},
		},
		&fakeTxManager{},
		&fakeKafkaProducer{},
	)

	if err := svc.SendCode(context.Background(), "13800000000"); err != nil {
		t.Fatalf("SendCode() error = %v", err)
	}
	if capturedKey != withdrawCacheKey+"13800000000" {
		t.Fatalf("SendCode() key = %q, want cached withdraw key", capturedKey)
	}
	if len(capturedCode) != 6 {
		t.Fatalf("SendCode() code len = %d, want 6", len(capturedCode))
	}
	if capturedTTL != 5*time.Minute {
		t.Fatalf("SendCode() ttl = %v, want 5m", capturedTTL)
	}
}
