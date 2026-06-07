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
type ExchangeCoinRepository struct {
	db *sqlx.DB
}

func NewExchangeCoinRepository(db *sqlx.DB) *ExchangeCoinRepository {
	return &ExchangeCoinRepository{db: db}
}

func (r *ExchangeCoinRepository) FindVisible(ctx context.Context) ([]*model.ExchangeCoin, error) {
	var list []*model.ExchangeCoin
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM exchange_coin WHERE visible=?", 1); err != nil {
		return nil, fmt.Errorf("query visible exchange coins: %w", err)
	}
	return list, nil
}

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
