// Package logic 定义了 exchange-rpc 的业务逻辑层。
// 每个 Logic 结构体负责处理特定的 RPC 请求，协调领域服务和外部依赖。
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
type AddLogic struct {
	exchangeLogicBase
}

// NewAddLogic 创建 AddLogic 实例。
func NewAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddLogic {
	return &AddLogic{exchangeLogicBase: newExchangeLogicBase(ctx, svcCtx)}
}

// Add 执行新增订单的业务逻辑。
// 处理流程：
// 1. 查询会员信息，验证用户状态
// 2. 查询交易对信息，验证交易对是否支持
// 3. 查询用户钱包信息
// 4. 验证订单请求参数
// 5. 检查用户当前委托订单数量限制
// 6. 构建并保存新订单
func (l *AddLogic) Add(req *orderpb.OrderReq) (*orderpb.AddOrderRes, error) {
	// 查询会员信息
	memberInfo, err := l.svcCtx.MemberClient.FindMemberById(l.ctx, &memberpb.MemberReq{MemberId: req.UserId})
	if err != nil {
		return nil, err
	}

	// 查询交易对信息
	exchangeCoin, err := l.svcCtx.MarketClient.FindSymbolInfo(l.ctx, &marketpb.MarketReq{Symbol: req.Symbol})
	if err != nil || exchangeCoin == nil {
		return nil, errors.New("nonsupport coin")
	}

	// 获取基础币种和交易币种
	baseSymbol := exchangeCoin.GetBaseSymbol()
	coinSymbol := exchangeCoin.GetCoinSymbol()

	// 查询基础币种钱包
	baseWallet, err := l.svcCtx.AssetClient.FindWalletBySymbol(l.ctx, &assetpb.AssetReq{
		UserId:   req.UserId,
		CoinName: baseSymbol,
	})
	if err != nil {
		return nil, fmt.Errorf("find base wallet: %w", err)
	}
	// 查询交易币种钱包
	coinWallet, err := l.svcCtx.AssetClient.FindWalletBySymbol(l.ctx, &assetpb.AssetReq{
		UserId:   req.UserId,
		CoinName: coinSymbol,
	})
	if err != nil {
		return nil, fmt.Errorf("find coin wallet: %w", err)
	}

	// 验证订单请求参数
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
	count, err := l.svcCtx.OrderService.FindCurrentTradingCount(l.ctx, req.UserId, req.Symbol, req.Direction)
	if err != nil {
		return nil, err
	}
	if exchangeCoin.GetMaxTradingOrder() > 0 && count >= exchangeCoin.GetMaxTradingOrder() {
		return nil, fmt.Errorf("too many trading orders, max=%d", exchangeCoin.GetMaxTradingOrder())
	}

	// 构建新订单
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
	if err := l.svcCtx.OrderService.AddOrder(l.ctx, order); err != nil {
		return nil, err
	}

	return &orderpb.AddOrderRes{OrderId: order.OrderId}, nil
}
