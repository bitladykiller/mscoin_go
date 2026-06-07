package model

import (
	"fmt"
	"strconv"
	"time"

	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

const (
	transactionRecharge = iota
	transactionWithdraw
	transactionTransferAccounts
	transactionExchange
)

var transactionTypeNames = map[int]string{
	transactionRecharge:         "RECHARGE",
	transactionWithdraw:         "WITHDRAW",
	transactionTransferAccounts: "TRANSFER_ACCOUNTS",
	transactionExchange:         "EXCHANGE",
}

type MemberTransaction struct {
	Id          int64   `db:"id" gorm:"column:id"`
	Address     string  `db:"address" gorm:"column:address"`
	Amount      float64 `db:"amount" gorm:"column:amount"`
	CreateTime  int64   `db:"create_time" gorm:"column:create_time"`
	Fee         float64 `db:"fee" gorm:"column:fee"`
	Flag        int32   `db:"flag" gorm:"column:flag"`
	MemberId    int64   `db:"member_id" gorm:"column:member_id"`
	Symbol      string  `db:"symbol" gorm:"column:symbol"`
	Type        int32   `db:"type" gorm:"column:type"`
	DiscountFee string  `db:"discount_fee" gorm:"column:discount_fee"`
	RealFee     string  `db:"real_fee" gorm:"column:real_fee"`
}

func (m *MemberTransaction) ToProto() *assetpb.MemberTransaction {
	return &assetpb.MemberTransaction{
		Id:          m.Id,
		Address:     m.Address,
		Amount:      m.Amount,
		CreateTime:  formatMillis(m.CreateTime),
		Fee:         m.Fee,
		Flag:        m.Flag,
		MemberId:    m.MemberId,
		Symbol:      m.Symbol,
		Type:        transactionTypeName(int(m.Type)),
		DiscountFee: m.DiscountFee,
		RealFee:     m.RealFee,
	}
}

func transactionTypeName(code int) string {
	if name, ok := transactionTypeNames[code]; ok {
		return name
	}
	return ""
}

func ParseTransactionType(raw string) (int32, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse transaction type: %w", err)
	}
	return int32(value), nil
}

func formatMillis(millis int64) string {
	if millis <= 0 {
		return ""
	}
	return time.UnixMilli(millis).Format("2006-01-02 15:04:05")
}
