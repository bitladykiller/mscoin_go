package model

import (
	"fmt"
	"strconv"
	"time"

	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// 交易类型常量
const (
	transactionRecharge         = iota // 充值
	transactionWithdraw                // 提现
	transactionTransferAccounts        // 转账
	transactionExchange                // 兑换
)

// transactionTypeNames 交易类型名称映射
var transactionTypeNames = map[int]string{
	transactionRecharge:         "RECHARGE",
	transactionWithdraw:         "WITHDRAW",
	transactionTransferAccounts: "TRANSFER_ACCOUNTS",
	transactionExchange:         "EXCHANGE",
}

// MemberTransaction 会员交易记录模型
type MemberTransaction struct {
	Id          int64   `db:"id" gorm:"column:id"`                   // 交易 ID
	Address     string  `db:"address" gorm:"column:address"`         // 交易地址
	Amount      float64 `db:"amount" gorm:"column:amount"`           // 交易金额
	CreateTime  int64   `db:"create_time" gorm:"column:create_time"` // 创建时间
	Fee         float64 `db:"fee" gorm:"column:fee"`                 // 手续费
	Flag        int32   `db:"flag" gorm:"column:flag"`               // 标记
	MemberId    int64   `db:"member_id" gorm:"column:member_id"`     // 会员 ID
	Symbol      string  `db:"symbol" gorm:"column:symbol"`           // 币种符号
	Type        int32   `db:"type" gorm:"column:type"`               // 交易类型
	DiscountFee string  `db:"discount_fee" gorm:"column:discount_fee"` // 折扣手续费
	RealFee     string  `db:"real_fee" gorm:"column:real_fee"`       // 实际手续费
}

// ToProto 转换为 protobuf 消息
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

// transactionTypeName 获取交易类型名称
func transactionTypeName(code int) string {
	if name, ok := transactionTypeNames[code]; ok {
		return name
	}
	return ""
}

// ParseTransactionType 解析交易类型
func ParseTransactionType(raw string) (int32, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse transaction type: %w", err)
	}
	return int32(value), nil
}

// formatMillis 格式化毫秒时间戳为字符串
func formatMillis(millis int64) string {
	if millis <= 0 {
		return ""
	}
	return time.UnixMilli(millis).Format("2006-01-02 15:04:05")
}
