package model

// Withdraw 提现状态码，保持与原有数据库语义一致，
// 以便异步工作者能与现有管理页面和历史查询协同工作。
const (
	WithdrawStatusProcessing int32 = iota
	WithdrawStatusWaiting
	WithdrawStatusFail
	WithdrawStatusSuccess
)

// WithdrawRecord 映射 `withdraw_record` 表的行数据，
// jobcenter 在链上执行后更新此记录。
type WithdrawRecord struct {
	Id                int64   `db:"id" gorm:"column:id"`
	MemberId          int64   `db:"member_id" gorm:"column:member_id"`
	CoinId            int64   `db:"coin_id" gorm:"column:coin_id"`
	TotalAmount       float64 `db:"total_amount" gorm:"column:total_amount"`
	Fee               float64 `db:"fee" gorm:"column:fee"`
	ArrivedAmount     float64 `db:"arrived_amount" gorm:"column:arrived_amount"`
	Address           string  `db:"address" gorm:"column:address"`
	Remark            string  `db:"remark" gorm:"column:remark"`
	TransactionNumber string  `db:"transaction_number" gorm:"column:transaction_number"`
	CanAutoWithdraw   int32   `db:"can_auto_withdraw" gorm:"column:can_auto_withdraw"`
	IsAuto            int32   `db:"isAuto" gorm:"column:isAuto"`
	Status            int32   `db:"status" gorm:"column:status"`
	CreateTime        int64   `db:"create_time" gorm:"column:create_time"`
	DealTime          int64   `db:"deal_time" gorm:"column:deal_time"`
}

// WithdrawRecordEvent 是 `ucenter-rpc` 发出的当前 Kafka 消息载荷。
//
// JSON 标签刻意与 Go 字段名保持一致，因为当前生产者直接序列化
// 持久化结构体，未使用小写的 `json` 标签。
type WithdrawRecordEvent struct {
	Id                int64   `json:"Id"`
	MemberId          int64   `json:"MemberId"`
	CoinId            int64   `json:"CoinId"`
	TotalAmount       float64 `json:"TotalAmount"`
	Fee               float64 `json:"Fee"`
	ArrivedAmount     float64 `json:"ArrivedAmount"`
	Address           string  `json:"Address"`
	Remark            string  `json:"Remark"`
	TransactionNumber string  `json:"TransactionNumber"`
	CanAutoWithdraw   int32   `json:"CanAutoWithdraw"`
	IsAuto            int32   `json:"IsAuto"`
	Status            int32   `json:"Status"`
	CreateTime        int64   `json:"CreateTime"`
	DealTime          int64   `json:"DealTime"`
}
