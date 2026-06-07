// Package repository 包含 market RPC 服务的持久化实现。
//
// Repository 层职责：
//   - 封装所有数据库访问逻辑
//   - 提供 CRUD 操作接口
//   - 隔离 SQL/NoSQL 细节，对上层透明
//   - 处理数据库错误转换
//
// 调用链路：
//
//	Domain Service -> Repository -> Database
//
// 本包包含三个 Repository：
//   - CoinRepository：币种数据访问（MySQL）
//   - ExchangeCoinRepository：交易对数据访问（MySQL）
//   - KlineRepository：K 线数据访问（MongoDB）
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"mscoin_go/app/market/rpc/internal/model"

	"github.com/jmoiron/sqlx"
)

// CoinRepository 封装所有对 `coin` 表的直接 SQL 访问。
//
// 数据源：MySQL
// 表名：coin
//
// 提供的查询方法：
//   - FindByUnit：根据 unit 查询单个币种
//   - FindByID：根据 ID 查询单个币种
//   - FindAll：查询所有币种
type CoinRepository struct {
	db *sqlx.DB
}

// NewCoinRepository 创建 CoinRepository 实例。
//
// 参数：
//   - db：MySQL 数据库连接（sqlx 包装）
func NewCoinRepository(db *sqlx.DB) *CoinRepository {
	return &CoinRepository{db: db}
}

// FindByUnit 根据 unit 查询币种。
//
// 查询规则：
//   - unit 为币种单位标识，如 "BTC"、"ETH"、"USDT"
//   - 使用 LIMIT 1 确保只返回一条记录
//   - 记录不存在时返回 nil, nil（而非错误）
//
// 参数：
//   - ctx：请求上下文
//   - unit：币种单位标识
//
// 返回：
//   - *model.Coin：币种信息（不存在时为 nil）
//   - error：数据库错误（不含"记录不存在"）
func (r *CoinRepository) FindByUnit(ctx context.Context, unit string) (*model.Coin, error) {
	var coin model.Coin
	err := r.db.GetContext(ctx, &coin, "SELECT * FROM coin WHERE unit=? LIMIT 1", unit)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query coin by unit %s: %w", unit, err)
	}
	return &coin, nil
}

// FindByID 根据 ID 查询币种。
//
// 查询规则：
//   - ID 为数据库主键
//   - 使用 LIMIT 1 确保只返回一条记录
//   - 记录不存在时返回 nil, nil（而非错误）
//
// 参数：
//   - ctx：请求上下文
//   - id：币种 ID
//
// 返回：
//   - *model.Coin：币种信息（不存在时为 nil）
//   - error：数据库错误（不含"记录不存在"）
func (r *CoinRepository) FindByID(ctx context.Context, id int64) (*model.Coin, error) {
	var coin model.Coin
	err := r.db.GetContext(ctx, &coin, "SELECT * FROM coin WHERE id=? LIMIT 1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query coin by id %d: %w", id, err)
	}
	return &coin, nil
}

// FindAll 查询所有币种。
//
// 查询规则：
//   - 返回所有币种，不区分状态
//   - 不分页，全量返回
//
// 参数：
//   - ctx：请求上下文
//
// 返回：
//   - []*model.Coin：币种列表（可能为空切片）
//   - error：数据库错误
func (r *CoinRepository) FindAll(ctx context.Context) ([]*model.Coin, error) {
	var list []*model.Coin
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM coin"); err != nil {
		return nil, fmt.Errorf("query all coins: %w", err)
	}
	return list, nil
}
