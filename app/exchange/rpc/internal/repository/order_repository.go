// Package repository 定义了 exchange-rpc 的数据仓库层。
// OrderRepository 负责所有与订单相关的数据库操作，使用 sqlx 进行 SQL 查询。
//
// 仓库层设计原则：
// 1. 仓库层只负责数据访问，不包含业务逻辑
// 2. 使用 sqlx 进行类型安全的 SQL 查询
// 3. 所有方法接收 context.Context 参数，支持超时和取消
// 4. 错误使用 fmt.Errorf 包装，保留原始错误信息便于调试
//
// 数据库表结构：
// - exchange_order 表存储订单数据
// - 主要字段：order_id, symbol, member_id, direction, type, price, amount, status 等
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
//
// 数据访问方法：
// - Save: 保存新订单
// - FindOrderHistory: 查询历史订单（已完成/已取消）
// - FindOrderCurrent: 查询当前委托订单（交易中）
// - FindCurrentTradingCount: 查询当前委托订单数量
// - FindByOrderID: 根据 ID 查询订单
// - UpdateStatus: 更新订单状态
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
//
// 插入字段：
// - order_id: 业务订单号
// - amount: 订单数量
// - base_symbol: 基础币种符号
// - coin_symbol: 交易币种符号
// - direction: 订单方向（0:买入, 1:卖出）
// - member_id: 会员 ID
// - price: 订单价格
// - status: 订单状态
// - symbol: 交易对符号
// - time: 创建时间戳
// - type: 订单类型（0:市价, 1:限价）
// - traded_amount: 已成交数量
// - turnover: 已成交金额
// - use_discount: 折扣使用
// - canceled_time: 取消时间（初始为0）
// - completed_time: 完成时间（初始为0）
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
//
// 查询条件：
// - symbol: 交易对符号，为空时查询所有交易对
// - member_id: 会员 ID，限定查询范围
//
// 分页参数：
// - page: 页码，从 1 开始
// - size: 每页记录数
// - offset: 偏移量，计算公式为 (page - 1) * size
//
// 返回值：
// - list: 订单实体列表
// - total: 符合条件的订单总数
//
// 注意：当前实现查询所有状态的订单，未过滤 COMPLETED/CANCELED/OVERTIMED 状态。
// 历史订单应只包含非 TRADING 状态的订单。
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
// 只返回状态为 OrderTrading 的订单，表示订单正在撮合队列中等待成交。
//
// 查询条件：
// - symbol: 交易对符号
// - member_id: 会员 ID
// - status: 订单状态，固定为 OrderTrading（交易中）
//
// 分页参数：
// - page: 页码，从 1 开始
// - size: 每页记录数
// - offset: 偏移量，计算公式为 (page - 1) * size
//
// 返回值：
// - list: 订单实体列表（仅包含 TRADING 状态）
// - total: 符合条件的订单总数
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
//
// 查询条件：
// - member_id: 会员 ID
// - symbol: 交易对符号
// - direction: 订单方向（0:买入, 1:卖出）
// - status: 订单状态，固定为 OrderTrading
//
// 使用场景：
// - 下单前检查：用户同一交易对同一方向的委托订单数量是否超过限制
// - 限制配置：由 market-rpc 的 ExchangeCoin.maxTradingOrder 提供
//
// 返回值：
// - total: 当前委托订单数量
func (r *OrderRepository) FindCurrentTradingCount(ctx context.Context, memberID int64, symbol string, direction int) (int64, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM exchange_order WHERE symbol=? AND member_id=? AND direction=? AND status=?", symbol, memberID, direction, model.OrderTrading)
	if err != nil {
		return 0, fmt.Errorf("count trading orders: %w", err)
	}
	return total, nil
}

// FindByOrderID 根据订单 ID 查询订单。
// 用于查询特定订单的详细信息，如取消订单前验证订单是否存在。
//
// 查询条件：
// - order_id: 业务订单号（格式为 "E" + 时间戳纳秒）
//
// 返回值：
// - 订单存在：返回订单实体
// - 订单不存在：返回 nil（不返回错误）
// - 查询失败：返回错误
//
// 设计说明：
// 订单不存在时返回 nil 而非错误，便于上层逻辑判断。
// 上层可以进一步判断返回值是否为 nil 来决定是否返回"订单不存在"的错误。
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
//
// 使用场景：
// 1. 取消订单：将状态从 TRADING 更新为 CANCELED
// 2. 订单完成：将状态从 TRADING 更新为 COMPLETED
// 3. 订单超时：将状态从 TRADING 更新为 OVERTIMED
//
// 参数：
// - order_id: 业务订单号
// - status: 目标状态（使用 model 中的状态常量）
//
// 注意：当前实现只更新状态字段，未更新 canceled_time 或 completed_time。
// 实际应用中应在取消/完成时记录相应的时间戳。
func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID string, status int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE exchange_order SET status=? WHERE order_id=?", status, orderID)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return nil
}
