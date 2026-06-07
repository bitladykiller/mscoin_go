// Package repository 定义提现记录仓储层。
//
// WithdrawRepository 封装提现记录表（withdraw_record）的数据库操作。
// 提现记录是会员提现申请的持久化凭证，记录提现的全生命周期。
//
// 提现流程与仓储交互：
//  1. Apply：在事务中保存提现记录（状态为 Processing）
//  2. jobcenter：更新状态为 Success 或 Fail
//  3. 前端查询：展示提现历史
//
// 事务安全设计：
//   - Save 方法支持事务执行器
//   - 在提现申请事务中与 FreezeBalance 原子执行
//   - 确保记录创建和余额冻结同时成功或失败
package repository

import (
	"context"
	"fmt"

	"mscoin_go/app/ucenter/rpc/internal/model"
	"mscoin_go/pkg/db/mysqlx"

	"github.com/jmoiron/sqlx"
)

// WithdrawRepository 封装所有对提现历史记录的直接 SQL 访问。
// 提供提现记录的查询和保存功能。
type WithdrawRepository struct {
	db *sqlx.DB // 数据库连接池
}

// NewWithdrawRepository 创建提现仓储实例
// 参数 db 为数据库连接池，由 ServiceContext 提供
func NewWithdrawRepository(db *sqlx.DB) *WithdrawRepository {
	return &WithdrawRepository{db: db}
}

// FindByMemberID 根据会员 ID 分页查询提现记录
// 用于前端展示会员的提现历史
//
// 参数：
//   - ctx: 请求上下文
//   - memberID: 会员 ID
//   - page: 页码，从 1 开始
//   - pageSize: 每页条数
//
// 返回：
//   - list: 提现记录列表
//   - total: 总记录数（用于分页计算）
//   - error: 错误信息
//
// 排序规则：按创建时间倒序，最新申请在前
func (r *WithdrawRepository) FindByMemberID(ctx context.Context, memberID int64, page int64, pageSize int64) ([]*model.WithdrawRecord, int64, error) {
	// 设置默认分页参数
	// 页码从 1 开始，pageSize 默认 10
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 查询总数
	// 用于前端分页组件计算总页数
	var total int64
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM withdraw_record WHERE member_id=?", memberID); err != nil {
		return nil, 0, fmt.Errorf("count withdraw records: %w", err)
	}

	// 查询列表
	// 按创建时间倒序排列，最新申请在前
	// 使用 LIMIT 和 OFFSET 实现分页
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
//
// 事务支持：
//   - exec 参数可以是 DB 或事务执行器
//   - 在提现申请流程中，使用事务执行器
//   - 确保记录创建和余额冻结在同一事务中
//
// 参数：
//   - ctx: 请求上下文
//   - exec: 执行器（可以是 DB 或事务）
//   - record: 提现记录对象
//
// 创建后会回填自增 ID 到 record.Id
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
	// 新创建的记录需要知道自己的 ID，用于后续更新状态
	if insertedID, lastInsertErr := result.LastInsertId(); lastInsertErr == nil {
		record.Id = insertedID
	}
	return nil
}