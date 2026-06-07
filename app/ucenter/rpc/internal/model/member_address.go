package model

import withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"

// MemberAddress 存储会员维护的提现地址簿条目。
//
// 为何此模型与钱包表分离：
//   - `member_wallet` 表示平台控制的链上钱包，由交易所账户拥有
//   - `member_address` 表示用户配置的提现目标地址
//   - 将两者分离与旧版 schema 保持一致，避免将截然不同的生命周期规则混入同一聚合
type MemberAddress struct {
	Id         int64  `db:"id" gorm:"column:id"`                   // 地址 ID
	MemberId   int64  `db:"member_id" gorm:"column:member_id"`       // 会员 ID
	CoinId     int64  `db:"coin_id" gorm:"column:coin_id"`           // 币种 ID
	Address    string `db:"address" gorm:"column:address"`           // 提现地址
	Remark     string `db:"remark" gorm:"column:remark"`             // 备注说明
	Status     int32  `db:"status" gorm:"column:status"`             // 地址状态
	CreateTime int64  `db:"create_time" gorm:"column:create_time"` // 创建时间
	DeleteTime int64  `db:"delete_time" gorm:"column:delete_time"`   // 删除时间
}

// ToProto 转换为 protobuf 消息
func (m *MemberAddress) ToProto() *withdrawpb.AddressSimple {
	return &withdrawpb.AddressSimple{
		Remark:  m.Remark,
		Address: m.Address,
	}
}
