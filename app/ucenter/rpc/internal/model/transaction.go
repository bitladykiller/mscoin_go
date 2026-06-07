// Package model 定义会员交易记录模型。
//
// MemberTransaction 记录会员的所有资产变动，包括：
//   - 充值：外部转账到会员钱包
//   - 提现：会员申请提取资产到外部地址
//   - 转账：会员间内部转账
//   - 兑换：币种兑换
//
// 交易记录是会员资产变动的历史凭证，用于：
//   - 账单查询
//   - 资产审计
//   - 风控分析
package model

import (
	"fmt"
	"strconv"
	"time"

	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// --- 交易类型常量 ---

const (
	transactionRecharge         = iota // 充值：外部转入
	transactionWithdraw                // 提现：转出到外部
	transactionTransferAccounts        // 转账：会员间内部转账
	transactionExchange                // 兑换：币种兑换
)

// transactionTypeNames 交易类型名称映射
// 用于将数字类型转换为字符串表示，便于前端展示
var transactionTypeNames = map[int]string{
	transactionRecharge:         "RECHARGE",
	transactionWithdraw:         "WITHDRAW",
	transactionTransferAccounts: "TRANSFER_ACCOUNTS",
	transactionExchange:         "EXCHANGE",
}

// MemberTransaction 会员交易记录模型
// 记录每一笔资产变动的详细信息
//
// 交易状态由 Type 和 Flag 共同决定：
//   - Type：交易类型（充值、提现等）
//   - Flag：交易标记（成功、失败、处理中等）
type MemberTransaction struct {
	Id          int64   `db:"id" gorm:"column:id"`                   // 交易 ID，自增主键
	Address     string  `db:"address" gorm:"column:address"`         // 交易地址，充值/提现的目标地址
	Amount      float64 `db:"amount" gorm:"column:amount"`           // 交易金额
	CreateTime  int64   `db:"create_time" gorm:"column:create_time"` // 创建时间（毫秒时间戳）
	Fee         float64 `db:"fee" gorm:"column:fee"`                 // 手续费
	Flag        int32   `db:"flag" gorm:"column:flag"`               // 标记：交易状态标记
	MemberId    int64   `db:"member_id" gorm:"column:member_id"`     // 会员 ID，关联会员表
	Symbol      string  `db:"symbol" gorm:"column:symbol"`           // 币种符号，如 BTC、ETH
	Type        int32   `db:"type" gorm:"column:type"`               // 交易类型：0-充值，1-提现，2-转账，3-兑换
	DiscountFee string  `db:"discount_fee" gorm:"column:discount_fee"` // 折扣手续费
	RealFee     string  `db:"real_fee" gorm:"column:real_fee"`       // 实际手续费
}

// ToProto 转换为 protobuf 消息
// 用于 RPC 响应，返回交易记录列表时调用
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
// 将数字类型编码转换为字符串名称
func transactionTypeName(code int) string {
	if name, ok := transactionTypeNames[code]; ok {
		return name
	}
	return ""
}

// ParseTransactionType 解析交易类型
// 将字符串类型转换为数字编码
// 用于处理前端传入的交易类型筛选参数
func ParseTransactionType(raw string) (int32, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse transaction type: %w", err)
	}
	return int32(value), nil
}

// formatMillis 格式化毫秒时间戳为字符串
// 输出格式：2006-01-02 15:04:05
func formatMillis(millis int64) string {
	if millis <= 0 {
		return ""
	}
	return time.UnixMilli(millis).Format("2006-01-02 15:04:05")
}
