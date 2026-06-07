package model

import (
	"math"
	"time"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

// 提现状态常量
const (
	// WithdrawStatusProcessing 表示提现申请已被 `ucenter` 接受，
	// 正在等待下游链上处理工作器处理。
	WithdrawStatusProcessing int32 = iota
	// WithdrawStatusWaiting 保留历史状态码映射，供后续 `jobcenter` 阶段使用，
	// 尽管当前迁移尚未发出此状态。
	WithdrawStatusWaiting
	// WithdrawStatusFail 表示最终失败的提现申请。
	WithdrawStatusFail
	// WithdrawStatusSuccess 表示已完全完成的链上提现。
	WithdrawStatusSuccess
)

// WithdrawRecord 表示一条持久化到 MySQL 的用户提现申请记录。
//
// 重构保留了较宽的旧版表结构，因为下游异步处理、管理流程和历史页面
// 都从同一条记录读取。过早缩减此结构会使 `SELECT *` 读取不安全，
// 并增加迁移风险。
type WithdrawRecord struct {
	Id                int64   `db:"id" gorm:"column:id"`                         // 记录 ID
	MemberId          int64   `db:"member_id" gorm:"column:member_id"`           // 会员 ID
	CoinId            int64   `db:"coin_id" gorm:"column:coin_id"`               // 币种 ID
	TotalAmount       float64 `db:"total_amount" gorm:"column:total_amount"`     // 提现总额
	Fee               float64 `db:"fee" gorm:"column:fee"`                       // 手续费
	ArrivedAmount     float64 `db:"arrived_amount" gorm:"column:arrived_amount"` // 到账金额
	Address           string  `db:"address" gorm:"column:address"`               // 提现地址
	Remark            string  `db:"remark" gorm:"column:remark"`                 // 备注
	TransactionNumber string  `db:"transaction_number" gorm:"column:transaction_number"` // 交易号
	CanAutoWithdraw   int32   `db:"can_auto_withdraw" gorm:"column:can_auto_withdraw"`   // 是否可自动提现
	IsAuto            int32   `db:"isAuto" gorm:"column:isAuto"`                 // 是否自动处理
	Status            int32   `db:"status" gorm:"column:status"`                 // 提现状态
	CreateTime        int64   `db:"create_time" gorm:"column:create_time"`       // 创建时间
	DealTime          int64   `db:"deal_time" gorm:"column:deal_time"`           // 处理时间
}

// NewWithdrawRecordForApply 从一条已验证的用户申请构建初始持久化提现记录。
//
// 为何初始记录在 model 包中构建：
//   - 所有写侧入口点必须创建相同的默认状态结构
//   - 下游 Kafka 消费者依赖稳定的持久化字段
//   - 将默认值保持在持久化模型附近，可防止业务逻辑散落原始状态值和时间语义
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

// ToProto 转换为 protobuf 消息
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

// toWithdrawCoin 转换为提现币种 protobuf 消息
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

// floorFloat 对浮点数进行向下取整
func floorFloat(value float64, precision int) float64 {
	if precision < 0 {
		precision = 0
	}
	ratio := math.Pow10(precision)
	return math.Floor(value*ratio) / ratio
}
