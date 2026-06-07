// Package logic 定义了 exchange-rpc 的业务逻辑层。
// 每个 Logic 结构体负责处理特定的 RPC 请求，协调领域服务和外部依赖。
//
// 架构说明：
// - RPC 层的 Logic 结构体作为业务逻辑的入口点
// - 通过 ServiceContext 访问数据库、缓存和其他 RPC 服务
// - 订单相关的核心业务规则封装在 OrderService 领域服务中
//
// 与其他 RPC 服务的调用关系：
// - ucenter-rpc: 查询会员信息、钱包信息
// - market-rpc: 查询交易对配置、市场行情
package logic

import (
	"context"
	"errors"
	"fmt"

	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
	marketpb "mscoin_go/app/market/rpc/pb/market"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
)

// AddLogic 处理新增订单的 RPC 请求。
// 这是订单创建的核心入口，负责协调多个服务完成订单创建。
type AddLogic struct {
	exchangeLogicBase
}

// NewAddLogic 创建 AddLogic 实例。
func NewAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddLogic {
	return &AddLogic{exchangeLogicBase: newExchangeLogicBase(ctx, svcCtx)}
}

// Add 执行新增订单的业务逻辑。
//
// 订单创建流程（完整）：
// 1. 调用 ucenter-rpc 查询会员信息
//    - 验证用户是否存在
//    - 验证用户交易状态是否正常（transactionStatus == 1）
//
// 2. 调用 market-rpc 查询交易对信息
//    - 验证交易对是否存在且可交易（exchangeable == 1 && enable == 1）
//    - 获取基础币种和交易币种符号
//    - 获取价格限制（minSellPrice, maxBuyPrice）
//    - 获取市价交易支持配置（enableMarketBuy, enableMarketSell）
//    - 获取最大委托订单数限制（maxTradingOrder）
//
// 3. 调用 ucenter-rpc 查询用户钱包信息
//    - 查询基础币种钱包（如 USDT 钱包）
//    - 查询交易币种钱包（如 BTC 钱包）
//    - 验证钱包是否被锁定（isLock == 0）
//
// 4. 调用 OrderService.ValidateAddRequest 验证订单请求参数
//    - 验证用户交易状态
//    - 验证限价单价格 > 0
//    - 验证订单数量 > 0
//    - 验证交易对是否可交易
//    - 验证钱包是否被锁定
//    - 验证卖出价格是否 >= minSellPrice
//    - 验证买入价格是否 <= maxBuyPrice
//    - 验证市价交易是否支持
//
// 5. 检查用户当前委托订单数量
//    - 查询用户当前正在交易中的订单数量
//    - 如果超过 maxTradingOrder 限制，返回错误
//
// 6. 调用 OrderService.BuildNewOrder 构建新订单实体
//    - 生成唯一订单 ID（格式：E + 时间戳纳秒）
//    - 设置订单初始状态为 OrderInit
//    - 设置订单方向、类型、价格、数量等字段
//    - 市价单价格设置为 0
//
// 7. 调用 OrderService.AddOrder 保存订单到数据库
//
// 与 ucenter-rpc 的调用关系：
// - MemberClient.FindMemberById(): 查询会员信息
// - AssetClient.FindWalletBySymbol(): 查询钱包信息
//
// 与 market-rpc 的调用关系：
// - MarketClient.FindSymbolInfo(): 查询交易对配置
//
// 返回值：
// - 成功：返回订单 ID
// - 失败：返回错误信息（如"nonsupport coin"、"wallet locked"、"too many trading orders"等）
func (l *AddLogic) Add(req *orderpb.OrderReq) (*orderpb.AddOrderRes, error) {
	// 查询会员信息
	// 调用 ucenter-rpc 的 MemberClient.FindMemberById 方法
	// 用于验证用户是否存在、用户交易状态是否正常
	memberInfo, err := l.svcCtx.MemberClient.FindMemberById(l.ctx, &memberpb.MemberReq{MemberId: req.UserId})
	if err != nil {
		return nil, err
	}

	// 查询交易对信息
	// 调用 market-rpc 的 MarketClient.FindSymbolInfo 方法
	// 用于验证交易对是否可交易、获取价格限制和配置信息
	exchangeCoin, err := l.svcCtx.MarketClient.FindSymbolInfo(l.ctx, &marketpb.MarketReq{Symbol: req.Symbol})
	if err != nil || exchangeCoin == nil {
		return nil, errors.New("nonsupport coin")
	}

	// 获取基础币种和交易币种
	// 例如：交易对 BTCUSDT，baseSymbol=USDT，coinSymbol=BTC
	baseSymbol := exchangeCoin.GetBaseSymbol()
	coinSymbol := exchangeCoin.GetCoinSymbol()

	// 查询基础币种钱包
	// 调用 ucenter-rpc 的 AssetClient.FindWalletBySymbol 方法
	// 用于验证用户是否有足够的基础币种进行买入
	baseWallet, err := l.svcCtx.AssetClient.FindWalletBySymbol(l.ctx, &assetpb.AssetReq{
		UserId:   req.UserId,
		CoinName: baseSymbol,
	})
	if err != nil {
		return nil, fmt.Errorf("find base wallet: %w", err)
	}
	// 查询交易币种钱包
	// 调用 ucenter-rpc 的 AssetClient.FindWalletBySymbol 方法
	// 用于验证用户是否有足够的交易币种进行卖出
	coinWallet, err := l.svcCtx.AssetClient.FindWalletBySymbol(l.ctx, &assetpb.AssetReq{
		UserId:   req.UserId,
		CoinName: coinSymbol,
	})
	if err != nil {
		return nil, fmt.Errorf("find coin wallet: %w", err)
	}

	// 验证订单请求参数
	// 调用 OrderService.ValidateAddRequest 方法
	// 验证内容包括：用户交易状态、限价单价格、订单数量、交易对状态、钱包状态、价格范围、市价交易支持
	if err := l.svcCtx.OrderService.ValidateAddRequest(
		req.Direction,
		req.Type,
		req.Price,
		req.Amount,
		memberInfo,
		exchangeCoin,
		baseWallet,
		coinWallet,
	); err != nil {
		return nil, err
	}

	// 检查用户当前委托订单数量
	// 调用 OrderService.FindCurrentTradingCount 方法
	// 验证用户是否超过交易对配置的最大委托订单数限制
	count, err := l.svcCtx.OrderService.FindCurrentTradingCount(l.ctx, req.UserId, req.Symbol, req.Direction)
	if err != nil {
		return nil, err
	}
	if exchangeCoin.GetMaxTradingOrder() > 0 && count >= exchangeCoin.GetMaxTradingOrder() {
		return nil, fmt.Errorf("too many trading orders, max=%d", exchangeCoin.GetMaxTradingOrder())
	}

	// 构建新订单
	// 调用 OrderService.BuildNewOrder 方法
	// 生成订单 ID，设置订单属性，初始化状态为 OrderInit
	order := l.svcCtx.OrderService.BuildNewOrder(
		req.UserId,
		req.Symbol,
		baseSymbol,
		coinSymbol,
		req.Type,
		req.Direction,
		req.Price,
		req.Amount,
	)
	// 保存订单到数据库
	// 调用 OrderService.AddOrder 方法
	// 将订单实体持久化到 exchange_order 表
	if err := l.svcCtx.OrderService.AddOrder(l.ctx, order); err != nil {
		return nil, err
	}

	return &orderpb.AddOrderRes{OrderId: order.OrderId}, nil
}
