// Package repository 提供数据访问层实现。
//
// Repository 模式的优势：
//   - 封装数据库操作细节，隔离 SQL 实现
//   - 便于单元测试时 mock 数据访问
//   - 集中管理数据访问逻辑，符合单一职责原则
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"mscoin_go/app/jobcenter/internal/model"

	"github.com/jmoiron/sqlx"
)

// WithdrawRepository 封装提现记录的数据访问操作。
//
// 为什么在 jobcenter 中直接操作此表：
//   - 目前避免重新生成 protobuf，因为仓库使用较旧的 grpc 工具链
//   - 提现表是异步执行状态的权威数据源
//   - 更新操作通过狭窄的 Repository 接口保持跨服务耦合明确和局部化
//   - 后续可引入专用 RPC 契约来解耦
//
// 线程安全：
//   - Repository 本身无状态，所有方法通过传入 context 支持并发调用
//   - 底层 sqlx.DB 连接池是线程安全的
type WithdrawRepository struct {
	// db MySQL 数据库连接
	db *sqlx.DB
}

// NewWithdrawRepository 创建提现记录 Repository 实例。
//
// 参数：
//   - db: MySQL 数据库连接（来自 ServiceContext）
//
// 返回：Repository 实例，可直接使用
func NewWithdrawRepository(db *sqlx.DB) *WithdrawRepository {
	return &WithdrawRepository{db: db}
}

// FindByID 根据主键查询提现记录。
//
// 查询逻辑：
//   - 使用 SELECT * 查询所有字段
//   - LIMIT 1 确保只返回一条记录
//   - 如果记录不存在，返回 nil, nil（非错误）
//
// 参数：
//   - ctx: 上下文，支持超时和取消
//   - id: 提现记录主键
//
// 返回：
//   - *model.WithdrawRecord: 找到的记录，不存在时为 nil
//   - error: 查询错误，不存在记录时返回 nil
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

// MarkSuccess 标记提现记录为成功状态。
//
// 更新逻辑：
//   - 设置 transaction_number（链上交易哈希）
//   - 设置 status 为 WithdrawStatusSuccess
//   - 设置 deal_time（处理完成时间）
//   - 仅当当前状态为 WithdrawStatusProcessing 时才更新
//
// 并发安全：
//   - 使用 WHERE status=? 条件实现乐观锁
//   - 防止重复 Kafka 投递覆盖已完成的记录
//   - 防止人工后台更正被自动处理覆盖
//
// 参数：
//   - ctx: 上下文，支持超时和取消
//   - id: 提现记录主键
//   - txID: 链上交易哈希
//   - dealTime: 处理完成时间戳（毫秒）
//
// 返回：
//   - bool: 是否实际更新了记录（false 表示状态已变更或记录不存在）
//   - error: 更新错误
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
