package model

import (
	"testing"

	marketpb "mscoin_go/app/market/rpc/pb/market"
)

func TestMemberWalletToProto(t *testing.T) {
	wallet := &MemberWallet{
		Id:             1,
		Address:        "wallet-address",
		Balance:        10.5,
		FrozenBalance:  1.25,
		ReleaseBalance: 0.75,
		IsLock:         0,
		MemberId:       100,
		Version:        2,
		ToReleased:     8.5,
	}
	coin := &marketpb.Coin{
		Id:                9,
		Name:              "Bitcoin",
		Unit:              "BTC",
		CnyRate:           123.45,
		UsdRate:           17.89,
		CanRecharge:       1,
		CanWithdraw:       1,
		CanTransfer:       1,
		CanAutoWithdraw:   0,
		EnableRpc:         1,
		IsPlatformCoin:    0,
		MaxTxFee:          1.1,
		MaxWithdrawAmount: 1000,
		MinTxFee:          0.01,
		MinWithdrawAmount: 0.1,
		NameCn:            "比特币",
		Sort:              1,
		Status:            1,
		WithdrawThreshold: 0.5,
		HasLegal:          1,
		ColdWalletAddress: "cold-wallet",
		MinerFee:          0.001,
		WithdrawScale:     8,
		AccountType:       1,
		DepositAddress:    "deposit-address",
		Infolink:          "https://example.com/btc",
		Information:       "coin-info",
		MinRechargeAmount: 0.2,
	}

	result := wallet.ToProto(coin)
	if result.Id != wallet.Id {
		t.Fatalf("ToProto().Id = %d, want %d", result.Id, wallet.Id)
	}
	if result.MemberId != wallet.MemberId {
		t.Fatalf("ToProto().MemberId = %d, want %d", result.MemberId, wallet.MemberId)
	}
	if result.Coin == nil {
		t.Fatal("ToProto().Coin = nil, want non-nil")
	}
	if result.Coin.Unit != coin.Unit {
		t.Fatalf("ToProto().Coin.Unit = %q, want %q", result.Coin.Unit, coin.Unit)
	}
	if result.Coin.MinRechargeAmount != coin.MinRechargeAmount {
		t.Fatalf("ToProto().Coin.MinRechargeAmount = %v, want %v", result.Coin.MinRechargeAmount, coin.MinRechargeAmount)
	}
}
