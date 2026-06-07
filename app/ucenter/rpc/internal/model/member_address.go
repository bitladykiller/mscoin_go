// Package model 定义会员提现地址簿模型。
//
// MemberAddress 存储会员配置的提现目标地址，与 MemberWallet 的区别：
//   - MemberWallet：平台控制的链上钱包，由交易所账户拥有，用于充值
//   - MemberAddress：用户配置的提现目标地址，用于提现到外部钱包
//
// 这种设计将充值和提现的地址概念分离，符合交易所的业务模型。
package model

import withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"

// MemberAddress 存储会员维护的提现地址簿条目。
//
// 为何此模型与钱包表分离：
//   - `member_wallet` 表示平台控制的链上钱包，由交易所账户拥有
//   - `member_address` 表示用户配置的提现目标地址
//   - 将两者分离与旧版 schema 保持一致，避免将截然不同的生命周期规则混入同一聚合
//
// 业务场景：
//   - 会员首次提现时添加地址
//   - 会员可以维护多个币种的多个提现地址
//   - 地址簿便于会员快速选择常用提现地址
type MemberAddress struct {
	Id         int64  `db:"id" gorm:"column:id"`                   // 地址 ID，自增主键
	MemberId   int64  `db:"member_id" gorm:"column:member_id"`     // 会员 ID，关联会员表
	CoinId     int64  `db:"coin_id" gorm:"column:coin_id"`         // 币种 ID，关联币种表
	Address    string `db:"address" gorm:"column:address"`         // 提现地址，区块链钱包地址
	Remark     string `db:"remark" gorm:"column:remark"`           // 备注说明，会员自定义地址名称
	Status     int32  `db:"status" gorm:"column:status"`           // 地址状态：0-正常，1-已删除
	CreateTime int64  `db:"create_time" gorm:"column:create_time"` // 创建时间（毫秒时间戳）
	DeleteTime int64  `db:"delete_time" gorm:"column:delete_time"` // 删除时间（毫秒时间戳），软删除
}

// ToProto 转换为 protobuf 消息
// 用于 RPC 响应，返回地址簿列表时调用
func (m *MemberAddress) ToProto() *withdrawpb.AddressSimple {
	return &withdrawpb.AddressSimple{
		Remark:  m.Remark,
		Address: m.Address,
	}
}
