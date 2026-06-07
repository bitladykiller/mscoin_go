package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"mscoin_go/app/jobcenter/internal/model"

	"github.com/jmoiron/sqlx"
)

// WithdrawRepository 封装 jobcenter 提现状态终结所需的直接 SQL 操作。
//
// 为什么在当前迁移阶段 jobcenter 直接操作此表：
//   - 目前刻意避免重新生成 protobuf，因为仓库使用较旧的 grpc 工具链
//   - 提现表仍是异步执行状态的权威数据源
//   - 更新一个窄仓库使跨服务耦合保持明确和局部化，直至后续引入专用 RPC 契约
type WithdrawRepository struct {
	db *sqlx.DB
}

func NewWithdrawRepository(db *sqlx.DB) *WithdrawRepository {
	return &WithdrawRepository{db: db}
}

func (r *WithdrawRepository) FindByID(ctx context.Context, id int64) (*model.WithdrawRecord, error) {
	var record model.WithdrawRecord
	err := r.db.GetContext(ctx, &record, "SELECT * FROM withdraw_record WHERE id=? LIMIT 1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query withdraw record: %w", err)
	}
	return &record, nil
}

// MarkSuccess 持久化最终的 txid 和成功状态。
//
// 更新操作受当前状态保护，因此重复的 Kafka 投递或人工后台更正
// 不会覆盖已脱离处理状态的记录。
func (r *WithdrawRepository) MarkSuccess(ctx context.Context, id int64, txID string, dealTime int64) (bool, error) {
	result, err := r.db.ExecContext(
		ctx,
		"UPDATE withdraw_record SET transaction_number=?, status=?, deal_time=? WHERE id=? AND status=?",
		txID,
		model.WithdrawStatusSuccess,
		dealTime,
		id,
		model.WithdrawStatusProcessing,
	)
	if err != nil {
		return false, fmt.Errorf("update withdraw record success: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read withdraw record update rows: %w", err)
	}
	return rowsAffected > 0, nil
}
