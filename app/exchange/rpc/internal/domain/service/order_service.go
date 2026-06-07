package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mscoin_go/app/exchange/rpc/internal/model"
	"mscoin_go/app/exchange/rpc/internal/repository"
	marketpb "mscoin_go/app/market/rpc/pb/market"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
)

// OrderService owns exchange order business rules.
type OrderService struct {
	repo *repository.OrderRepository
}

func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) FindOrderHistory(ctx context.Context, symbol string, page int64, size int64, memberID int64) ([]*model.OrderView, int64, error) {
	list, total, err := s.repo.FindOrderHistory(ctx, symbol, page, size, memberID)
	if err != nil {
		return nil, 0, err
	}

	views := make([]*model.OrderView, len(list))
	for i, item := range list {
		views[i] = item.ToView()
	}
	return views, total, nil
}

func (s *OrderService) FindOrderCurrent(ctx context.Context, symbol string, page int64, size int64, memberID int64) ([]*model.OrderView, int64, error) {
	list, total, err := s.repo.FindOrderCurrent(ctx, symbol, page, size, memberID)
	if err != nil {
		return nil, 0, err
	}

	views := make([]*model.OrderView, len(list))
	for i, item := range list {
		views[i] = item.ToView()
	}
	return views, total, nil
}

func (s *OrderService) FindCurrentTradingCount(ctx context.Context, memberID int64, symbol string, direction string) (int64, error) {
	return s.repo.FindCurrentTradingCount(ctx, memberID, symbol, model.DirectionCode(direction))
}

// ValidateAddRequest keeps all user-facing business validation for order
// placement in one place.
func (s *OrderService) ValidateAddRequest(
	reqDirection string,
	reqType string,
	price float64,
	amount float64,
	memberInfo *memberpb.MemberInfo,
	exchangeCoin *marketpb.ExchangeCoin,
	baseWallet *assetpb.MemberWallet,
	coinWallet *assetpb.MemberWallet,
) error {
	if memberInfo.GetTransactionStatus() == 0 {
		return errors.New("this user is forbidden to trade")
	}
	if reqType == model.TypeLabels[model.TypeLimitPrice] && price <= 0 {
		return errors.New("limit price mode requires price > 0")
	}
	if amount <= 0 {
		return errors.New("amount must be > 0")
	}
	if exchangeCoin.GetExchangeable() != 1 && exchangeCoin.GetEnable() != 1 {
		return errors.New("coin forbidden")
	}
	if baseWallet.GetIsLock() == 1 || coinWallet.GetIsLock() == 1 {
		return errors.New("wallet locked")
	}
	if reqDirection == model.DirectionLabels[model.DirectionSell] && exchangeCoin.GetMinSellPrice() > 0 && price < exchangeCoin.GetMinSellPrice() {
		return fmt.Errorf("price must be >= %f", exchangeCoin.GetMinSellPrice())
	}
	if reqDirection == model.DirectionLabels[model.DirectionBuy] && exchangeCoin.GetMaxBuyPrice() > 0 && price > exchangeCoin.GetMaxBuyPrice() {
		return fmt.Errorf("price must be <= %f", exchangeCoin.GetMaxBuyPrice())
	}
	if reqType == model.TypeLabels[model.TypeMarketPrice] {
		if reqDirection == model.DirectionLabels[model.DirectionBuy] && exchangeCoin.GetEnableMarketBuy() == 0 {
			return errors.New("market buy is not supported")
		}
		if reqDirection == model.DirectionLabels[model.DirectionSell] && exchangeCoin.GetEnableMarketSell() == 0 {
			return errors.New("market sell is not supported")
		}
	}
	return nil
}

func (s *OrderService) BuildNewOrder(
	memberID int64,
	symbol string,
	baseSymbol string,
	coinSymbol string,
	reqType string,
	reqDirection string,
	price float64,
	amount float64,
) *model.ExchangeOrder {
	order := &model.ExchangeOrder{
		OrderId:      fmt.Sprintf("E%d", time.Now().UnixNano()),
		Amount:       amount,
		BaseSymbol:   baseSymbol,
		CoinSymbol:   coinSymbol,
		Direction:    model.DirectionCode(reqDirection),
		MemberId:     memberID,
		Price:        price,
		Status:       model.OrderInit,
		Symbol:       symbol,
		Time:         time.Now().UnixMilli(),
		TradedAmount: 0,
		Turnover:     0,
		Type:         model.TypeCode(reqType),
		UseDiscount:  "0",
	}
	if order.Type == model.TypeMarketPrice {
		order.Price = 0
	}
	return order
}

func (s *OrderService) AddOrder(ctx context.Context, order *model.ExchangeOrder) error {
	return s.repo.Save(ctx, order)
}

func (s *OrderService) FindByOrderID(ctx context.Context, orderID string) (*model.ExchangeOrder, error) {
	order, err := s.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("orderId not found")
	}
	return order, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID string) error {
	return s.repo.UpdateStatus(ctx, orderID, model.OrderCanceled)
}
