package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"mscoin_go/app/ucenter/rpc/internal/model"

	"github.com/jmoiron/sqlx"
)

type TransactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

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
	where := []string{"member_id = ?"}
	args := []any{memberID}

	if transactionType != "" {
		parsedType, err := model.ParseTransactionType(transactionType)
		if err != nil {
			return nil, 0, err
		}
		where = append(where, "`type` = ?")
		args = append(args, parsedType)
	}

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

	if symbol != "" {
		where = append(where, "symbol = ?")
		args = append(args, symbol)
	}

	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	whereSQL := strings.Join(where, " AND ")
	countQuery := "SELECT COUNT(*) FROM member_transaction WHERE " + whereSQL

	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("count member transaction: %w", err)
	}

	offset := (pageNo - 1) * pageSize
	listQuery := "SELECT * FROM member_transaction WHERE " + whereSQL + " ORDER BY create_time DESC LIMIT ? OFFSET ?"
	listArgs := append(append([]any{}, args...), pageSize, offset)

	var list []*model.MemberTransaction
	if err := r.db.SelectContext(ctx, &list, listQuery, listArgs...); err != nil {
		return nil, 0, fmt.Errorf("query member transaction: %w", err)
	}

	return list, total, nil
}

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
