// Package model 定义 jobcenter 服务使用的数据模型。
//
// 该包包含：
//   - 数据库映射结构体（ORM）
//   - Kafka 消息事件结构
//   - 业务状态常量
package model

// 提现状态常量定义。
//
// 这些状态码保持与原有数据库语义一致，确保：
//   - 异步工作者能与现有管理页面协同
//   - 历史查询数据一致性
//   - 状态流转可追溯
const (
	// WithdrawStatusProcessing 处理中状态（0）。
	// 提现申请已提交，等待 jobcenter 处理。
	// 在此状态下，jobcenter 会尝试执行链上转账。
	WithdrawStatusProcessing int32 = iota

	// WithdrawStatusWaiting 等待状态（1）。
	// 提现申请等待人工审核或其他条件满足。
	WithdrawStatusWaiting

	// WithdrawStatusFail 失败状态（2）。
	// 提现执行失败，可能是链上交易失败或其他错误。
	WithdrawStatusFail

	// WithdrawStatusSuccess 成功状态（3）。
	// 提现已成功执行，链上交易已确认。
	WithdrawStatusSuccess
)

// WithdrawRecord 映射 withdraw_record 表的行数据。
//
// 该结构体用于：
//   - 从数据库读取提现记录
//   - 查询提现状态和详情
//   - 作为业务处理的数据源
//
// 字段说明：
//   - Id: 主键，唯一标识一条提现记录
//   - MemberId: 用户ID，关联用户表
//   - CoinId: 币种ID，用于查询币种信息
//   - TotalAmount: 提现总额（包含手续费）
//   - Fee: 手续费金额
//   - ArrivedAmount: 到账金额（TotalAmount - Fee）
//   - Address: 目标提现地址
//   - Remark: 提现备注
//   - TransactionNumber: 链上交易哈希（执行成功后填充）
//   - CanAutoWithdraw: 是否允许自动提现标志
//   - IsAuto: 是否为自动提现标志
//   - Status: 当前状态（见状态常量）
//   - CreateTime: 创建时间戳（毫秒）
//   - DealTime: 处理完成时间戳（毫秒）
type WithdrawRecord struct {
	Id                int64   `db:"id"`
	MemberId          int64   `db:"member_id"`
	CoinId            int64   `db:"coin_id"`
	TotalAmount       float64 `db:"total_amount"`
	Fee               float64 `db:"fee"`
	ArrivedAmount     float64 `db:"arrived_amount"`
	Address           string  `db:"address"`
	Remark            string  `db:"remark"`
	TransactionNumber string  `db:"transaction_number"`
	CanAutoWithdraw   int32   `db:"can_auto_withdraw"`
	IsAuto            int32   `db:"isAuto"`
	Status            int32   `db:"status"`
	CreateTime        int64   `db:"create_time"`
	DealTime          int64   `db:"deal_time"`
}

// WithdrawRecordEvent 是 ucenter-rpc 发出的 Kafka 消息载荷。
//
// 该结构体用于：
//   - 反序列化 Kafka 消息
//   - 传递提现事件给领域服务处理
//
// JSON 标签说明：
//   - 标签名与 Go 字段名保持一致（首字母大写）
//   - 原因：当前生产者直接序列化持久化结构体，未使用小写 json 标签
//   - 保持一致性避免反序列化失败
//
// 注意：未来如需修改消息格式，需要同步修改生产端序列化逻辑。
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
