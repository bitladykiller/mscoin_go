// Package model 定义会员钱包模型。
//
// MemberWallet 是会员资产管理的核心模型，存储会员的币种钱包信息。
// 每个会员可以拥有多个币种的钱包，每个币种对应一条 MemberWallet 记录。
//
// 钱包与充提关系：
//   - 充值：会员向钱包地址转账，链上确认后增加余额
//   - 提现：会员申请提现，冻结余额后由后台处理链上转账
//   - 转账：会员间内部转账，直接修改余额
package model

import (
	marketpb "mscoin_go/app/market/rpc/pb/market"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// MemberWallet 会员钱包模型
// 存储会员的币种钱包信息，包括余额、冻结余额等
//
// 余额字段说明：
//   - Balance：可用余额，会员可用于提现或转账
//   - FrozenBalance：冻结余额，提现申请时冻结，提现完成或失败后解冻
//   - ReleaseBalance：释放余额，用于锁仓释放场景
//   - ToReleased：待释放金额，锁仓金额逐步释放
//
// 安全设计：
//   - Version：乐观锁版本号，防止并发更新
//   - IsLock：钱包锁定状态，用于风控
//   - AddressPrivateKey：地址私钥（已废弃，私钥由 Bitcoin Core 管理）
type MemberWallet struct {
	Id                int64   `db:"id" gorm:"column:id"`                                 // 钱包 ID，自增主键
	Address           string  `db:"address" gorm:"column:address"`                       // 钱包地址，用于充值
	Balance           float64 `db:"balance" gorm:"column:balance"`                       // 可用余额，会员可自由支配
	FrozenBalance     float64 `db:"frozen_balance" gorm:"column:frozen_balance"`         // 冻结余额，提现申请时冻结
	ReleaseBalance    float64 `db:"release_balance" gorm:"column:release_balance"`       // 释放余额，锁仓释放场景
	IsLock            int32   `db:"is_lock" gorm:"column:is_lock"`                       // 是否锁定：0-正常，1-锁定
	MemberId          int64   `db:"member_id" gorm:"column:member_id"`                   // 会员 ID，关联会员表
	Version           int32   `db:"version" gorm:"column:version"`                       // 版本号，乐观锁
	CoinId            int64   `db:"coin_id" gorm:"column:coin_id"`                       // 币种 ID，关联币种表
	ToReleased        float64 `db:"to_released" gorm:"column:to_released"`               // 待释放金额
	CoinName          string  `db:"coin_name" gorm:"column:coin_name"`                   // 币种名称，如 BTC、ETH
	AddressPrivateKey string  `db:"address_private_key" gorm:"column:address_private_key"` // 地址私钥（已废弃）
}

// ToProto 转换为 protobuf 消息
// 用于 RPC 响应，返回钱包信息时调用
// 参数 coin 为币种市场信息，用于丰富钱包展示数据
func (w *MemberWallet) ToProto(coin *marketpb.Coin) *assetpb.MemberWallet {
	return &assetpb.MemberWallet{
		Id:             w.Id,
		Address:        w.Address,
		Balance:        w.Balance,
		FrozenBalance:  w.FrozenBalance,
		ReleaseBalance: w.ReleaseBalance,
		IsLock:         w.IsLock,
		MemberId:       w.MemberId,
		Version:        w.Version,
		Coin: &assetpb.Coin{
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
		},
		ToReleased: w.ToReleased,
	}
}