package logic

import (
	"context"

	"mscoin_go/app/exchange/rpc/internal/svc"
)

// exchangeLogicBase 是所有 Logic 结构体的基础类。
// 封装了公共的上下文和服务上下文引用，减少代码重复。
//
// 设计模式：模板方法模式
// - 基类提供公共字段和初始化方法
// - 子类继承基类，实现各自的业务逻辑
// - 通过组合 ServiceContext，Logic 可以访问数据库、缓存和 RPC 客户端
type exchangeLogicBase struct {
	// ctx 是请求上下文，用于传递超时和取消信号。
	ctx context.Context
	// svcCtx 是服务上下文，包含所有运行时依赖。
	// 通过 svcCtx 可以访问：
	// - OrderService: 订单领域服务，封装订单业务逻辑
	// - MemberClient: ucenter-rpc 的会员客户端，查询会员信息
	// - AssetClient: ucenter-rpc 的资产客户端，查询钱包信息
	// - MarketClient: market-rpc 的市场客户端，查询交易对信息
	svcCtx *svc.ServiceContext
}

// newExchangeLogicBase 创建 exchangeLogicBase 实例。
func newExchangeLogicBase(ctx context.Context, svcCtx *svc.ServiceContext) exchangeLogicBase {
	return exchangeLogicBase{ctx: ctx, svcCtx: svcCtx}
}
