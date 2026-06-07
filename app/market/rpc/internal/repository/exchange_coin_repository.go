package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"mscoin_go/app/market/rpc/internal/model"

	"github.com/jmoiron/sqlx"
)

// ExchangeCoinRepository 持有可见交易对数据的持久化访问。
//
// 数据源：MySQL
// 表名：exchange_coin
//
// 提供的查询方法：
//   - FindVisible：查询所有可见的交易对
//   - FindBySymbol：根据 symbol 查询单个交易对
type ExchangeCoinRepository struct {
	db *sqlx.DB
}

// NewExchangeCoinRepository 创建 ExchangeCoinRepository 实例。
//
// 参数：
//   - db：MySQL 数据库连接（sqlx 包装）
func NewExchangeCoinRepository(db *sqlx.DB) *ExchangeCoinRepository {
	return &ExchangeCoinRepository{db: db}
}

// FindVisible 查询所有可见的交易对。
//
// 查询规则：
//   - 只返回 visible=1 的交易对
//   - 不分页，全量返回
//
// 参数：
//   - ctx：请求上下文
//
// 返回：
//   - []*model.ExchangeCoin：可见交易对列表（可能为空切片）
//   - error：数据库错误
func (r *ExchangeCoinRepository) FindVisible(ctx context.Context) ([]*model.ExchangeCoin, error) {
	var list []*model.ExchangeCoin
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM exchange_coin WHERE visible=?", 1); err != nil {
		return nil, fmt.Errorf("query visible exchange coins: %w", err)
	}
	return list, nil
}

// FindBySymbol 根据 symbol 查询交易对。
//
// 查询规则：
//   - symbol 为交易对标识，格式为 "BASEQUOTE"，如 "BTCUSDT"
//   - 使用 LIMIT 1 确保只返回一条记录
//   - 记录不存在时返回 nil, nil（而非错误）
//
// 参数：
//   - ctx：请求上下文
//   - symbol：交易对标识
//
// 返回：
//   - *model.ExchangeCoin：交易对信息（不存在时为 nil）
//   - error：数据库错误（不含"记录不存在"）
func (r *ExchangeCoinRepository) FindBySymbol(ctx context.Context, symbol string) (*model.ExchangeCoin, error) {
	var item model.ExchangeCoin
	err := r.db.GetContext(ctx, &item, "SELECT * FROM exchange_coin WHERE symbol=? LIMIT 1", symbol)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query exchange coin by symbol %s: %w", symbol, err)
	}
	return &item, nil
}
