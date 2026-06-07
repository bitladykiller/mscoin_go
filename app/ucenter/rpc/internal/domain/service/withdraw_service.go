package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/model"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
	"mscoin_go/pkg/db/mysqlx"
	"mscoin_go/pkg/mq/kafka"
)

const withdrawCacheKey = "WITHDRAW::"

type withdrawMemberRepository interface {
	FindByID(ctx context.Context, memberID int64) (*model.Member, error)
}

type withdrawWalletRepository interface {
	FindByMemberIDAndCoinName(ctx context.Context, memberID int64, coinName string) (*model.MemberWallet, error)
	FindByMemberIDAndCoinNameForUpdate(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string) (*model.MemberWallet, error)
	FreezeBalance(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string, amount float64) error
}

type withdrawAddressRepository interface {
	FindByMemberIDAndCoinID(ctx context.Context, memberID int64, coinID int64) ([]*model.MemberAddress, error)
}

type withdrawRecordRepository interface {
	FindByMemberID(ctx context.Context, memberID int64, page int64, pageSize int64) ([]*model.WithdrawRecord, int64, error)
	Save(ctx context.Context, exec mysqlx.ExtContext, record *model.WithdrawRecord) error
}

type withdrawCache interface {
	GetCtx(ctx context.Context, key string, value any) error
	SetWithExpireCtx(ctx context.Context, key string, value any, ttl time.Duration) error
}

// WithdrawService groups the migrated read-side withdraw workflows.
//
// Why read flows are centralized here:
//   - both gRPC handlers and future async workers need the same mapping rules
//   - keeping market-coin enrichment out of repository code preserves clear
//     layering between persistence and orchestration
//   - Redis-backed verification code handling belongs to the domain service layer
//     rather than transport adapters
//   - write-side withdraw apply logic must also stay here so transaction
//     orchestration, cache validation, and Kafka dispatch remain reusable
type WithdrawService struct {
	memberRepo  withdrawMemberRepository
	walletRepo  withdrawWalletRepository
	addressRepo withdrawAddressRepository
	recordRepo  withdrawRecordRepository
	cache       withdrawCache
	txManager   mysqlx.TxManager
	queue       kafka.Producer
}

func NewWithdrawService(
	memberRepo withdrawMemberRepository,
	walletRepo withdrawWalletRepository,
	addressRepo withdrawAddressRepository,
	recordRepo withdrawRecordRepository,
	cache withdrawCache,
	txManager mysqlx.TxManager,
	queue kafka.Producer,
) *WithdrawService {
	return &WithdrawService{
		memberRepo:  memberRepo,
		walletRepo:  walletRepo,
		addressRepo: addressRepo,
		recordRepo:  recordRepo,
		cache:       cache,
		txManager:   txManager,
		queue:       queue,
	}
}

func (s *WithdrawService) FindAddressByCoinID(ctx context.Context, memberID int64, coinID int64) ([]*withdrawpb.AddressSimple, error) {
	list, err := s.addressRepo.FindByMemberIDAndCoinID(ctx, memberID, coinID)
	if err != nil {
		return nil, err
	}

	resp := make([]*withdrawpb.AddressSimple, 0, len(list))
	for _, item := range list {
		resp = append(resp, item.ToProto())
	}
	return resp, nil
}

func (s *WithdrawService) SendCode(ctx context.Context, phone string) error {
	if phone == "" {
		return errors.New("phone is required")
	}

	code, err := generateNumericCode(6)
	if err != nil {
		return errors.New("generate verification code failed")
	}
	if err := s.cache.SetWithExpireCtx(ctx, withdrawCacheKey+phone, code, 5*time.Minute); err != nil {
		return errors.New("send withdraw verification code failed")
	}
	return nil
}

// Apply executes the write-side withdraw application workflow.
//
// Current migration strategy:
//   - validate Redis verification code and transaction password first
//   - lock the member wallet row inside one SQL transaction
//   - freeze the requested balance and persist one withdraw record
//   - publish the Kafka event before commit so message delivery failure still
//     aborts the balance freeze in this migration phase
//
// This mirrors the legacy atomic intent while the full outbox/consumer
// refactor is still pending in `jobcenter`.
func (s *WithdrawService) Apply(ctx context.Context, req *withdrawpb.WithdrawReq) error {
	if req == nil {
		return errors.New("withdraw request is required")
	}
	if err := validateWithdrawApplyRequest(req); err != nil {
		return err
	}

	member, err := s.memberRepo.FindByID(ctx, req.UserId)
	if err != nil {
		return err
	}
	if member == nil {
		return errors.New("member not found")
	}
	if strings.TrimSpace(member.MobilePhone) == "" {
		return errors.New("member phone is unavailable")
	}

	var cachedCode string
	if err := s.cache.GetCtx(ctx, withdrawCacheKey+member.MobilePhone, &cachedCode); err != nil {
		return errors.New("verification code unavailable")
	}
	if cachedCode != req.Code {
		return errors.New("verification code mismatch")
	}
	if member.JyPassword != req.JyPassword {
		return errors.New("wrong transaction password")
	}

	return s.txManager.WithinTx(ctx, func(exec mysqlx.ExtContext) error {
		wallet, err := s.walletRepo.FindByMemberIDAndCoinNameForUpdate(ctx, exec, req.UserId, req.Unit)
		if err != nil {
			return err
		}
		if wallet == nil {
			return errors.New("wallet not found")
		}
		if wallet.Balance < req.Amount {
			return errors.New("insufficient balance")
		}

		if err := s.walletRepo.FreezeBalance(ctx, exec, req.UserId, req.Unit, req.Amount); err != nil {
			return err
		}

		record := model.NewWithdrawRecordForApply(time.Now(), wallet, req)
		if err := s.recordRepo.Save(ctx, exec, record); err != nil {
			return err
		}

		message, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("marshal withdraw event: %w", err)
		}
		if err := s.queue.PushWithKey(ctx, strconv.FormatInt(req.UserId, 10), string(message)); err != nil {
			return fmt.Errorf("publish withdraw event: %w", err)
		}
		return nil
	})
}

func (s *WithdrawService) FindRecordList(ctx context.Context, memberID int64, page int64, pageSize int64, findCoin func(context.Context, int64) (*marketpb.Coin, error)) ([]*withdrawpb.WithdrawRecord, int64, error) {
	list, total, err := s.recordRepo.FindByMemberID(ctx, memberID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]*withdrawpb.WithdrawRecord, 0, len(list))
	for _, record := range list {
		coin, err := findCoin(ctx, record.CoinId)
		if err != nil {
			return nil, 0, err
		}
		if coin == nil {
			return nil, 0, fmt.Errorf("coin %d not found", record.CoinId)
		}
		resp = append(resp, record.ToProto(coin))
	}

	return resp, total, nil
}

func validateWithdrawApplyRequest(req *withdrawpb.WithdrawReq) error {
	if req.UserId <= 0 {
		return errors.New("user id is required")
	}
	if strings.TrimSpace(req.Unit) == "" {
		return errors.New("coin unit is required")
	}
	if strings.TrimSpace(req.Address) == "" {
		return errors.New("withdraw address is required")
	}
	if req.Amount <= 0 {
		return errors.New("withdraw amount must be greater than zero")
	}
	if req.Fee < 0 {
		return errors.New("withdraw fee cannot be negative")
	}
	if req.Fee > req.Amount {
		return errors.New("withdraw fee exceeds amount")
	}
	if strings.TrimSpace(req.JyPassword) == "" {
		return errors.New("transaction password is required")
	}
	if strings.TrimSpace(req.Code) == "" {
		return errors.New("verification code is required")
	}
	return nil
}
