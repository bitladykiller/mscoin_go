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
	db *sqlx.DB
}

// NewTransactionRepository 创建交易仓储实例
func NewTransactionRepository(db *sqlx.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// FindTransaction 查询会员交易记录
// 支持按交易类型、时间范围、币种筛选，支持分页
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
	where := []string{"member_id = ?"}
	args := []any{memberID}

	// 添加交易类型筛选条件
	if transactionType != "" {
		parsedType, err := model.ParseTransactionType(transactionType)
		if err != nil {
			return nil, 0, err
		}
		where = append(where, "`type` = ?")
		args = append(args, parsedType)
	}

	// 添加时间范围筛选条件
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
	if symbol != "" {
		where = append(where, "symbol = ?")
		args = append(args, symbol)
	}

	// 设置默认分页参数
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 查询总数
	whereSQL := strings.Join(where, " AND ")
	countQuery := "SELECT COUNT(*) FROM member_transaction WHERE " + whereSQL

	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("count member transaction: %w", err)
	}

	// 查询列表
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
		if candidate.dateOnly && isEnd {
			value = value.Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999*time.Millisecond)
		}
		return value.UnixMilli(), nil
	}

	return 0, fmt.Errorf("parse time %q failed", raw)
}