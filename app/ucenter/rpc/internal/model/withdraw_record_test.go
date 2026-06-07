package model

import (
	"testing"
	"time"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

func TestWithdrawRecordToProto(t *testing.T) {
	record := &WithdrawRecord{
		Id:                1,
		MemberId:          2,
		TotalAmount:       10.5,
		Fee:               0.1,
		ArrivedAmount:     10.4,
		Address:           "wallet-address",
		Remark:            "test",
		TransactionNumber: "tx-hash",
		CanAutoWithdraw:   1,
		IsAuto:            0,
		Status:            3,
		CreateTime:        1710000000000,
		DealTime:          1710003600000,
	}
	coin := &marketpb.Coin{
		Id:                9,
		Name:              "Bitcoin",
		NameCn:            "比特币",
		Unit:              "BTC",
		WithdrawScale:     8,
		AccountType:       1,
		MinTxFee:          0.01,
		MaxTxFee:          1.2,
		MinWithdrawAmount: 0.1,
		MaxWithdrawAmount: 100,
	}

	payload := record.ToProto(coin)
	if payload.Id != record.Id {
		t.Fatalf("ToProto().Id = %d, want %d", payload.Id, record.Id)
	}
	if payload.Coin == nil || payload.Coin.Unit != "BTC" {
		t.Fatalf("ToProto().Coin = %+v, want BTC coin", payload.Coin)
	}
	if payload.CreateTime == "" || payload.DealTime == "" {
		t.Fatalf("ToProto() time fields should be formatted, got create=%q deal=%q", payload.CreateTime, payload.DealTime)
	}
}

func TestNewWithdrawRecordForApply(t *testing.T) {
	now := time.Unix(1700000000, 0)
	record := NewWithdrawRecordForApply(now, &MemberWallet{CoinId: 9}, &withdrawpb.WithdrawReq{
		UserId:  7,
		Address: "tb1-destination",
		Amount:  10.2,
		Fee:     0.2,
	})

	if record.MemberId != 7 {
		t.Fatalf("NewWithdrawRecordForApply().MemberId = %d, want 7", record.MemberId)
	}
	if record.CoinId != 9 {
		t.Fatalf("NewWithdrawRecordForApply().CoinId = %d, want 9", record.CoinId)
	}
	if record.ArrivedAmount != 10 {
		t.Fatalf("NewWithdrawRecordForApply().ArrivedAmount = %v, want 10", record.ArrivedAmount)
	}
	if record.Status != WithdrawStatusProcessing {
		t.Fatalf("NewWithdrawRecordForApply().Status = %d, want %d", record.Status, WithdrawStatusProcessing)
	}
	if record.CreateTime != now.UnixMilli() {
		t.Fatalf("NewWithdrawRecordForApply().CreateTime = %d, want %d", record.CreateTime, now.UnixMilli())
	}
}
