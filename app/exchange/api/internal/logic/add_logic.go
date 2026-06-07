// Package logic 定义了 exchange-api 的业务逻辑层。
// 每个 Logic 结构体负责处理特定的业务请求，调用 RPC 服务完成实际操作。
//
// 架构说明：
// - API 层的 Logic 结构体作为业务逻辑的入口点
// - 通过 ServiceContext 中持有的 RPC 客户端调用下游服务
// - exchange-rpc 是订单服务的核心实现，负责订单的 CRUD 和业务规则验证
package logic

import (
	"context"
	"errors"

	"mscoin_go/app/exchange/api/internal/middleware"
	"mscoin_go/app/exchange/api/internal/svc"
	"mscoin_go/app/exchange/api/internal/types"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

// AddLogic 处理新增订单的业务逻辑。
type AddLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAddLogic 创建 AddLogic 实例。
func NewAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddLogic {
	return &AddLogic{ctx: ctx, svcCtx: svcCtx}
}

// Add 执行新增订单的业务逻辑。
//
// 处理流程：
// 1. 验证请求参数的有效性（Direction 和 Type 必须非空）
// 2. 从上下文获取用户 ID（由认证中间件解析 JWT 后存入）
// 3. 调用 exchange-rpc 的 Add 方法创建订单
// 4. 返回订单 ID
//
// 与 exchange-rpc 的调用关系：
// - 本方法调用 OrderClient.Add() 将订单创建请求转发到 exchange-rpc
// - exchange-rpc 会执行完整的业务规则验证，包括：
//   - 调用 ucenter-rpc.MemberClient.FindMemberById() 查询会员信息，验证用户交易状态
//   - 调用 market-rpc.MarketClient.FindSymbolInfo() 查询交易对配置，验证交易对是否可交易
//   - 调用 ucenter-rpc.AssetClient.FindWalletBySymbol() 查询钱包信息，验证钱包是否被锁定
//   - 检查用户当前委托订单数量是否超过限制
//   - 构建订单实体并保存到数据库
//
// 订单类型说明：
// - MARKET_PRICE: 市价单，按市场当前最优价格立即成交
// - LIMIT_PRICE: 限价单，按指定价格挂单等待成交
//
// 订单方向说明：
// - BUY: 买入方向，用基础币种（如 USDT）购买交易币种（如 BTC）
// - SELL: 卖出方向，用交易币种（如 BTC）换取基础币种（如 USDT）
func (l *AddLogic) Add(req *types.ExchangeReq) (string, error) {
	// 验证订单请求参数
	if !req.OrderValid() {
		return "", errors.New("invalid request")
	}

	// 从上下文获取已认证的用户 ID
	userID := middleware.UserIDFromContext(l.ctx)
	// 调用 RPC 服务创建订单
	resp, err := l.svcCtx.OrderClient.Add(l.ctx, &orderpb.OrderReq{
		Symbol:    req.Symbol,
		UserId:    userID,
		Direction: req.Direction,
		Type:      req.Type,
		Price:     req.Price,
		Amount:    req.Amount,
	})
	if err != nil {
		return "", err
	}
	return resp.OrderId, nil
}
