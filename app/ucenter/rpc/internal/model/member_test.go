package model

import (
	"testing"
	"time"
)

func TestMemberLevelText(t *testing.T) {
	testCases := []struct {
		name        string
		memberLevel int64
		want        string
	}{
		{name: "general", memberLevel: generalLevel, want: "普通会员"},
		{name: "real_name", memberLevel: realNameLevel, want: "实名"},
		{name: "merchant", memberLevel: identificationLevel, want: "认证商家"},
		{name: "unknown", memberLevel: 99, want: ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			member := &Member{MemberLevel: testCase.memberLevel}
			if got := member.MemberLevelText(); got != testCase.want {
				t.Fatalf("MemberLevelText() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestMemberRate(t *testing.T) {
	testCases := []struct {
		name         string
		superPartner string
		want         int32
	}{
		{name: "normal", superPartner: normalPartner, want: 0},
		{name: "super", superPartner: superPartner, want: 1},
		{name: "parent_super", superPartner: pSuperPartner, want: 2},
		{name: "unknown", superPartner: "9", want: 0},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			member := &Member{SuperPartner: testCase.superPartner}
			if got := member.MemberRate(); got != testCase.want {
				t.Fatalf("MemberRate() = %d, want %d", got, testCase.want)
			}
		})
	}
}

func TestFillSuperPartner(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		wantPartner string
		wantStatus  int64
	}{
		{name: "empty_partner_defaults_to_normal", input: "", wantPartner: normalPartner, wantStatus: normalMemberStatus},
		{name: "normal_partner_stays_normal", input: normalPartner, wantPartner: normalPartner, wantStatus: normalMemberStatus},
		{name: "super_partner_marks_illegal", input: superPartner, wantPartner: superPartner, wantStatus: illegalMemberStatus},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			member := &Member{}
			member.FillSuperPartner(testCase.input)

			if member.SuperPartner != testCase.wantPartner {
				t.Fatalf("FillSuperPartner() partner = %q, want %q", member.SuperPartner, testCase.wantPartner)
			}
			if member.Status != testCase.wantStatus {
				t.Fatalf("FillSuperPartner() status = %d, want %d", member.Status, testCase.wantStatus)
			}
		})
	}
}

func TestNewMemberForRegister(t *testing.T) {
	member := NewMemberForRegister(
		time.Unix(1700000000, 0),
		"13800000000",
		"alice",
		"CN",
		"encoded-password",
		"salt-value",
		"",
		"PROMO",
	)

	if member.MobilePhone != "13800000000" {
		t.Fatalf("NewMemberForRegister().MobilePhone = %q, want %q", member.MobilePhone, "13800000000")
	}
	if member.Username != "alice" {
		t.Fatalf("NewMemberForRegister().Username = %q, want %q", member.Username, "alice")
	}
	if member.MemberLevel != generalLevel {
		t.Fatalf("NewMemberForRegister().MemberLevel = %d, want %d", member.MemberLevel, generalLevel)
	}
	if member.SuperPartner != normalPartner {
		t.Fatalf("NewMemberForRegister().SuperPartner = %q, want %q", member.SuperPartner, normalPartner)
	}
	if member.Avatar == "" {
		t.Fatal("NewMemberForRegister().Avatar is empty")
	}
}
