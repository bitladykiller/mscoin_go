package repository

import (
	"context"
	"fmt"

	"mscoin_go/app/ucenter/rpc/internal/model"

	"github.com/jmoiron/sqlx"
)

// MemberAddressRepository 拥有提现地址簿表的 SQL 访问能力。
type MemberAddressRepository struct {
	db *sqlx.DB
}

// NewMemberAddressRepository 创建会员地址仓储实例
func NewMemberAddressRepository(db *sqlx.DB) *MemberAddressRepository {
	return &MemberAddressRepository{db: db}
}

// FindByMemberIDAndCoinID 根据会员 ID 和币种 ID 查询提现地址列表
func (r *MemberAddressRepository) FindByMemberIDAndCoinID(ctx context.Context, memberID int64, coinID int64) ([]*model.MemberAddress, error) {
	var list []*model.MemberAddress
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM member_address WHERE member_id=? AND coin_id=?", memberID, coinID); err != nil {
		return nil, fmt.Errorf("query member addresses: %w", err)
	}
	return list, nil
}
