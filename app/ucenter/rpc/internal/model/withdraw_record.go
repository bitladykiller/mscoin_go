package model

import (
	"math"
	"time"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

const (
	// WithdrawStatusProcessing means the withdraw application has been accepted
	// by `ucenter` and is waiting for downstream chain-processing workers.
	WithdrawStatusProcessing int32 = iota
	// WithdrawStatusWaiting keeps the historical status code map intact for later
	// jobcenter phases even though the current migration does not emit it yet.
	WithdrawStatusWaiting
	// WithdrawStatusFail represents a final failed withdraw application.
	WithdrawStatusFail
	// WithdrawStatusSuccess represents a fully completed on-chain withdraw.
	WithdrawStatusSuccess
)

// WithdrawRecord represents one user withdraw application persisted in MySQL.
//
// The refactor keeps the wide legacy table shape because downstream async
// processing, admin workflows, and history pages all read from the same record.
// Narrowing this struct prematurely would make `SELECT *` reads unsafe and
// increase migration risk.
type WithdrawRecord struct {
	Id                int64   `db:"id" gorm:"column:id"`
	MemberId          int64   `db:"member_id" gorm:"column:member_id"`
	CoinId            int64   `db:"coin_id" gorm:"column:coin_id"`
	TotalAmount       float64 `db:"total_amount" gorm:"column:total_amount"`
	Fee               float64 `db:"fee" gorm:"column:fee"`
	ArrivedAmount     float64 `db:"arrived_amount" gorm:"column:arrived_amount"`
	Address           string  `db:"address" gorm:"column:address"`
	Remark            string  `db:"remark" gorm:"column:remark"`
	TransactionNumber string  `db:"transaction_number" gorm:"column:transaction_number"`
	CanAutoWithdraw   int32   `db:"can_auto_withdraw" gorm:"column:can_auto_withdraw"`
	IsAuto            int32   `db:"isAuto" gorm:"column:isAuto"`
	Status            int32   `db:"status" gorm:"column:status"`
	CreateTime        int64   `db:"create_time" gorm:"column:create_time"`
	DealTime          int64   `db:"deal_time" gorm:"column:deal_time"`
}

// NewWithdrawRecordForApply constructs the initial persisted withdraw record
// from one validated user application.
//
// Why the initial record is built in the model package:
//   - every write-side entry point must create the same default status shape
//   - downstream Kafka consumers rely on stable persisted fields
//   - keeping the defaults close to the persistence model prevents business
//     logic from scattering raw status values and time semantics
func NewWithdrawRecordForApply(now time.Time, wallet *MemberWallet, req *withdrawpb.WithdrawReq) *WithdrawRecord {
	return &WithdrawRecord{
		MemberId:          req.UserId,
		CoinId:            wallet.CoinId,
		TotalAmount:       req.Amount,
		Fee:               req.Fee,
		ArrivedAmount:     floorFloat(req.Amount-req.Fee, 10),
		Address:           req.Address,
		Remark:            "",
		TransactionNumber: "",
		CanAutoWithdraw:   0,
		IsAuto:            0,
		Status:            WithdrawStatusProcessing,
		CreateTime:        now.UnixMilli(),
		DealTime:          0,
	}
}

func (r *WithdrawRecord) ToProto(coin *marketpb.Coin) *withdrawpb.WithdrawRecord {
	return &withdrawpb.WithdrawRecord{
		Id:                r.Id,
		MemberId:          r.MemberId,
		Coin:              toWithdrawCoin(coin),
		TotalAmount:       r.TotalAmount,
		Fee:               r.Fee,
		ArrivedAmount:     r.ArrivedAmount,
		Address:           r.Address,
		Remark:            r.Remark,
		TransactionNumber: r.TransactionNumber,
		CanAutoWithdraw:   r.CanAutoWithdraw,
		IsAuto:            r.IsAuto,
		Status:            r.Status,
		CreateTime:        formatMillis(r.CreateTime),
		DealTime:          formatMillis(r.DealTime),
	}
}

func toWithdrawCoin(coin *marketpb.Coin) *withdrawpb.Coin {
	if coin == nil {
		return nil
	}

	return &withdrawpb.Coin{
		Id:                coin.Id,
		Name:              coin.Name,
		CanAutoWithdraw:   coin.CanAutoWithdraw,
		CanRecharge:       coin.CanRecharge,
		CanTransfer:       coin.CanTransfer,
		CanWithdraw:       coin.CanWithdraw,
		CnyRate:           coin.CnyRate,
		EnableRpc:         coin.EnableRpc,
		IsPlatformCoin:    coin.IsPlatformCoin,
		MaxTxFee:          coin.MaxTxFee,
		MaxWithdrawAmount: coin.MaxWithdrawAmount,
		MinTxFee:          coin.MinTxFee,
		MinWithdrawAmount: coin.MinWithdrawAmount,
		NameCn:            coin.NameCn,
		Sort:              coin.Sort,
		Status:            coin.Status,
		Unit:              coin.Unit,
		UsdRate:           coin.UsdRate,
		WithdrawThreshold: coin.WithdrawThreshold,
		HasLegal:          coin.HasLegal,
		ColdWalletAddress: coin.ColdWalletAddress,
		MinerFee:          coin.MinerFee,
		WithdrawScale:     coin.WithdrawScale,
		AccountType:       coin.AccountType,
		DepositAddress:    coin.DepositAddress,
		Infolink:          coin.Infolink,
		Information:       coin.Information,
		MinRechargeAmount: coin.MinRechargeAmount,
	}
}

func floorFloat(value float64, precision int) float64 {
	if precision < 0 {
		precision = 0
	}
	ratio := math.Pow10(precision)
	return math.Floor(value*ratio) / ratio
}
