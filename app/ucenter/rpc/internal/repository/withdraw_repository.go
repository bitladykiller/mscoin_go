package repository

import (
	"context"
	"fmt"

	"mscoin_go/app/ucenter/rpc/internal/model"
	"mscoin_go/pkg/db/mysqlx"

	"github.com/jmoiron/sqlx"
)

// WithdrawRepository 封装所有对提现历史记录的直接 SQL 访问。
type WithdrawRepository struct {
	db *sqlx.DB
}

// NewWithdrawRepository 创建提现仓储实例
func NewWithdrawRepository(db *sqlx.DB) *WithdrawRepository {
	return &WithdrawRepository{db: db}
}

// FindByMemberID 根据会员 ID 分页查询提现记录
func (r *WithdrawRepository) FindByMemberID(ctx context.Context, memberID int64, page int64, pageSize int64) ([]*model.WithdrawRecord, int64, error) {
	// 设置默认分页参数
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 查询总数
	var total int64
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM withdraw_record WHERE member_id=?", memberID); err != nil {
		return nil, 0, fmt.Errorf("count withdraw records: %w", err)
	}

	// 查询列表
	offset := (page - 1) * pageSize
	var list []*model.WithdrawRecord
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM withdraw_record WHERE member_id=? ORDER BY create_time DESC LIMIT ? OFFSET ?", memberID, pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("query withdraw records: %w", err)
	}

	return list, total, nil
}

// Save 持久化一条提现申请记录。
//
// 写入路径在发布 Kafka 消息前存储记录，以便后续异步工作器
// 即使消息投递需要重试，也始终拥有持久的可信源。
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

	// 回填自增 ID
	if insertedID, lastInsertErr := result.LastInsertId(); lastInsertErr == nil {
		record.Id = insertedID
	}
	return nil
}