package model

import "testing"

func TestMemberAddressToProto(t *testing.T) {
	address := &MemberAddress{
		Remark:  "common wallet",
		Address: "1BitcoinAddr",
	}

	payload := address.ToProto()
	if payload.Remark != address.Remark {
		t.Fatalf("ToProto().Remark = %q, want %q", payload.Remark, address.Remark)
	}
	if payload.Address != address.Address {
		t.Fatalf("ToProto().Address = %q, want %q", payload.Address, address.Address)
	}
}
