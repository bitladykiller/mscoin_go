package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"mscoin_go/app/exchange/rpc/internal/model"

	"github.com/jmoiron/sqlx"
)

// OrderRepository owns all direct SQL access to exchange orders.
type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Save(ctx context.Context, order *model.ExchangeOrder) error {
	query := `INSERT INTO exchange_order (
		order_id, amount, base_symbol, canceled_time, coin_symbol, completed_time,
		direction, member_id, price, status, symbol, time, traded_amount, turnover,
		type, use_discount
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(
		ctx,
		query,
		order.OrderId,
		order.Amount,
		order.BaseSymbol,
		order.CanceledTime,
		order.CoinSymbol,
		order.CompletedTime,
		order.Direction,
		order.MemberId,
		order.Price,
		order.Status,
		order.Symbol,
		order.Time,
		order.TradedAmount,
		order.Turnover,
		order.Type,
		order.UseDiscount,
	)
	if err != nil {
		return fmt.Errorf("insert exchange order: %w", err)
	}
	return nil
}

func (r *OrderRepository) FindOrderHistory(ctx context.Context, symbol string, page int64, size int64, memberID int64) ([]*model.ExchangeOrder, int64, error) {
	offset := (page - 1) * size
	var list []*model.ExchangeOrder
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM exchange_order WHERE symbol=? AND member_id=? LIMIT ? OFFSET ?", symbol, memberID, size, offset); err != nil {
		return nil, 0, fmt.Errorf("query order history: %w", err)
	}

	var total int64
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM exchange_order WHERE symbol=? AND member_id=?", symbol, memberID); err != nil {
		return nil, 0, fmt.Errorf("count order history: %w", err)
	}
	return list, total, nil
}

func (r *OrderRepository) FindOrderCurrent(ctx context.Context, symbol string, page int64, size int64, memberID int64) ([]*model.ExchangeOrder, int64, error) {
	offset := (page - 1) * size
	var list []*model.ExchangeOrder
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM exchange_order WHERE symbol=? AND member_id=? AND status=? LIMIT ? OFFSET ?", symbol, memberID, model.OrderTrading, size, offset); err != nil {
		return nil, 0, fmt.Errorf("query current orders: %w", err)
	}

	var total int64
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM exchange_order WHERE symbol=? AND member_id=? AND status=?", symbol, memberID, model.OrderTrading); err != nil {
		return nil, 0, fmt.Errorf("count current orders: %w", err)
	}
	return list, total, nil
}

func (r *OrderRepository) FindCurrentTradingCount(ctx context.Context, memberID int64, symbol string, direction int) (int64, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM exchange_order WHERE symbol=? AND member_id=? AND direction=? AND status=?", symbol, memberID, direction, model.OrderTrading)
	if err != nil {
		return 0, fmt.Errorf("count trading orders: %w", err)
	}
	return total, nil
}

func (r *OrderRepository) FindByOrderID(ctx context.Context, orderID string) (*model.ExchangeOrder, error) {
	var order model.ExchangeOrder
	err := r.db.GetContext(ctx, &order, "SELECT * FROM exchange_order WHERE order_id=? LIMIT 1", orderID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query order by id: %w", err)
	}
	return &order, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID string, status int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE exchange_order SET status=? WHERE order_id=?", status, orderID)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return nil
}
