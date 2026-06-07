package model

import "testing"

func TestMemberTransactionToProto(t *testing.T) {
	transaction := &MemberTransaction{
		Id:          1,
		Address:     "wallet-address",
		Amount:      12.5,
		CreateTime:  1710000000000,
		Fee:         0.2,
		Flag:        1,
		MemberId:    100,
		Symbol:      "BTC",
		Type:        transactionWithdraw,
		DiscountFee: "0.01",
		RealFee:     "0.19",
	}

	payload := transaction.ToProto()
	if payload.Id != transaction.Id {
		t.Fatalf("ToProto().Id = %d, want %d", payload.Id, transaction.Id)
	}
	if payload.Type != "WITHDRAW" {
		t.Fatalf("ToProto().Type = %q, want %q", payload.Type, "WITHDRAW")
	}
	if payload.CreateTime == "" {
		t.Fatal("ToProto().CreateTime is empty")
	}
}

func TestParseTransactionType(t *testing.T) {
	value, err := ParseTransactionType("2")
	if err != nil {
		t.Fatalf("ParseTransactionType() error = %v", err)
	}
	if value != 2 {
		t.Fatalf("ParseTransactionType() = %d, want 2", value)
	}
	if _, err = ParseTransactionType("invalid"); err == nil {
		t.Fatal("ParseTransactionType() error = nil, want parse error")
	}
}
