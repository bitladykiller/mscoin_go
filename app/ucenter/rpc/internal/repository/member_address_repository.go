package repository

import (
	"context"
	"fmt"

	"mscoin_go/app/ucenter/rpc/internal/model"

	"github.com/jmoiron/sqlx"
)

// MemberAddressRepository owns SQL access for the withdraw address book table.
type MemberAddressRepository struct {
	db *sqlx.DB
}

func NewMemberAddressRepository(db *sqlx.DB) *MemberAddressRepository {
	return &MemberAddressRepository{db: db}
}

func (r *MemberAddressRepository) FindByMemberIDAndCoinID(ctx context.Context, memberID int64, coinID int64) ([]*model.MemberAddress, error) {
	var list []*model.MemberAddress
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM member_address WHERE member_id=? AND coin_id=?", memberID, coinID); err != nil {
		return nil, fmt.Errorf("query member addresses: %w", err)
	}
	return list, nil
}
