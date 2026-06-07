// Package service 定义了 exchange-rpc 的领域服务层。
// OrderService 封装了交易订单的核心业务规则，是业务逻辑的核心实现。
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

// OrderService 封装交易订单的核心业务规则。
// 该服务负责协调仓库层完成订单相关的业务操作，是业务逻辑的核心。
type OrderService struct {
	// repo 是订单数据仓库，负责数据库操作。
	repo *repository.OrderRepository
}

// NewOrderService 创建 OrderService 实例。
func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// FindOrderHistory 查询用户的历史订单列表（已完成/已取消）。
// 返回订单视图列表和总数，支持分页。
func (s *OrderService) FindOrderHistory(ctx context.Context, symbol string, page int64, size int64, memberID int64) ([]*model.OrderView, int64, error) {
	list, total, err := s.repo.FindOrderHistory(ctx, symbol, page, size, memberID)
	if err != nil {
		return nil, 0, err
	}

	// 将数据库实体转换为视图对象
	views := make([]*model.OrderView, len(list))
	for i, item := range list {
		views[i] = item.ToView()
	}
	return views, total, nil
}

// FindOrderCurrent 查询用户的当前委托订单列表（交易中）。
// 返回订单视图列表和总数，支持分页。
func (s *OrderService) FindOrderCurrent(ctx context.Context, symbol string, page int64, size int64, memberID int64) ([]*model.OrderView, int64, error) {
	list, total, err := s.repo.FindOrderCurrent(ctx, symbol, page, size, memberID)
	if err != nil {
		return nil, 0, err
	}

	// 将数据库实体转换为视图对象
	views := make([]*model.OrderView, len(list))
	for i, item := range list {
		views[i] = item.ToView()
	}
	return views, total, nil
}

// FindCurrentTradingCount 查询用户当前正在交易中的订单数量。
// 用于检查用户是否超过最大委托订单数限制。
func (s *OrderService) FindCurrentTradingCount(ctx context.Context, memberID int64, symbol string, direction string) (int64, error) {
	return s.repo.FindCurrentTradingCount(ctx, memberID, symbol, model.DirectionCode(direction))
}

// ValidateAddRequest 集中验证下单请求的所有业务规则。
// 验证内容包括：
// - 用户交易状态是否正常
// - 限价单价格是否有效
// - 订单数量是否有效
// - 交易对是否可交易
// - 钱包是否被锁定
// - 价格是否在允许范围内
// - 是否支持市价交易
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
	// 验证用户交易状态
	if memberInfo.GetTransactionStatus() == 0 {
		return errors.New("this user is forbidden to trade")
	}
	// 验证限价单价格
	if reqType == model.TypeLabels[model.TypeLimitPrice] && price <= 0 {
		return errors.New("limit price mode requires price > 0")
	}
	// 验证订单数量
	if amount <= 0 {
		return errors.New("amount must be > 0")
	}
	// 验证交易对是否可交易
	if exchangeCoin.GetExchangeable() != 1 && exchangeCoin.GetEnable() != 1 {
		return errors.New("coin forbidden")
	}
	// 验证钱包是否被锁定
	if baseWallet.GetIsLock() == 1 || coinWallet.GetIsLock() == 1 {
		return errors.New("wallet locked")
	}
	// 验证卖出价格下限
	if reqDirection == model.DirectionLabels[model.DirectionSell] && exchangeCoin.GetMinSellPrice() > 0 && price < exchangeCoin.GetMinSellPrice() {
		return fmt.Errorf("price must be >= %f", exchangeCoin.GetMinSellPrice())
	}
	// 验证买入价格上限
	if reqDirection == model.DirectionLabels[model.DirectionBuy] && exchangeCoin.GetMaxBuyPrice() > 0 && price > exchangeCoin.GetMaxBuyPrice() {
		return fmt.Errorf("price must be <= %f", exchangeCoin.GetMaxBuyPrice())
	}
	// 验证市价交易支持
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

// BuildNewOrder 构建新订单实体。
// 根据请求参数创建订单对象，并设置初始状态。
// 市价单的价格会被设置为 0，由撮合引擎在成交时确定。
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
		OrderId:      fmt.Sprintf("E%d", time.Now().UnixNano()), // 生成唯一订单 ID
		Amount:       amount,                                    // 订单数量
		BaseSymbol:   baseSymbol,                                // 基础币种符号
		CoinSymbol:   coinSymbol,                                // 交易币种符号
		Direction:    model.DirectionCode(reqDirection),         // 交易方向
		MemberId:     memberID,                                  // 会员 ID
		Price:        price,                                     // 订单价格
		Status:       model.OrderInit,                           // 初始状态
		Symbol:       symbol,                                    // 交易对符号
		Time:         time.Now().UnixMilli(),                    // 创建时间戳（毫秒）
		TradedAmount: 0,                                          // 已成交数量
		Turnover:     0,                                          // 已成交金额
		Type:         model.TypeCode(reqType),                   // 订单类型
		UseDiscount:  "0",                                        // 折扣使用
	}
	// 市价单价格为 0，由撮合引擎确定
	if order.Type == model.TypeMarketPrice {
		order.Price = 0
	}
	return order
}

// AddOrder 保存新订单到数据库。
func (s *OrderService) AddOrder(ctx context.Context, order *model.ExchangeOrder) error {
	return s.repo.Save(ctx, order)
}

// FindByOrderID 根据订单 ID 查询订单。
// 如果订单不存在，返回错误。
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

// CancelOrder 取消订单，将订单状态更新为已取消。
func (s *OrderService) CancelOrder(ctx context.Context, orderID string) error {
	return s.repo.UpdateStatus(ctx, orderID, model.OrderCanceled)
}
