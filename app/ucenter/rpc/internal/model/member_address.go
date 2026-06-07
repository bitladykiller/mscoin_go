package model

import withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"

// MemberAddress stores a member-maintained withdraw address book entry.
//
// Why this model stays separate from the wallet table:
//   - `member_wallet` represents the platform-controlled on-chain wallet owned by
//     the exchange account
//   - `member_address` represents user-configured destination addresses for
//     withdraw operations
//   - keeping them separate mirrors the legacy schema and avoids mixing very
//     different lifecycle rules into one aggregate
type MemberAddress struct {
	Id         int64  `db:"id" gorm:"column:id"`
	MemberId   int64  `db:"member_id" gorm:"column:member_id"`
	CoinId     int64  `db:"coin_id" gorm:"column:coin_id"`
	Address    string `db:"address" gorm:"column:address"`
	Remark     string `db:"remark" gorm:"column:remark"`
	Status     int32  `db:"status" gorm:"column:status"`
	CreateTime int64  `db:"create_time" gorm:"column:create_time"`
	DeleteTime int64  `db:"delete_time" gorm:"column:delete_time"`
}

func (m *MemberAddress) ToProto() *withdrawpb.AddressSimple {
	return &withdrawpb.AddressSimple{
		Remark:  m.Remark,
		Address: m.Address,
	}
}
