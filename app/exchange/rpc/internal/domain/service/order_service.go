// Package service 定义了 exchange-rpc 的领域服务层。
// OrderService 封装了交易订单的核心业务规则，是业务逻辑的核心实现。
//
// 领域服务设计原则：
// - 服务层封装复杂的业务规则和跨实体操作
// - 服务层不持有状态，只提供方法
// - 服务层通过仓库层访问数据，不直接操作数据库
// - 服务层可被多个 Logic 结构体共享，避免代码重复
//
// 与仓库层的职责划分：
// - OrderRepository: 数据访问层，负责 CRUD 操作
// - OrderService: 领域服务层，负责业务规则验证和实体构建
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
//
// 服务职责：
// 1. 订单查询：查询历史订单、当前订单、根据 ID 查询订单
// 2. 订单验证：验证下单请求参数的有效性
// 3. 订单构建：构建新订单实体，设置初始状态
// 4. 订单创建：保存订单到数据库
// 5. 订单取消：取消订单，更新状态
//
// 与其他服务的关系：
// - 不直接调用 ucenter-rpc 或 market-rpc，由 Logic 层负责调用
// - Logic 层获取会员信息、钱包信息、交易对信息后，传递给 OrderService 进行验证
// - 这种设计使得 OrderService 可以独立于外部 RPC 服务，便于测试和复用
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
// 该方法接收外部服务查询的结果作为参数，进行统一的业务规则验证。
//
// 验证内容包括：
// 1. 用户交易状态验证
//    - memberInfo.transactionStatus == 1 表示用户可以正常交易
//    - transactionStatus == 0 表示用户被禁止交易
//
// 2. 限价单价格验证
//    - 限价单（LIMIT_PRICE）必须指定价格，且价格 > 0
//    - 市价单（MARKET_PRICE）价格由撮合引擎确定，传入的价格会被忽略
//
// 3. 订单数量验证
//    - 订单数量必须 > 0
//
// 4. 交易对状态验证
//    - exchangeCoin.exchangeable == 1 表示交易对可交易
//    - exchangeCoin.enable == 1 表示交易对已启用
//
// 5. 钱包状态验证
//    - baseWallet.isLock == 0 表示基础币种钱包未锁定
//    - coinWallet.isLock == 0 表示交易币种钱包未锁定
//
// 6. 价格范围验证
//    - 卖出方向：价格必须 >= exchangeCoin.minSellPrice
//    - 买入方向：价格必须 <= exchangeCoin.maxBuyPrice
//
// 7. 市价交易支持验证
//    - 买入市价：exchangeCoin.enableMarketBuy == 1
//    - 卖出市价：exchangeCoin.enableMarketSell == 1
//
// 参数来源说明：
// - memberInfo: 由 Logic 层调用 ucenter-rpc.MemberClient.FindMemberById() 获取
// - exchangeCoin: 由 Logic 层调用 market-rpc.MarketClient.FindSymbolInfo() 获取
// - baseWallet: 由 Logic 层调用 ucenter-rpc.AssetClient.FindWalletBySymbol() 获取
// - coinWallet: 由 Logic 层调用 ucenter-rpc.AssetClient.FindWalletBySymbol() 获取
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
//
// 订单实体构建规则：
// 1. 订单 ID 生成：格式为 "E" + 当前时间戳纳秒，保证全局唯一
// 2. 初始状态：OrderInit（初始化状态），订单创建后由撮合引擎处理
// 3. 时间戳：使用 Unix 毫秒时间戳记录订单创建时间
// 4. 成交信息：初始时 tradedAmount=0, turnover=0
// 5. 折扣使用：初始值为 "0"，表示未使用折扣
// 6. 市价单价格：市价单价格为 0，由撮合引擎在成交时确定
//
// 参数说明：
// - memberID: 会员 ID，标识订单所属用户
// - symbol: 交易对符号，如 "BTCUSDT"
// - baseSymbol: 基础币种符号，如 "USDT"
// - coinSymbol: 交易币种符号，如 "BTC"
// - reqType: 订单类型，"MARKET_PRICE" 或 "LIMIT_PRICE"
// - reqDirection: 订单方向，"BUY" 或 "SELL"
// - price: 订单价格，市价单传 0
// - amount: 订单数量
//
// 返回值：
// - 返回完整的 ExchangeOrder 实体，尚未持久化
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
