// Package service contains business-oriented domain services for the market
// module.
package service

import (
	"context"
	"errors"

	"mscoin_go/app/market/rpc/internal/model"
	"mscoin_go/app/market/rpc/internal/repository"
)

// CoinService owns coin-related business rules.
type CoinService struct {
	repo *repository.CoinRepository
}

func NewCoinService(repo *repository.CoinRepository) *CoinService {
	return &CoinService{repo: repo}
}

func (s *CoinService) FindCoinInfo(ctx context.Context, unit string) (*model.Coin, error) {
	coin, err := s.repo.FindByUnit(ctx, unit)
	if err != nil {
		return nil, err
	}
	if coin == nil {
		return nil, errors.New("not support this coin: " + unit)
	}
	return coin, nil
}

func (s *CoinService) FindCoinByID(ctx context.Context, id int64) (*model.Coin, error) {
	coin, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if coin == nil {
		return nil, errors.New("not support this coin")
	}
	return coin, nil
}

func (s *CoinService) FindAll(ctx context.Context) ([]*model.Coin, error) {
	return s.repo.FindAll(ctx)
}
