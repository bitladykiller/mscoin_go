package model

import (
	marketpb "mscoin_go/app/market/rpc/pb/market"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

type MemberWallet struct {
	Id                int64   `db:"id" gorm:"column:id"`
	Address           string  `db:"address" gorm:"column:address"`
	Balance           float64 `db:"balance" gorm:"column:balance"`
	FrozenBalance     float64 `db:"frozen_balance" gorm:"column:frozen_balance"`
	ReleaseBalance    float64 `db:"release_balance" gorm:"column:release_balance"`
	IsLock            int32   `db:"is_lock" gorm:"column:is_lock"`
	MemberId          int64   `db:"member_id" gorm:"column:member_id"`
	Version           int32   `db:"version" gorm:"column:version"`
	CoinId            int64   `db:"coin_id" gorm:"column:coin_id"`
	ToReleased        float64 `db:"to_released" gorm:"column:to_released"`
	CoinName          string  `db:"coin_name" gorm:"column:coin_name"`
	AddressPrivateKey string  `db:"address_private_key" gorm:"column:address_private_key"`
}

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
