// Package repository 定义交易记录仓储层。
//
// TransactionRepository 封装会员交易记录表（member_transaction）的数据库操作。
// 交易记录是会员资产变动的历史凭证，用于账单查询、资产审计、风控分析。
//
// 交易类型包括：
//   - 充值（RECHARGE）：外部转入
//   - 提现（WITHDRAW）：转出到外部
//   - 转账（TRANSFER_ACCOUNTS）：会员间内部转账
//   - 兑换（EXCHANGE）：币种兑换
//
// 仓储职责：
//   - 查询交易记录：支持按币种、时间、类型筛选
//   - 分页查询：支持分页参数
package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"mscoin_go/app/ucenter/rpc/internal/model"

	"github.com/jmoiron/sqlx"
)

// TransactionRepository 交易仓储
// 负责交易记录数据的持久化操作
type TransactionRepository struct {
	db *sqlx.DB // 数据库连接池
}

// NewTransactionRepository 创建交易仓储实例
// 参数 db 为数据库连接池，由 ServiceContext 提供
func NewTransactionRepository(db *sqlx.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// FindTransaction 查询会员交易记录
// 支持按交易类型、时间范围、币种筛选，支持分页
//
// 参数：
//   - ctx: 请求上下文
//   - memberID: 会员 ID
//   - pageNo: 页码，从 1 开始
//   - pageSize: 每页条数
//   - symbol: 币种符号筛选（可选）
//   - startTime: 开始时间筛选（可选）
//   - endTime: 结束时间筛选（可选）
//   - transactionType: 交易类型筛选（可选）
//
// 返回：
//   - list: 交易记录列表
//   - total: 总记录数（用于分页计算）
//   - error: 错误信息
//
// 时间格式支持：
//   - "2006-01-02 15:04:05"
//   - "2006-01-02 15:04"
//   - "2006-01-02"
//   - RFC3339 格式
func (r *TransactionRepository) FindTransaction(
	ctx context.Context,
	memberID int64,
	pageNo int64,
	pageSize int64,
	symbol string,
	startTime string,
	endTime string,
	transactionType string,
) ([]*model.MemberTransaction, int64, error) {
	// 构建查询条件
	// 基础条件：会员 ID
	where := []string{"member_id = ?"}
	args := []any{memberID}

	// 添加交易类型筛选条件
	// 将字符串类型转换为数字编码
	if transactionType != "" {
		parsedType, err := model.ParseTransactionType(transactionType)
		if err != nil {
			return nil, 0, err
		}
		where = append(where, "`type` = ?")
		args = append(args, parsedType)
	}

	// 添加时间范围筛选条件
	// 支持多种时间格式，自动解析
	if startTime != "" && endTime != "" {
		startMillis, err := parseBoundaryMillis(startTime, false)
		if err != nil {
			return nil, 0, err
		}
		endMillis, err := parseBoundaryMillis(endTime, true)
		if err != nil {
			return nil, 0, err
		}
		where = append(where, "create_time >= ? AND create_time <= ?")
		args = append(args, startMillis, endMillis)
	}

	// 添加币种筛选条件
	// 按币种符号精确匹配
	if symbol != "" {
		where = append(where, "symbol = ?")
		args = append(args, symbol)
	}

	// 设置默认分页参数
	// 页码从 1 开始，pageSize 默认 10
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 查询总数
	// 用于前端分页组件计算总页数
	whereSQL := strings.Join(where, " AND ")
	countQuery := "SELECT COUNT(*) FROM member_transaction WHERE " + whereSQL

	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("count member transaction: %w", err)
	}

	// 查询列表
	// 按创建时间倒序排列，最新记录在前
	// 使用 LIMIT 和 OFFSET 实现分页
	offset := (pageNo - 1) * pageSize
	listQuery := "SELECT * FROM member_transaction WHERE " + whereSQL + " ORDER BY create_time DESC LIMIT ? OFFSET ?"
	listArgs := append(append([]any{}, args...), pageSize, offset)

	var list []*model.MemberTransaction
	if err := r.db.SelectContext(ctx, &list, listQuery, listArgs...); err != nil {
		return nil, 0, fmt.Errorf("query member transaction: %w", err)
	}

	return list, total, nil
}

// parseBoundaryMillis 解析时间边界为毫秒时间戳
// isEnd 为 true 时，对于日期格式会自动设置到当天结束时间
//
// 支持的时间格式：
//   - "2006-01-02 15:04:05" -> 完整时间
//   - "2006-01-02 15:04" -> 省略秒
//   - "2006-01-02" -> 仅日期，isEnd 时自动补充到 23:59:59.999
//   - RFC3339 -> 标准格式
//
// 参数：
//   - raw: 时间字符串
//   - isEnd: 是否为结束时间边界
//
// 返回：毫秒时间戳
func parseBoundaryMillis(raw string, isEnd bool) (int64, error) {
	layouts := []struct {
		layout   string
		dateOnly bool
	}{
		{layout: "2006-01-02 15:04:05"},
		{layout: "2006-01-02 15:04"},
		{layout: "2006-01-02", dateOnly: true},
		{layout: time.RFC3339},
	}

	for _, candidate := range layouts {
		value, err := time.ParseInLocation(candidate.layout, raw, time.Local)
		if err != nil {
			continue
		}
		// 对于仅日期的格式，如果是结束时间边界
		// 自动补充到当天最后一毫秒
		if candidate.dateOnly && isEnd {
			value = value.Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999*time.Millisecond)
		}
		return value.UnixMilli(), nil
	}

	return 0, fmt.Errorf("parse time %q failed", raw)
}