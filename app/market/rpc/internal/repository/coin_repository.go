// Package repository contains persistence implementations for the market RPC
// service.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"mscoin_go/app/market/rpc/internal/model"

	"github.com/jmoiron/sqlx"
)

// CoinRepository encapsulates all direct SQL access to the `coin` table.
type CoinRepository struct {
	db *sqlx.DB
}

func NewCoinRepository(db *sqlx.DB) *CoinRepository {
	return &CoinRepository{db: db}
}

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

func (r *CoinRepository) FindAll(ctx context.Context) ([]*model.Coin, error) {
	var list []*model.Coin
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM coin"); err != nil {
		return nil, fmt.Errorf("query all coins: %w", err)
	}
	return list, nil
}
