package model

import (
	marketpb "mscoin_go/app/market/rpc/pb/market"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// MemberWallet 会员钱包模型
// 存储会员的币种钱包信息，包括余额、冻结余额等
type MemberWallet struct {
	Id                int64   `db:"id" gorm:"column:id"`                                 // 钱包 ID
	Address           string  `db:"address" gorm:"column:address"`                       // 钱包地址
	Balance           float64 `db:"balance" gorm:"column:balance"`                       // 可用余额
	FrozenBalance     float64 `db:"frozen_balance" gorm:"column:frozen_balance"`         // 冻结余额
	ReleaseBalance    float64 `db:"release_balance" gorm:"column:release_balance"`       // 释放余额
	IsLock            int32   `db:"is_lock" gorm:"column:is_lock"`                       // 是否锁定
	MemberId          int64   `db:"member_id" gorm:"column:member_id"`                   // 会员 ID
	Version           int32   `db:"version" gorm:"column:version"`                       // 版本号
	CoinId            int64   `db:"coin_id" gorm:"column:coin_id"`                       // 币种 ID
	ToReleased        float64 `db:"to_released" gorm:"column:to_released"`               // 待释放金额
	CoinName          string  `db:"coin_name" gorm:"column:coin_name"`                   // 币种名称
	AddressPrivateKey string  `db:"address_private_key" gorm:"column:address_private_key"` // 地址私钥
}

// ToProto 转换为 protobuf 消息
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