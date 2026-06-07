// Package repository 定义了 exchange-rpc 的数据仓库层。
// OrderRepository 负责所有与订单相关的数据库操作，使用 sqlx 进行 SQL 查询。
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"mscoin_go/app/exchange/rpc/internal/model"

	"github.com/jmoiron/sqlx"
)

// OrderRepository 负责所有订单相关的数据库操作。
// 使用 sqlx 进行类型安全的 SQL 查询。
type OrderRepository struct {
	// db 是数据库连接池。
	db *sqlx.DB
}

// NewOrderRepository 创建 OrderRepository 实例。
func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Save 保存新订单到数据库。
// 将订单实体的所有字段插入 exchange_order 表。
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

// FindOrderHistory 查询用户的历史订单列表（已完成/已取消）。
// 支持按交易对过滤和分页查询。
func (r *OrderRepository) FindOrderHistory(ctx context.Context, symbol string, page int64, size int64, memberID int64) ([]*model.ExchangeOrder, int64, error) {
	// 计算分页偏移量
	offset := (page - 1) * size
	var list []*model.ExchangeOrder
	// 查询订单列表
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM exchange_order WHERE symbol=? AND member_id=? LIMIT ? OFFSET ?", symbol, memberID, size, offset); err != nil {
		return nil, 0, fmt.Errorf("query order history: %w", err)
	}

	// 查询总数
	var total int64
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM exchange_order WHERE symbol=? AND member_id=?", symbol, memberID); err != nil {
		return nil, 0, fmt.Errorf("count order history: %w", err)
	}
	return list, total, nil
}

// FindOrderCurrent 查询用户的当前委托订单列表（交易中）。
// 只返回状态为 OrderTrading 的订单。
func (r *OrderRepository) FindOrderCurrent(ctx context.Context, symbol string, page int64, size int64, memberID int64) ([]*model.ExchangeOrder, int64, error) {
	// 计算分页偏移量
	offset := (page - 1) * size
	var list []*model.ExchangeOrder
	// 查询当前委托订单列表
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM exchange_order WHERE symbol=? AND member_id=? AND status=? LIMIT ? OFFSET ?", symbol, memberID, model.OrderTrading, size, offset); err != nil {
		return nil, 0, fmt.Errorf("query current orders: %w", err)
	}

	// 查询总数
	var total int64
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM exchange_order WHERE symbol=? AND member_id=? AND status=?", symbol, memberID, model.OrderTrading); err != nil {
		return nil, 0, fmt.Errorf("count current orders: %w", err)
	}
	return list, total, nil
}

// FindCurrentTradingCount 查询用户当前正在交易中的订单数量。
// 用于检查用户是否超过最大委托订单数限制。
func (r *OrderRepository) FindCurrentTradingCount(ctx context.Context, memberID int64, symbol string, direction int) (int64, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM exchange_order WHERE symbol=? AND member_id=? AND direction=? AND status=?", symbol, memberID, direction, model.OrderTrading)
	if err != nil {
		return 0, fmt.Errorf("count trading orders: %w", err)
	}
	return total, nil
}

// FindByOrderID 根据订单 ID 查询订单。
// 如果订单不存在，返回 nil 而非错误。
func (r *OrderRepository) FindByOrderID(ctx context.Context, orderID string) (*model.ExchangeOrder, error) {
	var order model.ExchangeOrder
	err := r.db.GetContext(ctx, &order, "SELECT * FROM exchange_order WHERE order_id=? LIMIT 1", orderID)
	// 订单不存在时返回 nil
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query order by id: %w", err)
	}
	return &order, nil
}

// UpdateStatus 更新订单状态。
// 用于取消订单或标记订单完成。
func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID string, status int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE exchange_order SET status=? WHERE order_id=?", status, orderID)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return nil
}
