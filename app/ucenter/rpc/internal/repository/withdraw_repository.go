package repository

import (
	"context"
	"fmt"

	"mscoin_go/app/ucenter/rpc/internal/model"
	"mscoin_go/pkg/db/mysqlx"

	"github.com/jmoiron/sqlx"
)

// WithdrawRepository encapsulates all direct SQL access to withdraw history.
type WithdrawRepository struct {
	db *sqlx.DB
}

func NewWithdrawRepository(db *sqlx.DB) *WithdrawRepository {
	return &WithdrawRepository{db: db}
}

func (r *WithdrawRepository) FindByMemberID(ctx context.Context, memberID int64, page int64, pageSize int64) ([]*model.WithdrawRecord, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var total int64
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM withdraw_record WHERE member_id=?", memberID); err != nil {
		return nil, 0, fmt.Errorf("count withdraw records: %w", err)
	}

	offset := (page - 1) * pageSize
	var list []*model.WithdrawRecord
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM withdraw_record WHERE member_id=? ORDER BY create_time DESC LIMIT ? OFFSET ?", memberID, pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("query withdraw records: %w", err)
	}

	return list, total, nil
}

// Save persists one withdraw application record.
//
// The write path stores the record before publishing Kafka so later async
// workers always have a durable source of truth even if message delivery needs
// retry.
func (r *WithdrawRepository) Save(ctx context.Context, exec mysqlx.ExtContext, record *model.WithdrawRecord) error {
	if exec == nil {
		return fmt.Errorf("sql executor is nil")
	}
	if record == nil {
		return fmt.Errorf("withdraw record is nil")
	}

	const query = `INSERT INTO withdraw_record (
		member_id, coin_id, total_amount, fee, arrived_amount, address, remark,
		transaction_number, can_auto_withdraw, isAuto, status, create_time, deal_time
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := exec.ExecContext(
		ctx,
		query,
		record.MemberId,
		record.CoinId,
		record.TotalAmount,
		record.Fee,
		record.ArrivedAmount,
		record.Address,
		record.Remark,
		record.TransactionNumber,
		record.CanAutoWithdraw,
		record.IsAuto,
		record.Status,
		record.CreateTime,
		record.DealTime,
	)
	if err != nil {
		return fmt.Errorf("insert withdraw record: %w", err)
	}

	if insertedID, lastInsertErr := result.LastInsertId(); lastInsertErr == nil {
		record.Id = insertedID
	}
	return nil
}
