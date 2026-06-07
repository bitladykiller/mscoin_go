package service

import (
	"context"
	"errors"

	"mscoin_go/app/market/rpc/internal/model"
	"mscoin_go/app/market/rpc/internal/repository"
)

// ExchangeCoinService owns business rules around visible trading pairs.
type ExchangeCoinService struct {
	repo *repository.ExchangeCoinRepository
}

func NewExchangeCoinService(repo *repository.ExchangeCoinRepository) *ExchangeCoinService {
	return &ExchangeCoinService{repo: repo}
}

func (s *ExchangeCoinService) FindVisible(ctx context.Context) ([]*model.ExchangeCoin, error) {
	return s.repo.FindVisible(ctx)
}

func (s *ExchangeCoinService) FindBySymbol(ctx context.Context, symbol string) (*model.ExchangeCoin, error) {
	item, err := s.repo.FindBySymbol(ctx, symbol)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("trading pair not found")
	}
	return item, nil
}
